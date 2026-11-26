// Package query holds the account module's read-side use cases: one file and one
// handler per use case, each with its own result model.
package query

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// UsernameAvailability asks whether a username can still be registered. The
// client calls it while the user types; the answer is a hint only, because the
// username is not reserved until the account is created.
type UsernameAvailability struct {
	Username string
}

// UsernameChecker reports whether an account already uses a username.
type UsernameChecker interface {
	UsernameExists(ctx context.Context, username string) (bool, error)
}

type UsernameAvailabilityHandler struct {
	users     UsernameChecker
	usernames domain.UsernameRules
}

func NewUsernameAvailabilityHandler(users UsernameChecker, usernames domain.UsernameRules) *UsernameAvailabilityHandler {
	return &UsernameAvailabilityHandler{users: users, usernames: usernames}
}

// Handle returns domain.ErrInvalidUsername for a malformed or reserved name
// (ACCOUNT-R5), otherwise whether no account uses it.
func (h *UsernameAvailabilityHandler) Handle(ctx context.Context, q UsernameAvailability) (bool, error) {
	username, err := h.usernames.Parse(q.Username)
	if err != nil {
		return false, err
	}
	exists, err := h.users.UsernameExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("username availability: %w", err)
	}
	return !exists, nil
}
