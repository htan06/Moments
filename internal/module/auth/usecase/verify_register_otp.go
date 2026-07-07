package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/domain"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
)

type VerifyRegisterOTPCmd struct {
	Email string
	OTP   string
}

type VerifyRegisterOTPRes struct {
	AccessToken  string
	RefreshToken string
}

type VerifyRegisterOTPUsecase struct {
	userRepo    domain.UserRepository
	cacheRepo   domain.CacheRepository
	jwtProvider *security.JWTProvier
}

func NewVerifyRegisterOTPUsecase(
	userRepo domain.UserRepository,
	cacheRepo domain.CacheRepository,
	jwtProvider *security.JWTProvier,
) *VerifyRegisterOTPUsecase {
	return &VerifyRegisterOTPUsecase{
		userRepo:    userRepo,
		cacheRepo:   cacheRepo,
		jwtProvider: jwtProvider,
	}
}

func (vru *VerifyRegisterOTPUsecase) Excute(ctx context.Context, cmd VerifyRegisterOTPCmd) (VerifyRegisterOTPRes, error) {
	key := fmt.Sprintf("auth-register:%s", cmd.Email)
	userPending, err := vru.cacheRepo.GetUserPending(ctx, key)
	if err != nil {
		return VerifyRegisterOTPRes{}, err
	}

	if userPending.OTP != cmd.OTP {
		return VerifyRegisterOTPRes{}, errs.NewError(errs.Invalid, nil, errs.IncorrectOTP)
	}

	user := domain.User{
		Name:         userPending.Name,
		Email:        userPending.Email,
		Username:     userPending.Username,
		PasswordHash: userPending.PasswordHash,
	}

	if err := vru.userRepo.Create(ctx, user); err != nil {
		return VerifyRegisterOTPRes{}, err
	}

	accessToken, err := vru.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return VerifyRegisterOTPRes{}, err
	}

	refreshToken, err := vru.jwtProvider.GenerateRefreshToken(user)
	if err != nil {
		return VerifyRegisterOTPRes{}, err
	}

	return VerifyRegisterOTPRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
