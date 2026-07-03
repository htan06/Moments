package auth

type requireOTPReq struct {
	Email string `json:"email"`
}

type VerifyOTPReq struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type TokenType string

const (
	LoginType    TokenType = "LOGIN"
	RegisterType TokenType = "REGISTER"
)

type VerifyOTPResp struct {
	Type          TokenType `json:"type"`
	RegisterToken string    `json:"register_token,omitempty"`
	AccessToken   string    `json:"access_token,omitempty"`
	RefreshToken  string    `json:"refresh_token,omitempty"`
}

type RegisterUserReq struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Username      string `json:"username"`
	PhoneNumber   string `json:"phone_number"`
	RegisterToken string `json:"register_token"`
}
