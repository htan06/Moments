package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProfile(t *testing.T) {
	tests := []struct {
		testName string
		userID   int64
		username string
		name     string
		wantErr  error
	}{
		// ========================
		// Valid cases
		// ========================
		{
			testName: "valid profile",
			userID:   1,
			username: "valid_user",
			name:     "Valid Name",
			wantErr:  nil,
		},
		{
			testName: "invalid userID zero",
			userID:   0,
			username: "valid_user",
			name:     "Valid Name",
			wantErr:  ErrUserIDInvalid,
		},
		{
			testName: "invalid userID negative",
			userID:   -1,
			username: "valid_user",
			name:     "Valid Name",
			wantErr:  ErrUsernameInvalid,
		},
		{
			testName: "username too short",
			userID:   1,
			username: "user",
			name:     "Valid Name",
			wantErr:  ErrUsernameInvalid,
		},
		{
			testName: "username too long",
			userID:   1,
			username: strings.Repeat("a", 31),
			name:     "Valid Name",
			wantErr:  ErrUsernameInvalid,
		},
		{
			testName: "username contains special character",
			userID:   1,
			username: "user@name",
			name:     "Valid Name",
			wantErr:  ErrUsernameInvalid,
		},
		{
			testName: "username contains whitespace",
			userID:   1,
			username: "user name",
			name:     "Valid Name",
			wantErr:  ErrUsernameInvalid,
		},
		{
			testName: "name too short",
			userID:   1,
			username: "valid_user",
			name:     "A",
			wantErr:  ErrNameInvalid,
		},
		{
			testName: "name too long",
			userID:   1,
			username: "valid_user",
			name:     strings.Repeat("a", 101),
			wantErr:  ErrNameInvalid,
		},
		{
			testName: "empty username",
			userID:   1,
			username: "",
			name:     "Valid Name",
			wantErr:  ErrUsernameInvalid,
		},
		{
			testName: "empty name",
			userID:   1,
			username: "valid_user",
			name:     "",
			wantErr:  ErrNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := NewProfile(
				tt.userID,
				tt.name,
				tt.username,
			)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if got != nil {
					t.Fatal("expected nil Profile when error occurs")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil {
				t.Fatal("expected Profile, got nil")
			}

			// Verify fields
			assert.Equal(t, tt.userID, got.UserID())
			assert.Equal(t, tt.name, got.Name())
			assert.Equal(t, tt.username, got.Username())
			assert.EqualValues(t, 0, got.FollowersCount())
			assert.EqualValues(t, 0, got.FollowingCount())
			assert.EqualValues(t, 0, got.PostsCount())
			assert.Equal(t, false, got.CreatedAt().IsZero())
			assert.Equal(t, false, got.UpdatedAt().IsZero())
		})
	}
}
