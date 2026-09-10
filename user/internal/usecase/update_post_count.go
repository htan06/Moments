package usecase

import (
	"context"
	"log"
	"user-service/internal/domain"
)

type UpdatePostCountUC struct {
	userRepo         domain.UserRepository
	userPostConsumer domain.UserPostConsumer
}

func NewUpdatePostCountUC(
	userRepo domain.UserRepository,
	userPostConsumer domain.UserPostConsumer,
) *UpdatePostCountUC {
	return &UpdatePostCountUC{
		userRepo:         userRepo,
		userPostConsumer: userPostConsumer,
	}
}

func (u *UpdatePostCountUC) Run(ctx context.Context) {
	for {
		postEvent, err := u.userPostConsumer.ReadMessage(ctx)
		if err != nil {
			log.Println("Error: %s", err.Error())
			continue
		}

		switch postEvent.Type {
		case domain.Created:
			err = u.userRepo.IncPostCount(ctx, postEvent.AuthorID)
		case domain.Deleted:
			err = u.userRepo.DecPostCount(ctx, postEvent.AuthorID)
		}
		if err != nil {
			log.Println("Error: %s", err.Error())
		}
	}
}
