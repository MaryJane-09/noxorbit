package profile

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "log"
    "mime"
    "net/http"
    "strings"
    "unicode/utf8"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/MaryJane-09/noxorbit/backend/internal/validate"
    "github.com/gofrs/uuid"
)

type Store interface{
    UpdateName(context.Context,uuid.UUID,string)(Profile,error)
    ChangePassword(context.Context,uuid.UUID,string,string)error
}
type Handler struct{store Store}
func NewHandler(s Store)*Handler{return &Handler{store:s}}
func fail(w http.ResponseWriter,code int,msg string){
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(code)
    _=json.NewEncoder(w).Encode(map[string]string{"error":msg})
}
func decode(w http.ResponseWriter,r *http.Request,target any)bool {
    media,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
    if err!=nil||media!="application/json"{fail(w,415,"Content-Type must be application/json");return false}
    defer r.Body.Close()
    d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
    d.DisallowUnknownFields()
    if err:=d.Decode(target);err!=nil{fail(w,400,"Invalid request");return false}
    var extra any
    if err:=d.Decode(&extra);!errors.Is(err,io.EOF){fail(w,400,"Invalid request");return false}
    return true
}
func(h *Handler)Name(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPatch{w.Header().Set("Allow","PATCH");fail(w,405,"Method not allowed");return}
    uid,ok:=auth.UserIDFromContext(r.Context())
    if !ok{fail(w,401,"Not logged in");return}
    var input struct{Name string `json:"name"`}
    if !decode(w,r,&input){return}
    name:=strings.Join(strings.Fields(input.Name)," ")
    if utf8.RuneCountInString(name)<1||utf8.RuneCountInString(name)>100{
        fail(w,400,"Name must be between 1 and 100 characters");return
    }
    changed,err:=h.store.UpdateName(r.Context(),uid,name)
    if err!=nil{
        if errors.Is(err,ErrNotFound){fail(w,401,"Account not found");return}
        log.Println("Update profile:",err);fail(w,500,"Could not save profile");return
    }
    w.Header().Set("Content-Type","application/json")
    _=json.NewEncoder(w).Encode(changed)
}
func(h *Handler)Password(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");fail(w,405,"Method not allowed");return}
    uid,ok:=auth.UserIDFromContext(r.Context())
    if !ok{fail(w,401,"Not logged in");return}
    var input struct {
        CurrentPassword string `json:"current_password"`
        NewPassword string `json:"new_password"`
    }
    if !decode(w,r,&input){return}
    if input.CurrentPassword==""{fail(w,400,"Enter your current password");return}
    if err:=validate.ValidatePassword(input.NewPassword);err!=nil{fail(w,400,err.Error());return}
    err:=h.store.ChangePassword(r.Context(),uid,input.CurrentPassword,input.NewPassword)
    switch{
    case errors.Is(err,ErrWrongPassword): fail(w,401,"Incorrect current password")
    case errors.Is(err,ErrSamePassword): fail(w,400,"Choose a different password")
    case errors.Is(err,ErrNotFound): fail(w,401,"Account not found")
    case err!=nil: log.Println("Change profile password:",err);fail(w,500,"Could not change password")
    default:
        http.SetCookie(w,&http.Cookie{Name:"session",Value:"",Path:"/",MaxAge:-1,HttpOnly:true,SameSite:http.SameSiteLaxMode})
        w.WriteHeader(http.StatusNoContent)
    }
}
var _ Store=(*Repository)(nil)
