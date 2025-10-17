package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/internal/user"
	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrGuestDisabled       = errors.New("guest sign in disabled")
	ErrGuestLimited        = errors.New("guest sign in rate limited")
	ErrLocked              = errors.New("account locked")
	ErrExpiredRefreshToken = errors.New("expired refresh token")
)

type AuthUsecase interface {
	SignUp(ctx context.Context, req SignUpRequest) error
	SignIn(ctx context.Context, req SignInRequest, userAgent string) (*SignInResponse, error)
	SignInGuest(ctx context.Context, req SignInGuestRequest, userAgent string) (*SignInGuestResponse, error)
	SignOut(ctx context.Context, sessionId string) error
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
}

type authUsecase struct {
	cfg      *config.Config
	log      *logger.Logger
	pool     *pgxpool.Pool
	authRepo AuthRepository
	userRepo user.UserRepository
}

func NewAuthUsecase(cfg *config.Config, log *logger.Logger, pool *pgxpool.Pool, authRepo AuthRepository, userRepo user.UserRepository) AuthUsecase {
	return &authUsecase{cfg, log, pool, authRepo, userRepo}
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
		uc.log.Warn("signup: create account failed, rolling back", "email", email, "error", err)
		return err
	}

	// Create user profile
	user := req.ToUserEntity(accountID)

	_, err = uc.authRepo.CreateUser(ctx, tx, user)
	if err != nil {
		uc.log.Warn("signup: create user failed, rolling back", "account_id", accountID, "error", err)
		return err // tx rollback by defer
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		uc.log.Error("signup: commit transaction failed", "email", email, "error", err)
		return err
	}

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

	// revoke another session
	if err := uc.authRepo.RevokeSessionByAccountId(ctx, auth.AccountID, userAgent); err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
	}

	// create session with refresh token
	accessToken, err := uc.createSessionToken(ctx, "user", userAgent, &auth.AccountID)
	if err != nil {
		return nil, err
	}

	return &SignInResponse{
		Name:         auth.Name,
		Email:        auth.Email,
		Age:          auth.AgeYears,
		Height:       auth.HeightCM,
		Weight:       auth.WeightKG,
		Token:        accessToken.Token,
		RefreshToken: accessToken.RefreshToken,
		ExpiresIn:    accessToken.ExpiresInMs,
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

	accessToken, err := uc.createSessionToken(ctx, "guest", userAgent, nil)
	if err != nil {
		return nil, err
	}

	return &SignInGuestResponse{
		Name:         "Guest",
		Weight:       req.Weight,
		Height:       req.Height,
		Age:          req.Age,
		Token:        accessToken.Token,
		RefreshToken: accessToken.RefreshToken,
		ExpiresIn:    accessToken.ExpiresInMs,
	}, nil
}

func (uc *authUsecase) SignOut(ctx context.Context, sessionId string) error {
	if err := uc.authRepo.RevokeSessionById(ctx, sessionId); err != nil {
		if err != pgx.ErrNoRows {
			return err
		}
	}

	return nil
}

func (uc *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	session, err := uc.authRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrExpiredRefreshToken
		}
		return nil, err
	}

	err = uc.authRepo.RevokeSessionById(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	accessToken, err := uc.createSessionToken(ctx, session.Kind, session.UserAgent, session.AccountID)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		Token:        accessToken.Token,
		RefreshToken: accessToken.RefreshToken,
		ExpiresIn:    accessToken.ExpiresInMs,
	}, nil
}

func (uc *authUsecase) createSessionToken(ctx context.Context, kind, userAgent string, accountId *string) (*AccessToken, error) {
	// create session with refresh token
	session, err := NewSession(&uc.cfg.Auth, userAgent, accountId)
	if err != nil {
		return nil, err
	}

	var sessionId string
	var userId *string
	if kind == "guest" || accountId == nil {
		sessionId, err = uc.authRepo.CreateGuestSession(ctx, session)
		if err != nil {
			return nil, err
		}
	} else {
		userId, err = uc.userRepo.GetIdByAccountId(ctx, *accountId)
		if err != nil {
			return nil, err
		}

		sessionId, err = uc.authRepo.CreateUserSession(ctx, session)
		if err != nil {
			return nil, err
		}
	}

	accessToken, exp, err := security.NewAccessToken(uc.cfg.Auth.JWTSecret, uc.cfg.Auth.JWTAccessTTL, sessionId, kind, accountId, userId)
	if err != nil {
		return nil, err
	}

	return &AccessToken{
		Token:        accessToken,
		RefreshToken: session.RefreshTokenHash,
		ExpiresInMs:  time.Until(exp).Milliseconds(),
	}, nil
}
