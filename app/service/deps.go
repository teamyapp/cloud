package service

import (
	"github.com/teamyapp/cloud/app/dao"
)

type Dependencies struct {
	permissionDao        dao.Permission
	securityGroupUserDao dao.SecurityGroupUser
	resourceOperation    dao.ResourceOperation
	resource             dao.Resource
}

func NewDependencies(
	permissionDao dao.Permission,
	securityGroupUserDao dao.SecurityGroupUser,
	resourceOperationDao dao.ResourceOperation,
	resourceDao dao.Resource) Dependencies {
	return Dependencies{
		permissionDao:        permissionDao,
		securityGroupUserDao: securityGroupUserDao,
		resourceOperation:    resourceOperationDao,
		resource:             resourceDao,
	}
}
