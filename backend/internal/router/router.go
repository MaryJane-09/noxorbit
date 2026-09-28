package router

import (
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/config"
	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/health"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/pending"
	"github.com/MaryJane-09/noxorbit/backend/internal/register"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func New(pool *pgxpool.Pool) http.Handler {

	sender := email.EmailSender{
		GmailAddress:     config.AppConfig.GmailAddress,
		GmailAppPassword: config.AppConfig.GmailAppPassword,
		SMTPHost:         "smtp.gmail.com",
		SMTPPort:         "587",
	}
	repo := user.NewRepository(pool)
	otpRepo := otp.NewRepository(pool)
	pendingRepo := pending.NewRepository(pool)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.HealthHandler)
	mux.HandleFunc("/register", register.RegisterHandler(repo, otpRepo, pendingRepo, &sender))
	mux.HandleFunc("/users", user.UsersHandler(repo))
	mux.HandleFunc("/verify-otp", register.VerifyOTPHandler(repo, otpRepo, pendingRepo))
	return corsMiddleware(mux)
}
