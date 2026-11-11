package project

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
	create func(context.Context, uuid.UUID, uuid.UUID, Input) (Project, error)
	list func(context.Context, uuid.UUID, uuid.UUID) ([]Project, error)
	get func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Project, error)
	update func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, Input) (Project, error)
	status func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) (Project, error)
}
func (f fakeStore) Create(c context.Context, uid, wid uuid.UUID, input Input) (Project,error) { if f.create==nil {panic("unexpected Create")};return f.create(c,uid,wid,input) }
func (f fakeStore) ListForUser(c context.Context, uid,wid uuid.UUID) ([]Project,error) { if f.list==nil {panic("unexpected List")};return f.list(c,uid,wid) }
func (f fakeStore) GetForUser(c context.Context,uid,wid,pid uuid.UUID) (Project,error) { if f.get==nil {panic("unexpected Get")};return f.get(c,uid,wid,pid) }
func (f fakeStore) UpdateForMember(c context.Context,uid,wid,pid uuid.UUID,input Input) (Project,error) { if f.update==nil {panic("unexpected Update")};return f.update(c,uid,wid,pid,input) }
func (f fakeStore) UpdateStatusForOwner(c context.Context,uid,wid,pid uuid.UUID,status string) (Project,error) { if f.status==nil {panic("unexpected Status")};return f.status(c,uid,wid,pid,status) }

func id(value string) uuid.UUID {
	out, err := uuid.FromString(value)
	if err!=nil { panic(err) }
	return out
}
var uid=id("14efcf1e-bb4d-43b1-a09d-ebed749d2fb7")
var wid=id("052f92b9-e53d-4499-b7f7-6d8c4b157cb1")
var pid=id("bba96e70-f1b2-478f-a4c3-6684fd7f26ee")

func request(method, url, body string, authorized bool) *http.Request {
	req := httptest.NewRequest(method,url,strings.NewReader(body))
	req.Header.Set("Content-Type","application/json")
	req.SetPathValue("workspace_id",wid.String())
	req.SetPathValue("project_id",pid.String())
	if authorized { req=req.WithContext(auth.WithUserID(req.Context(),uid)) }
	return req
}

func TestProjectInputValidation(t *testing.T) {
	if input,err:=Normalize(Input{Name:"  New Project  ", Description:"  Some notes  "});err!=nil ||input.Name!="New Project" ||input.Description!="Some notes" { t.Fatalf("normalize = %+v, %v",input,err) }
	for _, name:=range []string{"", "  ",strings.Repeat("a",101)} {
		if _,err:=Normalize(Input{Name:name}); !errors.Is(err,ErrInvalidName) {t.Fatalf("expected name error for %q, got %v",name,err)}
	}
	if _,err:=Normalize(Input{Name:"Valid",Description:strings.Repeat("d",2001)});!errors.Is(err,ErrInvalidDescription){t.Fatalf("description cap: %v",err)}
	for _, status:=range []string{"", "pending", "deleted"} {
		if _,err:=NormalizeStatus(status);!errors.Is(err,ErrInvalidStatus) {t.Fatalf("status %q: %v",status,err)}
	}
}

func TestProjectCreateUsesSessionAndWorkspaceRoute(t *testing.T) {
	calls:=0
	h:=NewHandler(fakeStore{create:func(_ context.Context, userID,workspaceID uuid.UUID,input Input)(Project,error){
		calls++
		if userID!=uid||workspaceID!=wid||input.Name!="Demo project"||input.Description!="Notes"{t.Fatalf("incorrect params: %s %s %+v",userID,workspaceID,input)}
		return Project{ID:pid,WorkspaceID:wid,Name:input.Name,Status:"active"},nil
	}})
	rec:=httptest.NewRecorder()
	h.Collection(rec,request(http.MethodPost,"/projects",`{"name":"  Demo project ","description":" Notes "}`,true))
	if rec.Code!=201||calls!=1{t.Fatalf("code=%d body=%s calls=%d",rec.Code,rec.Body.String(),calls)}
}

func TestProjectRejectsUnknownFieldsAndInvalidBodies(t *testing.T) {
	h:=NewHandler(fakeStore{})
	for _, body:=range []string{
		`{"name":"Valid","workspace_id":"spoof"}`,
		`{"name":"Valid"} {"name":"another"}`,
		`{"name": `,
		`{"name":""}`,
	} {
		rec:=httptest.NewRecorder()
		h.Collection(rec,request(http.MethodPost,"/projects",body,true))
		if rec.Code!=400 {t.Fatalf("body %s: %d %s",body,rec.Code,rec.Body.String())}
	}
}

func TestProjectRequiresSession(t *testing.T) {
	h:=NewHandler(fakeStore{})
	for _, tc:=range []struct{
		run func(http.ResponseWriter,*http.Request)
		method string
	}{
		{h.Collection,http.MethodGet},
		{h.Collection,http.MethodPost},
		{h.Detail,http.MethodGet},
		{h.Detail,http.MethodPatch},
		{h.Status,http.MethodPatch},
	} {
		rec:=httptest.NewRecorder()
		tc.run(rec,request(tc.method,"/projects",`{"name":"Valid"}`,false))
		if rec.Code!=401 {t.Fatalf("%s: %d %s",tc.method,rec.Code,rec.Body.String())}
	}
}

func TestProjectListScopeAndEmpty(t *testing.T) {
	h:=NewHandler(fakeStore{list:func(_ context.Context,userID,workspaceID uuid.UUID)([]Project,error){
		if userID!=uid||workspaceID!=wid{t.Fatal("wrong scope")}
		return []Project{},nil
	}})
	rec:=httptest.NewRecorder()
	h.Collection(rec,request(http.MethodGet,"/projects","",true))
	if rec.Code!=200 || strings.TrimSpace(rec.Body.String())!="[]" {t.Fatalf("%d %s",rec.Code,rec.Body.String())}
}

func TestProjectStatusOwnerOnly(t *testing.T) {
	for _, c:=range []struct{ err error; want int }{
		{ErrForbidden,403},
		{ErrNotFound,404},
		{ErrWorkspaceNotFound,404},
		{errors.New("internal database credentials"),500},
	} {
		h:=NewHandler(fakeStore{status:func(_ context.Context,userID,workspaceID,projectID uuid.UUID,status string)(Project,error){
			if userID!=uid||workspaceID!=wid||projectID!=pid||status!="archived" {t.Fatal("incorrect status scope")}
			return Project{},c.err
		}})
		rec:=httptest.NewRecorder()
		h.Status(rec,request(http.MethodPatch,"/status",`{"status":"ARCHIVED"}`,true))
		if rec.Code!=c.want || strings.Contains(rec.Body.String(),"internal database credentials") {
			t.Fatalf("code=%d want=%d body=%s",rec.Code,c.want,rec.Body.String())
		}
	}
}
