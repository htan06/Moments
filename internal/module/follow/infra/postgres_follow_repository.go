package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/follow/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresFollowRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresFollowRepository(conn *pgxpool.Pool) *PostgresFollowRepository {
	return &PostgresFollowRepository{
		conn: conn,
	}
}

func (pfr *PostgresFollowRepository) CreateFollow(ctx context.Context, follow domain.Follow) (*int64, error) {
	insertQuery := `INSERT INTO social.follows (follower_id, following_id) 
				VALUES ($1, $2) 
				RETURNING id;`

	var id int64
	if err := pfr.conn.QueryRow(ctx, insertQuery, follow.FollowerID, follow.FollowingID).Scan(&id); err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.CreateFollow: %w", err)

	}

	return &id, nil
}

func (pfr *PostgresFollowRepository) GetFollowing(ctx context.Context, username string, limit int32, offset int32) ([]domain.UserSummary, error) {
	var userID int64
	if err := pfr.conn.QueryRow(ctx, "SELECT user_id FROM profile.users WHERE username = $1;", username).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewError(errs.NotFound, nil, domain.UserNotFound)
		}
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	query := `SELECT f.id as follow_id, u.user_id, u.username, u.name, u.avatar_thumbnail_id, f.created_at
				FROM social.follows f
				JOIN profile.users u
					ON f.following_id = u.user_id
				WHERE f.follower_id = $1
				LIMIT $2
				OFFSET $3;`

	rows, err := pfr.conn.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows[domain.UserSummary](rows, pgx.RowToStructByName)

	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	return users, nil
}

func (pfr *PostgresFollowRepository) GetFollowers(ctx context.Context, username string, limit int32, offset int32) ([]domain.UserSummary, error) {
	var userID int64
	if err := pfr.conn.QueryRow(ctx, "SELECT user_id FROM profile.users WHERE username = $1;", username).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewError(errs.NotFound, nil, domain.UserNotFound)
		}
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	query := `SELECT f.id as follow_id, u.user_id, u.username, u.name, u.avatar_thumbnail_id, f.created_at
				FROM social.follows f
				JOIN profile.users u
					ON f.follower_id = u.user_id
				WHERE f.following_id = $1
				LIMIT $2
				OFFSET $3;`

	rows, err := pfr.conn.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows[domain.UserSummary](rows, pgx.RowToStructByName)

	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	return users, nil
}

func (pfr *PostgresFollowRepository) RemoveByFollowerID(ctx context.Context, followID int64, followerID int64) (domain.Follow, error) {
	query := `DELETE FROM social.follows WHERE id = $1 AND follower_id = $2 RETURNING follower_id, following_id;`

	var follow domain.Follow
	if err := pfr.conn.QueryRow(ctx, query, followID, followerID).Scan(&follow.FollowerID, &follow.FollowingID); err != nil {
		return domain.Follow{}, fmt.Errorf("PostgresFollowRepository.RemoveByFollowerID: %w", err)
	}
	return follow, nil
}

func (pfr *PostgresFollowRepository) RemoveByFollowingID(ctx context.Context, followID int64, followingID int64) (domain.Follow, error) {
	query := `DELETE FROM social.follows WHERE id = $1 AND following_id = $2 RETURNING follower_id, following_id;`

	var follow domain.Follow
	if err := pfr.conn.QueryRow(ctx, query, followID, followingID).Scan(&follow.FollowerID, &follow.FollowingID); err != nil {
		return domain.Follow{}, fmt.Errorf("PostgresFollowRepository.RemoveByFollowerID: %w", err)
	}

	return follow, nil
}
