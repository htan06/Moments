package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/domain"
)

type UnfollowCmd struct {
	UserID   int64
	FollowID int64
}

type UnfollowUsecase struct {
	followRepo domain.FollowRepository
}

func NewUnfollowUsecase(followRepo domain.FollowRepository) *UnfollowUsecase {
	return &UnfollowUsecase{
		followRepo: followRepo,
	}
}

func (ufu *UnfollowUsecase) Execute(ctx context.Context, cmd UnfollowCmd) error {

	if err := ufu.followRepo.RemoveByFollowerID(ctx, cmd.FollowID, cmd.UserID); err != nil {
		return fmt.Errorf("UnfollowUsecase.Execute: %w", err)
	}
	return nil
}
