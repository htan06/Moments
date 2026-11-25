package feature

import (
	"context"
	"fmt"
	"regexp"

	"github.com/htan06/Moments/user/internal/domain"
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
	profileUpdater ProfileUpdater
}

func NewUpdateProfile(profileUpdater ProfileUpdater) *UpdateProfile {
	return &UpdateProfile{
		profileUpdater: profileUpdater,
	}
}

func (gpu *UpdateProfile) Execute(ctx context.Context, cmd UpdateProfileCmd) (UpdateProfileRes, error) {
	fieldsUpdates := map[string]interface{}{}
	res := UpdateProfileRes{}

	if cmd.Username != nil && !usernameRegex.MatchString(*cmd.Username) {
		return UpdateProfileRes{}, domain.ErrUsernameInvalid
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

	if err := gpu.profileUpdater.UpdateProfile(ctx, cmd.UserID, fieldsUpdates); err != nil {
		return UpdateProfileRes{}, fmt.Errorf("UpdateProfile.Execute: %w", err)
	}
	return res, nil
}
