package feature

// import (
// 	"context"
// 	"log"
// 	"user-service/internal/application/domain/event"
// 	"user-service/internal/ports/inbound"
// 	"user-service/internal/ports/outbound"
// )

// type UpdateFollowCount struct {
// 	userRepo           outbound.UserRepository
// 	userFollowConsumer inbound.UserFollowConsumer
// }

// func NewUpdateFollowCount(
// 	userRepo outbound.UserRepository,
// 	userFollowConsumer inbound.UserFollowConsumer,
// ) *UpdateFollowCount {
// 	return &UpdateFollowCount{
// 		userRepo:           userRepo,
// 		userFollowConsumer: userFollowConsumer,
// 	}
// }

// func (u *UpdateFollowCount) Run(ctx context.Context) {
// 	for {
// 		followCreated, err := u.userFollowConsumer.ReadMessage(ctx)
// 		if err != nil {
// 			log.Println("Update Follow Count  Error: %v", err)
// 			continue
// 		}

// 		switch followCreated.Type {
// 		case event.FollowCreated:
// 			err = u.userRepo.IncFollowCount(ctx, followCreated.FollowerID, followCreated.FollowingID)
// 		case event.FollowDeleted:
// 			err = u.userRepo.DecFollowCount(ctx, followCreated.FollowerID, followCreated.FollowingID)
// 		}

// 		if err != nil {
// 			log.Println("Update Follow Count  Error: %v", err)
// 		}
// 	}
// }
