package gql

import (
	_ "embed"
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graph-gophers/graphql-go/trace/tracer"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Service[Resolver any] struct {
	dataCollector telemetry.DataCollector
	graphQLTracer tracer.Tracer
	schema        string
	resolver      *Resolver
	pathPrefix    string
}

var _ runner.Service = (*Service[any])(nil)

func (s Service[Resolver]) Start(rn *runner.ServiceRunner) *errs.Error {
	schema, err := graphql.ParseSchema(s.schema, s.resolver,
		graphql.UseFieldResolvers(),
		graphql.UseStringDescriptions(),
		graphql.Tracer(s.graphQLTracer))
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

func NewService[Resolver any](
	dataCollector telemetry.DataCollector,
	graphQLTracer tracer.Tracer,
	schema string,
	resolver *Resolver,
	pathPrefix string,
) Service[Resolver] {
	return Service[Resolver]{
		dataCollector: dataCollector,
		graphQLTracer: graphQLTracer,
		schema:        schema,
		resolver:      resolver,
		pathPrefix:    pathPrefix,
	}
}
