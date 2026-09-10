package errs

import "fmt"

const (
	UsernameInvalid    ErrorCode = "USERNAME_INVALID"
	EmailInvalid       ErrorCode = "EMAIL_INVALID"
	PhoneNumberInvalid ErrorCode = "PHONE_NUMBER_INVALID"
	NameInvalid        ErrorCode = "NAME_INVALID"
	PasswordHashEmpty  ErrorCode = "PASSWORD_HASH_EMPTY"

	UsernameAlreadyUsed       ErrorCode = "USERNAME_ALREADY_USED"
	EmailAlreadyUsed          ErrorCode = "EMAIL_ALREADY_USED"
	PhoneNumberAlreadyUsed    ErrorCode = "PHONE_NUMBER_ALREADY_USED"
	UserNotFound              ErrorCode = "USER_NOT_FOUND"
	OTPIncorrect              ErrorCode = "OTP_INCORRCET"
	UsernameOrPasswordInvalid ErrorCode = "USERNAME_OR_PASSWORD_INVALID"
	UserNonActiveErr          ErrorCode = "USER_NON_ACTIVE"
	TokenInvalid              ErrorCode = "TOKEN_INVALID"
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
