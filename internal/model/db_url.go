package model

import (
	"time"
)

type DBURL struct {
	ID           int        `csv:"id"`
	OriginalURL  string     `csv:"original_url"`
	ShortURL     string     `csv:"short_url"`
	CreatedAt    time.Time  `csv:"created_at"`
	UserID       int64      `csv:"user_id"`
	DeletedAtUTC *time.Time `csv:"deleted_at_utc"`
}
