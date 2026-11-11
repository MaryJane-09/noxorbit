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

func TestWorkspaceRoutesRequireAuthentication(t *testing.T) {
	handler := New(nil)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/workspaces"},
		{http.MethodPost, "/api/v1/workspaces"},
		{http.MethodGet, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1"},
		{http.MethodPatch, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: status %d, want 401", tc.method, tc.path, rec.Code)
		}
	}
}

func TestProjectRoutesRequireAuthentication(t *testing.T) {
	handler := New(nil)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1/projects"},
		{http.MethodPost, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1/projects"},
		{http.MethodGet, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1/projects/bba96e70-f1b2-478f-a4c3-6684fd7f26ee"},
		{http.MethodPatch, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1/projects/bba96e70-f1b2-478f-a4c3-6684fd7f26ee"},
		{http.MethodPatch, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1/projects/bba96e70-f1b2-478f-a4c3-6684fd7f26ee/status"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s returned %d, want 401", tc.method, tc.path, rec.Code)
		}
	}
}
