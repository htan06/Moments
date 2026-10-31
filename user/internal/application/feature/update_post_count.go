package feature

// import (
// 	"context"
// 	"log"
// 	"user-service/internal/application/domain/event"
// 	"user-service/internal/ports/inbound"
// 	"user-service/internal/ports/outbound"
// )

// type UpdatePostCount struct {
// 	userRepo         outbound.UserRepository
// 	userPostConsumer inbound.UserPostConsumer
// }

// func NewUpdatePostCount(
// 	userRepo outbound.UserRepository,
// 	userPostConsumer inbound.UserPostConsumer,
// ) *UpdatePostCount {
// 	return &UpdatePostCount{
// 		userRepo:         userRepo,
// 		userPostConsumer: userPostConsumer,
// 	}
// }

// func (u *UpdatePostCount) Run(ctx context.Context) {
// 	for {
// 		postEvent, err := u.userPostConsumer.ReadMessage(ctx)
// 		if err != nil {
// 			log.Println("Error: %s", err.Error())
// 			continue
// 		}

// 		switch postEvent.Action {
// 		case event.Created:
// 			err = u.userRepo.IncPostCount(ctx, postEvent.AuthorID)
// 		case event.Deleted:
// 			err = u.userRepo.DecPostCount(ctx, postEvent.AuthorID)
// 		}
// 		if err != nil {
// 			log.Println("Error: %s", err.Error())
// 		}
// 	}
// }
