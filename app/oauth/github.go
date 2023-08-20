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
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/web"
)

const GitHubName = "github"

// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps
var githubAuthorizationURL = "https://github.com/login/oauth/authorize"
var githubAccessTokenURL = "https://github.com/login/oauth/access_token"
var githubUserURL = "https://api.github.com/user"

type GitHub struct {
	httpClient   web.HTTPClient
	clientID     string
	clientSecret string
	redirectURI  string
}

var _ Provider = (*GitHub)(nil)

func (g GitHub) GetName() string {
	return GitHubName
}

func (g GitHub) GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, *errs.Error) {
	// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-
	// users-for-github-apps#2-users-are-redirected-back-to-your-site-by-github
	accessToken, internalErr := g.getAccessToken(ct, authorizationCode)
	if internalErr != nil {
		return entity.ExternalUser{}, internalErr
	}

	// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-
	// users-for-github-apps#3-your-github-app-accesses-the-api-with-the-users-access-token
	req, err := http.NewRequest(http.MethodGet, githubUserURL, nil)
	if err != nil {
		return entity.ExternalUser{}, errs.NewError(errs.Unknown, err.Error())
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := g.httpClient.Do(req)
	if err != nil {
		return entity.ExternalUser{}, errs.NewError(errs.Unknown, err.Error())
	}

	internalErr = errs.GetFromHTTPErr(res)
	if internalErr != nil {
		internalErr.Message = fmt.Sprintf("fail to obtain user ID: authProviderName=%v, httpStatusCode=%v",
			g.GetName(),
			res.StatusCode)
		return entity.ExternalUser{}, internalErr
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return entity.ExternalUser{}, errs.NewError(errs.IO, err.Error())
	}

	var body struct {
		NodeID string `json:"node_id"`
		Login  string `json:"login"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		return entity.ExternalUser{}, errs.NewError(errs.Deserialization, err.Error())
	}

	return entity.ExternalUser{
		ID:    body.NodeID,
		Label: body.Login,
	}, nil
}

func (g GitHub) GetStateID(ct context.Context, fullURL *url.URL) (uint64, *errs.Error) {
	num, err := strconv.ParseUint(fullURL.Query().Get("state"), 10, 64)
	if err != nil {
		return 0, errs.NewError(errs.InvalidFormat, err.Error())
	}

	return num, nil
}

func (g GitHub) GetAuthorizationCode(ct context.Context, fullURL *url.URL) string {
	return fullURL.Query().Get("code")
}

func (g GitHub) GetSignInURL(ct context.Context, stateID uint64) (string, *errs.Error) {
	baseURL, err := url.Parse(githubAuthorizationURL)
	if err != nil {
		return "", errs.NewError(errs.InvalidFormat, err.Error())
	}

	// Github app does not require "scopes" in your authorization request
	query := baseURL.Query()
	query.Add("client_id", g.clientID)
	query.Add("redirect_uri", g.redirectURI)
	query.Add("state", strconv.Itoa(int(stateID)))
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func (g GitHub) getAccessToken(ct context.Context, authorizationCode string) (string, *errs.Error) {
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
		return "", errs.NewError(errs.Serialization, err.Error())
	}

	req, err := http.NewRequest("POST", githubAccessTokenURL, bytes.NewReader(buf))
	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	res, err := g.httpClient.Do(req)
	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	defer res.Body.Close()

	internalErr := errs.GetFromHTTPErr(res)
	if internalErr != nil {
		internalErr.Message = fmt.Sprintf("fail to obtain user ID: authProviderName=%v, httpStatusCode=%v",
			g.GetName(),
			res.StatusCode)
		return "", internalErr
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		return "", errs.NewError(errs.IO, err.Error())
	}

	body := struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		return "", errs.NewError(errs.Deserialization, err.Error())
	}

	return body.AccessToken, nil
}

func NewGitHub(
	httpClient web.HTTPClient,
	webAPIBaseURL string,
	clientID string,
	clientSecret string,
) GitHub {
	return GitHub{
		httpClient:   httpClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, GitHubName),
	}
}
