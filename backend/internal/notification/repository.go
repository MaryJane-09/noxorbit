package notification

import (
 "context"
 "time"

 "github.com/gofrs/uuid"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
)

type Preferences struct {
 TaskAssigned bool `json:"task_assigned"`
 DueTomorrow bool `json:"due_tomorrow"`
 DueToday bool `json:"due_today"`
 Overdue bool `json:"overdue"`
}
type Item struct {
 ID uuid.UUID `json:"id"`
 Kind string `json:"kind"`
 Title string `json:"title"`
 Body string `json:"body"`
 WorkspaceID *uuid.UUID `json:"workspace_id"`
 ProjectID *uuid.UUID `json:"project_id"`
 TaskID *uuid.UUID `json:"task_id"`
 CreatedAt time.Time `json:"created_at"`
 ReadAt *time.Time `json:"read_at"`
}
type Repository struct{ pool *pgxpool.Pool }
func NewRepository(pool *pgxpool.Pool)*Repository{return &Repository{pool:pool}}
func (r *Repository) Preferences(ctx context.Context,user uuid.UUID)(Preferences,error){
 var p Preferences
 err:=r.pool.QueryRow(ctx,`
 SELECT COALESCE((SELECT task_assigned FROM public.notification_preferences WHERE user_id=$1),true),
 COALESCE((SELECT due_tomorrow FROM public.notification_preferences WHERE user_id=$1),true),
 COALESCE((SELECT due_today FROM public.notification_preferences WHERE user_id=$1),true),
 COALESCE((SELECT overdue FROM public.notification_preferences WHERE user_id=$1),true)
 `,user).Scan(&p.TaskAssigned,&p.DueTomorrow,&p.DueToday,&p.Overdue)
 return p,err
}
func (r *Repository) SavePreferences(ctx context.Context,user uuid.UUID,p Preferences)(Preferences,error){
 var updated Preferences
 err:=r.pool.QueryRow(ctx,`
 INSERT INTO public.notification_preferences(user_id,task_assigned,due_tomorrow,due_today,overdue)
 VALUES ($1,$2,$3,$4,$5)
 ON CONFLICT(user_id) DO UPDATE SET
 task_assigned=EXCLUDED.task_assigned,due_tomorrow=EXCLUDED.due_tomorrow,
 due_today=EXCLUDED.due_today,overdue=EXCLUDED.overdue,updated_at=NOW()
 RETURNING task_assigned,due_tomorrow,due_today,overdue
 `,user,p.TaskAssigned,p.DueTomorrow,p.DueToday,p.Overdue).
 Scan(&updated.TaskAssigned,&updated.DueTomorrow,&updated.DueToday,&updated.Overdue)
 return updated,err
}
// Reminders are generated on authenticated notification reads, not via a
// fictitious background scheduler. A daily per-task key prevents duplicates.
// Africa/Lagos is temporary until user-specific time zones are implemented.
func (r *Repository) generateDueReminders(ctx context.Context,user uuid.UUID)error{
 _,err:=r.pool.Exec(ctx,`
 WITH preference AS (
  SELECT COALESCE((SELECT due_tomorrow FROM public.notification_preferences WHERE user_id=$1),true) AS tomorrow,
  COALESCE((SELECT due_today FROM public.notification_preferences WHERE user_id=$1),true) AS today,
  COALESCE((SELECT overdue FROM public.notification_preferences WHERE user_id=$1),true) AS overdue
 ), today AS (
  SELECT (NOW() AT TIME ZONE 'Africa/Lagos')::date AS current_day
 )
 INSERT INTO public.notifications(user_id,workspace_id,project_id,task_id,kind,title,body,dedupe_key)
 SELECT $1,t.workspace_id,t.project_id,t.id,
 CASE WHEN t.due_date=today.current_day+1 THEN 'due_tomorrow'
      WHEN t.due_date=today.current_day THEN 'due_today'
      ELSE 'overdue' END,
 CASE WHEN t.due_date=today.current_day+1 THEN 'Task due tomorrow'
      WHEN t.due_date=today.current_day THEN 'Task due today'
      ELSE 'Overdue task' END,
 LEFT(t.title,200),
 'due:'||t.id::text||':'||today.current_day::text||':'||
 CASE WHEN t.due_date=today.current_day+1 THEN 'due_tomorrow'
      WHEN t.due_date=today.current_day THEN 'due_today'
      ELSE 'overdue' END
 FROM public.tasks t
 JOIN public.workspace_memberships m
  ON m.workspace_id=t.workspace_id AND m.user_id=$1
 JOIN public.projects pr ON pr.id=t.project_id AND pr.status='active'
 CROSS JOIN preference CROSS JOIN today
 WHERE t.assignee_id=$1 AND t.status<>'done' AND t.deleted_at IS NULL
   AND t.due_date IS NOT NULL
   AND ( (t.due_date=today.current_day+1 AND preference.tomorrow)
      OR (t.due_date=today.current_day AND preference.today)
      OR (t.due_date<today.current_day AND preference.overdue) )
 ON CONFLICT(user_id,dedupe_key) DO NOTHING
 `,user)
 return err
}
func (r *Repository) List(ctx context.Context,user uuid.UUID)([]Item,error){
 if err:=r.generateDueReminders(ctx,user);err!=nil{return nil,err}
 rows,err:=r.pool.Query(ctx,`
 SELECT id,kind,title,body,workspace_id,project_id,task_id,created_at,read_at
 FROM public.notifications WHERE user_id=$1
 ORDER BY created_at DESC,id DESC LIMIT 100
 `,user)
 if err!=nil{return nil,err}
 defer rows.Close()
 items:=make([]Item,0)
 for rows.Next(){
  var it Item
  if err:=rows.Scan(&it.ID,&it.Kind,&it.Title,&it.Body,&it.WorkspaceID,
   &it.ProjectID,&it.TaskID,&it.CreatedAt,&it.ReadAt);err!=nil{return nil,err}
  items=append(items,it)
 }
 return items,rows.Err()
}
func (r *Repository) MarkRead(ctx context.Context,user,id uuid.UUID)(bool,error){
 tag,err:=r.pool.Exec(ctx,`UPDATE public.notifications SET read_at=COALESCE(read_at,NOW())
 WHERE user_id=$1 AND id=$2`,user,id)
 return tag.RowsAffected()>0,err
}
func (r *Repository) MarkAllRead(ctx context.Context,user uuid.UUID)error{
 _,err:=r.pool.Exec(ctx,`UPDATE public.notifications SET read_at=NOW()
 WHERE user_id=$1 AND read_at IS NULL`,user)
 return err
}
var _ pgx.Row
