package dto

type FollowPayload struct {
	FollowerID int64 `json:"follower_id"`
}

type UnFollowPayload struct {
	FollowerID int64 `json:"follower_id"`
}
