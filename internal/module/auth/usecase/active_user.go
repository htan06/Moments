package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/auth/domain"
	"github.com/htan06/Moments/internal/security"
)

type ActiveUserUC struct {
	userRepo    domain.UserRepository
	jwtProvider *security.JWTProvier
}

func NewActiveUserUC(
	userRepo domain.UserRepository,
	jwtProvider *security.JWTProvier,
) *ActiveUserUC {
	return &ActiveUserUC{
		userRepo:    userRepo,
		jwtProvider: jwtProvider,
	}
}

func (uc *ActiveUserUC) Execute(ctx context.Context, userID int64) (string, error) {
	if err := uc.userRepo.UpdateStatusActiveIfExistsProfile(ctx, userID); err != nil {
		return "", fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}

	accessToken, err := uc.jwtProvider.GenerateAccessToken(u)
	if err != nil {
		return "", fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}
	return accessToken, nil
}
