package http_error

import "fmt"

type HTTPError struct {
	StatusCode int
	Message    string
}

func New(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%d:%s", e.StatusCode, e.Message)
}
