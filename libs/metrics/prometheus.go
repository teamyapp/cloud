package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Prometheus struct {
	httpIncomingRequestCountMetric *prometheus.CounterVec
	httpIncomingResponseTimeMetric *prometheus.HistogramVec
	httpOutgoingRequestCountMetric *prometheus.CounterVec
	httpOutgoingResponseTimeMetric *prometheus.HistogramVec
	grpcIncomingRequestCountMetric *prometheus.CounterVec
	grpcIncomingResponseTimeMetric *prometheus.HistogramVec
	grpcOutgoingRequestCountMetric *prometheus.CounterVec
	grpcOutgoingResponseTimeMetric *prometheus.HistogramVec
}

var _ middleware.ServerHTTPMetrics = (*Prometheus)(nil)
var _ middleware.ClientHTTPMetrics = (*Prometheus)(nil)
var _ middleware.ServerGRPCMetrics = (*Prometheus)(nil)
var _ middleware.ClientGRPCMetrics = (*Prometheus)(nil)

func (p Prometheus) ReportHTTPIncomingRequest(path string) {
	p.httpIncomingRequestCountMetric.WithLabelValues(path).Inc()
}

func (p Prometheus) ReportHTTPIncomingRequestResponseTime(path string, duration time.Duration) {
	p.httpIncomingResponseTimeMetric.WithLabelValues(path).Observe(float64(duration.Milliseconds()))
}

func (p Prometheus) ReportHTTPOutgoingRequest(target string, path string) {
	p.httpOutgoingRequestCountMetric.WithLabelValues(target, path).Inc()
}

func (p Prometheus) ReportHTTPOutgoingRequestResponseTime(target string, path string, duration time.Duration) {
	p.httpOutgoingResponseTimeMetric.WithLabelValues(target, path).Observe(float64(duration.Milliseconds()))
}

func (p Prometheus) ReportGRPCIncomingRequest(method string) {
	p.grpcIncomingRequestCountMetric.WithLabelValues(method).Inc()
}

func (p Prometheus) ReportGRPCIncomingRequestResponseTime(method string, duration time.Duration) {
	p.grpcIncomingResponseTimeMetric.WithLabelValues(method).Observe(float64(duration.Milliseconds()))
}

func (p Prometheus) ReportGRPCOutgoingRequest(target string, method string) {
	p.grpcOutgoingRequestCountMetric.WithLabelValues(target, method).Inc()
}

func (p Prometheus) ReportGRPCOutgoingRequestResponseTime(target string, method string, duration time.Duration) {
	p.grpcOutgoingResponseTimeMetric.WithLabelValues(target, method).Observe(float64(duration.Milliseconds()))
}

func NewPrometheus(appMame string, serviceName string, environment env.Environment) Prometheus {
	httpIncomingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_incoming_request_count",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
		},
		[]string{"path"})
	httpIncomingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_incoming_request_response_time",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
			Buckets: telemetry.ResponseTimeBuckets,
		},
		[]string{"path"})
	httpOutgoingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_outgoing_request_count",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
		},
		[]string{"target", "path"})
	httpOutgoingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_outgoing_request_response_time",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
			Buckets: telemetry.ResponseTimeBuckets,
		},
		[]string{"target", "path"})
	grpcIncomingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_incoming_request_count",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
		},
		[]string{"method"})
	grpcIncomingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_incoming_request_response_time",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
			Buckets: telemetry.ResponseTimeBuckets,
		},
		[]string{"method"})
	grpcOutgoingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_outgoing_request_count",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
		},
		[]string{"target", "method"})
	grpcOutgoingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_outgoing_request_response_time",
			ConstLabels: map[string]string{
				telemetry.EnvironmentLabel: string(environment),
			},
			Buckets: telemetry.ResponseTimeBuckets,
		},
		[]string{"target", "method"})
	prometheus.MustRegister(httpIncomingRequestCountMetric)
	prometheus.MustRegister(httpIncomingRequestResponseTimeMetric)
	prometheus.MustRegister(httpOutgoingRequestCountMetric)
	prometheus.MustRegister(httpOutgoingRequestResponseTimeMetric)
	prometheus.MustRegister(grpcIncomingRequestCountMetric)
	prometheus.MustRegister(grpcIncomingRequestResponseTimeMetric)
	prometheus.MustRegister(grpcOutgoingRequestCountMetric)
	prometheus.MustRegister(grpcOutgoingRequestResponseTimeMetric)
	return Prometheus{
		httpIncomingRequestCountMetric: httpIncomingRequestCountMetric,
		httpIncomingResponseTimeMetric: httpIncomingRequestResponseTimeMetric,
		httpOutgoingRequestCountMetric: httpOutgoingRequestCountMetric,
		httpOutgoingResponseTimeMetric: httpOutgoingRequestResponseTimeMetric,
		grpcIncomingRequestCountMetric: grpcIncomingRequestCountMetric,
		grpcIncomingResponseTimeMetric: grpcIncomingRequestResponseTimeMetric,
		grpcOutgoingRequestCountMetric: grpcOutgoingRequestCountMetric,
		grpcOutgoingResponseTimeMetric: grpcOutgoingRequestResponseTimeMetric,
	}
}
