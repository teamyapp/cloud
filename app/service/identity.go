package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type tokenPayload struct {
	UserID           uint64     `json:"user_id"`
	IssuedAt         *time.Time `json:"issued_at"`
	IsServiceAccount bool       `json:"is_service_account"`
	Secret           *string    `json:"secret"`
}

type Identity struct {
	logger            telemetry.Logger
	signInSessionDao  dao.SignInSession
	userLinkDao       dao.UserLink
	serviceAccountDao dao.ServiceAccount
	userIDGenerator   *gen.UniqueNumber
	stateIDGenerator  *gen.UniqueNumber
	jwtAuthority      security.JWTAuthority
	oauthProviders    map[string]oauth.Provider
	accessTokenTLL    time.Duration
}

func (i Identity) VerifyAccessToken(ct context.Context, accessToken string) (uint64, bool) {
	payload := tokenPayload{}
	err := i.jwtAuthority.DecodeToken(ct, accessToken, &payload)
	if err != nil {
		return 0, false
	}

	if payload.IsServiceAccount {
		serviceAccount, err := i.serviceAccountDao.FindServiceAccountByID(ct, payload.UserID)
		if err != nil {
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

func (i Identity) GenerateUnknownUserSignInURL(ct context.Context, authProviderName string, redirectURL string) (string, *errs.Error) {
	session := entity.SignInSession{
		Type:        entity.UnknownUserSignInSessionType,
		RedirectURL: redirectURL,
	}

	return i.generateSignInURL(ct, authProviderName, session)
}

func (i Identity) GenerateLinkUsersSignInURL(
	ct context.Context,
	authProviderName string,
	internalUserID uint64,
	redirectURL string,
) (string, *errs.Error) {
	session := entity.SignInSession{
		Type:           entity.LinkUsersSignInSessionType,
		InternalUserID: &internalUserID,
		RedirectURL:    redirectURL,
	}

	return i.generateSignInURL(ct, authProviderName, session)
}

func (i Identity) generateSignInURL(ct context.Context, authProviderName string, session entity.SignInSession) (string, *errs.Error) {
	provider, err := i.GetOAuthProvider(ct, authProviderName)
	if err != nil {
		return "", err
	}

	sessionID, err := i.stateIDGenerator.GenerateUniqueNumber(ct)
	if err != nil {
		return "", err
	}

	session.ID = sessionID
	err = i.signInSessionDao.CreateSignInSession(ct, session)
	if err != nil {
		return "", err
	}

	signInURL, err := provider.GetSignInURL(ct, sessionID)
	if err != nil {
		return "", err
	}

	i.logger.InfoWithContext(ct, fmt.Sprintf("SignInURL=%v", signInURL))
	return signInURL, nil
}

func (i Identity) GetOAuthProvider(ct context.Context, authProviderName string) (oauth.Provider, *errs.Error) {
	provider, ok := i.oauthProviders[authProviderName]
	if !ok {
		return nil, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("authProvider not found: AuthProvider=%v", provider))
	}

	return provider, nil
}

func (i Identity) FinishOAuthSignIn(ct context.Context, authProviderName string, authorizationCode string, sessionID uint64) (string, *errs.Error) {
	session, internalErr := i.signInSessionDao.FindSignInSessionByID(ct, sessionID)
	if internalErr != nil {
		return "", internalErr
	}

	internalErr = i.signInSessionDao.DeleteSignInSession(ct, sessionID)
	if internalErr != nil {
		return "", internalErr
	}

	provider, internalErr := i.GetOAuthProvider(ct, authProviderName)
	if internalErr != nil {
		return "", internalErr
	}

	externalUser, internalErr := provider.GetUser(ct, authorizationCode)
	if internalErr != nil {
		return "", internalErr
	}

	u, err := url.Parse(session.RedirectURL)
	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	switch session.Type {
	case entity.UnknownUserSignInSessionType:
		return i.signInUnknownUser(ct, authProviderName, externalUser, u)
	case entity.LinkUsersSignInSessionType:
		if session.InternalUserID == nil {
			return "", errs.NewError(errs.InvalidValue, "internal user ID cannot nil")
		}

		internalErr = i.linkUsers(ct, authProviderName, externalUser, *session.InternalUserID)
		if internalErr != nil {
			return "", internalErr
		}

		return u.String(), nil
	default:
		return "", errs.NewError(
			errs.InvalidValue,
			fmt.Sprintf("unsupported sign in session type: sessionType=%v", session.Type))
	}
}

func (i Identity) getOrLinkInternalUserID(ct context.Context, authProvider string, externalUser entity.ExternalUser) (uint64, *errs.Error) {
	internalUserID, err := i.GetInternalUserID(ct, authProvider, externalUser.ID)
	if err == nil {
		return internalUserID, nil
	}

	if err.Code != errs.NotFound {
		return 0, err
	}

	internalUserID, err = i.userIDGenerator.GenerateUniqueNumber(ct)
	if err != nil {
		return 0, err
	}

	userLink := entity.UserLink{
		AuthProvider:      authProvider,
		InternalUserID:    internalUserID,
		ExternalUserID:    externalUser.ID,
		ExternalUserLabel: externalUser.Label,
	}

	err = i.userLinkDao.CreateUserLink(ct, userLink)
	if err != nil {
		return 0, err
	}

	return internalUserID, err
}

func (i Identity) GetInternalUserID(ct context.Context, authProvider string, externalUserID string) (uint64, *errs.Error) {
	userLink, err := i.userLinkDao.FindUserLinkByExternalUserID(ct, authProvider, externalUserID)
	if err != nil {
		return 0, err
	}

	return userLink.InternalUserID, nil
}

func (i Identity) ListServiceAccounts(ct context.Context, accountOwnerID uint64) ([]entity.ServiceAccount, *errs.Error) {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(ct, accountOwnerID)
	if err != nil {
		return nil, err
	}

	return collect.Map(serviceAccounts, func(serviceAccount entity.ServiceAccount, _ int) entity.ServiceAccount {
		serviceAccount.Secret = nil
		return serviceAccount
	}), nil
}

func (i Identity) CreateServiceAccount(ct context.Context, accountOwnerID uint64, serviceAccountName string) *errs.Error {
	serviceAccountID, err := i.userIDGenerator.GenerateUniqueNumber(ct)
	if err != nil {
		return err
	}

	account := entity.ServiceAccount{
		ID:          serviceAccountID,
		Name:        serviceAccountName,
		OwnerUserID: accountOwnerID,
		CreatedAt:   time.Now().UTC(),
	}

	return i.serviceAccountDao.CreateServiceAccount(ct, account)
}

func (i Identity) GenerateServiceToken(ct context.Context, accountOwnerID uint64, serviceAccountID uint64) (string, *errs.Error) {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(ct, accountOwnerID)
	if err != nil {
		return "", err
	}

	foundServiceAccounts := collect.Filter(serviceAccounts, func(account entity.ServiceAccount) bool {
		return account.ID == serviceAccountID
	})
	if len(foundServiceAccounts) < 1 {
		return "", errs.NewError(
			errs.NotFound,
			fmt.Sprintf("service account not found: userID=%v, serviceAccountID=%v",
				accountOwnerID,
				serviceAccountID))
	}

	serviceAccount := serviceAccounts[0]
	secret := randgen.String(randgen.Base62, 10)
	serviceAccount.Secret = &secret
	err = i.serviceAccountDao.UpdateServiceAccount(ct, serviceAccount)
	if err != nil {
		return "", err
	}

	payload := tokenPayload{
		UserID:           serviceAccountID,
		IsServiceAccount: true,
		Secret:           serviceAccount.Secret,
	}

	return i.jwtAuthority.GenerateToken(ct, payload)
}

func (i Identity) DeleteServiceAccount(ct context.Context, accountOwnerID uint64, serviceAccountID uint64) *errs.Error {
	serviceAccounts, err := i.serviceAccountDao.FindAllServiceAccounts(ct, accountOwnerID)
	if err != nil {
		return err
	}

	foundServiceAccounts := collect.Filter(serviceAccounts, func(account entity.ServiceAccount) bool {
		return account.ID == serviceAccountID
	})
	if len(foundServiceAccounts) < 1 {
		return errs.NewError(
			errs.NotFound,
			fmt.Sprintf("service account not found: userID=%v, serviceAccountID=%v",
				accountOwnerID,
				serviceAccountID))
	}

	return i.serviceAccountDao.DeleteServiceAccount(ct, serviceAccountID)
}

func (i Identity) ListUserLinks(ct context.Context, internalUserID uint64) ([]entity.UserLink, *errs.Error) {
	return i.userLinkDao.FindUserLinksByInternalUserID(ct, internalUserID)
}

func (i Identity) DeleteUserLink(ct context.Context, userID uint64, authProviderName string) *errs.Error {
	return i.userLinkDao.DeleteUserLink(ct, authProviderName, userID)
}

func (i Identity) signInUnknownUser(
	ct context.Context,
	authProviderName string,
	externalUser entity.ExternalUser,
	redirectURL *url.URL,
) (string, *errs.Error) {
	userID, err := i.getOrLinkInternalUserID(ct, authProviderName, externalUser)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	payload := tokenPayload{
		UserID:   userID,
		IssuedAt: &now,
	}

	accessToken, internalErr := i.jwtAuthority.GenerateToken(ct, payload)
	if internalErr != nil {
		return "", internalErr
	}

	query := redirectURL.Query()
	query.Add("accessToken", accessToken)
	redirectURL.RawQuery = query.Encode()
	return redirectURL.String(), nil
}

func (i Identity) linkUsers(
	ct context.Context,
	authProviderName string,
	externalUser entity.ExternalUser,
	internalUserID uint64,
) *errs.Error {
	userLink := entity.UserLink{
		AuthProvider:      authProviderName,
		InternalUserID:    internalUserID,
		ExternalUserID:    externalUser.ID,
		ExternalUserLabel: externalUser.Label,
	}

	return i.userLinkDao.CreateUserLink(ct, userLink)
}

func NewIdentity(
	logger telemetry.Logger,
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
		return Identity{}, err.ToError()
	}

	stateIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("stateID")
	if err != nil {
		return Identity{}, err.ToError()
	}

	oauthProviderMap := make(map[string]oauth.Provider)
	for _, oauthProvider := range oauthProviders {
		oauthProviderMap[oauthProvider.GetName()] = oauthProvider
	}

	return Identity{
		logger:            logger,
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
