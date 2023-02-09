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

type UserLink struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
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
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"user link not found: authProvider=%v externalUserID=%v",
				authProvider,
				externalUserID),
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserLink{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.UserLink{}, internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewUserLink(dataCollector telemetry.DataCollector, sqlDB *sql.DB) UserLink {
	return UserLink{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
