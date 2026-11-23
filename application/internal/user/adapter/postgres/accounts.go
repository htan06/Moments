// Package postgres stores the user module's data in the `users` schema.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/htan06/Moments/application/internal/platform/pg"
	"github.com/htan06/Moments/application/internal/user/adapter/postgres/sqlc"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// Unique index names from migrations/00001_init.sql.
const (
	emailUniqueIndex    = "accounts_email_uq"
	usernameUniqueIndex = "accounts_username_uq"
)

type Accounts struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewAccounts(db *pgxpool.Pool) *Accounts {
	return &Accounts{db: db, q: sqlc.New(db)}
}

func (r *Accounts) EmailExists(ctx context.Context, email string) (bool, error) {
	ok, err := r.q.EmailExists(ctx, email)
	if err != nil {
		return false, fmt.Errorf("postgres: email exists: %w", err)
	}
	return ok, nil
}

// Create inserts Account and Profile in one transaction (USER-R7).
func (r *Accounts) Create(ctx context.Context, a domain.Account, p domain.Profile) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("postgres: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)

	row, err := q.CreateAccount(ctx, sqlc.CreateAccountParams{
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

	if err := q.CreateProfile(ctx, sqlc.CreateProfileParams{UserID: row.ID, Name: p.Name}); err != nil {
		return 0, fmt.Errorf("postgres: create profile: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("postgres: commit: %w", err)
	}
	return row.ID, nil
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

func toAccount(row sqlc.UsersAccount, err error) (domain.Account, error) {
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
