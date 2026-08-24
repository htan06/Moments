package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/auth/domain"
	"github.com/htan06/Moments/internal/security"
	"golang.org/x/crypto/bcrypt"
)

type LoginPasswordCmd struct {
	Email    string
	Password string
}

type LoginPasswordRes struct {
	UserID       int64
	Status       domain.UserStatus
	Username     *string
	AccessToken  string
	RefreshToken string
}

type LoginPasswordUsecase struct {
	userRepo    domain.UserRepository
	jwtProvider *security.JWTProvier
}

func NewLoginPasswordUsecase(
	userRepo domain.UserRepository,
	jwtProvider *security.JWTProvier,
) *LoginPasswordUsecase {
	return &LoginPasswordUsecase{
		userRepo:    userRepo,
		jwtProvider: jwtProvider,
	}
}

func (lu *LoginPasswordUsecase) Execute(ctx context.Context, cmd LoginPasswordCmd) (LoginPasswordRes, error) {
	user, err := lu.userRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return LoginPasswordRes{}, errs.NewError(errs.Invalid, nil, domain.UsernameOrPasswordInvalid)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return LoginPasswordRes{}, errs.NewError(errs.Invalid, nil, domain.UsernameOrPasswordInvalid)
	}

	if user.Status == domain.UserStatusInactive {
		return LoginPasswordRes{}, errs.NewError(errs.AuthenticationFailure, nil, domain.UserNonActiveErr)
	}

	user.LastLoginAt = time.Now().UTC()
	if err := lu.userRepo.UpdateLastLogin(ctx, user); err != nil {
		return LoginPasswordRes{}, errs.NewError(errs.Invalid, nil, domain.UsernameOrPasswordInvalid)
	}

	accessToken, err := lu.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return LoginPasswordRes{}, fmt.Errorf("LoginPasswordUsecase.Excute %w", err)
	}

	refreshToken, err := lu.jwtProvider.GenerateRefreshToken(user)
	if err != nil {
		return LoginPasswordRes{}, fmt.Errorf("LoginPasswordUsecase.Excute %w", err)
	}

	return LoginPasswordRes{
		UserID:       user.ID,
		Status:       user.Status,
		Username:     user.UserName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
