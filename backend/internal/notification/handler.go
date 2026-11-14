package notification

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "log"
 "mime"
 "net/http"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)
type Store interface {
 Preferences(context.Context,uuid.UUID)(Preferences,error)
 SavePreferences(context.Context,uuid.UUID,Preferences)(Preferences,error)
 List(context.Context,uuid.UUID)([]Item,error)
 MarkRead(context.Context,uuid.UUID,uuid.UUID)(bool,error)
 MarkAllRead(context.Context,uuid.UUID)error
}
type Handler struct{store Store}
func NewHandler(store Store)*Handler{return &Handler{store:store}}
func send(w http.ResponseWriter,status int,payload any){
 w.Header().Set("Content-Type","application/json")
 w.WriteHeader(status)
 _=json.NewEncoder(w).Encode(payload)
}
func fail(w http.ResponseWriter,status int,message string){send(w,status,map[string]string{"error":message})}
func getUser(w http.ResponseWriter,r *http.Request)(uuid.UUID,bool){
 user,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in")}
 return user,ok
}
func readPreferences(w http.ResponseWriter,r *http.Request)(Preferences,bool){
 var p Preferences
 media,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
 if err!=nil||media!="application/json"{fail(w,415,"Content-Type must be application/json");return p,false}
 defer r.Body.Close()
 var input struct{
  TaskAssigned *bool `json:"task_assigned"`
  DueTomorrow *bool `json:"due_tomorrow"`
  DueToday *bool `json:"due_today"`
  Overdue *bool `json:"overdue"`
 }
 dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4<<10))
 dec.DisallowUnknownFields()
 if err:=dec.Decode(&input);err!=nil{fail(w,400,"Invalid notification preferences");return p,false}
 var extra any
 if err:=dec.Decode(&extra);!errors.Is(err,io.EOF){fail(w,400,"Invalid notification preferences");return p,false}
 if input.TaskAssigned==nil||input.DueTomorrow==nil||input.DueToday==nil||input.Overdue==nil{
  fail(w,400,"All notification preferences are required");return p,false
 }
 return Preferences{TaskAssigned:*input.TaskAssigned,DueTomorrow:*input.DueTomorrow,
  DueToday:*input.DueToday,Overdue:*input.Overdue},true
}
func (h *Handler) Preferences(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet&&r.Method!=http.MethodPut{
  w.Header().Set("Allow","GET, PUT");fail(w,405,"Method not allowed");return
 }
 user,ok:=getUser(w,r);if !ok{return}
 if r.Method==http.MethodGet{
  p,err:=h.store.Preferences(r.Context(),user)
  if err!=nil{log.Println("notification preferences:",err);fail(w,500,"Could not load notification preferences");return}
  send(w,200,p);return
 }
 p,ok:=readPreferences(w,r);if !ok{return}
 saved,err:=h.store.SavePreferences(r.Context(),user,p)
 if err!=nil{log.Println("save notification preferences:",err);fail(w,500,"Could not save notification preferences");return}
 send(w,200,saved)
}
func (h *Handler) List(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{w.Header().Set("Allow","GET");fail(w,405,"Method not allowed");return}
 user,ok:=getUser(w,r);if !ok{return}
 items,err:=h.store.List(r.Context(),user)
 if err!=nil{log.Println("list notifications:",err);fail(w,500,"Could not load notifications");return}
 send(w,200,items)
}
func (h *Handler) MarkRead(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPatch{w.Header().Set("Allow","PATCH");fail(w,405,"Method not allowed");return}
 user,ok:=getUser(w,r);if !ok{return}
 id,err:=uuid.FromString(r.PathValue("notification_id"))
 if err!=nil||id==uuid.Nil{fail(w,400,"Invalid notification ID");return}
 found,err:=h.store.MarkRead(r.Context(),user,id)
 if err!=nil{log.Println("mark notification read:",err);fail(w,500,"Could not update notification");return}
 if !found{fail(w,404,"Notification not found");return}
 w.WriteHeader(204)
}
func (h *Handler) MarkAllRead(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");fail(w,405,"Method not allowed");return}
 user,ok:=getUser(w,r);if !ok{return}
 if err:=h.store.MarkAllRead(r.Context(),user);err!=nil{
  log.Println("mark all notifications read:",err);fail(w,500,"Could not update notifications");return
 }
 w.WriteHeader(204)
}
var _ Store=(*Repository)(nil)
