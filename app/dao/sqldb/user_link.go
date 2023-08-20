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

type UserLink struct {
	db *sql.DB
}

var _ dao.UserLink = (*UserLink)(nil)

func (u UserLink) FindUserLinkByExternalUserID(ct context.Context, authProvider string, externalUserID string) (entity.UserLink, *errs.Error) {
	row := u.db.QueryRow(`
		SELECT
		    auth_provider,
		    external_user_id,
		    external_user_label,
		    internal_user_id
		FROM identity_user_link
		WHERE auth_provider = $1 AND external_user_id = $2;
`,
		authProvider,
		externalUserID)

	var userLink entity.UserLink
	err := row.Scan(
		&userLink.AuthProvider,
		&userLink.ExternalUserID,
		&userLink.ExternalUserLabel,
		&userLink.InternalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserLink{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf(
				"user link not found: authProvider=%v externalUserID=%v",
				authProvider,
				externalUserID))
	}

	if err != nil {
		return entity.UserLink{}, errs.NewError(errs.Unknown, err.Error())
	}

	return userLink, nil
}

func (u UserLink) FindUserLinksByInternalUserID(ct context.Context, internalUserID uint64) ([]entity.UserLink, *errs.Error) {
	rows, err := u.db.Query(
		`
		SELECT
		    auth_provider,
		    external_user_id,
		    external_user_label,
		    internal_user_id
		FROM identity_user_link
		WHERE internal_user_id = $1;
`,
		internalUserID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	userLinks := make([]entity.UserLink, 0)
	for rows.Next() {
		userLink := entity.UserLink{}
		err = rows.Scan(
			&userLink.AuthProvider,
			&userLink.ExternalUserID,
			&userLink.ExternalUserLabel,
			&userLink.InternalUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		userLinks = append(userLinks, userLink)
	}

	return userLinks, nil
}

func (u UserLink) CreateUserLink(ct context.Context, userLink entity.UserLink) *errs.Error {
	_, err := u.db.Exec(`
		INSERT INTO identity_user_link 
		(
		 	auth_provider,
		 	external_user_id,
		 	external_user_label,
		 	internal_user_id
		)
		VALUES ($1, $2, $3, $4);
		`,
		userLink.AuthProvider,
		userLink.ExternalUserID,
		userLink.ExternalUserLabel,
		userLink.InternalUserID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (u UserLink) DeleteUserLink(ct context.Context, authProvider string, internalUserID uint64) *errs.Error {
	_, err := u.db.Exec(`
		DELETE 
		FROM identity_user_link
		WHERE auth_provider = $1 AND internal_user_id = $2;`,
		authProvider,
		internalUserID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewUserLink(sqlDB *sql.DB) UserLink {
	return UserLink{
		db: sqlDB,
	}
}
