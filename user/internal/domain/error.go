package domain

import (
	"errors"
)

var (
	ErrUserNotFound    error = errors.New("USER_NOT_FOUND")
	ErrUserIDInvalid   error = errors.New("USER_ID_INVALID")
	ErrUsernameInvalid error = errors.New("USERNAME_INVALID")
	ErrNameInvalid     error = errors.New("NAME_INVALID")
	ErrNegativeCount   error = errors.New("NEGATIVE_COUNT")
	ErrTokenInvalid    error = errors.New("TOKEN_INVALID")
)
