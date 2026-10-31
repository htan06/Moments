package feature

import (
	"context"
	"fmt"
	"regexp"
	"user-service/internal/errs"
	"user-service/internal/ports/outbound"
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

type UpdateProfileRes struct {
	Username *string `json:"username,omitempty"`
	Name     *string `json:"name,omitempty"`
	Bio      *string `json:"bio,omitempty"`
}

type UpdateProfile struct {
	userRepo outbound.UserRepository
}

func NewUpdateProfile(userRepo outbound.UserRepository) *UpdateProfile {
	return &UpdateProfile{
		userRepo: userRepo,
	}
}

func (gpu *UpdateProfile) Execute(ctx context.Context, cmd UpdateProfileCmd) (UpdateProfileRes, error) {
	fieldsUpdates := map[string]interface{}{}
	res := UpdateProfileRes{}

	if cmd.Username != nil && !usernameRegex.MatchString(*cmd.Username) {
		return UpdateProfileRes{}, errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}

	if cmd.Username != nil {
		fieldsUpdates["username"] = *cmd.Username
		res.Username = cmd.Username
	}

	if cmd.Name != nil {
		fieldsUpdates["name"] = *cmd.Name
		res.Name = cmd.Name
	}

	if cmd.Bio != nil {
		fieldsUpdates["bio"] = *cmd.Bio
		res.Bio = cmd.Bio
	}

	if err := gpu.userRepo.UpdateProfile(ctx, cmd.UserID, fieldsUpdates); err != nil {
		return UpdateProfileRes{}, fmt.Errorf("UpdateProfile.Execute: %w", err)
	}
	return res, nil
}
