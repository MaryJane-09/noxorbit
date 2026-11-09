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
	if got := ClientIP(req); got != "2001:db8::2" {
		t.Fatalf("got IP %q, want IPv6 peer", got)
	}
}
