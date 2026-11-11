package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaryJane-09/noxorbit/backend/internal/auth"
	"github.com/gofrs/uuid"
)

type fakeStore struct {
	create func(context.Context, uuid.UUID, string, string) (Workspace, error)
	list func(context.Context, uuid.UUID) ([]Workspace, error)
	get func(context.Context, uuid.UUID, uuid.UUID) (Workspace, error)
	update func(context.Context, uuid.UUID, uuid.UUID, string, string) (Workspace, error)
}

func (f fakeStore) Create(ctx context.Context, ownerID uuid.UUID, name, purpose string) (Workspace, error) {
	if f.create == nil { panic("unexpected Create call") }
	return f.create(ctx, ownerID, name, purpose)
}
func (f fakeStore) ListForUser(ctx context.Context, id uuid.UUID) ([]Workspace, error) {
	if f.list == nil { panic("unexpected ListForUser call") }
	return f.list(ctx, id)
}
func (f fakeStore) GetForUser(ctx context.Context, userID, workspaceID uuid.UUID) (Workspace, error) {
	if f.get == nil { panic("unexpected GetForUser call") }
	return f.get(ctx, userID, workspaceID)
}

func (f fakeStore) UpdateForOwner(ctx context.Context, userID, workspaceID uuid.UUID, name, purpose string) (Workspace, error) {
	if f.update == nil { panic("unexpected UpdateForOwner call") }
	return f.update(ctx, userID, workspaceID, name, purpose)
}

func fixtureID(t *testing.T, value string) uuid.UUID {
	t.Helper()
	id, err := uuid.FromString(value)
	if err != nil { t.Fatal(err) }
	return id
}

func signedIn(req *http.Request, userID uuid.UUID) *http.Request {
	return req.WithContext(auth.WithUserID(req.Context(), userID))
}

func TestNormalizeCreate(t *testing.T) {
	tests := []struct {
		name, inputName, purpose, wantName, wantPurpose string
		wantErr error
	}{
		{name: "default purpose", inputName: "  Study Group  ", wantName: "Study Group", wantPurpose: "general"},
		{name: "case normalized", inputName: "  NoxOrbit Team ", purpose: "  EDUCATION  ", wantName: "NoxOrbit Team", wantPurpose: "education"},
		{name: "unicode name", inputName: "Étude ✨", purpose: "development", wantName: "Étude ✨", wantPurpose: "development"},
		{name: "empty name", inputName: " \t  ", wantErr: ErrInvalidName},
		{name: "long name", inputName: strings.Repeat("a", 101), wantErr: ErrInvalidName},
		{name: "unsupported purpose", inputName: "Team", purpose: "private", wantErr: ErrInvalidPurpose},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCreate(CreateRequest{Name: tt.inputName, Purpose: tt.purpose})
			if !errors.Is(err, tt.wantErr) { t.Fatalf("error = %v, want %v", err, tt.wantErr) }
			if err == nil && (got.Name != tt.wantName || got.Purpose != tt.wantPurpose) {
				t.Fatalf("got %#v, want name %q and purpose %q", got, tt.wantName, tt.wantPurpose)
			}
		})
	}
}

func TestCollectionRequiresSession(t *testing.T) {
	h := NewHandler(fakeStore{})
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "/api/v1/workspaces", nil)
		rec := httptest.NewRecorder()
		h.Collection(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s: got %d, want 401", method, rec.Code)
		}
	}
}

