package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Masterminds/squirrel"
	"github.com/htan06/Moments/post/internal/domain"
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

	postContent, err := json.Marshal(post.Contents())
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	fmt.Printf("Log: %s\n", post.AspectRatio().String())

	var postID int64
	if err := tx.QueryRow(
		ctx,
		insertPost,
		post.AuthorID(), postContent, post.Visibility(), post.ThumbnailID(), post.MediaCount(), post.AspectRatio().String(), post.LikeCount(), post.CommentCount()).
		Scan(&postID); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	insertMedia := `INSERT INTO content.medias (post_id, type, media_id, display_order, width, height) 
					VALUES ($1, $2, $3, $4, $5, $6)`

	insertMention := `INSERT INTO content.post_mentions (post_id, user_id) VALUES ($1, $2);`

	batch := &pgx.Batch{}

	for _, m := range post.Medias() {
		batch.Queue(insertMedia, postID, m.Type, m.MediaID, m.DisplayOrder, m.Width, m.Height)
	}

	for _, userID := range post.Mentions() {
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
					ON p.author_id = u.user_id
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
						id AS m_id,
						type,
						media_id,
						display_order,
						width,
						height,
						duration,
						size,
						created_at AS m_created_at
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

func (pr *PostgresPostRepository) getMentionsByPostID(ctx context.Context, postID int64) (map[int64]string, error) {
	queryMentions := `SELECT 
						u.user_id, u.username 
					FROM content.post_mentions pm
					JOIN profile.users u
						ON pm.user_id = u.user_id
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
			uID, err := strconv.ParseInt(content[i].Text, 10, 64)
			if err != nil {
				return fmt.Errorf("PostgresPostRepository.enrichContent: %w", err)
			}

			username, ok := mentions[uID]
			if !ok {
				content[i].Text = "@deleted"
				continue
			}
			content[i].Text = "@" + username
		case domain.Hashtag:
			content[i].Text = "#" + content[i].Text
		}
	}
	return nil
}

func (pr *PostgresPostRepository) DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error {
	query := `DELETE FROM content.posts WHERE id = $1 AND author_id = $2;`

	_, err := pr.conn.Exec(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.DeletePostByUserIDAndPostID: %w", err)
	}
	return nil
}

func (pr *PostgresPostRepository) GetPostsByAuthorID(ctx context.Context, viewerID int64, authorID int64, cursor int64, size int) ([]domain.PostSummary, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	likeQry, likeArgs, err := squirrel.Expr("EXISTS(SELECT 1 FROM content.post_likes WHERE user_id = ? AND post_id = id) AS is_liked", viewerID).ToSql()

	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
	}

	postQry := psql.Select(
		"id AS p_id",
		"thumbnail_id",
		"content",
		"media_count",
		"like_count",
		"comment_count",
		"created_at AS p_created_at").
		Column(likeQry, likeArgs...).
		From("content.posts").
		Where(squirrel.Eq{"author_id": authorID}).
		OrderBy("created_at DESC").
		Limit(uint64(size))

	if cursor != 0 {
		postQry = postQry.Where("created_at < to_timestamp(?)", cursor)
	}

	qryPostMediaBuilder := psql.
		Select(
			"p.p_id",
			"p.thumbnail_id",
			"p.content",
			"p.media_count",
			"p.like_count",
			"p.comment_count",
			"p.is_liked",
			"p.p_created_at",

			"m.id AS m_id",
			"m.type",
			"m.media_id",
			"m.display_order",
			"m.width",
			"m.height",
			"m.duration",
			"m.size",
			"m.created_at AS media_created_at").
		FromSelect(postQry, "p").
		LeftJoin("content.medias AS m ON m.post_id = p.p_id").
		OrderBy("p.p_created_at DESC, m.display_order ASC")

	qry, args, err := qryPostMediaBuilder.ToSql()
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
	}

	rows, err := pr.conn.Query(ctx, qry, args...)
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
	}
	defer rows.Close()

	postIndex := map[int64]int{}
	posts := make([]domain.PostSummary, 0, size)
	for rows.Next() {
		post := domain.PostSummary{}
		media := domain.MediaReadModel{}

		if err := rows.Scan(
			&post.ID, &post.ThumbnailID, &post.Content, &post.MediaCount, &post.LikeCount, &post.CommentCount, &post.IsLiked, &post.CreatedAt,
			&media.ID, &media.Type, &media.MediaID, &media.DisplayOrder, &media.Width, &media.Height, &media.Duration, &media.Size, &media.CreatedAt,
		); err != nil {
			return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
		}

		index, exists := postIndex[post.ID]
		if !exists {
			posts = append(posts, post)
			index = len(posts) - 1
			postIndex[post.ID] = index
		}

		posts[index].Medias = append(posts[index].Medias, media)
	}

	return posts, nil
}

func (p *PostgresPostRepository) CreateLikePost(ctx context.Context, userID int64, postID int64) error {
	qry := `INSERT INTO content.post_likes (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	tag, err := p.conn.Exec(ctx, qry, userID, postID)

	if err != nil {
		return fmt.Errorf("PostgresPostRepository.CreateLikePost: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresPostRepository.CreateLikePost: %s", "Error insert like post")
	}
	return nil
}

func (p *PostgresPostRepository) DeleteLikePost(ctx context.Context, userID int64, postID int64) error {
	qry := `DELETE FROM content.post_likes WHERE user_id = $1 AND post_id = $2;`
	tag, err := p.conn.Exec(ctx, qry, userID, postID)
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.DeleteLikePost: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresPostRepository.DeleteLikePost: %s", "Error delete like post")
	}
	return nil
}

func (pr *PostgresPostRepository) GetRepostsByUsername(ctx context.Context, username string) ([]domain.PostSummary, error) {
	qry := `SELECT 
					p.id,
					p.thumbnail_id,
					p.media_count,
					p.like_count,
					p.comment_count,
					rp.created_at
				FROM content.posts p
				RIGHT JOIN content.reposts rp 
					ON p.id = rp.post_id
				WHERE rp.user_id = (SELECT user_id FROM profile.users WHERE username = $1) 
				ORDER BY rp.created_at DESC;`

	rows, err := pr.conn.Query(ctx, qry, username)
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("GetPostByUsername.GetPostsByUsername: %w", err)
	}
	defer rows.Close()

	posts, err := pgx.CollectRows[domain.PostSummary](rows, pgx.RowToStructByName)
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("GetPostByUsername.GetPostsByUsername: %w", err)
	}
	return posts, nil
}

