package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPostRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresPostRepository(conn *pgxpool.Pool) *PostgresPostRepository {
	return &PostgresPostRepository{
		conn: conn,
	}
}

func (pr *PostgresPostRepository) CreatePost(ctx context.Context, post domain.Post) (*int64, error) {
	tx, err := pr.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}
	defer tx.Rollback(ctx)

	insertPost := `INSERT INTO content.posts (author_id, content, visibility, thumbnail_id, media_count, aspect_ratio, like_count, comment_count)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;`

	postContent, err := json.Marshal(post.Contents)
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	var postID int64
	if err := tx.QueryRow(
		ctx,
		insertPost,
		post.AuthorID, postContent, post.Visibility, post.ThumbnailID, post.MediaCount, post.AspectRatio, post.LikeCount, post.CommentCount).
		Scan(&postID); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	insertMedia := `INSERT INTO content.medias (post_id, type, media_id, display_order, width, height) 
					VALUES ($1, $2, $3, $4, $5, $6)`

	insertMention := `INSERT INTO content.post_mentions (post_id, user_id) VALUES ($1, $2);`

	batch := &pgx.Batch{}

	for _, m := range post.Medias {
		batch.Queue(insertMedia, postID, m.Type, m.MediaID, m.DisplayOrder, m.Width, m.Height)
	}

	for _, userID := range post.Mentions {
		batch.Queue(insertMention, postID, userID)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}
	return &postID, nil
}

func (pr *PostgresPostRepository) GetPost(ctx context.Context, postID int64) (domain.PostReadModel, error) {
	postQuery := `SELECT 
					p.id,
					p.author_id,
					p.content,
					p.visibility,
					p.thumbnail_id,
					p.media_count,
					p.aspect_ratio,
					p.like_count,
					p.comment_count,
					p.created_at,
					p.updated_at,
					
					u.name AS author_name,
					u.username AS author_username,
					u.avatar_thumbnail_id AS author_avatar_thumbnail_id

				FROM content.posts p
				JOIN profile.users u 
					ON p.author_id = u.id
				WHERE p.id = $1;`

	postRows, err := pr.conn.Query(ctx, postQuery, postID)
	if err != nil {
		return domain.PostReadModel{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
	}
	defer postRows.Close()

	post, err := pgx.CollectOneRow[domain.PostReadModel](postRows, pgx.RowToStructByName)
	if err != nil {
		return domain.PostReadModel{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
	}

	medias, err := pr.getMediasByPostID(ctx, postID)
	if err != nil {
		return domain.PostReadModel{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
	}

	post.Medias = medias

	mentions, err := pr.getMentionsByPostID(ctx, postID)
	if err != nil {
		return domain.PostReadModel{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
	}

	if err := pr.enrichContent(post.Content, mentions); err != nil {
		return domain.PostReadModel{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
	}

	return post, nil
}

func (pr *PostgresPostRepository) getMediasByPostID(ctx context.Context, postID int64) ([]domain.MediaReadModel, error) {
	mediasQuery := `SELECT 
						id,
						type,
						media_id,
						display_order,
						width,
						height,
						duration,
						size,
						created_at
					FROM content.medias
					WHERE post_id = $1
					ORDER BY display_order ASC;`

	mediaRows, err := pr.conn.Query(ctx, mediasQuery, postID)
	if err != nil {
		return []domain.MediaReadModel{}, fmt.Errorf("PostgresPostRepository.getMediasByPostID: %w", err)
	}
	defer mediaRows.Close()

	medias, err := pgx.CollectRows[domain.MediaReadModel](mediaRows, pgx.RowToStructByName)
	if err != nil {
		return []domain.MediaReadModel{}, fmt.Errorf("PostgresPostRepository.getMediasByPostID: %w", err)
	}
	return medias, nil
}

type Mentions struct {
}

func (pr *PostgresPostRepository) getMentionsByPostID(ctx context.Context, postID int64) (map[int64]string, error) {
	queryMentions := `SELECT 
						u.id as user_id, u.username 
					FROM content.post_mentions pm
					JOIN profile.users u
						ON pm.user_id = u.id
					WHERE pm.post_id = $1;`

	mentionRows, err := pr.conn.Query(ctx, queryMentions, postID)
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.getMentionsByPostID: %w", err)
	}
	defer mentionRows.Close()

	mentions := make(map[int64]string, 0)
	for mentionRows.Next() {
		var id int64
		var username string

		err := mentionRows.Scan(&id, &username)
		if err != nil {
			return nil, fmt.Errorf("PostgresPostRepository.getMentionsByPostID: %w", err)
		}

		mentions[id] = username
	}
	return mentions, nil
}

func (pr *PostgresPostRepository) enrichContent(content []domain.Content, mentions map[int64]string) error {
	for i := range content {
		switch content[i].Type {
		case domain.Mention:
			uID, err := strconv.ParseInt(content[i].Value, 10, 64)
			if err != nil {
				return fmt.Errorf("PostgresPostRepository.enrichContent: %w", err)
			}

			username, ok := mentions[uID]
			if !ok {
				content[i].Value = "@deleted"
				continue
			}
			content[i].Value = "@" + username
		case domain.Hashtag:
			content[i].Value = "#" + content[i].Value
		}
	}
	return nil
}

func (pr *PostgresPostRepository) GetPostsByUsername(ctx context.Context, username string) ([]domain.PostSummary, error) {
	postQuery := `SELECT 
					id,
					thumbnail_id,
					media_count,
					like_count,
					comment_count
				FROM content.posts
				WHERE p.author_id = (SELECT id FROM profile.users WHERE username = $1);`

	rows, err := pr.conn.Query(ctx, postQuery, username)
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("GetPostByUsername.Get: %w", err)
	}
	defer rows.Close()

	posts, err := pgx.CollectRows[domain.PostSummary](rows, pgx.RowToStructByName)
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("GetPostByUsername.Get: %w", err)
	}
	return posts, nil
}
