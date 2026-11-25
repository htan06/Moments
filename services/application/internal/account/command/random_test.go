package command

import (
	"regexp"
	"testing"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestCryptoRandom_SessionIDIsUUIDv4(t *testing.T) {
	seen := map[string]bool{}
	for range 100 {
		id, err := CryptoRandom{}.SessionID()
		if err != nil {
			t.Fatal(err)
		}
		if !uuidV4.MatchString(id) {
			t.Fatalf("SessionID() = %q, not a UUID v4", id)
		}
		if seen[id] {
			t.Fatalf("SessionID() repeated %q", id)
		}
		seen[id] = true
	}
}
