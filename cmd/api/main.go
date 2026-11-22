package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaksimKorneplod/MyInvestor/internal/config"
	"github.com/MaksimKorneplod/MyInvestor/internal/database"
	"github.com/MaksimKorneplod/MyInvestor/internal/handler"
	"github.com/MaksimKorneplod/MyInvestor/internal/repository"
	"github.com/MaksimKorneplod/MyInvestor/internal/repository/rss"
	"github.com/MaksimKorneplod/MyInvestor/internal/service"
)

func serverStart(server *http.Server) error {
	slog.Info("Starting server", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Connecting to the database")
	dbPool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to the database successfully")
	defer dbPool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", handler.HealthHandler)
	mux.HandleFunc("POST /api/v1/schools/register", handler.SchoolRegistration(service.Registration{Store: repository.SchoolStore{Pool: dbPool}}))
	mux.HandleFunc("POST /api/v1/schools/login", handler.SchoolLogin(service.Login{Store: repository.SchoolStore{Pool: dbPool}}))
	feedURL := os.Getenv("RSS_FEED_URL")
	if feedURL == "" {
		feedURL = rss.DefaultFeedURL
	}
	newsClient := rss.NewClient(&http.Client{Timeout: 10 * time.Second}, feedURL)
	mux.HandleFunc("GET /api/v1/news", handler.NewsHandler(newsClient))

	port := cfg.HTTPAddr

	server := &http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- serverStart(server)
	}()

	select {
	case <-ctx.Done():
		slog.Info("Received shutdown signal")
		slog.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown failed", "error", err)
			_ = server.Close()
		}
		if err := <-errCh; err != nil {
			slog.Error("Server stopped with error", "error", err)
		}
		slog.Info("Server shutdown complete")
	case err := <-errCh:
		if err != nil {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}

}
