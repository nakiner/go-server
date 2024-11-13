package server

import (
	"context"
	"os"

	"github.com/nakiner/go-logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type OtelJSONErrorHandler struct{}

var (
	otelTracerProvider *sdktrace.TracerProvider
)

func OtelTracer() trace.Tracer {
	if otelTracerProvider == nil {
		otelTracerProvider = sdktrace.NewTracerProvider()
	}
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		return otelTracerProvider.Tracer(serviceName)
	}
	return otelTracerProvider.Tracer("unknown")
}

func OtelTracerProvider() *sdktrace.TracerProvider {
	if otelTracerProvider == nil {
		otelTracerProvider = sdktrace.NewTracerProvider()
	}
	return otelTracerProvider
}

func initTracerProvider(serviceName string, otlpEndpoint string) {
	otel.SetErrorHandler(&OtelJSONErrorHandler{})

	if otlpEndpoint == "" {
		otelTracerProvider = sdktrace.NewTracerProvider()
		return
	}

	ctx := context.Background()

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithCompressor("gzip"),
	)
	if err != nil {
		otelTracerProvider = sdktrace.NewTracerProvider()
		return
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		otelTracerProvider = sdktrace.NewTracerProvider()
		return
	}

	otelTracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)
}

func (h *OtelJSONErrorHandler) Handle(err error) {
	if err != nil {
		logger.ErrorKV(context.Background(), "otel err", "err", err)
	}
}
