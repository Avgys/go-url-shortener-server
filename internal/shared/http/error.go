package httphelper

// ShowHTTPError is an error with an associated HTTP status code for [HandleErr].
type ShowHTTPError struct {
	text       string
	StatusCode int
}

// NewError wraps text and statusCode so callers can return typed HTTP errors.
func NewError(text string, statusCode int) error {
	return &ShowHTTPError{text, statusCode}
}

// Error returns the client-facing error message.
func (e *ShowHTTPError) Error() string {
	return e.text
}