func (p *PostgresPostRepository) CreateRepost(ctx context.Context, userID int64, postID int64) (int64, error) {
	qry := `INSERT INTO content.reposts (user_id, post_id) VALUES ($1, $2) RETURNING id;`

	var id int64
	if err := p.conn.QueryRow(ctx, qry, userID, postID).Scan(&id); err != nil {
		return 0, fmt.Errorf("PostgresPostRepository.CreateRepost: %w", err)
	}
	return id, nil
}

func (p *PostgresPostRepository) DeleteRepost(ctx context.Context, userID int64, repostID int64) error {
	qry := `DELETE FROM content.posts WHERE user_id = $1 AND id = $2;`
	tag, err := p.conn.Exec(ctx, qry, userID, repostID)
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.DeleteRepost: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresPostRepository.DeleteRepost: %s", "Error delete like post")
	}
	return nil
}

func (p *PostgresPostRepository) UpadateBatchLikeCount(ctx context.Context, list map[int64]int64) error {
	psqlStatement := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	expr := squirrel.Case()
	for postID, likeCount := range list {
		expr = expr.When(squirrel.Eq{"id": postID}, squirrel.Expr("like_count + ?", likeCount))
	}
	expr = expr.Else("like_count")

	qryBuilder := psqlStatement.Update("content.posts").Set("like_count", expr)
	qry, args, err := qryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.UpadateBatchLikeCount: %w", err)
	}

	if _, err := p.conn.Exec(ctx, qry, args...); err != nil {
		return fmt.Errorf("PostgresPostRepository.UpadateBatchLikeCount: %w", err)
	}

	// if tag.RowsAffected() != int64(len(list)) {
	// 	return fmt.Errorf("PostgresPostRepository.UpadateBatchLikeCount: %s", "Error update like count")
	// }
	return nil
}

// func (p *PostgresPostRepository) UpadateBatchCommentCount(ctx context.Context, list map[int64]int64) error {
// }

// func (p *PostgresPostRepository) UpadateBatchRepostCount(ctx context.Context, list map[int64]int64) error {
// }
func (p *PostgresPostRepository) GetLikeCount(ctx context.Context, postID int64) (int64, error) {
	qry := `SELECT like_count FROM content.posts where id = $1`

	var likeCount int64
	if err := p.conn.QueryRow(ctx, qry, postID).Scan(&likeCount); err != nil {
		return 0, fmt.Errorf("PostgresPostRepository.GetLikeCount: %w", err)
	}

	return likeCount, nil
}

