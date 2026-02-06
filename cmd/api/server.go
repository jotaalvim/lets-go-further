package main

import (
	"context"
	"errors"
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

	shutdownError := make(chan error)

	// start a background goroutine
	go func() {
		quit := make(chan os.Signal, 1)

		// notify listens to incomimg singint and sigterm and relay them to the quit channel.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// blocks until it's possible to read a sigal from the channel
		s := <-quit

		app.logger.Info("shutingdown server", "signal", s.String())
		// exits sucessfuly

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		//we only send it to shutdown channel if it returns an error
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr)

		//wait until the WaitGroup counter is zero
		app.wg.Wait()

		shutdownError <- nil

	}()

	app.logger.Info("Starting server", slog.String("hosted_at", "http://localhost"+srv.Addr))

	// calling Shutdown will immidatly return an err on Listen and Server()
	err := srv.ListenAndServe() //blocks

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// er wait to receive a shutdown error from its channel
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("Stopped server", "addr", srv.Addr)

	return nil
}
