package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/htan06/Moments/application/internal/user/command"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// maxUpdateRetries bounds optimistic-lock retries when the same session is
// updated concurrently.
const maxUpdateRetries = 5

// VerificationSessions stores sessions as JSON with a TTL equal to their
// remaining lifetime.
type VerificationSessions struct {
	rdb *redis.Client
	now func() time.Time
}

func NewVerificationSessions(rdb *redis.Client) *VerificationSessions {
	return &VerificationSessions{rdb: rdb, now: time.Now}
}

// record is the stored form. Only the OTP hash is stored, never the OTP.
type record struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	Email          string    `json:"email"`
	OTPHash        string    `json:"otp_hash"`
	Status         string    `json:"status"`
	FailedAttempts int       `json:"failed_attempts"`
	LastSentAt     time.Time `json:"last_sent_at"`
	IsDummy        bool      `json:"is_dummy"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func (r *VerificationSessions) Create(ctx context.Context, s domain.VerificationSession) error {
	ttl := s.ExpiresAt.Sub(r.now())
	if ttl <= 0 {
		return errors.New("session: create expired verification session")
	}
	b, err := json.Marshal(toRecord(s))
	if err != nil {
		return fmt.Errorf("session: encode: %w", err)
	}
	ok, err := r.rdb.SetNX(ctx, verificationKey(s.ID), b, ttl).Result()
	if err != nil {
		return fmt.Errorf("session: create: %w", err)
	}
	if !ok {
		return errors.New("session: verification session id collision")
	}
	return nil
}

func (r *VerificationSessions) Get(ctx context.Context, id string) (domain.VerificationSession, error) {
	return decode(r.rdb.Get(ctx, verificationKey(id)).Bytes())
}

func (r *VerificationSessions) Delete(ctx context.Context, id string) error {
	if err := r.rdb.Del(ctx, verificationKey(id)).Err(); err != nil {
		return fmt.Errorf("session: delete: %w", err)
	}
	return nil
}

// Update runs fn under WATCH so concurrent updates of one session (e.g. two
// OTP attempts at once) cannot lose a failed-attempt increment.
func (r *VerificationSessions) Update(ctx context.Context, id string,
	fn func(*domain.VerificationSession) (command.Change, error)) error {
	key := verificationKey(id)
	for range maxUpdateRetries {
		var fnErr error
		err := r.rdb.Watch(ctx, func(tx *redis.Tx) error {
			s, err := decode(tx.Get(ctx, key).Bytes())
			if err != nil {
				return err
			}
			change, ferr := fn(&s)
			fnErr = ferr
			switch change {
			case command.SaveSession:
				b, err := json.Marshal(toRecord(s))
				if err != nil {
					return fmt.Errorf("session: encode: %w", err)
				}
				_, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
					// KEEPTTL: updates never extend the session.
					p.SetArgs(ctx, key, b, redis.SetArgs{KeepTTL: true, Mode: "XX"})
					return nil
				})
				return err
			case command.DeleteSession:
				_, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
					p.Del(ctx, key)
					return nil
				})
				return err
			default:
				return nil
			}
		}, key)
		if errors.Is(err, redis.TxFailedErr) {
			continue // the session changed underneath: retry with fresh data
		}
		if err != nil {
			if errors.Is(err, domain.ErrSessionNotFound) {
				return err
			}
			return fmt.Errorf("session: update: %w", err)
		}
		return fnErr
	}
	return errors.New("session: update: too much contention")
}

func decode(b []byte, err error) (domain.VerificationSession, error) {
	if errors.Is(err, redis.Nil) {
		return domain.VerificationSession{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.VerificationSession{}, fmt.Errorf("session: get: %w", err)
	}
	var rec record
	if err := json.Unmarshal(b, &rec); err != nil {
		return domain.VerificationSession{}, fmt.Errorf("session: decode: %w", err)
	}
	return domain.VerificationSession{
		ID:             rec.ID,
		Type:           domain.VerificationType(rec.Type),
		Email:          rec.Email,
		OTPHash:        rec.OTPHash,
		Status:         domain.VerificationStatus(rec.Status),
		FailedAttempts: rec.FailedAttempts,
		LastSentAt:     rec.LastSentAt,
		IsDummy:        rec.IsDummy,
		CreatedAt:      rec.CreatedAt,
		ExpiresAt:      rec.ExpiresAt,
	}, nil
}

func toRecord(s domain.VerificationSession) record {
	return record{
		ID:             s.ID,
		Type:           string(s.Type),
		Email:          s.Email,
		OTPHash:        s.OTPHash,
		Status:         string(s.Status),
		FailedAttempts: s.FailedAttempts,
		LastSentAt:     s.LastSentAt,
		IsDummy:        s.IsDummy,
		CreatedAt:      s.CreatedAt,
		ExpiresAt:      s.ExpiresAt,
	}
}
