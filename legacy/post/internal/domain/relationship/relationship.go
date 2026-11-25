// Package relationship mirrors the Relationship enum owned by the User
// context. It is read left to right: Following between A and B means
// A follows B.
package relationship

import "github.com/htan06/Moments/post/internal/domain"

type Relationship string

const (
	Following Relationship = "FOLLOWING"
	Follower  Relationship = "FOLLOWER"
	None      Relationship = "NONE"
	Blocking  Relationship = "BLOCKING"
	BlockedBy Relationship = "BLOCKED_BY"
)

func Parse(s string) (Relationship, error) {
	r := Relationship(s)
	if !r.Valid() {
		return "", domain.ErrRelationshipInvalid
	}
	return r, nil
}

func (r Relationship) Valid() bool {
	switch r {
	case Following, Follower, None, Blocking, BlockedBy:
		return true
	}
	return false
}

// IsBlocked reports whether either side has blocked the other.
func (r Relationship) IsBlocked() bool { return r == Blocking || r == BlockedBy }

func (r Relationship) String() string { return string(r) }
