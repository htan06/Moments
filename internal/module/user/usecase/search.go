package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/user/domain"
)

type SearchUC struct {
	userRepo domain.UserRepository
}

func NewSearchUC(userRepo domain.UserRepository) *SearchUC {
	return &SearchUC{
		userRepo: userRepo,
	}
}

func (s *SearchUC) Execute(ctx context.Context, username string) ([]domain.ProfileSummaryReadModel, error) {
	profiles, err := s.userRepo.FindProfilesByUsername(ctx, username)
	if err != nil {
		return []domain.ProfileSummaryReadModel{}, fmt.Errorf("SearchUC.Execute: %w", err)
	}

	for _, p := range profiles {
		if p.AvatarThumbnailURL != nil {
			*p.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *p.AvatarThumbnailURL)
		}
	}
	return profiles, nil
}
