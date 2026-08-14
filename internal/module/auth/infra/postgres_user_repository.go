package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/auth/domain"
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

func (pur *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT u.id, u.name, u.username, ui.email, ui.phone_number, ui.password_hash, ui.status 
				FROM identity.user_identity ui
				JOIN profile.users u ON ui.user_id = u.id
				WHERE email = $1;`

	var user domain.User
	if err := pur.conn.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.PhoneNumber, &user.PasswordHash, &user.Status); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errs.NewError(errs.NotFound, err, domain.UserNotFound)
		}
		return domain.User{}, fmt.Errorf("PostgresUserRepository.GetByEmail.: %w", err)
	}

	return user, nil
}

func (pur *PostgresUserRepository) GetByID(ctx context.Context, id int64) (domain.User, error) {
	query := `SELECT u.id, u.name, u.username, ui.email, ui.phone_number, ui.password_hash, ui.status 
				FROM identity.user_identity ui
				JOIN profile.users u ON ui.user_id = u.id
				WHERE u.id = $1;`

	var user domain.User
	if err := pur.conn.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.PhoneNumber, &user.PasswordHash, &user.Status); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errs.NewError(errs.NotFound, err, domain.UserNotFound)
		}
		return domain.User{}, fmt.Errorf("PostgresUserRepository.GetByID: %w", err)
	}

	return user, nil
}

func (pur *PostgresUserRepository) Create(ctx context.Context, user domain.User) error {

	tx, err := pur.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.Create: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	insertProfile := `INSERT INTO profile.users (username, name) VALUES ($1, $2) RETURNING id;`

	if err := tx.QueryRow(ctx, insertProfile, user.Username, user.Name).Scan(&id); err != nil {
		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok && pgerr.Code == "23505" {
			switch pgerr.ConstraintName {
			case "users_username_key":
				return errs.NewError(errs.Conflict, err, domain.UsernameAlreadyUsed)
			}
		}
		return fmt.Errorf("PostgresUserRepository.Create: %w", err)
	}

	insertIdentity := `INSERT INTO identity.user_identity (user_id, email, password_hash) VALUES ($1, $2, $3);`
	cmd, err := tx.Exec(ctx, insertIdentity, id, user.Email, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.Create: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("PostgresUserRepository.Create: insert user_identity failure")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("PostgresUserRepository.Create: commit failure")
	}
	return nil
}

func (pur *PostgresUserRepository) UpdateLastLogin(ctx context.Context, user domain.User) error {
	query := `UPDATE identity.user_identity SET last_login_at = $1 WHERE user_id = $2;`

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
	query := `UPDATE identity.user_identity SET password_hash = $1 WHERE user_id = $2;`

	cmd, err := pur.conn.Exec(ctx, query, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("PostgresUserRepository.UpdatePassword: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository.UpdatePassword: update user_identity failure")
	}
	return nil
}
