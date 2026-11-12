package account

import (
    "context"
    "errors"
    "fmt"

    "github.com/gofrs/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "golang.org/x/crypto/bcrypt"
)

var (
    ErrNotFound = errors.New("account not found")
    ErrBadPassword = errors.New("incorrect password")
    ErrSharedOwner = errors.New("transfer ownership of shared workspaces before deleting your account")
    ErrConfirmationRequired = errors.New("confirm deletion of sole-member workspaces")
)

type WorkspaceImpact struct {
    ID uuid.UUID `json:"id"`
    Name string `json:"name"`
    MemberCount int `json:"member_count"`
}
type Impact struct {
    SoleOwned []WorkspaceImpact `json:"sole_owned"`
    SharedOwned []WorkspaceImpact `json:"shared_owned"`
}

type Repository struct { pool *pgxpool.Pool }
func NewRepository(pool *pgxpool.Pool)*Repository { return &Repository{pool:pool} }

func (r *Repository) Impact(ctx context.Context, userID uuid.UUID)(Impact,error){
    var result Impact
    result.SoleOwned=make([]WorkspaceImpact,0)
    result.SharedOwned=make([]WorkspaceImpact,0)
    rows,err:=r.pool.Query(ctx,`
        SELECT w.id, w.name, (
            SELECT COUNT(*) FROM public.workspace_memberships all_members
            WHERE all_members.workspace_id=w.id
        ) AS members
        FROM public.workspaces w
        JOIN public.workspace_memberships m ON m.workspace_id=w.id
        WHERE m.user_id=$1 AND m.role='owner'
        ORDER BY w.name,w.id
    `,userID)
    if err!=nil{return result,err}
    defer rows.Close()
    for rows.Next(){
        var ws WorkspaceImpact
        if err:=rows.Scan(&ws.ID,&ws.Name,&ws.MemberCount);err!=nil{return result,err}
        if ws.MemberCount>1{result.SharedOwned=append(result.SharedOwned,ws)
        }else{result.SoleOwned=append(result.SoleOwned,ws)}
    }
    return result,rows.Err()
}

// Delete is transactional: a bad password, shared workspace, missing consent,
// FK error or concurrent membership change cannot leave a partial deletion.
func(r *Repository) Delete(ctx context.Context,userID uuid.UUID,password string,deleteSoleOwned bool)error {
    tx,err:=r.pool.BeginTx(ctx,pgx.TxOptions{IsoLevel:pgx.Serializable})
    if err!=nil{return err}
    defer tx.Rollback(ctx)

    var passwordHash,email string
    err=tx.QueryRow(ctx,`SELECT password,email FROM public.users WHERE id=$1 FOR UPDATE`,userID).Scan(&passwordHash,&email)
    if errors.Is(err,pgx.ErrNoRows){return ErrNotFound}
    if err!=nil{return err}
    if bcrypt.CompareHashAndPassword([]byte(passwordHash),[]byte(password))!=nil{return ErrBadPassword}

    // Lock owned workspace rows before counting memberships. No other member
    // can successfully join a workspace while it is being deleted.
    rows,err:=tx.Query(ctx,`
        SELECT w.id,w.name
        FROM public.workspaces w
        JOIN public.workspace_memberships m ON m.workspace_id=w.id
        WHERE m.user_id=$1 AND m.role='owner'
        ORDER BY w.id FOR UPDATE OF w
    `,userID)
    if err!=nil{return err}
    owned:=make([]WorkspaceImpact,0)
    for rows.Next(){
        var ws WorkspaceImpact
        if err:=rows.Scan(&ws.ID,&ws.Name);err!=nil{rows.Close();return err}
        owned=append(owned,ws)
    }
    if err:=rows.Err();err!=nil{rows.Close();return err}
    rows.Close()

    for _,ws:=range owned {
        var count int
        if err:=tx.QueryRow(ctx,`SELECT COUNT(*) FROM public.workspace_memberships WHERE workspace_id=$1`,ws.ID).Scan(&count);err!=nil{return err}
        if count>1{return ErrSharedOwner}
        if !deleteSoleOwned{return ErrConfirmationRequired}
    }
    // Remove explicitly consented sole-member workspaces. Their projects,
    // tasks, memberships, invitations and activity cascade by schema design.
    for _,ws:=range owned{
        tag,err:=tx.Exec(ctx,`DELETE FROM public.workspaces WHERE id=$1`,ws.ID)
        if err!=nil{return err}
        if tag.RowsAffected()!=1{return fmt.Errorf("workspace disappeared during account deletion")}
    }

    // Pending invitations created in other people's workspaces have an FK to
    // this user without ON DELETE SET NULL. Revoke these links by deletion.
    if _,err:=tx.Exec(ctx,`DELETE FROM public.workspace_invitations WHERE created_by=$1`,userID);err!=nil{return err}
    // Revoke every browser session *before* deleting the user. Never assume
    // a legacy sessions table has ON DELETE CASCADE.
    if _,err:=tx.Exec(ctx,`DELETE FROM public.sessions WHERE user_id=$1`,userID);err!=nil{return err}
    if _,err:=tx.Exec(ctx,`DELETE FROM public.otps WHERE lower(email)=lower($1)`,email);err!=nil{return err}
    tag,err:=tx.Exec(ctx,`DELETE FROM public.users WHERE id=$1`,userID)
    if err!=nil{return err}
    if tag.RowsAffected()!=1{return ErrNotFound}
    return tx.Commit(ctx)
}
