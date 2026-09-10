package usecase

import (
	"context"
	"log"
	"user-service/internal/domain"
)

type UpdateFollowCountUC struct {
	userRepo           domain.UserRepository
	userFollowConsumer domain.UserFollowConsumer
}

func NewUpdateFollowCountUC(
	userRepo domain.UserRepository,
	userFollowConsumer domain.UserFollowConsumer,
) *UpdateFollowCountUC {
	return &UpdateFollowCountUC{
		userRepo:           userRepo,
		userFollowConsumer: userFollowConsumer,
	}
}

func (u *UpdateFollowCountUC) Run(ctx context.Context) {
	for {
		followCreated, err := u.userFollowConsumer.ReadMessage(ctx)
		if err != nil {
			log.Println("Update Follow Count UC Error: %v", err)
			continue
		}

		switch followCreated.Type {
		case domain.FollowCreated:
			err = u.userRepo.IncFollowCount(ctx, followCreated.FollowerID, followCreated.FollowingID)
		case domain.FollowDeleted:
			err = u.userRepo.DecFollowCount(ctx, followCreated.FollowerID, followCreated.FollowingID)
		}

		if err != nil {
			log.Println("Update Follow Count UC Error: %v", err)
		}
	}
}
