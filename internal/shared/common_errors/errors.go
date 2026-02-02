package common_errors

import "errors"

var (
	ErrInvalidURL     = errors.New("url in wrong format")
	ErrCollision      = errors.New("could not find free space to store url")
	ErrNotFound       = errors.New("url not found")
	ErrEmptyHost      = errors.New("empty host")
	ErrNotValidScheme = errors.New("not valid scheme")
	ErrEmptyURL       = errors.New("empty url")
)
