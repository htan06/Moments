package user

import "github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"

type ProfileResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarURL      string `json:"avatar_url"`
	Bio            string `json:"bio"`
	FollowersCount int64  `json:"followers_count"`
	FollowingCount int64  `json:"following_count"`
	PostsCount     int64  `json:"posts_count"`
}

func ToProfileResponse(profile *domain.UserProfile) *ProfileResponse {
	if profile == nil {
		return nil
	}

	avatarURL := ""
	if profile.AvatarURL != nil {
		avatarURL = *profile.AvatarURL
	}

	bio := ""
	if profile.Bio != nil {
		bio = *profile.Bio
	}

	return &ProfileResponse{
		ID:             profile.ID,
		Username:       profile.Username,
		Name:           profile.Name,
		AvatarURL:      avatarURL,
		Bio:            bio,
		FollowersCount: profile.FollowersCount,
		FollowingCount: profile.FollowingCount,
		PostsCount:     profile.PostsCount,
	}
}

type UpdateProfileReq struct {
	Username  *string `json:"username"`
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}
