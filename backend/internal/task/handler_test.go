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

type fakeStore struct {
	create func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, DetailsInput) (Task,error)
	list func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]Task,error)
	get func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (Task,error)
	update func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, DetailsInput) (Task,error)
	status func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, string) (Task,error)
	deleteTask func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error
}
func (f fakeStore) Create(c context.Context,a,b,d uuid.UUID,i DetailsInput)(Task,error){if f.create==nil{panic("unexpected create")};return f.create(c,a,b,d,i)}
func (f fakeStore) List(c context.Context,a,b,d uuid.UUID)([]Task,error){if f.list==nil{panic("unexpected list")};return f.list(c,a,b,d)}
func (f fakeStore) Get(c context.Context,a,b,d,e uuid.UUID)(Task,error){if f.get==nil{panic("unexpected get")};return f.get(c,a,b,d,e)}
func (f fakeStore) UpdateDetails(c context.Context,a,b,d,e uuid.UUID,i DetailsInput)(Task,error){if f.update==nil{panic("unexpected edit")};return f.update(c,a,b,d,e,i)}
func (f fakeStore) UpdateStatus(c context.Context,a,b,d,e uuid.UUID,s string)(Task,error){if f.status==nil{panic("unexpected status")};return f.status(c,a,b,d,e,s)}
func (f fakeStore) Delete(c context.Context,a,b,d,e uuid.UUID)error{if f.deleteTask==nil{panic("unexpected delete")};return f.deleteTask(c,a,b,d,e)}

var testUser=uuid.Must(uuid.NewV4())
var testWorkspace=uuid.Must(uuid.NewV4())
var testProject=uuid.Must(uuid.NewV4())
var testTask=uuid.Must(uuid.NewV4())

func req(method,body string,signedIn bool)*http.Request{
	r:=httptest.NewRequest(method,"/tasks",strings.NewReader(body))
	r.Header.Set("Content-Type","application/json")
	r.SetPathValue("workspace_id",testWorkspace.String())
	r.SetPathValue("project_id",testProject.String())
	r.SetPathValue("task_id",testTask.String())
	if signedIn {r=r.WithContext(auth.WithUserID(r.Context(),testUser))}
	return r
}
func TestTaskRoutesRequireSession(t *testing.T){
	h:=NewHandler(fakeStore{})
	for _,tc:=range []struct{ method string; run func(http.ResponseWriter,*http.Request) }{
		{http.MethodGet,h.Collection},{http.MethodPost,h.Collection},
		{http.MethodGet,h.Detail},{http.MethodPatch,h.Detail},{http.MethodDelete,h.Detail},
		{http.MethodPatch,h.Status},
	}{
		rec:=httptest.NewRecorder()
		tc.run(rec,req(tc.method,"",false))
		if rec.Code!=401 {t.Fatalf("%s: expected 401 got %d",tc.method,rec.Code)}
	}
}
func TestCreateUsesSessionIdentityAndRouteScope(t *testing.T){
	h:=NewHandler(fakeStore{create:func(_ context.Context,uid,wid,pid uuid.UUID,in DetailsInput)(Task,error){
		if uid!=testUser||wid!=testWorkspace||pid!=testProject||in.Title!="Write proposal"||in.Priority!="high" {t.Fatalf("incorrect input or scope: %+v",in)}
		return Task{ID:testTask,ProjectID:pid,WorkspaceID:wid,Title:in.Title,Priority:in.Priority,Status:"todo"},nil
	}})
	rec:=httptest.NewRecorder()
	h.Collection(rec,req(http.MethodPost,`{"title":" Write proposal ","priority":"HIGH"}`,true))
	if rec.Code!=201 {t.Fatalf("expected 201 got %d: %s",rec.Code,rec.Body.String())}
}
func TestRejectUnknownAndMalformedTaskBodies(t *testing.T){
	h:=NewHandler(fakeStore{})
	for _,body:=range []string{
		`{"title":"Valid","created_by":"forged"}`,
		`{"title":"Valid","workspace_id":"forged"}`,
		`{"title":"Valid"} {"title":"second"}`,
		`{"title":""}`, `{"title":"Valid","priority":"invalid"}`,
	}{
		rec:=httptest.NewRecorder()
		h.Collection(rec,req(http.MethodPost,body,true))
		if rec.Code!=400 {t.Fatalf("body %s: %d %s",body,rec.Code,rec.Body.String())}
	}
}
func TestStatusAndDeletePermissions(t *testing.T){
	for _,tc:=range []struct{err error;code int}{
		{ErrForbidden,403},{ErrNotFound,404},{ErrArchived,409},{ErrInvalidAssignee,400},
		{errors.New("private database error"),500},
	}{
		h:=NewHandler(fakeStore{status:func(_ context.Context,uid,wid,pid,tid uuid.UUID,status string)(Task,error){
			if uid!=testUser||wid!=testWorkspace||pid!=testProject||tid!=testTask||status!="done" {t.Fatal("wrong status scope")}
			return Task{},tc.err
		}})
		rec:=httptest.NewRecorder()
		h.Status(rec,req(http.MethodPatch,`{"status":"DONE"}`,true))
		if rec.Code!=tc.code||strings.Contains(rec.Body.String(),"private database error"){
			t.Fatalf("want %d got %d %s",tc.code,rec.Code,rec.Body.String())
		}
	}
	h:=NewHandler(fakeStore{deleteTask:func(_ context.Context,uid,wid,pid,tid uuid.UUID)error{
		if uid!=testUser||wid!=testWorkspace||pid!=testProject||tid!=testTask {t.Fatal("wrong delete scope")}
		return nil
	}})
	rec:=httptest.NewRecorder()
	h.Detail(rec,req(http.MethodDelete,"",true))
	if rec.Code!=204||rec.Body.Len()!=0 {t.Fatalf("delete: %d %s",rec.Code,rec.Body.String())}
}
func TestListReturnsRealEmptyArray(t *testing.T){
	h:=NewHandler(fakeStore{list:func(_ context.Context,_,_,_ uuid.UUID)([]Task,error){return []Task{},nil}})
	rec:=httptest.NewRecorder()
	h.Collection(rec,req(http.MethodGet,"",true))
	if rec.Code!=200||strings.TrimSpace(rec.Body.String())!="[]" {t.Fatalf("%d %s",rec.Code,rec.Body.String())}
}
