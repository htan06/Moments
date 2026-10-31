package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestProfile(t *testing.T) *Profile {
	t.Helper()

	profile, err := NewProfile(
		1,
		"user111",
		"user1111",
	)

	require.NoError(t, err)

	return profile
}

func TestNewProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		profile, err := NewProfile(
			1,
			"user111",
			"user1111",
		)

		require.NoError(t, err)
		require.NotNil(t, profile)

		require.Equal(t, int64(1), profile.UserID())
		require.Equal(t, "user1111", profile.Username())
		require.Equal(t, "user111", profile.Name())
		require.Equal(t, int64(0), profile.FollowersCount())
		require.Equal(t, int64(0), profile.FollowingCount())
		require.Equal(t, int64(0), profile.PostsCount())
	})
}

func TestProfile_ChangeUsername(t *testing.T) {
	profile := newTestProfile(t)

	err := profile.ChangeUsername("new_username")

	require.NoError(t, err)
	require.Equal(t, "new_username", profile.Username())
}
