package membership

import (
    "context"
    "errors"
    "time"

    "github.com/gofrs/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var (
    ErrNotMember = errors.New("member not found")
    ErrCannotRemoveOwner = errors.New("workspace owner cannot be removed")
)

type InviteSummary struct {
    ID uuid.UUID `json:"id"`
    Kind string `json:"kind"`
    InvitedEmail *string `json:"invited_email"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
    RevokedAt *time.Time `json:"revoked_at"`
    MaxUses int `json:"max_uses"`
    UseCount int `json:"use_count"`
}

// These operations never return raw tokens, and owners can only manage their
// own workspace's invitations.
func (r *LinkRepository) ListInvites(ctx context.Context, owner, workspace uuid.UUID) ([]InviteSummary, error) {
    var allowed bool
    err := r.pool.QueryRow(ctx,`
        SELECT EXISTS (SELECT 1 FROM public.workspace_memberships
            WHERE workspace_id=$1 AND user_id=$2 AND role='owner')
    `,workspace,owner).Scan(&allowed)
    if err != nil {return nil,err}
    if !allowed {return nil,ErrOwnerRequired}
    rows,err:=r.pool.Query(ctx,`
        SELECT id,kind,invited_email,created_at,expires_at,revoked_at,max_uses,use_count
        FROM public.workspace_invitations
        WHERE workspace_id=$1 AND EXISTS (
            SELECT 1 FROM public.workspace_memberships
            WHERE workspace_id=$1 AND user_id=$2 AND role='owner')
        ORDER BY created_at DESC,id DESC LIMIT 100
    `,workspace,owner)
    if err != nil {return nil,err}
    defer rows.Close()
    invites:=make([]InviteSummary,0)
    for rows.Next() {
        var item InviteSummary
        if err:=rows.Scan(&item.ID,&item.Kind,&item.InvitedEmail,&item.CreatedAt,
            &item.ExpiresAt,&item.RevokedAt,&item.MaxUses,&item.UseCount);err!=nil{return nil,err}
        invites=append(invites,item)
    }
    return invites,rows.Err()
}

// UPDATE checks ownership in the same statement as the mutation.
func (r *LinkRepository) RevokeInvite(ctx context.Context, owner, workspace, invitation uuid.UUID) error {
    var changed uuid.UUID
    err:=r.pool.QueryRow(ctx,`
        UPDATE public.workspace_invitations AS i
        SET revoked_at=NOW()
        WHERE i.id=$3 AND i.workspace_id=$1 AND i.revoked_at IS NULL
          AND EXISTS (SELECT 1 FROM public.workspace_memberships m
            WHERE m.workspace_id=i.workspace_id AND m.user_id=$2 AND m.role='owner')
        RETURNING i.id
    `,workspace,owner,invitation).Scan(&changed)
    if errors.Is(err,pgx.ErrNoRows) {return ErrNotFound}
    return err
}

type MemberManagementRepository struct { pool *pgxpool.Pool }
func NewMemberManagementRepository(pool *pgxpool.Pool)*MemberManagementRepository{
    return &MemberManagementRepository{pool:pool}
}

// Target must be a member, never the workspace owner. Existing task
// assignments are automatically cleared by the schema's FK ON DELETE SET NULL.
func (r *MemberManagementRepository) RemoveMember(ctx context.Context, owner, workspace, target uuid.UUID) error {
    var removed uuid.UUID
    err:=r.pool.QueryRow(ctx,`
        WITH removed AS (
            DELETE FROM public.workspace_memberships AS target
            WHERE target.workspace_id=$1 AND target.user_id=$3 AND target.role='member'
              AND EXISTS (SELECT 1 FROM public.workspace_memberships m
                WHERE m.workspace_id=$1 AND m.user_id=$2 AND m.role='owner')
            RETURNING target.workspace_id, target.user_id
        ), recorded AS (
            INSERT INTO public.activity_events
                (workspace_id, actor_id, action, entity_type, entity_id, entity_name)
            SELECT workspace_id, $2, 'member.removed', 'member', user_id, 'Team member'
            FROM removed
            RETURNING entity_id
        )
        SELECT entity_id FROM recorded
    `,workspace,owner,target).Scan(&removed)
    if err==nil {return nil}
    if !errors.Is(err,pgx.ErrNoRows) {return err}
    var callerOwner, targetOwner bool
    err=r.pool.QueryRow(ctx,`
        SELECT EXISTS(SELECT 1 FROM public.workspace_memberships WHERE workspace_id=$1 AND user_id=$2 AND role='owner'),
               EXISTS(SELECT 1 FROM public.workspace_memberships WHERE workspace_id=$1 AND user_id=$3 AND role='owner')
    `,workspace,owner,target).Scan(&callerOwner,&targetOwner)
    if err!=nil {return err}
    if !callerOwner {return ErrOwnerRequired}
    if targetOwner {return ErrCannotRemoveOwner}
    return ErrNotMember
}
