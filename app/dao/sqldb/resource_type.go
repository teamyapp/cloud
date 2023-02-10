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

type ResourceType struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.ResourceType = (*ResourceType)(nil)

func (r ResourceType) FindResourceType(ct context.Context, resourceTypeName string) (entity.ResourceType, *errs.Error) {
	resourceType := entity.ResourceType{}
	err := r.db.QueryRow(`
		SELECT
			resource_type,
			created_at,
			creator_user_id
		FROM resource_type
		WHERE resource_type = $1;`,
		resourceTypeName).
		Scan(
			&resourceType.ResourceTypeName,
			&resourceType.CreatedAt,
			&resourceType.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"resource type not found: resource_type=%v",
				resourceTypeName),
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ResourceType{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ResourceType{}, internalErr
	}

	return resourceType, nil
}

func (r ResourceType) FindAllResourceTypes(ct context.Context) ([]entity.ResourceType, *errs.Error) {
	rows, err := r.db.Query(`
	SELECT
		resource_type,
		created_at,
		creator_user_id
	FROM resource_type;
`)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	resourceTypeEntities := make([]entity.ResourceType, 0)
	for rows.Next() {
		resourceTypeEntity := entity.ResourceType{}
		err = rows.Scan(
			&resourceTypeEntity.ResourceTypeName,
			&resourceTypeEntity.CreatedAt,
			&resourceTypeEntity.CreatorUserID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			r.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		resourceTypeEntities = append(resourceTypeEntities, resourceTypeEntity)
	}

	return resourceTypeEntities, nil
}

func (r ResourceType) CreateResourceType(ct context.Context, resourceTypeEntity entity.ResourceType) *errs.Error {
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

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (r ResourceType) DeleteResourceType(ct context.Context, resourceTypeName string) *errs.Error {
	_, err := r.db.Exec(`
		DELETE FROM resource_type
		WHERE resource_type = $1;
		`,
		resourceTypeName)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewResourceType(dataCollector telemetry.DataCollector, sqlDB *sql.DB) ResourceType {
	return ResourceType{dataCollector: dataCollector, db: sqlDB}
}
