package profile

import (
 "context"
 "errors"
 "strings"
 "time"

 "github.com/MaryJane-09/noxorbit/backend/internal/otp"
 "github.com/gofrs/uuid"
 "github.com/jackc/pgx/v5"
 "github.com/jackc/pgx/v5/pgconn"
 "golang.org/x/crypto/bcrypt"
)

var (
 ErrEmailUnchanged = errors.New("new email matches current email")
 ErrEmailTaken = errors.New("email already in use")
 ErrEmailCooldown = errors.New("wait before requesting another code")
 ErrEmailCodeInvalid = errors.New("invalid or expired verification code")
 ErrEmailCodeLocked = errors.New("too many verification attempts")
)

func(r *Repository)IssueEmailChange(ctx context.Context,uid uuid.UUID,password,newEmail,codeHash string)(string,error){
 var name,email,stored string
 err:=r.pool.QueryRow(ctx,`SELECT name,email,password FROM public.users WHERE id=$1`,uid).Scan(&name,&email,&stored)
 if errors.Is(err,pgx.ErrNoRows){return "",ErrNotFound}
 if err!=nil{return "",err}
 if bcrypt.CompareHashAndPassword([]byte(stored),[]byte(password))!=nil{return "",ErrWrongPassword}
 if strings.EqualFold(email,newEmail){return "",ErrEmailUnchanged}
 var taken bool
 err=r.pool.QueryRow(ctx,`SELECT EXISTS(SELECT 1 FROM public.users WHERE lower(email)=lower($1) AND id<>$2)`,newEmail,uid).Scan(&taken)
 if err!=nil{return "",err}
 if taken{return "",ErrEmailTaken}
 tag,err:=r.pool.Exec(ctx,`
    INSERT INTO public.pending_email_changes (user_id,new_email,code_hash,requested_at,expires_at,attempts)
    VALUES ($1,$2,$3,NOW(),NOW()+INTERVAL '5 minutes',0)
    ON CONFLICT (user_id) DO UPDATE
    SET new_email=EXCLUDED.new_email,code_hash=EXCLUDED.code_hash,
        requested_at=NOW(),expires_at=EXCLUDED.expires_at,attempts=0
    WHERE pending_email_changes.requested_at <= NOW()-INTERVAL '1 minute'
 `,uid,newEmail,codeHash)
 if err!=nil{return "",err}
 if tag.RowsAffected()==0{return "",ErrEmailCooldown}
 return name,nil
}

// If sending the code fails, remove only the corresponding pending request.
// Never remove a newer code submitted by another request.
func(r *Repository)DiscardEmailChange(ctx context.Context,uid uuid.UUID,codeHash string)error{
 _,err:=r.pool.Exec(ctx,`DELETE FROM public.pending_email_changes WHERE user_id=$1 AND code_hash=$2`,uid,codeHash)
 return err
}

// Confirm holds the pending row lock through verification and email mutation.
// Incorrect guesses are committed and capped in SQL to resist concurrent guesses.
func(r *Repository)ConfirmEmailChange(ctx context.Context,uid uuid.UUID,code string)(Profile,error){
 var result Profile
 tx,err:=r.pool.Begin(ctx)
 if err!=nil{return result,err}
 defer tx.Rollback(ctx)
 var address,storedHash string
 var attempts int
 var expiry time.Time
 err=tx.QueryRow(ctx,`
    SELECT new_email,code_hash,attempts,expires_at
    FROM public.pending_email_changes WHERE user_id=$1 FOR UPDATE
 `,uid).Scan(&address,&storedHash,&attempts,&expiry)
 if errors.Is(err,pgx.ErrNoRows){return result,ErrEmailCodeInvalid}
 if err!=nil{return result,err}
 if time.Now().After(expiry){return result,ErrEmailCodeInvalid}
 if attempts>=otp.MaxAttempts{return result,ErrEmailCodeLocked}
 if bcrypt.CompareHashAndPassword([]byte(storedHash),[]byte(code))!=nil{
    _,err=tx.Exec(ctx,`UPDATE public.pending_email_changes SET attempts=attempts+1 WHERE user_id=$1`,uid)
    if err!=nil{return result,err}
    if err=tx.Commit(ctx);err!=nil{return result,err}
    if attempts+1>=otp.MaxAttempts{return result,ErrEmailCodeLocked}
    return result,ErrEmailCodeInvalid
 }
 // Lock account to serialize concurrent account changes with verified email.
 err=tx.QueryRow(ctx,`
    UPDATE public.users SET email=$2 WHERE id=$1
    RETURNING id,name,email
 `,uid,address).Scan(&result.ID,&result.Name,&result.Email)
 if errors.Is(err,pgx.ErrNoRows){return Profile{},ErrNotFound}
 if err!=nil{
    var pgerr *pgconn.PgError
    if errors.As(err,&pgerr)&&pgerr.Code=="23505"{return Profile{},ErrEmailTaken}
    return Profile{},err
 }
 if _,err=tx.Exec(ctx,`DELETE FROM public.pending_email_changes WHERE user_id=$1`,uid);err!=nil{return Profile{},err}
 // Email is the login identity; revoke prior sessions after it changes.
 if _,err=tx.Exec(ctx,`DELETE FROM public.sessions WHERE user_id=$1`,uid);err!=nil{return Profile{},err}
 return result,tx.Commit(ctx)
}
