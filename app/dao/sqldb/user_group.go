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

type UserGroup struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.UserGroup = (*UserGroup)(nil)

func (u UserGroup) FindGroupByID(ct context.Context, groupID uint64) (entity.UserGroup, *errs.Error) {
	group := entity.UserGroup{}
	err := u.db.QueryRow(`
		SELECT
			id,
			name,
			description,
			created_at,
			creator_user_id,
			updated_at
		FROM user_group
		WHERE id = $1;`,
		groupID).
		Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatedAt,
			&group.CreatorUserID,
			&group.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"user group not found: group_id=%d",
				groupID),
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserGroup{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserGroup{}, internalErr
	}

	return group, nil
}

func (u UserGroup) FindAllGroups(ct context.Context) ([]entity.UserGroup, *errs.Error) {
	rows, err := u.db.Query(`
		SELECT
			id,
			name,
			description,
			created_at,
			creator_user_id,
			updated_at
		FROM user_group;
	`)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	groups := make([]entity.UserGroup, 0)
	for rows.Next() {
		group := entity.UserGroup{}
		err = rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatedAt,
			&group.CreatorUserID,
			&group.UpdatedAt,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: newInternalErr})
			continue
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (u UserGroup) CreateGroup(ct context.Context, group entity.UserGroup) (entity.UserGroup, *errs.Error) {
	_, err := u.db.Exec(`
		INSERT INTO user_group
		(
			id,
			name,
			description,
			created_at,
			creator_user_id,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		group.ID,
		group.Name,
		group.Description,
		group.CreatedAt,
		group.CreatorUserID,
		group.UpdatedAt,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserGroup{}, internalErr
	}

	return group, nil
}

func (u UserGroup) UpdateGroup(ct context.Context, group entity.UserGroup) *errs.Error {
	_, err := u.db.Exec(`
		UPDATE user_group
		SET
			name = $1,
			description = $2,
			created_at = $3,
			creator_user_id = $4,
			updated_at = $5
		WHERE id = $6;`,
		group.Name,
		group.Description,
		group.CreatedAt,
		group.CreatorUserID,
		group.UpdatedAt,
		group.ID,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (u UserGroup) DeleteGroup(ct context.Context, groupID uint64) *errs.Error {
	_, err := u.db.Exec(`
		DELETE FROM user_group
		WHERE id = $1;
		`,
		groupID)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewUserGroup(dataCollector telemetry.DataCollector, sqlDB *sql.DB) UserGroup {
	return UserGroup{dataCollector: dataCollector, db: sqlDB}
}
