package grpc

import (
	"context"

	"github.com/htan06/Moments/grpc/user"
	"github.com/htan06/Moments/user/internal/application/feature"
)

type ProfileHandler struct {
	user.UnimplementedUserServiceServer
	getProfilesSummary *feature.GetProfilesSummary
}

func NewProfileHandler(getProfilesSummary *feature.GetProfilesSummary) *ProfileHandler {
	return &ProfileHandler{
		getProfilesSummary: getProfilesSummary,
	}
}

func (p *ProfileHandler) GetProfiles(ctx context.Context, input *user.GetProfilesSummaryInput) (*user.GetProfilesSummaryOutput, error) {
	proriles, err := p.getProfilesSummary.Execute(ctx, input.CurrentUserID, input.Ids)
	if err != nil {
		return nil, err
	}

	var profilesOutput []*user.ProfileSummary
	for _, p := range proriles {
		profilesOutput = append(profilesOutput, &user.ProfileSummary{
			UserID:          p.UserID,
			Username:        p.Username,
			Name:            p.Name,
			AvatarThumbnail: *p.AvatarThumbnailURL,
		})
	}
	return &user.GetProfilesSummaryOutput{
		Profiles: profilesOutput,
	}, nil
}
