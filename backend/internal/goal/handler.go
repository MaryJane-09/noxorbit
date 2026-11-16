package goal

import (
    "encoding/json"
    "errors"
    "io"
    "log"
    "mime"
    "net/http"
    "strings"
    "time"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type Handler struct{ store Store }
func NewHandler(store Store)*Handler {return &Handler{store:store}}

func reply(w http.ResponseWriter,status int,body any){
    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(status)
    _=json.NewEncoder(w).Encode(body)
}
func reject(w http.ResponseWriter,status int,message string){
    reply(w,status,map[string]string{"error":message})
}
func readJSON(w http.ResponseWriter,r *http.Request,dst any)bool{
    ct,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
    if err!=nil||ct!="application/json"{reject(w,415,"Content-Type must be application/json");return false}
    defer r.Body.Close()
    dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,16<<10))
    dec.DisallowUnknownFields()
    if err:=dec.Decode(dst);err!=nil{reject(w,400,"Invalid request");return false}
    var trailing any
    if err:=dec.Decode(&trailing);!errors.Is(err,io.EOF){reject(w,400,"Invalid request");return false}
    return true
}
func goalID(w http.ResponseWriter,r *http.Request,key string)(uuid.UUID,bool){
    id,err:=uuid.FromString(r.PathValue(key))
    if err!=nil||id==uuid.Nil{reject(w,400,"Invalid goal or milestone ID");return uuid.Nil,false}
    return id,true
}
func validateTitle(raw string,max int)(string,bool){
    name:=strings.Join(strings.Fields(raw)," ")
    return name,len([]rune(name))>=1&&len([]rune(name))<=max
}
type goalInput struct{
    Title string `json:"title"`
    Description string `json:"description"`
    TargetDate *string `json:"target_date"`
}
func validateGoal(in goalInput)(goalInput,bool){
    title,ok:=validateTitle(in.Title,120)
    if !ok||len([]rune(in.Description))>1000{return in,false}
    if in.TargetDate!=nil{
        date:=*in.TargetDate
        parsed,err:=time.Parse("2006-01-02",date)
        if err!=nil||parsed.Format("2006-01-02")!=date{return in,false}
    }
    in.Title=title
    in.Description=strings.TrimSpace(in.Description)
    return in,true
}
func dbError(w http.ResponseWriter,err error)bool{
    if err==nil{return false}
    log.Println("goals:",err)
    reject(w,500,"Could not save or load goals")
    return true
}
func(h *Handler) Collection(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodGet&&r.Method!=http.MethodPost{
        w.Header().Set("Allow","GET, POST");reject(w,405,"Method not allowed");return
    }
    user,ok:=auth.UserIDFromContext(r.Context())
    if !ok{reject(w,401,"Not logged in");return}
    if r.Method==http.MethodGet{
        goals,err:=h.store.List(r.Context(),user)
        if dbError(w,err){return}
        reply(w,200,goals);return
    }
    var in goalInput
    if !readJSON(w,r,&in){return}
    in,ok=validateGoal(in)
    if !ok{reject(w,400,"Title, description or target date is invalid");return}
    created,err:=h.store.Create(r.Context(),user,in.Title,in.Description,in.TargetDate)
    if dbError(w,err){return}
    reply(w,201,created)
}
func(h *Handler) Detail(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPatch&&r.Method!=http.MethodDelete{
        w.Header().Set("Allow","PATCH, DELETE");reject(w,405,"Method not allowed");return
    }
    user,ok:=auth.UserIDFromContext(r.Context())
    if !ok{reject(w,401,"Not logged in");return}
    id,ok:=goalID(w,r,"goal_id")
    if !ok{return}
    var found bool
    var err error
    if r.Method==http.MethodDelete{
        found,err=h.store.Delete(r.Context(),user,id)
    }else{
        var in goalInput
        if !readJSON(w,r,&in){return}
        in,ok=validateGoal(in)
        if !ok{reject(w,400,"Title, description or target date is invalid");return}
        found,err=h.store.Update(r.Context(),user,id,in.Title,in.Description,in.TargetDate)
    }
    if dbError(w,err){return}
    if !found{reject(w,404,"Goal not found");return}
    w.WriteHeader(204)
}
func(h *Handler) Milestones(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPost{
        w.Header().Set("Allow","POST");reject(w,405,"Method not allowed");return
    }
    user,ok:=auth.UserIDFromContext(r.Context())
    if !ok{reject(w,401,"Not logged in");return}
    id,ok:=goalID(w,r,"goal_id")
    if !ok{return}
    var in struct{Title string `json:"title"`}
    if !readJSON(w,r,&in){return}
    title,valid:=validateTitle(in.Title,160)
    if !valid{reject(w,400,"Milestone title must be 1–160 characters");return}
    item,err:=h.store.AddMilestone(r.Context(),user,id,title)
    if errors.Is(err,ErrNotFound){reject(w,404,"Goal not found");return}
    if dbError(w,err){return}
    reply(w,201,item)
}
func(h *Handler) Milestone(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPatch&&r.Method!=http.MethodDelete{
        w.Header().Set("Allow","PATCH, DELETE");reject(w,405,"Method not allowed");return
    }
    user,ok:=auth.UserIDFromContext(r.Context())
    if !ok{reject(w,401,"Not logged in");return}
    id,ok:=goalID(w,r,"goal_id");if !ok{return}
    mid,ok:=goalID(w,r,"milestone_id");if !ok{return}
    var found bool
    var err error
    if r.Method==http.MethodPatch{
        var in struct{Completed *bool `json:"completed"`}
        if !readJSON(w,r,&in){return}
        if in.Completed==nil{reject(w,400,"Completed must be true or false");return}
        found,err=h.store.SetMilestone(r.Context(),user,id,mid,*in.Completed)
    }else{
        found,err=h.store.DeleteMilestone(r.Context(),user,id,mid)
    }
    if dbError(w,err){return}
    if !found{reject(w,404,"Milestone not found");return}
    w.WriteHeader(204)
}
var _ Store=(*Repository)(nil)

func(h *Handler) Activity(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodGet{
        w.Header().Set("Allow","GET")
        reject(w,405,"Method not allowed")
        return
    }
    user,ok:=auth.UserIDFromContext(r.Context())
    if !ok{reject(w,401,"Not logged in");return}
    events,err:=h.store.ListActivity(r.Context(),user)
    if dbError(w,err){return}
    reply(w,200,events)
}
