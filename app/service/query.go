package service

import (
	"time"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type ResourceTypeQuery struct {
	ResourceType      *string
	CreatorUserID     *uint64
	StartCreationTime *time.Time
	EndCreationTime   *time.Time
	Limit             *uint64
}

func queryResourceTypes(resourceTypeEntities []entity.ResourceType, resourceTypeQuery ResourceTypeQuery) []entity.ResourceType {
	return collect.Filter(resourceTypeEntities, func(resourceTypeEntity entity.ResourceType) bool {
		if resourceTypeQuery.ResourceType != nil && *resourceTypeQuery.ResourceType != resourceTypeEntity.ResourceType {
			return false
		}

		if resourceTypeQuery.CreatorUserID != nil && *resourceTypeQuery.CreatorUserID != resourceTypeEntity.CreatorUserID {
			return false
		}

		if resourceTypeQuery.StartCreationTime != nil && (*resourceTypeQuery.StartCreationTime).After(resourceTypeEntity.CreatedAt) {
			return false
		}

		if resourceTypeQuery.EndCreationTime != nil && (*resourceTypeQuery.EndCreationTime).Before(resourceTypeEntity.CreatedAt) {
			return false
		}

		return true
	})
}
