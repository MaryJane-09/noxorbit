package preference

import (
 "context"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)
type stubStore struct{
 get func(context.Context,uuid.UUID)(Settings,error)
 save func(context.Context,uuid.UUID,Settings)(Settings,error)
}
func(s stubStore)Get(ctx context.Context,id uuid.UUID)(Settings,error){return s.get(ctx,id)}
func(s stubStore)Save(ctx context.Context,id uuid.UUID,p Settings)(Settings,error){return s.save(ctx,id,p)}
func testID()uuid.UUID{
 id,err:=uuid.FromString("126a39be-7194-47b0-9071-b750d11f17c7")
 if err!=nil{panic(err)}
 return id
}
func TestPreferencesRequireSession(t *testing.T){
 h:=NewHandler(stubStore{})
 for _,method:=range []string{"GET","PUT"}{
  r:=httptest.NewRequest(method,"/api/v1/preferences",nil)
  w:=httptest.NewRecorder()
  h.Handle(w,r)
  if w.Code!=http.StatusUnauthorized{t.Fatalf("%s status %d",method,w.Code)}
 }
}
func TestPreferencesRejectUnsupportedValues(t *testing.T){
 h:=NewHandler(stubStore{save:func(context.Context,uuid.UUID,Settings)(Settings,error){
  t.Fatal("must reject invalid input before store");return Settings{},nil
 }})
 for _,body:=range []string{
  `{"time_zone":"Invalid/Region","date_format":"YYYY-MM-DD"}`,
  `{"time_zone":"UTC","date_format":"random"}`,
  `{"time_zone":"UTC"}`,
  `{"time_zone":"UTC","date_format":"YYYY-MM-DD","extra":"not allowed"}`,
 }{
  r:=httptest.NewRequest("PUT","/api/v1/preferences",strings.NewReader(body))
  r=r.WithContext(auth.WithUserID(r.Context(),testID()))
  r.Header.Set("Content-Type","application/json")
  w:=httptest.NewRecorder()
  h.Handle(w,r)
  if w.Code!=400{t.Fatalf("body %s status %d",body,w.Code)}
 }
}
func TestPreferencesSaveUsesSessionAccount(t *testing.T){
 user:=testID()
 h:=NewHandler(stubStore{save:func(_ context.Context,id uuid.UUID,p Settings)(Settings,error){
  if id!=user||p.TimeZone!="Asia/Tokyo"||p.DateFormat!="DD/MM/YYYY"{
   t.Fatalf("wrong account or payload: %v %#v",id,p)
  }
  return p,nil
 }})
 r:=httptest.NewRequest("PUT","/api/v1/preferences",strings.NewReader(`{"time_zone":"Asia/Tokyo","date_format":"DD/MM/YYYY"}`))
 r=r.WithContext(auth.WithUserID(r.Context(),user))
 r.Header.Set("Content-Type","application/json")
 w:=httptest.NewRecorder()
 h.Handle(w,r)
 if w.Code!=200||!strings.Contains(w.Body.String(),"Asia/Tokyo"){t.Fatalf("response %d %s",w.Code,w.Body.String())}
}
func TestPreferencesListUsesSessionAccount(t *testing.T){
 user:=testID()
 h:=NewHandler(stubStore{get:func(_ context.Context,id uuid.UUID)(Settings,error){
  if id!=user{t.Fatalf("wrong requester: %s",id)}
  return Settings{TimeZone:"UTC",DateFormat:"MM/DD/YYYY"},nil
 }})
 r:=httptest.NewRequest("GET","/api/v1/preferences",nil)
 r=r.WithContext(auth.WithUserID(r.Context(),user))
 w:=httptest.NewRecorder()
 h.Handle(w,r)
 if w.Code!=200||!strings.Contains(w.Body.String(),"MM/DD/YYYY"){t.Fatalf("response %d %s",w.Code,w.Body.String())}
}
