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
	GetTrainings(ctx context.Context, query *TrainingsQuery) (trainingItems []TrainingItemResponse, totalPages int, err error)
	CreateTraining(ctx context.Context, req *TrainingRequest) (*TrainingResponse, error)
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
	training, err := uc.trainingRepo.GetLastByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	if training == nil {
		return nil, ErrTrainingSessionNotFound
	}

	return (*TrainingSessionResponse)(training), nil
}

func (u *trainingUsecase) GetTrainings(ctx context.Context, query *TrainingsQuery) (trainingItems []TrainingItemResponse, totalPages int, err error) {
	trainings, total, err := u.trainingRepo.GetList(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	if len(trainings) == 0 {
		return nil, 0, ErrTrainingNotFound
	}

	for _, training := range trainings {
		trainingItems = append(trainingItems, TrainingItemResponse{
			ID:           training.ID,
			Level:        training.Level,
			Name:         training.Name,
			Descriptions: training.Descriptions,
			ThumbnailURL: training.ThumbnailURL,
		})
	}

	totalPages = 0
	if total > 0 {
		totalPages = (total + query.Limit - 1) / query.Limit
	}

	return trainingItems, totalPages, nil
}

func (u *trainingUsecase) CreateTraining(ctx context.Context, req *TrainingRequest) (*TrainingResponse, error) {
	training, err := u.trainingRepo.Create(ctx, &Training{
		CategoryCode: req.CategoryCode,
		Level:        req.Level,
		Name:         req.Name,
		Descriptions: req.Descriptions,
		TimeLabel:    req.Time,
		CaloriesKcal: req.Calories,
		ThumbnailURL: req.ThumbnailURL,
		VideoURL:     &req.VideoURL,
		ContentHTML:  req.Content,
	})
	if err != nil {
		return nil, err
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
