package repo

import (
	"context"
	"database/sql"
	"errors"
	"stocks_calculator/internal/model"
	"stocks_calculator/internal/server/app"

	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

func (r *Repo) CreateNewUser(ctx context.Context, user string, hash string) (int, error) {
	var id int
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id",
		user, hash,
	).Scan(&id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pqerror.UniqueViolation {
			return 0, app.ErrUserExists
		}
		return 0, err
	}

	return id, nil
}

func (r *Repo) FindUserByUsername(ctx context.Context, username string) (model.User, error) {
	user := model.User{Name: username}
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, password FROM users WHERE name = $1",
		username,
	).Scan(&user.Id, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, app.ErrWrongCredentials
		}
		return user, err
	}
	return user, nil
}

func (r *Repo) UserIdByRefreshToken(ctx context.Context, refToken string) (int, error) {
	var id int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT user_id FROM refresh_tokens WHERE token = $1",
		refToken,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, app.ErrInvalidRefreshToken
		}
		return 0, err
	}
	return id, nil
}
