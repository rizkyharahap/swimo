package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken     = errors.New("invalid token format")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrExpiredToken     = errors.New("token expired")
)

type Claim struct {
	Sub  string
	Aid  *string
	Uid  *string
	Kind string
	Iat  int64
	Exp  int64
}

func NewAccessToken(secret string, ttl time.Duration, sessionId string, kind string, accountId, userId *string) (token string, exp time.Time, err error) {
	now := time.Now()
	exp = now.Add(ttl)

	claims := Claim{
		Sub:  sessionId,
		Aid:  accountId,
		Uid:  userId,
		Kind: kind,
		Iat:  now.Unix(),
		Exp:  exp.Unix(),
	}

	token, err = signJWT(&claims, secret)
	return token, exp, err
}

func NewRefreshToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

func VerifyJWT(token, secret string) (*Claim, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	data := parts[0] + "." + parts[1]
	expectedSig := signHMACSHA256(data, secret)

	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return nil, ErrInvalidSignature
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claim
	if err := json.Unmarshal(payloadBytes, &claims); err != nil || claims.Sub == "" {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > claims.Exp {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

// signJWT builds and signs a JWT string (header.payload.signature)
func signJWT(claims *Claim, secret string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payloadEnc := base64.RawURLEncoding.EncodeToString(payload)
	data := header + "." + payloadEnc
	signature := signHMACSHA256(data, secret)

	return data + "." + signature, nil
}

// signHMACSHA256 returns the Base64URL-encoded HMAC-SHA256 signature
func signHMACSHA256(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
