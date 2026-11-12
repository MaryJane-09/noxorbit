package profile

import (
    "context"
    "errors"
    "strings"

    "github.com/gofrs/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "golang.org/x/crypto/bcrypt"
)

var (
    ErrNotFound = errors.New("account not found")
    ErrWrongPassword = errors.New("incorrect current password")
    ErrSamePassword = errors.New("new password must be different")
)

type Profile struct {
    ID uuid.UUID `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
}
type Repository struct {pool *pgxpool.Pool}
func NewRepository(pool *pgxpool.Pool)*Repository{return &Repository{pool:pool}}

// Only the current session's account can be edited.
func(r *Repository)UpdateName(ctx context.Context, uid uuid.UUID, name string)(Profile,error){
    var p Profile
    err:=r.pool.QueryRow(ctx,`
        UPDATE public.users SET name=$2 WHERE id=$1
        RETURNING id,name,email
    `,uid,strings.TrimSpace(name)).Scan(&p.ID,&p.Name,&p.Email)
    if errors.Is(err,pgx.ErrNoRows){return p,ErrNotFound}
    return p,err
}

// Locking the account row serializes password changes. All sessions (including
// this one) are revoked in the SAME transaction as the password replacement.
func(r *Repository)ChangePassword(ctx context.Context,uid uuid.UUID,current,newPassword string)error{
    tx,err:=r.pool.Begin(ctx)
    if err!=nil{return err}
    defer tx.Rollback(ctx)
    var stored string
    err=tx.QueryRow(ctx,`SELECT password FROM public.users WHERE id=$1 FOR UPDATE`,uid).Scan(&stored)
    if errors.Is(err,pgx.ErrNoRows){return ErrNotFound}
    if err!=nil{return err}
    if bcrypt.CompareHashAndPassword([]byte(stored),[]byte(current))!=nil{return ErrWrongPassword}
    if bcrypt.CompareHashAndPassword([]byte(stored),[]byte(newPassword))==nil{return ErrSamePassword}
    hash,err:=bcrypt.GenerateFromPassword([]byte(newPassword),bcrypt.DefaultCost)
    if err!=nil{return err}
    if _,err:=tx.Exec(ctx,`UPDATE public.users SET password=$2 WHERE id=$1`,uid,string(hash));err!=nil{return err}
    if _,err:=tx.Exec(ctx,`DELETE FROM public.sessions WHERE user_id=$1`,uid);err!=nil{return err}
    return tx.Commit(ctx)
}
