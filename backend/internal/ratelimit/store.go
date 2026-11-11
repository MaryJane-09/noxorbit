// Package ratelimit enforces database-backed request quotas.
// It deliberately uses the actual TCP peer address, not untrusted
// X-Forwarded-For headers. Trusted proxy handling must be configured
// explicitly when deploying behind a reverse proxy.
package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/netip"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Allow atomically counts attempts in a fixed time window shared by every
// server instance using the same database. All requests count, not just
// unsuccessful ones; this keeps expensive password checks bounded.
func (s *Store) Allow(ctx context.Context, key string, max int, window time.Duration) (bool, error) {
	hash := sha256.Sum256([]byte(key))
	hashKey := hex.EncodeToString(hash[:])
	windowSeconds := int(window / time.Second)

	var attempts int
	err := s.pool.QueryRow(ctx, `INSERT INTO auth_rate_limits (key_hash, window_started_at, attempts)
		VALUES ($1, NOW(), 1)
		ON CONFLICT (key_hash) DO UPDATE
		SET attempts = CASE
		        WHEN auth_rate_limits.window_started_at <= NOW() - make_interval(secs => $2::int)
		        THEN 1 ELSE auth_rate_limits.attempts + 1
		    END,
		    window_started_at = CASE
		        WHEN auth_rate_limits.window_started_at <= NOW() - make_interval(secs => $2::int)
		        THEN NOW() ELSE auth_rate_limits.window_started_at
		    END
		RETURNING attempts`, hashKey, windowSeconds).Scan(&attempts)
	if err != nil {
		return false, err
	}
	return attempts <= max, nil
}

// ClientIP does not trust client-supplied forwarding headers. Configure the
// real client IP at the deployment proxy rather than accepting spoofed values.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	addr, parseErr := netip.ParseAddr(host)
	if parseErr != nil {
		return host
	}
	addr = addr.Unmap()
	if addr.Is6() {
		// Group IPv6 clients by /64 to prevent inexpensive address rotation.
		return netip.PrefixFrom(addr, 64).Masked().String()
	}
	return addr.String()
}

// LimitIP applies a separate quota to each HTTP endpoint and TCP peer.
// It fails closed if the rate-limit database query fails.
func (s *Store) LimitIP(scope string, max int, window time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		key := "ip:" + scope + ":" + ClientIP(r)
		allowed, err := s.Allow(r.Context(), key, max, window)
		if err != nil {
			log.Println("Rate limiter failed:", err)
			writeJSONError(w, http.StatusServiceUnavailable, "Service temporarily unavailable")
			return
		}
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(int(window/time.Second)))
			writeJSONError(w, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
