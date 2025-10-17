package training

import (
	"context"
	"errors"
)

var (
	ErrTrainingNotFound        = errors.New("training not found")
	ErrTrainingSessionNotFound = errors.New("no training sessions found")
)

type TrainingUsecase interface {
	GetById(ctx context.Context, id string) (*TrainingResponse, error)
	GetLastTraining(ctx context.Context, userId string) (*TrainingSessionResponse, error)
	// GetList(ctx context.Context, req *TrainingListRequest) ([]TrainingItemResponse, int, error)
}

type trainingUsecase struct {
	trainingRepo TrainingRepository
}

func NewTrainingUsecase(trainingRepo TrainingRepository) TrainingUsecase {
	return &trainingUsecase{trainingRepo}
}

func (u *trainingUsecase) GetById(ctx context.Context, id string) (*TrainingResponse, error) {
	training, err := u.trainingRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	if training == nil {
		return nil, ErrTrainingNotFound
	}

	return &TrainingResponse{
		ID:           training.ID,
		Level:        training.Level,
		Name:         training.Name,
		Descriptions: training.Descriptions,
		Time:         training.TimeLabel,
		Calories:     training.CaloriesKcal,
		ThumbnailURL: training.ThumbnailURL,
		VideoURL:     training.VideoURL,
		Content:      training.ContentHTML,
		CategoryCode: training.CategoryCode,
	}, nil
}

func (uc *trainingUsecase) GetLastTraining(ctx context.Context, userId string) (*TrainingSessionResponse, error) {
	session, err := uc.trainingRepo.GetLastByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, ErrTrainingSessionNotFound
	}

	return (*TrainingSessionResponse)(session), nil
}
