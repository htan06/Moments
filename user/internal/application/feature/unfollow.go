package feature

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/user/internal/domain"
)

type UnfollowCmd struct {
	FollowerID  int64
	FollowingID int64
}

type Unfollow struct {
	followDeleter     FollowDeleter
	followProducer FollowProducer
}

func NewUnfollow(
	followDeleter FollowDeleter,
	followProducer FollowProducer,
) *Unfollow {
	return &Unfollow{
		followDeleter:     followDeleter,
		followProducer: followProducer,
	}
}

func (ufu *Unfollow) Execute(ctx context.Context, cmd UnfollowCmd) error {
	follow, err := domain.NewFollow(cmd.FollowerID, cmd.FollowingID)
	if err != nil {
		return fmt.Errorf("Unfollow.Execute: %w", err)
	}

	if err := ufu.followDeleter.RemoveByIDAndFollowerID(ctx, *follow); err != nil {
		return fmt.Errorf("Unfollow.Execute: %w", err)
	}

	fdevent, err := domain.NewFollowDeleted(follow.FollowerID(), follow.FollowingID(), time.Now().UTC())
	if err != nil {
		return err
	}

	if err := ufu.followProducer.PublishFollowDeleted(ctx, *fdevent); err != nil {
		return fmt.Errorf("Unfollow.Execute: %w", err)
	}

	return nil
}
