package post

import "context"

// Repository persists the Post aggregate (with its medias).
type Repository interface {
	// Create inserts a new post and returns the sequence id; the caller
	// then calls Post.AssignID.
	Create(ctx context.Context, p *Post) (int64, error)
	// FindByID returns domain.ErrPostNotFound when the post does not exist.
	FindByID(ctx context.Context, id int64) (*Post, error)
	Update(ctx context.Context, p *Post) error
	Delete(ctx context.Context, id int64) error
}
