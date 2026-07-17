package service

import "fmt"

// ValidationError marks an error caused by invalid user input. The API layer
// maps it to HTTP 400 and surfaces its message to the client; every other
// (unexpected) error is treated as internal and hidden behind a generic 500.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

func validationErrorf(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}
