package feature

import (
	"context"
	"fmt"
	"time"
	"user-service/internal/application/domain"
	"user-service/internal/application/domain/entity"
	"user-service/internal/application/domain/event"
	"user-service/internal/errs"
	"user-service/internal/ports/outbound"
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
	followRepo     outbound.FollowRepository
	followProducer outbound.FollowProducer
}

func NewCreateFollow(
	followRepo outbound.FollowRepository,
	followProducer outbound.FollowProducer,
) *CreateFollow {
	return &CreateFollow{
		followRepo:     followRepo,
		followProducer: followProducer,
	}
}

func (fuu *CreateFollow) Execute(ctx context.Context, cmd CreateFollowCmd) error {
	if cmd.FollowerID == cmd.FollowingID {
		return errs.NewError(errs.Invalid, nil, domain.FollowInvalid)
	}

	follow, err := entity.NewFollow(cmd.FollowerID, cmd.FollowingID)
	if err != nil {
		return fmt.Errorf("CreateFollow.Execute: %w", err)
	}

	if err := fuu.followRepo.CreateFollow(ctx, *follow); err != nil {
		return fmt.Errorf("CreateFollow.Execute: %w", err)
	}

	fcevent, err := event.NewFollowCreated(follow.FollowerID(), follow.FollowingID(), follow.CreatedAt())
	if err != nil {
		return fmt.Errorf("CreateFollow.Execute: %w", err)
	}

	if err := fuu.followProducer.PublishFollowCreated(ctx, *fcevent); err != nil {
		return fmt.Errorf("CreateFollowUC.Execute: %w", err)
	}

	return nil
}
