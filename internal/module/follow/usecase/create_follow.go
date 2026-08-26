package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/follow/domain"
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

type CreateFollowUsecase struct {
	followRepo     domain.FollowRepository
	followProducer domain.FollowProducer
}

func NewCreateFollowUsecase(
	followRepo domain.FollowRepository,
	followProducer domain.FollowProducer,
) *CreateFollowUsecase {
	return &CreateFollowUsecase{
		followRepo:     followRepo,
		followProducer: followProducer,
	}
}

func (fuu *CreateFollowUsecase) Execute(ctx context.Context, cmd CreateFollowCmd) (*int64, error) {
	if cmd.FollowerID == cmd.FollowingID {
		return nil, errs.NewError(errs.Invalid, nil, domain.FollowInvalid)
	}

	follow := domain.Follow{
		FollowerID:  cmd.FollowerID,
		FollowingID: cmd.FollowingID,
	}

	id, err := fuu.followRepo.CreateFollow(ctx, follow)
	if err != nil {
		return nil, err
	}

	if err := fuu.followProducer.SendMessage(ctx, domain.FollowEvent{
		FollowerID:  follow.FollowerID,
		FollowingID: follow.FollowingID,
		Type:        domain.FollowCreated,
	}); err != nil {
		return nil, fmt.Errorf("CreateFollowUC.Execute: %w", err)
	}

	return id, nil
}
