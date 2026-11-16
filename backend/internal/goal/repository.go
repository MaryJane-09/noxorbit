package goal

import (
    "context"
    "errors"
    "time"

    "github.com/gofrs/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("goal or milestone not found")

type Goal struct {
    ID uuid.UUID `json:"id"`
    Title string `json:"title"`
    Description string `json:"description"`
    TargetDate *string `json:"target_date"`
    CreatedAt time.Time `json:"created_at"`
    Milestones []Milestone `json:"milestones"`
}
type Milestone struct {
    ID uuid.UUID `json:"id"`
    Title string `json:"title"`
    Completed bool `json:"completed"`
}
type ActivityEvent struct {
    ID uuid.UUID `json:"id"`
    GoalID uuid.UUID `json:"goal_id"`
    GoalTitle string `json:"goal_title"`
    MilestoneTitle *string `json:"milestone_title"`
    Action string `json:"action"`
    CreatedAt time.Time `json:"created_at"`
}

type Store interface {
    List(context.Context,uuid.UUID)([]Goal,error)
    ListActivity(context.Context,uuid.UUID)([]ActivityEvent,error)
    Create(context.Context,uuid.UUID,string,string,*string)(Goal,error)
    Update(context.Context,uuid.UUID,uuid.UUID,string,string,*string)(bool,error)
    Delete(context.Context,uuid.UUID,uuid.UUID)(bool,error)
    AddMilestone(context.Context,uuid.UUID,uuid.UUID,string)(Milestone,error)
    SetMilestone(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,bool)(bool,error)
    DeleteMilestone(context.Context,uuid.UUID,uuid.UUID,uuid.UUID)(bool,error)
}
type Repository struct { pool *pgxpool.Pool }
func NewRepository(pool *pgxpool.Pool)*Repository{return &Repository{pool:pool}}

// Every query enforces the user ID from the session, even when a caller knows
// someone else's UUID. Milestones are never accessible without goal ownership.
func(r *Repository) List(ctx context.Context,user uuid.UUID)([]Goal,error){
    rows,err:=r.pool.Query(ctx,`
      SELECT id,title,description,target_date::text,created_at
      FROM public.personal_goals WHERE user_id=$1
      ORDER BY created_at DESC,id DESC`,user)
    if err!=nil{return nil,err}
    items:=make([]Goal,0)
    positions:=make(map[uuid.UUID]int)
    for rows.Next(){
        var g Goal
        g.Milestones=make([]Milestone,0)
        if err=rows.Scan(&g.ID,&g.Title,&g.Description,&g.TargetDate,&g.CreatedAt);err!=nil{break}
        positions[g.ID]=len(items)
        items=append(items,g)
    }
    if err==nil{err=rows.Err()}
    rows.Close()
    if err!=nil{return nil,err}
    if len(items)==0{return items,nil}
    milestoneRows,err:=r.pool.Query(ctx,`
      SELECT m.goal_id,m.id,m.title,m.completed
      FROM public.goal_milestones m JOIN public.personal_goals g ON g.id=m.goal_id
      WHERE g.user_id=$1
      ORDER BY m.position,m.created_at,m.id`,user)
    if err!=nil{return nil,err}
    defer milestoneRows.Close()
    for milestoneRows.Next(){
        var goalID uuid.UUID
        var m Milestone
        if err:=milestoneRows.Scan(&goalID,&m.ID,&m.Title,&m.Completed);err!=nil{return nil,err}
        if index,ok:=positions[goalID];ok{items[index].Milestones=append(items[index].Milestones,m)}
    }
    if err:=milestoneRows.Err();err!=nil{return nil,err}
    return items,nil
}
func(r *Repository) Create(ctx context.Context,user uuid.UUID,title,description string,due *string)(Goal,error){
    var g Goal
    g.Milestones=make([]Milestone,0)
    err:=r.pool.QueryRow(ctx,`
      INSERT INTO public.personal_goals(user_id,title,description,target_date)
      VALUES($1,$2,$3,$4::date)
      RETURNING id,title,description,target_date::text,created_at`,user,title,description,due).
      Scan(&g.ID,&g.Title,&g.Description,&g.TargetDate,&g.CreatedAt)
    return g,err
}
func(r *Repository) Update(ctx context.Context,user,id uuid.UUID,title,description string,due *string)(bool,error){
    tag,err:=r.pool.Exec(ctx,`
      UPDATE public.personal_goals
      SET title=$3,description=$4,target_date=$5::date,updated_at=now()
      WHERE id=$1 AND user_id=$2`,id,user,title,description,due)
    return err==nil&&tag.RowsAffected()>0,err
}
func(r *Repository) Delete(ctx context.Context,user,id uuid.UUID)(bool,error){
    tag,err:=r.pool.Exec(ctx,`DELETE FROM public.personal_goals WHERE id=$1 AND user_id=$2`,id,user)
    return err==nil&&tag.RowsAffected()>0,err
}
func(r *Repository) AddMilestone(ctx context.Context,user,goalID uuid.UUID,title string)(Milestone,error){
    var m Milestone
    err:=r.pool.QueryRow(ctx,`
      INSERT INTO public.goal_milestones(goal_id,title,position)
      SELECT g.id,$3,COALESCE((SELECT max(position)+1 FROM public.goal_milestones WHERE goal_id=g.id),0)
      FROM public.personal_goals g WHERE g.id=$1 AND g.user_id=$2
      RETURNING id,title,completed`,goalID,user,title).
      Scan(&m.ID,&m.Title,&m.Completed)
    if errors.Is(err,pgx.ErrNoRows){return Milestone{},ErrNotFound}
    return m,err
}
func(r *Repository) SetMilestone(ctx context.Context,user,goalID,milestoneID uuid.UUID,completed bool)(bool,error){
    tag,err:=r.pool.Exec(ctx,`
      UPDATE public.goal_milestones m SET completed=$4,updated_at=now()
      FROM public.personal_goals g
      WHERE m.id=$3 AND m.goal_id=g.id AND g.id=$2 AND g.user_id=$1
    `,user,goalID,milestoneID,completed)
    return err==nil&&tag.RowsAffected()>0,err
}
func(r *Repository) DeleteMilestone(ctx context.Context,user,goalID,milestoneID uuid.UUID)(bool,error){
    tag,err:=r.pool.Exec(ctx,`
      DELETE FROM public.goal_milestones m USING public.personal_goals g
      WHERE m.id=$3 AND m.goal_id=g.id AND g.id=$2 AND g.user_id=$1
    `,user,goalID,milestoneID)
    return err==nil&&tag.RowsAffected()>0,err
}

func(r *Repository) ListActivity(ctx context.Context,user uuid.UUID)([]ActivityEvent,error){
    rows,err:=r.pool.Query(ctx,`
        SELECT id,goal_id,goal_title,milestone_title,action,created_at
        FROM public.personal_goal_activity
        WHERE user_id=$1
        ORDER BY created_at DESC,id DESC
        LIMIT 30
    `,user)
    if err!=nil{return nil,err}
    defer rows.Close()
    events:=make([]ActivityEvent,0)
    for rows.Next(){
        var e ActivityEvent
        if err:=rows.Scan(&e.ID,&e.GoalID,&e.GoalTitle,&e.MilestoneTitle,&e.Action,&e.CreatedAt);err!=nil{return nil,err}
        events=append(events,e)
    }
    if err:=rows.Err();err!=nil{return nil,err}
    return events,nil
}
