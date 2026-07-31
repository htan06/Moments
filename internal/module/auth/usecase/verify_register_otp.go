package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/config"
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

func (vru *VerifyRegisterOTPUsecase) Execute(ctx context.Context, cmd VerifyRegisterOTPCmd) error {
	key := fmt.Sprintf("%s:%s", config.UserRegisterPrefix, cmd.Email)
	userPending, err := vru.cacheRepo.GetUserPending(ctx, key)
	if err != nil {
		return err
	}

	if userPending.OTP != cmd.OTP {
		return errs.NewError(errs.Invalid, nil, domain.OTPIncorrect)
	}

	user, err := domain.NewUser(
		userPending.Username,
		userPending.Name,
		userPending.Email,
		nil,
		userPending.PasswordHash,
	)

	if err != nil {
		return err
	}

	if err := vru.userRepo.Create(ctx, *user); err != nil {
		return err
	}

	return nil
}
