package gql

import (
	"context"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/trace/tracer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/metrics"
)

type PrometheusTracer struct {
	requestCountMetric *prometheus.CounterVec
	responseTimeMetric *prometheus.HistogramVec
}

var _ tracer.Tracer = (*PrometheusTracer)(nil)

func (p PrometheusTracer) TraceQuery(
	ctx context.Context,
	queryString string,
	operationName string,
	variables map[string]interface{},
	varTypes map[string]*introspection.Type,
) (context.Context, tracer.QueryFinishFunc) {
	return ctx, func(errors []*errors.QueryError) {}
}

func (p PrometheusTracer) TraceField(
	ctx context.Context,
	label string,
	typeName string,
	fieldName string,
	trivial bool,
	args map[string]interface{},
) (context.Context, tracer.FieldFinishFunc) {
	path := fmt.Sprintf("%v.%v", typeName, fieldName)
	p.requestCountMetric.
		WithLabelValues(path).
		Inc()
	startAt := time.Now()
	return ctx, func(queryError *errors.QueryError) {
		responseTime := time.Now().Sub(startAt)
		p.responseTimeMetric.
			WithLabelValues(path).
			Observe(float64(responseTime.Milliseconds()))
	}
}

func NewPrometheusTracer(appMame string, serviceName string, environment env.Environment) PrometheusTracer {
	requestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "graphql_incoming_request_count",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
		},
		[]string{"path"})
	responseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "graphql_incoming_response_time",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
			Buckets: metrics.ResponseTimeBuckets,
		},
		[]string{"path"})
	prometheus.MustRegister(requestCountMetric)
	prometheus.MustRegister(responseTimeMetric)
	return PrometheusTracer{
		requestCountMetric: requestCountMetric,
		responseTimeMetric: responseTimeMetric,
	}
}
