package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/auth/internal/domain"
	"github.com/htan06/Moments/auth/internal/security"
)

type ActiveUserRes struct {
	UserID       int64
	Status       domain.UserStatus
	Username     *string
	AccessToken  string
	RefreshToken string
}

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

func (uc *ActiveUserUC) Execute(ctx context.Context, userID int64) (ActiveUserRes, error) {
	if err := uc.userRepo.UpdateStatusActiveIfExistsProfile(ctx, userID); err != nil {
		return ActiveUserRes{}, fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ActiveUserRes{}, fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}

	accessToken, err := uc.jwtProvider.GenerateAccessToken(u)
	if err != nil {
		return ActiveUserRes{}, fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}

	reafreshToken, err := uc.jwtProvider.GenerateRefreshToken(u)
	if err != nil {
		return ActiveUserRes{}, fmt.Errorf("ActiveUserUC.Execute: %w", err)
	}

	return ActiveUserRes{
		UserID:       u.ID,
		Username:     u.UserName,
		Status:       u.Status,
		AccessToken:  accessToken,
		RefreshToken: reafreshToken,
	}, nil
}
