package repo

import "context"

func (r *Repo) InsertRefreshToken(ctx context.Context, userId int, ref string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO refresh_tokens (user_id, token) VALUES ($1, $2)",
		userId, ref,
	)
	return err
}
