package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type block struct {
	typ string
	der []byte
}

func genKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func privDER(t *testing.T, k *rsa.PrivateKey) []byte {
	t.Helper()
	b, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func pubDER(t *testing.T, k *rsa.PrivateKey) []byte {
	t.Helper()
	b, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func writePEM(t *testing.T, dir, name string, blocks ...block) string {
	t.Helper()
	var sb strings.Builder
	for _, b := range blocks {
		if err := pem.Encode(&sb, &pem.Block{Type: b.typ, Bytes: b.der}); err != nil {
			t.Fatal(err)
		}
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(sb.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// fixture writes a valid key pair file set and returns the paths and keys.
type fixture struct {
	dir               string
	access, refresh   *rsa.PrivateKey
	privPath, pubPath string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	f := fixture{dir: t.TempDir(), access: genKey(t), refresh: genKey(t)}
	f.privPath = writePEM(t, f.dir, "private.pem",
		block{blockAccessPrivate, privDER(t, f.access)},
		block{blockRefreshPrivate, privDER(t, f.refresh)})
	f.pubPath = writePEM(t, f.dir, "public.pem",
		block{blockAccessPublic, pubDER(t, f.access)},
		block{blockRefreshPublic, pubDER(t, f.refresh)})
	return f
}

func newManager(t *testing.T, f fixture, now time.Time) *Manager {
	t.Helper()
	keys, err := LoadKeys(f.privPath, f.pubPath)
	if err != nil {
		t.Fatalf("LoadKeys: %v", err)
	}
	m := NewManager(keys, 15*time.Minute, 30*24*time.Hour)
	m.now = func() time.Time { return now }
	return m
}

func TestIssueAndParse_RoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := newManager(t, newFixture(t), now)
	want := Claims{UserID: 42, Username: "alice", Status: "ACTIVE"}

	pair, err := m.Issue(want)
	if err != nil {
		t.Fatal(err)
	}
	if !pair.AccessExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Errorf("access exp = %v", pair.AccessExpiresAt)
	}
	if !pair.RefreshExpiresAt.Equal(now.Add(30 * 24 * time.Hour)) {
		t.Errorf("refresh exp = %v", pair.RefreshExpiresAt)
	}

	got, err := m.ParseAccess(pair.AccessToken)
	if err != nil || got != want {
		t.Errorf("ParseAccess = %+v, %v", got, err)
	}
	got, err = m.ParseRefresh(pair.RefreshToken)
	if err != nil || got != want {
		t.Errorf("ParseRefresh = %+v, %v", got, err)
	}
}

func TestParse_AccessAndRefreshNotInterchangeable(t *testing.T) {
	m := newManager(t, newFixture(t), time.Now())
	pair, err := m.Issue(Claims{UserID: 1, Username: "a", Status: "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ParseRefresh(pair.AccessToken); !errors.Is(err, ErrInvalid) {
		t.Errorf("access as refresh: err = %v, want ErrInvalid", err)
	}
	if _, err := m.ParseAccess(pair.RefreshToken); !errors.Is(err, ErrInvalid) {
		t.Errorf("refresh as access: err = %v, want ErrInvalid", err)
	}
}

func TestParse_Expired(t *testing.T) {
	issued := time.Now()
	f := newFixture(t)
	m := newManager(t, f, issued)
	pair, err := m.Issue(Claims{UserID: 1, Username: "a", Status: "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
	m.now = func() time.Time { return issued.Add(16 * time.Minute) }
	if _, err := m.ParseAccess(pair.AccessToken); !errors.Is(err, ErrExpired) {
		t.Errorf("err = %v, want ErrExpired", err)
	}
}

func TestParse_RejectsOtherAlgorithms(t *testing.T) {
	m := newManager(t, newFixture(t), time.Now())
	claims := jwtClaims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	hs, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	for name, tok := range map[string]string{"none": none, "HS256": hs, "garbage": "a.b.c"} {
		if _, err := m.ParseAccess(tok); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", name, err)
		}
	}
}

func TestParse_MissingUserID(t *testing.T) {
	f := newFixture(t)
	m := newManager(t, f, time.Now())
	tok, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}).SignedString(f.access)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ParseAccess(tok); !errors.Is(err, ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestLoadKeys_Errors(t *testing.T) {
	f := newFixture(t)
	other := genKey(t)

	tests := []struct {
		name    string
		priv    []block
		pub     []block
		wantErr string
	}{
		{
			name:    "missing refresh private block",
			priv:    []block{{blockAccessPrivate, privDER(t, f.access)}},
			pub:     []block{{blockAccessPublic, pubDER(t, f.access)}, {blockRefreshPublic, pubDER(t, f.refresh)}},
			wantErr: "missing PEM block",
		},
		{
			name: "unexpected block type",
			priv: []block{{blockAccessPrivate, privDER(t, f.access)}, {blockRefreshPrivate, privDER(t, f.refresh)},
				{"PRIVATE KEY", privDER(t, other)}},
			pub:     []block{{blockAccessPublic, pubDER(t, f.access)}, {blockRefreshPublic, pubDER(t, f.refresh)}},
			wantErr: "unexpected PEM block",
		},
		{
			name: "PKCS1 private key",
			priv: []block{{blockAccessPrivate, x509.MarshalPKCS1PrivateKey(f.access)},
				{blockRefreshPrivate, privDER(t, f.refresh)}},
			pub:     []block{{blockAccessPublic, pubDER(t, f.access)}, {blockRefreshPublic, pubDER(t, f.refresh)}},
			wantErr: "not a PKCS#8",
		},
		{
			name:    "public does not match private",
			priv:    []block{{blockAccessPrivate, privDER(t, f.access)}, {blockRefreshPrivate, privDER(t, f.refresh)}},
			pub:     []block{{blockAccessPublic, pubDER(t, other)}, {blockRefreshPublic, pubDER(t, f.refresh)}},
			wantErr: "does not match",
		},
		{
			name:    "access and refresh share a key pair",
			priv:    []block{{blockAccessPrivate, privDER(t, f.access)}, {blockRefreshPrivate, privDER(t, f.access)}},
			pub:     []block{{blockAccessPublic, pubDER(t, f.access)}, {blockRefreshPublic, pubDER(t, f.access)}},
			wantErr: "different key pairs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			priv := writePEM(t, dir, "private.pem", tt.priv...)
			pub := writePEM(t, dir, "public.pem", tt.pub...)
			_, err := LoadKeys(priv, pub)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
