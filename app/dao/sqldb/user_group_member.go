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

type UserGroupMember struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.UserGroupMember = (*UserGroupMember)(nil)

func (u UserGroupMember) FindGroupIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error) {
	rows, err := u.db.Query(`
		SELECT
			group_id
		FROM user_group_member
		WHERE user_id = $1;`,
		userID)
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
	groupIDs := make([]uint64, 0)
	for rows.Next() {
		var groupID uint64
		err = rows.Scan(
			&groupID,
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

		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs, nil
}

func (u UserGroupMember) FindUserGroupMembersByUserID(ct context.Context, userID uint64) ([]entity.UserGroupMember, *errs.Error) {
	rows, err := u.db.Query(`
		SELECT
			group_id,
			user_id,
			created_at,
			creator_user_id
		FROM user_group_member
		WHERE user_id = $1;`,
		userID)
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
	userGroupMembers := make([]entity.UserGroupMember, 0)
	for rows.Next() {
		userGroupMember := entity.UserGroupMember{}
		err = rows.Scan(
			&userGroupMember.GroupID,
			&userGroupMember.UserID,
			&userGroupMember.CreatedAt,
			&userGroupMember.CreatorUserID,
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

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, nil
}

func (u UserGroupMember) FindUserGroupMembersByGroupID(ct context.Context, groupID uint64) ([]entity.UserGroupMember, *errs.Error) {
	rows, err := u.db.Query(`
		SELECT
			group_id,
			user_id,
			created_at,
			creator_user_id
		FROM user_group_member
		WHERE group_id = $1;`,
		groupID)
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
	userGroupMembers := make([]entity.UserGroupMember, 0)
	for rows.Next() {
		userGroupMember := entity.UserGroupMember{}
		err = rows.Scan(
			&userGroupMember.GroupID,
			&userGroupMember.UserID,
			&userGroupMember.CreatedAt,
			&userGroupMember.CreatorUserID,
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

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, nil
}

func (u UserGroupMember) FindUserGroupMember(ct context.Context, groupID uint64, userID uint64) (entity.UserGroupMember, *errs.Error) {
	userGroupMember := entity.UserGroupMember{}
	err := u.db.QueryRow(`
		SELECT
			group_id,
			user_id,
			created_at,
			creator_user_id
		FROM user_group_member
		WHERE group_id = $1 AND user_id = $2;`,
		groupID, userID).
		Scan(
			&userGroupMember.GroupID,
			&userGroupMember.UserID,
			&userGroupMember.CreatorUserID,
			&userGroupMember.CreatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"user group member not found: group_id=%d, user_id=%d",
				groupID,
				userID),
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserGroupMember{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserGroupMember{}, internalErr
	}

	return userGroupMember, nil
}

func (u UserGroupMember) FindAllUserGroupMembers(ct context.Context) ([]entity.UserGroupMember, *errs.Error) {
	rows, err := u.db.Query(`
		SELECT
			group_id,
			user_id,
			created_at,
			creator_user_id
		FROM user_group_member;
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
	userGroupMembers := make([]entity.UserGroupMember, 0)
	for rows.Next() {
		userGroupMember := entity.UserGroupMember{}
		err = rows.Scan(
			&userGroupMember.GroupID,
			&userGroupMember.UserID,
			&userGroupMember.CreatedAt,
			&userGroupMember.CreatorUserID,
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

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, nil
}

func (u UserGroupMember) CreateUserGroupMember(ct context.Context, userGroupMember entity.UserGroupMember) *errs.Error {
	_, err := u.db.Exec(`
		INSERT INTO user_group_member
		(
			group_id,
		 	user_id,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4);`,
		userGroupMember.GroupID,
		userGroupMember.UserID,
		userGroupMember.CreatedAt,
		userGroupMember.CreatorUserID,
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

func (u UserGroupMember) DeleteUserGroupMember(ct context.Context, groupID uint64, userID uint64) *errs.Error {
	_, err := u.db.Exec(`
		DELETE FROM user_group_member
		WHERE group_id = $1 AND user_id = $2;
		`,
		groupID, userID)

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

func NewUserGroupMember(dataCollector telemetry.DataCollector, sqlDB *sql.DB) UserGroupMember {
	return UserGroupMember{dataCollector: dataCollector, db: sqlDB}
}
