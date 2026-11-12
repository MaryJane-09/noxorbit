package membership

import (
 "context"
 "errors"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)

type fakeEmailStore struct {
 create func(context.Context,uuid.UUID,uuid.UUID,string)(string,uuid.UUID,string,error)
 revoked bool
}
func (f *fakeEmailStore) CreateEmail(ctx context.Context,actor,workspace uuid.UUID,address string)(string,uuid.UUID,string,error) {
 return f.create(ctx,actor,workspace,address)
}
func (f *fakeEmailStore) RevokeFailedEmail(context.Context,uuid.UUID)error { f.revoked=true;return nil }

type fakeEmailSender struct{ failure error; sent bool }
func (f *fakeEmailSender) SendWorkspaceInvitation(address,name,link string)error{
 f.sent=true
 if !strings.Contains(link,"#invite/") {return errors.New("missing token")}
 return f.failure
}

func TestNormalizeInviteEmail(t *testing.T) {
 for _,input:=range []string{"","not-an-email","a@example.com,b@example.com","Tester <a@example.com>","a@example.com\nBcc:test@example.com"} {
  if _,err:=NormalizeInviteEmail(input);err==nil {t.Errorf("accepted %q",input)}
 }
 address,err:=NormalizeInviteEmail("  MEMBER@Example.com  ")
 if err!=nil||address!="member@example.com"{t.Fatalf("%q %v",address,err)}
}
func TestTrustedInvitationBase(t *testing.T){
 for _,address:=range []string{"https://noxorbit.example","http://localhost:5173"} {
  if _,err:=trustedInvitationBase(address);err!=nil{t.Fatal(err)}
 }
 for _,address:=range []string{"http://example.com","https://example.com/other","https://example.com#bad"}{
  if _,err:=trustedInvitationBase(address);err==nil{t.Errorf("accepted %s",address)}
 }
}
func TestEmailInviteRequiresOwner(t *testing.T){
 t.Setenv("FRONTEND_URL","https://noxorbit.example")
 uid:=uuid.Must(uuid.NewV4())
 wid:=uuid.Must(uuid.NewV4())
 store:=&fakeEmailStore{create:func(_ context.Context,actor,workspace uuid.UUID,email string)(string,uuid.UUID,string,error){
  if uid!=actor||wid!=workspace||email!="member@example.com"{t.Fatal("unexpected request")}
  return "",uuid.Nil,"",ErrOwnerRequired
 }}
 sender:=&fakeEmailSender{}
 handler:=NewEmailInviteHandler(store,sender)
 req:=httptest.NewRequest(http.MethodPost,"/invite",strings.NewReader(`{"email":"member@example.com"}`))
 req.SetPathValue("workspace_id",wid.String())
 req.Header.Set("Content-Type","application/json")
 req=req.WithContext(auth.WithUserID(req.Context(),uid))
 rec:=httptest.NewRecorder()
 handler.Send(rec,req)
 if rec.Code!=403 || sender.sent {t.Fatalf("status %d sent %v",rec.Code,sender.sent)}
}
func TestEmailFailureRevokesInvitation(t *testing.T){
 t.Setenv("FRONTEND_URL","https://noxorbit.example")
 uid:=uuid.Must(uuid.NewV4())
 wid:=uuid.Must(uuid.NewV4())
 inviteID:=uuid.Must(uuid.NewV4())
 store:=&fakeEmailStore{create:func(_ context.Context,_,_ uuid.UUID,_ string)(string,uuid.UUID,string,error){
  return "secret",inviteID,"Workspace",nil
 }}
 sender:=&fakeEmailSender{failure:errors.New("SMTP offline")}
 handler:=NewEmailInviteHandler(store,sender)
 req:=httptest.NewRequest(http.MethodPost,"/invite",strings.NewReader(`{"email":"member@example.com"}`))
 req.Header.Set("Content-Type","application/json")
 req.SetPathValue("workspace_id",wid.String())
 req=req.WithContext(auth.WithUserID(req.Context(),uid))
 rec:=httptest.NewRecorder()
 handler.Send(rec,req)
 if rec.Code!=502 || !store.revoked {t.Fatalf("status %d, revoked %v",rec.Code,store.revoked)}
}
