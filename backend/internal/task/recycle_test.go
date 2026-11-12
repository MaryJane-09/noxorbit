package task

import (
    "context"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type fakeRecycleStore struct {
    list func(context.Context,uuid.UUID)([]RecycledTask,error)
    restore func(context.Context,uuid.UUID,uuid.UUID)error
    deleteForever func(context.Context,uuid.UUID,uuid.UUID)error
}
func(f fakeRecycleStore)ListRecycled(ctx context.Context,u uuid.UUID)([]RecycledTask,error){return f.list(ctx,u)}
func(f fakeRecycleStore)Restore(ctx context.Context,u,t uuid.UUID)error{return f.restore(ctx,u,t)}
func(f fakeRecycleStore)DeleteForever(ctx context.Context,u,t uuid.UUID)error{
    if f.deleteForever == nil { panic("unexpected DeleteForever call") }
    return f.deleteForever(ctx,u,t)
}
func recycleContext(r *http.Request, uid uuid.UUID)*http.Request{return r.WithContext(auth.WithUserID(r.Context(),uid))}

func TestRecycleBinRequiresLogin(t *testing.T){
    called:=false
    h:=NewRecycleHandler(fakeRecycleStore{
        list:func(context.Context,uuid.UUID)([]RecycledTask,error){called=true;return nil,nil},
        restore:func(context.Context,uuid.UUID,uuid.UUID)error{called=true;return nil},
    })
    list:=httptest.NewRecorder()
    h.List(list,httptest.NewRequest(http.MethodGet,"/api/v1/recycle-bin/tasks",nil))
    if list.Code!=401||called{t.Fatalf("list: code=%d called=%v",list.Code,called)}
    restore:=httptest.NewRecorder()
    restoreRequest:=httptest.NewRequest(http.MethodPost,"/restore",nil)
    restoreRequest.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    h.Restore(restore,restoreRequest)
    if restore.Code!=401||called{t.Fatalf("restore: code=%d called=%v",restore.Code,called)}
}

func TestRecycleBinUsesSessionIdentity(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    tid:=uuid.Must(uuid.NewV4())
    h:=NewRecycleHandler(fakeRecycleStore{
        list:func(_ context.Context,got uuid.UUID)([]RecycledTask,error){
            if got!=uid{t.Fatalf("list ID mismatch %v",got)}
            return []RecycledTask{},nil
        },
        restore:func(_ context.Context,got,taskID uuid.UUID)error{
            if got!=uid||taskID!=tid{t.Fatalf("wrong restore IDs %s %s",got,taskID)}
            return nil
        },
    })
    list:=httptest.NewRecorder()
    h.List(list,recycleContext(httptest.NewRequest(http.MethodGet,"/api/v1/recycle-bin/tasks?user_id=other",nil),uid))
    if list.Code!=200||strings.TrimSpace(list.Body.String())!="[]"{t.Fatalf("list status %d, body %q",list.Code,list.Body.String())}
    restore:=httptest.NewRecorder()
    request:=recycleContext(httptest.NewRequest(http.MethodPost,"/restore?user_id=other",nil),uid)
    request.SetPathValue("task_id",tid.String())
    h.Restore(restore,request)
    if restore.Code!=204{t.Fatalf("restore status %d",restore.Code)}
}

func TestRecycleBinInvalidIDsAndNotFound(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    h:=NewRecycleHandler(fakeRecycleStore{
        list:func(context.Context,uuid.UUID)([]RecycledTask,error){return nil,nil},
        restore:func(context.Context,uuid.UUID,uuid.UUID)error{return ErrNotFound},
    })
    req:=recycleContext(httptest.NewRequest(http.MethodPost,"/restore",nil),uid)
    req.SetPathValue("task_id","invalid")
    invalid:=httptest.NewRecorder()
    h.Restore(invalid,req)
    if invalid.Code!=400{t.Fatalf("invalid task id: %d",invalid.Code)}
    req.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    missing:=httptest.NewRecorder()
    h.Restore(missing,req)
    if missing.Code!=404{t.Fatalf("inaccessible task: %d",missing.Code)}
}

func TestRecycleBinDoesNotExposeInternalErrors(t *testing.T){
    uid:=uuid.Must(uuid.NewV4())
    h:=NewRecycleHandler(fakeRecycleStore{
        list:func(context.Context,uuid.UUID)([]RecycledTask,error){return nil,errors.New("sensitive database details")},
        restore:func(context.Context,uuid.UUID,uuid.UUID)error{return errors.New("sensitive database details")},
    })
    list:=httptest.NewRecorder()
    h.List(list,recycleContext(httptest.NewRequest(http.MethodGet,"/bin",nil),uid))
    if list.Code!=500||strings.Contains(list.Body.String(),"sensitive"){t.Fatal("list error disclosure")}
    req:=recycleContext(httptest.NewRequest(http.MethodPost,"/restore",nil),uid)
    req.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    restore:=httptest.NewRecorder()
    h.Restore(restore,req)
    if restore.Code!=500||strings.Contains(restore.Body.String(),"sensitive"){t.Fatal("restore error disclosure")}
}

func TestRecycleDeleteForeverRequiresAuthenticatedSession(t *testing.T) {
    called := false
    h := NewRecycleHandler(fakeRecycleStore{deleteForever:func(context.Context,uuid.UUID,uuid.UUID)error{
        called = true
        return nil
    }})
    req := httptest.NewRequest(http.MethodDelete,"/api/v1/recycle-bin/tasks/ignored",nil)
    req.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    res := httptest.NewRecorder()
    h.DeleteForever(res,req)
    if res.Code != http.StatusUnauthorized || called { t.Fatalf("unexpected response %d called=%v", res.Code,called) }
}

func TestRecycleDeleteForeverScopesRequestToSession(t *testing.T) {
    actor := uuid.Must(uuid.NewV4())
    taskID := uuid.Must(uuid.NewV4())
    called := 0
    h := NewRecycleHandler(fakeRecycleStore{deleteForever:func(_ context.Context,userID,id uuid.UUID)error{
        called++
        if userID != actor || id != taskID { t.Fatal("must use authenticated identity and URL task ID") }
        return nil
    }})
    req := recycleContext(httptest.NewRequest(http.MethodDelete,"/api/v1/recycle-bin/tasks/"+taskID.String()+"?user_id=other",nil),actor)
    req.SetPathValue("task_id",taskID.String())
    res := httptest.NewRecorder()
    h.DeleteForever(res,req)
    if res.Code != http.StatusNoContent || called != 1 { t.Fatalf("unexpected status %d, called %d",res.Code,called) }
}

func TestRecycleDeleteForeverRejectsInvalidRequests(t *testing.T) {
    actor := uuid.Must(uuid.NewV4())
    h := NewRecycleHandler(fakeRecycleStore{deleteForever:func(context.Context,uuid.UUID,uuid.UUID)error{
        return ErrNotFound
    }})
    req := recycleContext(httptest.NewRequest(http.MethodPost,"/bin",nil),actor)
    req.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    res := httptest.NewRecorder()
    h.DeleteForever(res,req)
    if res.Code != http.StatusMethodNotAllowed { t.Fatalf("method: %d",res.Code) }
    req = recycleContext(httptest.NewRequest(http.MethodDelete,"/bin",nil),actor)
    req.SetPathValue("task_id","invalid")
    res = httptest.NewRecorder()
    h.DeleteForever(res,req)
    if res.Code != http.StatusBadRequest { t.Fatalf("invalid UUID: %d",res.Code) }
    req.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    res = httptest.NewRecorder()
    h.DeleteForever(res,req)
    if res.Code != http.StatusNotFound { t.Fatalf("not found: %d",res.Code) }
}

func TestRecycleDeleteForeverDoesNotLeakDatabaseError(t *testing.T) {
    actor := uuid.Must(uuid.NewV4())
    h := NewRecycleHandler(fakeRecycleStore{deleteForever:func(context.Context,uuid.UUID,uuid.UUID)error{
        return errors.New("sensitive database details")
    }})
    req := recycleContext(httptest.NewRequest(http.MethodDelete,"/bin",nil),actor)
    req.SetPathValue("task_id",uuid.Must(uuid.NewV4()).String())
    res := httptest.NewRecorder()
    h.DeleteForever(res,req)
    if res.Code != http.StatusInternalServerError || strings.Contains(res.Body.String(),"sensitive") { t.Fatal("internal error leaked") }
}
