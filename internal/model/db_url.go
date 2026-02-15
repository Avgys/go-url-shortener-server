package model

import "time"

type DbURL struct {
	Id       int
	FullURL  string
	ShortURL string
	UpdateAt time.Time
}
