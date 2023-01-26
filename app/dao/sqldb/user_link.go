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

type UserLink struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.UserLink = (*UserLink)(nil)

func (u UserLink) FindUserLinkByExternalUserID(ct context.Context, authProvider string, externalUserID string) (entity.UserLink, error) {
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
		return entity.UserLink{}, dao.ErrNotFound(fmt.Sprintf(
			"user link not found: authProvider=%v externalUserID=%v",
			authProvider,
			externalUserID))
	}

	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return userLink, err
}

func (u UserLink) FindUserLinksByInternalUserID(ct context.Context, internalUserID uint64) ([]entity.UserLink, error) {
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
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		userLinks = append(userLinks, userLink)
	}

	return userLinks, nil
}

func (u UserLink) CreateUserLink(ct context.Context, userLink entity.UserLink) error {
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
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (u UserLink) DeleteUserLink(ct context.Context, authProvider string, internalUserID uint64) error {
	_, err := u.db.Exec(`
		DELETE 
		FROM identity_user_link
		WHERE auth_provider = $1 AND internal_user_id = $2;`,
		authProvider,
		internalUserID)

	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewUserLink(dataCollector obs.DataCollector, sqlDB *sql.DB) UserLink {
	return UserLink{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
