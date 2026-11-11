package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllUsersListingIsNotExposed(t *testing.T) {
	// Router construction does not require a live database; this route
	// should return 404 without invoking authentication or any repository.
	handler := New(nil)
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "/users", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s /users returned %d, want 404", method, rec.Code)
		}
	}
}
