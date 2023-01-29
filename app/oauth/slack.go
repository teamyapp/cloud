package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/telemetry"
)

const SlackName = "slack"

// https://api.slack.com/authentication/sign-in-with-slack
var slackAuthURLString = "https://slack.com/openid/connect/authorize"
var slackTokenURLString = "https://slack.com/api/openid.connect.token"

type Slack struct {
	dataCollector telemetry.DataCollector
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
	idToken, err := s.getIDToken(ct, authorizationCode)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	query := baseURL.Query()
	query.Add("client_id", s.clientID)
	query.Add("redirect_uri", s.redirectURI)
	query.Add("response_type", "code")
	query.Add("state", strconv.Itoa(int(stateID)))
	query.Add("scope", "openid email")
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func (s Slack) getIDToken(ct context.Context, authorizationCode string) (string, error) {
	// API accepts request in encoded form data only
	// https://api.slack.com/methods/openid.connect.token
	formData := url.Values{}
	formData.Set("client_id", s.clientID)
	formData.Set("client_secret", s.clientSecret)
	formData.Set("code", authorizationCode)
	formData.Set("grant_type", "authorization_code")
	formData.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequest("POST", slackTokenURLString, strings.NewReader(formData.Encode()))
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode > 300 || res.StatusCode < 200 {
		err = fmt.Errorf("fail to obtain %s access token", s.GetName())
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: err,
			"OauthProviderName": s.GetName(),
			"HttpStatusCode":    res.StatusCode,
		})
		return "", err
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}
	if !body.Ok {
		err = errors.New(body.Error)
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err.Error()})
	}

	return body.IDToken, err
}

func NewSlack(
	dataCollector telemetry.DataCollector,
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
