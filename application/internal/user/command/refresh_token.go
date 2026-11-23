package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// RefreshToken exchanges a refresh token for a new token pair.
type RefreshToken struct {
	RefreshToken string
}

type RefreshTokenHandler struct {
	accounts Accounts
	tokens   Tokens
}

func NewRefreshTokenHandler(accounts Accounts, tokens Tokens) *RefreshTokenHandler {
	return &RefreshTokenHandler{accounts: accounts, tokens: tokens}
}

// Handle reloads username and status instead of copying them from the old
// token (USER-R6).
func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshToken) (token.Pair, error) {
	claims, err := h.tokens.ParseRefresh(cmd.RefreshToken)
	if err != nil {
		return token.Pair{}, domain.ErrInvalidToken
	}
	account, err := h.accounts.ByID(ctx, claims.UserID)
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return token.Pair{}, domain.ErrInvalidToken
	case err != nil:
		return token.Pair{}, fmt.Errorf("refresh token: %w", err)
	}
	if !account.Active() {
		return token.Pair{}, domain.ErrAccountInactive
	}
	pair, err := h.tokens.Issue(claimsOf(account))
	if err != nil {
		return token.Pair{}, fmt.Errorf("refresh token: %w", err)
	}
	return pair, nil
}
