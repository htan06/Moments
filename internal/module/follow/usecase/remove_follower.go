package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/domain"
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

func (ufu *RemovefollowerUsecase) Excute(ctx context.Context, cmd RemoveFollowCmd) error {

	if err := ufu.followRepo.RemoveByFollowingID(ctx, cmd.FollowID, cmd.UserID); err != nil {
		return fmt.Errorf("RemovefollowerUsecase.Excute: %w", err)
	}
	return nil
}
