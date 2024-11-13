package server

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

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
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(),
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
