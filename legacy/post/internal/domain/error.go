// Package domain holds what every aggregate of the Post context shares:
// business errors and the Event contract.
package domain

import (
	"errors"
	"time"
)

// Event is a fact that happened inside an aggregate. Event payloads are
// plain structs so adapters can serialize them however they need.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// Validation errors.
var (
	ErrUserIDInvalid       = errors.New("USER_ID_INVALID")
	ErrPostIDInvalid       = errors.New("POST_ID_INVALID")
	ErrVisibilityInvalid   = errors.New("VISIBILITY_INVALID")
	ErrAspectRatioInvalid  = errors.New("ASPECT_RATIO_INVALID")
	ErrPostStatusInvalid   = errors.New("POST_STATUS_INVALID")
	ErrThumbnailKeyEmpty   = errors.New("THUMBNAIL_KEY_EMPTY")
	ErrCounterInvalid      = errors.New("COUNTER_INVALID")
	ErrMediaKeyEmpty       = errors.New("MEDIA_KEY_EMPTY")
	ErrMediaTypeInvalid    = errors.New("MEDIA_TYPE_INVALID")
	ErrMediaStatusInvalid  = errors.New("MEDIA_STATUS_INVALID")
	ErrMediaCountInvalid   = errors.New("MEDIA_COUNT_INVALID")
	ErrMaxMediasExceeded   = errors.New("MAX_MEDIAS_EXCEEDED")
	ErrDisplayOrderInvalid = errors.New("DISPLAY_ORDER_INVALID")
	ErrWidthInvalid        = errors.New("WIDTH_INVALID")
	ErrHeightInvalid       = errors.New("HEIGHT_INVALID")
	ErrContentInvalid      = errors.New("CONTENT_INVALID")
	ErrNodeInvalid         = errors.New("NODE_INVALID")
	ErrContentHasHashtag   = errors.New("CONTENT_HAS_HASHTAG")
	ErrRelationshipInvalid = errors.New("RELATIONSHIP_INVALID")
	ErrSessionExpired      = errors.New("SESSION_EXPIRED")
	ErrPresignExpiry       = errors.New("PRESIGN_EXPIRY_INVALID")
	ErrMediaNotInSession   = errors.New("MEDIA_NOT_IN_SESSION")
)

// Invariant / business rule errors.
var (
	// INV-POST-01
	ErrDuplicateDisplayOrder = errors.New("DUPLICATE_DISPLAY_ORDER")
	ErrDuplicateMediaKey     = errors.New("DUPLICATE_MEDIA_KEY")
	// INV-POST-02 / INV-MEDIA-01 state transitions.
	ErrMediaAlreadyPublished = errors.New("MEDIA_ALREADY_PUBLISHED")
	ErrMediaNotFound         = errors.New("MEDIA_NOT_FOUND")
	// Only the author may change or delete a post.
	ErrNotPostOwner = errors.New("NOT_POST_OWNER")
	// BR-01: the post is not visible / not interactable for this user.
	ErrPostNotAccessible = errors.New("POST_NOT_ACCESSIBLE")
	// A repost is always of someone else's post.
	ErrRepostOwnPost = errors.New("REPOST_OWN_POST")
	// INV-LIKE-01 / INV-COMMENT-01 / INV-REPOST-01, returned by repositories.
	ErrAlreadyExists = errors.New("ALREADY_EXISTS")
	ErrPostNotFound  = errors.New("POST_NOT_FOUND")
	ErrNotFound      = errors.New("NOT_FOUND")
)

// EventMeta is embedded by every event to carry its occurrence time.
type EventMeta struct {
	At time.Time `json:"occurred_at"`
}

func NewEventMeta() EventMeta { return EventMeta{At: time.Now().UTC()} }

func (m EventMeta) OccurredAt() time.Time { return m.At }
