package server

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nakiner/go-logger"
)

func interruptActor(delay time.Duration, sig ...os.Signal) (func() error, func(error)) {
	sig = append(sig, syscall.SIGTERM, syscall.SIGINT)
	sigCh := make(chan os.Signal, 1)
	doneCh := make(chan struct{})
	return func() error {
			signal.Notify(sigCh, sig...)
			select {
			case <-doneCh:
			case <-sigCh:
				logger.Logger().Warnf("waiting graceful delay: %s", delay)
			}
			select {
			case <-time.After(delay):
			case <-sigCh:
				logger.Logger().Fatal("force quit")
			}
			return nil
		}, func(err error) {
			signal.Stop(sigCh)
			close(doneCh)
		}
}
