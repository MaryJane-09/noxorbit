package membership

import (
    "context"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "strings"
    "time"

    "github.com/gofrs/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

var (
    ErrOwnerRequired = errors.New("only workspace owners may invite members")
    ErrInvalidInvite = errors.New("invalid invitation")
    ErrExpiredInvite = errors.New("invitation expired, revoked, or already used")
)

// LinkRepository is backed by the same trusted database connection as the
// membership directory, but invitation secrets are never stored in plaintext.
type LinkRepository struct { pool *pgxpool.Pool }
func NewLinkRepository(pool *pgxpool.Pool) *LinkRepository { return &LinkRepository{pool:pool} }

func (r *LinkRepository) CreateLink(ctx context.Context,actor,workspaceID uuid.UUID)(string,error) {
    secret,digest,err:=newInviteSecret()
    if err!=nil{return "",err}
    tx,err:=r.pool.Begin(ctx)
    if err!=nil{return "",err}
    defer tx.Rollback(ctx)
    var owner bool
    err=tx.QueryRow(ctx,`SELECT EXISTS(SELECT 1 FROM public.workspace_memberships WHERE workspace_id=$1 AND user_id=$2 AND role='owner')`,workspaceID,actor).Scan(&owner)
    if err!=nil{return "",err}
    if !owner{return "",ErrOwnerRequired}
    if err=invitationQuota(ctx,tx,workspaceID);err!=nil{return "",err}
    var id uuid.UUID
    err=tx.QueryRow(ctx,`
        INSERT INTO public.workspace_invitations
            (workspace_id,created_by,kind,token_hash,expires_at,max_uses)
        VALUES ($1,$2,'link',$3,NOW()+INTERVAL '7 days',1)
        RETURNING id
    `,workspaceID,actor,digest).Scan(&id)
    if err!=nil{return "",err}
    if err=tx.Commit(ctx);err!=nil{return "",err}
    return secret,nil
}

// AcceptLink serializes claims on a single invitation with a row lock so
// concurrent requests cannot exceed its max_uses. Login is mandatory.
func (r *LinkRepository) AcceptLink(ctx context.Context,userID uuid.UUID,secret string)(uuid.UUID,error) {
    decoded,err:=base64.RawURLEncoding.DecodeString(secret)
    if err!=nil || len(decoded)!=32 {return uuid.Nil,ErrInvalidInvite}
    digest:=sha256.Sum256([]byte(secret))
    tx,err:=r.pool.BeginTx(ctx,pgx.TxOptions{})
    if err!=nil{return uuid.Nil,err}
    defer tx.Rollback(ctx)
    var inviteID,workspaceID uuid.UUID
    var expires time.Time
    var revoked *time.Time
    var maxUses,uses int
    var kind string
    var invitedEmail *string
    err=tx.QueryRow(ctx,`
        SELECT id,workspace_id,expires_at,revoked_at,max_uses,use_count,kind,invited_email
        FROM public.workspace_invitations
        WHERE token_hash=$1 FOR UPDATE
    `,digest[:]).Scan(&inviteID,&workspaceID,&expires,&revoked,&maxUses,&uses,&kind,&invitedEmail)
    if errors.Is(err,pgx.ErrNoRows){return uuid.Nil,ErrInvalidInvite}
    if err!=nil{return uuid.Nil,err}
    if kind=="email" {
        var accountEmail string
        err=tx.QueryRow(ctx,`SELECT email FROM public.users WHERE id=$1`,userID).Scan(&accountEmail)
        if errors.Is(err,pgx.ErrNoRows){return uuid.Nil,ErrInvalidInvite}
        if err!=nil{return uuid.Nil,err}
        if invitedEmail==nil || strings.ToLower(strings.TrimSpace(accountEmail))!=*invitedEmail {
            return uuid.Nil,ErrEmailMismatch
        }
    }
    var alreadyMember bool
    err=tx.QueryRow(ctx,`
        SELECT EXISTS(SELECT 1 FROM public.workspace_memberships
        WHERE workspace_id=$1 AND user_id=$2)
    `,workspaceID,userID).Scan(&alreadyMember)
    if err!=nil{return uuid.Nil,err}
    if alreadyMember {
        if err:=tx.Commit(ctx);err!=nil{return uuid.Nil,err}
        return workspaceID,nil
    }
    if revoked!=nil || !time.Now().Before(expires) || uses>=maxUses {
        return uuid.Nil,ErrExpiredInvite
    }
    tag,err:=tx.Exec(ctx,`
        INSERT INTO public.workspace_memberships(workspace_id,user_id,role)
        VALUES ($1,$2,'member') ON CONFLICT DO NOTHING
    `,workspaceID,userID)
    if err!=nil{return uuid.Nil,err}
    if tag.RowsAffected()==1 {
        _,err=tx.Exec(ctx,`
            UPDATE public.workspace_invitations SET use_count=use_count+1
            WHERE id=$1
        `,inviteID)
        if err!=nil{return uuid.Nil,err}
    }
    if err:=tx.Commit(ctx);err!=nil{return uuid.Nil,err}
    return workspaceID,nil
}
