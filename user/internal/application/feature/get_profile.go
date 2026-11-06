package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/application/readmodel"
	"github.com/htan06/Moments/user/internal/domain"
)

type GetProfileQry struct {
	CurrentUserID   int64
	CurrentUsername string
	TargetUsername  string
}

type GetProfile struct {
	profileReader ProfileReader
}

func NewGetProfile(profileReader ProfileReader) *GetProfile {
	return &GetProfile{
		profileReader: profileReader,
	}
}

func (gpu *GetProfile) Execute(ctx context.Context, qry GetProfileQry) (readmodel.Profile, error) {
	if qry.TargetUsername == "" {
		return readmodel.Profile{}, domain.ErrUserNotFound
	}

	var p readmodel.Profile
	var err error

	if qry.CurrentUsername == qry.TargetUsername {
		p, err = gpu.profileReader.GetSelfProfileByUsername(ctx, qry.TargetUsername)
		p.IsOwner = true
	} else {
		p, err = gpu.profileReader.GetOtherProfileByUsername(ctx, qry.CurrentUserID, qry.TargetUsername)
		p.IsOwner = false
	}

	if err != nil {
		return readmodel.Profile{}, fmt.Errorf("NewGetProfile.Execute: %w", err)
	}

	if p.AvatarURL != nil {
		*p.AvatarURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *p.AvatarURL)
	}

	if p.AvatarThumbnailURL != nil {
		*p.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *p.AvatarThumbnailURL)
	}

	return p, nil
}
