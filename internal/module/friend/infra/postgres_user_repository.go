package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/friend/model"
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

func (pur *PostgresUserRepository) CreateFriendRequest(ctx context.Context, friendRequqest model.FriendRequest) (model.FriendRequest, error) {
	query := `INSERT INTO identity.friend_requests (sender_id, receiver_id) 
				SELECT $1, $2 WHERE NOT EXISTS (
					SELECT 1 FROM identity.friend_requests WHERE sender_id = $2 AND receiver_id = $1) 
			  ON CONFLICT DO NOTHING RETURNING id, sender_id, receiver_id, created_at, updated_at;`

	fr := model.FriendRequest{}

	if err := pur.conn.QueryRow(ctx, query, friendRequqest.SenderID, friendRequqest.ReceiverID).
		Scan(&fr.ID, &fr.SenderID, &fr.ReceiverID, &fr.CreatedAt, &fr.UpdatedAt); err != nil {

		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok && pgerr.ConstraintName == "friend_requests_receiver_id_fkey" {
			return model.FriendRequest{}, errs.NewError(errs.NotFound, pgerr, errs.ReceiverNotFound)
		}
		return model.FriendRequest{}, fmt.Errorf("PostgresUserRepository.CreateFriendRequest: %w", err)
	}
	return fr, nil
}

func (pur *PostgresUserRepository) GetSentFriendRequests(ctx context.Context, userID int64) ([]model.FriendRequestProfile, error) {
	query := `SELECT fr.id as friend_request_id, u.id as user_id, u.username, u.first_name, u.last_name, u.avatar_url, u.cover_photo_url
				FROM identity.friend_requests fr 
				JOIN identity.users u ON u.id = fr.receiver_id
				WHERE fr.sender_id = $1;`

	rows, err := pur.conn.Query(ctx, query, userID)
	if err != nil {
		return []model.FriendRequestProfile{}, fmt.Errorf("PostgresUserRepository.GetSentFriendRequests: %w", err)
	}

	friendRequests := make([]model.FriendRequestProfile, 0)
	for rows.Next() {
		fr := model.FriendRequestProfile{}
		rows.Scan(&fr.ID, &fr.UserID, &fr.Username, &fr.FirstName, &fr.LastName, &fr.AvatarURL, &fr.CoverPhotoURL)
		friendRequests = append(friendRequests, fr)
	}
	rows.Close()
	return friendRequests, nil
}

func (pur *PostgresUserRepository) GetReceivedFriendRequests(ctx context.Context, userID int64) ([]model.FriendRequestProfile, error) {
	query := `SELECT fr.id as friend_request_id, u.id as user_id, u.username, u.first_name, u.last_name, u.avatar_url, u.cover_photo_url
				FROM identity.friend_requests fr 
				JOIN identity.users u ON u.id = fr.sender_id
				WHERE fr.receiver_id = $1;`

	rows, err := pur.conn.Query(ctx, query, userID)
	if err != nil {
		return []model.FriendRequestProfile{}, fmt.Errorf("PostgresUserRepository.GetReceivedFriendRequests: %w", err)
	}

	friendRequests := make([]model.FriendRequestProfile, 0)
	for rows.Next() {
		fr := model.FriendRequestProfile{}
		rows.Scan(&fr.ID, &fr.UserID, &fr.Username, &fr.FirstName, &fr.LastName, &fr.AvatarURL, &fr.CoverPhotoURL)
		friendRequests = append(friendRequests, fr)
	}
	rows.Close()
	return friendRequests, nil
}

func (pur *PostgresUserRepository) CreateFriend(ctx context.Context, lowUserID int64, highUserID int64) error {
	query := `INSERT INTO identity.friends (low_user_id, high_user_id) VALUES ($1, $2);`

	if _, err := pur.conn.Exec(ctx, query, lowUserID, highUserID); err != nil {
		return fmt.Errorf("PostgresUserRepository.CreateFriend: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) GetListFriends(ctx context.Context, userID int64) ([]model.UserProfile, error) {
	query := `SELECT u.id, u.username, u.first_name, u.last_name, u.avatar_url, u.cover_photo_url
				FROM identity.friends f 
				JOIN identity.users u 
					ON u.id = CASE WHEN f.low_user_id = $1 THEN f.high_user_id ELSE f.low_user_id END
				WHERE $1 IN (f.low_user_id, f.high_user_id)`

	rows, err := pur.conn.Query(ctx, query, userID)
	if err != nil {
		return []model.UserProfile{}, fmt.Errorf("PostgresUserRepository.GetListFriends: %w", err)
	}

	friendRequests := make([]model.UserProfile, 0)
	for rows.Next() {
		fr := model.UserProfile{}
		rows.Scan(&fr.UserID, &fr.Username, &fr.FirstName, &fr.LastName, &fr.AvatarURL, &fr.CoverPhotoURL)
		friendRequests = append(friendRequests, fr)
	}
	rows.Close()
	return friendRequests, nil
}

func (pur *PostgresUserRepository) AcceptFriendRequest(ctx context.Context, frID int64, userID int64) error {
	queryFind := `SELECT sender_id, receiver_id FROM identity.friend_requests WHERE id = $1;`
	queryInsert := `INSERT INTO identity.friends(low_user_id, high_user_id) VALUES($1, $2);`
	queryDelete := `DELETE FROM identity.friend_requests WHERE id = $1;`
	var senderID, receiverID int64

	tx, err := pur.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, queryFind, frID).Scan(&senderID, &receiverID); err != nil {
		return err
	}

	if receiverID != userID {
		return errs.NewError(errs.NotFound, nil, errs.FriendRequestNotFound)
	}

	lowID, highID := func(id1 int64, id2 int64) (int64, int64) {
		if id1 < id2 {
			return id1, id2
		}
		return id2, id1
	}(senderID, receiverID)

	if _, err := tx.Exec(ctx, queryInsert, lowID, highID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, queryDelete, frID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (pur *PostgresUserRepository) RejectFriendRequest(ctx context.Context, frID int64, userID int64) error {
	query := `DELETE FROM identity.friend_requests WHERE id = $1 AND receiver_id = $2;`
	if _, err := pur.conn.Exec(ctx, query, frID, userID); err != nil {
		return fmt.Errorf("PostgresUserRepository.RejectFriendRequest: %w", err)
	}
	return nil
}

func (pur *PostgresUserRepository) CancelFriendRequest(ctx context.Context, frID int64, userID int64) error {
	query := `DELETE FROM identity.friend_requests WHERE id = $1 AND sender_id = $2;`
	if _, err := pur.conn.Exec(ctx, query, frID, userID); err != nil {
		return fmt.Errorf("PostgresUserRepository.CancelFriendRequest: %w", err)
	}
	return nil
}
