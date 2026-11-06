package feature

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/user/internal/domain"
)

type CreateFollowCmd struct {
	FollowerID  int64
	FollowingID int64
}

type CreateFollowRes struct {
	FollowID  int64
	UserID    int64
	Username  string
	Name      string
	AvatarUrl string
	CreatedAt time.Time
}

type CreateFollow struct {
	followCreator  FollowCreator
	followProducer FollowProducer
}

func NewCreateFollow(
	followCreator FollowCreator,
	followProducer FollowProducer,
) *CreateFollow {
	return &CreateFollow{
		followCreator:  followCreator,
		followProducer: followProducer,
	}
}

func (fuu *CreateFollow) Execute(ctx context.Context, cmd CreateFollowCmd) error {
	if cmd.FollowerID == cmd.FollowingID {
		return domain.ErrUserIDInvalid
	}

	follow, err := domain.NewFollow(cmd.FollowerID, cmd.FollowingID)
	if err != nil {
		return fmt.Errorf("CreateFollow.Execute: %w", err)
	}

	if err := fuu.followCreator.CreateFollow(ctx, *follow); err != nil {
		return fmt.Errorf("CreateFollow.Execute: %w", err)
	}

	fcevent, err := domain.NewFollowCreated(follow.FollowerID(), follow.FollowingID(), follow.CreatedAt())
	if err != nil {
		return fmt.Errorf("CreateFollow.Execute: %w", err)
	}

	if err := fuu.followProducer.PublishFollowCreated(ctx, *fcevent); err != nil {
		return fmt.Errorf("CreateFollowUC.Execute: %w", err)
	}

	return nil
}
