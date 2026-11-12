package membership

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "log"
    "net/http"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type InvitationStore interface {
    CreateLink(context.Context,uuid.UUID,uuid.UUID)(string,error)
    AcceptLink(context.Context,uuid.UUID,string)(uuid.UUID,error)
}

type InvitationHandler struct { store InvitationStore }
func NewInvitationHandler(store InvitationStore)*InvitationHandler { return &InvitationHandler{store:store} }

func inviteReply(w http.ResponseWriter,code int,payload any) {
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(code)
    _=json.NewEncoder(w).Encode(payload)
}
func inviteError(w http.ResponseWriter,code int,message string) {
    inviteReply(w,code,map[string]string{"error":message})
}

func (h *InvitationHandler) Create(w http.ResponseWriter,r *http.Request) {
    if r.Method!=http.MethodPost {
        w.Header().Set("Allow","POST")
        inviteError(w,405,"Method not allowed")
        return
    }
    actor,ok:=auth.UserIDFromContext(r.Context())
    if !ok {inviteError(w,401,"Not logged in");return}
    workspaceID,err:=uuid.FromString(r.PathValue("workspace_id"))
    if err!=nil {inviteError(w,400,"Invalid workspace ID");return}
    secret,err:=h.store.CreateLink(r.Context(),actor,workspaceID)
    switch {
    case errors.Is(err,ErrOwnerRequired): inviteError(w,403,"Only workspace owners can create invitations")
    case err!=nil: log.Println("Create invitation:",err);inviteError(w,500,"Could not create invitation")
    default:
        w.Header().Set("Cache-Control","no-store")
        inviteReply(w,201,map[string]string{"token":secret,"expires_in":"7 days","max_uses":"1"})
    }
}
func (h *InvitationHandler) Accept(w http.ResponseWriter,r *http.Request) {
    if r.Method!=http.MethodPost {
        w.Header().Set("Allow","POST")
        inviteError(w,405,"Method not allowed")
        return
    }
    actor,ok:=auth.UserIDFromContext(r.Context())
    if !ok {inviteError(w,401,"Not logged in");return}
    contentType, _, contentErr := mime.ParseMediaType(r.Header.Get("Content-Type"))
    if contentErr != nil || contentType != "application/json" {
        inviteError(w,415,"Content-Type must be application/json")
        return
    }
    defer r.Body.Close()
    var input struct { Token string `json:"token"` }
    decoder:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
    decoder.DisallowUnknownFields()
    if err:=decoder.Decode(&input);err!=nil {inviteError(w,400,"Invalid invitation request");return}
    var extra any
    if err:=decoder.Decode(&extra);!errors.Is(err,io.EOF) {inviteError(w,400,"Invalid invitation request");return}
    id,err:=h.store.AcceptLink(r.Context(),actor,input.Token)
    switch {
    case errors.Is(err,ErrInvalidInvite): inviteError(w,404,"Invitation not found")
    case errors.Is(err,ErrExpiredInvite): inviteError(w,410,"Invitation expired or already used")
    case err!=nil:log.Println("Accept invitation:",err);inviteError(w,500,"Could not accept invitation")
    default:inviteReply(w,200,map[string]string{"workspace_id":id.String()})
    }
}
