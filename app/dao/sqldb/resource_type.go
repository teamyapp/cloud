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

type ResourceType struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.ResourceType = (*ResourceType)(nil)

func (r ResourceType) FindResourceType(ct context.Context, resourceTypeName string) (entity.ResourceType, error) {
	resourceTypeEntity := entity.ResourceType{}
	err := r.db.QueryRow(`
		SELECT
			resource_type,
			created_at,
			creator_user_id
		FROM resource_type
		WHERE resource_type = $1;`,
		resourceTypeName).
		Scan(
			&resourceTypeEntity.ResourceTypeName,
			&resourceTypeEntity.CreatedAt,
			&resourceTypeEntity.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.ResourceType{}, dao.ErrNotFound(fmt.Sprintf(
			"resource type not found: resource_type=%v",
			resourceTypeName))
	}

	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return resourceTypeEntity, err
}

func (r ResourceType) FindAllResourceTypes(ct context.Context) ([]entity.ResourceType, error) {
	rows, err := r.db.Query(`
	SELECT
		resource_type,
		created_at,
		creator_user_id
	FROM resource_type;
`)
	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	resourceTypeEntities := make([]entity.ResourceType, 0)
	for rows.Next() {
		resourceTypeEntity := entity.ResourceType{}
		err = rows.Scan(
			&resourceTypeEntity.ResourceTypeName,
			&resourceTypeEntity.CreatedAt,
			&resourceTypeEntity.CreatorUserID,
		)
		if err != nil {
			r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		resourceTypeEntities = append(resourceTypeEntities, resourceTypeEntity)
	}

	return resourceTypeEntities, nil
}

func (r ResourceType) CreateResourceType(ct context.Context, resourceTypeEntity entity.ResourceType) error {
	_, err := r.db.Exec(`
		INSERT INTO resource_type
		(
			resource_type,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3);`,
		resourceTypeEntity.ResourceTypeName,
		resourceTypeEntity.CreatedAt,
		resourceTypeEntity.CreatorUserID,
	)
	return err
}

func (r ResourceType) DeleteResourceType(ct context.Context, resourceTypeName string) error {
	_, err := r.db.Exec(`
		DELETE FROM resource_type
		WHERE resource_type = $1;
		`,
		resourceTypeName)
	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewResourceType(dataCollector obs.DataCollector, sqlDB *sql.DB) ResourceType {
	return ResourceType{dataCollector: dataCollector, db: sqlDB}
}
