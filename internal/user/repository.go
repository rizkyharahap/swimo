package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository interface {
	GetIdByAccountId(ctx context.Context, accountId string) (*string, error)
	GetUserById(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, tx pgx.Tx, user *User) (*User, error)
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

func (r *userRepository) GetUserById(ctx context.Context, id string) (*User, error) {
	const q = `
		SELECT id, name, weight_kg, height_cm, age_years, gender
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var user User
	if err := r.db.QueryRow(ctx, q, id).Scan(&user.ID, &user.Name, &user.WeightKG, &user.HeightCM, &user.AgeYears, &user.Gender); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, tx pgx.Tx, user *User) (*User, error) {
	const q = `
		INSERT INTO users (account_id, name, gender, weight_kg, height_cm, age_years)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id`

	if err := tx.QueryRow(ctx, q,
		&user.AccountID,
		&user.Name,
		&user.Gender,
		&user.WeightKG,
		&user.HeightCM,
		&user.AgeYears,
	).Scan(&user.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrUserExists
		}

		return nil, err
	}

	return user, nil
}
