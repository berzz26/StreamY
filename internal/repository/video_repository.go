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

func (r *VideoRepository) UpdateVideoStatus(
	videoID string,
	status string,
	processedPath string,
) error {

	query := `
	UPDATE videos
	SET status = $1,
		updated_at = NOW(),
		processed_path = $2
	WHERE id = $3
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		status,
		processedPath,
		videoID,
	)

	return err
}
func (r *VideoRepository) MarkVideoFailed(
	videoID string,
	errorMessage string,
) error {

	query := `
	UPDATE videos
	SET status = $1,
		error_message = $2,
		updated_at = NOW()
	WHERE id = $3
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		models.StatusFailed,
		errorMessage,
		videoID,
	)

	return err
}

func (r *VideoRepository) CreateVideoProbe(probe *models.VideoProbe) error {
	query := `
        INSERT INTO video_probe (
			id,
            video_id,
            format_name,
            format_long_name,
            format_bitrate,
            format_duration,
            video_codec,
            video_codec_long_name,
            video_profile,
            video_level,
            width,
            height,
            video_bitrate,
            frame_rate_num,
            frame_rate_den,
            pixel_format,
            color_space,
            color_transfer,
            color_primaries,
            sample_aspect_ratio,
            display_aspect_ratio,
            audio_codec,
            audio_codec_long_name,
            audio_bitrate,
            audio_sample_rate,
            audio_channels,
            audio_channel_layout,
            probed_at,
            created_at
        )
        VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21,
			$22, $23, $24, $25, $26,
			$27, $28, $29
        )
    `

	_, err := r.db.Exec(
		context.Background(),
		query,
		probe.ID,
		probe.VideoID,
		probe.FormatName,
		probe.FormatLongName,
		probe.FormatBitrate,
		probe.FormatDuration,
		probe.VideoCodec,
		probe.VideoCodecLongName,
		probe.VideoProfile,
		probe.VideoLevel,
		probe.Width,
		probe.Height,
		probe.VideoBitrate,
		probe.FrameRateNum,
		probe.FrameRateDen,
		probe.PixelFormat,
		probe.ColorSpace,
		probe.ColorTransfer,
		probe.ColorPrimaries,
		probe.SampleAspectRatio,
		probe.DisplayAspectRatio,
		probe.AudioCodec,
		probe.AudioCodecLongName,
		probe.AudioBitrate,
		probe.AudioSampleRate,
		probe.AudioChannels,
		probe.AudioChannelLayout,
		probe.ProbedAt,
		probe.CreatedAt,
	)

	return err
}

func (r *VideoRepository) UpdateVideoDuration(videoID string, duration float64) error {
	query := `
	UPDATE videos
	SET duration_seconds = $1,
		updated_at = NOW()
	WHERE id = $2
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		duration,
		videoID,
	)

	return err
}
