package training

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrorTrainingExists = errors.New("training already exists")
)

type TrainingRepository interface {
	GetById(ctx context.Context, id string) (*Training, error)
	GetLastByUserId(ctx context.Context, userID string) (*TrainingSession, error)
	GetList(ctx context.Context, query *TrainingsQuery) ([]*TrainingItem, int, error)
	Create(ctx context.Context, training *Training) (*Training, error)
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
		args   []any
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

func (r *trainingRepository) Create(ctx context.Context, training *Training) (*Training, error) {
	const q = `
		WITH cat AS (
				SELECT id, code, name
				FROM training_categories
				WHERE code = $1
				LIMIT 1
		),
		ins AS (
				INSERT INTO trainings (
					category_id, level, name, descriptions, time_label,
					calories_kcal, thumbnail_url, video_url, content_html
				)
				SELECT
					cat.id, $2, $3, $4, $5, $6, $7, $8, $9
				FROM cat
				RETURNING
					id, category_id, level, name, descriptions,
					time_label, calories_kcal, thumbnail_url, video_url, content_html
		)
		SELECT
				ins.id,
				cat.code AS category_code,
				cat.name AS category_name,
				ins.level,
				ins.name,
				ins.descriptions,
				ins.time_label,
				ins.calories_kcal,
				ins.thumbnail_url,
				ins.video_url,
				ins.content_html
		FROM ins
		JOIN cat ON ins.category_id = cat.id;
		`

	err := r.db.QueryRow(ctx, q,
		training.CategoryCode,
		training.Level,
		training.Name,
		training.Descriptions,
		training.VideoURL,
		training.CaloriesKcal,
		training.ThumbnailURL,
		training.VideoURL,
		training.ContentHTML,
	).Scan(
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrorTrainingExists
		}

		return nil, err
	}

	return training, nil
}
