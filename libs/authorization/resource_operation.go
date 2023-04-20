package authorization

type ResourceOperation struct {
	ResourceType string
	Operation    string
	ResourceID   uint64
}
