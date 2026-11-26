package query

import (
	"context"
	"errors"
	"testing"

	"github.com/htan06/Moments/application/internal/account/domain"
)

type fakeStatuses struct {
	m   map[int64]AccountStatusRow
	err error
}

func (f fakeStatuses) Status(_ context.Context, id int64) (AccountStatusRow, error) {
	if f.err != nil {
		return AccountStatusRow{}, f.err
	}
	if r, ok := f.m[id]; ok {
		return r, nil
	}
	return AccountStatusRow{}, domain.ErrAccountNotFound
}

func TestAccountStatus(t *testing.T) {
	store := fakeStatuses{m: map[int64]AccountStatusRow{
		1: {Username: "alice", Status: domain.StatusActive},
		2: {Username: "bob", Status: "BANNED"}, // a future status
	}}
	tests := []struct {
		name string
		id   int64
		want AccountStatusResult
	}{
		{"active", 1, AccountStatusResult{Exists: true, Active: true, Username: "alice"}},
		{"not active", 2, AccountStatusResult{Exists: true, Active: false, Username: "bob"}},
		{"missing", 9, AccountStatusResult{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccountStatusHandler(store).Handle(context.Background(), AccountStatus{UserID: tt.id})
			if err != nil || got != tt.want {
				t.Fatalf("Handle = %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
}

func TestAccountStatus_StoreError(t *testing.T) {
	store := fakeStatuses{err: errors.New("db down")}
	if _, err := NewAccountStatusHandler(store).Handle(context.Background(), AccountStatus{UserID: 1}); err == nil {
		t.Fatal("want error")
	}
}
