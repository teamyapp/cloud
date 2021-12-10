package entity

import (
	"time"

	"github.com/teamyapp/one/entity"
)

type User struct {
	ID        entity.ID
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type ExternalUser struct {
	ID string
}
