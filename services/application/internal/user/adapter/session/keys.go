// Package session stores the user module's short-lived state in Redis:
// OTP verification sessions and failed-login counters.
package session

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// Every Redis key of the user module is built here and nowhere else.
const prefix = "user:"

// verificationKey holds one VerificationSession (JSON), expiring with it.
func verificationKey(sessionID string) string {
	return prefix + "verify:" + sessionID
}

// loginFailAccountKey counts failed logins of an account (email and username
// share it).
func loginFailAccountKey(accountID int64) string {
	return prefix + "login_fail:acct:" + strconv.FormatInt(accountID, 10)
}

// loginFailIdentifierKey counts failed logins of an identifier matching no
// account. The identifier is hashed so emails never appear in key names.
func loginFailIdentifierKey(identifier string) string {
	h := sha256.Sum256([]byte(identifier))
	return prefix + "login_fail:ident:" + hex.EncodeToString(h[:])
}
