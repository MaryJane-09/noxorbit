package profile

import (
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "image"
 "image/color"
 "image/png"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 "github.com/MaryJane-09/noxorbit/backend/internal/auth"
 "github.com/gofrs/uuid"
)

type avatarTestStore struct {
 save func(context.Context,uuid.UUID,Avatar)error
 load func(context.Context,uuid.UUID)(Avatar,error)
 remove func(context.Context,uuid.UUID)error
}
func(s avatarTestStore)SaveAvatar(ctx context.Context,id uuid.UUID,a Avatar)error{return s.save(ctx,id,a)}
func(s avatarTestStore)LoadAvatar(ctx context.Context,id uuid.UUID)(Avatar,error){return s.load(ctx,id)}
func(s avatarTestStore)DeleteAvatar(ctx context.Context,id uuid.UUID)error{return s.remove(ctx,id)}
func(s avatarTestStore)AvatarPresent(context.Context,uuid.UUID)(bool,error){return false,nil}
func avatarRequest(method string,body []byte,uid uuid.UUID)*http.Request{
 r:=httptest.NewRequest(method,"/api/v1/profile/avatar",bytes.NewReader(body))
 return r.WithContext(auth.WithUserID(r.Context(),uid))
}
func pngFixture(t *testing.T)[]byte{
 t.Helper()
 img:=image.NewRGBA(image.Rect(0,0,3,3))
 img.Set(1,1,color.RGBA{R:255,A:255})
 var b bytes.Buffer
 if err:=png.Encode(&b,img);err!=nil{t.Fatal(err)}
 return b.Bytes()
}
func TestAvatarRejectsUnauthenticatedUpload(t *testing.T){
 called:=false
 h:=NewAvatarHandler(avatarTestStore{save:func(context.Context,uuid.UUID,Avatar)error{called=true;return nil}})
 w:=httptest.NewRecorder()
 h.Avatar(w,httptest.NewRequest("POST","/avatar",strings.NewReader("anything")))
 if w.Code!=401||called{t.Fatalf("status %d called %v",w.Code,called)}
}
func TestAvatarRejectsInvalidMediaAndSize(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 called:=false
 h:=NewAvatarHandler(avatarTestStore{save:func(context.Context,uuid.UUID,Avatar)error{called=true;return nil}})
 for _,data:=range [][]byte{
     []byte("<svg><script>alert(1)</script></svg>"),
     bytes.Repeat([]byte{0},maxAvatarUploadBytes+1),
     []byte("not an image"),
 }{
     w:=httptest.NewRecorder()
     h.Avatar(w,avatarRequest("POST",data,uid))
     if w.Code<400||called{t.Fatalf("invalid avatar was accepted %d",w.Code)}
 }
}
func TestAvatarSavesAndServesOnlyCurrentUser(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 data:=pngFixture(t)
 saved:=Avatar{}
 h:=NewAvatarHandler(avatarTestStore{
    save:func(_ context.Context,id uuid.UUID,a Avatar)error{
        if id!=uid{t.Fatal("wrong account")}
        saved=a
        return nil
    },
    load:func(_ context.Context,id uuid.UUID)(Avatar,error){
        if id!=uid{t.Fatal("wrong account")}
        return saved,nil
    },
 })
 w:=httptest.NewRecorder()
 h.Avatar(w,avatarRequest("POST",data,uid))
 if w.Code!=204||saved.Mime!="image/png"||len(saved.Data)==0{t.Fatalf("upload %d mime=%s size=%d",w.Code,saved.Mime,len(saved.Data))}
 w=httptest.NewRecorder()
 h.Avatar(w,avatarRequest("GET",nil,uid))
 if w.Code!=200||w.Header().Get("Content-Type")!="image/png"||!bytes.Equal(w.Body.Bytes(),saved.Data){t.Fatalf("read %d",w.Code)}
}
func TestAvatarNoRecordReturns404(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 h:=NewAvatarHandler(avatarTestStore{load:func(context.Context,uuid.UUID)(Avatar,error){return Avatar{},ErrAvatarMissing}})
 w:=httptest.NewRecorder()
 h.Avatar(w,avatarRequest("GET",nil,uid))
 if w.Code!=404{t.Fatalf("want 404 got %d",w.Code)}
}
func TestAvatarFailureDoesNotLeakInternalErrors(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 h:=NewAvatarHandler(avatarTestStore{save:func(context.Context,uuid.UUID,Avatar)error{return errors.New("private SQL details")}})
 w:=httptest.NewRecorder()
 h.Avatar(w,avatarRequest("POST",pngFixture(t),uid))
 var output map[string]string
 _=json.Unmarshal(w.Body.Bytes(),&output)
 if w.Code!=500||strings.Contains(output["error"],"private SQL"){t.Fatalf("error leaked %d %q",w.Code,w.Body.String())}
}

