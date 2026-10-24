package domain

import "github.com/htan06/Moments/post/internal/errs"

const (
	FollowInvalid        errs.ErrorCode = "FOLLOW_INVALID"

	UserNotFound errs.ErrorCode = "USER_NOT_FOUND"
	MediaIDEmpty errs.ErrorCode = ""

	MediaCountInvalid  errs.ErrorCode = "MEDIA_COUNT_INVALID"
	AspectRatioInvalid errs.ErrorCode = "ASPECT_RATIO_INVALID"
)
