package logger

type HttpError struct {
	text       string
	StatusCode int
}

func NewError(text string, statusCode int) error {
	return &HttpError{text, statusCode}
}

func (e *HttpError) Error() string {
	return e.text
}
