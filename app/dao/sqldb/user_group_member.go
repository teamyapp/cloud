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

type UserGroupMember struct {
	db *sql.DB
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	groupIDs := make([]uint64, 0)
	for rows.Next() {
		var groupID uint64
		err = rows.Scan(
			&groupID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
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
		return entity.UserGroupMember{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf(
				"user group member not found: group_id=%d, user_id=%d",
				groupID,
				userID))
	}

	if err != nil {
		return entity.UserGroupMember{}, errs.NewError(errs.Unknown, err.Error())
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewUserGroupMember(sqlDB *sql.DB) UserGroupMember {
	return UserGroupMember{db: sqlDB}
}
