package server

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type grpcServer struct {
	listener        net.Listener
	server          *grpc.Server
	port            int
	gracefulTimeout time.Duration
	opts            []grpc.ServerOption
}

func newGRPCServer(port int, gracefulTimeout time.Duration) *grpcServer {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(errors.Wrap(err, "grpc listener init failure"))
	}

	r := &grpcServer{
		listener:        listener,
		port:            port,
		gracefulTimeout: gracefulTimeout,
	}

	return r
}

func (r *grpcServer) SetServerOptions(opt ...grpc.ServerOption) {
	r.opts = append(r.opts, opt...)
}

func (r *grpcServer) setupServer() {
	r.server = grpc.NewServer(r.opts...)
}

func (r *grpcServer) actor() (func() error, func(error)) {
	return func() error {
			appLogger.Warnw("grpc server started", "port", r.port)
			return errors.Wrap(r.server.Serve(r.listener), "grpc server")
		}, func(err error) {
			doneCh := make(chan struct{})
			go func() {
				r.server.GracefulStop()
				close(doneCh)
			}()

			select {
			case <-time.After(r.gracefulTimeout):
				appLogger.Error(errors.Wrap(context.DeadlineExceeded, "grpc server graceful stop timed out"))
				r.server.Stop()
				appLogger.Warn("grpc server stopped (force)")
			case <-doneCh:
				appLogger.Warn("grpc server: gracefully stopped")
			}
		}
}
