package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start a background goroutine
	go func() {
		quit := make(chan os.Signal, 1)

		// notify listens to incomimg singint and sigterm and relay them to the quit channel.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// blocks until it's possible to read a sigal from the channel
		s := <-quit

		app.logger.Info("caught signal", "signal", s)
		// exits sucessfuly
		os.Exit(0)
	}()

	app.logger.Info("Starting server", slog.String("hosted_at", "http://localhost"+srv.Addr))

	return srv.ListenAndServe()
}
