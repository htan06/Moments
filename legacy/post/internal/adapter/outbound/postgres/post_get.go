package postgresadapter

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/htan06/Moments/post/internal/adapter/outbound/mapper"
	"github.com/htan06/Moments/post/internal/application/readmodel"
	"github.com/jackc/pgx/v5"
)

func (pr *PostgresPostRepository) getMediasByPostID(ctx context.Context, postID int64) ([]readmodel.Media, error) {
	mediasQuery := `SELECT 
						id AS m_id,
						type,
						media_key,
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
		return []readmodel.Media{}, fmt.Errorf("PostgresPostRepository.getMediasByPostID: %w", err)
	}
	defer mediaRows.Close()

	medias, err := pgx.CollectRows[readmodel.Media](mediaRows, pgx.RowToStructByName)
	if err != nil {
		return []readmodel.Media{}, fmt.Errorf("PostgresPostRepository.getMediasByPostID: %w", err)
	}
	return medias, nil
}

func (pr *PostgresPostRepository) getMentionsByPostID(ctx context.Context, postID int64) (map[int64]string, error) {
	queryMentions := `SELECT 
						u.user_id, u.username 
					FROM content.post_mentions pm
					JOIN content.authors u
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

func (pr *PostgresPostRepository) GetPostsByAuthorID(ctx context.Context, viewerID int64, authorID int64, cursor int64, size int) ([]readmodel.Post, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	likeQry, likeArgs, err := squirrel.Expr("EXISTS(SELECT 1 FROM content.post_likes WHERE user_id = ? AND post_id = id) AS is_liked", viewerID).ToSql()

	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
	}

	postQry := psql.Select(
		"id AS p_id",
		"thumbnail_key",
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
			"p.thumbnail_key",
			"p.content",
			"p.media_count",
			"p.like_count",
			"p.comment_count",
			"p.is_liked",
			"p.p_created_at",

			"m.id AS m_id",
			"m.type",
			"m.media_key",
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
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
	}

	rows, err := pr.conn.Query(ctx, qry, args...)
	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
	}
	defer rows.Close()

	postIndex := map[int64]int{}
	posts := make([]readmodel.Post, 0, size)
	for rows.Next() {
		post := readmodel.Post{}
		media := readmodel.Media{}

		if err := rows.Scan(
			&post.ID, &post.ThumbnailID, &post.Content, &post.MediaCount, &post.LikeCount, &post.CommentCount, &post.IsLiked, &post.CreatedAt,
			&media.ID, &media.Type, &media.MediaID, &media.DisplayOrder, &media.Width, &media.Height, &media.Duration, &media.Size, &media.CreatedAt,
		); err != nil {
			return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByAuthorID: %w", err)
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

func (p *PostgresPostRepository) GetLikeCount(ctx context.Context, postID int64) (int64, error) {
	qry := `SELECT like_count FROM content.posts where id = $1`

	var likeCount int64
	if err := p.conn.QueryRow(ctx, qry, postID).Scan(&likeCount); err != nil {
		return 0, fmt.Errorf("PostgresPostRepository.GetLikeCount: %w", err)
	}

	return likeCount, nil
}

func (p *PostgresPostRepository) GetPostsByFollowing(ctx context.Context, viewerID int64, cursor int64, size int) ([]readmodel.Post, error) {
	psqlBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	likeQry, likeArgs, err := squirrel.Expr("EXISTS(SELECT 1 FROM content.post_likes pl WHERE pl.user_id = ? AND pl.post_id = p.id) AS is_liked", viewerID).ToSql()

	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	selectPosts := psqlBuilder.
		Select("p.id AS p_id",
			"u.name AS author_name",
			"u.username AS author_username",
			"u.avatar_thumbnail_key AS author_avatar_thumbnail_key",
			"p.thumbnail_key",
			"p.content",
			"p.media_count",
			"p.like_count",
			"p.comment_count",
			// "is_liked",
			"p.created_at AS p_created_at",

			"m.id AS m_id",
			"m.type",
			"m.media_key",
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
		LeftJoin("content.authors AS a ON a.user_id = p.author_id").
		OrderBy("p.created_at DESC, m.display_order ASC")

	if cursor != 0 {
		selectPosts = selectPosts.Where("p.created_at < to_timestamp(?)", cursor)
	}

	qry, args, err := selectPosts.ToSql()
	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	rows, err := p.conn.Query(ctx, qry, args...)
	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}
	defer rows.Close()
	posts, err := mapper.RowsToPosts(rows)
	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	return posts, nil
}

func (p *PostgresPostRepository) GetBatchPosts(ctx context.Context, viewerID int64, ids []int64) ([]readmodel.Post, error) {
	psqlBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	likeQry, likeArgs, err := squirrel.Expr("EXISTS(SELECT 1 FROM content.post_likes pl WHERE pl.user_id = ? AND pl.post_id = p.id) AS is_liked", viewerID).ToSql()

	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	selectPosts := psqlBuilder.
		Select("p.id AS p_id",
			"p.thumbnail_key",
			"p.content",
			"p.media_count",
			"p.like_count",
			"p.comment_count",
			// "is_liked",
			"p.created_at AS p_created_at",

			"m.id AS m_id",
			"m.type",
			"m.media_key",
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
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	rows, err := p.conn.Query(ctx, qry, args...)
	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}
	defer rows.Close()

	posts, err := mapper.RowsToPosts(rows)
	if err != nil {
		return []readmodel.Post{}, fmt.Errorf("PostgresPostRepository.GetPostsByFollowing: %w", err)
	}

	return posts, nil
}
