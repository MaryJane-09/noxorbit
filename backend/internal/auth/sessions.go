package auth

import (
    "context"
    "encoding/json"
    "log"
    "net/http"

    "github.com/MaryJane-09/noxorbit/backend/internal/session"
    "github.com/gofrs/uuid"
)

// A narrow interface lets handlers be tested without a live database.
type SessionManager interface {
    ListForUser(context.Context,uuid.UUID,string)([]session.DeviceSession,error)
    Revoke(context.Context,uuid.UUID,uuid.UUID)(bool,error)
    RevokeOthers(context.Context,uuid.UUID,string)(int64,error)
}
type SessionsHandler struct { store SessionManager }
func NewSessionsHandler(store SessionManager)*SessionsHandler{return &SessionsHandler{store:store}}

func sessionIdentity(r *http.Request)(uuid.UUID,string,bool){
    uid,ok:=UserIDFromContext(r.Context())
    if !ok{return uuid.Nil,"",false}
    cookie,err:=r.Cookie(cookieName)
    if err!=nil{return uuid.Nil,"",false}
    return uid,session.HashToken(cookie.Value),true
}
func (h *SessionsHandler) List(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodGet{
        w.Header().Set("Allow","GET")
        writeError(w,405,"Method not allowed");return
    }
    uid,hash,ok:=sessionIdentity(r)
    if !ok{writeError(w,401,"Not logged in");return}
    items,err:=h.store.ListForUser(r.Context(),uid,hash)
    if err!=nil{log.Println("List sessions:",err);writeError(w,500,"Could not load sessions");return}
    writeJSON(w,200,map[string]any{"sessions":items})
}

func (h *SessionsHandler) Revoke(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodDelete{
        w.Header().Set("Allow","DELETE")
        writeError(w,405,"Method not allowed");return
    }
    uid,hash,ok:=sessionIdentity(r)
    if !ok{writeError(w,401,"Not logged in");return}
    id,err:=uuid.FromString(r.PathValue("session_id"))
    if err!=nil||id==uuid.Nil{writeError(w,400,"Invalid session");return}

    // A session is looked up only for the authenticated user. The manager
    // never accepts another account's user ID from a request payload.
    items,err:=h.store.ListForUser(r.Context(),uid,hash)
    if err!=nil{log.Println("Check session:",err);writeError(w,500,"Could not revoke session");return}
    current:=false
    found:=false
    for _,item:=range items{
        if item.ID==id{found=true;current=item.Current;break}
    }
    if !found{writeError(w,404,"Session not found");return}
    deleted,err:=h.store.Revoke(r.Context(),uid,id)
    if err!=nil{log.Println("Revoke session:",err);writeError(w,500,"Could not revoke session");return}
    if !deleted{writeError(w,404,"Session not found");return}
    if current{
        http.SetCookie(w,&http.Cookie{Name:cookieName,Value:"",Path:"/",MaxAge:-1,HttpOnly:true,SameSite:http.SameSiteLaxMode})
    }
    w.WriteHeader(http.StatusNoContent)
}

func (h *SessionsHandler) RevokeOtherSessions(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPost{
        w.Header().Set("Allow","POST")
        writeError(w,405,"Method not allowed");return
    }
    uid,hash,ok:=sessionIdentity(r)
    if !ok{writeError(w,401,"Not logged in");return}
    // Reject cross-site simple-form POSTs. Caller sends application/json.
    if r.Header.Get("Content-Type")!="application/json"{
        writeError(w,415,"Content-Type must be application/json");return
    }
    count,err:=h.store.RevokeOthers(r.Context(),uid,hash)
    if err!=nil{log.Println("Revoke other sessions:",err);writeError(w,500,"Could not revoke other sessions");return}
    w.Header().Set("Content-Type","application/json")
    _=json.NewEncoder(w).Encode(map[string]int64{"revoked":count})
}
var _ SessionManager=(*session.SessionRepository)(nil)
