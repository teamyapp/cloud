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
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

const GitHubName = "github"

// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps
var githubAuthorizationURL = "https://github.com/login/oauth/authorize"
var githubAccessTokenURL = "https://github.com/login/oauth/access_token"
var githubUserURL = "https://api.github.com/user"

type GitHub struct {
	dataCollector telemetry.DataCollector
	httpClient    web.HTTPClient
	clientID      string
	clientSecret  string
	redirectURI   string
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
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ExternalUser{}, internalErr
	}

	// https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-
	// users-for-github-apps#3-your-github-app-accesses-the-api-with-the-users-access-token
	req, err := http.NewRequest(http.MethodGet, githubUserURL, nil)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ExternalUser{}, internalErr
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := g.httpClient.Do(req)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ExternalUser{}, internalErr
	}

	internalErr = errs.GetFromHTTPErr(res)
	if internalErr != nil {
		internalErr.Message = fmt.Sprintf("fail to obtain user ID: authProviderName=%v, httpStatusCode=%v",
			g.GetName(),
			res.StatusCode)
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ExternalUser{}, internalErr
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ExternalUser{}, internalErr
	}

	var body struct {
		UserID uint64 `json:"id"`
		Login  string `json:"login"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.ExternalUser{}, internalErr
	}

	return entity.ExternalUser{
		ID:    strconv.FormatUint(body.UserID, 10),
		Label: body.Login,
	}, nil
}

func (g GitHub) GetStateID(ct context.Context, fullURL *url.URL) (uint64, *errs.Error) {
	num, err := strconv.ParseUint(fullURL.Query().Get("state"), 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidFormat,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	return num, nil
}

func (g GitHub) GetAuthorizationCode(ct context.Context, fullURL *url.URL) string {
	return fullURL.Query().Get("code")
}

func (g GitHub) GetSignInURL(ct context.Context, stateID uint64) (string, *errs.Error) {
	baseURL, err := url.Parse(githubAuthorizationURL)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidFormat,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	req, err := http.NewRequest("POST", githubAccessTokenURL, bytes.NewReader(buf))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	res, err := g.httpClient.Do(req)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	defer res.Body.Close()

	internalErr := errs.GetFromHTTPErr(res)
	if internalErr != nil {
		internalErr.Message = fmt.Sprintf("fail to obtain user ID: authProviderName=%v, httpStatusCode=%v",
			g.GetName(),
			res.StatusCode)
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	buf, err = io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	body := struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	return body.AccessToken, nil
}

func NewGitHub(
	dataCollector telemetry.DataCollector,
	httpClient web.HTTPClient,
	webAPIBaseURL string,
	clientID string,
	clientSecret string,
) GitHub {
	return GitHub{
		dataCollector: dataCollector,
		httpClient:    httpClient,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, GitHubName),
	}
}
