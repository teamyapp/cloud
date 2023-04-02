package oauth

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/fakeapi"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/cloud/libs/web/webtest"
)

func TestGithub_GetUser(t *testing.T) {
	virtualNetwork := networktest.NewVirtualNetwork()
	githubTestKitConfig := fakeapi.GithubTestKitConfig{
		WebServerPort:  80,
		GRPCServerPort: 81,
	}
	githubTestKit := fakeapi.NewGithubTestKit(githubTestKitConfig, virtualNetwork)
	fakeapi.StartGithubServiceInstance(
		githubTestKitConfig.WebServerPort,
		virtualNetwork,
		githubTestKit.ServiceInstanceRunner)

	httpClient := webtest.InsecureHTTPClient(virtualNetwork)

	clientID := "123"
	secret := "randomSecret"
	fakeGithubAPI := githubTestKit.Refs.FakeGithubAPI
	fakeGithubAPI.RegisterClient(clientID, secret)
	user1NodeID := base64.StdEncoding.EncodeToString([]byte("04:User1"))
	githubUser1 := fakeapi.GithubUser{
		ID:     1,
		NodeID: user1NodeID,
		Login:  "Test",
	}
	fakeGithubAPI.RegisterUser(&githubUser1)
	githubOAuth := NewGitHub(httpClient, "http://localhost", clientID, secret)

	ct := context.Background()
	signInURL, internalErr := githubOAuth.GetSignInURL(ct, 1)
	assert.Nil(t, internalErr)
	if internalErr != nil {
		return
	}

	// Simulate browser to navigate to the signInURL
	req, err := http.NewRequest(http.MethodGet, signInURL, nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	response, err := httpClient.Do(req)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, http.StatusOK, response.StatusCode)
	buf, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// Simulate user select and sign in an account
	selectUserURI := string(buf)
	req, err = http.NewRequest(http.MethodPost, selectUserURI, nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	selectUserBody := struct {
		UserID int64 `json:"user_id"`
	}{
		UserID: 1,
	}
	web.WriteJSONToRequest(ct, req, selectUserBody)
	response, err = httpClient.Do(req)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// Simulate user is redirected to cloud identity API after selecting external
	// OAuth account
	rawRedirectURI := response.Header.Get("Location")
	redirectURI, err := url.Parse(rawRedirectURI)
	assert.Nil(t, err)
	if err != nil {
		return
	}

	authorizationCode := githubOAuth.GetAuthorizationCode(ct, redirectURI)
	externalUser, internalErr := githubOAuth.GetUser(ct, authorizationCode)
	assert.Nil(t, internalErr)
	if err != nil {
		return
	}

	expectedExternalUser := entity.ExternalUser{
		ID:    githubUser1.NodeID,
		Label: githubUser1.Login,
	}
	assert.Equal(t, expectedExternalUser, externalUser)
}
