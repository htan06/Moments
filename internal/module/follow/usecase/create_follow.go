package usecase

import (
	"context"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/domain"
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
	followRepo domain.FollowRepository
}

func NewCreateFollowUsecase(followRepo domain.FollowRepository) *CreateFollowUsecase {
	return &CreateFollowUsecase{
		followRepo: followRepo,
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

	return id, nil
}
