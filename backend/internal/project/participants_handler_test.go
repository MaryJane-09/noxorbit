package project

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
type fakeParticipants struct{
 list func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)([]Participant,error)
 add func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID,string)(Participant,error)
 setRole func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID,string)(Participant,error)
 remove func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID)error
}
func(f fakeParticipants)List(c context.Context,a,b,d uuid.UUID)([]Participant,error){if f.list==nil{panic("unexpected list")};return f.list(c,a,b,d)}
func(f fakeParticipants)Add(c context.Context,a,b,d,e uuid.UUID,role string)(Participant,error){if f.add==nil{panic("unexpected add")};return f.add(c,a,b,d,e,role)}
func(f fakeParticipants)SetRole(c context.Context,a,b,d,e uuid.UUID,role string)(Participant,error){if f.setRole==nil{panic("unexpected role")};return f.setRole(c,a,b,d,e,role)}
func(f fakeParticipants)Remove(c context.Context,a,b,d,e uuid.UUID)error{if f.remove==nil{panic("unexpected remove")};return f.remove(c,a,b,d,e)}
func participantRequest(method,body string,authorized bool)*http.Request{
 r:=httptest.NewRequest(method,"/api/v1/workspaces/"+wid.String()+"/projects/"+pid.String()+"/members",strings.NewReader(body))
 r.Header.Set("Content-Type","application/json")
 r.SetPathValue("workspace_id",wid.String())
 r.SetPathValue("project_id",pid.String())
 r.SetPathValue("user_id",uid.String())
 if authorized {r=r.WithContext(auth.WithUserID(r.Context(),uid))}
 return r
}
func TestParticipantsRequireAuthentication(t *testing.T){
 h:=NewParticipantsHandler(fakeParticipants{})
 for _,method:=range []string{"GET","POST"}{
  w:=httptest.NewRecorder()
  h.Collection(w,participantRequest(method,`{"user_id":"14efcf1e-bb4d-43b1-a09d-ebed749d2fb7","role":"member"}`,false))
  if w.Code!=401{t.Fatalf("%s got %d",method,w.Code)}
 }
}
func TestParticipantsListScopedToSession(t *testing.T){
 h:=NewParticipantsHandler(fakeParticipants{list:func(_ context.Context,a,b,d uuid.UUID)([]Participant,error){
  if a!=uid||b!=wid||d!=pid{t.Fatalf("unexpected scope %v %v %v",a,b,d)}
  return []Participant{{UserID:uid,Name:"Member",Role:"lead"}},nil
 }})
 w:=httptest.NewRecorder()
 h.Collection(w,participantRequest("GET","",true))
 if w.Code!=200||!strings.Contains(w.Body.String(),"Member"){t.Fatalf("status %d body %s",w.Code,w.Body.String())}
}
func TestParticipantsRejectInvalidRole(t *testing.T){
 called:=false
 h:=NewParticipantsHandler(fakeParticipants{add:func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID,string)(Participant,error){
  called=true;return Participant{},nil
 }})
 for _,body:=range []string{
  `{"user_id":"14efcf1e-bb4d-43b1-a09d-ebed749d2fb7","role":"owner"}`,
  `{"user_id":"14efcf1e-bb4d-43b1-a09d-ebed749d2fb7","role":"member","workspace_id":"other"}`,
 }{
  w:=httptest.NewRecorder()
  h.Collection(w,participantRequest("POST",body,true))
  if w.Code!=400||called{t.Fatalf("unexpected request accepted: %s %d",body,w.Code)}
 }
}
func TestParticipantsAddAndRoleReturnStatus(t *testing.T){
 h:=NewParticipantsHandler(fakeParticipants{
  add:func(_ context.Context,a,b,d,e uuid.UUID,role string)(Participant,error){
   if a!=uid||b!=wid||d!=pid||e!=uid||role!="member"{t.Fatal("bad input")}
   return Participant{UserID:uid,Name:"Member",Role:"member"},nil
  },
  setRole:func(_ context.Context,a,b,d,e uuid.UUID,role string)(Participant,error){
   if a!=uid||b!=wid||d!=pid||e!=uid||role!="lead"{t.Fatal("bad input")}
   return Participant{UserID:uid,Name:"Member",Role:"lead"},nil
  },
 })
 w:=httptest.NewRecorder()
 h.Collection(w,participantRequest("POST",`{"user_id":"14efcf1e-bb4d-43b1-a09d-ebed749d2fb7","role":"member"}`,true))
 if w.Code!=201{t.Fatalf("add returned %d: %s",w.Code,w.Body.String())}
 w=httptest.NewRecorder()
 h.Detail(w,participantRequest("PATCH",`{"role":"lead"}`,true))
 if w.Code!=200{t.Fatalf("role returned %d",w.Code)}
}
func TestParticipantsForbiddenAndPrivateError(t *testing.T){
 for _,tc:=range []struct{err error;want int}{
  {ErrForbidden,403},{ErrInvalidProjectMember,400},{ErrProjectMemberConflict,409},{ErrNotFound,404},{errors.New("private db details"),500},
 }{
  h:=NewParticipantsHandler(fakeParticipants{list:func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)([]Participant,error){return nil,tc.err}})
  w:=httptest.NewRecorder()
  h.Collection(w,participantRequest("GET","",true))
  if w.Code!=tc.want||strings.Contains(w.Body.String(),"private db details"){t.Fatalf("status %d body %s",w.Code,w.Body.String())}
 }
}
func TestParticipantsRemoveIsScoped(t *testing.T){
 called:=false
 h:=NewParticipantsHandler(fakeParticipants{remove:func(_ context.Context,a,b,d,e uuid.UUID)error{
  called=true
  if a!=uid||b!=wid||d!=pid||e!=uid{t.Fatal("bad scope")}
  return nil
 }})
 w:=httptest.NewRecorder()
 h.Detail(w,participantRequest("DELETE","",true))
 if w.Code!=204||!called{t.Fatalf("remove %d %v",w.Code,called)}
}
