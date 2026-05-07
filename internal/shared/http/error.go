package http

type ShowHTTPError struct {
	text       string
	StatusCode int
}

func NewError(text string, statusCode int) error {
	return &ShowHTTPError{text, statusCode}
}

func (e *ShowHTTPError) Error() string {
	return e.text
}
