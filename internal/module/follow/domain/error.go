package domain

import "github.com/htan06/echo-messenger-rest-api/internal/errs"

const (
	ReceiverNotFound      errs.ErrorCode = "RECEIVER_NOT_FOUND"
	FriendRequestNotFound errs.ErrorCode = "FRIEND_REQEST_NOT_FOUND"

	FriendRequestInvalid errs.ErrorCode = "FRIEND_REQUEST_INVALID"
	FollowInvalid        errs.ErrorCode = "FOLLOW_INVALID"

	UserNotFound errs.ErrorCode = "USER_NOT_FOUND"

	PaginationInvalid errs.ErrorCode = "PAGINATION_INVALID"
)
