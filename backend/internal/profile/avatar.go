package profile

import (
 "context"
 "errors"

 "github.com/gofrs/uuid"
 "github.com/jackc/pgx/v5"
)
const maxAvatarBytes = 512 * 1024
var ErrAvatarMissing = errors.New("profile picture not found")

type Avatar struct {
 Data []byte
 Mime string
}

func(r *Repository)SaveAvatar(ctx context.Context,uid uuid.UUID,pic Avatar)error{
 tag,err:=r.pool.Exec(ctx,`UPDATE public.users SET avatar_data=$2,avatar_mime=$3 WHERE id=$1`,uid,pic.Data,pic.Mime)
 if err!=nil{return err}
 if tag.RowsAffected()==0{return ErrNotFound}
 return nil
}
func(r *Repository)LoadAvatar(ctx context.Context,uid uuid.UUID)(Avatar,error){
 var a Avatar
 err:=r.pool.QueryRow(ctx,`SELECT avatar_data,avatar_mime FROM public.users WHERE id=$1 AND avatar_data IS NOT NULL`,uid).Scan(&a.Data,&a.Mime)
 if errors.Is(err,pgx.ErrNoRows){return Avatar{},ErrAvatarMissing}
 return a,err
}
func(r *Repository)DeleteAvatar(ctx context.Context,uid uuid.UUID)error{
 tag,err:=r.pool.Exec(ctx,`UPDATE public.users SET avatar_data=NULL,avatar_mime=NULL WHERE id=$1`,uid)
 if err!=nil{return err}
 if tag.RowsAffected()==0{return ErrNotFound}
 return nil
}
// Version allows clients to invalidate their cached avatar after an update.
// It doesn't disclose image content or affect authentication.
func(r *Repository)AvatarPresent(ctx context.Context,uid uuid.UUID)(bool,error){
 var exists bool
 err:=r.pool.QueryRow(ctx,`SELECT avatar_data IS NOT NULL FROM public.users WHERE id=$1`,uid).Scan(&exists)
 if errors.Is(err,pgx.ErrNoRows){return false,ErrNotFound}
 return exists,err
}
