package usecase

import (
	"context"
	"fmt"
	"regexp"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

type UpdateProfileCmd struct {
	UserID   int64
	Username *string
	Name     *string
	Bio      *string
}

type UpdateProfileUsecase struct {
	userRepo domain.UserRepository
}

func NewUpdateProfileUsecase(userRepo domain.UserRepository) *UpdateProfileUsecase {
	return &UpdateProfileUsecase{
		userRepo: userRepo,
	}
}

func (gpu *UpdateProfileUsecase) Execute(ctx context.Context, cmd UpdateProfileCmd) error {
	fieldsUpdates := map[string]interface{}{}

	if cmd.Username != nil && !usernameRegex.MatchString(*cmd.Username) {
		return errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}

	if cmd.Username != nil {
		fieldsUpdates["username"] = *cmd.Username
	}

	if cmd.Name != nil {
		fieldsUpdates["name"] = *cmd.Name
	}

	if cmd.Bio != nil {
		fieldsUpdates["bio"] = *cmd.Bio
	}

	if err := gpu.userRepo.UpdateProfile(ctx, cmd.UserID, fieldsUpdates); err != nil {
		return fmt.Errorf("UpdateProfileUsecase.Execute: %w", err)
	}
	return nil
}
