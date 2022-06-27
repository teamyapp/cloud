package entity

type ResourceRelation struct {
	ID                 uint64
	ResourceType       string
	ParentResourceID   uint64
	ParentResourceType string
}
