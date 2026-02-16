package repository

import (
	"errors"
)

var (
	ErrCollision = errors.New("slot in dictionary taken")
	ErrNotFound  = errors.New("url not found in store")
)
