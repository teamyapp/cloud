package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/obs"
)

type Permission struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.Permission = (*Permission)(nil)

func (p Permission) FindPermission(ct context.Context, query entity.PermissionQuery) (entity.Permission, error) {
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
		return entity.Permission{}, dao.ErrNotFound(fmt.Sprintf(
			"permission not found: resource_type=%v, resource_id=%d, operation=%v, group_id=%d",
			query.ResourceType, query.ResourceID, query.Operation, query.GroupID))
	}

	if err != nil {
		p.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return permission, err
}

func (p Permission) FindAllPermissions(ct context.Context) ([]entity.Permission, error) {
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
		p.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			p.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (p Permission) CreatePermission(ct context.Context, permission entity.Permission) error {
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
		p.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (p Permission) DeletePermission(ct context.Context, resourceType string, resourceID uint64, operation string, groupID uint64) error {
	_, err := p.db.Exec(`
		DELETE FROM permission
		WHERE resource_type = $1 AND resource_id = $2 AND operation = $3 AND group_id = $4;
		`,
		resourceType, resourceID, operation, groupID)
	if err != nil {
		p.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewPermission(dataCollector obs.DataCollector, sqlDB *sql.DB) Permission {
	return Permission{dataCollector: dataCollector, db: sqlDB}
}
