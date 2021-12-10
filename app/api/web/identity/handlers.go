package identity

import (
	_ "embed"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/channel"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/one/entity"
)

//go:embed sign-in-succeed.html
var signInSucceededView []byte

func newOAuthSignInHandler(identityService service.Identity) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		oauthProviderName := mux.Vars(request)["oauth-provider"]
		query := request.URL.Query()
		sessionID, err := strconv.Atoi(query.Get("session-id"))
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		url, err := identityService.RequestOAuthSignInURL(oauthProviderName, entity.ID(sessionID))
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		http.Redirect(writer, request, url, http.StatusSeeOther)
	}
}

func newOAuthSignInFinishHandler(identityService service.Identity) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// TODO: research gorilla router to get the path parameter
		providerName := mux.Vars(request)["oauth-provider"]
		oauth, err := identityService.GetOAuthProvider(providerName)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionID, err := strconv.Atoi(oauth.GetStateID(request))
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		authorizationCode := oauth.GetAuthorizationCode(request)
		err = identityService.FinishOAuthSignIn(authorizationCode, entity.ID(sessionID), providerName)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			writer.WriteHeader(http.StatusAccepted)
			writer.Write(signInSucceededView)
		}
	}
}

func newGetNewSessionIDHandler(identityService service.Identity) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		sessionID, err := identityService.NewSessionID()
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Generated new session ID: %d", sessionID)
		writer.Write([]byte(strconv.Itoa(int(sessionID))))
	}
}

func newSubscribeSessionHandler(
	webSocketOriginChecker channel.WebSocketOriginChecker,
	identityService service.Identity,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		sessionID, err := strconv.Atoi(mux.Vars(request)["session-id"])
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		webSocketChannel, err := channel.NewWebSocket(webSocketOriginChecker, writer, request)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		identityService.ClientSubscribe(webSocketChannel, entity.ID(sessionID))
	}
}

func newVerifyAccessTokenHandler(identityService service.Identity) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, isValid := identityService.VerifyAccessToken(string(buf))
		if isValid {
			writer.WriteHeader(http.StatusAccepted)
			writer.Write([]byte(strconv.Itoa(int(userID))))
		} else {
			writer.WriteHeader(http.StatusUnauthorized)
		}
	}
}
