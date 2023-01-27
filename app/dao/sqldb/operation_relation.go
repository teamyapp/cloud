package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type OperationRelation struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.OperationRelation = (*OperationRelation)(nil)

func (o OperationRelation) FindOperationRelation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) (entity.OperationRelation, error) {
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
		WHERE child_resource_type = $1 AND child_operation = $2 AND parent_resource_type = $3 AND parent_operation = $4;`,
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
		return entity.OperationRelation{}, dao.ErrNotFound(fmt.Sprintf(
			"resource relation not found: child_resource_type=%v, child_operation=%v, parent_resource_type=%v, parent_operation=%v",
			childResourceType, childOperation, parentResourceType, parentOperation))
	}

	if err != nil {
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return operationRelation, err
}

func (o OperationRelation) FindOperationRelations(ct context.Context, childResourceType string, childOperation string) ([]entity.OperationRelation, error) {
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
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
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
			o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		operationRelations = append(operationRelations, operationRelation)
	}

	return operationRelations, nil
}

func (o OperationRelation) FindAllOperationRelations(ct context.Context) ([]entity.OperationRelation, error) {
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
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
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
			o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		operationRelations = append(operationRelations, operationRelation)
	}

	return operationRelations, nil
}

func (o OperationRelation) CreateOperationRelation(ct context.Context, operationRelation entity.OperationRelation) error {
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
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (o OperationRelation) DeleteOperationRelation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) error {
	_, err := o.db.Exec(`
		DELETE FROM operation_relation
		WHERE child_resource_type = $1 AND child_operation = $2 AND parent_resource_type = $3 AND parent_operation = $4;
		`,
		childResourceType, childOperation, parentResourceType, parentOperation)

	if err != nil {
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewOperationRelation(dataCollector telemetry.DataCollector, sqlDB *sql.DB) OperationRelation {
	return OperationRelation{dataCollector: dataCollector, db: sqlDB}
}
