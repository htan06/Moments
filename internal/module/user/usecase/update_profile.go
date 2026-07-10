package usecase

import (
	"context"
	"regexp"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

type UpdateProfileCmd struct {
	UserID    int64
	Username  *string
	Name      *string
	AvatarURL *string
	Bio       *string
}

type UpdateProfileUsecase struct {
	userRepo domain.UserRepository
}

func NewUpdateProfileUsecase(userRepo domain.UserRepository) *UpdateProfileUsecase {
	return &UpdateProfileUsecase{
		userRepo: userRepo,
	}
}

func (gpu *UpdateProfileUsecase) Excute(ctx context.Context, cmd UpdateProfileCmd) (domain.UserProfile, error) {
	fieldsUpdates := map[string]interface{}{}

	if cmd.Username != nil && !usernameRegex.MatchString(*cmd.Username) {
		return domain.UserProfile{}, errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}
	
	if cmd.Username != nil {
		fieldsUpdates["username"] = *cmd.Username
	}

	if cmd.Name != nil {
		fieldsUpdates["name"] = *cmd.Name
	}

	if cmd.AvatarURL != nil {
		fieldsUpdates["avatar_url"] = *cmd.AvatarURL
	}

	if cmd.Bio != nil {
		fieldsUpdates["bio"] = *cmd.Bio
	}

	up, err := gpu.userRepo.UpdateProfile(ctx, cmd.UserID, fieldsUpdates)
	if err != nil {
		return domain.UserProfile{}, err
	}
	return up, nil
}
