package feature

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/application/domain/readmodel"
	"user-service/internal/ports/outbound"
)

type Search struct {
	userRepo outbound.UserRepository
}

func NewSearch(userRepo outbound.UserRepository) *Search {
	return &Search{
		userRepo: userRepo,
	}
}

func (s *Search) Execute(ctx context.Context, username string) ([]readmodel.ProfileSummary, error) {
	profiles, err := s.userRepo.FindProfilesByUsername(ctx, username)
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
