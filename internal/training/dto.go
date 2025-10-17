package training

import (
	"strings"

	"github.com/rizkyharahap/swimo/pkg/validator"
)

type CreateTrainingRequest struct {
	CategoryCode string  `json:"categoryCode" example:"BREASTSTROKE"`
	Level        string  `json:"level" example:"beginner"`
	Name         string  `json:"name" example:"Breaststroke Basics"`
	Descriptions string  `json:"descriptions" example:"Dasar gaya dada untuk pemula"`
	Time         string  `json:"time" example:"10-15 min"`                                               // required (maps to time_label)
	Calories     int     `json:"calories" example:"120"`                                                 // required (kcal) - positive integer
	ThumbnailURL string  `json:"thumbnailUrl" example:"https://cdn.example.com/thumbs/breaststroke.png"` // required, valid URL
	VideoURL     *string `json:"videoUrl" example:"https://cdn.example.com/videos/breaststroke.mp4"`     // required, valid URL
	Content      string  `json:"content" example:"<p>HTML content here</p>"`                             // required (HTML string)
}

type TrainingResponse struct {
	ID           string  `json:"id" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	CategoryCode string  `json:"categoryCode" example:"BREASTSTROKE"`
	Level        string  `json:"level" example:"beginner"`
	Name         string  `json:"name" example:"Breaststroke Basics"`
	Descriptions string  `json:"descriptions" example:"Short description about this training"`
	Time         string  `json:"time" example:"10-15 min"`
	Calories     int     `json:"calories" example:"120"`
	ThumbnailURL string  `json:"thumbnailUrl" example:"https://cdn.example.com/thumbs/breaststroke.png"`
	VideoURL     *string `json:"videoUrl" example:"https://cdn.example.com/videos/breaststroke.mp4"`
	Content      string  `json:"content" example:"<p>HTML content here</p>"`
}

type TrainingSessionResponse struct {
	ID         string  `json:"id" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	UserID     string  `json:"userId" example:"a1b2c3d4-e5f6-7890-1234-567890abcdef"`
	TrainingID *string `json:"trainingId,omitempty" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	Distance   int     `json:"distance" example:"1500"`
	Time       int     `json:"time" example:"1800"`
	Pace       float64 `json:"pace" example:"1.2"`
}

type TrainingItemResponse struct {
	ID           string `json:"id" example:"8c4a2d27-56e2-4ef3-8a6e-43b812345abc"`
	Level        string `json:"level" example:"beginner"`
	Name         string `json:"name" example:"Breaststroke Basics"`
	Descriptions string `json:"descriptions" example:"Short description about this training"`
	ThumbnailURL string `json:"thumbnailUrl" example:"https://cdn.example.com/thumbs/breaststroke.png"`
}

func (r *CreateTrainingRequest) Validate() error {
	errors := make(map[string]string)

	// CategoryCode
	if strings.TrimSpace(r.CategoryCode) == "" {
		errors["categoryCode"] = "CategoryCode is required"
	}

	// Level
	if strings.TrimSpace(r.Level) == "" {
		errors["level"] = "Level is required"
	} else if len(r.Level) > 50 {
		errors["level"] = "Level must not exceed 50 characters"
	}

	// Name
	if strings.TrimSpace(r.Name) == "" {
		errors["name"] = "Name is required"
	} else if len(r.Name) > 100 {
		errors["name"] = "Name must not exceed 100 characters"
	}

	// Descriptions
	if strings.TrimSpace(r.Descriptions) == "" {
		errors["descriptions"] = "Descriptions is required"
	}

	// Time (required, simple non-empty)
	if strings.TrimSpace(r.Time) == "" {
		errors["time"] = "Time is required"
	}

	// Calories (required, positive)
	if r.Calories <= 0 {
		errors["calories"] = "Calories must be a positive integer"
	}

	// ThumbnailURL (required + URL format)
	if strings.TrimSpace(r.ThumbnailURL) == "" {
		errors["thumbnailUrl"] = "ThumbnailURL is required"
	} else if !validator.URLPattern.MatchString(strings.TrimSpace(r.ThumbnailURL)) {
		errors["thumbnailUrl"] = "ThumbnailURL is not a valid URL"
	}

	// VideoURL (required + URL format)
	if r.VideoURL == nil || strings.TrimSpace(*r.VideoURL) == "" {
		errors["videoUrl"] = "VideoURL is required"
	} else if !validator.URLPattern.MatchString(strings.TrimSpace(*r.VideoURL)) {
		errors["videoUrl"] = "VideoURL is not a valid URL"
	}

	// Content (required)
	if strings.TrimSpace(r.Content) == "" {
		errors["content"] = "Content is required"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
