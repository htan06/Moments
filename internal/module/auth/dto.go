package auth

import "github.com/htan06/echo-messenger-rest-api/internal/module/auth/usecase"

type RegisterReq struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r RegisterReq) ToRegisterCmd() usecase.RegisterCmd {
	return usecase.RegisterCmd{
		Email:    r.Email,
		Password: r.Password,
		Name:     r.Name,
		Username: r.Username,
	}
}

type VerifyRegisterOTPReq struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (r VerifyRegisterOTPReq) ToVerifyRegisterOTPCmd() usecase.VerifyRegisterOTPCmd {
	return usecase.VerifyRegisterOTPCmd{
		Email: r.Email,
		OTP:   r.OTP,
	}
}

type LoginPasswordReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginPasswordReq) ToLoginPasswordCmd() usecase.LoginPasswordCmd {
	return usecase.LoginPasswordCmd{
		Email:    r.Email,
		Password: r.Password,
	}
}

type ChangePasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
