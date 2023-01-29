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
	"github.com/teamyapp/cloud/libs/telemetry"
)

const GitHubName = "github"

// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps
var githubAuthorizationURL = "https://github.com/login/oauth/authorize"
var githubAccessTokenURL = "https://github.com/login/oauth/access_token"
var githubUserURL = "https://api.github.com/user"

type GitHub struct {
	dataCollector telemetry.DataCollector
	clientID      string
	clientSecret  string
	redirectURI   string
}

var _ Provider = (*GitHub)(nil)

func (g GitHub) GetName() string {
	return GitHubName
}

func (g GitHub) GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, error) {
	// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-
	// users-for-github-apps#2-users-are-redirected-back-to-your-site-by-github
	accessToken, err := g.getAccessToken(ct, authorizationCode)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-
	// users-for-github-apps#3-your-github-app-accesses-the-api-with-the-users-access-token
	req, err := http.NewRequest("GET", githubUserURL, nil)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	if res.StatusCode > 300 || res.StatusCode < 200 {
		err = fmt.Errorf("fail to obtain %s user ID", g.GetName())
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: err,
			"AuthProviderName":  g.GetName(),
			"HttpStatusCode":    res.StatusCode,
		})
		return entity.ExternalUser{}, err
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	var body struct {
		UserID uint64 `json:"id"`
		Login  string `json:"login"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.ExternalUser{}, err
	}

	return entity.ExternalUser{
		ID:    strconv.FormatUint(body.UserID, 10),
		Label: body.Login,
	}, nil
}

func (g GitHub) GetStateID(request *http.Request) (uint64, error) {
	return strconv.ParseUint(request.URL.Query().Get("state"), 10, 64)
}

func (g GitHub) GetAuthorizationCode(request *http.Request) string {
	return request.URL.Query().Get("code")
}

func (g GitHub) GetSignInURL(ct context.Context, stateID uint64) (string, error) {
	baseURL, err := url.Parse(githubAuthorizationURL)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	// Github app does not require "scopes" in your authorization request
	query := baseURL.Query()
	query.Add("client_id", g.clientID)
	query.Add("redirect_uri", g.redirectURI)
	query.Add("state", strconv.Itoa(int(stateID)))
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func (g GitHub) getAccessToken(ct context.Context, authorizationCode string) (string, error) {
	tokenBody := struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}{
		ClientID:     g.clientID,
		ClientSecret: g.clientSecret,
		Code:         authorizationCode,
	}

	buf, err := json.Marshal(tokenBody)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	req, err := http.NewRequest("POST", githubAccessTokenURL, bytes.NewReader(buf))
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}
	res.Body.Close()

	if res.StatusCode > 300 || res.StatusCode < 200 {
		err = fmt.Errorf("fail to obtain %s access token", g.GetName())
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: err,
			"OauthProviderName": g.GetName(),
			"HttpStatusCode":    res.StatusCode,
		})
		return "", err
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	body := struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return body.AccessToken, err
}

func NewGitHub(dataCollector telemetry.DataCollector, webAPIBaseURL string, clientID string, clientSecret string) GitHub {
	return GitHub{
		dataCollector: dataCollector,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, GitHubName),
	}
}
