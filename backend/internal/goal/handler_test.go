package goal

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

type fakeStore struct {
    list func(context.Context,uuid.UUID)([]Goal,error)
    create func(context.Context,uuid.UUID,string,string,*string)(Goal,error)
    update func(context.Context,uuid.UUID,uuid.UUID,string,string,*string)(bool,error)
    delete func(context.Context,uuid.UUID,uuid.UUID)(bool,error)
    add func(context.Context,uuid.UUID,uuid.UUID,string)(Milestone,error)
    set func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,bool)(bool,error)
    remove func(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)(bool,error)
}
func(f fakeStore)List(c context.Context,u uuid.UUID)([]Goal,error){return f.list(c,u)}
func(f fakeStore)Create(c context.Context,u uuid.UUID,t,d string,date *string)(Goal,error){return f.create(c,u,t,d,date)}
func(f fakeStore)Update(c context.Context,u,id uuid.UUID,t,d string,date *string)(bool,error){return f.update(c,u,id,t,d,date)}
func(f fakeStore)Delete(c context.Context,u,id uuid.UUID)(bool,error){return f.delete(c,u,id)}
func(f fakeStore)AddMilestone(c context.Context,u,id uuid.UUID,t string)(Milestone,error){return f.add(c,u,id,t)}
func(f fakeStore)SetMilestone(c context.Context,u,id,mid uuid.UUID,done bool)(bool,error){return f.set(c,u,id,mid,done)}
func(f fakeStore)DeleteMilestone(c context.Context,u,id,mid uuid.UUID)(bool,error){return f.remove(c,u,id,mid)}

var testUser=uuid.Must(uuid.FromString("853ec9ec-b211-4be2-99d7-e76078ed663a"))
var testGoal=uuid.Must(uuid.FromString("322eefc3-251d-46d8-aa73-fde9e6ca07bb"))
var testMilestone=uuid.Must(uuid.FromString("4910a77e-887b-4bf9-a1c4-b888a2cd929c"))
func request(method,path,body string)*http.Request{
    r:=httptest.NewRequest(method,path,strings.NewReader(body))
    r.Header.Set("Content-Type","application/json")
    r.SetPathValue("goal_id",testGoal.String())
    r.SetPathValue("milestone_id",testMilestone.String())
    return r.WithContext(auth.WithUserID(r.Context(),testUser))
}
func TestGoalsRequireLogin(t *testing.T){
    h:=NewHandler(fakeStore{})
    for _,tt:=range []struct{handler http.HandlerFunc;method string}{
        {h.Collection,"GET"},{h.Collection,"POST"},{h.Detail,"PATCH"},
        {h.Milestones,"POST"},{h.Milestone,"DELETE"},
    }{
        w:=httptest.NewRecorder()
        tt.handler(w,httptest.NewRequest(tt.method,"/api/v1/goals",nil))
        if w.Code!=401{t.Fatalf("%s got %d want 401",tt.method,w.Code)}
    }
}
func TestGoalCreateUsesSignedInAccount(t *testing.T){
    called:=false
    h:=NewHandler(fakeStore{create:func(_ context.Context,user uuid.UUID,title,description string,date *string)(Goal,error){
        called=true
        if user!=testUser||title!="Finish AI foundations"||description!="Learn the math"||date==nil||*date!="2027-02-01"{
            t.Fatalf("unexpected create args %v %q %q %v",user,title,description,date)
        }
        return Goal{ID:testGoal,Title:title,Description:description,TargetDate:date,Milestones:[]Milestone{}},nil
    }})
    w:=httptest.NewRecorder()
    h.Collection(w,request("POST","/api/v1/goals",`{"title":"  Finish   AI foundations ","description":"Learn the math","target_date":"2027-02-01"}`))
    if w.Code!=201||!called{t.Fatalf("got %d: %s",w.Code,w.Body.String())}
}
func TestGoalBadInputsNeverReachStore(t *testing.T){
    h:=NewHandler(fakeStore{})
    for _,body:=range []string{
        `{"title":""}`,`{"title":"   "}`,`{"title":"OK","target_date":"2027-02-30"}`,
        `{"title":"OK","target_date":"2027-2-1"}`,
        `{"title":"OK","owner_id":"unknown"}`,
        `{"title":"OK"}{"title":"another"}`,`{"title":"`+strings.Repeat("x",121)+`"}`,
    }{
        w:=httptest.NewRecorder()
        h.Collection(w,request("POST","/api/v1/goals",body))
        if w.Code!=400{t.Fatalf("body %q returned %d",body,w.Code)}
    }
}
func TestGoalDeleteAndMilestoneOwnership(t *testing.T){
    called:=false
    h:=NewHandler(fakeStore{set:func(_ context.Context,user,id,mid uuid.UUID,done bool)(bool,error){
        called=true
        if user!=testUser||id!=testGoal||mid!=testMilestone||!done{t.Fatal("wrong authenticated identity, goal or milestone")}
        return true,nil
    }})
    w:=httptest.NewRecorder()
    h.Milestone(w,request("PATCH","/api/v1/goals/"+testGoal.String()+"/milestones/"+testMilestone.String(),`{"completed":true}`))
    if w.Code!=204||!called{t.Fatalf("got %d %s",w.Code,w.Body.String())}
    // Unknown goal (including another user's goal) must be indistinguishable.
    missing:=NewHandler(fakeStore{delete:func(context.Context,uuid.UUID,uuid.UUID)(bool,error){return false,nil}})
    w=httptest.NewRecorder()
    missing.Detail(w,request("DELETE","/api/v1/goals/"+testGoal.String(),""))
    if w.Code!=404{t.Fatalf("got %d want 404",w.Code)}
}
func TestMilestoneRequiresBooleanAndValidName(t *testing.T){
    h:=NewHandler(fakeStore{})
    for _,body:=range []string{`{}`,`{"completed":null}`,`{"completed":"yes"}`,`{"completed":true,"user_id":"fake"}`}{
        w:=httptest.NewRecorder()
        h.Milestone(w,request("PATCH","/api/v1/goals/1/milestones/2",body))
        if w.Code!=400{t.Fatalf("bad milestone update %s gave %d",body,w.Code)}
    }
    w:=httptest.NewRecorder()
    h.Milestones(w,request("POST","/api/v1/goals/1/milestones",`{"title":"   "}`))
    if w.Code!=400{t.Fatalf("blank milestone gave %d",w.Code)}
}
func TestListIsAccountScoped(t *testing.T){
    h:=NewHandler(fakeStore{list:func(_ context.Context,user uuid.UUID)([]Goal,error){
        if user!=testUser{t.Fatal("not scoped to signed-in account")}
        return []Goal{},nil
    }})
    w:=httptest.NewRecorder()
    h.Collection(w,request("GET","/api/v1/goals",""))
    if w.Code!=200||strings.TrimSpace(w.Body.String())!="[]"{t.Fatalf("%d: %s",w.Code,w.Body.String())}
}
