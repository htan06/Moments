package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/model"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
)

type AuthenticationService struct {
	userRepo       UserReposiotry
	cacheRepo      CacheRepository
	emailOTPSender EmailOTPSender
	jwtProvider    *security.JWTProvier
	secureRand     *security.OTPProvider
}

func NewAuthenticationService(
	userRepo UserReposiotry,
	cacheRepo CacheRepository,
	emailOTPSender EmailOTPSender,
	jwtProvider *security.JWTProvier,
	secureRand *security.OTPProvider,
) *AuthenticationService {
	return &AuthenticationService{
		userRepo:       userRepo,
		cacheRepo:      cacheRepo,
		emailOTPSender: emailOTPSender,
		jwtProvider:    jwtProvider,
		secureRand:     secureRand,
	}
}

func (as *AuthenticationService) RequireOTP(ctx context.Context, email string) error {
	otp := as.secureRand.RandOTP()

	key := "auth-otp-" + email
	if err := as.cacheRepo.SetIfNotExists(ctx, key, otp, time.Minute*5); err != nil {
		return fmt.Errorf("AuthenticationService[SendOTP]: %w", err)
	}

	go as.emailOTPSender.Send(ctx, email, otp)

	return nil
}

func (as *AuthenticationService) VerifyOTP(ctx context.Context, email string, receivedOtp string) (VerifyOTPResp, error) {
	key := "auth-otp-" + email
	otp, err := as.cacheRepo.Get(ctx, key)

	if err != nil || otp != receivedOtp {
		return VerifyOTPResp{}, errs.NewError(errs.AuthenticationFailure, nil, errs.IncorrectOTP)
	}

	if err := as.cacheRepo.Remove(ctx, key); err != nil {
		return VerifyOTPResp{}, fmt.Errorf("AuthenticationService[VerifyOTP]: %w", err)
	}

	user, err := as.userRepo.GetByEmail(ctx, email)

	if e, ok := errors.AsType[*errs.Error](err); ok && e.Type == errs.NotFound  {
		registerToken, err := as.jwtProvider.GenerateRegisterToken(email)
		if err != nil {
			return VerifyOTPResp{}, fmt.Errorf("AuthenticationService[VerifyOTP]: %w", err)
		}

		return VerifyOTPResp{
			Type: RegisterType,
			RegisterToken: registerToken,
		}, nil
	}

	if err != nil {
		return VerifyOTPResp{}, fmt.Errorf("AuthenticationService[VerifyOTP]: %w", err)
	}

	if user.Status != model.UserActive {
		return VerifyOTPResp{}, errs.NewError(errs.AuthenticationFailure, nil, errs.UserNonActive)
	}

	accessToken, err := as.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return VerifyOTPResp{}, err
	}

	refreshToken, err := as.jwtProvider.GenerateRefreshToken(user)
	if err != nil {
		return VerifyOTPResp{}, err
	}

	return VerifyOTPResp{
		Type: LoginType,
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthenticationService) RegisterUser(ctx context.Context, req RegisterUserReq) (map[string]string, error) {
	parsedToken, err := as.jwtProvider.ParseRegisterToken(req.RegisterToken)
	if err != nil {
		return nil, fmt.Errorf("AuthenticationService[RegisterUser]: %w", err)
	}

	email, err := parsedToken.Claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("AuthenticationService[RegisterUser]: err get email(subject field): %w", err)
	}

	user := model.User{
		FirstName:   req.FirstName,
		LastName:    &req.LastName,
		Username:    req.Username,
		PhoneNumber: req.PhoneNumber,
		Email:       email,
	}

	if err := as.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("AuthenticationService[RegisterUser]: %w", err)
	}
	
	accessToken, err := as.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := as.jwtProvider.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		}, nil
}

func (as *AuthenticationService) RefreshToken(ctx context.Context, currenUserID int64) (string, error) {
	u, err := as.userRepo.GetByID(ctx, currenUserID)
	if err != nil {
		return "", fmt.Errorf("AuthenticationService.RefreshToken: %w", err)
	}

	accessToken, err := as.jwtProvider.GenerateAccessToken(u)
	if err != nil {
		return "", fmt.Errorf("AuthenticationService.RefreshToken: %w", err)
	}
	return accessToken, nil
}