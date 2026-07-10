package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresUserRepository(conn *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		conn: conn,
	}
}

func (pur *PostgresUserRepository) UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) (domain.UserProfile, error) {
	queryBuilder := squirrel.Update("profile.users").
		SetMap(fieldUpdates).
		Where("id = ?", userID).
		Suffix("RETURNING id, username, name, avatar_url, bio").
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return domain.UserProfile{}, err
	}

	var up domain.UserProfile
	if err := pur.conn.QueryRow(ctx, query, args...).Scan(&up.ID, &up.Username, &up.Name, &up.AvatarURL, &up.Bio); err != nil {
		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgerr.Code == "23505" && pgerr.ConstraintName == "users_username_key" {
				return domain.UserProfile{}, errs.NewError(errs.Conflict, pgerr, errs.UsernameAlreadyUsed)
			}
		}
		return domain.UserProfile{}, fmt.Errorf("PostgresUserRepository[UpdateUsername]: %w", err)
	}
	return up, nil
}

func (pur *PostgresUserRepository) GetProfileByUsername(ctx context.Context, username string) (domain.UserProfile, error) {
	query := `SELECT id, username, name, avatar_url, bio, followers_count, following_count, posts_count
				FROM profile.users
				WHERE username = $1;`

	var up domain.UserProfile
	if err := pur.conn.QueryRow(ctx, query, username).
		Scan(&up.ID, &up.Username, &up.Name, &up.AvatarURL, &up.Bio, &up.FollowersCount, &up.FollowingCount, &up.PostsCount); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserProfile{}, errs.NewError(errs.NotFound, err, errs.UserNotFound)
		}
		return domain.UserProfile{}, fmt.Errorf("PostgresUserRepository[FindByUsername]: %w", err)
	}

	return up, nil
}
