package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/model"
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

func (pur *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	query := `SELECT id, username, email, phone_number, first_name, last_name, status FROM identity.users WHERE email = $1;`

	user := model.User{}
	if err := pur.conn.QueryRow(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.PhoneNumber, &user.FirstName, &user.LastName, &user.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, errs.NewError(errs.NotFound, err, errs.UserNotFound)
		}
		return model.User{}, fmt.Errorf("PostgresUserRepository[GetByEmail]: %w", err)
	}

	return user, nil
}

func (pur *PostgresUserRepository) Create(ctx context.Context, user model.User) error {
	query := `INSERT INTO identity.users(username, email, phone_number, first_name, last_name) VALUES ($1, $2, $3, $4, $5);`

	cmdTag, err := pur.conn.Exec(ctx, query, user.Username, user.Email, user.PhoneNumber, user.FirstName, user.LastName)
	if err != nil {
		e := errs.NewError(errs.Conflict, err)
		if pgerr, ok := errors.AsType[*pgconn.PgError](err); ok && pgerr.Code == "23505" {
			switch pgerr.ConstraintName {
			case "users_email_key":
				e.AddCode(errs.EmailAlreadyUsed)
			case "users_phone_number_key":
				e.AddCode(errs.PhoneNumberAlreadyUsed)
			case "users_username_key":
				e.AddCode(errs.UsernameAlreadyUsed)
			}
			return e
		}
		return fmt.Errorf("PostgresUserRepository[GetByEmail]: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresUserRepository[GetByEmail]: Can not insert user")
	}
	return nil
}
