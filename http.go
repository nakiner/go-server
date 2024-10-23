package server

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type httpServer struct {
	listener        net.Listener
	debugListener   net.Listener
	server          *http.Server
	debugServer     *http.Server
	port            int
	debugPort       int
	gracefulTimeout time.Duration
}

func newHTTPServer(port int, debugPort int, gracefulTimeout time.Duration) *httpServer {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(errors.Wrap(err, "http listener init failure"))
	}
	debugListener, err := net.Listen("tcp", ":"+strconv.Itoa(debugPort))
	if err != nil {
		panic(errors.Wrap(err, "http debug listener init failure"))
	}
	server := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
	}
	debugServer := &http.Server{
		ReadHeaderTimeout: 30 * time.Second,
	}

	r := &httpServer{
		listener:        listener,
		debugListener:   debugListener,
		server:          server,
		debugServer:     debugServer,
		port:            port,
		debugPort:       debugPort,
		gracefulTimeout: gracefulTimeout,
	}

	return r
}

func (r *httpServer) actor() (func() error, func(error)) {
	return func() error {
			appLogger.Warnw("http server started", "port", r.port)
			return errors.Wrap(r.server.Serve(r.listener), "http server")
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), r.gracefulTimeout)
			defer cancel()

			r.server.SetKeepAlivesEnabled(false)
			if err := errors.Wrap(r.server.Shutdown(ctx), "http server: error during shutdown"); err != nil {
				appLogger.Error("http server: stop failure", err)
			} else {
				appLogger.Warn("http server: gracefully stopped")
			}
		}
}

func (r *httpServer) debugActor() (func() error, func(error)) {
	return func() error {
			appLogger.Warnw("http debug server started", "port", r.debugPort)
			return errors.Wrap(r.debugServer.Serve(r.debugListener), "http debug server")
		}, func(err error) {
			ctx, cancel := context.WithTimeout(context.Background(), r.gracefulTimeout)
			defer cancel()

			r.debugServer.SetKeepAlivesEnabled(true)
			if err := errors.Wrap(r.debugServer.Shutdown(ctx), "http debug server: error during shutdown"); err != nil {
				appLogger.Error("http debug server: stop failure", err)
			} else {
				appLogger.Warn("http debug server: gracefully stopped")
			}
		}
}
