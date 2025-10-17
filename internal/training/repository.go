package training

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainingRepository interface {
	GetById(ctx context.Context, id string) (*Training, error)
	GetLastByUserId(ctx context.Context, userID string) (*TrainingSession, error)
	GetList(ctx context.Context, query *TrainingsQuery) ([]*TrainingItem, int, error)
}

type trainingRepository struct{ db *pgxpool.Pool }

func NewTrainingRepositry(db *pgxpool.Pool) TrainingRepository { return &trainingRepository{db: db} }

func (r *trainingRepository) GetById(ctx context.Context, id string) (*Training, error) {
	const q = `
		SELECT
			t.id, tc.code, tc.name,
			t.level, t.name, t.descriptions, t.time_label,
			t.calories_kcal, t.thumbnail_url, t.video_url, t.content_html
		FROM trainings t
		LEFT JOIN training_categories tc ON t.category_id = tc.id
		WHERE t.id = $1
		LIMIT 1
	`

	var training Training
	err := r.db.QueryRow(ctx, q, id).Scan(
		&training.ID,
		&training.CategoryCode,
		&training.CategoryName,
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

func (r *trainingRepository) GetList(ctx context.Context, query *TrainingsQuery) ([]*TrainingItem, int, error) {
	var (
		whereQ string
		args   []interface{}
		baseQ  = `
		SELECT
			id, level, name, descriptions, time_label, thumbnail_url
		FROM trainings
	`
		countQ = `SELECT COUNT(*) FROM trainings`
	)

	// Filter (search)
	if query.Search != "" {
		whereQ = ` WHERE (name ILIKE $1 OR descriptions ILIKE $1 OR level ILIKE $1)`
		args = append(args, "%"+query.Search+"%")
	}

	// Order by
	orderMap := map[string]string{
		"name.asc":        " ORDER BY name ASC",
		"name.desc":       " ORDER BY name DESC",
		"level.asc":       " ORDER BY level ASC",
		"level.desc":      " ORDER BY level DESC",
		"created_at.asc":  " ORDER BY created_at ASC",
		"created_at.desc": " ORDER BY created_at DESC",
	}
	orderByQ := orderMap[query.Sort]
	if orderByQ == "" {
		orderByQ = " ORDER BY created_at DESC"
	}

	// Pagination
	offset := (query.Page - 1) * query.Limit
	finalQ := fmt.Sprintf("%s%s%s LIMIT $%d OFFSET $%d",
		baseQ, whereQ, orderByQ,
		len(args)+1, len(args)+2,
	)

	rows, err := r.db.Query(ctx, finalQ, append(args, query.Limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	trainings := make([]*TrainingItem, 0, query.Limit)
	for rows.Next() {
		var t TrainingItem
		if err := rows.Scan(
			&t.ID,
			&t.Level,
			&t.Name,
			&t.Descriptions,
			&t.TimeLabel,
			&t.ThumbnailURL,
		); err != nil {
			return nil, 0, err
		}

		trainings = append(trainings, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(trainings) == 0 {
		return nil, 0, nil
	}

	var total int
	if len(args) > 0 {
		err = r.db.QueryRow(ctx, countQ+whereQ, args...).Scan(&total)
	} else {
		err = r.db.QueryRow(ctx, countQ).Scan(&total)
	}

	if err != nil {
		return nil, 0, err
	}

	return trainings, total, nil
}