func TestCreateWorkspaceUsesSessionIdentity(t *testing.T) {
	signedInID := fixtureID(t, "d4a3a5cc-3dd4-4b99-920a-199629ca73a3")
	called := false
	h := NewHandler(fakeStore{create: func(_ context.Context, owner uuid.UUID, name, purpose string) (Workspace, error) {
		called = true
		if owner != signedInID { t.Fatalf("owner = %v, want session id", owner) }
		if name != "Study Group" || purpose != "education" {
			t.Fatalf("normalized name/purpose = %q, %q", name, purpose)
		}
		return Workspace{ID: fixtureID(t, "bba96e70-f1b2-478f-a4c3-6684fd7f26ee"), Name: name, Purpose: purpose, Role: "owner"}, nil
	}})
	req := signedIn(httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(`{"name":"  Study Group  ","purpose":"EDUCATION"}`)), signedInID)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Collection(rec, req)
	if rec.Code != http.StatusCreated || !called {
		t.Fatalf("got status %d, called = %v, body %s", rec.Code, called, rec.Body.String())
	}
	var created Workspace
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil { t.Fatal(err) }
	if created.Role != "owner" || created.Purpose != "education" { t.Fatalf("unexpected created workspace: %#v", created) }
}

func TestCreateWorkspaceRejectsInvalidBodies(t *testing.T) {
	uid := fixtureID(t, "cf1796c4-626e-440f-9d9f-701846f610c3")
	tests := []struct{ title, body, contentType string; want int }{
		{"missing name", `{"purpose":"general"}`, "application/json", 400},
		{"unknown purpose", `{"name":"Team","purpose":"secret"}`, "application/json", 400},
		{"attempt to spoof owner", `{"name":"Team","owner_id":"fake"}`, "application/json", 400},
		{"extra JSON", `{"name":"Team"} {"name":"Another"}`, "application/json", 400},
		{"malformed JSON", `{"name":`, "application/json", 400},
		{"body too large", `{"name":"` + strings.Repeat("a", maxCreateBodyBytes) + `"}`, "application/json", 400},
		{"form body", "name=Team", "application/x-www-form-urlencoded", 415},
	}
	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			h := NewHandler(fakeStore{})
			req := signedIn(httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(tc.body)), uid)
			req.Header.Set("Content-Type", tc.contentType)
			rec := httptest.NewRecorder()
			h.Collection(rec, req)
			if rec.Code != tc.want { t.Fatalf("got %d, want %d: %s", rec.Code, tc.want, rec.Body.String()) }
		})
	}
}

func TestListWorkspacesOnlyUsesSignedInID(t *testing.T) {
	uid := fixtureID(t, "2ac8bfe6-a12f-4e99-a5c7-13b25d311f05")
	h := NewHandler(fakeStore{list: func(_ context.Context, userID uuid.UUID) ([]Workspace, error) {
		if userID != uid { t.Fatalf("list id = %v, want %v", userID, uid) }
		return []Workspace{}, nil
	}})
	req := signedIn(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil), uid)
	rec := httptest.NewRecorder()
	h.Collection(rec, req)
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Fatalf("got %d %s, want 200 []", rec.Code, rec.Body.String())
	}
}

func TestDetailIsMembershipScopedAndHidesUnknown(t *testing.T) {
	uid := fixtureID(t, "14efcf1e-bb4d-43b1-a09d-ebed749d2fb7")
	wid := fixtureID(t, "052f92b9-e53d-4499-b7f7-6d8c4b157cb1")
	h := NewHandler(fakeStore{get: func(_ context.Context, userID, workspaceID uuid.UUID) (Workspace, error) {
		if userID != uid || workspaceID != wid { t.Fatalf("got identities %s, %s", userID, workspaceID) }
		return Workspace{}, ErrNotFound
	}})
	req := signedIn(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+wid.String(), nil), uid)
	req.SetPathValue("id", wid.String())
	rec := httptest.NewRecorder()
	h.Detail(rec, req)
	if rec.Code != http.StatusNotFound { t.Fatalf("status = %d, want 404", rec.Code) }

	bad := signedIn(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/not-a-uuid", nil), uid)
	bad.SetPathValue("id", "not-a-uuid")
	out := httptest.NewRecorder()
	h.Detail(out, bad)
	if out.Code != http.StatusBadRequest { t.Fatalf("invalid ID status = %d, want 400", out.Code) }
}

