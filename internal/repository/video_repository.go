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
