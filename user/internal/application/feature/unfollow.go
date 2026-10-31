package feature

import (
	"context"
	"fmt"
	"time"
	"user-service/internal/application/domain/entity"
	"user-service/internal/application/domain/event"
	"user-service/internal/ports/outbound"
)

type UnfollowCmd struct {
	FollowerID  int64
	FollowingID int64
}

type Unfollow struct {
	followRepo     outbound.FollowRepository
	followProducer outbound.FollowProducer
}

func NewUnfollow(
	followRepo outbound.FollowRepository,
	followProducer outbound.FollowProducer,
) *Unfollow {
	return &Unfollow{
		followRepo:     followRepo,
		followProducer: followProducer,
	}
}

func (ufu *Unfollow) Execute(ctx context.Context, cmd UnfollowCmd) error {
	follow, err := entity.NewFollow(cmd.FollowerID, cmd.FollowingID)
	if err != nil {
		return fmt.Errorf("Unfollow.Execute: %w", err)
	}

	if err := ufu.followRepo.RemoveByIDAndFollowerID(ctx, *follow); err != nil {
		return fmt.Errorf("Unfollow.Execute: %w", err)
	}

	fdevent, err := event.NewFollowDeleted(follow.FollowerID(), follow.FollowingID(), time.Now().UTC())
	if err != nil {
		return err
	}

	if err := ufu.followProducer.PublishFollowDeleted(ctx, *fdevent); err != nil {
		return fmt.Errorf("Unfollow.Execute: %w", err)
	}

	return nil
}
