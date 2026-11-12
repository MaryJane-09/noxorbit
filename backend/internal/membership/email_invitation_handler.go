package membership

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "log"
 "mime"
 "net/http"
 "net/url"
 "os"
 "strings"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)

type EmailInviteStore interface {
 CreateEmail(context.Context,uuid.UUID,uuid.UUID,string)(string,uuid.UUID,string,error)
 RevokeFailedEmail(context.Context,uuid.UUID)error
}
type WorkspaceInviteMailer interface { SendWorkspaceInvitation(to,workspaceName,link string)error }
type EmailInviteHandler struct { store EmailInviteStore; mailer WorkspaceInviteMailer; frontendURL string }
func NewEmailInviteHandler(store EmailInviteStore,mailer WorkspaceInviteMailer)*EmailInviteHandler{
 return &EmailInviteHandler{store:store,mailer:mailer,frontendURL:os.Getenv("FRONTEND_URL")}
}

func trustedInvitationBase(configured string)(string,error){
 if configured=="" {return "",errors.New("FRONTEND_URL is required")}
 parsed,err:=url.Parse(configured)
 if err!=nil || parsed.Host=="" || parsed.User!=nil || parsed.RawQuery!="" || parsed.Fragment!="" || parsed.Path!="" ||
   (parsed.Scheme!="https" && !(parsed.Scheme=="http" && (parsed.Hostname()=="localhost" || parsed.Hostname()=="127.0.0.1"))) {
  return "",errors.New("FRONTEND_URL must be an HTTPS origin or local development origin")
 }
 return strings.TrimRight(parsed.String(),"/"),nil
}

func (h *EmailInviteHandler) Send(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost {w.Header().Set("Allow","POST");inviteError(w,405,"Method not allowed");return}
 actor,ok:=auth.UserIDFromContext(r.Context())
 if !ok {inviteError(w,401,"Not logged in");return}
 media,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
 if err!=nil || media!="application/json" {inviteError(w,415,"Content-Type must be application/json");return}
 wid,err:=uuid.FromString(r.PathValue("workspace_id"))
 if err!=nil {inviteError(w,400,"Invalid workspace ID");return}
 base,err:=trustedInvitationBase(h.frontendURL)
 if err!=nil {log.Println("Email invitation frontend configuration:",err);inviteError(w,503,"Invitation emails are not configured");return}
 defer r.Body.Close()
 decoder:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
 decoder.DisallowUnknownFields()
 var input struct{Email string `json:"email"`}
 if err:=decoder.Decode(&input);err!=nil{inviteError(w,400,"Invalid invitation request");return}
 var extra any
 if err:=decoder.Decode(&extra);!errors.Is(err,io.EOF){inviteError(w,400,"Invalid invitation request");return}
 address,err:=NormalizeInviteEmail(input.Email)
 if err!=nil{inviteError(w,400,ErrInvalidEmail.Error());return}
 secret,id,workspaceName,err:=h.store.CreateEmail(r.Context(),actor,wid,address)
 switch {
 case errors.Is(err,ErrOwnerRequired):inviteError(w,403,"Only workspace owners may invite members");return
 case errors.Is(err,ErrInviteLimit):inviteError(w,429,"Invitation limit reached. Try again later.");return
 case err!=nil:log.Println("Create email invitation:",err);inviteError(w,500,"Could not create email invitation");return
 }
 link:=base+"/#invite/"+url.PathEscape(secret)
 if err=h.mailer.SendWorkspaceInvitation(address,workspaceName,link);err!=nil{
  log.Println("Send workspace invitation:",err)
  if revokeErr:=h.store.RevokeFailedEmail(r.Context(),id);revokeErr!=nil{log.Println("Revoke failed email invitation:",revokeErr)}
  inviteError(w,502,"Could not send invitation email")
  return
 }
 inviteReply(w,201,map[string]string{"message":"Invitation email sent"})
}
