package project

import (
 "context"
 "errors"
 "log"
 "net/http"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)
type ParticipantsStore interface{
 List(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)([]Participant,error)
 Add(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID,string)(Participant,error)
 SetRole(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID,string)(Participant,error)
 Remove(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,uuid.UUID)error
}
type ParticipantsHandler struct{store ParticipantsStore}
func NewParticipantsHandler(store ParticipantsStore)*ParticipantsHandler{return &ParticipantsHandler{store:store}}
func participantError(w http.ResponseWriter,err error){
 switch{
 case errors.Is(err,ErrNotFound):fail(w,404,"Project not found")
 case errors.Is(err,ErrNotProjectMember):fail(w,404,"Project participant not found")
 case errors.Is(err,ErrForbidden):fail(w,403,"You do not have permission to manage project members")
 case errors.Is(err,ErrInvalidProjectMember):fail(w,400,"Choose a current workspace member")
 case errors.Is(err,ErrProjectMemberConflict):fail(w,409,"Already a project participant")
 case errors.Is(err,ErrSelfRemoval):fail(w,409,"You cannot remove yourself from a project")
 case errors.Is(err,ErrWorkspaceOwnerProtected):fail(w,409,"The workspace owner cannot be removed from the project")
 default:log.Println("Project members:",err);fail(w,500,"Could not manage project members")
 }
}
func memberRoute(w http.ResponseWriter,r *http.Request)(uuid.UUID,uuid.UUID,uuid.UUID,bool){
 actor,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in");return uuid.Nil,uuid.Nil,uuid.Nil,false}
 wid,err:=uuid.FromString(r.PathValue("workspace_id"))
 if err!=nil||wid==uuid.Nil{fail(w,400,"Invalid workspace ID");return uuid.Nil,uuid.Nil,uuid.Nil,false}
 pid,err:=uuid.FromString(r.PathValue("project_id"))
 if err!=nil||pid==uuid.Nil{fail(w,400,"Invalid project ID");return uuid.Nil,uuid.Nil,uuid.Nil,false}
 return actor,wid,pid,true
}
func (h *ParticipantsHandler) Collection(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet&&r.Method!=http.MethodPost{
  w.Header().Set("Allow","GET, POST");fail(w,405,"Method not allowed");return
 }
 actor,wid,pid,ok:=memberRoute(w,r);if !ok{return}
 if r.Method==http.MethodGet{
  items,err:=h.store.List(r.Context(),actor,wid,pid)
  if err!=nil{participantError(w,err);return}
  respond(w,200,items);return
 }
 var input struct{UserID uuid.UUID `json:"user_id"`;Role string `json:"role"`}
 if !decode(w,r,&input){return}
 if input.UserID==uuid.Nil{fail(w,400,"Choose a workspace member");return}
 if input.Role!="lead"&&input.Role!="member"{fail(w,400,"Role must be lead or member");return}
 item,err:=h.store.Add(r.Context(),actor,wid,pid,input.UserID,input.Role)
 if err!=nil{participantError(w,err);return}
 respond(w,201,item)
}
func(h *ParticipantsHandler) Detail(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPatch&&r.Method!=http.MethodDelete{
  w.Header().Set("Allow","PATCH, DELETE");fail(w,405,"Method not allowed");return
 }
 actor,wid,pid,ok:=memberRoute(w,r);if !ok{return}
 target,err:=uuid.FromString(r.PathValue("user_id"))
 if err!=nil||target==uuid.Nil{fail(w,400,"Invalid member ID");return}
 if r.Method==http.MethodDelete{
  if err:=h.store.Remove(r.Context(),actor,wid,pid,target);err!=nil{participantError(w,err);return}
  w.WriteHeader(204);return
 }
 var input struct{Role string `json:"role"`}
 if !decode(w,r,&input){return}
 if input.Role!="member"&&input.Role!="lead"{fail(w,400,"Role must be lead or member");return}
 changed,err:=h.store.SetRole(r.Context(),actor,wid,pid,target,input.Role)
 if err!=nil{participantError(w,err);return}
 respond(w,200,changed)
}
var _ ParticipantsStore=(*ParticipantsRepository)(nil)
