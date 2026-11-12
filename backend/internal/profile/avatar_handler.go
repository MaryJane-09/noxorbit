package profile

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "image"
 _ "image/jpeg"
 _ "image/png"
 "io"
 "log"
 "net/http"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)

type AvatarStore interface{
 SaveAvatar(context.Context,uuid.UUID,Avatar)error
 LoadAvatar(context.Context,uuid.UUID)(Avatar,error)
 DeleteAvatar(context.Context,uuid.UUID)error
 AvatarPresent(context.Context,uuid.UUID)(bool,error)
}

func avatarETag(data []byte) string {
 sum:=sha256.Sum256(data)
 return "\"" + hex.EncodeToString(sum[:]) + "\""
}

// Avatar is a private, authenticated resource. POST accepts only a single
// image payload, while DELETE resets the account to its initials avatar.
func(h *Handler)Avatar(w http.ResponseWriter,r *http.Request){
 uid,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in");return}
 if r.Method==http.MethodGet{
    picture,err:=h.store.LoadAvatar(r.Context(),uid)
    if errors.Is(err,ErrAvatarMissing){http.NotFound(w,r);return}
    if err!=nil{log.Println("Load avatar:",err);fail(w,500,"Could not load picture");return}
    tag:=avatarETag(picture.Data)
    w.Header().Set("Cache-Control","private, no-cache")
    w.Header().Set("ETag",tag)
    w.Header().Set("X-Content-Type-Options","nosniff")
    if r.Header.Get("If-None-Match")==tag{w.WriteHeader(304);return}
    w.Header().Set("Content-Type",picture.Mime)
    _,_=w.Write(picture.Data)
    return
 }
 if r.Method==http.MethodDelete{
    if err:=h.store.DeleteAvatar(r.Context(),uid);err!=nil{
        if errors.Is(err,ErrNotFound){fail(w,401,"Account not found");return}
        log.Println("Delete avatar:",err);fail(w,500,"Could not remove picture");return
    }
    w.WriteHeader(204);return
 }
 if r.Method!=http.MethodPost{
    w.Header().Set("Allow","GET, POST, DELETE")
    fail(w,405,"Method not allowed");return
 }
 r.Body=http.MaxBytesReader(w,r.Body,maxAvatarBytes+1)
 defer r.Body.Close()
 data,err:=io.ReadAll(r.Body)
 if err!=nil||len(data)==0||len(data)>maxAvatarBytes{
    fail(w,413,"Choose an image no larger than 512 KB");return
 }
 mime:=http.DetectContentType(data)
 if mime!="image/jpeg"&&mime!="image/png"{
    fail(w,400,"Only PNG and JPEG profile pictures are supported");return
 }
 cfg,format,err:=image.DecodeConfig(bytes.NewReader(data))
 if err!=nil||cfg.Width<1||cfg.Height<1||cfg.Width>1024||cfg.Height>1024{
    fail(w,400,"Image must be valid and at most 1024 by 1024 pixels");return
 }
 if (format=="png"&&mime!="image/png")||(format=="jpeg"&&mime!="image/jpeg"){
    fail(w,400,"Invalid image format");return
 }
 if err:=h.store.SaveAvatar(r.Context(),uid,Avatar{Data:data,Mime:mime});err!=nil{
    if errors.Is(err,ErrNotFound){fail(w,401,"Account not found");return}
    log.Println("Save avatar:",err);fail(w,500,"Could not save picture");return
 }
 w.WriteHeader(204)
}

func(h *Handler)AvatarStatus(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{w.Header().Set("Allow","GET");fail(w,405,"Method not allowed");return}
 uid,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in");return}
 exists,err:=h.store.AvatarPresent(r.Context(),uid)
 if err!=nil{log.Println("Avatar status:",err);fail(w,500,"Could not load picture status");return}
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]bool{"has_avatar":exists})
}
