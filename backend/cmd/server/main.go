package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"activist-base/src/config"
	pgstore "activist-base/src/repository/postgres"
	transporthttp "activist-base/src/transport/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "server startup failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := pgstore.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer store.Close()

	handler := transporthttp.NewRouter(transporthttp.Deps{
		Config: cfg,
		Logger: logger,
		Store:  store,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting http server", "addr", cfg.HTTPAddr)
	serveErrCh := make(chan error, 1)
	go func() {
		serveErrCh <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErrCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("listen and serve: %w", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		if err := <-serveErrCh; err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("wait server shutdown: %w", err)
		}
	}

	return nil
}

func parseLogLevel(raw string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return slog.LevelInfo
	}
	return level
}
