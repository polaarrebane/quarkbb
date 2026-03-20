// Package main implements the QuarkBB bulletin board system server.
// This is the entry point for the application that sets up HTTP server,
// database connections, dependency injection, and graceful shutdown.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/ronia/quarkbb/internal/auth"
	"codeberg.org/ronia/quarkbb/internal/repository"
	"codeberg.org/ronia/quarkbb/internal/repository/sqlc"
	"codeberg.org/ronia/quarkbb/internal/router"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

// main is the application entry point. It initializes configuration,
// database connection, HTTP server with all routes, and handles graceful shutdown.
// The server supports user registration and authentication endpoints.
func main() {
	viper.SetConfigName("main")
	viper.AddConfigPath("./config/")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	dsn := viper.GetString("dsn")
	host := viper.GetString("host")
	port := viper.GetString("port")

	startCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
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

	sqlcQueries := sqlc.New(pool)
	authRepo := repository.NewUserRepository(sqlcQueries)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewAuthHandler(*authSvc)

	r := router.NewRouter(authHandler)

	srv := &http.Server{
		Addr:              host + ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
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
