package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/htan06/Moments/application/internal/user/command"
)

// LoginAttempts counts failed logins in a fixed window that starts at the
// first failure (USER-R3). The counter is never reset by a successful login;
// it disappears when the window ends.
type LoginAttempts struct {
	rdb    *redis.Client
	max    int64
	window time.Duration
}

func NewLoginAttempts(rdb *redis.Client, max int, window time.Duration) *LoginAttempts {
	return &LoginAttempts{rdb: rdb, max: int64(max), window: window}
}

func (l *LoginAttempts) Blocked(ctx context.Context, s command.LoginSubject) (time.Duration, error) {
	key := loginFailKey(s)
	n, err := l.rdb.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("session: login attempts: %w", err)
	}
	if n < l.max {
		return 0, nil
	}
	ttl, err := l.rdb.PTTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("session: login attempts ttl: %w", err)
	}
	if ttl <= 0 {
		return 0, nil // expired between the two calls
	}
	return ttl, nil
}

// incrScript increments the counter and starts the window on the first failure.
var incrScript = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return n`)

func (l *LoginAttempts) Fail(ctx context.Context, s command.LoginSubject) error {
	if err := incrScript.Run(ctx, l.rdb, []string{loginFailKey(s)}, l.window.Milliseconds()).Err(); err != nil {
		return fmt.Errorf("session: login fail: %w", err)
	}
	return nil
}

func loginFailKey(s command.LoginSubject) string {
	if s.AccountID > 0 {
		return loginFailAccountKey(s.AccountID)
	}
	return loginFailIdentifierKey(s.Identifier)
}
