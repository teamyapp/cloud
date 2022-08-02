package entity

import "time"

type ResourceType struct {
	ResourceType  string
	CreatedAt     time.Time
	CreatorUserID uint64
}
