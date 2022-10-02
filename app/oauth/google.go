package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/security"
)

const GoogleName = "google"

// https://developers.google.com/identity/protocols/oauth2/web-server#httprest_1
var googleAuthURLString = "https://accounts.google.com/o/oauth2/v2/auth"
var googleTokenURLString = "https://oauth2.googleapis.com/token"

type Google struct {
	dataCollector obs.DataCollector
	jwtAuthority  security.JWTAuthority
	clientID      string
	clientSecret  string
	redirectURI   string
}

var _ Provider = (*Google)(nil)

func (g Google) GetName() string {
	return GoogleName
}

func (g Google) GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, error) {
	// https://developers.google.com/identity/protocols/oauth2/openid-connect#exchangecode
	idToken, err := g.getIDToken(ct, authorizationCode)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	// https://developers.google.com/identity/protocols/oauth2/openid-connect#obtainuserinfo
	tokenPayload := struct {
		UserID         string `json:"sub"`
		Issuer         string `json:"iss"`
		ExpirationTime int    `json:"exp"`
		IssuedAt       int    `json:"iat"`
		Email          string `json:"email"`
		EmailVerified  bool   `json:"email_verified"`
	}{}

	err = g.jwtAuthority.DecodeUnverifiedToken(ct, idToken, &tokenPayload)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return entity.ExternalUser{
		ID:    tokenPayload.UserID,
		Label: tokenPayload.Email,
	}, err
}

func (g Google) GetStateID(request *http.Request) (uint64, error) {
	return strconv.ParseUint(request.URL.Query().Get("state"), 10, 64)
}

func (g Google) GetAuthorizationCode(request *http.Request) string {
	return request.URL.Query().Get("code")
}

func (g Google) GetSignInURL(ct context.Context, stateID uint64) (string, error) {
	baseURL, err := url.Parse(googleAuthURLString)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	query := baseURL.Query()
	query.Add("client_id", g.clientID)
	query.Add("redirect_uri", g.redirectURI)
	query.Add("response_type", "code")
	query.Add("state", strconv.Itoa(int(stateID)))
	query.Add("scope", "openid email")
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func (g Google) getIDToken(ct context.Context, authorizationCode string) (string, error) {
	tokenBody := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
		Code         string `json:"code"`
		GrantType    string `json:"grant_type"`
	}{
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		Code:         authorizationCode,
		GrantType:    "authorization_code",
		RedirectURI:  g.redirectURI,
	}

	buf, err := json.Marshal(tokenBody)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	res, err := http.Post(googleTokenURLString, "application/json", bytes.NewReader(buf))
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	if res.StatusCode > 300 || res.StatusCode < 200 {
		err = fmt.Errorf("fail to obtain %s access token", g.GetName())
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"oauthProviderName": g.GetName(),
				"httpStatusCode":    res.StatusCode,
			},
		})
		return "", err
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	body := struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return body.IDToken, err
}

func NewGoogle(
	dataCollector obs.DataCollector,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL string,
	clientID string,
	clientSecret string,
) Google {
	return Google{
		dataCollector: dataCollector,
		jwtAuthority:  jwtAuthority,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, GoogleName),
	}
}
