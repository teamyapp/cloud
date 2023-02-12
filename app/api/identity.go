package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

var skippedIdentityAPIPrefixes = []string{
	path.Join(identityPathPrefix, "verify-token"),
	path.Join(identityPathPrefix, "sign-in"),
}

func IncludeIdentityWebFunc(request *http.Request) bool {
	for _, skippedIdentityAPIPrefix := range skippedIdentityAPIPrefixes {
		if strings.HasPrefix(request.URL.Path, skippedIdentityAPIPrefix) {
			return false
		}
	}

	return true
}

var _ middleware.IncludeIdentityWebFunc = IncludeIdentityWebFunc

type Identity struct {
	dataCollector   telemetry.DataCollector
	identityService service.Identity
	proto.UnimplementedIdentityServer
}

var _ runner.Service = (*Identity)(nil)
var _ proto.IdentityServer = (*Identity)(nil)

func (i Identity) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(identityPathPrefix, "verify-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.webVerifyToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in", "oauth", "{provider}"),
			Method:      http.MethodGet,
			HandlerFunc: i.webOAuthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in", "oauth", "{provider}", "finish"),
			Method:      http.MethodGet,
			HandlerFunc: i.webFinishOAuthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "user-links"),
			Method:      http.MethodGet,
			HandlerFunc: i.webListUserLinks,
		},
		{
			Path:        path.Join(identityPathPrefix, "user-links", "{provider}", "create"),
			Method:      http.MethodGet,
			HandlerFunc: i.webCreateUserLink,
		},
		{
			Path:        path.Join(identityPathPrefix, "user-links", "{provider}", "delete"),
			Method:      http.MethodDelete,
			HandlerFunc: i.webDeleteUserLink,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts"),
			Method:      http.MethodGet,
			HandlerFunc: i.webListServiceAccounts,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "create"),
			Method:      http.MethodPost,
			HandlerFunc: i.webCreateServiceAccount,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "{serviceAccountId}", "generate-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.webGenerateServiceToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "{serviceAccountId}", "delete"),
			Method:      http.MethodDelete,
			HandlerFunc: i.webDeleteServiceAccount,
		},
	})
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterIdentityServer(server, i)
	})
	return nil
}

func (i Identity) webVerifyToken(w http.ResponseWriter, r *http.Request) {
	ct := r.Context()
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, w)
		return
	}

	userID, isValid := i.identityService.VerifyAccessToken(ct, string(buf))
	if isValid {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(strconv.Itoa(int(userID))))
	} else {
		internalErr := &errs.Error{
			Code:     errs.Unauthenticated,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, w)
	}
}

func (i Identity) webOAuthSignIn(w http.ResponseWriter, r *http.Request) {
	ct := r.Context()
	authProviderName := mux.Vars(r)["provider"]
	query := r.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("missing redirectUrl"),
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, w)
		return
	}

	url, err := i.identityService.GenerateUnknownUserSignInURL(ct, authProviderName, redirectURL)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, w)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (i Identity) webFinishOAuthSignIn(w http.ResponseWriter, r *http.Request) {
	ct := r.Context()
	providerName := mux.Vars(r)["provider"]
	provider, err := i.identityService.GetOAuthProvider(ct, providerName)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, w)
		return
	}

	stateID, err := provider.GetStateID(ct, r)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, w)
		return
	}

	authorizationCode := provider.GetAuthorizationCode(ct, r)
	url, err := i.identityService.FinishOAuthSignIn(ct, providerName, authorizationCode, stateID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, w)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (i Identity) webListUserLinks(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	userLinks, err := i.identityService.ListUserLinks(ct, userID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	web.WriteJSON(ct, i.dataCollector, writer, userLinks)
}

func (i Identity) webCreateUserLink(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	authProviderName := mux.Vars(request)["provider"]
	query := request.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("missing redirectUrl"),
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	url, err := i.identityService.GenerateLinkUsersSignInURL(ct, authProviderName, userID, redirectURL)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	writer.Write([]byte(url))
}

func (i Identity) webDeleteUserLink(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	authProviderName := mux.Vars(request)["provider"]
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	err := i.identityService.DeleteUserLink(ct, userID, authProviderName)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) webListServiceAccounts(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceAccounts, err := i.identityService.ListServiceAccounts(ct, userID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	web.WriteJSON(ct, i.dataCollector, writer, serviceAccounts)
}

func (i Identity) webCreateServiceAccount(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr := i.identityService.CreateServiceAccount(ct, userID, body.Name)
	if internalErr != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) webGenerateServiceToken(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceAccountIDParam := mux.Vars(request)["serviceAccountId"]
	serviceAccountID, err := strconv.ParseUint(serviceAccountIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("invalid serviceAccountId"),
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceToken, internalErr := i.identityService.GenerateServiceToken(ct, userID, serviceAccountID)
	if internalErr != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.Write([]byte(serviceToken))
}

func (i Identity) webDeleteServiceAccount(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "userID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceAccountIDParam := mux.Vars(request)["serviceAccountId"]
	serviceAccountID, err := strconv.ParseUint(serviceAccountIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("invalid serviceAccountId"),
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr := i.identityService.DeleteServiceAccount(ct, userID, serviceAccountID)
	if internalErr != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) GetInternalUserId(ct context.Context, req *proto.GetInternalUserIdRequest) (*proto.GetInternalUserIdResponse, error) {
	internalUserID, err := i.identityService.GetInternalUserID(ct, req.AuthProvider, req.ExternalUserId)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.GetInternalUserIdResponse{InternalUserId: internalUserID}, nil
}

func (i Identity) ListUserLinks(ct context.Context, req *proto.ListUserLinksRequest) (*proto.ListUserLinksResponse, error) {
	userLinks, err := i.identityService.ListUserLinks(ct, req.InternalUserId)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	protoUserLinks := collect.Map(userLinks, func(userLink entity.UserLink, index int) *proto.UserLink {
		return &proto.UserLink{
			AuthProvider:      userLink.AuthProvider,
			InternalUserId:    userLink.InternalUserID,
			ExternalUserId:    userLink.ExternalUserID,
			ExternalUserLabel: userLink.ExternalUserLabel,
		}
	})
	return &proto.ListUserLinksResponse{UserLinks: protoUserLinks}, nil
}

func NewIdentity(dataCollector telemetry.DataCollector, identityService service.Identity) Identity {
	return Identity{
		dataCollector:   dataCollector,
		identityService: identityService,
	}
}
