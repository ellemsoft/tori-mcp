package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func serveHTTP(addr string, handler http.Handler) {
	srv := &http.Server{Addr: addr, Handler: handler}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		serverLogf("[shutdown] flushing stats and stopping HTTP server")
		flushStats()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			serverLogf("[shutdown] forced HTTP shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		serverLogf("[server] listen failed: %v", err)
		flushStats()
		os.Exit(1)
	}
	flushStats()
}
