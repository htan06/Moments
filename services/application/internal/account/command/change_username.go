package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/platform/token"
)

// ChangeUsername renames the signed-in user. UserID comes from the access
// token, never from the request body.
type ChangeUsername struct {
	UserID   int64
	Username string
}

type ChangeUsernameHandler struct {
	accounts  AccountReadUsernameChanger
	profiles  ProfileUsernameUpdater
	tokens    TokenIssuer
	usernames domain.UsernameRules
}

func NewChangeUsernameHandler(accounts AccountReadUsernameChanger, profiles ProfileUsernameUpdater, tokens TokenIssuer,
	usernames domain.UsernameRules) *ChangeUsernameHandler {
	return &ChangeUsernameHandler{accounts: accounts, profiles: profiles, tokens: tokens, usernames: usernames}
}

// Handle renames the account and its profile copy together (ACCOUNT-R5):
//  1. the account row is updated inside a transaction;
//  2. still inside it, the user module updates the profile copy;
//  3. the transaction commits only if step 2 succeeded, otherwise it rolls
//     back and the account keeps the old username.
//
// The old username is free as soon as the commit lands. A new token pair is
// returned, carrying the new username (ACCOUNT-R6). Renaming to the current
// username changes nothing and still returns a pair.
//
// Errors: domain.ErrInvalidUsername, domain.ErrUsernameTaken (held by another
// account or another profile copy), domain.ErrAccountNotFound,
// domain.ErrAccountInactive.
func (h *ChangeUsernameHandler) Handle(ctx context.Context, cmd ChangeUsername) (token.Pair, error) {
	username, err := h.usernames.Parse(cmd.Username)
	if err != nil {
		return token.Pair{}, err
	}
	account, err := h.accounts.ByID(ctx, cmd.UserID)
	if err != nil {
		return token.Pair{}, err
	}
	if !account.Active() {
		return token.Pair{}, domain.ErrAccountInactive
	}
	old := account.Username
	changed, err := account.Rename(username)
	if err != nil {
		return token.Pair{}, err
	}

	if changed {
		// The profile update commits in the user module before the account
		// transaction commits. synced records that it did, so a failed
		// account commit can be undone on the profile side.
		synced := false
		err = h.accounts.ChangeUsername(ctx, account.ID, username, func(ctx context.Context) error {
			if err := h.profiles.UpdateProfileUsername(ctx, account.ID, username); err != nil {
				return err
			}
			synced = true
			return nil
		})
		if err != nil {
			if synced {
				h.undoProfile(ctx, account.ID, old, &err)
			}
			if errors.Is(err, domain.ErrUsernameTaken) || errors.Is(err, domain.ErrAccountNotFound) {
				return token.Pair{}, err
			}
			return token.Pair{}, fmt.Errorf("change username: %w", err)
		}
	}

	pair, err := h.tokens.Issue(claimsOf(account))
	if err != nil {
		// The rename is done; the client keeps its old tokens until it signs
		// in or refreshes, which reads the current username.
		return token.Pair{}, fmt.Errorf("change username: %w", err)
	}
	return pair, nil
}

// undoProfile puts the old username back in the profile copy after the
// account transaction failed to commit, so both modules agree again. It runs
// even if the request was cancelled. If it fails too, the two copies differ
// and err says so for the logs.
func (h *ChangeUsernameHandler) undoProfile(ctx context.Context, id int64, old string, err *error) {
	if uerr := h.profiles.UpdateProfileUsername(context.WithoutCancel(ctx), id, old); uerr != nil {
		*err = fmt.Errorf("%w; undo profile username of account %d: %v", *err, id, uerr)
	}
}
