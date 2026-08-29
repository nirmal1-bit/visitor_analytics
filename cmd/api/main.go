package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"visitor_analytics/internal/data"
	"visitor_analytics/internal/geo"
)

const version = "1.0.0"

type config struct {
	port    int
	env     string
	baseURL string
}

type application struct {
	config      config
	logger      *slog.Logger
	models      data.Models
	geoProvider geo.Provider
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.baseURL, "base-url", "", "Public base URL (e.g. https://analytics.example.com). If empty, falls back to http://localhost:<port>")
	flag.Parse()

	// Default fallback to http://localhost:<port> if not provided
	if cfg.baseURL == "" {
		cfg.baseURL = fmt.Sprintf("http://localhost:%d", cfg.port)
	}


	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	models := data.NewModels()
	geoProvider := geo.NewIPAPIProvider(5 * time.Second)

	app := &application{
		config:      cfg,
		logger:      logger,
		models:      models,
		geoProvider: geoProvider,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
