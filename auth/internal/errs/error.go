package errs

import (
	"fmt"
)

type ErrorType string
type ErrorCode string

const (
	Invalid               ErrorType = "INVALID"
	Conflict              ErrorType = "CONFLICT"
	NotFound              ErrorType = "NOT_FOUND"
	AuthenticationFailure ErrorType = "AUTHENTICATION_FAILURE"
	Incorrect             ErrorType = "INCORRECT"
)

type Error struct {
	Type ErrorType
	Code ErrorCode
	Err  error
}

func NewError(errType ErrorType, err error, code ErrorCode) *Error {
	return &Error{
		Type: errType,
		Code: code,
		Err:  err,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:[%s]", e.Type, e.Code)
}

func (e *Error) Unwrap() error {
	return e.Err
}
