package account

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
type Store interface{
    Impact(context.Context,uuid.UUID)(Impact,error)
    Delete(context.Context,uuid.UUID,string,bool)error
}
type Handler struct {store Store}
func NewHandler(s Store)*Handler{return &Handler{store:s}}
type deleteRequest struct{
    Password string `json:"password"`
    Confirmation string `json:"confirmation"`
    DeleteSoleOwnedWorkspaces bool `json:"delete_sole_owned_workspaces"`
}
func respond(w http.ResponseWriter,code int,payload any){
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(code)
    _=json.NewEncoder(w).Encode(payload)
}
func errResponse(w http.ResponseWriter,code int,msg string){
    respond(w,code,map[string]string{"error":msg})
}
func(h *Handler)Handle(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodGet && r.Method!=http.MethodDelete{
        w.Header().Set("Allow","GET, DELETE")
        errResponse(w,http.StatusMethodNotAllowed,"Method not allowed")
        return
    }
    userID,ok:=auth.UserIDFromContext(r.Context())
    if !ok{errResponse(w,http.StatusUnauthorized,"Not logged in");return}
    if r.Method==http.MethodGet{
        impact,err:=h.store.Impact(r.Context(),userID)
        if err!=nil{
            log.Println("Account deletion preview:",err)
            errResponse(w,http.StatusInternalServerError,"Could not check account deletion impact")
            return
        }
        respond(w,http.StatusOK,impact)
        return
    }
    mediaType,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
    if err!=nil||mediaType!="application/json"{
        errResponse(w,http.StatusUnsupportedMediaType,"Content-Type must be application/json")
        return
    }
    defer r.Body.Close()
    decoder:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
    decoder.DisallowUnknownFields()
    var input deleteRequest
    if err:=decoder.Decode(&input);err!=nil{
        errResponse(w,http.StatusBadRequest,"Invalid account deletion request")
        return
    }
    var extra any
    if err:=decoder.Decode(&extra);!errors.Is(err,io.EOF){
        errResponse(w,http.StatusBadRequest,"Invalid account deletion request")
        return
    }
    if input.Confirmation!="DELETE MY ACCOUNT"||input.Password==""{
        errResponse(w,http.StatusBadRequest,"Enter your password and type DELETE MY ACCOUNT")
        return
    }
    err=h.store.Delete(r.Context(),userID,input.Password,input.DeleteSoleOwnedWorkspaces)
    switch {
    case errors.Is(err,ErrNotFound):
        errResponse(w,http.StatusUnauthorized,"Account not found")
    case errors.Is(err,ErrBadPassword):
        errResponse(w,http.StatusUnauthorized,"Incorrect password")
    case errors.Is(err,ErrSharedOwner):
        errResponse(w,http.StatusConflict,"Transfer ownership of shared workspaces before deleting your account")
    case errors.Is(err,ErrConfirmationRequired):
        errResponse(w,http.StatusConflict,"Confirm that your sole-member workspaces and their data will be deleted")
    case err!=nil:
        log.Println("Account deletion:",err)
        errResponse(w,http.StatusInternalServerError,"Could not delete account. No changes were saved.")
    default:
        http.SetCookie(w,&http.Cookie{Name:"session",Value:"",Path:"/",MaxAge:-1,HttpOnly:true,SameSite:http.SameSiteLaxMode})
        w.WriteHeader(http.StatusNoContent)
    }
}
var _ Store=(*Repository)(nil)
