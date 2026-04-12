package errs

import "errors"

var (
	ErrForbidden          = errors.New("forbidden")
	ErrEmailTaken         = errors.New("email taken")
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
