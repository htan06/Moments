package outbound

import (
	"context"
	"errors"
	"fmt"
	"time"
	"user-service/internal/adapter/outbound/postgres/sqlc/queries"
	"user-service/internal/application/domain"
	"user-service/internal/application/domain/entity"
	"user-service/internal/application/domain/readmodel"
	"user-service/internal/errs"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresFollowRepository struct {
	conn    *pgxpool.Pool
	queries *queries.Queries
}

func NewPostgresFollowRepository(conn *pgxpool.Pool) *PostgresFollowRepository {
	return &PostgresFollowRepository{
		conn:    conn,
		queries: queries.New(conn),
	}
}

func (pfr *PostgresFollowRepository) CreateFollow(ctx context.Context, follow entity.Follow) error {
	if err := pfr.queries.CreateFollow(ctx, queries.CreateFollowParams{
		FollowerID:  follow.FollowerID(),
		FollowingID: follow.FollowingID(),
	}); err != nil {
		return fmt.Errorf("PostgresFollowRepository.CreateFollow: %w", err)
	}
	return nil
}

func (pfr *PostgresFollowRepository) GetFollowing(ctx context.Context, username string, limit int32, cursor int64) ([]readmodel.UserSummary, error) {
	var userID int64
	if err := pfr.conn.QueryRow(ctx, "SELECT user_id FROM profile.users WHERE username = $1;", username).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewError(errs.NotFound, nil, domain.UserNotFound)
		}
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	qry := psql.Select("u.user_id", "u.username", "u.name", "u.avatar_thumbnail_key", "f.created_at").
		From("social.follows f").
		Join("profile.users u ON f.following_id = u.user_id").
		Where(squirrel.Eq{"f.follower_id": userID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit))

	if cursor != 0 {
		qry = qry.Where(squirrel.Lt{"f.created_at": time.UnixMicro(cursor)})
	}

	sql, args, err := qry.ToSql()
	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowing: %w", err)
	}

	rows, err := pfr.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowing: %w", err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows[readmodel.UserSummary](rows, pgx.RowToStructByName)

	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowing: %w", err)
	}

	return users, nil
}

func (pfr *PostgresFollowRepository) GetFollowers(ctx context.Context, username string, limit int32, cursor int64) ([]readmodel.UserSummary, error) {
	var userID int64
	if err := pfr.conn.QueryRow(ctx, "SELECT user_id FROM profile.users WHERE username = $1;", username).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NewError(errs.NotFound, nil, domain.UserNotFound)
		}
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	qry := psql.Select("u.user_id", "u.username", "u.name", "u.avatar_thumbnail_key", "f.created_at").
		From("social.follows f").
		Join("profile.users u ON f.follower_id = u.user_id").
		Where(squirrel.Eq{"f.following_id": userID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit))

	if cursor != 0 {
		qry = qry.Where(squirrel.Lt{"f.created_at": time.UnixMicro(cursor)})
	}

	sql, args, err := qry.ToSql()
	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowing: %w", err)
	}

	rows, err := pfr.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows[readmodel.UserSummary](rows, pgx.RowToStructByName)

	if err != nil {
		return nil, fmt.Errorf("PostgresFollowRepository.GetFollowers: %w", err)
	}

	return users, nil
}

func (pfr *PostgresFollowRepository) RemoveByIDAndFollowerID(ctx context.Context, f entity.Follow) error {
	err := pfr.queries.RemoveByFollowerIDAndFollowingID(ctx, queries.RemoveByFollowerIDAndFollowingIDParams{
		FollowerID:  f.FollowerID(),
		FollowingID: f.FollowingID(),
	})
	if err != nil {
		return fmt.Errorf("PostgresFollowRepository.RemoveByFollowerID: %w", err)
	}
	return nil
}

func (pfr *PostgresFollowRepository) RemoveByIDAndFollowingID(ctx context.Context, f entity.Follow) error {
	err := pfr.queries.RemoveByFollowerIDAndFollowingID(ctx, queries.RemoveByFollowerIDAndFollowingIDParams{
		FollowerID:  f.FollowerID(),
		FollowingID: f.FollowingID(),
	})
	if err != nil {
		return fmt.Errorf("PostgresFollowRepository.RemoveByFollowingID: %w", err)
	}
	return nil
}
