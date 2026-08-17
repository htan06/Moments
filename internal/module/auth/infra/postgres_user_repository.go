package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/auth/domain"
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

func (pur *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT id, email, phone_number, password_hash, status 
				FROM identity.users
				WHERE email = $1;`

	var user domain.User
	if err := pur.conn.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PhoneNumber, &user.PasswordHash, &user.Status); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errs.NewError(errs.NotFound, err, domain.UserNotFound)
		}
		return domain.User{}, fmt.Errorf("PostgresUserRepository.GetByEmail.: %w", err)
	}

	return user, nil
}

func (pur *PostgresUserRepository) GetByID(ctx context.Context, id int64) (domain.User, error) {
	query := `SELECT id, email, phone_number, password_hash, status 
				FROM identity.users
				WHERE id = $1;`

	var user domain.User
	if err := pur.conn.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PhoneNumber, &user.PasswordHash, &user.Status); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errs.NewError(errs.NotFound, err, domain.UserNotFound)
		}
		return domain.User{}, fmt.Errorf("PostgresUserRepository.GetByID: %w", err)
	}

	return user, nil
}

func (pur *PostgresUserRepository) Create(ctx context.Context, user domain.User) error {
	insertQry := `INSERT INTO identity.users (email, password_hash, status) VALUES ($1, $2, $3);`
	cmd, err := pur.conn.Exec(ctx, insertQry, user.Email, user.PasswordHash, user.Status)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.Create: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("PostgresUserRepository.Create: insert user_identity failure")
	}
	return nil
}

func (pur *PostgresUserRepository) UpdateLastLogin(ctx context.Context, user domain.User) error {
	query := `UPDATE identity.users SET last_login_at = $1 WHERE id = $2;`

	cmd, err := pur.conn.Exec(ctx, query, user.LastLoginAt, user.ID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateLastLogin: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository.UpdateLastLogin: update user_identity failure")
	}
	return nil
}

func (pur *PostgresUserRepository) UpdatePassword(ctx context.Context, user domain.User) error {
	query := `UPDATE identity.users SET password_hash = $1 WHERE id = $2;`

	cmd, err := pur.conn.Exec(ctx, query, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdatePassword: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository.UpdatePassword: update iedntity.user failure")
	}
	return nil
}

func (pur *PostgresUserRepository) UpdateStatusActiveIfExistsProfile(ctx context.Context, userID int64) error {
	query := `UPDATE identity.users 
				SET status = $1 
				WHERE 
					id = $2
					AND EXISTS (
						SELECT 1 
						FROM profile.users 
						WHERE 
							user_id = $2
					);`

	cmd, err := pur.conn.Exec(ctx, query, domain.UserStatusActive, userID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdateStatusActiveIfExistsProfile: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository.UpdateStatusActiveIfExistsProfile: update iedntity.user.status failure")
	}
	return nil
}

func (pur *PostgresUserRepository) GetUserDetailByID(ctx context.Context, userID int64) (domain.UserDetail, error) {
	query := `SELECT pu.username, iu.email, iu.phone_number, iu.last_login_at
				FROM identity.users iu
				JOIN profile.users pu
					ON iu.id = pu.user_id
				WHERE id = $1;`

	var ud domain.UserDetail
	if err := pur.conn.QueryRow(ctx, query, userID).Scan(&ud.Username, &ud.Email, &ud.PhoneNumber, &ud.LastLoginAt); err != nil {
		return domain.UserDetail{}, fmt.Errorf("PostgresUserRepository.GetUserDetailByID: %w", err)
	}
	return ud, nil
}
