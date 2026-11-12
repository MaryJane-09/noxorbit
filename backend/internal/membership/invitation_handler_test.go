package membership

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)
type fakeInvites struct {
    create func(context.Context,uuid.UUID,uuid.UUID)(string,error)
    accept func(context.Context,uuid.UUID,string)(uuid.UUID,error)
}
func (f fakeInvites) CreateLink(ctx context.Context,u,w uuid.UUID)(string,error){return f.create(ctx,u,w)}
func (f fakeInvites) AcceptLink(ctx context.Context,u uuid.UUID,t string)(uuid.UUID,error){return f.accept(ctx,u,t)}

func TestCreateInvitationRequiresOwner(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    wid:=uuid.Must(uuid.NewV4())
    handler:=NewInvitationHandler(fakeInvites{create:func(_ context.Context,userID,workspaceID uuid.UUID)(string,error){
        if userID!=uid||workspaceID!=wid {t.Fatal("wrong scope")}
        return "",ErrOwnerRequired
    }})
    request:=httptest.NewRequest(http.MethodPost,"/invitations",nil)
    request.SetPathValue("workspace_id",wid.String())
    rec:=httptest.NewRecorder()
    handler.Create(rec,request)
    if rec.Code!=401 {t.Fatalf("anonymous invite: %d",rec.Code)}
    request=request.WithContext(auth.WithUserID(request.Context(),uid))
    rec=httptest.NewRecorder()
    handler.Create(rec,request)
    if rec.Code!=403 {t.Fatalf("non-owner invite: %d",rec.Code)}
}
func TestAcceptInvitationRejectsInvalidBodies(t *testing.T) {
    uid:=uuid.Must(uuid.NewV4())
    handler:=NewInvitationHandler(fakeInvites{accept:func(context.Context,uuid.UUID,string)(uuid.UUID,error){
        t.Fatal("invalid request reached store")
        return uuid.Nil,nil
    }})
    for _,body:=range []string{`{"token":"value"} {"token":"extra"}`,`{"wrong":"value"}`,`{"token":`} {
        req:=httptest.NewRequest(http.MethodPost,"/accept",strings.NewReader(body))
        req.Header.Set("Content-Type","application/json")
        req=req.WithContext(auth.WithUserID(req.Context(),uid))
        rec:=httptest.NewRecorder()
        handler.Accept(rec,req)
        if rec.Code!=400 {t.Fatalf("%q: %d %s",body,rec.Code,rec.Body.String())}
    }
}
func TestAcceptInvitationUsesAuthenticatedIdentity(t *testing.T) {
    uid:=uuid.Must(uuid.NewV4())
    wid:=uuid.Must(uuid.NewV4())
    handler:=NewInvitationHandler(fakeInvites{accept:func(_ context.Context,userID uuid.UUID,token string)(uuid.UUID,error){
        if userID!=uid || token!="testtoken" {t.Fatal("unexpected identity or token")}
        return wid,nil
    }})
    req:=httptest.NewRequest(http.MethodPost,"/accept",strings.NewReader(`{"token":"testtoken"}`))
    req.Header.Set("Content-Type","application/json")
    req=req.WithContext(auth.WithUserID(req.Context(),uid))
    rec:=httptest.NewRecorder()
    handler.Accept(rec,req)
    if rec.Code!=200 || !strings.Contains(rec.Body.String(),wid.String()) {
        t.Fatalf("%d %s",rec.Code,rec.Body.String())
    }
}
