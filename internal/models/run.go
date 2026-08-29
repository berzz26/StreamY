package models

import "time"

type Run struct {
	ID           string `db:"id"`
	VideoID      string `db:"video_id"`
	Status       string `db:"status"`
	ErrorMessage string `db:"error_message"`

	ProbeMS     int64 `db:"probe_ms"`
	TranscodeMS int64 `db:"transcode_ms"`
	UploadMS    int64 `db:"upload_ms"`
	DbMS        int64 `db:"db_ms"`
	TotalMS     int64 `db:"total_ms"`

	RenditionsCount int              `db:"renditions_count"`
	RenditionTimes  map[string]int64 `db:"rendition_times"`

	OriginalSize    int64   `db:"original_size"`
	DurationSeconds float64 `db:"duration_seconds"`

	StartedAt  time.Time `db:"started_at"`
	FinishedAt time.Time `db:"finished_at"`
	CreatedAt  time.Time `db:"created_at"`
}