package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/model"
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

func (pur *PostgresUserRepository) GetInfo(ctx context.Context, id int64) (model.User, error) {
	query := `SELECT id, username, first_name, last_name, avatar_url, cover_photo_url FROM identity.users WHERE id = $1`

	var user model.User
	if err := pur.conn.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.AvatarURL, &user.CoverPhotoURL); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, errs.NewError(errs.NotFound, err, errs.UserNotFound)
		}

		return model.User{}, fmt.Errorf("PostgresUserRepository[GetInfo]: %w", err)
	}
	return user, nil
}

func (pur *PostgresUserRepository) UpdateInfo(ctx context.Context, user model.User) error {
	query := `UPDATE identity.users SET first_name = $1, last_name = $2, avatar_url = $3, cover_photo_url = $4 WHERE id = $5;`

	_, err := pur.conn.Exec(ctx, query, user.FirstName, user.LastName, user.AvatarURL, user.CoverPhotoURL, user.ID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository[UpdateInfo]: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) ChangeReadStatus(ctx context.Context, user model.User) error {
	query := `UPDATE identity.users SET read_status = $1 WHERE id = $2;`

	_, err := pur.conn.Exec(ctx, query, user.ReadStatus, user.ID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository[ChangeReadStatus]: %w", err)
	}
	return nil
}
func (pur *PostgresUserRepository) UpdateUsername(ctx context.Context, user model.User) error {
	query := `UPDATE identity.users SET username = $1 WHERE id = $2;`

	_, err := pur.conn.Exec(ctx, query, user.Username, user.ID)
	if err != nil {
		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgerr.Code == "23505" && pgerr.ConstraintName == "users_username_key" {
				return errs.NewError(errs.Conflict, pgerr, errs.UsernameAlreadyUsed)
			}
		}
		return fmt.Errorf("PostgresUserRepository[UpdateUsername]: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (model.UserProfile, error) {
	query := `SELECT id, username, first_name, last_name, avatar_url, cover_photo_url 
				FROM identity.users
				WHERE username = $1;`

	var up model.UserProfile
	if err := pur.conn.QueryRow(ctx, query, username).
		Scan(&up.ID, &up.Username, &up.FirstName, &up.LastName, &up.AvatarURL, &up.CoverPhotoURL); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserProfile{}, errs.NewError(errs.NotFound, err, errs.UserNotFound)
		}
		return model.UserProfile{}, fmt.Errorf("PostgresUserRepository[FindByUsername]: %w", err)
	}

	return up, nil
}