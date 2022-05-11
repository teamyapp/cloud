package web

import (
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/service"
)

const identityPathPrefix = "/identity"

type IdentityAPI struct {
	identityService service.Identity
}

var _ Service = (*IdentityAPI)(nil)

func (i IdentityAPI) getRoutes() []Route {
	return []Route{
		{
			Path:        path.Join(identityPathPrefix, "verify-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.verifyToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in/oauth/{provider}"),
			Method:      http.MethodGet,
			HandlerFunc: i.oauthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in/oauth/{provider}/finish"),
			Method:      http.MethodGet,
			HandlerFunc: i.finishOAuthSignIn,
		},
	}
}

func (i IdentityAPI) verifyToken(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, isValid := i.identityService.VerifyAccessToken(string(buf))
	if isValid {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(strconv.Itoa(int(userID))))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (i IdentityAPI) oauthSignIn(w http.ResponseWriter, r *http.Request) {
	authProviderName := mux.Vars(r)["provider"]
	query := r.URL.Query()
	redirectURL := query.Get("redirectUrl")

	url, err := i.identityService.GenerateSignInURL(authProviderName, redirectURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (i IdentityAPI) finishOAuthSignIn(w http.ResponseWriter, r *http.Request) {
	providerName := mux.Vars(r)["provider"]
	provider, err := i.identityService.GetOAuthProvider(providerName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stateID, err := provider.GetStateID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizationCode := provider.GetAuthorizationCode(r)
	url, err := i.identityService.FinishOAuthSignIn(providerName, authorizationCode, stateID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func NewIdentityAPI(identityService service.Identity) IdentityAPI {
	return IdentityAPI{
		identityService: identityService,
	}
}
