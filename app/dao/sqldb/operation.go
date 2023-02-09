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

type Operation struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.Operation = (*Operation)(nil)

func (o Operation) FindOperation(ct context.Context, resourceTypeName string, operationName string) (entity.Operation, *errs.Error) {
	operation := entity.Operation{}
	err := o.db.QueryRow(`
		SELECT
			resource_type,
			operation,
			created_at,
			creator_user_id
		FROM operation
		WHERE resource_type = $1 AND operation = $2;`,
		resourceTypeName, operationName).
		Scan(
			&operation.ResourceTypeName,
			&operation.OperationName,
			&operation.CreatedAt,
			&operation.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("resource not found: resource_type=%v, operation=%v", resourceTypeName, operationName),
		}
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Operation{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Operation{}, internalErr
	}

	return operation, nil
}

func (o Operation) FindAllOperations(ct context.Context) ([]entity.Operation, *errs.Error) {
	rows, err := o.db.Query(`
		SELECT
			resource_type,
			operation,
			created_at,
			creator_user_id
		FROM operation;
	`)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	operations := make([]entity.Operation, 0)
	for rows.Next() {
		operation := entity.Operation{}
		err = rows.Scan(
			&operation.ResourceTypeName,
			&operation.OperationName,
			&operation.CreatedAt,
			&operation.CreatorUserID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: newInternalErr})
			continue
		}

		operations = append(operations, operation)
	}

	return operations, nil
}

func (o Operation) CreateOperation(ct context.Context, operation entity.Operation) *errs.Error {
	_, err := o.db.Exec(`
		INSERT INTO operation
		(
			resource_type,
		 	operation,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4);`,
		operation.ResourceTypeName,
		operation.OperationName,
		operation.CreatedAt,
		operation.CreatorUserID,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (o Operation) DeleteOperation(ct context.Context, resourceTypeName string, operationName string) *errs.Error {
	_, err := o.db.Exec(`
		DELETE FROM operation
		WHERE resource_type = $1 AND operation = $2;
		`,
		resourceTypeName, operationName)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		o.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewOperation(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Operation {
	return Operation{dataCollector: dataCollector, db: sqlDB}
}
