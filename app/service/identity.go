package service

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/security"
)

type tokenPayload struct {
	UserID           uint64     `json:"user_id"`
	IssuedAt         *time.Time `json:"issued_at"`
	IsServiceAccount bool       `json:"is_service_account"`
	Secret           *string    `json:"secret"`
}

type Identity struct {
	dataCollector     obs.DataCollector
	signInSessionDao  dao.SignInSession
	userLinkDao       dao.UserLink
	serviceAccountDao dao.ServiceAccount
	userIDGenerator   *gen.UniqueNumber
	stateIDGenerator  *gen.UniqueNumber
	jwtAuthority      security.JWTAuthority
	oauthProviders    map[string]oauth.Provider
	accessTokenTLL    time.Duration
}

func (i Identity) VerifyAccessToken(accessToken string) (uint64, bool) {
	payload := tokenPayload{}
	err := i.jwtAuthority.DecodeToken(accessToken, &payload)
	if err != nil {
		return 0, false
	}

	if payload.IsServiceAccount {
		serviceAccount, err := i.serviceAccountDao.FindServiceAccountByID(payload.UserID)
		if err != nil {
			i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return 0, false
		}

		if serviceAccount.Secret == nil || payload.Secret == nil {
			return 0, false
		}

		return serviceAccount.ID, *serviceAccount.Secret == *payload.Secret
	} else {
		issuedAt := *payload.IssuedAt
		if issuedAt.Add(i.accessTokenTLL).Before(time.Now()) {
			return 0, false
		}
	}

	return payload.UserID, true
}

func (i Identity) GenerateUnknownUserSignInURL(authProviderName string, redirectURL string) (string, error) {
	session := entity.SignInSession{
		Type:        entity.UnknownUserSignInSessionType,
		RedirectURL: redirectURL,
	}

	return i.generateSignInURL(authProviderName, session)
}

func (i Identity) GenerateLinkUsersSignInURL(
	authProviderName string,
	internalUserID uint64,
	redirectURL string,
) (string, error) {
	session := entity.SignInSession{
		Type:           entity.LinkUsersSignInSessionType,
		InternalUserID: &internalUserID,
		RedirectURL:    redirectURL,
	}

	return i.generateSignInURL(authProviderName, session)
}

func (i Identity) generateSignInURL(authProviderName string, session entity.SignInSession) (string, error) {
	provider, err := i.GetOAuthProvider(authProviderName)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	sessionID, err := i.stateIDGenerator.GenerateUniqueNumber()
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	session.ID = sessionID
	err = i.signInSessionDao.CreateSignInSession(session)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	signInURL, err := provider.GetSignInURL(sessionID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	i.dataCollector.Logger.Log(obs.Info, obs.Props{
		obs.MessageProp: obs.Props{
			"signInURL": signInURL,
		},
	})
	return signInURL, nil
}

func (i Identity) GetOAuthProvider(authProviderName string) (oauth.Provider, error) {
	provider, ok := i.oauthProviders[authProviderName]
	if !ok {
		err := fmt.Errorf("authProvider not found")
		i.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"authProvider": provider,
			},
		})
		return nil, err
	}

	return provider, nil
}

func (i Identity) FinishOAuthSignIn(authProviderName string, authorizationCode string, sessionID uint64) (string, error) {
	session, err := i.signInSessionDao.FindSignInSessionByID(sessionID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	err = i.signInSessionDao.DeleteSignInSession(sessionID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	provider, err := i.GetOAuthProvider(authProviderName)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	externalUser, err := provider.GetUser(authorizationCode)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	u, err := url.Parse(session.RedirectURL)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	switch session.Type {
	case entity.UnknownUserSignInSessionType:
		return i.signInUnknownUser(authProviderName, externalUser, u)
	case entity.LinkUsersSignInSessionType:
		if session.InternalUserID == nil {
			return "", errors.New("internal user ID cannot nil")
		}

		err = i.linkUsers(authProviderName, externalUser, *session.InternalUserID)
		if err != nil {
			i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return "", err
		}

		return u.String(), nil
	default:
		err = errors.New("unsupported sign in session type")
		i.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			"sessionType": session.Type,
		})
		return "", err
	}
}

func (i Identity) getOrLinkInternalUserID(authProvider string, externalUser entity.ExternalUser) (uint64, error) {
	internalUserID, err := i.GetInternalUserID(authProvider, externalUser.ID)
	switch err.(type) {
	case nil:
		return internalUserID, nil
	case dao.ErrNotFound:
		internalUserID, err = i.userIDGenerator.GenerateUniqueNumber()
		if err != nil {
			i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return 0, err
		}

		userLink := entity.UserLink{
			AuthProvider:      authProvider,
			InternalUserID:    internalUserID,
			ExternalUserID:    externalUser.ID,
			ExternalUserLabel: externalUser.Label,
		}

		err = i.userLinkDao.CreateUserLink(userLink)
		if err != nil {
			i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		}

		return internalUserID, err
	default:
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}
}

