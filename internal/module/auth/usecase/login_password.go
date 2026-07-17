package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/domain"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
	"golang.org/x/crypto/bcrypt"
)

type LoginPasswordCmd struct {
	Email    string
	Password string
}

type LoginPasswordRes struct {
	UserID       int64
	Username     string
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

func (lu *LoginPasswordUsecase) Excute(ctx context.Context, cmd LoginPasswordCmd) (LoginPasswordRes, error) {
	user, err := lu.userRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return LoginPasswordRes{}, errs.NewError(errs.Invalid, nil, errs.InvalidUsernameOrPassword)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return LoginPasswordRes{}, errs.NewError(errs.Invalid, nil, errs.InvalidUsernameOrPassword)
	}

	if user.Status != domain.UserActive {
		return LoginPasswordRes{}, errs.NewError(errs.AuthenticationFailure, nil, errs.UserNonActive)
	}

	user.LastLoginAt = time.Now()
	if err := lu.userRepo.UpdateLastLogin(ctx, user); err != nil {
		return LoginPasswordRes{}, errs.NewError(errs.Invalid, nil, errs.InvalidUsernameOrPassword)
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
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
