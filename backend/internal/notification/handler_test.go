package notification

import (
 "context"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)
type testStore struct{
 pref func(context.Context,uuid.UUID)(Preferences,error)
 save func(context.Context,uuid.UUID,Preferences)(Preferences,error)
 list func(context.Context,uuid.UUID)([]Item,error)
 one func(context.Context,uuid.UUID,uuid.UUID)(bool,error)
 all func(context.Context,uuid.UUID)error
}
func(f testStore)Preferences(c context.Context,u uuid.UUID)(Preferences,error){if f.pref==nil{panic("unexpected preferences")};return f.pref(c,u)}
func(f testStore)SavePreferences(c context.Context,u uuid.UUID,p Preferences)(Preferences,error){if f.save==nil{panic("unexpected save")};return f.save(c,u,p)}
func(f testStore)List(c context.Context,u uuid.UUID)([]Item,error){if f.list==nil{panic("unexpected list")};return f.list(c,u)}
func(f testStore)MarkRead(c context.Context,u,i uuid.UUID)(bool,error){if f.one==nil{panic("unexpected read")};return f.one(c,u,i)}
func(f testStore)MarkAllRead(c context.Context,u uuid.UUID)error{if f.all==nil{panic("unexpected all")};return f.all(c,u)}
func testUUID(value string)uuid.UUID{
 id,err:=uuid.FromString(value)
 if err!=nil{panic(err)}
 return id
}
var userID=testUUID("f49ab8ea-cae0-4e1b-a1be-737316b96917")
var notificationID=testUUID("a699b087-0b92-4c0d-becd-b9799cd2460d")
func TestPreferencesRequireLogin(t *testing.T){
 h:=NewHandler(testStore{})
 r:=httptest.NewRequest("GET","/api/v1/notifications/preferences",nil)
 w:=httptest.NewRecorder()
 h.Preferences(w,r)
 if w.Code!=401{t.Fatalf("got %d",w.Code)}
}
func TestPreferencesRejectMissingFields(t *testing.T){
 h:=NewHandler(testStore{})
 for _,body:=range []string{`{"task_assigned":true}`,`{"task_assigned":true,"due_tomorrow":true,"due_today":true,"overdue":false,"extra":1}`}{
  r:=httptest.NewRequest("PUT","/api/v1/notifications/preferences",strings.NewReader(body))
  r.Header.Set("Content-Type","application/json")
  r=r.WithContext(auth.WithUserID(r.Context(),userID))
  w:=httptest.NewRecorder()
  h.Preferences(w,r)
  if w.Code!=400{t.Fatalf("got %d for %s",w.Code,body)}
 }
}
func TestPreferencesPutPreservesFalseValues(t *testing.T){
 called:=false
 h:=NewHandler(testStore{save:func(_ context.Context,u uuid.UUID,p Preferences)(Preferences,error){
  called=true
  if u!=userID||p.TaskAssigned||!p.DueTomorrow||p.DueToday||!p.Overdue{
   t.Fatalf("invalid preferences %#v",p)
  }
  return p,nil
 }})
 r:=httptest.NewRequest("PUT","/api/v1/notifications/preferences",strings.NewReader(`{"task_assigned":false,"due_tomorrow":true,"due_today":false,"overdue":true}`))
 r.Header.Set("Content-Type","application/json")
 r=r.WithContext(auth.WithUserID(r.Context(),userID))
 w:=httptest.NewRecorder()
 h.Preferences(w,r)
 if w.Code!=200||!called{t.Fatalf("status %d called %v",w.Code,called)}
}
func TestNotificationsAreUserScoped(t *testing.T){
 h:=NewHandler(testStore{list:func(_ context.Context,u uuid.UUID)([]Item,error){
  if u!=userID{t.Fatal("request must use session identity")}
  return []Item{},nil
 }})
 r:=httptest.NewRequest("GET","/api/v1/notifications",nil)
 r=r.WithContext(auth.WithUserID(r.Context(),userID))
 w:=httptest.NewRecorder()
 h.List(w,r)
 if w.Code!=200||strings.TrimSpace(w.Body.String())!="[]"{t.Fatalf("got %d %s",w.Code,w.Body.String())}
}
func TestReadUsesSessionAndPathID(t *testing.T){
 h:=NewHandler(testStore{one:func(_ context.Context,u,id uuid.UUID)(bool,error){
  if u!=userID||id!=notificationID{t.Fatal("wrong scope")}
  return true,nil
 }})
 r:=httptest.NewRequest("PATCH","/api/v1/notifications/anything/read",nil)
 r.SetPathValue("notification_id",notificationID.String())
 r=r.WithContext(auth.WithUserID(r.Context(),userID))
 w:=httptest.NewRecorder()
 h.MarkRead(w,r)
 if w.Code!=204{t.Fatalf("got %d",w.Code)}
}
func TestReadMissingIs404(t *testing.T){
 h:=NewHandler(testStore{one:func(context.Context,uuid.UUID,uuid.UUID)(bool,error){return false,nil}})
 r:=httptest.NewRequest("PATCH","/",nil)
 r.SetPathValue("notification_id",notificationID.String())
 r=r.WithContext(auth.WithUserID(r.Context(),userID))
 w:=httptest.NewRecorder()
 h.MarkRead(w,r)
 if w.Code!=404{t.Fatalf("got %d",w.Code)}
}