func TestAvatarLargeSourceIsResizedToSquare(t *testing.T) {
 uid:=uuid.Must(uuid.NewV4())
 img:=image.NewRGBA(image.Rect(0,0,1400,900))
 for y:=0;y<900;y++{
  for x:=0;x<1400;x++{
   img.SetRGBA(x,y,color.RGBA{R:uint8(x%256),G:uint8(y%256),B:120,A:255})
  }
 }
 var input bytes.Buffer
 if err:=png.Encode(&input,img);err!=nil{t.Fatal(err)}
 if input.Len()>maxAvatarUploadBytes{t.Fatal("fixture too large")}
 var stored Avatar
 h:=NewAvatarHandler(avatarTestStore{
  save:func(_ context.Context,id uuid.UUID,a Avatar)error{
   if id!=uid{t.Fatal("wrong user")}
   stored=a
   return nil
  },
 })
 w:=httptest.NewRecorder()
 h.Avatar(w,avatarRequest("POST",input.Bytes(),uid))
 if w.Code!=204{t.Fatalf("upload rejected: %d %s",w.Code,w.Body.String())}
 if len(stored.Data)>maxAvatarBytes{t.Fatalf("stored %d bytes exceeds 512 KiB",len(stored.Data))}
 config,format,err:=image.DecodeConfig(bytes.NewReader(stored.Data))
 if err!=nil{t.Fatal(err)}
 if config.Width!=512||config.Height!=512{t.Fatalf("wanted square 512x512, got %dx%d",config.Width,config.Height)}
 if format!="jpeg"||stored.Mime!="image/jpeg"{t.Fatalf("wanted optimized JPEG, got %s %s",format,stored.Mime)}
}

func TestAvatarRejectsSourceOverDimensionLimit(t *testing.T){
 uid:=uuid.Must(uuid.NewV4())
 img:=image.NewRGBA(image.Rect(0,0,maxAvatarSourceDimension+1,1))
 var input bytes.Buffer
 if err:=png.Encode(&input,img);err!=nil{t.Fatal(err)}
 called:=false
 h:=NewAvatarHandler(avatarTestStore{save:func(context.Context,uuid.UUID,Avatar)error{called=true;return nil}})
 w:=httptest.NewRecorder()
 h.Avatar(w,avatarRequest("POST",input.Bytes(),uid))
 if w.Code!=400||called{t.Fatalf("oversize dimension accepted: %d",w.Code)}
}

func TestAvatarAcceptsTransparentPNG(t *testing.T){
 img:=image.NewNRGBA(image.Rect(0,0,256,256))
 img.SetNRGBA(128,128,color.NRGBA{R:255,A:150})
 data,mime,err:=prepareAvatar(img,"image/png")
 if err!=nil{t.Fatal(err)}
 if mime!="image/png"{t.Fatalf("transparent PNG became %q",mime)}
 decoded,_,err:=image.Decode(bytes.NewReader(data))
 if err!=nil{t.Fatal(err)}
 if _,_,_,a:=decoded.At(0,0).RGBA();a!=0{t.Fatal("transparency lost")}
}
