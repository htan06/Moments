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

const (
	UsernameAlreadyUsed    ErrorCode = "USERNAME_ALREADY_USED"
	EmailAlreadyUsed       ErrorCode = "EMAIL_ALREADY_USED"
	PhoneNumberAlreadyUsed ErrorCode = "PHONE_NUMBER_ALREADY_USED"

	UserNotFound          ErrorCode = "USER_NOT_FOUND"
	ReceiverNotFound      ErrorCode = "RECEIVER_NOT_FOUND"
	FriendRequestNotFound ErrorCode = "FRIEND_REQEST_NOT_FOUND"

	IncorrectOTP ErrorCode = "INCORRCET_OTP"

	InvalidUsernameOrPassword ErrorCode = "INVALID_USERNAME_OR_PASSWORD"
	UserNonActive             ErrorCode = "USER_NON_ACTIVE"
	TokenInvalid              ErrorCode = "TOKEN_INVALID"
	FriendRequestInvalid      ErrorCode = "FRIEND_REQUEST_INVALID"
	UsernameInvalid           ErrorCode = "USERNAME_INVALID"
	FollowInvalid             ErrorCode = "FOLLOW_INVALID"
)

type Error struct {
	Type  ErrorType
	Codes []ErrorCode
	Err   error
}

func NewError(errType ErrorType, err error, codes ...ErrorCode) *Error {
	return &Error{
		Type:  errType,
		Codes: codes,
		Err:   err,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:[%v]", e.Type, e.Codes)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) AddCode(code ErrorCode) {
	e.Codes = append(e.Codes, code)
}
