package model

import "time"

type DBURL struct {
	ID       int
	FullURL  string
	ShortURL string
	UpdateAt time.Time
}
