// Package session stores the account module's short-lived state in Redis:
// OTP verification sessions and failed-login counters.
package session

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/htan06/Moments/application/internal/account/command"
)

// Every Redis key of the account module is built here and nowhere else.
const prefix = "account:"

// verificationKey holds one VerificationSession (JSON), expiring with it.
func verificationKey(sessionID string) string {
	return prefix + "verify:" + sessionID
}

// loginFailKey holds the failed-login counter of a subject (ACCOUNT-R3):
//   - user:login_fail:acct:<account id> when the identifier matches an
//     account, so its email and username share one counter;
//   - user:login_fail:ident:<sha256 hex> otherwise. The identifier is hashed
//     so emails never appear in key names.
func loginFailKey(s command.LoginSubject) string {
	if s.AccountID > 0 {
		return prefix + "login_fail:acct:" + strconv.FormatInt(s.AccountID, 10)
	}
	h := sha256.Sum256([]byte(s.Identifier))
	return prefix + "login_fail:ident:" + hex.EncodeToString(h[:])
}
