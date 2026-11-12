package membership

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaryJane-09/noxorbit/backend/internal/auth"
	"github.com/gofrs/uuid"
)

type fakeStore struct {
	list func(context.Context,uuid.UUID,uuid.UUID)([]Member,error)
}
func (f fakeStore) List(ctx context.Context,userID,workspaceID uuid.UUID)([]Member,error) {
	return f.list(ctx,userID,workspaceID)
}
func TestMembersHandlerAuthenticationAndScope(t *testing.T) {
	userID:=uuid.Must(uuid.NewV4())
	workspaceID:=uuid.Must(uuid.NewV4())
	h:=NewHandler(fakeStore{list:func(_ context.Context,uid,wid uuid.UUID)([]Member,error) {
		if uid!=userID || wid!=workspaceID { t.Fatal("scope mismatch") }
		return []Member{{UserID:userID,Name:"Member",Role:"owner"}},nil
	}})
	testRequest:=func(authenticated bool)*http.Request{
		r:=httptest.NewRequest(http.MethodGet,"/members",nil)
		r.SetPathValue("workspace_id",workspaceID.String())
		if authenticated {r=r.WithContext(auth.WithUserID(r.Context(),userID))}
		return r
	}
	unauthorized:=httptest.NewRecorder()
	h.List(unauthorized,testRequest(false))
	if unauthorized.Code!=401 {t.Fatalf("unauthenticated: %d",unauthorized.Code)}
	rec:=httptest.NewRecorder()
	h.List(rec,testRequest(true))
	if rec.Code!=200 {t.Fatalf("status: %d %s",rec.Code,rec.Body.String())}
	var members []Member
	if err:=json.Unmarshal(rec.Body.Bytes(),&members);err!=nil {t.Fatal(err)}
	if len(members)!=1||members[0].Role!="owner" {t.Fatalf("unexpected members %+v",members)}
}
func TestMembersHandlerDoesNotLeakOtherWorkspaces(t *testing.T){
	h:=NewHandler(fakeStore{list:func(context.Context,uuid.UUID,uuid.UUID)([]Member,error){
		return nil,ErrNotFound
	}})
	r:=httptest.NewRequest(http.MethodGet,"/members",nil)
	r.SetPathValue("workspace_id",uuid.Must(uuid.NewV4()).String())
	r=r.WithContext(auth.WithUserID(r.Context(),uuid.Must(uuid.NewV4())))
	rec:=httptest.NewRecorder()
	h.List(rec,r)
	if rec.Code!=404 {t.Fatalf("inaccessible workspace: %d",rec.Code)}
}
