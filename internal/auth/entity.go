package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCreds = errors.New("invalid email or passwords")
)

type User struct {
	ID        string
	AccountID string
	Name      string
	WeightKG  *float64
	HeightCM  *float64
	AgeYears  *int16
}

type Auth struct {
	AccountID    string
	Email        string
	PasswordHash string
	IsLocked     bool
	Name         string
	WeightKG     *float64
	HeightCM     *float64
	AgeYears     *int16
}

func (u *Auth) ComparePassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCreds
	}

	return nil
}