func (i Identity) GetInternalUserID(authProvider string, externalUserID string) (uint64, error) {
	userLink, err := i.userLinkDao.FindUserLinkByExternalUserID(authProvider, externalUserID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	return userLink.InternalUserID, nil
}

func (i Identity) ListServiceAccounts(accountOwnerID uint64) ([]entity.ServiceAccount, error) {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(accountOwnerID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return collect.Map(serviceAccounts, func(serviceAccount entity.ServiceAccount, _ int) entity.ServiceAccount {
		serviceAccount.Secret = nil
		return serviceAccount
	}), nil
}

func (i Identity) CreateServiceAccount(accountOwnerID uint64, serviceAccountName string) error {
	serviceAccountID, err := i.userIDGenerator.GenerateUniqueNumber()
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	account := entity.ServiceAccount{
		ID:          serviceAccountID,
		Name:        serviceAccountName,
		OwnerUserID: accountOwnerID,
		CreatedAt:   time.Now().UTC(),
	}

	return i.serviceAccountDao.CreateServiceAccount(account)
}

func (i Identity) GenerateServiceToken(accountOwnerID uint64, serviceAccountID uint64) (string, error) {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(accountOwnerID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	foundServiceAccounts := collect.Filter(serviceAccounts, func(account entity.ServiceAccount) bool {
		return account.ID == serviceAccountID
	})
	if len(foundServiceAccounts) < 1 {
		err = errors.New("service account not found")
		i.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp:      err,
			"userID":           accountOwnerID,
			"serviceAccountID": serviceAccountID,
		})
		return "", err
	}

	serviceAccount := serviceAccounts[0]
	secret := randgen.String(randgen.Base62, 10)
	serviceAccount.Secret = &secret
	err = i.serviceAccountDao.UpdateServiceAccount(serviceAccount)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	payload := tokenPayload{
		UserID:           serviceAccountID,
		IsServiceAccount: true,
		Secret:           serviceAccount.Secret,
	}

	return i.jwtAuthority.GenerateToken(payload)
}

func (i Identity) DeleteServiceAccount(accountOwnerID uint64, serviceAccountID uint64) error {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(accountOwnerID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	foundServiceAccounts := collect.Filter(serviceAccounts, func(account entity.ServiceAccount) bool {
		return account.ID == serviceAccountID
	})
	if len(foundServiceAccounts) < 1 {
		err = errors.New("service account not found")
		i.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp:      err,
			"userID":           accountOwnerID,
			"serviceAccountID": serviceAccountID,
		})
		return err
	}

	return i.serviceAccountDao.DeleteServiceAccount(serviceAccountID)
}

func (i Identity) ListUserLinks(internalUserID uint64) ([]entity.UserLink, error) {
	return i.userLinkDao.FindUserLinksByInternalUserID(internalUserID)
}

func (i Identity) DeleteUserLink(userID uint64, authProviderName string) error {
	return i.userLinkDao.DeleteUserLink(authProviderName, userID)
}

func (i Identity) signInUnknownUser(
	authProviderName string,
	externalUser entity.ExternalUser,
	redirectURL *url.URL,
) (string, error) {
	userID, err := i.getOrLinkInternalUserID(authProviderName, externalUser)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	now := time.Now()
	payload := tokenPayload{
		UserID:   userID,
		IssuedAt: &now,
	}

	accessToken, err := i.jwtAuthority.GenerateToken(payload)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	query := redirectURL.Query()
	query.Add("accessToken", accessToken)
	redirectURL.RawQuery = query.Encode()
	return redirectURL.String(), nil
}

func (i Identity) linkUsers(
	authProviderName string,
	externalUser entity.ExternalUser,
	internalUserID uint64,
) error {
	userLink := entity.UserLink{
		AuthProvider:      authProviderName,
		InternalUserID:    internalUserID,
		ExternalUserID:    externalUser.ID,
		ExternalUserLabel: externalUser.Label,
	}

	return i.userLinkDao.CreateUserLink(userLink)
}

func NewIdentity(
	dataCollector obs.DataCollector,
	signInSessionDao dao.SignInSession,
	userLinkDao dao.UserLink,
	serviceAccountDao dao.ServiceAccount,
	uniqueNumberFactory gen.UniqueNumberFactory,
	jwtAuthority security.JWTAuthority,
	oauthProviders []oauth.Provider,
	accessTokenTLL time.Duration,
) (Identity, error) {
	userIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("userID")
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Identity{}, err
	}

	stateIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("stateID")
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Identity{}, err
	}

	oauthProviderMap := make(map[string]oauth.Provider)
	for _, oauthProvider := range oauthProviders {
		oauthProviderMap[oauthProvider.GetName()] = oauthProvider
	}

	return Identity{
		signInSessionDao:  signInSessionDao,
		userLinkDao:       userLinkDao,
		serviceAccountDao: serviceAccountDao,
		userIDGenerator:   userIDGenerator,
		stateIDGenerator:  stateIDGenerator,
		jwtAuthority:      jwtAuthority,
		oauthProviders:    oauthProviderMap,
		accessTokenTLL:    accessTokenTLL,
	}, nil
}
