package project

import (
 "context"
 "errors"
 "time"
 "github.com/gofrs/uuid"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgxpool"
)

var ErrProjectMemberConflict=errors.New("already a project participant")
var ErrNotProjectMember=errors.New("project participant not found")
var ErrInvalidProjectMember=errors.New("user must belong to this workspace")

type Participant struct{
 UserID uuid.UUID `json:"user_id"`
 Name string `json:"name"`
 Role string `json:"role"`
 JoinedAt time.Time `json:"joined_at"`
}
type ParticipantsRepository struct{pool *pgxpool.Pool}
func NewParticipantsRepository(pool *pgxpool.Pool)*ParticipantsRepository{return &ParticipantsRepository{pool:pool}}

func (r *ParticipantsRepository) access(ctx context.Context,actor,wid,pid uuid.UUID)(bool,bool,error){
 var visible,manager bool
 err:=r.pool.QueryRow(ctx,`
 SELECT EXISTS(
   SELECT 1 FROM public.projects p JOIN public.workspace_memberships m
    ON m.workspace_id=p.workspace_id AND m.user_id=$1
   WHERE p.workspace_id=$2 AND p.id=$3
 ), EXISTS(
   SELECT 1 FROM public.projects p JOIN public.workspace_memberships m
     ON m.workspace_id=p.workspace_id AND m.user_id=$1
   WHERE p.workspace_id=$2 AND p.id=$3 AND
     (m.role='owner' OR EXISTS(SELECT 1 FROM public.project_memberships pm
       WHERE pm.project_id=p.id AND pm.user_id=$1 AND pm.role='lead'))
 )
 `,actor,wid,pid).Scan(&visible,&manager)
 return visible,manager,err
}
func (r *ParticipantsRepository) List(ctx context.Context,actor,wid,pid uuid.UUID)([]Participant,error){
 visible,_,err:=r.access(ctx,actor,wid,pid)
 if err!=nil{return nil,err}
 if !visible{return nil,ErrNotFound}
 rows,err:=r.pool.Query(ctx,`
 SELECT pm.user_id,u.name,pm.role,pm.joined_at
 FROM public.project_memberships pm
 JOIN public.users u ON u.id=pm.user_id
 JOIN public.projects p ON p.id=pm.project_id AND p.workspace_id=pm.workspace_id
 WHERE pm.project_id=$3 AND pm.workspace_id=$2 AND EXISTS(
   SELECT 1 FROM public.workspace_memberships m
   WHERE m.workspace_id=p.workspace_id AND m.user_id=$1
 )
 ORDER BY CASE pm.role WHEN 'lead' THEN 0 ELSE 1 END,u.name,pm.user_id
 `,actor,wid,pid)
 if err!=nil{return nil,err}
 defer rows.Close()
 items:=make([]Participant,0)
 for rows.Next(){
  var v Participant
  if err:=rows.Scan(&v.UserID,&v.Name,&v.Role,&v.JoinedAt);err!=nil{return nil,err}
  items=append(items,v)
 }
 return items,rows.Err()
}
func (r *ParticipantsRepository) Add(ctx context.Context,actor,wid,pid,target uuid.UUID,role string)(Participant,error){
 var v Participant
 err:=r.pool.QueryRow(ctx,`
 WITH allowed AS (
  SELECT p.id,p.workspace_id FROM public.projects p
  JOIN public.workspace_memberships a
   ON a.workspace_id=p.workspace_id AND a.user_id=$1
  WHERE p.id=$3 AND p.workspace_id=$2 AND p.status='active'
    AND (a.role='owner' OR EXISTS(
      SELECT 1 FROM public.project_memberships lead
      WHERE lead.project_id=p.id AND lead.user_id=$1 AND lead.role='lead'))
    AND ($5='member' OR a.role='owner')
 ), inserted AS (
  INSERT INTO public.project_memberships(project_id,workspace_id,user_id,role)
  SELECT allowed.id,allowed.workspace_id,t.user_id,$5
  FROM allowed JOIN public.workspace_memberships t
   ON t.workspace_id=allowed.workspace_id AND t.user_id=$4
  ON CONFLICT(project_id,user_id) DO NOTHING
  RETURNING user_id,role,joined_at
 )
 SELECT inserted.user_id,u.name,inserted.role,inserted.joined_at
 FROM inserted JOIN public.users u ON u.id=inserted.user_id
 `,actor,wid,pid,target,role).Scan(&v.UserID,&v.Name,&v.Role,&v.JoinedAt)
 if err==nil{return v,nil}
 if !errors.Is(err,pgx.ErrNoRows){return v,err}
 visible,manager,e:=r.access(ctx,actor,wid,pid)
 if e!=nil{return v,e}
 if !visible{return v,ErrNotFound}
 if !manager{return v,ErrForbidden}
 if role=="lead" {
  var owner bool
  e=r.pool.QueryRow(ctx,`SELECT EXISTS(SELECT 1 FROM public.workspace_memberships
   WHERE workspace_id=$1 AND user_id=$2 AND role='owner')`,wid,actor).Scan(&owner)
  if e!=nil{return v,e}
  if !owner{return v,ErrForbidden}
 }
 var member,participant bool
 e=r.pool.QueryRow(ctx,`
 SELECT EXISTS(SELECT 1 FROM public.workspace_memberships WHERE workspace_id=$1 AND user_id=$2),
 EXISTS(SELECT 1 FROM public.project_memberships WHERE project_id=$3 AND user_id=$2)
 `,wid,target,pid).Scan(&member,&participant)
 if e!=nil{return v,e}
 if participant{return v,ErrProjectMemberConflict}
 if !member{return v,ErrInvalidProjectMember}
 return v,ErrForbidden
}
func (r *ParticipantsRepository) SetRole(ctx context.Context,actor,wid,pid,target uuid.UUID,role string)(Participant,error){
 var v Participant
 err:=r.pool.QueryRow(ctx,`
 WITH changed AS (
 UPDATE public.project_memberships pm SET role=$5
 FROM public.projects p JOIN public.workspace_memberships a
  ON a.workspace_id=p.workspace_id AND a.user_id=$1 AND a.role='owner'
 WHERE pm.project_id=p.id AND pm.workspace_id=$2 AND p.id=$3 AND pm.user_id=$4
 AND p.status='active'
 RETURNING pm.user_id,pm.role,pm.joined_at
 )
 SELECT changed.user_id,u.name,changed.role,changed.joined_at
 FROM changed JOIN public.users u ON u.id=changed.user_id
 `,actor,wid,pid,target,role).Scan(&v.UserID,&v.Name,&v.Role,&v.JoinedAt)
 if err==nil{return v,nil}
 if !errors.Is(err,pgx.ErrNoRows){return v,err}
 visible,_,e:=r.access(ctx,actor,wid,pid)
 if e!=nil{return v,e}
 if !visible{return v,ErrNotFound}
 return v,ErrForbidden
}
func (r *ParticipantsRepository) Remove(ctx context.Context,actor,wid,pid,target uuid.UUID)error{
 // The workspace owner can remove any participant. A project lead may
 // remove ordinary members but cannot demote/remove other leads.
 tag,err:=r.pool.Exec(ctx,`
 DELETE FROM public.project_memberships pm
 USING public.projects p,public.workspace_memberships a
 WHERE pm.project_id=p.id AND pm.workspace_id=p.workspace_id
 AND p.workspace_id=$2 AND p.id=$3 AND pm.user_id=$4 AND p.status='active'
 AND a.workspace_id=p.workspace_id AND a.user_id=$1
 AND (a.role='owner' OR (
    pm.role='member' AND EXISTS(SELECT 1 FROM public.project_memberships lead
      WHERE lead.project_id=p.id AND lead.user_id=$1 AND lead.role='lead')
 ))
 `,actor,wid,pid,target)
 if err!=nil{return err}
 if tag.RowsAffected()!=0{return nil}
 visible,manager,e:=r.access(ctx,actor,wid,pid)
 if e!=nil{return e}
 if !visible{return ErrNotFound}
 if !manager{return ErrForbidden}
 return ErrNotProjectMember
}
