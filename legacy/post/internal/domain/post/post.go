package post

import (
	"strings"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/content"
)

// MaxMedias is the maximum number of medias of a post (and of a Session,
// INV-SESSION-02).
const MaxMedias = 20

// Post is the aggregate root of a user's post.
type Post struct {
	id           int64
	authorID     int64
	content      content.Content
	visibility   Visibility
	aspectRatio  AspectRatio
	thumbnailKey string
	medias       []Media
	likeCount    int64
	commentCount int64
	repostCount  int64
	status       Status
	createdAt    time.Time
	updatedAt    time.Time
}

type CreateParams struct {
	AuthorID     int64
	Content      content.Content // may be null
	Visibility   Visibility
	AspectRatio  AspectRatio
	ThumbnailKey string
	Medias       []Media
}

// Create builds a new PENDING post. Its id comes from a sequence, so it
// is 0 until the repository assigns one through AssignID.
func Create(p CreateParams) (*Post, error) {
	now := time.Now().UTC()
	post := &Post{
		authorID:     p.AuthorID,
		content:      p.Content,
		visibility:   p.Visibility,
		aspectRatio:  p.AspectRatio,
		thumbnailKey: p.ThumbnailKey,
		medias:       append([]Media(nil), p.Medias...),
		status:       StatusPending,
		createdAt:    now,
		updatedAt:    now,
	}
	if err := post.validate(); err != nil {
		return nil, err
	}
	return post, nil
}

