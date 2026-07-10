package user

// import (
// 	"context"
// 	"fmt"

// 	"github.com/htan06/echo-messenger-rest-api/internal/module/user/model"
// )

// type UserService struct {
// 	userRepo UserRepository
// }

// func NewUserService(
// 	userRepo UserRepository,
// ) *UserService {
// 	return &UserService{
// 		userRepo: userRepo,
// 	}
// }

// func (us *UserService) GetProfileById(ctx context.Context, userId int64) (model.UserProfile, error) {
// 	user, err := us.userRepo.GetInfo(ctx, userId)
// 	if err != nil {
// 		return model.UserProfile{}, fmt.Errorf("UserService[GetInfo]: %w", err)
// 	}
// 	return model.UserProfile{
// 		ID:            user.ID,
// 		Username:      user.Username,
// 		FirstName:     user.FirstName,
// 		LastName:      user.LastName,
// 		AvatarURL:     user.AvatarURL,
// 		CoverPhotoURL: user.CoverPhotoURL,
// 	}, nil
// }

// func (us *UserService) UpdateInfo(ctx context.Context, userId int64, req UpdateInfoReq) error {
// 	user := model.User{
// 		ID:            userId,
// 		FirstName:     req.FirstName,
// 		LastName:      req.LastName,
// 		AvatarURL:     &req.AvatarURL,
// 		CoverPhotoURL: &req.CoverPhotoURL,
// 	}

// 	if err := us.userRepo.UpdateInfo(ctx, user); err != nil {
// 		return fmt.Errorf("UserService[UpdateInfo]: %w", err)
// 	}
// 	return nil
// }

// func (us *UserService) ChangeReadStatus(ctx context.Context, userId int64, req ChangeReadStatusReq) error {
// 	user := model.User{
// 		ID: userId,
// 		ReadStatus: &req.Status,
// 	}

// 	if err := us.userRepo.ChangeReadStatus(ctx, user); err != nil {
// 		return fmt.Errorf("UserService[ChangeReadStatus]: %w", err)
// 	}
// 	return nil
// }

// func (us *UserService) UpdateUsername(ctx context.Context, userId int64, req UpdateUsernameReq) error {
// 	user := model.User{
// 		ID: userId,
// 		Username: req.Username,
// 	}

// 	if err := us.userRepo.UpdateUsername(ctx, user); err != nil {
// 		return fmt.Errorf("UserService[UpdateUsername]: %w", err)
// 	}
// 	return nil
// }

// func (us *UserService) GetProfileByUserName(ctx context.Context, username string) (model.UserProfile, error) {
// 	up, err := us.userRepo.FindByUsername(ctx, username)
// 	if err != nil {
// 		return model.UserProfile{}, fmt.Errorf("FriendService[GetProfileByUserName]: %w", err)
// 	}
// 	return up, nil
// }