package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/auth/internal/domain"
	"github.com/htan06/Moments/auth/internal/security"
)

type RefreshTokenUsecase struct {
	userRepo    domain.UserRepository
	jwtProvider *security.JWTProvier
}

func NewRefreshTokenUsecase(
	userRepo domain.UserRepository,
	jwtProvider *security.JWTProvier,
) *RefreshTokenUsecase {
	return &RefreshTokenUsecase{
		userRepo:    userRepo,
		jwtProvider: jwtProvider,
	}
}

func (ru *RefreshTokenUsecase) Execute(ctx context.Context, userID int64) (string, error) {
	user, err := ru.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("RefreshTokenUsecase.Excute %w", err)
	}

	accessToken, err := ru.jwtProvider.GenerateAccessToken(user)
	if err != nil {
		return "", fmt.Errorf("RefreshTokenUsecase.Excute %w", err)
	}
	return accessToken, nil
}
