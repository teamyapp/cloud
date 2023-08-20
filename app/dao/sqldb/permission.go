package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
)

type Permission struct {
	db *sql.DB
}

var _ dao.Permission = (*Permission)(nil)

func (p Permission) FindPermission(ct context.Context, query entity.PermissionQuery) (entity.Permission, *errs.Error) {
	permission := entity.Permission{}
	err := p.db.QueryRow(`
		SELECT
			resource_type,
			resource_id,
			operation,
			group_id,
			created_at,
			creator_user_id
		FROM permission
		WHERE resource_type = $1 AND resource_id = $2 AND operation = $3 AND group_id = $4;`,
		query.ResourceType, query.ResourceID, query.Operation, query.GroupID).
		Scan(
			&permission.ResourceType,
			&permission.ResourceID,
			&permission.Operation,
			&permission.GroupID,
			&permission.CreatedAt,
			&permission.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Permission{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("permission not found: resource_type=%v, resource_id=%d, operation=%v, group_id=%d",
				query.ResourceType,
				query.ResourceID,
				query.Operation,
				query.GroupID))
	}

	if err != nil {
		return entity.Permission{}, errs.NewError(errs.Unknown, err.Error())
	}

	return permission, nil
}

func (p Permission) FindAllPermissions(ct context.Context) ([]entity.Permission, *errs.Error) {
	rows, err := p.db.Query(`
		SELECT
			resource_type,
			resource_id,
			operation,
			group_id,
			created_at,
			creator_user_id
		FROM permission;
	`)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	permissions := make([]entity.Permission, 0)
	for rows.Next() {
		permission := entity.Permission{}
		err = rows.Scan(
			&permission.ResourceType,
			&permission.ResourceID,
			&permission.Operation,
			&permission.GroupID,
			&permission.CreatedAt,
			&permission.CreatorUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (p Permission) CreatePermission(ct context.Context, permission entity.Permission) *errs.Error {
	_, err := p.db.Exec(`
		INSERT INTO permission
		(
			resource_type,
		 	resource_id,
		 	operation,
		 	group_id,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		permission.ResourceType,
		permission.ResourceID,
		permission.Operation,
		permission.GroupID,
		permission.CreatedAt,
		permission.CreatorUserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p Permission) DeletePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) *errs.Error {
	_, err := p.db.Exec(`
		DELETE FROM permission
		WHERE resource_type = $1 AND resource_id = $2 AND operation = $3 AND group_id = $4;
		`,
		resourceType, resourceID, operation, groupID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewPermission(sqlDB *sql.DB) Permission {
	return Permission{db: sqlDB}
}
