package task

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type fakeAssignedStore struct{
    list func(context.Context,uuid.UUID)([]AssignedTask,error)
}
func (f fakeAssignedStore) ListAssigned(ctx context.Context,userID uuid.UUID)([]AssignedTask,error){
    return f.list(ctx,userID)
}

func TestAssignedTasksRequireSession(t *testing.T){
    called:=false
    handler:=NewAssignedHandler(fakeAssignedStore{list:func(context.Context,uuid.UUID)([]AssignedTask,error){
        called=true
        return nil,nil
    }})
    rec:=httptest.NewRecorder()
    handler.List(rec,httptest.NewRequest(http.MethodGet,"/api/v1/my-tasks",nil))
    if rec.Code!=401 || called {t.Fatalf("unauthorized request: %d called=%t",rec.Code,called)}
}

func TestAssignedTasksUsesAuthenticatedIdentity(t *testing.T){
    userID:=uuid.Must(uuid.NewV4())
    taskID:=uuid.Must(uuid.NewV4())
    handler:=NewAssignedHandler(fakeAssignedStore{list:func(_ context.Context,requested uuid.UUID)([]AssignedTask,error){
        if requested!=userID {t.Fatal("query used wrong user ID")}
        return []AssignedTask{{Task:Task{ID:taskID,AssigneeID:&userID,Title:"Review notes"},ProjectName:"Research",WorkspaceName:"Class"}},nil
    }})
    request:=httptest.NewRequest(http.MethodGet,"/api/v1/my-tasks?user_id=someone-else",nil)
    request=request.WithContext(auth.WithUserID(request.Context(),userID))
    rec:=httptest.NewRecorder()
    handler.List(rec,request)
    if rec.Code!=200 {t.Fatalf("unexpected status %d",rec.Code)}
    var items []AssignedTask
    if err:=json.Unmarshal(rec.Body.Bytes(),&items);err!=nil{t.Fatal(err)}
    if len(items)!=1 || items[0].AssigneeID==nil || *items[0].AssigneeID!=userID || items[0].WorkspaceName!="Class" {
        t.Fatalf("unexpected response: %+v",items)
    }
}

func TestAssignedTasksFailureDoesNotLeakDatabaseDetails(t *testing.T){
    handler:=NewAssignedHandler(fakeAssignedStore{list:func(context.Context,uuid.UUID)([]AssignedTask,error){
        return nil,errors.New("secret database detail")
    }})
    request:=httptest.NewRequest(http.MethodGet,"/api/v1/my-tasks",nil)
    request=request.WithContext(auth.WithUserID(request.Context(),uuid.Must(uuid.NewV4())))
    rec:=httptest.NewRecorder()
    handler.List(rec,request)
    if rec.Code!=500 {t.Fatalf("unexpected status: %d",rec.Code)}
    if rec.Body.String()=="" {t.Fatal("expected error body")}
    if rec.Body.String()=="secret database detail" {t.Fatal("leaked database error")}
}

func TestAssignedTasksRejectsWrites(t *testing.T){
    handler:=NewAssignedHandler(fakeAssignedStore{list:func(context.Context,uuid.UUID)([]AssignedTask,error){return nil,nil}})
    request:=httptest.NewRequest(http.MethodPost,"/api/v1/my-tasks",nil)
    request=request.WithContext(auth.WithUserID(request.Context(),uuid.Must(uuid.NewV4())))
    rec:=httptest.NewRecorder()
    handler.List(rec,request)
    if rec.Code!=405 {t.Fatalf("unexpected status %d",rec.Code)}
}
