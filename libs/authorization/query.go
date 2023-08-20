package authorization

import (
	"fmt"
)

type Query struct {
	ResourceType string
	ResourceID   uint64
	Operation    string
	UserID       uint64
}

func (q Query) String() string {
	return fmt.Sprintf("[Query UserID=%v Operation=%v ResourceType=%v ResourceID=%v]", q.UserID, q.Operation, q.ResourceType, q.ResourceID)
}
