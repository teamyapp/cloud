package entity

import "time"

type Resource struct {
	ResourceType  string
	ResourceID    uint64
	CreatedAt     time.Time
	CreatorUserID uint64
}
