package training

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainingRepository interface {
	GetById(ctx context.Context, id string) (*Training, error)
	GetLastByUserId(ctx context.Context, userID string) (*TrainingSession, error)
	// GetList(ctx context.Context, req *TrainingListRequest) ([]*Training, int, error)
}

type trainingRepository struct{ db *pgxpool.Pool }

func NewTrainingRepositry(db *pgxpool.Pool) TrainingRepository { return &trainingRepository{db: db} }

func (r *trainingRepository) GetById(ctx context.Context, id string) (*Training, error) {
	const q = `
		SELECT
			t.id, t.level, t.name, t.descriptions, t.time_label,
			t.calories_kcal, t.thumbnail_url, t.video_url, t.content_html,
			tc.code
		FROM trainings t
		LEFT JOIN training_categories tc ON t.category_id = tc.id
		WHERE t.id = $1
		LIMIT 1
	`

	var training Training
	err := r.db.QueryRow(ctx, q, id).Scan(
		&training.ID,
		&training.CategoryCode,
		&training.Level,
		&training.Name,
		&training.Descriptions,
		&training.TimeLabel,
		&training.CaloriesKcal,
		&training.ThumbnailURL,
		&training.VideoURL,
		&training.ContentHTML,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &training, nil
}

func (r *trainingRepository) GetLastByUserId(ctx context.Context, userID string) (*TrainingSession, error) {
	const q = `
		SELECT
			id, user_id, training_id, distance, time, pace, created_at
		FROM training_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var trainingSession TrainingSession
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&trainingSession.ID,
		&trainingSession.UserID,
		&trainingSession.TrainingID,
		&trainingSession.Distance,
		&trainingSession.Time,
		&trainingSession.Pace,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &trainingSession, nil
}
