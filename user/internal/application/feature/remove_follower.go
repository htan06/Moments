package feature

import (
	"context"
	"fmt"
	"time"
	"user-service/internal/application/domain/entity"
	"user-service/internal/application/domain/event"
	"user-service/internal/ports/outbound"
)

type RemoveFollowCmd struct {
	FollowerID  int64
	FollowingID int64
}

type Removefollower struct {
	followRepo     outbound.FollowRepository
	followProducer outbound.FollowProducer
}

func NewRemovefollower(
	followRepo outbound.FollowRepository,
	followProducer outbound.FollowProducer,
) *Removefollower {
	return &Removefollower{
		followRepo: followRepo,
	}
}

func (ufu *Removefollower) Execute(ctx context.Context, cmd RemoveFollowCmd) error {
	follow, err := entity.NewFollow(cmd.FollowerID, cmd.FollowingID)
	if err != nil {
		return fmt.Errorf("Removefollower.Execute: %w", err)
	}

	if err := ufu.followRepo.RemoveByIDAndFollowingID(ctx, *follow); err != nil {
		return fmt.Errorf("Removefollower.Execute: %w", err)
	}

	fdevent, err := event.NewFollowDeleted(follow.FollowerID(), follow.FollowingID(), time.Now().UTC())
	if err != nil {
		return err
	}

	if err := ufu.followProducer.PublishFollowDeleted(ctx, *fdevent); err != nil {
		return fmt.Errorf("CreateFollowUC.Execute: %w", err)
	}

	return nil
}
