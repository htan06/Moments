package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/follow/domain"
)

type RemoveFollowCmd struct {
	UserID   int64
	FollowID int64
}

type RemovefollowerUsecase struct {
	followRepo domain.FollowRepository
}

func NewRemovefollowerUsecase(followRepo domain.FollowRepository) *RemovefollowerUsecase {
	return &RemovefollowerUsecase{
		followRepo: followRepo,
	}
}

func (ufu *RemovefollowerUsecase) Execute(ctx context.Context, cmd RemoveFollowCmd) error {

	if err := ufu.followRepo.RemoveByFollowingID(ctx, cmd.FollowID, cmd.UserID); err != nil {
		return fmt.Errorf("RemovefollowerUsecase.Execute: %w", err)
	}
	return nil
}
