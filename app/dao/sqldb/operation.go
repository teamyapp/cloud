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

type Operation struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.Operation = (*Operation)(nil)

func (o Operation) FindOperation(ct context.Context, resourceTypeName string, operationName string) (entity.Operation, error) {
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
		return entity.Operation{}, dao.ErrNotFound(fmt.Sprintf(
			"resource not found: resource_type=%v, operation=%v",
			resourceTypeName, operationName))
	}

	if err != nil {
		o.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return operation, err
}

func (o Operation) FindAllOperations(ct context.Context) ([]entity.Operation, error) {
	rows, err := o.db.Query(`
		SELECT
			resource_type,
			operation,
			created_at,
			creator_user_id
		FROM operation;
	`)
	if err != nil {
		o.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
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
			o.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		operations = append(operations, operation)
	}

	return operations, nil
}

func (o Operation) CreateOperation(ct context.Context, operation entity.Operation) error {
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
		o.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (o Operation) DeleteOperation(ct context.Context, resourceTypeName string, operationName string) error {
	_, err := o.db.Exec(`
		DELETE FROM operation
		WHERE resource_type = $1 AND operation = $2;
		`,
		resourceTypeName, operationName)

	if err != nil {
		o.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewOperation(dataCollector obs.DataCollector, sqlDB *sql.DB) Operation {
	return Operation{dataCollector: dataCollector, db: sqlDB}
}
