package profile

import (
 "context"
 "encoding/json"
 "errors"
 "log"
 "net/http"
 "strings"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/MaryJane-09/noxorbit/backend/internal/otp"
 "github.com/MaryJane-09/noxorbit/backend/internal/validate"
 "github.com/gofrs/uuid"
 "golang.org/x/crypto/bcrypt"
)

type EmailChangeStore interface {
 IssueEmailChange(context.Context,uuid.UUID,string,string,string)(string,error)
 DiscardEmailChange(context.Context,uuid.UUID,string)error
 ConfirmEmailChange(context.Context,uuid.UUID,string)(Profile,error)
}
type EmailChangeSender interface{SendEmailChangeCode(string,string,string)error}
type EmailChangeHandler struct{store EmailChangeStore; sender EmailChangeSender}
func NewEmailChangeHandler(store EmailChangeStore,sender EmailChangeSender)*EmailChangeHandler{
 return &EmailChangeHandler{store:store,sender:sender}
}
func(h *EmailChangeHandler)Request(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");fail(w,405,"Method not allowed");return}
 uid,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in");return}
 var in struct{
    NewEmail string `json:"new_email"`
    Password string `json:"password"`
 }
 if !decode(w,r,&in){return}
 address:=strings.ToLower(strings.TrimSpace(in.NewEmail))
 if err:=validate.ValidateEmail(address);err!=nil{fail(w,400,"Enter a valid email address");return}
 if in.Password==""{fail(w,400,"Enter your current password");return}
 code,err:=otp.Generate(otp.CodeLength)
 if err!=nil{log.Println("Email change code:",err);fail(w,500,"Could not start email verification");return}
 hash,err:=bcrypt.GenerateFromPassword([]byte(code),bcrypt.DefaultCost)
 if err!=nil{log.Println("Hash email change code:",err);fail(w,500,"Could not start email verification");return}
 name,err:=h.store.IssueEmailChange(r.Context(),uid,in.Password,address,string(hash))
 switch{
 case errors.Is(err,ErrWrongPassword):fail(w,401,"Incorrect current password");return
 case errors.Is(err,ErrEmailUnchanged):fail(w,400,"This is already your email address");return
 case errors.Is(err,ErrEmailTaken):fail(w,409,"This email address is already in use");return
 case errors.Is(err,ErrEmailCooldown):w.Header().Set("Retry-After","60");fail(w,429,"Wait one minute before requesting another verification code");return
 case errors.Is(err,ErrNotFound):fail(w,401,"Account not found");return
 case err!=nil:log.Println("Request email change:",err);fail(w,500,"Could not start email verification");return
 }
 if err:=h.sender.SendEmailChangeCode(name,address,code);err!=nil{
    log.Println("Send email-change verification:",err)
    if e:=h.store.DiscardEmailChange(r.Context(),uid,string(hash));e!=nil{log.Println("Discard failed email-change request:",e)}
    fail(w,502,"Could not send verification email. Try again later.")
    return
 }
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]string{"message":"A verification code was sent to the new email address.","expires_in":"5 minutes"})
}
func(h *EmailChangeHandler)Confirm(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");fail(w,405,"Method not allowed");return}
 uid,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in");return}
 var in struct{Code string `json:"code"`}
 if !decode(w,r,&in){return}
 code:=strings.TrimSpace(in.Code)
 if len(code)!=otp.CodeLength{fail(w,400,"Invalid or expired verification code");return}
 profile,err:=h.store.ConfirmEmailChange(r.Context(),uid,code)
 switch{
 case errors.Is(err,ErrEmailCodeLocked):fail(w,429,"Too many incorrect codes. Request another after the cooldown.");return
 case errors.Is(err,ErrEmailCodeInvalid):fail(w,400,"Invalid or expired verification code");return
 case errors.Is(err,ErrEmailTaken):fail(w,409,"This email address is already in use");return
 case errors.Is(err,ErrNotFound):fail(w,401,"Account not found");return
 case err!=nil:log.Println("Confirm email change:",err);fail(w,500,"Could not change email address");return
 }
 http.SetCookie(w,&http.Cookie{Name:"session",Value:"",Path:"/",MaxAge:-1,HttpOnly:true,SameSite:http.SameSiteLaxMode})
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(profile)
}
