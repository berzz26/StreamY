package models

import "time"

type Video struct {
	ID              string
	Title           string
	Status          string
	OriginalPath    string
	DurationSeconds float64
	ProcessedPath   string
	ErrorMessage    string
	OriginalSize    int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
