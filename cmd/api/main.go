package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"greenlight/internal/data"   //Postgrees go driver
	"greenlight/internal/mailer" //Postgrees go driver

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string

	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer *mailer.Mailer
}

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug, // Debug (descartadas), Info, Warn, Error
	}))

	var config config

	flag.IntVar(&config.port, "port", 4000, "HTTP network adress  ")
	flag.StringVar(&config.env, "env", "development", "Enviroment(developement|staging|production")

	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_NAME"), "PostgresSQL")

	flag.IntVar(&config.db.maxOpenConns, "db-max-open-conns", 25, "Postgres max amount of open connection")
	flag.IntVar(&config.db.maxIdleConns, "db-max-idle-conns", 25, "Postgres max amount of connection idle")
	flag.DurationVar(&config.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgres max connection idle time")

	flag.StringVar(&config.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&config.smtp.port, "smtp-port", 2525, "SMTP posrt")
	flag.StringVar(&config.smtp.username, "smtp-username", "b2d6588c9ee528", "SMTP username")
	flag.StringVar(&config.smtp.password, "smtp-password", "1ce6d8c9fdee78", "SMTP password")
	flag.StringVar(&config.smtp.sender, "smtp-sender", "Magic Elves <ola@example.com>", "SMTP sender")

	flag.Parse()

	db, err := openDB(config)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	mailer, err := mailer.New(config.smtp.host, config.smtp.port, config.smtp.username, config.smtp.password, config.smtp.sender)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config: config,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("database conection established")
	logger.Info("Starting server", slog.String("hosted_at", "http://localhost"+srv.Addr))

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)

}

func openDB(config config) (*sql.DB, error) {

	db, err := sql.Open("postgres", config.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(config.db.maxIdleTime)
	db.SetMaxOpenConns(config.db.maxOpenConns)
	db.SetMaxIdleConns(config.db.maxIdleConns)

	// create a contex
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	// if the connection was not successful in 5 seconds this returns an error
	err = db.PingContext(ctx)

	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil

}
