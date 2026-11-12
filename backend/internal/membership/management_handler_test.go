package membership

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type managementFake struct {
    list func(context.Context, uuid.UUID, uuid.UUID)([]InviteSummary,error)
    revoke func(context.Context, uuid.UUID,uuid.UUID,uuid.UUID)error
}
func (f managementFake) ListInvites(ctx context.Context,a,w uuid.UUID)([]InviteSummary,error){return f.list(ctx,a,w)}
func (f managementFake) RevokeInvite(ctx context.Context,a,w,i uuid.UUID)error{return f.revoke(ctx,a,w,i)}

type removalFake struct {
    remove func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)error
}
func (f removalFake) RemoveMember(ctx context.Context,a,w,i uuid.UUID)error{return f.remove(ctx,a,w,i)}

func managementRequest(method,pathVar string,withAuth bool)*http.Request{
    r:=httptest.NewRequest(method,"/management",nil)
    r.SetPathValue("workspace_id",uuid.Must(uuid.NewV4()).String())
    r.SetPathValue("invitation_id",pathVar)
    r.SetPathValue("member_id",pathVar)
    if withAuth {r=r.WithContext(auth.WithUserID(r.Context(),uuid.Must(uuid.NewV4())))}
    return r
}
func TestManagementListRequiresOwner(t *testing.T){
    h:=NewManagementHandler(managementFake{
        list:func(context.Context,uuid.UUID,uuid.UUID)([]InviteSummary,error){return nil,ErrOwnerRequired},
    },removalFake{})
    rec:=httptest.NewRecorder()
    h.List(rec,managementRequest("GET","",false))
    if rec.Code!=401 {t.Fatalf("anonymous list: %d",rec.Code)}
    rec=httptest.NewRecorder()
    h.List(rec,managementRequest("GET","",true))
    if rec.Code!=403 {t.Fatalf("member list: %d",rec.Code)}
}
func TestManagementRevokeAndRemove(t *testing.T){
    id:=uuid.Must(uuid.NewV4())
    h:=NewManagementHandler(managementFake{
        revoke:func(_ context.Context,_,_,inviteID uuid.UUID)error{
            if inviteID!=id {t.Fatal("wrong invitation")}
            return nil
        },
    },removalFake{remove:func(_ context.Context,_,_,target uuid.UUID)error{
        if target!=id {t.Fatal("wrong member")}
        return ErrCannotRemoveOwner
    }})
    rec:=httptest.NewRecorder()
    h.Revoke(rec,managementRequest("DELETE",id.String(),true))
    if rec.Code!=204 {t.Fatalf("revoke: %d",rec.Code)}
    rec=httptest.NewRecorder()
    h.Remove(rec,managementRequest("DELETE",id.String(),true))
    if rec.Code!=403 {t.Fatalf("owner protected: %d",rec.Code)}
}
func TestManagementListDoesNotReturnToken(t *testing.T) {
    id:=uuid.Must(uuid.NewV4())
    h:=NewManagementHandler(managementFake{
        list:func(context.Context,uuid.UUID,uuid.UUID)([]InviteSummary,error){
            return []InviteSummary{{ID:id,Kind:"link",ExpiresAt:time.Now().Add(time.Hour),MaxUses:1}},nil
        },
    },removalFake{})
    rec:=httptest.NewRecorder()
    h.List(rec,managementRequest("GET","",true))
    if rec.Code!=200 {t.Fatalf("list: %d",rec.Code)}
    if contains:=rec.Body.String(); len(contains)==0 {t.Fatal("empty response")}
}
