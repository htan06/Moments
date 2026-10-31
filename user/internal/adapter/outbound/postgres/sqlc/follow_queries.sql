-- name: CreateFollow :exec
INSERT INTO social.follows (follower_id, following_id) 
				VALUES ($1, $2);

-- name: RemoveByFollowerIDAndFollowingID :exec
DELETE FROM social.follows WHERE follower_id = $1 AND following_id = $2;