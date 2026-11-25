package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/application/readmodel"
)

type GetProfilesSummary struct {
	profileReader ProfileReader
}

func NewGetProfilesSummary(profileReader ProfileReader) *GetProfilesSummary {
	return &GetProfilesSummary{
		profileReader: profileReader,
	}
}

func (t *GetProfilesSummary) Execute(ctx context.Context, currentUserID int64, ids []int64) ([]readmodel.ProfileSummary, error) {
	profiles, err := t.profileReader.GetProfilesSummary(ctx, currentUserID, ids)
	if err != nil {
		return nil, fmt.Errorf("GetProfilesSummary.Execute: %w", err)
	}

	for _, p := range profiles {
		*p.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *p.AvatarThumbnailURL)
	}
	return profiles, nil
}
