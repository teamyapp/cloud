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

type ResourceQuery struct {
	ResourceType      *string
	ResourceID        *uint64
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

func queryResources(resources []entity.Resource, resourceQuery ResourceQuery) []entity.Resource {
	return collect.Filter(resources, func(resource entity.Resource) bool {
		if resourceQuery.ResourceType != nil && *resourceQuery.ResourceType != resource.ResourceType {
			return false
		}

		if resourceQuery.ResourceID != nil && *resourceQuery.ResourceID != resource.ResourceID {
			return false
		}

		if resourceQuery.CreatorUserID != nil && *resourceQuery.CreatorUserID != resource.CreatorUserID {
			return false
		}

		if resourceQuery.StartCreationTime != nil && (*resourceQuery.StartCreationTime).After(resource.CreatedAt) {
			return false
		}

		if resourceQuery.EndCreationTime != nil && (*resourceQuery.EndCreationTime).Before(resource.CreatedAt) {
			return false
		}

		return true
	})
}
