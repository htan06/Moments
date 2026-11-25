package command

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/platform/token"
)

var (
	testNow    = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	testPolicy = domain.OTPPolicy{TTL: 10 * time.Minute, ResendCooldown: time.Minute, MaxAttempts: 5}
	testLog    = slog.New(slog.NewTextHandler(io.Discard, nil))
)

// syncAsync runs background work inline so tests see its effect.
func syncAsync(fn func(ctx context.Context)) { fn(context.Background()) }

type fakeRandom struct {
	otp string
	n   int
}

func (r *fakeRandom) SessionID() (string, error) {
	r.n++
	return "sid" + strconv.Itoa(r.n), nil
}
func (r *fakeRandom) OTP() (string, error) { return r.otp, nil }

type fakeAccounts struct {
	byID      map[int64]domain.Account
	createErr error
	deleteErr error
	nextID    int64
}

func newFakeAccounts(accounts ...domain.Account) *fakeAccounts {
	f := &fakeAccounts{byID: map[int64]domain.Account{}, nextID: 100}
	for _, a := range accounts {
		f.byID[a.ID] = a
	}
	return f
}

func (f *fakeAccounts) EmailExists(_ context.Context, email string) (bool, error) {
	_, err := f.ByEmail(context.Background(), email)
	return err == nil, nil
}

func (f *fakeAccounts) Create(_ context.Context, a domain.Account) (int64, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.nextID++
	a.ID = f.nextID
	f.byID[a.ID] = a
	return a.ID, nil
}

func (f *fakeAccounts) Delete(_ context.Context, id int64) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.byID, id)
	return nil
}

// fakeProfiles stands in for the user module: it trims names like the real
// rule and records created profiles.
type fakeProfiles struct {
	names     map[int64]string
	usernames map[int64]string
	createErr error
}

func newFakeProfiles() *fakeProfiles {
	return &fakeProfiles{names: map[int64]string{}, usernames: map[int64]string{}}
}

func (f *fakeProfiles) ValidateName(name string) (string, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return "", domain.ErrInvalidName
	}
	return n, nil
}

func (f *fakeProfiles) CreateProfile(_ context.Context, userID int64, username, name string) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.names[userID] = name
	f.usernames[userID] = username
	return nil
}

func (f *fakeAccounts) ByID(_ context.Context, id int64) (domain.Account, error) {
	if a, ok := f.byID[id]; ok {
		return a, nil
	}
	return domain.Account{}, domain.ErrAccountNotFound
}

func (f *fakeAccounts) ByEmail(_ context.Context, email string) (domain.Account, error) {
	for _, a := range f.byID {
		if a.Email == email {
			return a, nil
		}
	}
	return domain.Account{}, domain.ErrAccountNotFound
}

func (f *fakeAccounts) ByUsername(_ context.Context, username string) (domain.Account, error) {
	for _, a := range f.byID {
		if a.Username == username {
			return a, nil
		}
	}
	return domain.Account{}, domain.ErrAccountNotFound
}

type fakeSessions struct {
	m map[string]domain.VerificationSession
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{m: map[string]domain.VerificationSession{}}
}

func (f *fakeSessions) Create(_ context.Context, s domain.VerificationSession) error {
	f.m[s.ID] = s
	return nil
}

func (f *fakeSessions) Update(_ context.Context, id string, fn func(*domain.VerificationSession) (Change, error)) error {
	s, ok := f.m[id]
	if !ok {
		return domain.ErrSessionNotFound
	}
	change, err := fn(&s)
	switch change {
	case SaveSession:
		f.m[id] = s
	case DeleteSession:
		delete(f.m, id)
	}
	return err
}

func (f *fakeSessions) Delete(_ context.Context, id string) error {
	delete(f.m, id)
	return nil
}

type sentOTP struct {
	email, otp string
}

type fakeSender struct {
	sent []sentOTP
	err  error
}

func (f *fakeSender) SendOTP(_ context.Context, email, otp string, _ time.Time) error {
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, sentOTP{email, otp})
	return nil
}

// fakePasswords "hashes" by prefixing; compared records every hash it checked
// and hashed counts Hash calls.
type fakePasswords struct {
	compared []string
	hashed   int
}

func (f *fakePasswords) Hash(p string) (string, error) {
	f.hashed++
	return "hash:" + p, nil
}
func (f *fakePasswords) Matches(hash, p string) bool {
	f.compared = append(f.compared, hash)
	return hash == "hash:"+p
}

// dummyHash is what CompareDummy records, so tests can see it ran.
const dummyHash = "hash:dummy"

func (f *fakePasswords) CompareDummy(string) { f.compared = append(f.compared, dummyHash) }

type fakeAttempts struct {
	fails   map[LoginSubject]int
	blocked map[LoginSubject]time.Duration
}

func newFakeAttempts() *fakeAttempts {
	return &fakeAttempts{fails: map[LoginSubject]int{}, blocked: map[LoginSubject]time.Duration{}}
}

func (f *fakeAttempts) Blocked(_ context.Context, s LoginSubject) (time.Duration, error) {
	return f.blocked[s], nil
}

func (f *fakeAttempts) Fail(_ context.Context, s LoginSubject) error {
	f.fails[s]++
	return nil
}

type fakeTokens struct {
	issued []token.Claims
	parse  map[string]token.Claims
}

func (f *fakeTokens) Issue(c token.Claims) (token.Pair, error) {
	f.issued = append(f.issued, c)
	return token.Pair{AccessToken: "access-" + c.Username, RefreshToken: "refresh-" + c.Username}, nil
}

func (f *fakeTokens) ParseRefresh(s string) (token.Claims, error) {
	if c, ok := f.parse[s]; ok {
		return c, nil
	}
	return token.Claims{}, errors.New("bad token")
}

func activeAccount(id int64, email, username, password string) domain.Account {
	return domain.Account{ID: id, Email: email, Username: username, PasswordHash: "hash:" + password, Status: domain.StatusActive}
}