func TestWorkspaceErrorsNeverLeakDatabaseDetails(t *testing.T) {
	uid := fixtureID(t, "14efcf1e-bb4d-43b1-a09d-ebed749d2fb7")
	h := NewHandler(fakeStore{list: func(context.Context, uuid.UUID) ([]Workspace, error) {
		return nil, errors.New("database_secret: connection failed")
	}})
	req := signedIn(httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil), uid)
	rec := httptest.NewRecorder()
	h.Collection(rec, req)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "database_secret") {
		t.Fatalf("server error response: %d %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateWorkspaceRequiresOwnerAndValidInput(t *testing.T) {
	uid := fixtureID(t, "14efcf1e-bb4d-43b1-a09d-ebed749d2fb7")
	wid := fixtureID(t, "052f92b9-e53d-4499-b7f7-6d8c4b157cb1")
	var called int
	h := NewHandler(fakeStore{update: func(_ context.Context, userID, workspaceID uuid.UUID, name, purpose string) (Workspace, error) {
		called++
		if userID != uid || workspaceID != wid { t.Fatal("update must use session and route IDs") }
		if name != "Renamed Team" || purpose != "business" { t.Fatalf("unexpected settings %q/%q", name, purpose) }
		return Workspace{ID: wid, Name: name, Purpose: purpose, Role: "owner"}, nil
	}})
	makeRequest := func(body string) *http.Request {
		req := signedIn(httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wid.String(), strings.NewReader(body)), uid)
		req.SetPathValue("id", wid.String())
		req.Header.Set("Content-Type", "application/json")
		return req
	}
	rec := httptest.NewRecorder()
	h.Detail(rec, makeRequest(`{"name":"  Renamed Team  ","purpose":"BUSINESS"}`))
	if rec.Code != http.StatusOK || called != 1 { t.Fatalf("owner edit = %d %s", rec.Code, rec.Body.String()) }

	for _, badBody := range []string{
		`{"name":" "}`,
		`{"name":"Team","purpose":"unknown"}`,
		`{"name":"Team","user_id":"spoof"}`,
		`{"name":"Team"} {"name":"Other"}`,
	} {
		rec := httptest.NewRecorder()
		h.Detail(rec, makeRequest(badBody))
		if rec.Code != http.StatusBadRequest { t.Fatalf("bad body %s got %d", badBody, rec.Code) }
	}
	if called != 1 { t.Fatalf("invalid settings reached repository %d times", called-1) }
}

func TestUpdateWorkspaceMemberCannotRename(t *testing.T) {
	uid := fixtureID(t, "14efcf1e-bb4d-43b1-a09d-ebed749d2fb7")
	wid := fixtureID(t, "052f92b9-e53d-4499-b7f7-6d8c4b157cb1")
	for _, tc := range []struct{ repoErr error; want int }{
		{ErrForbidden, http.StatusForbidden},
		{ErrNotFound, http.StatusNotFound},
		{errors.New("hidden connection string"), http.StatusInternalServerError},
	} {
		h := NewHandler(fakeStore{update: func(context.Context, uuid.UUID, uuid.UUID, string, string) (Workspace, error) {
			return Workspace{}, tc.repoErr
		}})
		req := signedIn(httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/"+wid.String(), strings.NewReader(`{"name":"Team","purpose":"general"}`)), uid)
		req.SetPathValue("id", wid.String())
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Detail(rec, req)
		if rec.Code != tc.want || strings.Contains(rec.Body.String(), "hidden connection string") {
			t.Fatalf("update response status %d body %s", rec.Code, rec.Body.String())
		}
	}
}

func TestUpdateWorkspaceRequiresSession(t *testing.T) {
	h := NewHandler(fakeStore{})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/workspaces/052f92b9-e53d-4499-b7f7-6d8c4b157cb1", nil)
	req.SetPathValue("id", "052f92b9-e53d-4499-b7f7-6d8c4b157cb1")
	rec := httptest.NewRecorder()
	h.Detail(rec, req)
	if rec.Code != http.StatusUnauthorized { t.Fatalf("got %d, want 401", rec.Code) }
}
