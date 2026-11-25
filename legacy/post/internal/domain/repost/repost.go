package repost

import (
	"context"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/post"
	"github.com/htan06/Moments/post/internal/domain/relationship"
)

// Repost is an aggregate root: a user re-sharing someone else's post.
// INV-REPOST-01: one Repost per (userID, postID), guaranteed by Repository.
type Repost struct {
	userID    int64
	postID    int64
	createdAt time.Time
}

func (r *Repost) UserID() int64        { return r.userID }
func (r *Repost) PostID() int64        { return r.postID }
func (r *Repost) CreatedAt() time.Time { return r.createdAt }

// RepostPost is the Repost domain service (post.md §5): it checks BR-01
// then creates the Repost. The reposter never owns the post.
func RepostPost(userID int64, p *post.Post, rel relationship.Relationship) (*Repost, Created, error) {
	if userID == p.AuthorID() {
		return nil, Created{}, domain.ErrRepostOwnPost
	}
	if err := p.CheckAccess(userID, rel); err != nil {
		return nil, Created{}, err
	}
	r := &Repost{userID: userID, postID: p.ID(), createdAt: time.Now().UTC()}
	return r, Created{EventMeta: domain.NewEventMeta(), UserID: userID, PostID: p.ID(), AuthorID: p.AuthorID()}, nil
}

// Rehydrate rebuilds a Repost from persistence.
func Rehydrate(userID, postID int64, createdAt time.Time) (*Repost, error) {
	if userID <= 0 {
		return nil, domain.ErrUserIDInvalid
	}
	if postID <= 0 {
		return nil, domain.ErrPostIDInvalid
	}
	return &Repost{userID: userID, postID: postID, createdAt: createdAt}, nil
}

type Created struct {
	domain.EventMeta
	UserID   int64 `json:"user_id"`
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

func (Created) EventName() string { return "repost.created" }

var _ domain.Event = Created{}

type Repository interface {
	// Save returns domain.ErrAlreadyExists when (userID, postID) already
	// exists (INV-REPOST-01).
	Save(ctx context.Context, r *Repost) error
}
