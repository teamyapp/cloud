package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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

const oauthProviderParam = "oauthProvider"
const serviceAccountIDParam = "serviceAccountID"

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
	logger          telemetry.Logger
	identityService service.Identity
	proto.UnimplementedIdentityServer
}

var _ runner.Service = (*Identity)(nil)
var _ proto.IdentityServer = (*Identity)(nil)

func (i Identity) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(identityPathPrefix, "verify-token"),
			HandlerFunc: i.webVerifyToken,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(identityPathPrefix, "sign-in", "oauth", runner.Param(oauthProviderParam)),
			HandlerFunc: i.webOAuthSignIn,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(identityPathPrefix, "sign-in", "oauth", runner.Param(oauthProviderParam), "finish"),
			HandlerFunc: i.webFinishOAuthSignIn,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(identityPathPrefix, "user-links"),
			HandlerFunc: i.webListUserLinks,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(identityPathPrefix, "user-links", runner.Param(oauthProviderParam), "create"),
			HandlerFunc: i.webCreateUserLink,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     path.Join(identityPathPrefix, "user-links", runner.Param(oauthProviderParam), "delete"),
			HandlerFunc: i.webDeleteUserLink,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(identityPathPrefix, "service-accounts"),
			HandlerFunc: i.webListServiceAccounts,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(identityPathPrefix, "service-accounts", "create"),
			HandlerFunc: i.webCreateServiceAccount,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(identityPathPrefix, "service-accounts", runner.Param(serviceAccountIDParam), "generate-token"),
			HandlerFunc: i.webGenerateServiceToken,
		},
		{
			Method:      http.MethodDelete,
			Pattern:     path.Join(identityPathPrefix, "service-accounts", runner.Param(serviceAccountIDParam), "delete"),
			HandlerFunc: i.webDeleteServiceAccount,
		},
	})
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterIdentityServer(server, i)
	})
	return nil
}

func (i Identity) webVerifyToken(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := errs.NewError(errs.IO, err.Error())
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	userID, isValid := i.identityService.VerifyAccessToken(ct, string(buf))
	if isValid {
		writer.WriteHeader(http.StatusAccepted)
		writer.Write([]byte(strconv.Itoa(int(userID))))
	} else {
		internalErr := errs.NewError(errs.Unauthenticated, "invalid access token")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
	}
}

func (i Identity) webOAuthSignIn(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	oauthProviderName := chi.URLParam(request, oauthProviderParam)
	query := request.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		internalErr := errs.NewError(errs.InvalidArgument, "missing redirectUrl")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	url, err := i.identityService.GenerateUnknownUserSignInURL(ct, oauthProviderName, redirectURL)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	http.Redirect(writer, request, url, http.StatusSeeOther)
}

func (i Identity) webFinishOAuthSignIn(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	oauthProviderName := chi.URLParam(request, oauthProviderParam)
	provider, err := i.identityService.GetOAuthProvider(ct, oauthProviderName)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	stateID, err := provider.GetStateID(ct, request.URL)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	authorizationCode := provider.GetAuthorizationCode(ct, request.URL)
	url, err := i.identityService.FinishOAuthSignIn(ct, oauthProviderName, authorizationCode, stateID)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	http.Redirect(writer, request, url, http.StatusSeeOther)
}

func (i Identity) webListUserLinks(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	userLinks, err := i.identityService.ListUserLinks(ct, userID)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	web.WriteJSONToResponse(writer, userLinks)
}

func (i Identity) webCreateUserLink(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	oauthProviderName := chi.URLParam(request, oauthProviderParam)
	query := request.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		internalErr := errs.NewError(errs.InvalidArgument, "missing redirectUrl")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	url, err := i.identityService.GenerateLinkUsersSignInURL(ct, oauthProviderName, userID, redirectURL)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	writer.Write([]byte(url))
}

func (i Identity) webDeleteUserLink(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	oauthProviderName := chi.URLParam(request, oauthProviderParam)
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	err := i.identityService.DeleteUserLink(ct, userID, oauthProviderName)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) webListServiceAccounts(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceAccounts, err := i.identityService.ListServiceAccounts(ct, userID)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		errs.SetHTTPErr(err, writer)
		return
	}

	web.WriteJSONToResponse(writer, serviceAccounts)
}

func (i Identity) webCreateServiceAccount(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := errs.NewError(errs.IO, err.Error())
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr := errs.NewError(errs.Deserialization, err.Error())
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	_, internalErr := i.identityService.CreateServiceAccount(ct, userID, body.Name)
	if internalErr != nil {
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) webGenerateServiceToken(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceAccountIDRaw := chi.URLParam(request, serviceAccountIDParam)
	serviceAccountID, err := strconv.ParseUint(serviceAccountIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "invalid serviceAccountId")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceToken, internalErr := i.identityService.GenerateServiceToken(ct, userID, serviceAccountID)
	if internalErr != nil {
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.Write([]byte(serviceToken))
}

func (i Identity) webDeleteServiceAccount(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(request.Context())
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "userID not found")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	serviceAccountIDRaw := chi.URLParam(request, serviceAccountIDParam)
	serviceAccountID, err := strconv.ParseUint(serviceAccountIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "invalid serviceAccountId")
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr := i.identityService.DeleteServiceAccount(ct, userID, serviceAccountID)
	if internalErr != nil {
		i.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) GetInternalUserId(ct context.Context, req *proto.GetInternalUserIdRequest) (*proto.GetInternalUserIdResponse, error) {
	internalUserID, err := i.identityService.GetInternalUserID(ct, req.AuthProvider, req.ExternalUserId)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &proto.GetInternalUserIdResponse{InternalUserId: internalUserID}, nil
}

func (i Identity) ListUserLinks(ct context.Context, req *proto.ListUserLinksRequest) (*proto.ListUserLinksResponse, error) {
	userLinks, err := i.identityService.ListUserLinks(ct, req.InternalUserId)
	if err != nil {
		i.logger.ErrorWithContext(ct, err)
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

func NewIdentity(logger telemetry.Logger, identityService service.Identity) Identity {
	return Identity{
		logger:          logger,
		identityService: identityService,
	}
}
