package like

import (
	"context"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/post"
	"github.com/htan06/Moments/post/internal/domain/relationship"
)

// Like is an aggregate root.
// INV-LIKE-01: one Like per (userID, postID), guaranteed by Repository.
type Like struct {
	userID    int64
	postID    int64
	createdAt time.Time
}

func (l *Like) UserID() int64        { return l.userID }
func (l *Like) PostID() int64        { return l.postID }
func (l *Like) CreatedAt() time.Time { return l.createdAt }

// LikePost is the LikePost domain service (post.md §5): it checks BR-01
// then creates the Like.
func LikePost(userID int64, p *post.Post, rel relationship.Relationship) (*Like, Created, error) {
	if err := p.CheckAccess(userID, rel); err != nil {
		return nil, Created{}, err
	}
	l := &Like{userID: userID, postID: p.ID(), createdAt: time.Now().UTC()}
	return l, Created{EventMeta: domain.NewEventMeta(), UserID: userID, PostID: p.ID(), AuthorID: p.AuthorID()}, nil
}

// Rehydrate rebuilds a Like from persistence.
func Rehydrate(userID, postID int64, createdAt time.Time) (*Like, error) {
	if userID <= 0 {
		return nil, domain.ErrUserIDInvalid
	}
	if postID <= 0 {
		return nil, domain.ErrPostIDInvalid
	}
	return &Like{userID: userID, postID: postID, createdAt: createdAt}, nil
}

type Created struct {
	domain.EventMeta
	UserID   int64 `json:"user_id"`
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

func (Created) EventName() string { return "post.liked" }

var _ domain.Event = Created{}

type Repository interface {
	// Save returns domain.ErrAlreadyExists when (userID, postID) already
	// exists (INV-LIKE-01).
	Save(ctx context.Context, l *Like) error
}
