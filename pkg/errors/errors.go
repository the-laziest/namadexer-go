package errors

import (
	"errors"
	"fmt"
	"strings"
)

type Error struct {
	Err     error
	Caused  error
	Message string
}

var (
	Create = errors.New
	As     = errors.As
	Is     = errors.Is
	Unwrap = errors.Unwrap
)

func New(err error, messages ...string) error {
	if err == nil {
		return nil
	}
	message := strings.Join(messages, ": ")
	return &Error{
		Err:     fmt.Errorf("%s: %w", message, err),
		Caused:  err,
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Caused
}
