package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// AccountStatus asks what other modules need to know about an account:
// whether it exists, whether it is ACTIVE, and its username.
type AccountStatus struct {
	UserID int64
}

// AccountStatusResult is the read model of AccountStatus. A missing account
// has Exists = false and the other fields empty.
type AccountStatusResult struct {
	Exists   bool
	Active   bool
	Username string
}

// AccountStatusRow is what the store returns for one account.
type AccountStatusRow struct {
	Username string
	Status   string
}

// AccountStatusReader loads the status row. It returns
// domain.ErrAccountNotFound when absent.
type AccountStatusReader interface {
	Status(ctx context.Context, id int64) (AccountStatusRow, error)
}

type AccountStatusHandler struct {
	accounts AccountStatusReader
}

func NewAccountStatusHandler(accounts AccountStatusReader) *AccountStatusHandler {
	return &AccountStatusHandler{accounts: accounts}
}

// Handle reports a missing account as Exists = false, not as an error.
func (h *AccountStatusHandler) Handle(ctx context.Context, q AccountStatus) (AccountStatusResult, error) {
	row, err := h.accounts.Status(ctx, q.UserID)
	if errors.Is(err, domain.ErrAccountNotFound) {
		return AccountStatusResult{}, nil
	}
	if err != nil {
		return AccountStatusResult{}, fmt.Errorf("account status: %w", err)
	}
	return AccountStatusResult{
		Exists:   true,
		Active:   row.Status == domain.StatusActive,
		Username: row.Username,
	}, nil
}
