package entity

type PermissionQuery struct {
	ResourceType string
	ResourceID   uint64
	Operation    string
	GroupID      uint64
}
