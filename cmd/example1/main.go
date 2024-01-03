package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jjngx/otexample"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Handle ctrl+C gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// set up OT
	svcName := "dice"
	svcVersion := "0.1.0"
	shutdown, err := otexample.SetupOtelSDK(ctx, svcName, svcVersion)
	if err != nil {
		return err
	}

	// Handle shutdown properly, so nothing leaks
	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	// Start HTTP Server
	srv := &http.Server{
		Addr:         ":8085",
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for possible interruption
	select {
	case err = <-srvErr:
		// Error when starting the HTTP server
		return err
	case <-ctx.Done():
		// Wait for first CTRL+C
		stop()
	}

	// When shutdown is called, ListenAndServe immediately returns ErrServerClosed
	return srv.Shutdown(context.Background())
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	handleFunc := func(pattern string, handlerFn func(http.ResponseWriter, *http.Request)) {
		// http route for the instrumentation
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFn))
		mux.Handle(pattern, handler)
	}

	// register handlers
	handleFunc("/rolldice", otexample.RollDice)

	// add HTTP instrumentation for the whole server
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}
