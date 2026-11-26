// Package domain holds the user module's entities (Profile, Follow, Block),
// invariants and domain errors. It imports only the standard library.
package domain

import "errors"

var (
	// ErrInvalidName: profile name is empty or longer than allowed.
	ErrInvalidName = errors.New("user: invalid name")
	// ErrInvalidUsername: the username copy is empty. Its format is checked
	// by the account module, which owns usernames (ACCOUNT-R5).
	ErrInvalidUsername = errors.New("user: invalid username")
	// ErrUsernameTaken: another profile already holds this username copy.
	ErrUsernameTaken = errors.New("user: username taken")
	// ErrProfileNotFound: no profile for the user.
	ErrProfileNotFound = errors.New("user: profile not found")
)
