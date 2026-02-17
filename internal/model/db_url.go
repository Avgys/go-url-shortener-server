package model

import "time"

type DBURL struct {
	Id       int
	FullURL  string
	ShortURL string
	UpdateAt time.Time
}
