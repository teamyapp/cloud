package entity

import "time"

type ResourceUserGroupRelation struct {
	ResourceType  string
	ResourceID    uint64
	GroupID       uint64
	Key           *string
	CreatedAt     time.Time
	CreatorUserID uint64
}
