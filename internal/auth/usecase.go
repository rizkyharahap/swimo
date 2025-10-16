package auth

import (
	"context"
	"errors"
	"strings"
	"time"

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
	SignIn(ctx context.Context, req SignInRequest, userAgent string) (*SignInResponse, error)
	SignInGuest(ctx context.Context, req SignInGuestRequest, userAgent string) (*SignInGuestResponse, error)
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

func (uc *authUsecase) SignIn(ctx context.Context, req SignInRequest, userAgent string) (*SignInResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	auth, err := uc.authRepo.GetAuthByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if auth.IsLocked {
		return nil, ErrLocked
	}

	if err = auth.ComparePassword(req.Password); err != nil {
		return nil, err
	}

	// create session with refresh token
	session, err := NewSession(&uc.cfg.Auth, userAgent, auth.AccountID)
	if err != nil {
		return nil, err
	}

	sessionId, err := uc.authRepo.CreateUserSession(ctx, session)
	if err != nil {
		return nil, err
	}

	accessToken, exp, err := NewAccessToken(uc.cfg.Auth.JWTSecret, "user", auth.AccountID, sessionId, uc.cfg.Auth.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	return &SignInResponse{
		Name:         auth.Name,
		Email:        auth.Email,
		Age:          auth.AgeYears,
		Height:       auth.HeightCM,
		Weight:       auth.WeightKG,
		Token:        accessToken,
		RefreshToken: session.RefreshTokenHash,
		ExpiresInMs:  time.Until(exp).Milliseconds(),
	}, nil
}

func (uc *authUsecase) SignInGuest(ctx context.Context, req SignInGuestRequest, userAgent string) (*SignInGuestResponse, error) {
	if !uc.cfg.Auth.GuestEnabled {
		return nil, ErrGuestDisabled
	}

	if uc.cfg.Auth.GuestRatePerMinute > 0 {
		since := time.Now().UTC().Add(-1 * time.Minute)

		count, err := uc.authRepo.CountRecentGuestByUsertAgent(ctx, userAgent, since)
		if err == nil && count >= uc.cfg.Auth.GuestRatePerMinute {
			return nil, ErrGuestLimited
		}
	}

	// Create session with refresh token
	session, err := NewSession(&uc.cfg.Auth, userAgent, "")
	if err != nil {
		return nil, err
	}

	sessionId, err := uc.authRepo.CreateGuestSession(ctx, session)
	if err != nil {
		return nil, err
	}

	access, exp, err := NewAccessToken(uc.cfg.Auth.JWTSecret, "guest", "", sessionId, uc.cfg.Auth.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	return &SignInGuestResponse{
		Name:         "Guest",
		Weight:       req.Weight,
		Height:       req.Height,
		Age:          req.Age,
		Token:        access,
		RefreshToken: session.RefreshTokenHash,
		ExpiresInMs:  time.Until(exp).Milliseconds(),
	}, nil
}
