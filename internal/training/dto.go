package training

import (
	"strings"

	"github.com/rizkyharahap/swimo/pkg/validator"
)

type TrainingRequest struct {
	CategoryCode string `json:"categoryCode" example:"BREASTSTROKE"`
	Level        string `json:"level" example:"beginner"`
	Name         string `json:"name" example:"Breaststroke Basics"`
	Descriptions string `json:"descriptions" example:"Dasar gaya dada untuk pemula"`
	TimeLabel    string `json:"time" example:"10-15 min"`
	CaloriesKcal int    `json:"caloriesKcal" example:"120"`
	ThumbnailURL string `json:"thumbnailUrl" example:"https://cdn.example.com/thumbs/breaststroke.png"`
	VideoURL     string `json:"videoUrl" example:"https://cdn.example.com/videos/breaststroke.mp4"`
	Content      string `json:"content" example:"<p>HTML content here</p>"`
}

type TrainingResponse struct {
	ID           string  `json:"id" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	CategoryCode string  `json:"categoryCode" example:"BREASTSTROKE"`
	CategoryName string  `json:"categoryName" example:"Breaststroke"`
	Level        string  `json:"level" example:"beginner"`
	Name         string  `json:"name" example:"Breaststroke Basics"`
	Descriptions string  `json:"descriptions" example:"Short description about this training"`
	TimeLabel    string  `json:"timeLabel" example:"10-15 min"`
	CaloriesKcal int     `json:"caloriesKcal" example:"120"`
	ThumbnailURL string  `json:"thumbnailUrl" example:"https://cdn.example.com/thumbs/breaststroke.png"`
	VideoURL     *string `json:"videoUrl" example:"https://cdn.example.com/videos/breaststroke.mp4"`
	ContentHTML  string  `json:"content" example:"<p>HTML content here</p>"`
}

type TrainingSessionResponse struct {
	ID              string  `json:"id" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	UserID          string  `json:"userId" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	TrainingID      string  `json:"trainingId" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	DistanceMeters  int     `json:"distanceMeters" example:"1500"`
	DurationSeconds int     `json:"durationSeconds" example:"1800"`
	Pace            float64 `json:"pace" example:"1.2"`
	CaloriesKcal    int     `json:"caloriesKcal" example:"120"`
}

type TrainingItemResponse struct {
	ID           string `json:"id" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	Level        string `json:"level" example:"beginner"`
	Name         string `json:"name" example:"Breaststroke Basics"`
	Descriptions string `json:"descriptions" example:"Short description about this training"`
	ThumbnailURL string `json:"thumbnailUrl" example:"https://cdn.example.com/thumbs/breaststroke.png"`
}

type TrainingsQuery struct {
	Page   int    `query:"page" validate:"min=1"`
	Limit  int    `query:"limit" validate:"min=1,max=100"`
	Sort   string `query:"sort" validate:"oneof=name.asc name.desc level.asc level.desc created_at.asc created_at.desc"`
	Search string `query:"search"`
}

type TrainingFinishSessionRequest struct {
	DistanceMeters  int `json:"distanceMeters" example:"300"`
	DurationSeconds int `json:"durationSeconds" example:"50"`
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func (q *TrainingsQuery) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	if q.Page < 1 {
		errors["page"] = "Page must be at least 1"
	}

	if q.Limit < 1 {
		errors["limit"] = "Limit must be at least 1"
	} else if q.Limit > 100 {
		errors["limit"] = "Limit must not exceed 100"
	}

	validSorts := map[string]bool{
		"name.asc": true, "name.desc": true,
		"level.asc": true, "level.desc": true,
		"created_at.asc": true, "created_at.desc": true,
	}
	if q.Sort != "" && !validSorts[q.Sort] {
		errors["sort"] = "Sort must be one of: name.asc, name.desc, level.asc, level.desc, created_at.asc, created_at.desc"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}

func (r *TrainingRequest) Validate() error {
	errors := make(map[string]string)

	r.CategoryCode = trim(r.CategoryCode)
	if r.CategoryCode == "" {
		errors["categoryCode"] = "CategoryCode is required"
	}

	r.Level = trim(r.Level)
	if r.Level == "" {
		errors["level"] = "Level is required"
	} else if len(r.Level) > 50 {
		errors["level"] = "Level must not exceed 50 characters"
	}

	r.Name = trim(r.Name)
	if r.Name == "" {
		errors["name"] = "Name is required"
	} else if len(r.Name) > 100 {
		errors["name"] = "Name must not exceed 100 characters"
	}

	r.Descriptions = trim(r.Descriptions)
	if r.Descriptions == "" {
		errors["descriptions"] = "Descriptions is required"
	}

	r.TimeLabel = trim(r.TimeLabel)
	if r.TimeLabel == "" {
		errors["timeLabel"] = "TimeLabel is required"
	}

	if r.CaloriesKcal <= 0 {
		errors["caloriesKcal"] = "CaloriesKcal must be a positive integer"
	}

	r.ThumbnailURL = trim(r.ThumbnailURL)
	if r.ThumbnailURL == "" {
		errors["thumbnailUrl"] = "ThumbnailURL is required"
	} else if !validator.IsValidURL(r.ThumbnailURL) {
		errors["thumbnailUrl"] = "ThumbnailURL is not a valid URL"
	}

	r.VideoURL = trim(r.VideoURL)
	if r.VideoURL != "" && !validator.IsValidURL(r.VideoURL) {
		errors["videoUrl"] = "VideoURL is not a valid URL"
	}

	r.Content = trim(r.Content)
	if r.Content == "" {
		errors["content"] = "Content is required"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}

func (r *TrainingFinishSessionRequest) Validate() error {
	errors := make(map[string]string)

	if r.DistanceMeters <= 0 {
		errors["distanceMeters"] = "DistanceMeteres must be a positive integer"
	}

	if r.DurationSeconds <= 0 {
		errors["timeLabel"] = "TimeLabel must be a positive integer"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
