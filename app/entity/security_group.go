package entity

import "time"

type SecurityGroup struct {
	ID          uint64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
