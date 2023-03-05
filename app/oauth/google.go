package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

const GoogleName = "google"

// https://developers.google.com/identity/protocols/oauth2/web-server#httprest_1
var googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
var googleTokenURL = "https://oauth2.googleapis.com/token"

type Google struct {
	dataCollector telemetry.DataCollector
	httpClient    web.HTTPClient
	jwtAuthority  security.JWTAuthority
	clientID      string
	clientSecret  string
	redirectURI   string
}

var _ Provider = (*Google)(nil)

func (g Google) GetName() string {
	return GoogleName
}

func (g Google) GetUser(ct context.Context, authorizationCode string) (entity.ExternalUser, *errs.Error) {
	// https://developers.google.com/identity/protocols/oauth2/openid-connect#exchangecode
	idToken, err := g.getIDToken(ct, authorizationCode)
	if err != nil {
		g.dataCollector.Logger.ErrorWithContext(ct, err)
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
		g.dataCollector.Logger.ErrorWithContext(ct, err)
	}

	return entity.ExternalUser{
		ID:    tokenPayload.UserID,
		Label: tokenPayload.Email,
	}, err
}

func (g Google) GetStateID(ct context.Context, fullURL *url.URL) (uint64, *errs.Error) {
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

func (g Google) GetAuthorizationCode(ct context.Context, fullURL *url.URL) string {
	return fullURL.Query().Get("code")
}

func (g Google) GetSignInURL(ct context.Context, stateID uint64) (string, *errs.Error) {
	baseURL, err := url.Parse(googleAuthURL)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
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

func (g Google) getIDToken(ct context.Context, authorizationCode string) (string, *errs.Error) {
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
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	req, err := http.NewRequest(http.MethodPost, googleTokenURL, nil)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	web.WriteJSONToRequest(ct, g.dataCollector, req, buf)
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
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
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

	return body.IDToken, nil
}

func NewGoogle(
	dataCollector telemetry.DataCollector,
	httpClient web.HTTPClient,
	jwtAuthority security.JWTAuthority,
	webAPIBaseURL string,
	clientID string,
	clientSecret string,
) Google {
	return Google{
		dataCollector: dataCollector,
		httpClient:    httpClient,
		jwtAuthority:  jwtAuthority,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   fmt.Sprintf("%s/identity/sign-in/oauth/%s/finish", webAPIBaseURL, GoogleName),
	}
}
