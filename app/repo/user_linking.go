package repo

import (
	"github.com/teamyapp/cloud/app/errs"
	"github.com/teamyapp/one/entity"
)

type UserLinking interface {
	GetInternalUser(oauthProvider string, externalUserID string) (entity.ID, error)
	LinkUser(oauthProvider string, externalUserID string, internalUserID entity.ID) error
}

type userLinkingRow struct {
	oauthProvider  string
	externalUserID string
	internalUserID entity.ID
}

type InMemoryUserLinking struct {
	rows []userLinkingRow
}

var _ UserLinking = (*InMemoryUserLinking)(nil)

func (i InMemoryUserLinking) GetInternalUser(oauthProvider string, externalUserID string) (entity.ID, error) {
	for _, row := range i.rows {
		if row.oauthProvider == oauthProvider &&
			row.externalUserID == externalUserID {
			return row.internalUserID, nil
		}
	}

	return -1, errs.NotFound{Message: "external user not found"}
}

func (i *InMemoryUserLinking) LinkUser(oauthProvider string, externalUserID string, internalUserID entity.ID) error {
	for index, row := range i.rows {
		if row.oauthProvider == oauthProvider &&
			row.externalUserID == externalUserID {
			row.internalUserID = internalUserID
			i.rows[index] = row
			return nil
		}
	}

	i.rows = append(i.rows, userLinkingRow{
		oauthProvider:  oauthProvider,
		externalUserID: externalUserID,
		internalUserID: internalUserID,
	})
	return nil
}

func NewInMemoryUserLinking() *InMemoryUserLinking {
	return &InMemoryUserLinking{rows: make([]userLinkingRow, 0)}
}
