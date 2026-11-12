package membership

import (
 "context"
 "crypto/rand"
 "crypto/sha256"
 "encoding/base64"
 "errors"
 "net/mail"
 "strings"

 "github.com/gofrs/uuid"
 "github.com/jackc/pgx/v5"
)

var ErrInviteLimit = errors.New("too many recent invitations")
var ErrEmailMismatch = errors.New("this invitation is for another email address")
var ErrInvalidEmail = errors.New("enter a valid email address")

// NormalizeInviteEmail accepts a single mailbox address, not display names,
// header fragments, or comma-separated recipients.
func NormalizeInviteEmail(value string) (string,error) {
 value = strings.ToLower(strings.TrimSpace(value))
 if len(value)>254 || strings.ContainsAny(value,"\r\n,") {return "",ErrInvalidEmail}
 parsed,err:=mail.ParseAddress(value)
 if err!=nil || parsed.Address!=value || !strings.Contains(value,"@") {return "",ErrInvalidEmail}
 return value,nil
}

// Serialize all invitation issuance by workspace, then enforce a DB quota
// shared by email invitations and links, even across server processes.
func invitationQuota(ctx context.Context,tx pgx.Tx,workspace uuid.UUID) error {
 _,err:=tx.Exec(ctx,`SELECT pg_advisory_xact_lock(hashtext($1::text))`,workspace)
 if err!=nil{return err}
 var hourly,daily int
 err=tx.QueryRow(ctx,`
 SELECT COUNT(*) FILTER(WHERE created_at > NOW()-INTERVAL '1 hour'),
        COUNT(*) FILTER(WHERE created_at > NOW()-INTERVAL '1 day')
 FROM public.workspace_invitations WHERE workspace_id=$1
 `,workspace).Scan(&hourly,&daily)
 if err!=nil{return err}
 if hourly>=10 || daily>=30{return ErrInviteLimit}
 return nil
}

func newInviteSecret() (string,[]byte,error) {
 var nonce [32]byte
 if _,err:=rand.Read(nonce[:]);err!=nil{return "",nil,err}
 secret:=base64.RawURLEncoding.EncodeToString(nonce[:])
 hash:=sha256.Sum256([]byte(secret))
 return secret,hash[:],nil
}

func (r *LinkRepository) CreateEmail(ctx context.Context,actor,workspace uuid.UUID,address string)(string,uuid.UUID,string,error){
 address,err:=NormalizeInviteEmail(address)
 if err!=nil{return "",uuid.Nil,"",err}
 secret,digest,err:=newInviteSecret()
 if err!=nil{return "",uuid.Nil,"",err}
 tx,err:=r.pool.Begin(ctx)
 if err!=nil{return "",uuid.Nil,"",err}
 defer tx.Rollback(ctx)
 var name string
 err=tx.QueryRow(ctx,`
 SELECT w.name FROM public.workspaces w
 JOIN public.workspace_memberships m ON m.workspace_id=w.id
 WHERE w.id=$1 AND m.user_id=$2 AND m.role='owner'
 `,workspace,actor).Scan(&name)
 if errors.Is(err,pgx.ErrNoRows){return "",uuid.Nil,"",ErrOwnerRequired}
 if err!=nil{return "",uuid.Nil,"",err}
 if err=invitationQuota(ctx,tx,workspace);err!=nil{return "",uuid.Nil,"",err}
 var invitation uuid.UUID
 err=tx.QueryRow(ctx,`
 INSERT INTO public.workspace_invitations(workspace_id,created_by,kind,invited_email,token_hash,expires_at,max_uses)
 VALUES($1,$2,'email',$3,$4,NOW()+INTERVAL '7 days',1) RETURNING id
 `,workspace,actor,address,digest).Scan(&invitation)
 if err!=nil{return "",uuid.Nil,"",err}
 if err=tx.Commit(ctx);err!=nil{return "",uuid.Nil,"",err}
 return secret,invitation,name,nil
}

// If mail fails, invalidate its token. The revocation is deliberately scoped
// to the freshly created invitation and does not affect any other invite.
func (r *LinkRepository) RevokeFailedEmail(ctx context.Context,id uuid.UUID) error {
 _,err:=r.pool.Exec(ctx,`UPDATE public.workspace_invitations SET revoked_at=NOW() WHERE id=$1 AND kind='email' AND revoked_at IS NULL`,id)
 return err
}
