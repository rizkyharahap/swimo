package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/rizkyharahap/swimo/config"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCreds = errors.New("invalid email or passwords")
)

type User struct {
	ID        string
	AccountID string
	Name      string
	WeightKG  float64
	HeightCM  float64
	AgeYears  int16
}

type Auth struct {
	AccountID    string
	Email        string
	PasswordHash string
	IsLocked     bool
	Name         string
	WeightKG     float64
	HeightCM     float64
	AgeYears     int16
}

type Session struct {
	AccountID        string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
	UserAgent        string
}

type Claims struct {
	Kind      string
	SessionID string
	Sub       string
	IssuedAt  int64
	ExpiresAt int64
}

func (u *Auth) ComparePassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCreds
	}

	return nil
}

func NewRefreshToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	refreshTokenHash := sha256.Sum256(b)

	return hex.EncodeToString(refreshTokenHash[:]), nil
}

// signJWT builds and signs a JWT string (header.payload.signature)
func signJWT(claims *Claims, secret string) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(claims)

	headerEnc := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadJSON)

	data := headerEnc + "." + payloadEnc
	signature := signHMACSHA256(data, secret)
	return data + "." + signature, nil
}

// signHMACSHA256 returns the Base64URL-encoded HMAC-SHA256 signature
func signHMACSHA256(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	sig := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sig)
}

func NewAccessToken(secret string, kind string, accountID, sessionID string, ttl time.Duration) (token string, exp time.Time, err error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := Claims{
		Kind:      kind,
		SessionID: sessionID,
		Sub:       accountID,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  now.Unix(),
	}

	token, err = signJWT(&claims, secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func NewSession(cfg *config.AuthConfig, userAgent string, accountId string) (*Session, error) {
	refreshToken, err := NewRefreshToken(32)
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
