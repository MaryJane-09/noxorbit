package membership

import (
    "context"
    "errors"
    "log"
    "net/http"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type ManagementStore interface {
    ListInvites(context.Context,uuid.UUID,uuid.UUID)([]InviteSummary,error)
    RevokeInvite(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)error
}
type RemovalStore interface {
    RemoveMember(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)error
}
type ManagementHandler struct {
    invites ManagementStore
    members RemovalStore
}
func NewManagementHandler(invites ManagementStore,members RemovalStore)*ManagementHandler{
    return &ManagementHandler{invites:invites,members:members}
}
func (h *ManagementHandler) route(w http.ResponseWriter,r *http.Request,method string)(uuid.UUID,uuid.UUID,bool){
    if r.Method!=method {
        w.Header().Set("Allow",method)
        inviteError(w,http.StatusMethodNotAllowed,"Method not allowed")
        return uuid.Nil,uuid.Nil,false
    }
    actor,ok:=auth.UserIDFromContext(r.Context())
    if !ok {inviteError(w,http.StatusUnauthorized,"Not logged in");return uuid.Nil,uuid.Nil,false}
    wid,err:=uuid.FromString(r.PathValue("workspace_id"))
    if err!=nil {inviteError(w,http.StatusBadRequest,"Invalid workspace ID");return uuid.Nil,uuid.Nil,false}
    return actor,wid,true
}
func (h *ManagementHandler) List(w http.ResponseWriter,r *http.Request){
    actor,wid,ok:=h.route(w,r,http.MethodGet)
    if !ok{return}
    invites,err:=h.invites.ListInvites(r.Context(),actor,wid)
    switch {
    case errors.Is(err,ErrOwnerRequired): inviteError(w,403,"Only workspace owners can manage invitations")
    case err!=nil:log.Println("List invitations:",err);inviteError(w,500,"Could not load invitations")
    default:inviteReply(w,200,invites)
    }
}
func (h *ManagementHandler) Revoke(w http.ResponseWriter,r *http.Request){
    actor,wid,ok:=h.route(w,r,http.MethodDelete)
    if !ok{return}
    id,err:=uuid.FromString(r.PathValue("invitation_id"))
    if err!=nil{inviteError(w,400,"Invalid invitation ID");return}
    err=h.invites.RevokeInvite(r.Context(),actor,wid,id)
    switch{
    case errors.Is(err,ErrNotFound):inviteError(w,404,"Invitation not found or already revoked")
    case err!=nil:log.Println("Revoke invitation:",err);inviteError(w,500,"Could not revoke invitation")
    default:w.WriteHeader(http.StatusNoContent)
    }
}
func (h *ManagementHandler) Remove(w http.ResponseWriter,r *http.Request){
    actor,wid,ok:=h.route(w,r,http.MethodDelete)
    if !ok{return}
    id,err:=uuid.FromString(r.PathValue("member_id"))
    if err!=nil{inviteError(w,400,"Invalid member ID");return}
    err=h.members.RemoveMember(r.Context(),actor,wid,id)
    switch{
    case errors.Is(err,ErrOwnerRequired):inviteError(w,403,"Only workspace owners can remove members")
    case errors.Is(err,ErrCannotRemoveOwner):inviteError(w,403,"The workspace owner cannot be removed")
    case errors.Is(err,ErrNotMember):inviteError(w,404,"Member not found")
    case err!=nil:log.Println("Remove member:",err);inviteError(w,500,"Could not remove member")
    default:w.WriteHeader(http.StatusNoContent)
    }
}
