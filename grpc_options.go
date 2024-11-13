package server

import (
	"context"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/nakiner/go-logger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WithUnaryServerRecovery() grpc.UnaryServerInterceptor {
	recoveryHandler := recovery.WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) (err error) {
		logger.Errorf(ctx, "panic recovered: %+v, stack: %s", p, debug.Stack())
		return status.Errorf(codes.Internal, "panic triggered")
	})
	return recovery.UnaryServerInterceptor(recoveryHandler)
}

func WithStreamServerRecovery() grpc.StreamServerInterceptor {
	recoveryHandler := recovery.WithRecoveryHandlerContext(func(ctx context.Context, p interface{}) (err error) {
		logger.Errorf(ctx, "panic recovered: %+v, stack: %s", p, debug.Stack())
		return status.Errorf(codes.Internal, "panic triggered")
	})
	return recovery.StreamServerInterceptor(recoveryHandler)
}

func WithGrpcTracing() grpc.ServerOption {
	return grpc.StatsHandler(
		otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(OtelTracerProvider())),
	)
}
