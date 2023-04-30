package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type ResourceUserGroupRelation struct {
	db *sql.DB
}

var _ dao.ResourceUserGroupRelation = (*ResourceUserGroupRelation)(nil)

func (r ResourceUserGroupRelation) FindResourceUserGroupRelationByUserGroup(ct context.Context, userGroupID uint64) ([]entity.ResourceUserGroupRelation, *errs.Error) {
	rows, err := r.db.Query(`
		SELECT
			resource_type,
			resource_id,
			user_group_id,
			created_at,
			creator_user_id
		FROM resource_user_group_relation
		WHERE user_group_id = $1;
	`, userGroupID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	resourceUserGroupRelations := make([]entity.ResourceUserGroupRelation, 0)
	for rows.Next() {
		resourceUserGroupRelation := entity.ResourceUserGroupRelation{}
		err = rows.Scan(
			&resourceUserGroupRelation.ResourceType,
			&resourceUserGroupRelation.ResourceID,
			&resourceUserGroupRelation.GroupID,
			&resourceUserGroupRelation.Key,
			&resourceUserGroupRelation.CreatedAt,
			&resourceUserGroupRelation.CreatorUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		resourceUserGroupRelations = append(resourceUserGroupRelations, resourceUserGroupRelation)
	}

	return resourceUserGroupRelations, nil
}

func (r ResourceUserGroupRelation) FindResourceUserGroupRelationByResource(ct context.Context, resourceType string, resourceID uint64) ([]entity.ResourceUserGroupRelation, *errs.Error) {
	rows, err := r.db.Query(`
		SELECT
			resource_type,
			resource_id,
			user_group_id,
			created_at,
			creator_user_id
		FROM resource_user_group_relation
		WHERE resource_type = $1 AND resource_id = $2;
	`, resourceType, resourceID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	resourceUserGroupRelations := make([]entity.ResourceUserGroupRelation, 0)
	for rows.Next() {
		resourceUserGroupRelation := entity.ResourceUserGroupRelation{}
		err = rows.Scan(
			&resourceUserGroupRelation.ResourceType,
			&resourceUserGroupRelation.ResourceID,
			&resourceUserGroupRelation.GroupID,
			&resourceUserGroupRelation.Key,
			&resourceUserGroupRelation.CreatedAt,
			&resourceUserGroupRelation.CreatorUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		resourceUserGroupRelations = append(resourceUserGroupRelations, resourceUserGroupRelation)
	}

	return resourceUserGroupRelations, nil
}

func (r ResourceUserGroupRelation) FindAllResourceUserGroupRelations(ct context.Context) ([]entity.ResourceUserGroupRelation, *errs.Error) {
	rows, err := r.db.Query(`
		SELECT
			resource_type,
			resource_id,
			user_group_id,
			created_at,
			creator_user_id
		FROM resource_user_group_relation;
	`)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	resourceUserGroupRelations := make([]entity.ResourceUserGroupRelation, 0)
	for rows.Next() {
		resourceUserGroupRelation := entity.ResourceUserGroupRelation{}
		err = rows.Scan(
			&resourceUserGroupRelation.ResourceType,
			&resourceUserGroupRelation.ResourceID,
			&resourceUserGroupRelation.GroupID,
			&resourceUserGroupRelation.Key,
			&resourceUserGroupRelation.CreatedAt,
			&resourceUserGroupRelation.CreatorUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		resourceUserGroupRelations = append(resourceUserGroupRelations, resourceUserGroupRelation)
	}

	return resourceUserGroupRelations, nil
}

func (r ResourceUserGroupRelation) CreateResourceUserGroupRelation(ct context.Context, relation entity.ResourceUserGroupRelation) *errs.Error {
	_, err := r.db.Exec(`
		INSERT INTO resource_user_group_relation
		(
			resource_type,
		 	resource_id,
		 	user_group_id,
		 	key,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		relation.ResourceType,
		relation.ResourceID,
		relation.GroupID,
		relation.Key,
		relation.CreatedAt,
		relation.CreatorUserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (r ResourceUserGroupRelation) DeleteResourceUserGroupRelation(ct context.Context, resourceType string, resourceID uint64, userGroupID uint64) *errs.Error {
	_, err := r.db.Exec(`
		DELETE FROM resource_user_group_relation
		WHERE resource_type = $1 AND resource_id = $2 AND user_group_id = $3;
		`,
		resourceType, resourceID, userGroupID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewResourceUserGroupRelation(sqlDB *sql.DB) ResourceUserGroupRelation {
	return ResourceUserGroupRelation{db: sqlDB}
}
