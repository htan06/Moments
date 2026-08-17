package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/auth/domain"
)

type GetCurrentUserUC struct {
	userRepo domain.UserRepository
}

func NewGetCurrentUserUC(
	userRepo domain.UserRepository,
) *GetCurrentUserUC {
	return &GetCurrentUserUC{
		userRepo: userRepo,
	}
}

func (uc *GetCurrentUserUC) Execute(ctx context.Context, userID int64) (domain.UserDetail, error) {
	ud, err := uc.userRepo.GetUserDetailByID(ctx, userID)
	if err != nil {
		return domain.UserDetail{}, fmt.Errorf("GetCurrentUserUC.Execute: %w", err)
	}
	return ud, nil
}
