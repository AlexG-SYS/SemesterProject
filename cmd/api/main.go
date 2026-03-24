package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AlexG-SYS/semesterproject/internal/routes"
	_ "github.com/lib/pq"
)

func main() {
	var cfg struct {
		port int
		env  string
		db   struct {
			dsn string
		}
		limiter struct {
			rps     float64
			burst   int
			enabled bool
		}
		cors struct {
			trustedOrigins []string
		}
	}

	// Read flags (The Makefile will pass environment variables into these)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "Environment...")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiting")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(s string) error {
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if cfg.db.dsn == "" {
		logger.Error("DB_DSN must be provided")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		logger.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Update SetupRoutes to accept the new config parameters
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      routes.SetupRoutes(db, logger, cfg.limiter.rps, cfg.limiter.burst, cfg.limiter.enabled, cfg.cors.trustedOrigins),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}

	if err := <-shutdownError; err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("stopped server")
}
