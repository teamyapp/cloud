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

type Resource struct {
	db *sql.DB
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
		return entity.Resource{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("resource not found: resource_type=%v, resource_id=%d", resourceTypeName, resourceID))
	}

	if err != nil {
		return entity.Resource{}, errs.NewError(errs.Unknown, err.Error())
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
		return nil, errs.NewError(errs.Unknown, err.Error())
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
			return nil, errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewResource(sqlDB *sql.DB) Resource {
	return Resource{db: sqlDB}
}
