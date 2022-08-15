package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"log"
)

type UserGroupMember struct {
	db *sql.DB
}

var _ dao.UserGroupMember = (*UserGroupMember)(nil)

func (u UserGroupMember) FindGroupIDsByUserID(userID uint64) ([]uint64, error) {
	rows, err := u.db.Query(`
		SELECT
			group_id
		FROM user_group_member
		WHERE user_id = $1;`,
		userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	groupIDs := make([]uint64, 0)
	for rows.Next() {
		var groupID uint64
		err = rows.Scan(
			&groupID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs, err
}

func (u UserGroupMember) FindUserGroupMembersByUserID(userID uint64) ([]entity.UserGroupMember, error) {
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
		log.Println(err)
		return nil, err
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
			log.Println(err)
			continue
		}

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, err
}

func (u UserGroupMember) FindUserGroupMembersByGroupID(groupID uint64) ([]entity.UserGroupMember, error) {
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
		log.Println(err)
		return nil, err
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
			log.Println(err)
			continue
		}

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, err
}

func (u UserGroupMember) FindUserGroupMember(groupID uint64, userID uint64) (entity.UserGroupMember, error) {
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
		return entity.UserGroupMember{}, dao.ErrNotFound(fmt.Sprintf(
			"user group member not found: group_id=%d, user_id=%d",
			groupID, userID))
	}

	return userGroupMember, err
}

func (u UserGroupMember) FindAllUserGroupMembers() ([]entity.UserGroupMember, error) {
	rows, err := u.db.Query(`
		SELECT
			group_id,
			user_id,
			created_at,
			creator_user_id
		FROM user_group_member;
`)
	if err != nil {
		log.Println(err)
		return nil, err
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
			log.Println(err)
			continue
		}

		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	return userGroupMembers, err
}

func (u UserGroupMember) CreateUserGroupMember(userGroupMember entity.UserGroupMember) error {
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
	return err
}

func (u UserGroupMember) DeleteUserGroupMember(groupID uint64, userID uint64) error {
	_, err := u.db.Exec(`
		DELETE FROM user_group_member
		WHERE group_id = $1 AND user_id = $2;
		`,
		groupID, userID)
	return err
}

func NewUserGroupMember(sqlDB *sql.DB) UserGroupMember {
	return UserGroupMember{db: sqlDB}
}
