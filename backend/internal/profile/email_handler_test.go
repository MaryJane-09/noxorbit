package profile

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

type emailTestStore struct{
 issue func(context.Context,uuid.UUID,string,string,string)(string,error)
 discard func(context.Context,uuid.UUID,string)error
 confirm func(context.Context,uuid.UUID,string)(Profile,error)
}
func(s emailTestStore)IssueEmailChange(ctx context.Context,u uuid.UUID,p,a,h string)(string,error){return s.issue(ctx,u,p,a,h)}
func(s emailTestStore)DiscardEmailChange(ctx context.Context,u uuid.UUID,hash string)error{
 if s.discard==nil{return nil}
 return s.discard(ctx,u,hash)
}
func(s emailTestStore)ConfirmEmailChange(ctx context.Context,u uuid.UUID,code string)(Profile,error){return s.confirm(ctx,u,code)}
type emailTestSender struct{send func(string,string,string)error}
func(s emailTestSender)SendEmailChangeCode(name,to,code string)error{return s.send(name,to,code)}
func emailRequest(body string,uid uuid.UUID)*http.Request {
 r:=httptest.NewRequest("POST","/api/v1/profile/email/request",strings.NewReader(body))
 r.Header.Set("Content-Type","application/json")
 return r.WithContext(auth.WithUserID(r.Context(),uid))
}
func TestEmailRequestRequiresSession(t *testing.T){
 called:=false
 h:=NewEmailChangeHandler(emailTestStore{issue:func(context.Context,uuid.UUID,string,string,string)(string,error){
  called=true;return "",nil
 }},emailTestSender{send:func(string,string,string)error{called=true;return nil}})
 w:=httptest.NewRecorder()
 h.Request(w,httptest.NewRequest("POST","/email/request",nil))
 if w.Code!=401||called{t.Fatalf("expected 401, called=%v",called)}
}
func TestEmailRequestValidatesAddressAndPassword(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 touched:=false
 h:=NewEmailChangeHandler(emailTestStore{issue:func(context.Context,uuid.UUID,string,string,string)(string,error){
   touched=true;return "",nil
 }},emailTestSender{})
 for _,body:=range []string{
  `{"new_email":"bad","password":"pass"}`,
  `{"new_email":"test@example.com","password":""}`,
  `{"new_email":"test@example.com","password":"pass","user_id":"other"}`,
 }{
  w:=httptest.NewRecorder()
  h.Request(w,emailRequest(body,uid))
  if w.Code!=400||touched{t.Fatalf("invalid request %q: %d",body,w.Code)}
 }
}
func TestEmailRequestUsesIdentityAndVerifiedDestination(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 sent:=false
 h:=NewEmailChangeHandler(emailTestStore{
    issue:func(_ context.Context,id uuid.UUID,p,address,hash string)(string,error){
       if id!=uid||p!="secret"||address!="new@example.com"||hash==""{t.Fatal("wrong account or address")}
       return "Test User",nil
    },
 },emailTestSender{send:func(name,to,code string)error{
    if name!="Test User"||to!="new@example.com"||len(code)!=6{t.Fatal("bad notification")}
    sent=true;return nil
 }})
 w:=httptest.NewRecorder()
 h.Request(w,emailRequest(`{"password":"secret","new_email":" NEW@EXAMPLE.COM "}`,uid))
 if w.Code!=200||!sent{t.Fatalf("status %d, sent=%v: %s",w.Code,sent,w.Body.String())}
}
func TestEmailRequestDiscardsCodeWhenMailFails(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 discarded:=false
 h:=NewEmailChangeHandler(emailTestStore{
    issue:func(_ context.Context,_ uuid.UUID,_,_,_ string)(string,error){return "Test",nil},
    discard:func(_ context.Context,user uuid.UUID,hash string)error{
       discarded=user==uid&&hash!=""
       return nil
    },
 },emailTestSender{send:func(string,string,string)error{return errors.New("sensitive SMTP info")}})
 w:=httptest.NewRecorder()
 h.Request(w,emailRequest(`{"password":"secret","new_email":"new@example.com"}`,uid))
 if w.Code!=502||!discarded||strings.Contains(w.Body.String(),"sensitive SMTP"){t.Fatalf("email failure %d: %s",w.Code,w.Body.String())}
}
func TestEmailConfirmationRevokesCookie(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 h:=NewEmailChangeHandler(emailTestStore{
    confirm:func(_ context.Context,id uuid.UUID,code string)(Profile,error){
       if id!=uid||code!="ABC123"{t.Fatalf("wrong identity/code")}
       return Profile{ID:uid,Name:"Test",Email:"new@example.com"},nil
    },
 },emailTestSender{})
 w:=httptest.NewRecorder()
 req:=emailRequest(`{"code":"ABC123"}`,uid)
 h.Confirm(w,req)
 if w.Code!=200||!strings.Contains(w.Body.String(),"new@example.com"){t.Fatalf("verification %d: %s",w.Code,w.Body.String())}
 cookies:=w.Result().Cookies()
 if len(cookies)!=1||cookies[0].Name!="session"||cookies[0].MaxAge!=-1{t.Fatalf("cookie not cleared")}
}
func TestEmailConfirmationErrors(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 for _,tc:=range []struct{err error;want int}{
   {ErrEmailCodeInvalid,400},{ErrEmailCodeLocked,429},{ErrEmailTaken,409},{errors.New("private data"),500},
 }{
    h:=NewEmailChangeHandler(emailTestStore{confirm:func(context.Context,uuid.UUID,string)(Profile,error){
        return Profile{},tc.err
    }},emailTestSender{})
    w:=httptest.NewRecorder()
    h.Confirm(w,emailRequest(`{"code":"ABC123"}`,uid))
    if w.Code!=tc.want||strings.Contains(w.Body.String(),"private data"){t.Fatalf("status %d body %s",w.Code,w.Body.String())}
 }
}
