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

type SignInSession struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.SignInSession = (*SignInSession)(nil)

func (s SignInSession) FindSignInSessionByID(ct context.Context, sessionID uint64) (entity.SignInSession, error) {
	row := s.db.QueryRow(`
	SELECT 
	    id,
	    redirect_url,
	    type,
	    internal_user_id
	FROM identity_sign_in_session
	WHERE id = $1;
	`,
		sessionID)

	var signInSession entity.SignInSession
	err := row.Scan(
		&signInSession.ID,
		&signInSession.RedirectURL,
		&signInSession.Type,
		&signInSession.InternalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.SignInSession{}, dao.ErrNotFound(fmt.Sprintf(
			"sign in session not found: sessionID=%v",
			sessionID))
	}

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return signInSession, err
}

func (s SignInSession) CreateSignInSession(ct context.Context, session entity.SignInSession) error {
	_, err := s.db.Exec(`
	INSERT INTO identity_sign_in_session 
	(
	 	id,
	 	redirect_url,
	 	type,
	 	internal_user_id
	)
	VALUES ($1, $2, $3, $4);
	`,
		session.ID,
		session.RedirectURL,
		session.Type,
		session.InternalUserID)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (s SignInSession) UpdateSignInSession(ct context.Context, session entity.SignInSession) error {
	_, err := s.db.Exec(`
	UPDATE identity_sign_in_session
	SET 
	    redirect_url = $1,
	    type = $2,
	    internal_user_id = $3
	WHERE id = $4;
	`,
		session.RedirectURL,
		session.Type,
		session.InternalUserID,
		session.ID)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (s SignInSession) DeleteSignInSession(ct context.Context, sessionID uint64) error {
	_, err := s.db.Exec(`
	DELETE 
	FROM identity_sign_in_session
	WHERE id = $1;`,
		sessionID)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewSignInSession(dataCollector obs.DataCollector, sqlDB *sql.DB) SignInSession {
	return SignInSession{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
