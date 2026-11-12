package session

import (
	"context"
	"errors"
	"time"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("session not found")

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool: pool,
	}
}

func (r *SessionRepository) Create(userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
    return r.CreateWithAgent(userID, tokenHash, expiresAt, "")
}

// Record a short browser description but never persist the raw session token.
func (r *SessionRepository) CreateWithAgent(userID uuid.UUID, tokenHash string, expiresAt time.Time, agent string) error {
    if len(agent)>256 { agent=agent[:256] }
    _, err:=r.pool.Exec(context.Background(), `
        INSERT INTO public.sessions (token_hash,user_id,expires_at,user_agent)
        VALUES ($1,$2,$3,$4)
    `,tokenHash,userID,expiresAt,agent)
    return err
}

func (r *SessionRepository) FindUserID(tokenHash string) (uuid.UUID, error) {
	ctx := context.Background()
	var userID uuid.UUID

	err := r.pool.QueryRow(ctx, `SELECT user_id FROM sessions WHERE token_hash = $1 AND expires_at > now()`, tokenHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, err
	}

	return userID, nil
}

func (r *SessionRepository) Delete(tokenHash string) error {
	ctx := context.Background()

	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *SessionRepository) DeleteAllForUser(userID uuid.UUID) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// DeviceSession contains no token or token hash, so it is safe to return
// to the owner of the account after authentication.
type DeviceSession struct {
 ID uuid.UUID `json:"id"`
 UserAgent string `json:"user_agent"`
 CreatedAt *time.Time `json:"created_at"`
 ExpiresAt time.Time `json:"expires_at"`
 Current bool `json:"current"`
}

func(r *SessionRepository) ListForUser(ctx context.Context, uid uuid.UUID, currentHash string)([]DeviceSession,error){
 result:=make([]DeviceSession,0)
 rows,err:=r.pool.Query(ctx,`
    SELECT public_id,user_agent,created_at,expires_at,token_hash=$2
    FROM public.sessions
    WHERE user_id=$1 AND expires_at>NOW()
    ORDER BY created_at DESC,public_id DESC
 `,uid,currentHash)
 if err!=nil{return nil,err}
 defer rows.Close()
 for rows.Next(){
    var item DeviceSession
    if err:=rows.Scan(&item.ID,&item.UserAgent,&item.CreatedAt,&item.ExpiresAt,&item.Current);err!=nil{return nil,err}
    result=append(result,item)
 }
 if err:=rows.Err();err!=nil{return nil,err}
 return result,nil
}

// Session IDs are random, public-facing UUIDs and must always be scoped
// to the authenticated owner in the SQL predicate.
func(r *SessionRepository) Revoke(ctx context.Context,uid,id uuid.UUID)(bool,error){
 tag,err:=r.pool.Exec(ctx,`DELETE FROM public.sessions WHERE user_id=$1 AND public_id=$2`,uid,id)
 if err!=nil{return false,err}
 return tag.RowsAffected()==1,nil
}

func(r *SessionRepository) RevokeOthers(ctx context.Context,uid uuid.UUID,currentHash string)(int64,error){
 tag,err:=r.pool.Exec(ctx,`DELETE FROM public.sessions WHERE user_id=$1 AND token_hash<>$2`,uid,currentHash)
 if err!=nil{return 0,err}
 return tag.RowsAffected(),nil
}
