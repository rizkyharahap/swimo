package auth

import (
	"errors"
	"time"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/internal/user"
	"github.com/rizkyharahap/swimo/pkg/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCreds = errors.New("invalid email or passwords")
)

type Auth struct {
	AccountID    string
	Email        string
	PasswordHash string
	IsLocked     bool
	Name         string
	Gender       user.Gender
	WeightKG     float64
	HeightCM     float64
	AgeYears     int16
}

type Session struct {
	ID               string
	AccountID        *string
	Kind             string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
	UserAgent        string
	RevokedAt        *time.Time
}

type AccessToken struct {
	Token        string
	RefreshToken string
	ExpiresInMs  int64
}

func (u *Auth) ComparePassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCreds
	}

	return nil
}

func NewSession(cfg *config.AuthConfig, userAgent string, accountId *string) (*Session, error) {
	refreshToken, err := security.NewRefreshToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(cfg.JWTAccessTTL)
	refreshExpiresAt := now.Add(cfg.JWTRefreshTTL)

	return &Session{
		AccountID:        accountId,
		RefreshTokenHash: refreshToken,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		UserAgent:        userAgent,
	}, nil
}
