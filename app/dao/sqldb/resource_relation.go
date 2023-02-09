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

type ResourceRelation struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.ResourceRelation = (*ResourceRelation)(nil)

func (r ResourceRelation) FindResourceRelation(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) (entity.ResourceRelation, *errs.Error) {
	resourceRelation := entity.ResourceRelation{}
	err := r.db.QueryRow(`
		SELECT
			child_resource_type,
			child_resource_id,
			parent_resource_type,
			parent_resource_id,
			created_at,
			creator_user_id
		FROM resource_relation
		WHERE child_resource_type = $1 AND child_resource_id = $2 AND parent_resource_type = $3 AND parent_resource_id = $4;`,
		childResourceType, childResourceID, parentResourceType, parentResourceID).
		Scan(
			&resourceRelation.ChildResourceType,
			&resourceRelation.ChildResourceID,
			&resourceRelation.ParentResourceType,
			&resourceRelation.ParentResourceID,
			&resourceRelation.CreatedAt,
			&resourceRelation.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"resource relation not found: child_resource_type=%v, child_resource_id=%d, parent_resource_type=%v, parent_resource_id=%d",
				childResourceType,
				childResourceID,
				parentResourceType,
				parentResourceID),
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.ResourceRelation{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		r.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.ResourceRelation{}, internalErr
	}

	return resourceRelation, nil
}

func (r ResourceRelation) FindResourceRelations(ct context.Context, childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, *errs.Error) {
	rows, err := r.db.Query(`
		SELECT
			child_resource_type,
			child_resource_id,
			parent_resource_type,
			parent_resource_id,
			created_at,
			creator_user_id
		FROM resource_relation
		WHERE child_resource_type = $1 AND child_resource_id = $2;`,
		childResourceType, childResourceID)
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
	resourceRelations := make([]entity.ResourceRelation, 0)
	for rows.Next() {
		resourceRelation := entity.ResourceRelation{}
		err = rows.Scan(
			&resourceRelation.ChildResourceType,
			&resourceRelation.ChildResourceID,
			&resourceRelation.ParentResourceType,
			&resourceRelation.ParentResourceID,
			&resourceRelation.CreatedAt,
			&resourceRelation.CreatorUserID,
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

		resourceRelations = append(resourceRelations, resourceRelation)
	}

	return resourceRelations, nil
}

func (r ResourceRelation) FindAllResourceRelations(ct context.Context) ([]entity.ResourceRelation, *errs.Error) {
	rows, err := r.db.Query(`
		SELECT
			child_resource_type,
			child_resource_id,
			parent_resource_type,
			parent_resource_id,
			created_at,
			creator_user_id
		FROM resource_relation;
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
	resourceRelations := make([]entity.ResourceRelation, 0)
	for rows.Next() {
		resourceRelation := entity.ResourceRelation{}
		err = rows.Scan(
			&resourceRelation.ChildResourceType,
			&resourceRelation.ChildResourceID,
			&resourceRelation.ParentResourceType,
			&resourceRelation.ParentResourceID,
			&resourceRelation.CreatedAt,
			&resourceRelation.CreatorUserID,
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

		resourceRelations = append(resourceRelations, resourceRelation)
	}

	return resourceRelations, nil
}

func (r ResourceRelation) CreateResourceRelation(ct context.Context, resourceRelation entity.ResourceRelation) *errs.Error {
	_, err := r.db.Exec(`
		INSERT INTO resource_relation
		(
		 	child_resource_type,
		 	child_resource_id,
		 	parent_resource_type,
		 	parent_resource_id,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		resourceRelation.ChildResourceType,
		resourceRelation.ChildResourceID,
		resourceRelation.ParentResourceType,
		resourceRelation.ParentResourceID,
		resourceRelation.CreatedAt,
		resourceRelation.CreatorUserID,
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

func (r ResourceRelation) DeleteResourceRelation(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) *errs.Error {
	_, err := r.db.Exec(`
		DELETE FROM resource_relation
		WHERE 
		    child_resource_type = $1 AND 
		    child_resource_id = $2 AND 
		    parent_resource_type = $3 AND 
		    parent_resource_id = $4;
		`,
		childResourceType, childResourceID, parentResourceType, parentResourceID)

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

func NewResourceRelation(dataCollector telemetry.DataCollector, sqlDB *sql.DB) ResourceRelation {
	return ResourceRelation{dataCollector: dataCollector, db: sqlDB}
}
