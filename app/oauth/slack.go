package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

const slackName = "slack"

// https://api.slack.com/authentication/sign-in-with-slack
const slackAuthorizeURL = "https://slack.com/openid/connect/authorize"
const slackTokenURL = "https://slack.com/api/openid.connect.token"

type Slack struct {
	dataCollector telemetry.DataCollector
	httpClient    web.HTTPClient
	jwtAuthority  security.JWTAuthority
	clientID      string
	clientSecret  string
	redirectURI   string
}

var _ Provider = (*Slack)(nil)

func (s Slack) GetName() string {
	return slackName
}

func (s Slack) GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, *errs.Error) {
	idToken, err := s.getIDToken(ct, authorizationCode)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
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
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.ExternalUser{}, err
	}

	return entity.ExternalUser{
		ID:    tokenPayload.UserID,
		Label: tokenPayload.Email,
	}, nil
}

func (s Slack) GetStateID(ct context.Context, fullURL *url.URL) (uint64, *errs.Error) {
	num, err := strconv.ParseUint(fullURL.Query().Get("state"), 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidFormat,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	return num, nil
}

func (s Slack) GetAuthorizationCode(ct context.Context, fullURL *url.URL) string {
	return fullURL.Query().Get("code")
}

func (s Slack) GetSignInURL(ct context.Context, stateID uint64) (string, *errs.Error) {
	baseURL, err := url.Parse(slackAuthorizeURL)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
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

func (s Slack) getIDToken(ct context.Context, authorizationCode string) (string, *errs.Error) {
	// API accepts request in encoded form data only
	// https://api.slack.com/methods/openid.connect.token
	formData := url.Values{}
	formData.Set("client_id", s.clientID)
	formData.Set("client_secret", s.clientSecret)
	formData.Set("code", authorizationCode)
	formData.Set("grant_type", "authorization_code")
	formData.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequest("POST", slackTokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := s.httpClient.Do(req)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	defer res.Body.Close()

	internalErr := errs.GetFromHTTPErr(res)
	if internalErr != nil {
		internalErr.Message = fmt.Sprintf("fail to obtain access token: oauthProviderName=%v, httpStatusCode=%v",
			s.GetName(),
			res.StatusCode)
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	body := struct {
		OK          bool   `json:"ok"`
		Error       string `json:"error"`
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	if !body.OK {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	return body.IDToken, nil
}

func NewSlack(
	dataCollector telemetry.DataCollector,
	httpClient web.HTTPClient,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL string,
	clientID string,
	clientSecret string,
) Slack {
	return Slack{
		dataCollector: dataCollector,
		httpClient:    httpClient,
		jwtAuthority:  jwtAuthority,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, slackName),
	}
}
