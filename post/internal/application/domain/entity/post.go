package entity

import (
	"time"

	"github.com/htan06/Moments/post/internal/application/domain/vo"
)

type Visibility string

const (
	Private Visibility = "PRIVATE"
	Public  Visibility = "PUBLIC"
	Friend  Visibility = "FRIEND"
)

type Post struct {
	id           int64
	authorID     int64
	visibility   Visibility
	contents     []vo.Content
	aspectRatio  vo.AspectRatio
	thumbnailID  string
	medias       []Media
	mentions     []int64
	hashTags     []string
	likeCount    int
	mediaCount   int
	commentCount int
	createdAt    time.Time
	updatedAt    time.Time
}

func NewPost(
	authorID int64,
	visibility Visibility,
	contents []vo.Content,
	aspectRatioString string,
) (*Post, error) {

	aspectRatio, err := vo.NewAspectRatio(aspectRatioString)
	if err != nil {
		return nil, err
	}

	return &Post{
		authorID:     authorID,
		visibility:   visibility,
		contents:     contents,
		aspectRatio:  aspectRatio,
		medias:       make([]Media, 0),
		mentions:     make([]int64, 0),
		hashTags:     make([]string, 0),
		mediaCount:   0,
		likeCount:    0,
		commentCount: 0,
	}, nil
}

func (p *Post) AppendMention(mention int64) {
	p.mentions = append(p.mentions, mention)
}

func (p *Post) Appendhashtag(hashtag string) {
	p.hashTags = append(p.hashTags, hashtag)
}

func (p *Post) SetMedias(medias []Media) {
	p.medias = medias
	p.mediaCount = len(medias)
}

func (p *Post) SetThumbnailID(thumbnailID string) {
	p.thumbnailID = thumbnailID
}

func (p *Post) ID() int64 {
	return p.id
}

func (p *Post) AuthorID() int64 {
	return p.authorID
}

func (p *Post) Visibility() Visibility {
	return p.visibility
}

func (p *Post) Contents() []vo.Content {
	return p.contents
}

func (p *Post) AspectRatio() vo.AspectRatio {
	return p.aspectRatio
}

func (p *Post) ThumbnailID() string {
	return p.thumbnailID
}

func (p *Post) Medias() []Media {
	return p.medias
}

func (p *Post) Mentions() []int64 {
	return p.mentions
}

func (p *Post) HashTags() []string {
	return p.hashTags
}

func (p *Post) LikeCount() int {
	return p.likeCount
}

func (p *Post) MediaCount() int {
	return p.mediaCount
}

func (p *Post) CommentCount() int {
	return p.commentCount
}

func (p *Post) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Post) UpdatedAt() time.Time {
	return p.updatedAt
}
