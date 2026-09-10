package usecase

import (
	"context"
	"fmt"
	"user-service/internal/domain"
)

type RemoveFollowCmd struct {
	UserID   int64
	FollowID int64
}

type RemovefollowerUsecase struct {
	followRepo     domain.FollowRepository
	followProducer domain.FollowProducer
}

func NewRemovefollowerUsecase(
	followRepo domain.FollowRepository,
	followProducer domain.FollowProducer,
) *RemovefollowerUsecase {
	return &RemovefollowerUsecase{
		followRepo: followRepo,
	}
}

func (ufu *RemovefollowerUsecase) Execute(ctx context.Context, cmd RemoveFollowCmd) error {

	follow, err := ufu.followRepo.RemoveByFollowingID(ctx, cmd.FollowID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("RemovefollowerUsecase.Execute: %w", err)
	}

	if err := ufu.followProducer.SendMessage(ctx, domain.FollowEvent{
		FollowerID:  follow.FollowerID,
		FollowingID: follow.FollowingID,
		Type:        domain.FollowDeleted,
	}); err != nil {
		return fmt.Errorf("CreateFollowUC.Execute: %w", err)
	}

	return nil
}
