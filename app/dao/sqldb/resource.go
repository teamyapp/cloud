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

type Resource struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindResource(ct context.Context, resourceTypeName string, resourceID uint64) (entity.Resource, error) {
	resource := entity.Resource{}
	err := r.db.QueryRow(`
		SELECT
			resource_type,
			resource_id,
			created_at,
			creator_user_id
		FROM resource
		WHERE resource_type = $1 AND resource_id = $2;`,
		resourceTypeName, resourceID).
		Scan(
			&resource.ResourceTypeName,
			&resource.ResourceID,
			&resource.CreatedAt,
			&resource.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Resource{}, dao.ErrNotFound(fmt.Sprintf(
			"resource not found: resource_type=%v, resource_id=%d",
			resourceTypeName, resourceID))
	}

	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return resource, err
}

func (r Resource) FindAllResources(ct context.Context) ([]entity.Resource, error) {
	rows, err := r.db.Query(`
	SELECT
		resource_type,
		resource_id,
		created_at,
		creator_user_id
	FROM resource;
`)
	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	resources := make([]entity.Resource, 0)
	for rows.Next() {
		resource := entity.Resource{}
		err = rows.Scan(
			&resource.ResourceTypeName,
			&resource.ResourceID,
			&resource.CreatedAt,
			&resource.CreatorUserID,
		)
		if err != nil {
			r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (r Resource) CreateResource(ct context.Context, resource entity.Resource) error {
	_, err := r.db.Exec(`
		INSERT INTO resource
		(
			resource_type,
		 	resource_id,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4);`,
		resource.ResourceTypeName,
		resource.ResourceID,
		resource.CreatedAt,
		resource.CreatorUserID,
	)

	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (r Resource) DeleteResource(ct context.Context, resourceTypeName string, resourceID uint64) error {
	_, err := r.db.Exec(`
		DELETE FROM resource
		WHERE resource_type = $1 AND resource_id = $2;
		`,
		resourceTypeName, resourceID)
	if err != nil {
		r.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewResource(dataCollector obs.DataCollector, sqlDB *sql.DB) Resource {
	return Resource{dataCollector: dataCollector, db: sqlDB}
}