func (p *PostgresPostRepository) GetPostsByFollowing(ctx context.Context, viewerID int64, cursor int64, size int) ([]domain.PostAuth, error) {
	psqlBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	likeQry, likeArgs, err := squirrel.Expr("EXISTS(SELECT 1 FROM content.post_likes pl WHERE pl.user_id = ? AND pl.post_id = p.id) AS is_liked", viewerID).ToSql()

	if err != nil {
		return []domain.PostAuth{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	selectPosts := psqlBuilder.
		Select("p.id AS p_id",
			"u.name AS author_name",
			"u.username AS author_username",
			"u.avatar_thumbnail_id AS author_avatar_thumbnail_id",
			"p.thumbnail_id",
			"p.content",
			"p.media_count",
			"p.like_count",
			"p.comment_count",
			// "is_liked",
			"p.created_at AS p_created_at",

			"m.id AS m_id",
			"m.type",
			"m.media_id",
			"m.display_order",
			"m.width",
			"m.height",
			"m.duration",
			"m.size",
			"m.created_at AS m_created_at").
		Column(likeQry, likeArgs...).
		FromSelect(psqlBuilder.
			Select("following_id").
			From("social.follows").
			Where(squirrel.Eq{"follower_id": viewerID}), "f").
		Join("content.posts AS p ON p.author_id = f.following_id").
		LeftJoin("content.medias AS m ON m.post_id = p.id").
		LeftJoin("profile.users AS u ON u.user_id = p.author_id").
		OrderBy("p.created_at DESC, m.display_order ASC")

	if cursor != 0 {
		selectPosts = selectPosts.Where("p.created_at < to_timestamp(?)", cursor)
	}

	qry, args, err := selectPosts.ToSql()
	if err != nil {
		return []domain.PostAuth{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	rows, err := p.conn.Query(ctx, qry, args...)
	if err != nil {
		return []domain.PostAuth{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}
	defer rows.Close()

	postIndex := map[int64]int{}
	posts := make([]domain.PostAuth, 0, size)
	for rows.Next() {
		post := domain.PostAuth{}
		media := domain.MediaReadModel{}

		if err := rows.Scan(
			&post.ID, &post.AuthorName, &post.AuthorUsername, &post.AuthorAvatarThumbnailID, &post.ThumbnailID, &post.Content, &post.MediaCount, &post.LikeCount, &post.CommentCount,  &post.CreatedAt,
			&media.ID, &media.Type, &media.MediaID, &media.DisplayOrder, &media.Width, &media.Height, &media.Duration, &media.Size, &media.CreatedAt, &post.IsLiked,
		); err != nil {
			return []domain.PostAuth{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
		}

		index, exists := postIndex[post.ID]
		if !exists {
			posts = append(posts, post)
			index = len(posts) - 1
			postIndex[post.ID] = index
		}

		posts[index].Medias = append(posts[index].Medias, media)
	}

	return posts, nil
}

func (p *PostgresPostRepository) GetBatchPosts(ctx context.Context, viewerID int64, ids []int64) ([]domain.PostSummary, error) {
	psqlBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	likeQry, likeArgs, err := squirrel.Expr("EXISTS(SELECT 1 FROM content.post_likes pl WHERE pl.user_id = ? AND pl.post_id = p.id) AS is_liked", viewerID).ToSql()

	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	selectPosts := psqlBuilder.
		Select("p.id AS p_id",
			"p.thumbnail_id",
			"p.content",
			"p.media_count",
			"p.like_count",
			"p.comment_count",
			// "is_liked",
			"p.created_at AS p_created_at",

			"m.id AS m_id",
			"m.type",
			"m.media_id",
			"m.display_order",
			"m.width",
			"m.height",
			"m.duration",
			"m.size",
			"m.created_at AS m_created_at").
		Column(likeQry, likeArgs...).
		From("content.posts AS p").
		LeftJoin("content.medias AS m ON m.post_id = p.id").
		OrderBy("p.created_at DESC, m.display_order ASC").
		Where(squirrel.Eq{"id": ids})


	qry, args, err := selectPosts.ToSql()
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	rows, err := p.conn.Query(ctx, qry, args...)
	if err != nil {
		return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}
	defer rows.Close()

	postIndex := map[int64]int{}
	posts := make([]domain.PostSummary, 0)
	for rows.Next() {
		post := domain.PostSummary{}
		media := domain.MediaReadModel{}

		if err := rows.Scan(
			&post.ID, &post.ThumbnailID, &post.Content, &post.MediaCount, &post.LikeCount, &post.CommentCount,  &post.CreatedAt,
			&media.ID, &media.Type, &media.MediaID, &media.DisplayOrder, &media.Width, &media.Height, &media.Duration, &media.Size, &media.CreatedAt, &post.IsLiked,
		); err != nil {
			return []domain.PostSummary{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
		}

		index, exists := postIndex[post.ID]
		if !exists {
			posts = append(posts, post)
			index = len(posts) - 1
			postIndex[post.ID] = index
		}

		posts[index].Medias = append(posts[index].Medias, media)
	}

	return posts, nil
}
