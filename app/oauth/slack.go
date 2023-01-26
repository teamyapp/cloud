package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/security"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const SlackName = "slack"

// https://api.slack.com/authentication/sign-in-with-slack
var slackAuthURLString = "https://slack.com/openid/connect/authorize"
var slackTokenURLString = "https://slack.com/api/openid.connect.token"

//var slackAuthURLString = "https://slack.com/oauth/v2/authorize"
//var slackTokenURLString = "https://slack.com/api/oauth.v2.access"

type Slack struct {
	dataCollector obs.DataCollector
	jwtAuthority  security.JWTAuthority
	clientID      string
	clientSecret  string
	redirectURI   string
}

var _ Provider = (*Slack)(nil)

func (s Slack) GetName() string {
	return SlackName
}

func (s Slack) GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, error) {
	// https://api.slack.com/methods/openid.connect.token
	idToken, err := s.getIDToken(ct, authorizationCode)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	tokenPayload := struct {
		UserID         string `json:"sub"`
		Issuer         string `json:"iss"`
		ExpirationTime int    `json:"exp"`
		IssuedAt       int    `json:"iat"`
		Email          string `json:"email"`
		EmailVerified  bool   `json:"email_verified"`
	}{}

	err = s.jwtAuthority.DecodeUnverifiedToken(ct, idToken, &tokenPayload)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return entity.ExternalUser{
		ID:    tokenPayload.UserID,
		Label: tokenPayload.Email,
	}, err
}

func (s Slack) GetStateID(request *http.Request) (uint64, error) {
	return strconv.ParseUint(request.URL.Query().Get("state"), 10, 64)
}

func (s Slack) GetAuthorizationCode(request *http.Request) string {
	return request.URL.Query().Get("code")
}

func (s Slack) GetSignInURL(ct context.Context, stateID uint64) (string, error) {
	baseURL, err := url.Parse(slackAuthURLString)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	query := baseURL.Query()
	query.Add("client_id", s.clientID)
	// while redirect_uri is optional, it is required if your app passed it as a parameter
	// to /openid/connect/authorize in the first step of the Sign in with Slack flow.
	query.Add("redirect_uri", s.redirectURI)
	query.Add("response_type", "code")
	query.Add("state", strconv.Itoa(int(stateID)))
	query.Add("scope", "openid email profile")
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func (s Slack) getIDToken(ct context.Context, authorizationCode string) (string, error) {
	tokenBody := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
		Code         string `json:"code"`
		GrantType    string `json:"grant_type"`
	}{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Code:         authorizationCode,
		GrantType:    "authorization_code",
		RedirectURI:  s.redirectURI,
	}

	buf, err := json.Marshal(tokenBody)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	//data := url.Values{}
	//data.Set("client_id", s.clientID)
	//data.Set("client_secret", s.clientSecret)
	//data.Add("code", authorizationCode)
	//data.Add("grant_type", "authorization_code")
	//data.Add("redirect_uri", s.redirectURI)
	//encodedData := data.Encode()

	req, err := http.NewRequest("POST", slackTokenURLString, bytes.NewReader(buf))
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	if res.StatusCode > 300 || res.StatusCode < 200 {
		err = fmt.Errorf("fail to obtain %s access token", s.GetName())
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp:       err,
			"OauthProviderName": s.GetName(),
			"HttpStatusCode":    res.StatusCode,
		})
		return "", err
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	body := struct {
		Ok          bool   `json:"ok"`
		Error       string `json:"error"`
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}
	if !body.Ok {
		err = errors.New(body.Error)
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err.Error()})
	}

	return body.IDToken, err
}

func NewSlack(
	dataCollector obs.DataCollector,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL string,
	clientID string,
	clientSecret string,
) Slack {
	return Slack{
		dataCollector: dataCollector,
		jwtAuthority:  jwtAuthority,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, SlackName),
	}
}
