package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/friend/model"
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

func (pur *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (model.UserProfile, error) {
	query := `SELECT username, first_name, last_name, avatar_url, cover_photo_url 
				FROM identity.users
				WHERE username = $1;`

	var up model.UserProfile
	if err := pur.conn.QueryRow(ctx, query, username).
		Scan(&up.Username, &up.FirstName, &up.LastName, &up.AvatarURL, &up.CoverPhotoURL); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserProfile{}, errs.NewError(errs.NotFound, err, errs.UserNotFound)
		}
		return model.UserProfile{}, fmt.Errorf("PostgresUserRepository[FindByUsername]: %w", err)
	}

	return up, nil
}

func (pur *PostgresUserRepository) CreateFriendRequest(ctx context.Context, friendRequqest model.FriendRequest) error {
	query := `INSERT INTO identity.friend_requests (sender_id, receiver_id, status) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;`
	_, err := pur.conn.Exec(ctx, query, friendRequqest.SenderID, friendRequqest.ReceiverID, friendRequqest.Status)
	if err != nil {
		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok && pgerr.ConstraintName == "friend_requests_receiver_id_fkey" {
			return errs.NewError(errs.NotFound, pgerr, errs.ReceiverNotFound)
		}
		return fmt.Errorf("PostgresUserRepository.CreateFriendRequest: %w", err)
	}
	return nil
}
