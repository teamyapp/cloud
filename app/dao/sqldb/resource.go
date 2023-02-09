package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Resource struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindResource(ct context.Context, resourceTypeName string, resourceID uint64) (entity.Resource, *errs.Error) {
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
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"resource not found: resource_type=%v, resource_id=%d",
				resourceTypeName,
				resourceID),
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Resource{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Resource{}, internalErr
	}

	return resource, nil
}

func (r Resource) FindAllResources(ct context.Context) ([]entity.Resource, *errs.Error) {
	rows, err := r.db.Query(`
	SELECT
		resource_type,
		resource_id,
		created_at,
		creator_user_id
	FROM resource;
`)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
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
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: newInternalErr})
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (r Resource) CreateResource(ct context.Context, resource entity.Resource) *errs.Error {
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (r Resource) DeleteResource(ct context.Context, resourceTypeName string, resourceID uint64) *errs.Error {
	_, err := r.db.Exec(`
		DELETE FROM resource
		WHERE resource_type = $1 AND resource_id = $2;
		`,
		resourceTypeName, resourceID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewResource(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Resource {
	return Resource{dataCollector: dataCollector, db: sqlDB}
}
