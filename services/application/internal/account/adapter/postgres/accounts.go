// Package postgres stores the account module's data in the `account` schema.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/htan06/Moments/application/internal/account/adapter/postgres/sqlc"
	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/platform/pg"
)

// Unique index names from migrations/20261125000000_init.sql.
const (
	emailUniqueIndex    = "accounts_email_uq"
	usernameUniqueIndex = "accounts_username_uq"
)

type Accounts struct {
	q *sqlc.Queries
}

func NewAccounts(db *pgxpool.Pool) *Accounts {
	return &Accounts{q: sqlc.New(db)}
}

func (r *Accounts) EmailExists(ctx context.Context, email string) (bool, error) {
	ok, err := r.q.EmailExists(ctx, email)
	if err != nil {
		return false, fmt.Errorf("postgres: email exists: %w", err)
	}
	return ok, nil
}

func (r *Accounts) UsernameExists(ctx context.Context, username string) (bool, error) {
	ok, err := r.q.UsernameExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("postgres: username exists: %w", err)
	}
	return ok, nil
}

// Create inserts the account and returns its id. The profile is created by
// the user module right after (ACCOUNT-R7).
func (r *Accounts) Create(ctx context.Context, a domain.Account) (int64, error) {
	row, err := r.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		Email:        a.Email,
		Username:     a.Username,
		PasswordHash: a.PasswordHash,
		Status:       a.Status,
	})
	switch {
	case pg.IsUniqueViolation(err, emailUniqueIndex):
		return 0, domain.ErrEmailTaken
	case pg.IsUniqueViolation(err, usernameUniqueIndex):
		return 0, domain.ErrUsernameTaken
	case err != nil:
		return 0, fmt.Errorf("postgres: create account: %w", err)
	}
	return row.ID, nil
}

// Delete removes an account; it undoes Create when the profile could not be
// created (ACCOUNT-R7).
func (r *Accounts) Delete(ctx context.Context, id int64) error {
	if err := r.q.DeleteAccount(ctx, id); err != nil {
		return fmt.Errorf("postgres: delete account: %w", err)
	}
	return nil
}

func (r *Accounts) ByID(ctx context.Context, id int64) (domain.Account, error) {
	return toAccount(r.q.GetAccountByID(ctx, id))
}

func (r *Accounts) ByEmail(ctx context.Context, email string) (domain.Account, error) {
	return toAccount(r.q.GetAccountByEmail(ctx, email))
}

func (r *Accounts) ByUsername(ctx context.Context, username string) (domain.Account, error) {
	return toAccount(r.q.GetAccountByUsername(ctx, username))
}

func toAccount(row sqlc.AccountAccount, err error) (domain.Account, error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Account{}, domain.ErrAccountNotFound
	case err != nil:
		return domain.Account{}, fmt.Errorf("postgres: get account: %w", err)
	}
	return domain.Account{
		ID:           row.ID,
		Email:        row.Email,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Status:       row.Status,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}
