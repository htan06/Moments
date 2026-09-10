package domain

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/htan06/Moments/internal/errs"
)

type CreatePostSession struct {
	SessionID uuid.UUID              `json:"session_id"`
	UserID    int64                  `json:"user_id"`
	MediaIDs  map[uuid.UUID]struct{} `json:"media_ids"`
}

func NewCreatePostSession(userID int64, mediaCount int) (*CreatePostSession, error) {
	if userID < 0 {
		return nil, fmt.Errorf("CreatePostSession.NewCreatePostSession: User id invalid")
	}

	if mediaCount < 1 {
		return nil, errs.NewError(errs.Invalid, nil, MediaCountInvalid)
	}

	sessionID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("CreatePostSession.NewCreatePostSession: %w", err)
	}

	mediaIDs := make(map[uuid.UUID]struct{}, 0)
	for i := 0; i < mediaCount; i++ {
		mediaID, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("CreatePostSession.NewCreatePostSession: %w", err)
		}
		mediaIDs[mediaID] = struct{}{}
	}

	return &CreatePostSession{
		SessionID: sessionID,
		UserID:    userID,
		MediaIDs:  mediaIDs,
	}, nil
}

func (c *CreatePostSession) RequestMediaIDs(count int) (uuid.UUIDs, error) {
	if count < 1 {
		return nil, errs.NewError(errs.Invalid, nil, MediaCountInvalid)
	}

	var mediaIDs uuid.UUIDs
	for i := 0; i < count; i++ {
		mediaID, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("CreatePostSession.NewCreatePostSession: %w", err)
		}
		c.MediaIDs[mediaID] = struct{}{}
		mediaIDs = append(mediaIDs, mediaID)
	}
	return mediaIDs, nil
}
