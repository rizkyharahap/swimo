package user

import (
	"errors"
)

var ErrGenderInvalid = errors.New("invalid gender")

type Gender uint8

const (
	Male   Gender = iota // 0
	Female               // 1
)

func (g Gender) String() (string, error) {
	switch g {
	case Male:
		return "male", nil
	case Female:
		return "female", nil
	default:
		return "", ErrGenderInvalid
	}
}

func ParseGender(s string) (Gender, error) {
	switch s {
	case "male":
		return Male, nil
	case "female":
		return Female, nil
	default:
		return 0, ErrGenderInvalid
	}
}

type User struct {
	ID        string
	AccountID string
	Name      string
	Gender    Gender
	WeightKG  float64
	HeightCM  float64
	AgeYears  int16
}

func (u *User) GetBMR() float64 {
	var bmr float64

	if u.Gender == Male {
		bmr = 88.362 + (13.397 * u.WeightKG) + (4.799 * u.HeightCM) - (5.677 * float64(u.AgeYears))
	} else {
		bmr = 447.593 + (9.247 * u.WeightKG) + (3.098 * u.HeightCM) - (4.330 * float64(u.AgeYears))
	}

	return bmr
}
