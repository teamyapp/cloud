package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/teamyapp/cloud/app/channel"
	"github.com/teamyapp/cloud/app/errs"
	"github.com/teamyapp/cloud/app/idgen"
	"github.com/teamyapp/cloud/app/oauth"
	"github.com/teamyapp/cloud/app/pubsub"
	"github.com/teamyapp/cloud/app/repo"
	"github.com/teamyapp/cloud/app/security"
	oneEntity "github.com/teamyapp/one/entity"
)

type TokenPayload struct {
	UserID   oneEntity.ID `json:"user_id"`
	IssuedAt string       `json:"issued_at"`
}

type Identity struct {
	userIDGenerator    *idgen.IDGenerator
	sessionIDGenerator *idgen.IDGenerator
	jwtAuthority       security.JWTAuthority
	pubSub             pubsub.PubSub
	userLinkingRepo    repo.UserLinking
	oauthProviders     map[string]oauth.OAuth
	accessTokenTLL     time.Duration
	signInTimeOut      time.Duration
}

func (i Identity) RequestOAuthSignInURL(oauthProvider string, sessionID oneEntity.ID) (string, error) {
	oauthHandler, ok := i.oauthProviders[oauthProvider]
	if !ok {
		return "", errors.New("invalid oauthProvider")
	}

	signInURL, err := oauthHandler.GetSignInURL(sessionID)
	if err == nil {
		log.Printf("sign in URL: %s", signInURL)
	}
	return signInURL, err
}

func (i Identity) FinishOAuthSignIn(authorizationCode string, sessionID oneEntity.ID, oauthProvider string) error {
	provider, ok := i.oauthProviders[oauthProvider]
	if !ok {
		return fmt.Errorf("unknown oauth provider: %s", provider)
	}

	externalUser, err := provider.GetUser(authorizationCode)
	if err != nil {
		return err
	}

	userID, err := i.getInternalUserID(provider.GetName(), externalUser.ID)
	if err != nil {
		return err
	}

	payload := TokenPayload{
		UserID:   userID,
		IssuedAt: time.Now().Format(time.RFC3339),
	}

	jwt, err := i.jwtAuthority.GenerateToken(payload)
	if err != nil {
		return err
	}
	return i.pubSub.Publish(strconv.Itoa(int(sessionID)), jwt)
}

func (i Identity) ClientSubscribe(channel channel.Channel, sessionID oneEntity.ID) {
	go func() {
		defer channel.Disconnect()
		onSessionTokenReceived := make(chan string)

		subscription := i.pubSub.Subscribe(strconv.Itoa(int(sessionID)), func(data interface{}) {
			sessionToken := data.(string)
			onSessionTokenReceived <- sessionToken
		})
		defer subscription.Unsubscribe()
		log.Printf("client subscribed: session-id=%d", sessionID)

		select {
		case sessionToken := <-onSessionTokenReceived:
			err := channel.SendMessage(sessionToken)
			log.Printf("sent session token: session-id=%d", sessionID)
			if err != nil {
				log.Println(err)
			}
		case <-time.After(i.signInTimeOut):
		}
	}()
}

func (i Identity) VerifyAccessToken(accessToken string) (oneEntity.ID, bool) {
	payload := TokenPayload{}
	err := i.jwtAuthority.DecodeToken(accessToken, &payload)
	if err != nil {
		return -1, false
	}

	tm, err := time.Parse(time.RFC3339, payload.IssuedAt)
	if err != nil {
		return -1, false
	}

	if tm.Add(i.accessTokenTLL).Before(time.Now()) {
		return -1, false
	}
	return payload.UserID, true
}

func (i Identity) NewSessionID() (oneEntity.ID, error) {
	return i.sessionIDGenerator.GenerateUniqueID()
}

func (i Identity) GetOAuthProvider(providerName string) (oauth.OAuth, error) {
	provider, ok := i.oauthProviders[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", provider)
	}
	return provider, nil
}

func (i Identity) getInternalUserID(oauthProvider string, externalUserID string) (oneEntity.ID, error) {
	internalUserID, err := i.userLinkingRepo.GetInternalUser(oauthProvider, externalUserID)
	switch err.(type) {
	case nil:
		return internalUserID, nil
	case errs.NotFound:
		internalUserID, err = i.userIDGenerator.GenerateUniqueID()
		if err != nil {
			return -1, err
		}
		err = i.userLinkingRepo.LinkUser(oauthProvider, externalUserID, internalUserID)
		return internalUserID, err
	default:
		return -1, err
	}
}

func NewIdentity(
	jwtAuthority security.JWTAuthority,
	pubSub pubsub.PubSub,
	idGeneratorFactory idgen.Factory,
	userLinkingRepo repo.UserLinking,
	oauthProviders []oauth.OAuth,
	accessTokenTLL time.Duration,
	signInTimeOut time.Duration,
) (Identity, error) {
	providerMap := make(map[string]oauth.OAuth)
	for _, provider := range oauthProviders {
		providerMap[provider.GetName()] = provider
	}

	userIDGen, err := idGeneratorFactory.NewIDGenerator("userID")
	if err != nil {
		log.Println(err)
		return Identity{}, err
	}

	sessionIDGen, err := idGeneratorFactory.NewIDGenerator("sessionID")
	if err != nil {
		log.Println(err)
		return Identity{}, err
	}

	return Identity{
		jwtAuthority:       jwtAuthority,
		userIDGenerator:    userIDGen,
		sessionIDGenerator: sessionIDGen,
		oauthProviders:     providerMap,
		pubSub:             pubSub,
		userLinkingRepo:    userLinkingRepo,
		accessTokenTLL:     accessTokenTLL,
		signInTimeOut:      signInTimeOut,
	}, nil
}
