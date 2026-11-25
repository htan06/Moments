package feature

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/user/internal/domain"
)

type RemoveFollowCmd struct {
	FollowerID  int64
	FollowingID int64
}

type Removefollower struct {
	followDeleter     FollowDeleter
	followProducer FollowProducer
}

func NewRemovefollower(
	followDeleter FollowDeleter,
	followProducer FollowProducer,
) *Removefollower {
	return &Removefollower{
		followDeleter: followDeleter,
	}
}

func (ufu *Removefollower) Execute(ctx context.Context, cmd RemoveFollowCmd) error {
	follow, err := domain.NewFollow(cmd.FollowerID, cmd.FollowingID)
	if err != nil {
		return fmt.Errorf("Removefollower.Execute: %w", err)
	}

	if err := ufu.followDeleter.RemoveByIDAndFollowingID(ctx, *follow); err != nil {
		return fmt.Errorf("Removefollower.Execute: %w", err)
	}

	fdevent, err := domain.NewFollowDeleted(follow.FollowerID(), follow.FollowingID(), time.Now().UTC())
	if err != nil {
		return err
	}

	if err := ufu.followProducer.PublishFollowDeleted(ctx, *fdevent); err != nil {
		return fmt.Errorf("CreateFollowUC.Execute: %w", err)
	}

	return nil
}
