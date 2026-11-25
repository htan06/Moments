package postgresadapter

import (
	"github.com/htan06/Moments/post/internal/adapter/outbound/postgres/sqlc/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPostRepository struct {
	conn    *pgxpool.Pool
	queries *queries.Queries
}

func NewPostgresPostRepository(conn *pgxpool.Pool) *PostgresPostRepository {
	return &PostgresPostRepository{
		conn:    conn,
		queries: queries.New(conn),
	}
}

// func (pr *PostgresPostRepository) GetPost(ctx context.Context, postID int64) (readmodel.Post, error) {
// 	postQuery := `SELECT
// 					p.id,
// 					p.author_id,
// 					p.content,
// 					p.visibility,
// 					p.thumbnail_key,
// 					p.media_count,
// 					p.aspect_ratio,
// 					p.like_count,
// 					p.comment_count,
// 					p.created_at,
// 					p.updated_at,

// 					u.name AS author_name,
// 					u.username AS author_username,
// 					u.avatar_thumbnail_key AS author_avatar_thumbnail_key

// 				FROM content.posts p
// 				JOIN content.authors u
// 					ON p.author_id = u.user_id
// 				WHERE p.id = $1;`

// 	postRows, err := pr.conn.Query(ctx, postQuery, postID)
// 	if err != nil {
// 		return readmodel.Post{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
// 	}
// 	defer postRows.Close()

// 	post, err := pgx.CollectOneRow[readmodel.Post](postRows, pgx.RowToStructByName)
// 	if err != nil {
// 		return readmodel.Post{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
// 	}

// 	medias, err := pr.getMediasByPostID(ctx, postID)
// 	if err != nil {
// 		return readmodel.Post{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
// 	}

// 	post.Medias = medias

// 	mentions, err := pr.getMentionsByPostID(ctx, postID)
// 	if err != nil {
// 		return readmodel.Post{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
// 	}

// 	if err := pr.enrichContent(post.Content, mentions); err != nil {
// 		return readmodel.Post{}, fmt.Errorf("PostgresPostRepository.Get: %w", err)
// 	}

// 	return post, nil
// }

// func (pr *PostgresPostRepository) enrichContent(content []readmodel.Content, mentions map[int64]string) error {
// 	for i := range content {
// 		switch content[i].Type {
// 		case readmodel.Mention:
// 			uID, err := strconv.ParseInt(content[i].Text, 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("PostgresPostRepository.enrichContent: %w", err)
// 			}

// 			username, ok := mentions[uID]
// 			if !ok {
// 				content[i].Text = "@deleted"
// 				continue
// 			}
// 			content[i].Text = "@" + username
// 		case readmodel.Hashtag:
// 			content[i].Text = "#" + content[i].Text
// 		}
// 	}
// 	return nil
// }

// func (pr *PostgresPostRepository) GetRepostsByUsername(ctx context.Context, username string) ([]readmodel.Post, error) {
// 	qry := `SELECT
// 					p.id,
// 					p.thumbnail_key,
// 					p.media_count,
// 					p.like_count,
// 					p.comment_count,
// 					rp.created_at
// 				FROM content.posts p
// 				RIGHT JOIN content.reposts rp
// 					ON p.id = rp.post_id
// 				WHERE rp.user_id = (SELECT user_id FROM content.authors WHERE username = $1)
// 				ORDER BY rp.created_at DESC;`

// 	rows, err := pr.conn.Query(ctx, qry, username)
// 	if err != nil {
// 		return []readmodel.Post{}, fmt.Errorf("GetPostByUsername.GetPostsByUsername: %w", err)
// 	}
// 	defer rows.Close()

// 	posts, err := pgx.CollectRows[readmodel.Post](rows, pgx.RowToStructByName)
// 	if err != nil {
// 		return []readmodel.Post{}, fmt.Errorf("GetPostByUsername.GetPostsByUsername: %w", err)
// 	}
// 	return posts, nil
// }

// func (p *PostgresPostRepository) CreateRepost(ctx context.Context, userID int64, postID int64) (int64, error) {
// 	qry := `INSERT INTO content.reposts (user_id, post_id) VALUES ($1, $2) RETURNING id;`

// 	var id int64
// 	if err := p.conn.QueryRow(ctx, qry, userID, postID).Scan(&id); err != nil {
// 		return 0, fmt.Errorf("PostgresPostRepository.CreateRepost: %w", err)
// 	}
// 	return id, nil
// }

// func (p *PostgresPostRepository) DeleteRepost(ctx context.Context, userID int64, repostID int64) error {
// 	qry := `DELETE FROM content.posts WHERE user_id = $1 AND id = $2;`
// 	tag, err := p.conn.Exec(ctx, qry, userID, repostID)
// 	if err != nil {
// 		return fmt.Errorf("PostgresPostRepository.DeleteRepost: %w", err)
// 	}
// 	if tag.RowsAffected() != 1 {
// 		return fmt.Errorf("PostgresPostRepository.DeleteRepost: %s", "Error delete like post")
// 	}
// 	return nil
// }

// func (p *PostgresPostRepository) UpadateBatchCommentCount(ctx context.Context, list map[int64]int64) error {
// }

// func (p *PostgresPostRepository) UpadateBatchRepostCount(ctx context.Context, list map[int64]int64) error {
// }
