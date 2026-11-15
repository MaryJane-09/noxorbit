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

type Appearance struct{
 Theme string `json:"theme"`
 ReducedMotion bool `json:"reduced_motion"`
}
type AppearanceStore interface {
 Get(context.Context,uuid.UUID)(Appearance,error)
 Save(context.Context,uuid.UUID,Appearance)(Appearance,error)
}
type AppearanceRepository struct{ pool *pgxpool.Pool }
func NewAppearanceRepository(pool *pgxpool.Pool)*AppearanceRepository{return &AppearanceRepository{pool:pool}}
func(r *AppearanceRepository)Get(ctx context.Context,user uuid.UUID)(Appearance,error){
 a:=Appearance{}
 err:=r.pool.QueryRow(ctx,`
  SELECT COALESCE((SELECT theme FROM public.appearance_preferences WHERE user_id=$1),'dark'),
         COALESCE((SELECT reduced_motion FROM public.appearance_preferences WHERE user_id=$1),false)
 `,user).Scan(&a.Theme,&a.ReducedMotion)
 return a,err
}
func(r *AppearanceRepository)Save(ctx context.Context,user uuid.UUID,a Appearance)(Appearance,error){
 out:=Appearance{}
 err:=r.pool.QueryRow(ctx,`
  INSERT INTO public.appearance_preferences(user_id,theme,reduced_motion)
  VALUES($1,$2,$3)
  ON CONFLICT(user_id) DO UPDATE SET theme=EXCLUDED.theme,
   reduced_motion=EXCLUDED.reduced_motion,updated_at=NOW()
  RETURNING theme,reduced_motion
 `,user,a.Theme,a.ReducedMotion).Scan(&out.Theme,&out.ReducedMotion)
 return out,err
}
type AppearanceHandler struct{store AppearanceStore}
func NewAppearanceHandler(store AppearanceStore)*AppearanceHandler{return &AppearanceHandler{store:store}}
func(h *AppearanceHandler)Handle(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Content-Type","application/json")
 respond:=func(code int,result any){w.WriteHeader(code);_ = json.NewEncoder(w).Encode(result)}
 invalid:=func(code int,message string){respond(code,map[string]string{"error":message})}
 if r.Method!=http.MethodGet&&r.Method!=http.MethodPut{
  w.Header().Set("Allow","GET, PUT")
  invalid(http.StatusMethodNotAllowed,"Method not allowed")
  return
 }
 user,ok:=auth.UserIDFromContext(r.Context())
 if !ok{invalid(http.StatusUnauthorized,"Not logged in");return}
 if r.Method==http.MethodGet{
  a,err:=h.store.Get(r.Context(),user)
  if err!=nil{log.Printf("appearance get: %v",err);invalid(500,"Could not load appearance preferences");return}
  respond(200,a)
  return
 }
 media,_,err:=mime.ParseMediaType(r.Header.Get("Content-Type"))
 if err!=nil||media!="application/json"{
  invalid(415,"Expected application/json");return
 }
 defer r.Body.Close()
 var payload struct{
  Theme *string `json:"theme"`
  ReducedMotion *bool `json:"reduced_motion"`
 }
 dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,4096))
 dec.DisallowUnknownFields()
 if err:=dec.Decode(&payload);err!=nil{
  invalid(400,"Invalid appearance preferences");return
 }
 var extra any
 if err:=dec.Decode(&extra);!errors.Is(err,io.EOF){
  invalid(400,"Invalid appearance preferences");return
 }
 if payload.Theme==nil||payload.ReducedMotion==nil{
  invalid(400,"Theme and reduced motion preference are required");return
 }
 theme:=*payload.Theme
 if theme!="dark"&&theme!="light"&&theme!="system"{
  invalid(400,"Theme must be dark, light, or system");return
 }
 saved,err:=h.store.Save(r.Context(),user,Appearance{Theme:theme,ReducedMotion:*payload.ReducedMotion})
 if err!=nil{log.Printf("appearance save: %v",err);invalid(500,"Could not save appearance preferences");return}
 respond(200,saved)
}
var _ AppearanceStore=(*AppearanceRepository)(nil)
