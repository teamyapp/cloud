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

type UserGroup struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.UserGroup = (*UserGroup)(nil)

func (u UserGroup) FindGroupByID(ct context.Context, groupID uint64) (entity.UserGroup, error) {
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
		return entity.UserGroup{}, dao.ErrNotFound(fmt.Sprintf(
			"user group not found: group_id=%d",
			groupID))
	}

	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return group, err
}

func (u UserGroup) FindAllGroups(ct context.Context) ([]entity.UserGroup, error) {
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
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
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
			u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (u UserGroup) CreateGroup(ct context.Context, group entity.UserGroup) error {
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
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (u UserGroup) UpdateGroup(ct context.Context, group entity.UserGroup) error {
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
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (u UserGroup) DeleteGroup(ct context.Context, groupID uint64) error {
	_, err := u.db.Exec(`
		DELETE FROM user_group
		WHERE id = $1;
		`,
		groupID)

	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewUserGroup(dataCollector obs.DataCollector, sqlDB *sql.DB) UserGroup {
	return UserGroup{dataCollector: dataCollector, db: sqlDB}
}
