package service

import (
	"errors"
	"fmt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s %s", e.Field, e.Message)
}
