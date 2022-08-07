package service

import (
	"time"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type ResourceTypeQuery struct {
	ResourceTypeName  *string
	CreatorUserID     *uint64
	StartCreationTime *time.Time
	EndCreationTime   *time.Time
	Limit             *uint64
}

type ResourceQuery struct {
	ResourceTypeName  *string
	ResourceID        *uint64
	CreatorUserID     *uint64
	StartCreationTime *time.Time
	EndCreationTime   *time.Time
	Limit             *uint64
}

type ResourceRelationQuery struct {
	ChildResourceType  *string
	ChildResourceID    *uint64
	ParentResourceType *string
	ParentResourceID   *uint64
	CreatorUserID      *uint64
	StartCreationTime  *time.Time
	EndCreationTime    *time.Time
	Limit              *uint64
}

func queryResourceTypes(resourceTypeEntities []entity.ResourceType, resourceTypeQuery ResourceTypeQuery) []entity.ResourceType {
	return collect.Filter(resourceTypeEntities, func(resourceTypeEntity entity.ResourceType) bool {
		if resourceTypeQuery.ResourceTypeName != nil && *resourceTypeQuery.ResourceTypeName != resourceTypeEntity.ResourceTypeName {
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
		if resourceQuery.ResourceTypeName != nil && *resourceQuery.ResourceTypeName != resource.ResourceTypeName {
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

func queryResourceRelations(resourceRelations []entity.ResourceRelation, resourceRelationQuery ResourceRelationQuery) []entity.ResourceRelation {
	return collect.Filter(resourceRelations, func(resourceRelation entity.ResourceRelation) bool {
		if resourceRelationQuery.ChildResourceID != nil && *resourceRelationQuery.ChildResourceID != resourceRelation.ChildResourceID {
			return false
		}

		if resourceRelationQuery.ChildResourceType != nil && *resourceRelationQuery.ChildResourceType != resourceRelation.ChildResourceType {
			return false
		}

		if resourceRelationQuery.ParentResourceID != nil && *resourceRelationQuery.ParentResourceID != resourceRelation.ParentResourceID {
			return false
		}

		if resourceRelationQuery.ParentResourceType != nil && *resourceRelationQuery.ParentResourceType != resourceRelation.ParentResourceType {
			return false
		}

		if resourceRelationQuery.CreatorUserID != nil && *resourceRelationQuery.CreatorUserID != resourceRelation.CreatorUserID {
			return false
		}

		if resourceRelationQuery.StartCreationTime != nil && (*resourceRelationQuery.StartCreationTime).After(resourceRelation.CreatedAt) {
			return false
		}

		if resourceRelationQuery.EndCreationTime != nil && (*resourceRelationQuery.EndCreationTime).Before(resourceRelation.CreatedAt) {
			return false
		}

		return true
	})
}
