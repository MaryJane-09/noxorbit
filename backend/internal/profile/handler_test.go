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

type fakeProfileStore struct {
    update func(context.Context,uuid.UUID,string)(Profile,error)
    password func(context.Context,uuid.UUID,string,string)error
}
func(f fakeProfileStore)UpdateName(c context.Context,u uuid.UUID,n string)(Profile,error){return f.update(c,u,n)}
func(f fakeProfileStore)ChangePassword(c context.Context,u uuid.UUID,old,new string)error{return f.password(c,u,old,new)}
func profileRequest(method,path,body string,uid uuid.UUID)*http.Request{
    r:=httptest.NewRequest(method,path,strings.NewReader(body))
    r.Header.Set("Content-Type","application/json")
    return r.WithContext(auth.WithUserID(r.Context(),uid))
}
func TestProfileEndpointsRequireSession(t *testing.T){
    touched:=false
    store:=fakeProfileStore{
        update:func(context.Context,uuid.UUID,string)(Profile,error){touched=true;return Profile{},nil},
        password:func(context.Context,uuid.UUID,string,string)error{touched=true;return nil},
    }
    h:=NewHandler(store)
    for _,tc:=range []struct{handler http.HandlerFunc;method string}{
        {h.Name,http.MethodPatch},{h.Password,http.MethodPost},
    }{
        r:=httptest.NewRequest(tc.method,"/profile",nil)
        w:=httptest.NewRecorder()
        tc.handler(w,r)
        if w.Code!=401||touched{t.Fatalf("session: %d touched=%v",w.Code,touched)}
    }
}
func TestProfileNameValidationAndSessionIdentity(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    called:=0
    h:=NewHandler(fakeProfileStore{update:func(_ context.Context,got uuid.UUID,name string)(Profile,error){
        called++
        if got!=uid||name!="Mary Jane"{t.Fatalf("got %s %q",got,name)}
        return Profile{ID:uid,Name:name,Email:"test@example.com"},nil
    }})
    req:=profileRequest(http.MethodPatch,"/profile?user_id=malicious",`{"name":"  Mary   Jane  "}`,uid)
    w:=httptest.NewRecorder()
    h.Name(w,req)
    if w.Code!=200||called!=1||!strings.Contains(w.Body.String(),"Mary Jane"){t.Fatalf("status=%d body=%q",w.Code,w.Body.String())}
    for _,name:=range []string{ "",strings.Repeat("x",101) }{
        w:=httptest.NewRecorder()
        h.Name(w,profileRequest(http.MethodPatch,"/profile",`{"name":"`+name+`"}`,uid))
        if w.Code!=400||called!=1{t.Fatalf("invalid name %q status=%d",name,w.Code)}
    }
}
func TestChangePasswordValidatesBeforeDatabase(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    called:=false
    h:=NewHandler(fakeProfileStore{password:func(context.Context,uuid.UUID,string,string)error{
        called=true
        return nil
    }})
    for _,body:=range []string{
        `{"current_password":"","new_password":"example123"}`,
        `{"current_password":"old","new_password":"short"}`,
        `{"current_password":"old","new_password":"example123","user_id":"malicious"}`,
    }{
        w:=httptest.NewRecorder()
        h.Password(w,profileRequest(http.MethodPost,"/password",body,uid))
        if w.Code!=400||called{t.Fatalf("body %q code=%d called=%v",body,w.Code,called)}
    }
}
func TestPasswordChangeRevokesCurrentCookie(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    h:=NewHandler(fakeProfileStore{password:func(_ context.Context,user uuid.UUID,old,new string)error{
        if user!=uid||old!="old-secret"||new!="new-secret"{t.Fatal("unexpected identity or password")}
        return nil
    }})
    w:=httptest.NewRecorder()
    h.Password(w,profileRequest(http.MethodPost,"/password",`{"current_password":"old-secret","new_password":"new-secret"}`,uid))
    if w.Code!=204{t.Fatalf("code %d %s",w.Code,w.Body.String())}
    cookies:=w.Result().Cookies()
    if len(cookies)!=1||cookies[0].Name!="session"||cookies[0].MaxAge!=-1{t.Fatalf("expected cleared session cookie: %#v",cookies)}
}
func TestProfileErrorsDoNotLeakInternalDetails(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    for _,tc:=range []struct{err error;expected int}{
        {ErrWrongPassword,401},{ErrSamePassword,400},{errors.New("private database error"),500},
    }{
        h:=NewHandler(fakeProfileStore{password:func(context.Context,uuid.UUID,string,string)error{return tc.err}})
        w:=httptest.NewRecorder()
        h.Password(w,profileRequest(http.MethodPost,"/password",`{"current_password":"old-secret","new_password":"new-secret"}`,uid))
        if w.Code!=tc.expected||strings.Contains(w.Body.String(),"private database error"){
            t.Fatalf("error=%v status=%d body=%s",tc.err,w.Code,w.Body.String())
        }
    }
}
