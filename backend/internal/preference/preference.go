package preference

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
 "github.com/jackc/pgx/v5/pgxpool"
)

var Zones=map[string]bool{"Africa/Lagos":true,"UTC":true,"Europe/London":true,"Europe/Paris":true,
 "America/New_York":true,"America/Los_Angeles":true,"Asia/Dubai":true,"Asia/Kolkata":true,"Asia/Tokyo":true}
var Formats=map[string]bool{"YYYY-MM-DD":true,"DD/MM/YYYY":true,"MM/DD/YYYY":true}
type Settings struct {
 TimeZone string `json:"time_zone"`
 DateFormat string `json:"date_format"`
}
type Repository struct{pool *pgxpool.Pool}
func NewRepository(pool *pgxpool.Pool)*Repository{return &Repository{pool:pool}}
func(r *Repository) Get(ctx context.Context,user uuid.UUID)(Settings,error){
 var settings Settings
 err:=r.pool.QueryRow(ctx,`
 SELECT COALESCE((SELECT time_zone FROM public.user_preferences WHERE user_id=$1),'Africa/Lagos'),
        COALESCE((SELECT date_format FROM public.user_preferences WHERE user_id=$1),'YYYY-MM-DD')
 `,user).Scan(&settings.TimeZone,&settings.DateFormat)
 return settings,err
}
func(r *Repository) Save(ctx context.Context,user uuid.UUID,p Settings)(Settings,error){
 var updated Settings
 err:=r.pool.QueryRow(ctx,`
 INSERT INTO public.user_preferences(user_id,time_zone,date_format)
 VALUES($1,$2,$3)
 ON CONFLICT(user_id) DO UPDATE SET time_zone=EXCLUDED.time_zone,
  date_format=EXCLUDED.date_format,updated_at=NOW()
 RETURNING time_zone,date_format
 `,user,p.TimeZone,p.DateFormat).Scan(&updated.TimeZone,&updated.DateFormat)
 return updated,err
}
type Store interface{
 Get(context.Context,uuid.UUID)(Settings,error)
 Save(context.Context,uuid.UUID,Settings)(Settings,error)
}
type Handler struct{store Store}
func NewHandler(store Store)*Handler{return &Handler{store:store}}
func (h *Handler) Handle(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","application/json")
 reply:=func(status int,body any){w.WriteHeader(status);_ = json.NewEncoder(w).Encode(body)}
 if r.Method!=http.MethodGet&&r.Method!=http.MethodPut{
  w.Header().Set("Allow","GET, PUT");reply(405,map[string]string{"error":"Method not allowed"});return
 }
 user,ok:=auth.UserIDFromContext(r.Context())
 if !ok{reply(401,map[string]string{"error":"Not logged in"});return}
 if r.Method==http.MethodGet{
  p,err:=h.store.Get(r.Context(),user)
  if err!=nil{log.Println("preferences read:",err);reply(500,map[string]string{"error":"Could not load preferences"});return}
  reply(200,p);return
 }
 media,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
 if err!=nil||media!="application/json"{reply(415,map[string]string{"error":"Expected JSON"});return}
 defer r.Body.Close()
 decoder:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
 decoder.DisallowUnknownFields()
 var p Settings
 if err:=decoder.Decode(&p);err!=nil{reply(400,map[string]string{"error":"Invalid preferences"});return}
 var trailing any
 if err:=decoder.Decode(&trailing);!errors.Is(err,io.EOF){reply(400,map[string]string{"error":"Invalid preferences"});return}
 if !Zones[p.TimeZone]||!Formats[p.DateFormat]{
  reply(400,map[string]string{"error":"Unsupported time zone or date format"});return
 }
 updated,err:=h.store.Save(r.Context(),user,p)
 if err!=nil{log.Println("preferences save:",err);reply(500,map[string]string{"error":"Could not save preferences"});return}
 reply(200,updated)
}
