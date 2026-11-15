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

type stubAppearanceStore struct{
 get func(context.Context,uuid.UUID)(Appearance,error)
 save func(context.Context,uuid.UUID,Appearance)(Appearance,error)
}
func (s stubAppearanceStore) Get(c context.Context,id uuid.UUID)(Appearance,error){
 if s.get==nil {panic("unexpected appearance GET")}
 return s.get(c,id)
}
func (s stubAppearanceStore) Save(c context.Context,id uuid.UUID,a Appearance)(Appearance,error){
 if s.save==nil {panic("unexpected appearance PUT")}
 return s.save(c,id,a)
}
func appearanceReq(method,body string,authorized bool)*http.Request{
 r:=httptest.NewRequest(method,"/api/v1/appearance",strings.NewReader(body))
 r.Header.Set("Content-Type","application/json")
 if authorized {r=r.WithContext(auth.WithUserID(r.Context(),testID()))}
 return r
}
func TestAppearanceRequiresAuth(t *testing.T){
 h:=NewAppearanceHandler(stubAppearanceStore{})
 for _,method:=range []string{"GET","PUT"}{
  w:=httptest.NewRecorder()
  h.Handle(w,appearanceReq(method,"",false))
  if w.Code!=401 {t.Fatalf("%s returned %d",method,w.Code)}
 }
}
func TestAppearanceRejectsIncompleteOrUnexpectedInput(t *testing.T){
 h:=NewAppearanceHandler(stubAppearanceStore{})
 for _,body:=range []string{
  `{"theme":"blue","reduced_motion":false}`,
  `{"theme":"light"}`,
  `{"theme":"system","reduced_motion":null}`,
  `{"theme":"dark","reduced_motion":true,"user_id":"anything"}`,
  `{"theme":"dark","reduced_motion":false}{"theme":"light"}`,
 }{
  w:=httptest.NewRecorder()
  h.Handle(w,appearanceReq("PUT",body,true))
  if w.Code!=400 {t.Fatalf("body %q returned %d",body,w.Code)}
 }
}
func TestAppearanceSaveUsesAuthenticatedAccount(t *testing.T){
 called:=false
 h:=NewAppearanceHandler(stubAppearanceStore{
  save:func(_ context.Context,id uuid.UUID,a Appearance)(Appearance,error){
   called=true
   if id!=testID()||a.Theme!="system"||!a.ReducedMotion{t.Fatalf("invalid scope or preferences: %v %+v",id,a)}
   return a,nil
  },
 })
 w:=httptest.NewRecorder()
 h.Handle(w,appearanceReq("PUT",`{"theme":"system","reduced_motion":true}`,true))
 if !called||w.Code!=200||!strings.Contains(w.Body.String(),`"reduced_motion":true`){
  t.Fatalf("save failed: %d %s",w.Code,w.Body.String())
 }
}
func TestAppearanceLoadsOnlyCurrentUser(t *testing.T){
 h:=NewAppearanceHandler(stubAppearanceStore{
  get:func(_ context.Context,id uuid.UUID)(Appearance,error){
   if id!=testID(){t.Fatalf("wrong account: %s",id)}
   return Appearance{Theme:"light",ReducedMotion:false},nil
  },
 })
 w:=httptest.NewRecorder()
 h.Handle(w,appearanceReq("GET","",true))
 if w.Code!=200||!strings.Contains(w.Body.String(),`"theme":"light"`){t.Fatalf("GET: %d %s",w.Code,w.Body.String())}
}
