package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetIdByAccountId(ctx context.Context, accountId string) (*string, error)
}

type userRepository struct{ db *pgxpool.Pool }

func NewUserRepositry(db *pgxpool.Pool) UserRepository { return &userRepository{db: db} }

func (r *userRepository) GetIdByAccountId(ctx context.Context, accountId string) (id *string, err error) {
	const q = `
		SELECT id
		FROM users
		WHERE account_id = $1
		LIMIT 1
	`

	if err = r.db.QueryRow(ctx, q, accountId).Scan(&id); err != nil {
		return nil, err
	}

	return id, nil
}
