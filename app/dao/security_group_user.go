package dao

type SecurityGroupUser interface {
	FindGroupIDsByUserID(userID uint64) ([]uint64, error)
}
