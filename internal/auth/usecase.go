package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrGuestDisabled = errors.New("guest sign in disabled")
	ErrGuestLimited  = errors.New("guest sign in rate limited")
	ErrLocked        = errors.New("account locked")
)

type AuthUsecase interface {
	SignUp(ctx context.Context, req SignUpRequest) error
}

type authUsecase struct {
	cfg      *config.Config
	logger   *logger.Logger
	pool     *pgxpool.Pool
	authRepo AuthRepository
}

func NewAuthUsecase(cfg *config.Config, logger *logger.Logger, pool *pgxpool.Pool, authRepo AuthRepository) AuthUsecase {
	return &authUsecase{cfg, logger, pool, authRepo}
}

func (uc *authUsecase) SignUp(ctx context.Context, req SignUpRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Transaction Start
	tx, err := uc.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create account
	email := strings.TrimSpace(strings.ToLower(req.Email))

	accountID, err := uc.authRepo.CreateAccount(ctx, tx, email, string(hash))
	if err != nil {
		uc.logger.Warn("signup: create account failed, rolling back", "email", email, "error", err)
		return err
	}

	// Create user profile
	user := req.ToUserEntity(accountID)

	_, err = uc.authRepo.CreateUser(ctx, tx, user)
	if err != nil {
		uc.logger.Warn("signup: create user failed, rolling back", "account_id", accountID, "error", err)
		return err // tx rollback by defer
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		uc.logger.Error("signup: commit transaction failed", "email", email, "error", err)
		return err
	}

	uc.logger.Info("signup success", "email", email)
	return nil
}
