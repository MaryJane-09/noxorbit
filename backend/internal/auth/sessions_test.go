package auth

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/MaryJane-09/noxorbit/backend/internal/session"
    "github.com/gofrs/uuid"
)

type fakeSessionManager struct {
 list func(context.Context,uuid.UUID,string)([]session.DeviceSession,error)
 revoke func(context.Context,uuid.UUID,uuid.UUID)(bool,error)
 others func(context.Context,uuid.UUID,string)(int64,error)
}
func(f fakeSessionManager)ListForUser(c context.Context,u uuid.UUID,h string)([]session.DeviceSession,error){return f.list(c,u,h)}
func(f fakeSessionManager)Revoke(c context.Context,u,id uuid.UUID)(bool,error){return f.revoke(c,u,id)}
func(f fakeSessionManager)RevokeOthers(c context.Context,u uuid.UUID,h string)(int64,error){return f.others(c,u,h)}
func sessionReq(method,path string,uid uuid.UUID)*http.Request{
 r:=httptest.NewRequest(method,path,nil)
 r.AddCookie(&http.Cookie{Name:"session",Value:"test-secret"})
 return r.WithContext(WithUserID(r.Context(),uid))
}
func TestSessionListRequiresAuthentication(t *testing.T){
 called:=false
 h:=NewSessionsHandler(fakeSessionManager{list:func(context.Context,uuid.UUID,string)([]session.DeviceSession,error){called=true;return nil,nil}})
 w:=httptest.NewRecorder()
 h.List(w,httptest.NewRequest("GET","/api/v1/account/sessions",nil))
 if w.Code!=401||called{t.Fatalf("status %d called=%v",w.Code,called)}
}
func TestSessionListScopedToCurrentUser(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 h:=NewSessionsHandler(fakeSessionManager{list:func(_ context.Context,u uuid.UUID,hash string)([]session.DeviceSession,error){
    if u!=uid||hash!=session.HashToken("test-secret"){t.Fatal("untrusted identity or token")}
    return []session.DeviceSession{{ID:uuid.Must(uuid.NewV4()),UserAgent:"Browser",CreatedAt:time.Now(),Current:true}},nil
 }})
 w:=httptest.NewRecorder()
 h.List(w,sessionReq("GET","/api/v1/account/sessions?user_id=other",uid))
 if w.Code!=200||!strings.Contains(w.Body.String(),"Browser")||strings.Contains(w.Body.String(),session.HashToken("test-secret")){t.Fatalf("status %d body %s",w.Code,w.Body.String())}
}
func TestSessionRevokeRejectsForeignID(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 foreign:=uuid.Must(uuid.NewV4())
 revoked:=false
 h:=NewSessionsHandler(fakeSessionManager{
    list:func(context.Context,uuid.UUID,string)([]session.DeviceSession,error){return []session.DeviceSession{},nil},
    revoke:func(context.Context,uuid.UUID,uuid.UUID)(bool,error){revoked=true;return true,nil},
 })
 w:=httptest.NewRecorder()
 r:=sessionReq("DELETE","/api/v1/account/sessions/"+foreign.String(),uid)
 r.SetPathValue("session_id",foreign.String())
 h.Revoke(w,r)
 if w.Code!=404||revoked{t.Fatalf("status %d revoked=%v",w.Code,revoked)}
}
func TestRevokeCurrentClearsCookie(t *testing.T){
 uid:=uuid.Must(uuid.NewV4()); id:=uuid.Must(uuid.NewV4())
 h:=NewSessionsHandler(fakeSessionManager{
    list:func(context.Context,uuid.UUID,string)([]session.DeviceSession,error){return []session.DeviceSession{{ID:id,Current:true}},nil},
    revoke:func(_ context.Context,u,got uuid.UUID)(bool,error){
        if u!=uid||got!=id{t.Fatal("wrong session scope")}
        return true,nil
    },
 })
 w:=httptest.NewRecorder()
 r:=sessionReq("DELETE","/sessions/"+id.String(),uid)
 r.SetPathValue("session_id",id.String())
 h.Revoke(w,r)
 if w.Code!=204{t.Fatalf("status %d",w.Code)}
 cookies:=w.Result().Cookies()
 if len(cookies)!=1||cookies[0].Name!="session"||cookies[0].MaxAge!=-1{t.Fatal("cookie not cleared")}
}
func TestRevokeOthersRequiresJSON(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 called:=false
 h:=NewSessionsHandler(fakeSessionManager{others:func(context.Context,uuid.UUID,string)(int64,error){called=true;return 0,nil}})
 w:=httptest.NewRecorder()
 h.RevokeOtherSessions(w,sessionReq("POST","/api/v1/account/sessions/others",uid))
 if w.Code!=415||called{t.Fatalf("status %d called=%v",w.Code,called)}
}
func TestSessionsDoNotLeakInternalErrors(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 h:=NewSessionsHandler(fakeSessionManager{list:func(context.Context,uuid.UUID,string)([]session.DeviceSession,error){return nil,errors.New("secret database details")}})
 w:=httptest.NewRecorder()
 h.List(w,sessionReq("GET","/api/v1/account/sessions",uid))
 if w.Code!=500||strings.Contains(w.Body.String(),"secret database details"){t.Fatalf("status %d body %s",w.Code,w.Body.String())}
}
