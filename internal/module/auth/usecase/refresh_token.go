package usecase

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/domain"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
)

type RefreshTokenUsecase struct {
	userRepo domain.UserRepository
	jwtProvider *security.JWTProvier
}

func NewRefreshTokenUsecase(
	userRepo domain.UserRepository,
	jwtProvider *security.JWTProvier,
) *RefreshTokenUsecase {
	return &RefreshTokenUsecase{
		userRepo: userRepo,
		jwtProvider: jwtProvider,
	}
}

func (ru *RefreshTokenUsecase) Excute(ctx context.Context, userID int64) (string, error) {
	user, err := ru.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}

	accessToken, err := ru.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
