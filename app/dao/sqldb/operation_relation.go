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

type OperationRelation struct {
	db *sql.DB
}

var _ dao.OperationRelation = (*OperationRelation)(nil)

func (o OperationRelation) FindOperationRelation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) (entity.OperationRelation, *errs.Error) {
	operationRelation := entity.OperationRelation{}
	err := o.db.QueryRow(`
		SELECT
			child_resource_type,
			child_operation,
			parent_resource_type,
			parent_operation,
			created_at,
			creator_user_id
		FROM operation_relation
		WHERE child_resource_type = $1 
		  AND child_operation = $2 
		  AND parent_resource_type = $3 
		  AND parent_operation = $4;`,
		childResourceType, childOperation, parentResourceType, parentOperation).
		Scan(
			&operationRelation.ChildResourceType,
			&operationRelation.ChildOperation,
			&operationRelation.ParentResourceType,
			&operationRelation.ParentOperation,
			&operationRelation.CreatedAt,
			&operationRelation.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.OperationRelation{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("resource relation not found: child_resource_type=%v, child_operation=%v, parent_resource_type=%v, parent_operation=%v", childResourceType, childOperation, parentResourceType, parentOperation))
	}

	if err != nil {
		return entity.OperationRelation{}, errs.NewError(errs.Unknown, err.Error())
	}

	return operationRelation, nil
}

func (o OperationRelation) FindOperationRelations(ct context.Context, childResourceType string, childOperation string) ([]entity.OperationRelation, *errs.Error) {
	rows, err := o.db.Query(`
		SELECT
			child_resource_type,
			child_operation,
			parent_resource_type,
			parent_operation,
			created_at,
			creator_user_id
		FROM operation_relation
		WHERE child_resource_type = $1 AND child_operation = $2;`,
		childResourceType, childOperation)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	operationRelations := make([]entity.OperationRelation, 0)
	for rows.Next() {
		operationRelation := entity.OperationRelation{}
		err = rows.Scan(
			&operationRelation.ChildResourceType,
			&operationRelation.ChildOperation,
			&operationRelation.ParentResourceType,
			&operationRelation.ParentOperation,
			&operationRelation.CreatedAt,
			&operationRelation.CreatorUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		operationRelations = append(operationRelations, operationRelation)
	}

	return operationRelations, nil
}

func (o OperationRelation) FindAllOperationRelations(ct context.Context) ([]entity.OperationRelation, *errs.Error) {
	rows, err := o.db.Query(`
		SELECT
			child_resource_type,
			child_operation,
			parent_resource_type,
			parent_operation,
			created_at,
			creator_user_id
		FROM operation_relation;
	`)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	operationRelations := make([]entity.OperationRelation, 0)
	for rows.Next() {
		operationRelation := entity.OperationRelation{}
		err = rows.Scan(
			&operationRelation.ChildResourceType,
			&operationRelation.ChildOperation,
			&operationRelation.ParentResourceType,
			&operationRelation.ParentOperation,
			&operationRelation.CreatedAt,
			&operationRelation.CreatorUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		operationRelations = append(operationRelations, operationRelation)
	}

	return operationRelations, nil
}

func (o OperationRelation) CreateOperationRelation(ct context.Context, operationRelation entity.OperationRelation) *errs.Error {
	_, err := o.db.Exec(`
		INSERT INTO operation_relation
		(
		 	child_resource_type,
		 	child_operation,
		 	parent_resource_type,
		 	parent_operation,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		operationRelation.ChildResourceType,
		operationRelation.ChildOperation,
		operationRelation.ParentResourceType,
		operationRelation.ParentOperation,
		operationRelation.CreatedAt,
		operationRelation.CreatorUserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (o OperationRelation) DeleteOperationRelation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) *errs.Error {
	_, err := o.db.Exec(`
		DELETE FROM operation_relation
		WHERE child_resource_type = $1 AND child_operation = $2 AND parent_resource_type = $3 AND parent_operation = $4;
		`,
		childResourceType, childOperation, parentResourceType, parentOperation)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewOperationRelation(sqlDB *sql.DB) OperationRelation {
	return OperationRelation{db: sqlDB}
}
