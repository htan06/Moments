package comment

import (
	"context"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/content"
	"github.com/htan06/Moments/post/internal/domain/post"
	"github.com/htan06/Moments/post/internal/domain/relationship"
)

// Comment is an aggregate root.
// INV-COMMENT-01: one Comment per (userID, postID), guaranteed by
// Repository (see OQ-04).
// INV-COMMENT-02: content contains no hashtag node.
type Comment struct {
	userID    int64
	postID    int64
	content   content.Content
	createdAt time.Time
}

func (c *Comment) UserID() int64            { return c.userID }
func (c *Comment) PostID() int64            { return c.postID }
func (c *Comment) Content() content.Content { return c.content }
func (c *Comment) CreatedAt() time.Time     { return c.createdAt }

// CommentPost is the Comment domain service (post.md §5): it checks BR-01
// then creates the Comment.
func CommentPost(userID int64, p *post.Post, rel relationship.Relationship, c content.Content) (*Comment, Created, error) {
	if err := validateContent(c); err != nil {
		return nil, Created{}, err
	}
	if err := p.CheckAccess(userID, rel); err != nil {
		return nil, Created{}, err
	}
	cm := &Comment{userID: userID, postID: p.ID(), content: c, createdAt: time.Now().UTC()}
	return cm, Created{EventMeta: domain.NewEventMeta(), UserID: userID, PostID: p.ID(), AuthorID: p.AuthorID()}, nil
}

// Rehydrate rebuilds a Comment from persistence.
func Rehydrate(userID, postID int64, c content.Content, createdAt time.Time) (*Comment, error) {
	if userID <= 0 {
		return nil, domain.ErrUserIDInvalid
	}
	if postID <= 0 {
		return nil, domain.ErrPostIDInvalid
	}
	if err := validateContent(c); err != nil {
		return nil, err
	}
	return &Comment{userID: userID, postID: postID, content: c, createdAt: createdAt}, nil
}

func validateContent(c content.Content) error {
	if c.IsNull() {
		return domain.ErrContentInvalid
	}
	if c.HasHashtag() {
		return domain.ErrContentHasHashtag
	}
	return nil
}

type Created struct {
	domain.EventMeta
	UserID   int64 `json:"user_id"`
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

func (Created) EventName() string { return "comment.created" }

var _ domain.Event = Created{}

type Repository interface {
	// Save returns domain.ErrAlreadyExists when (userID, postID) already
	// exists (INV-COMMENT-01).
	Save(ctx context.Context, c *Comment) error
}
