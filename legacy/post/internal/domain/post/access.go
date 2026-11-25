package post

import (
	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/relationship"
)

// CheckAccess enforces BR-01: whether userID may see / interact with
// the post (Post, Like, Comment, Repost). rel is the relationship read
// from userID towards the author.
func (p *Post) CheckAccess(userID int64, rel relationship.Relationship) error {
	if userID <= 0 {
		return domain.ErrUserIDInvalid
	}
	if !rel.Valid() {
		return domain.ErrRelationshipInvalid
	}
	// 3. Status must be PUBLISHED.
	if p.status != StatusPublished {
		return domain.ErrPostNotAccessible
	}
	if userID == p.authorID {
		return nil
	}
	// 2. Relationship between the two users.
	if rel.IsBlocked() {
		return domain.ErrPostNotAccessible
	}
	// 1. Visibility.
	if p.visibility != Public {
		return domain.ErrPostNotAccessible
	}
	return nil
}
