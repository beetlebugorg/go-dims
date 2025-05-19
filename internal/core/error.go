package core

import "fmt"

type StatusError struct {
	StatusCode int
	Message    string
}

func NewStatusError(statusCode int, message string) *StatusError {
	return &StatusError{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("Error: %s (status: %d)", e.Message, e.StatusCode)
}
