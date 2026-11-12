package activity

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type fakeStore struct{ list func(context.Context, uuid.UUID, uuid.UUID)([]Event,error) }
func(f fakeStore)List(ctx context.Context,u,w uuid.UUID)([]Event,error){return f.list(ctx,u,w)}
func userRequest(method, url string, uid uuid.UUID)*http.Request{
    r:=httptest.NewRequest(method,url,nil)
    return r.WithContext(auth.WithUserID(r.Context(),uid))
}

func TestActivityRequiresSession(t *testing.T){
    called:=false
    h:=NewHandler(fakeStore{list:func(context.Context,uuid.UUID,uuid.UUID)([]Event,error){
        called=true
        return nil,nil
    }})
    r:=httptest.NewRequest(http.MethodGet,"/activity",nil)
    r.SetPathValue("workspace_id",uuid.Must(uuid.NewV4()).String())
    w:=httptest.NewRecorder()
    h.List(w,r)
    if w.Code!=401||called{t.Fatalf("code %d called %v",w.Code,called)}
}

func TestActivityUsesSessionAndWorkspacePath(t *testing.T){
    actor:=uuid.Must(uuid.NewV4())
    workspace:=uuid.Must(uuid.NewV4())
    h:=NewHandler(fakeStore{list:func(_ context.Context,uid,wid uuid.UUID)([]Event,error){
        if uid!=actor||wid!=workspace{t.Fatalf("wrong scoped IDs %s %s",uid,wid)}
        return []Event{},nil
    }})
    r:=userRequest(http.MethodGet,"/activity?user_id=other",actor)
    r.SetPathValue("workspace_id",workspace.String())
    w:=httptest.NewRecorder()
    h.List(w,r)
    if w.Code!=200 {t.Fatalf("expected 200 got %d",w.Code)}
    var events []Event
    if err:=json.Unmarshal(w.Body.Bytes(),&events);err!=nil{t.Fatal(err)}
    if events==nil||len(events)!=0{t.Fatalf("want empty array, got %#v",events)}
}

func TestActivityRejectsMethodsAndInvalidIDs(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    h:=NewHandler(fakeStore{list:func(context.Context,uuid.UUID,uuid.UUID)([]Event,error){t.Fatal("unexpected store call");return nil,nil}})
    method:=userRequest(http.MethodPost,"/activity",uid)
    method.SetPathValue("workspace_id",uuid.Must(uuid.NewV4()).String())
    mw:=httptest.NewRecorder()
    h.List(mw,method)
    if mw.Code!=405{t.Fatalf("wrong method %d",mw.Code)}
    bad:=userRequest(http.MethodGet,"/activity",uid)
    bad.SetPathValue("workspace_id","invalid")
    bw:=httptest.NewRecorder()
    h.List(bw,bad)
    if bw.Code!=400{t.Fatalf("invalid id %d",bw.Code)}
}

func TestActivityHidesInaccessibleWorkspaceAndDatabaseErrors(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    workspace:=uuid.Must(uuid.NewV4())
    for _,tc:=range []struct{dbErr error;want int}{
        {ErrNotFound,404},
        {errors.New("private database credentials"),500},
    }{
        h:=NewHandler(fakeStore{list:func(context.Context,uuid.UUID,uuid.UUID)([]Event,error){return nil,tc.dbErr}})
        req:=userRequest(http.MethodGet,"/activity",uid)
        req.SetPathValue("workspace_id",workspace.String())
        result:=httptest.NewRecorder()
        h.List(result,req)
        if result.Code!=tc.want{t.Fatalf("want %d got %d",tc.want,result.Code)}
        if strings.Contains(result.Body.String(),"private database credentials"){t.Fatal("internal details exposed")}
    }
}
