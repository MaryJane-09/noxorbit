package ratelimit

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPIgnoresSpoofedForwardingHeaders(t *testing.T) {
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "192.0.2.10:41002"
	req.Header.Set("X-Forwarded-For", "203.0.113.77")
	req.Header.Set("X-Real-IP", "203.0.113.88")
	if got := ClientIP(req); got != "192.0.2.10" {
		t.Fatalf("got IP %q, want TCP peer 192.0.2.10", got)
	}
}

func TestClientIPSupportsIPv6(t *testing.T) {
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "[2001:db8::2]:45000"
	if got := ClientIP(req); got != "2001:db8::/64" {
		t.Fatalf("got IP %q, want IPv6 /64 bucket", got)
	}
}

func TestClientIPGroupsIPv6Prefix(t *testing.T) {
	a := httptest.NewRequest("POST", "/login", nil)
	a.RemoteAddr = "[2001:db8:1234:5678::1]:41002"
	b := httptest.NewRequest("POST", "/login", nil)
	b.RemoteAddr = "[2001:db8:1234:5678:ffff::c]:41003"
	if got, want := ClientIP(a), ClientIP(b); got != want || got != "2001:db8:1234:5678::/64" {
		t.Fatalf("different IPv6 buckets: %q / %q", got, want)
	}
}
