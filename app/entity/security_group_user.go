package entity

import "time"

type SecurityGroupUser struct {
	GroupID   uint64
	UserID    uint64
	CreatedAt time.Time
}
