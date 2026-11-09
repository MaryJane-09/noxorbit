package register

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyOTPRequiresRegistrationDetails(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "old email and code only payload is rejected",
			body: `{"email":"person@example.com","code":"abc234"}`,
		},
		{
			name: "missing name is rejected",
			body: `{"email":"person@example.com","password":"strongPass123","code":"abc234"}`,
		},
		{
			name: "missing password is rejected",
			body: `{"name":"Sample Person","email":"person@example.com","code":"abc234"}`,
		},
		{
			name: "short password is rejected",
			body: `{"name":"Sample Person","email":"person@example.com","password":"short","code":"abc234"}`,
		},
		{
			name: "long password is rejected",
			body: `{"name":"Sample Person","email":"person@example.com","password":"` + strings.Repeat("x", 73) + `","code":"abc234"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/verify-otp", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()
			// Invalid inputs must be rejected before the database is touched.
			VerifyOTPHandler(nil, nil).ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestRegisterRejectsInvalidInputWithoutDatabase(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(
		`{"name":"Sample Person","email":"person@example.com","password":"123"}`,
	))
	recorder := httptest.NewRecorder()
	RegisterHandler(nil, nil).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
