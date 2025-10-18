package training

import (
	"errors"
	"math"
)

var (
	ErrInvalidCreds = errors.New("invalid email or passwords")
)

type TrainingCategory struct {
	ID          string
	Code        string
	Name        string
	Description *string
	MET         float32
}

type Training struct {
	ID           string
	CategoryCode string
	CategoryName *string
	Level        string
	Name         string
	Descriptions string
	TimeLabel    string
	CaloriesKcal int
	ThumbnailURL string
	VideoURL     *string
	ContentHTML  string
}

type TrainingSession struct {
	ID              string
	UserID          string
	TrainingID      string
	DistanceMeters  int
	DurationSeconds int
	Pace            float64
	CaloriesKcal    int
}

type TrainingItem struct {
	ID           string
	Level        string
	Name         string
	Descriptions string
	TimeLabel    string
	ThumbnailURL string
}

func NewTrainingSession(userID string, trainingID string, distanceMeters int, durationSeconds int, bmr float64, met float32) *TrainingSession {
	durationSecondsFloat := float64(durationSeconds)
	paceMinPer100m := (durationSecondsFloat / float64(distanceMeters)) * (100.0 / 60.0)
	durationHours := durationSecondsFloat / 3600.0

	return &TrainingSession{
		UserID:          userID,
		TrainingID:      trainingID,
		DistanceMeters:  distanceMeters,
		DurationSeconds: durationSeconds,
		Pace:            paceMinPer100m,
		CaloriesKcal:    calculateCalories(bmr, float64(met), durationHours),
	}
}

func calculateCalories(bmr float64, met float64, durationHours float64) int {
	bmrPerHour := bmr / 24.0
	calories := met * bmrPerHour * durationHours

	return int(math.Round(calories))
}
