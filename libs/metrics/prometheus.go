package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teamyapp/cloud/libs/env"
)

var responseTimeBuckets = prometheus.ExponentialBuckets(1, 2, 20)

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

var _ Metrics = (*Prometheus)(nil)

func (p Prometheus) ReportHTTPIncomingRequest(method string, pattern string) {
	p.httpIncomingRequestCountMetric.WithLabelValues(method, pattern).Inc()
}

func (p Prometheus) ReportHTTPIncomingRequestResponseTime(method string, pattern string, duration time.Duration) {
	p.httpIncomingResponseTimeMetric.WithLabelValues(method, pattern).Observe(float64(duration.Milliseconds()))
}

func (p Prometheus) ReportHTTPOutgoingRequest(target string, method string, pattern string) {
	p.httpOutgoingRequestCountMetric.WithLabelValues(target, method, pattern).Inc()
}

func (p Prometheus) ReportHTTPOutgoingRequestResponseTime(target string, method string, pattern string, duration time.Duration) {
	p.httpOutgoingResponseTimeMetric.WithLabelValues(target, method, pattern).Observe(float64(duration.Milliseconds()))
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
				EnvironmentLabel: string(environment),
			},
		},
		[]string{"method", "pattern"})
	httpIncomingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_incoming_request_response_time",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
			Buckets: responseTimeBuckets,
		},
		[]string{"method", "pattern"})
	httpOutgoingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_outgoing_request_count",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
		},
		[]string{"target", "method", "pattern"})
	httpOutgoingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "http_outgoing_request_response_time",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
			Buckets: responseTimeBuckets,
		},
		[]string{"target", "method", "pattern"})
	grpcIncomingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_incoming_request_count",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
		},
		[]string{"method"})
	grpcIncomingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_incoming_request_response_time",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
			Buckets: responseTimeBuckets,
		},
		[]string{"method"})
	grpcOutgoingRequestCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_outgoing_request_count",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
		},
		[]string{"target", "method"})
	grpcOutgoingRequestResponseTimeMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "grpc_outgoing_request_response_time",
			ConstLabels: map[string]string{
				EnvironmentLabel: string(environment),
			},
			Buckets: responseTimeBuckets,
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
