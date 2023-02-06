package gql

import (
	_ "embed"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/gql"
)

type Service struct {
	dataCollector telemetry.DataCollector
	schema        string
	resolver      interface{}
	pathPrefix    string
}

var _ runner.Service = (*Service)(nil)

func (s Service) Start(rn *runner.ServiceRunner) *errs.Error {
	schema, err := graphql.ParseSchema(s.schema, &s.resolver,
		graphql.UseFieldResolvers(),
		graphql.UseStringDescriptions())
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.NotReady,
			EmbedErr: err,
		}
		s.dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	relayHandler := relay.Handler{Schema: schema}
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        s.pathPrefix,
			Method:      http.MethodPost,
			HandlerFunc: relayHandler.ServeHTTP,
		},
	})
	return nil
}

func NewService(
	dataCollector telemetry.DataCollector,
	schema string,
	resolver gql.Resolver,
	pathPrefix string,
) Service {
	return Service{
		dataCollector: dataCollector,
		schema:        schema,
		resolver:      resolver,
		pathPrefix:    pathPrefix,
	}
}
