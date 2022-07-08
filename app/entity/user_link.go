package entity

type UserLink struct {
	AuthProvider      string `json:"auth_provider"`
	InternalUserID    uint64 `json:"internal_user_id"`
	ExternalUserID    string `json:"external_user_id"`
	ExternalUserLabel string `json:"external_user_label"`
}
