package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
	"github.com/jackc/pgx/v5"
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

func (pur *PostgresUserRepository) GetAvatarIDByUserID(ctx context.Context, userID int64) (string, error) {
	query := `SELECT avatar_id FROM profile.users WHERE id = $1;`

	var id string
	if err := pur.conn.QueryRow(ctx, query, userID).Scan(&id); err != nil {
		return "", fmt.Errorf("PostgresUserRepository.GetAvatarIDByUserID: %w", err)
	}
	return id, nil
}

func (pur *PostgresUserRepository) UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error {
	queryBuilder := squirrel.Update("profile.users").
		SetMap(fieldUpdates).
		Where("id = ?", userID).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateProfile: %w", err)
	}

	if _, err := pur.conn.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateProfile: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) GetProfileByUsername(ctx context.Context, username string) (domain.ProfileQry, error) {
	query := `SELECT id, username, name, avatar_id, bio, followers_count, following_count, posts_count
				FROM profile.users
				WHERE username = $1;`

	row, err := pur.conn.Query(ctx, query, username)
	if err != nil {
		return domain.ProfileQry{}, fmt.Errorf("PostgresUserRepository.GetProfileByUsername: %w", err)
	}

	profile, err := pgx.CollectOneRow[domain.ProfileQry](row, pgx.RowToStructByName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProfileQry{}, errs.NewError(errs.NotFound, err, errs.UserNotFound)
		}
		return domain.ProfileQry{}, fmt.Errorf("PostgresUserRepository[FindByUsername]: %w", err)
	}

	return profile, nil
}

func (pur *PostgresUserRepository) UpdateAvatarID(ctx context.Context, userID int64, avatarID string) error {
	query := `UPDATE profile.users SET avatar_id = $1 WHERE id = $2;`

	if _, err := pur.conn.Exec(ctx, query, avatarID, userID); err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateAvatarID: %w", err)
	}
	return nil
}
