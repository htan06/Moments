package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/follow/domain"
)

type UnfollowCmd struct {
	UserID   int64
	FollowID int64
}

type UnfollowUsecase struct {
	followRepo     domain.FollowRepository
	followProducer domain.FollowProducer
}

func NewUnfollowUsecase(
	followRepo domain.FollowRepository,
	followProducer domain.FollowProducer,
) *UnfollowUsecase {
	return &UnfollowUsecase{
		followRepo:     followRepo,
		followProducer: followProducer,
	}
}

func (ufu *UnfollowUsecase) Execute(ctx context.Context, cmd UnfollowCmd) error {

	follow, err := ufu.followRepo.RemoveByFollowerID(ctx, cmd.FollowID, cmd.UserID)
	if err != nil {
		return fmt.Errorf("UnfollowUsecase.Execute: %w", err)
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
