package entity

type AuthorizationQuery struct {
	PermissionType string
	ResourceId     string
	ResourceType   string
	UserOrGroupId  string
}

func (a AuthorizationQuery) IsValid() bool {
	return len(a.PermissionType) > 0 &&
		len(a.ResourceId) > 0 &&
		len(a.ResourceType) > 0 &&
		len(a.UserOrGroupId) > 0
}

func (a AuthorizationQuery) PermissionBindingFromAuthQuery() PermissionBinding {
	return PermissionBinding{
		PermissionType: a.PermissionType,
		ResourceId:     a.ResourceId,
		ResourceType:   a.ResourceType,
		UserOrGroupId:  a.UserOrGroupId,
	}
}
