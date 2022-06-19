package dao_test

import (
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type SecurityGroupUser struct {
	securityGroupUsers []entity.SecurityGroupUser
}

var _ dao.SecurityGroupUser = (*SecurityGroupUser)(nil)

func (s SecurityGroupUser) FindGroupIDsByUserID(userID uint64) ([]uint64, error) {
	groupIDs := make([]uint64, 0)
	for _, securityGroupUser := range s.securityGroupUsers {
		if securityGroupUser.UserID == userID {
			groupIDs = append(groupIDs, securityGroupUser.GroupID)
		}
	}

	return groupIDs, nil
}

func NewSecurityGroupUser(securityGroupUsers []entity.SecurityGroupUser) SecurityGroupUser {
	return SecurityGroupUser{
		securityGroupUsers: securityGroupUsers,
	}
}
