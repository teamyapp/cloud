package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const TeamyOpenTelemetryTracerName = "com.teamyapp.open-telemetry.tracer"

func InitTracerProvider(
	dataCollector DataCollector,
	traceCollectorEndpoint string,
	serviceName string,
) (func(ct context.Context) error, *errs.Error) {
	ct := context.Background()
	res, err := resource.New(
		ct,
		resource.WithAttributes(
			semconv.ServiceName(serviceName)))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(internalErr)
		return nil, internalErr
	}

	ct, _ = context.WithTimeout(ct, time.Second)
	conn, err := grpc.DialContext(
		ct,
		traceCollectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(internalErr)
		return nil, internalErr
	}

	dataCollector.Logger.Info(fmt.Sprintf("Connected to trace collector at %v", traceCollectorEndpoint))
	traceExporter, err := otlptracegrpc.New(ct, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Error(internalErr)
		return nil, internalErr
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp))

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{}))

	return tracerProvider.Shutdown, nil
}
