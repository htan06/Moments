package domain

import "github.com/htan06/Moments/auth/internal/errs"

const (
	UsernameInvalid    errs.ErrorCode = "USERNAME_INVALID"
	EmailInvalid       errs.ErrorCode = "EMAIL_INVALID"
	PhoneNumberInvalid errs.ErrorCode = "PHONE_NUMBER_INVALID"
	NameInvalid        errs.ErrorCode = "NAME_INVALID"
	PasswordHashEmpty  errs.ErrorCode = "PASSWORD_HASH_EMPTY"

	UsernameAlreadyUsed       errs.ErrorCode = "USERNAME_ALREADY_USED"
	EmailAlreadyUsed          errs.ErrorCode = "EMAIL_ALREADY_USED"
	PhoneNumberAlreadyUsed    errs.ErrorCode = "PHONE_NUMBER_ALREADY_USED"
	UserNotFound              errs.ErrorCode = "USER_NOT_FOUND"
	OTPIncorrect              errs.ErrorCode = "OTP_INCORRCET"
	UsernameOrPasswordInvalid errs.ErrorCode = "USERNAME_OR_PASSWORD_INVALID"
	UserNonActiveErr          errs.ErrorCode = "USER_NON_ACTIVE"
	TokenInvalid              errs.ErrorCode = "TOKEN_INVALID"
	PasswordTooShort          errs.ErrorCode = "PASSWORD_TOO_SHORT"
)
