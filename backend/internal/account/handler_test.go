package account

import (
    "bytes"
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
type fakeStore struct{
    impact func(context.Context,uuid.UUID)(Impact,error)
    delete func(context.Context,uuid.UUID,string,bool)error
}
func(f fakeStore)Impact(ctx context.Context,u uuid.UUID)(Impact,error){return f.impact(ctx,u)}
func(f fakeStore)Delete(ctx context.Context,u uuid.UUID,p string,b bool)error{return f.delete(ctx,u,p,b)}
func sessionRequest(method,body string,userID uuid.UUID)*http.Request{
    r:=httptest.NewRequest(method,"/api/v1/account?user_id=malicious",strings.NewReader(body))
    r.Header.Set("Content-Type","application/json")
    return r.WithContext(auth.WithUserID(r.Context(),userID))
}

func TestAccountRequiresSession(t *testing.T){
    called:=false
    h:=NewHandler(fakeStore{
        impact:func(context.Context,uuid.UUID)(Impact,error){called=true;return Impact{},nil},
        delete:func(context.Context,uuid.UUID,string,bool)error{called=true;return nil},
    })
    for _,method:=range []string{http.MethodGet,http.MethodDelete}{
        r:=httptest.NewRequest(method,"/api/v1/account",nil)
        w:=httptest.NewRecorder()
        h.Handle(w,r)
        if w.Code!=401||called{t.Fatalf("method=%s code=%d called=%v",method,w.Code,called)}
    }
}

func TestAccountImpactUsesSessionIdentity(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    h:=NewHandler(fakeStore{
        impact:func(_ context.Context,userID uuid.UUID)(Impact,error){
            if userID!=uid{t.Fatalf("wrong caller ID %s",userID)}
            return Impact{SoleOwned:[]WorkspaceImpact{},SharedOwned:[]WorkspaceImpact{}},nil
        },
    })
    w:=httptest.NewRecorder()
    h.Handle(w,sessionRequest(http.MethodGet,"",uid))
    if w.Code!=200{t.Fatalf("status %d: %s",w.Code,w.Body.String())}
    var response Impact
    if err:=json.Unmarshal(w.Body.Bytes(),&response);err!=nil{t.Fatal(err)}
    if response.SoleOwned==nil{t.Fatal("missing impact list")}
}

func TestAccountRejectsUnconfirmedDeletion(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    called:=false
    h:=NewHandler(fakeStore{delete:func(context.Context,uuid.UUID,string,bool)error{called=true;return nil}})
    for _,body:=range []string{
        `{"password":"sample","confirmation":"DELETE ACCOUNT"}`,
        `{"password":"","confirmation":"DELETE MY ACCOUNT"}`,
        `{"password":"sample","confirmation":"DELETE MY ACCOUNT","user_id":"someone-else"}`,
        `{"password":"sample","confirmation":"DELETE MY ACCOUNT"} {}`,
    }{
        w:=httptest.NewRecorder()
        h.Handle(w,sessionRequest(http.MethodDelete,body,uid))
        if w.Code!=400||called{t.Fatalf("body=%q code=%d called=%v",body,w.Code,called)}
    }
}

func TestAccountDeleteRechecksPasswordAndConsent(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    h:=NewHandler(fakeStore{
        delete:func(_ context.Context,got uuid.UUID,password string,allowSolo bool)error{
            if got!=uid||password!="secret"||!allowSolo{t.Fatalf("invalid server args")}
            return nil
        },
    })
    w:=httptest.NewRecorder()
    h.Handle(w,sessionRequest(http.MethodDelete,
        `{"password":"secret","confirmation":"DELETE MY ACCOUNT","delete_sole_owned_workspaces":true}`,uid))
    if w.Code!=http.StatusNoContent{t.Fatalf("status %d body %s",w.Code,w.Body.String())}
    cookies:=w.Result().Cookies()
    if len(cookies)!=1||cookies[0].Name!="session"||cookies[0].MaxAge!=-1{t.Fatalf("session not cleared %#v",cookies)}
}

func TestAccountDeletionErrorsAreSafe(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    cases:=[]struct{reason error;status int}{
        {ErrBadPassword,401},
        {ErrSharedOwner,409},
        {ErrConfirmationRequired,409},
        {errors.New("internal SQL private details"),500},
    }
    for _,tc:=range cases {
        h:=NewHandler(fakeStore{delete:func(context.Context,uuid.UUID,string,bool)error{return tc.reason}})
        w:=httptest.NewRecorder()
        h.Handle(w,sessionRequest(http.MethodDelete,
            `{"password":"test","confirmation":"DELETE MY ACCOUNT"}`,uid))
        if w.Code!=tc.status||bytes.Contains(w.Body.Bytes(),[]byte("internal SQL private details")){
            t.Fatalf("error %v code %d body %s",tc.reason,w.Code,w.Body.String())
        }
    }
}
