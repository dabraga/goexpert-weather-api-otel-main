package domain

import "errors"

type ServiceError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *ServiceError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func NewServiceError(statusCode int, message string) *ServiceError {
	return &ServiceError{
		Err:        errors.New(message),
		StatusCode: statusCode,
		Message:    message,
	}
}

var (
	ErrInvalidZipcode = errors.New("invalid zipcode")
)
