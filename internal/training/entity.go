package training

import (
	"errors"
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
	CategoryName string
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
	ID         string
	UserID     string
	TrainingID string
	Distance   int
	Time       int
	Pace       float64
}

type TrainingItem struct {
	ID           string
	Level        string
	Name         string
	Descriptions string
	TimeLabel    string
	ThumbnailURL string
}
