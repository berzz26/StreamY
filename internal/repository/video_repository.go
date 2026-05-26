package repository

import (
	"context"

	"github.com/berzz26/StreamY/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoRepository struct {
	db *pgxpool.Pool
}

func NewVideoRepository(
	db *pgxpool.Pool,
) *VideoRepository {

	return &VideoRepository{
		db: db,
	}

}

func (r *VideoRepository) CreateVideo(video models.Video) error {
	query := `
	INSERT INTO videos(
	id,
	title,
	status,
	original_path,
	original_size
	)
	VALUES ($1,$2,$3,$4,$5)
	
	`
	_, err := r.db.Exec(
		context.Background(),

		query,

		video.ID,
		video.Title,
		video.Status,
		video.OriginalPath,
		video.OriginalSize,
	)

	return err
}

func (r *VideoRepository) ClaimNextVideo() (*models.Video, error) {

	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(context.Background())
	// skip locked means the worker that is processing the video will lock the row.
	//this prevents other workers (when running the transcoder in parallel ) to not consume the same video
	query := `
	SELECT id, title, status, original_path
	FROM videos
	WHERE status = $1
	ORDER BY created_at ASC
	FOR UPDATE SKIP LOCKED
	LIMIT 1
	`

	row := tx.QueryRow(
		context.Background(),
		query,
		models.StatusUploaded,
	)

	var video models.Video

	err = row.Scan(
		&video.ID,
		&video.Title,
		&video.Status,
		&video.OriginalPath,
	)

	if err != nil {
		return nil, err
	}

	updateQuery := `
	UPDATE videos
	SET status = $1,
		updated_at = NOW()
	WHERE id = $2
	`

	_, err = tx.Exec(
		context.Background(),
		updateQuery,
		models.StatusProcessing,
		video.ID,
	)

	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	video.Status = models.StatusProcessing

	return &video, nil
}
