package password

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcrypt_HashMatches(t *testing.T) {
	b, err := NewBcrypt(bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	h, err := b.Hash("mậtkhẩu1")
	if err != nil {
		t.Fatal(err)
	}
	if !b.Matches(h, "mậtkhẩu1") {
		t.Fatal("correct password does not match")
	}
	if b.Matches(h, "mậtkhẩu2") {
		t.Fatal("wrong password matches")
	}
}

func TestBcrypt_DummyHashUsesConfiguredCost(t *testing.T) {
	const cost = bcrypt.MinCost + 1
	b, err := NewBcrypt(cost)
	if err != nil {
		t.Fatal(err)
	}
	got, err := bcrypt.Cost(b.dummyHash)
	if err != nil {
		t.Fatal(err)
	}
	if got != cost {
		t.Fatalf("dummy hash cost = %d, want %d (same work as a real hash)", got, cost)
	}
	b.CompareDummy("anything1") // must not panic
}

func TestNewBcrypt_CostOutOfRange(t *testing.T) {
	if _, err := NewBcrypt(bcrypt.MaxCost + 1); err == nil {
		t.Fatal("want error for cost above max")
	}
}
