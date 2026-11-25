package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/application/readmodel"
)

type Search struct {
	profileFinder ProfileReader
}

func NewSearch(profileFinder ProfileReader) *Search {
	return &Search{
		profileFinder: profileFinder,
	}
}

func (s *Search) Execute(ctx context.Context, username string) ([]readmodel.ProfileSummary, error) {
	profiles, err := s.profileFinder.FindProfilesByUsername(ctx, username)
	if err != nil {
		return []readmodel.ProfileSummary{}, fmt.Errorf("Search.Execute: %w", err)
	}

	for _, p := range profiles {
		if p.AvatarThumbnailURL != nil {
			*p.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *p.AvatarThumbnailURL)
		}
	}
	return profiles, nil
}
