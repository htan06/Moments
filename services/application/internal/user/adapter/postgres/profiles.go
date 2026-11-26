// Package postgres stores the user module's data in the `users` schema.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/htan06/Moments/application/internal/platform/pg"
	"github.com/htan06/Moments/application/internal/user/adapter/postgres/sqlc"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// Unique index name from migrations/20261125000000_init.sql.
const usernameUniqueIndex = "profiles_username_uq"

type Profiles struct {
	q *sqlc.Queries
}

func NewProfiles(db *pgxpool.Pool) *Profiles {
	return &Profiles{q: sqlc.New(db)}
}

// Create inserts the profile; an existing profile of the same user is left
// unchanged (ON CONFLICT DO NOTHING), so retries are safe. A username copy
// held by another profile returns domain.ErrUsernameTaken.
func (r *Profiles) Create(ctx context.Context, p domain.Profile) error {
	err := r.q.CreateProfile(ctx, sqlc.CreateProfileParams{UserID: p.UserID, Username: p.Username, Name: p.Name})
	switch {
	case pg.IsUniqueViolation(err, usernameUniqueIndex):
		return domain.ErrUsernameTaken
	case err != nil:
		return fmt.Errorf("postgres: create profile: %w", err)
	}
	return nil
}

// UpdateUsername replaces the username copy. It returns
// domain.ErrProfileNotFound when the user has no profile and
// domain.ErrUsernameTaken when another profile holds the username.
func (r *Profiles) UpdateUsername(ctx context.Context, userID int64, username string) error {
	n, err := r.q.UpdateProfileUsername(ctx, sqlc.UpdateProfileUsernameParams{UserID: userID, Username: username})
	switch {
	case pg.IsUniqueViolation(err, usernameUniqueIndex):
		return domain.ErrUsernameTaken
	case err != nil:
		return fmt.Errorf("postgres: update profile username: %w", err)
	case n == 0:
		return domain.ErrProfileNotFound
	}
	return nil
}
