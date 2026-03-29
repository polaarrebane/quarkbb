// Package main implements the QuarkBB bulletin board system server.
// This is the entry point for the application that sets up HTTP server,
// database connections, dependency injection, and graceful shutdown.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/ronia/quarkbb/internal/auth"
	"codeberg.org/ronia/quarkbb/internal/config"
	"codeberg.org/ronia/quarkbb/internal/repository"
	"codeberg.org/ronia/quarkbb/internal/repository/postgres"
	"codeberg.org/ronia/quarkbb/internal/router"
	"codeberg.org/ronia/quarkbb/internal/security"
	"codeberg.org/ronia/quarkbb/internal/validator"
	"github.com/jackc/pgx/v5/pgxpool"
)

// main is the application entry point. It initializes configuration,
// database connection, HTTP server with all routes, and handles graceful shutdown.
// The server supports user registration and authentication endpoints.
func main() {

	startCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	appConfig, err := config.New()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(appConfig.DSN())
	if err != nil {
		log.Fatalf("parse dsn: %v", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(startCtx, cfg)
	if err != nil {
		log.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	sqlcQueries := postgres.New(pool)
	authRepo := repository.NewUserRepository(sqlcQueries)
	rtRepo := repository.NewRefreshTokenRepository(sqlcQueries)
	sessionsRepo := repository.NewSessionRepository(sqlcQueries)

	js, err := security.NewJWTService(*appConfig)
	if err != nil {
		log.Fatalf("jwt service: %v", err)
	}
	authSvc := auth.NewService(authRepo, rtRepo, sessionsRepo, js)
	val := validator.New()

	authHandler := auth.NewHandler(authSvc, val)

	r := router.NewRouter(authHandler)

	srv := &http.Server{
		Addr:              appConfig.Host() + ":" + appConfig.Port(),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", appConfig.Port())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v\n", err)
		}
	}()

	// Setup graceful shutdown by listening for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Listen for Ctrl+C and termination signals
	<-quit                                               // Block until signal is received

	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("server stopped cleanly")
}