type RehydrateParams struct {
	ID           int64
	AuthorID     int64
	Content      content.Content
	Visibility   Visibility
	AspectRatio  AspectRatio
	ThumbnailKey string
	Medias       []Media
	LikeCount    int64
	CommentCount int64
	RepostCount  int64
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Rehydrate rebuilds a Post from persistence. It still checks the
// invariants so corrupted data never enters the domain.
func Rehydrate(p RehydrateParams) (*Post, error) {
	post := &Post{
		id:           p.ID,
		authorID:     p.AuthorID,
		content:      p.Content,
		visibility:   p.Visibility,
		aspectRatio:  p.AspectRatio,
		thumbnailKey: p.ThumbnailKey,
		medias:       append([]Media(nil), p.Medias...),
		likeCount:    p.LikeCount,
		commentCount: p.CommentCount,
		repostCount:  p.RepostCount,
		status:       p.Status,
		createdAt:    p.CreatedAt,
		updatedAt:    p.UpdatedAt,
	}
	if p.ID <= 0 {
		return nil, domain.ErrPostIDInvalid
	}
	if err := post.validate(); err != nil {
		return nil, err
	}
	return post, nil
}

func (p *Post) ID() int64                { return p.id }
func (p *Post) AuthorID() int64          { return p.authorID }
func (p *Post) Content() content.Content { return p.content }
func (p *Post) Visibility() Visibility   { return p.visibility }
func (p *Post) AspectRatio() AspectRatio { return p.aspectRatio }
func (p *Post) ThumbnailKey() string     { return p.thumbnailKey }
func (p *Post) Medias() []Media          { return append([]Media(nil), p.medias...) }
func (p *Post) LikeCount() int64         { return p.likeCount }
func (p *Post) CommentCount() int64      { return p.commentCount }
func (p *Post) RepostCount() int64       { return p.repostCount }
func (p *Post) Status() Status           { return p.status }
func (p *Post) CreatedAt() time.Time     { return p.createdAt }
func (p *Post) UpdatedAt() time.Time     { return p.updatedAt }
func (p *Post) IsPublished() bool        { return p.status == StatusPublished }

// AssignID sets the sequence id generated on persistence and emits
// Created. It can only be called once, on a new post.
func (p *Post) AssignID(id int64) (Created, error) {
	if id <= 0 || p.id != 0 {
		return Created{}, domain.ErrPostIDInvalid
	}
	p.id = id
	return Created{EventMeta: domain.NewEventMeta(), PostID: p.id, AuthorID: p.authorID}, nil
}

func (p *Post) UpdateContent(actorID int64, c content.Content) (ContentUpdated, error) {
	if err := p.ensureOwner(actorID); err != nil {
		return ContentUpdated{}, err
	}
	p.content = c
	p.updatedAt = time.Now().UTC()
	return ContentUpdated{EventMeta: domain.NewEventMeta(), PostID: p.id, AuthorID: p.authorID}, nil
}

func (p *Post) UpdateVisibility(actorID int64, v Visibility) (VisibilityUpdated, error) {
	if err := p.ensureOwner(actorID); err != nil {
		return VisibilityUpdated{}, err
	}
	if !v.Valid() {
		return VisibilityUpdated{}, domain.ErrVisibilityInvalid
	}
	p.visibility = v
	p.updatedAt = time.Now().UTC()
	return VisibilityUpdated{EventMeta: domain.NewEventMeta(), PostID: p.id, AuthorID: p.authorID, Visibility: v}, nil
}

// MarkMediaPublished publishes one processed media. When it was the last
// PROCESSING media the post moves to PUBLISHED (INV-POST-02) and the
// returned event is non-nil.
func (p *Post) MarkMediaPublished(mediaKey string, width, height int32) (*Published, error) {
	i := p.mediaIndex(mediaKey)
	if i < 0 {
		return nil, domain.ErrMediaNotFound
	}
	now := time.Now().UTC()
	if err := p.medias[i].markPublished(width, height, now); err != nil {
		return nil, err
	}
	p.updatedAt = now
	if p.status == StatusPending && p.allMediasPublished() {
		p.status = StatusPublished
		return &Published{EventMeta: domain.NewEventMeta(), PostID: p.id, AuthorID: p.authorID}, nil
	}
	return nil, nil
}

// Delete checks the actor may delete the post and emits Deleted; the
// repository removes it.
func (p *Post) Delete(actorID int64) (Deleted, error) {
	if err := p.ensureOwner(actorID); err != nil {
		return Deleted{}, err
	}
	keys := make([]string, 0, len(p.medias)+1)
	keys = append(keys, p.thumbnailKey)
	for _, m := range p.medias {
		keys = append(keys, m.mediaKey)
	}
	return Deleted{EventMeta: domain.NewEventMeta(), PostID: p.id, AuthorID: p.authorID, MediaKeys: keys}, nil
}

func (p *Post) ensureOwner(actorID int64) error {
	if actorID != p.authorID {
		return domain.ErrNotPostOwner
	}
	return nil
}

func (p *Post) mediaIndex(key string) int {
	for i := range p.medias {
		if p.medias[i].mediaKey == key {
			return i
		}
	}
	return -1
}

func (p *Post) allMediasPublished() bool {
	for _, m := range p.medias {
		if !m.IsPublished() {
			return false
		}
	}
	return true
}

func (p *Post) validate() error {
	if p.authorID <= 0 {
		return domain.ErrUserIDInvalid
	}
	if !p.visibility.Valid() {
		return domain.ErrVisibilityInvalid
	}
	if !p.aspectRatio.Valid() {
		return domain.ErrAspectRatioInvalid
	}
	if strings.TrimSpace(p.thumbnailKey) == "" {
		return domain.ErrThumbnailKeyEmpty
	}
	if !p.status.Valid() {
		return domain.ErrPostStatusInvalid
	}
	if p.likeCount < 0 || p.commentCount < 0 || p.repostCount < 0 {
		return domain.ErrCounterInvalid
	}
	if len(p.medias) == 0 {
		return domain.ErrMediaCountInvalid
	}
	if len(p.medias) > MaxMedias {
		return domain.ErrMaxMediasExceeded
	}
	orders := make(map[int32]struct{}, len(p.medias))
	keys := make(map[string]struct{}, len(p.medias))
	for _, m := range p.medias {
		if err := m.validate(); err != nil {
			return err
		}
		// INV-POST-01
		if _, dup := orders[m.displayOrder]; dup {
			return domain.ErrDuplicateDisplayOrder
		}
		orders[m.displayOrder] = struct{}{}
		if _, dup := keys[m.mediaKey]; dup {
			return domain.ErrDuplicateMediaKey
		}
		keys[m.mediaKey] = struct{}{}
	}
	// INV-POST-02
	if p.status == StatusPublished && !p.allMediasPublished() {
		return domain.ErrPostStatusInvalid
	}
	return nil
}
