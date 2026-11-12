package profile

import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "errors"
 "image"
 "image/jpeg"
 "image/color"
 "math"
 "image/png"
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
type AvatarHandler struct{ store AvatarStore }
func NewAvatarHandler(store AvatarStore)*AvatarHandler{return &AvatarHandler{store:store}}

func(h *AvatarHandler)Avatar(w http.ResponseWriter,r *http.Request){
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
 r.Body=http.MaxBytesReader(w,r.Body,maxAvatarUploadBytes+1)
 defer r.Body.Close()
 data,err:=io.ReadAll(r.Body)
 if err!=nil||len(data)==0||len(data)>maxAvatarUploadBytes{
    fail(w,413,"Choose an image no larger than 10 MB");return
 }
 mime:=http.DetectContentType(data)
 if mime!="image/jpeg"&&mime!="image/png"{
    fail(w,400,"Only PNG and JPEG profile pictures are supported");return
 }
 cfg,format,err:=image.DecodeConfig(bytes.NewReader(data))
 if err!=nil||cfg.Width<1||cfg.Height<1||cfg.Width>maxAvatarSourceDimension||cfg.Height>maxAvatarSourceDimension{
    fail(w,400,"Image must be valid and at most 6000 by 6000 pixels");return
 }
 if (format=="png"&&mime!="image/png")||(format=="jpeg"&&mime!="image/jpeg"){
    fail(w,400,"Invalid image format");return
 }
 // Fully decode to reject incomplete images, center-crop to a square,
 // resize to a compact avatar and strip uploaded metadata (including EXIF).
 pixels,_,err:=image.Decode(bytes.NewReader(data))
 if err!=nil{fail(w,400,"Invalid or incomplete image");return}
 normalized,mime,err:=prepareAvatar(pixels,mime)
 if err!=nil{fail(w,400,"Could not optimize picture. Choose a different image.");return}
 if err:=h.store.SaveAvatar(r.Context(),uid,Avatar{Data:normalized,Mime:mime});err!=nil{
    if errors.Is(err,ErrNotFound){fail(w,401,"Account not found");return}
    log.Println("Save avatar:",err);fail(w,500,"Could not save picture");return
 }
 w.WriteHeader(204)
}

func(h *AvatarHandler)AvatarStatus(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{w.Header().Set("Allow","GET");fail(w,405,"Method not allowed");return}
 uid,ok:=auth.UserIDFromContext(r.Context())
 if !ok{fail(w,401,"Not logged in");return}
 exists,err:=h.store.AvatarPresent(r.Context(),uid)
 if err!=nil{log.Println("Avatar status:",err);fail(w,500,"Could not load picture status");return}
 w.Header().Set("Content-Type","application/json")
 _=json.NewEncoder(w).Encode(map[string]bool{"has_avatar":exists})
}

// prepareAvatar never saves the original upload. A centered square is reduced
// to 512×512 pixels using bilinear sampling, then encoded without metadata.
// Transparent PNGs remain PNG. For high-entropy PNGs, smaller dimensions
// are attempted until the database's 512 KiB storage limit is met.
func prepareAvatar(source image.Image, mime string)([]byte,string,error){
 bounds:=source.Bounds()
 side:=bounds.Dx()
 if bounds.Dy()<side{side=bounds.Dy()}
 x0:=bounds.Min.X+(bounds.Dx()-side)/2
 y0:=bounds.Min.Y+(bounds.Dy()-side)/2
 dim:=avatarOutputDimension
 if side<dim{dim=side}
 if dim<1{return nil,"",errors.New("invalid dimensions")}
 opaque:=true
 if p,ok:=source.(interface{Opaque() bool});ok{opaque=p.Opaque()}
 outputMime:="image/jpeg"
 if mime=="image/png"&&!opaque{outputMime="image/png"}
 for {
    square:=image.NewNRGBA(image.Rect(0,0,dim,dim))
    // Bilinear interpolation: smooth downscaling without third-party libs.
    scale:=float64(side)/float64(dim)
    maxSrc:=float64(side-1)
    for y:=0;y<dim;y++{
      sy:=math.Min(math.Max((float64(y)+0.5)*scale-0.5,0),maxSrc)
      ya:=int(sy);yb:=ya+1;if yb>=side{yb=side-1}
      fy:=sy-float64(ya)
      for x:=0;x<dim;x++{
        sx:=math.Min(math.Max((float64(x)+0.5)*scale-0.5,0),maxSrc)
        xa:=int(sx);xb:=xa+1;if xb>=side{xb=side-1}
        fx:=sx-float64(xa)
        p00:=color.NRGBAModel.Convert(source.At(x0+xa,y0+ya)).(color.NRGBA)
        p10:=color.NRGBAModel.Convert(source.At(x0+xb,y0+ya)).(color.NRGBA)
        p01:=color.NRGBAModel.Convert(source.At(x0+xa,y0+yb)).(color.NRGBA)
        p11:=color.NRGBAModel.Convert(source.At(x0+xb,y0+yb)).(color.NRGBA)
        mix:=func(a,b,c,d uint8)uint8{
          v:=(float64(a)*(1-fx)+float64(b)*fx)*(1-fy)+(float64(c)*(1-fx)+float64(d)*fx)*fy
          return uint8(math.Round(v))
        }
        square.SetNRGBA(x,y,color.NRGBA{R:mix(p00.R,p10.R,p01.R,p11.R),
          G:mix(p00.G,p10.G,p01.G,p11.G),B:mix(p00.B,p10.B,p01.B,p11.B),
          A:mix(p00.A,p10.A,p01.A,p11.A)})
      }
    }
    var result bytes.Buffer
    var err error
    if outputMime=="image/png"{err=png.Encode(&result,square)
    }else{err=jpeg.Encode(&result,square,&jpeg.Options{Quality:82})}
    if err!=nil{return nil,"",err}
    if result.Len()<=maxAvatarBytes{return result.Bytes(),outputMime,nil}
    if dim<=1{break}
    dim/=2
 }
 return nil,"",errors.New("processed avatar exceeds storage limit")
}
