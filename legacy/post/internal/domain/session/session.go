// Package session implements the Session aggregate: a post creation
// session that hands out the media keys to upload with presigned URLs.
package session

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/post"
)

// MaxMedias is INV-SESSION-02: imageKeys.size + videoKeys.size ≤ 20.
const MaxMedias = post.MaxMedias

type Session struct {
	id        uuid.UUID
	userID    int64
	imageKeys []string
	videoKeys []string
	expiresAt time.Time
}

// New opens a session for userID (the future post author) valid for ttl.
func New(userID int64, ttl time.Duration) (*Session, error) {
	if userID <= 0 {
		return nil, domain.ErrUserIDInvalid
	}
	if ttl <= 0 {
		return nil, domain.ErrSessionExpired
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("session.New: %w", err)
	}
	return &Session{id: id, userID: userID, expiresAt: time.Now().UTC().Add(ttl)}, nil
}

// Rehydrate rebuilds a Session from persistence.
func Rehydrate(id uuid.UUID, userID int64, imageKeys, videoKeys []string, expiresAt time.Time) (*Session, error) {
	if userID <= 0 {
		return nil, domain.ErrUserIDInvalid
	}
	if len(imageKeys)+len(videoKeys) > MaxMedias {
		return nil, domain.ErrMaxMediasExceeded
	}
	return &Session{
		id:        id,
		userID:    userID,
		imageKeys: slices.Clone(imageKeys),
		videoKeys: slices.Clone(videoKeys),
		expiresAt: expiresAt,
	}, nil
}

func (s *Session) ID() uuid.UUID        { return s.id }
func (s *Session) UserID() int64        { return s.userID }
func (s *Session) ImageKeys() []string  { return slices.Clone(s.imageKeys) }
func (s *Session) VideoKeys() []string  { return slices.Clone(s.videoKeys) }
func (s *Session) ExpiresAt() time.Time { return s.expiresAt }
func (s *Session) MediaCount() int      { return len(s.imageKeys) + len(s.videoKeys) }

func (s *Session) IsExpired(now time.Time) bool { return !now.Before(s.expiresAt) }

// RequestUploadKeys reserves count new keys of mediaType whose presigned
// URLs will expire at presignExpiresAt.
func (s *Session) RequestUploadKeys(mediaType post.MediaType, count int, presignExpiresAt, now time.Time) ([]string, error) {
	if s.IsExpired(now) {
		return nil, domain.ErrSessionExpired
	}
	if !mediaType.Valid() {
		return nil, domain.ErrMediaTypeInvalid
	}
	if count < 1 {
		return nil, domain.ErrMediaCountInvalid
	}
	// INV-SESSION-01
	if !presignExpiresAt.Before(s.expiresAt) {
		return nil, domain.ErrPresignExpiry
	}
	// INV-SESSION-02
	if s.MediaCount()+count > MaxMedias {
		return nil, domain.ErrMaxMediasExceeded
	}
	keys := make([]string, 0, count)
	for range count {
		key, err := newMediaKey(mediaType)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if mediaType == post.Image {
		s.imageKeys = append(s.imageKeys, keys...)
	} else {
		s.videoKeys = append(s.videoKeys, keys...)
	}
	return keys, nil
}

// MediaTypeOf returns the type of a key reserved in this session.
func (s *Session) MediaTypeOf(key string) (post.MediaType, error) {
	switch {
	case slices.Contains(s.imageKeys, key):
		return post.Image, nil
	case slices.Contains(s.videoKeys, key):
		return post.Video, nil
	}
	return "", domain.ErrMediaNotInSession
}

func newMediaKey(t post.MediaType) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("session.newMediaKey: %w", err)
	}
	if t == post.Image {
		return id.String() + ".jpeg", nil
	}
	return id.String() + "/index.m3u8", nil
}

type Repository interface {
	Save(ctx context.Context, s *Session) error
	// Get returns domain.ErrNotFound when no session id belongs to userID.
	Get(ctx context.Context, userID int64, id uuid.UUID) (*Session, error)
}
