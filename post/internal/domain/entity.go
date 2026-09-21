package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/htan06/Moments/internal/errs"
)

type NodeType string
type MediaType string
type Visibility string

type AspectRatio struct {
	Width  int
	Height int
}

func NewAspectRatio(value string) (AspectRatio, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			nil,
			AspectRatioInvalid,
		)
	}

	Width, err := strconv.Atoi(parts[0])
	if err != nil {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			err,
			AspectRatioInvalid,
		)
	}

	Height, err := strconv.Atoi(parts[1])
	if err != nil {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			err,
			AspectRatioInvalid,
		)
	}

	ratio := AspectRatio{
		Width:  Width,
		Height: Height,
	}

	if !ratio.IsSupported() {
		return AspectRatio{}, errs.NewError(
			errs.Invalid,
			nil,
			AspectRatioInvalid,
		)
	}

	return ratio, nil
}

func (a AspectRatio) String() string {
	return fmt.Sprintf("%d:%d", a.Width, a.Height)
}

var supportedAspectRatios = map[AspectRatio]struct{}{
	{Width: 1, Height: 1}:  {},
	{Width: 3, Height: 4}:  {},
	{Width: 3, Height: 5}:  {},
	{Width: 4, Height: 3}:  {},
	{Width: 5, Height: 3}:  {},
	{Width: 16, Height: 9}: {},
}

func (a AspectRatio) Demensions() (widht int, Height int) {
	return a.Width, a.Height
}

func (a AspectRatio) IsSupported() bool {
	_, ok := supportedAspectRatios[a]
	return ok
}

const (
	Text    NodeType = "TEXT"
	Mention NodeType = "MENTION"
	Hashtag NodeType = "HASHTAG"
)

const (
	Image MediaType = "IMAGE"
	Video MediaType = "VIDEO"
)

const (
	Private Visibility = "PRIVATE"
	Public  Visibility = "PUBLIC"
	Friend  Visibility = "FRIEND"
)

type Content struct {
	Type NodeType `json:"type"`
	Text string   `json:"text"`
}

type Media struct {
	ID           int64
	MediaID      string
	Type         MediaType
	DisplayOrder int
	Width        int
	Height       int
	Duration     time.Duration
	Size         int
}

type Post struct {
	id           int64
	authorID     int64
	visibility   Visibility
	contents     []Content
	aspectRatio  AspectRatio
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
	contents []Content,
	aspectRatioString string,
) (*Post, error) {

	aspectRatio, err := NewAspectRatio(aspectRatioString)
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

func (p *Post) Contents() []Content {
	return p.contents
}

func (p *Post) AspectRatio() AspectRatio {
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

type Repost struct {
	id        int64
	userID    int64
	repostID  int64
	createdAt time.Time
}

func NewRepost(
	id int64,
	userID int64,
	repostID int64,
	createdAt time.Time,
) *Repost {
	return &Repost{
		id:        id,
		userID:    userID,
		repostID:  repostID,
		createdAt: createdAt,
	}
}

func (r *Repost) ID() int64 {
	return r.id
}

func (r *Repost) SetID(id int64) {
	r.id = id
}

func (r *Repost) UserID() int64 {
	return r.userID
}

func (r *Repost) SetUserID(userID int64) {
	r.userID = userID
}

func (r *Repost) RepostID() int64 {
	return r.repostID
}

func (r *Repost) SetRepostID(repostID int64) {
	r.repostID = repostID
}

func (r *Repost) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Repost) SetCreatedAt(createdAt time.Time) {
	r.createdAt = createdAt
}
