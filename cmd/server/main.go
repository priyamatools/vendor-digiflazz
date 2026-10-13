package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
	"github.com/addon-vendor-digiflazz/internal/config"
	"github.com/addon-vendor-digiflazz/internal/digiflazz"
	"github.com/addon-vendor-digiflazz/internal/handler"
	"github.com/addon-vendor-digiflazz/internal/migrate"
	"github.com/addon-vendor-digiflazz/internal/repository"
	"github.com/addon-vendor-digiflazz/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrationsDir := filepath.Join(projectRoot(), "migrations")
	if err := migrate.Up(ctx, pool, migrationsDir); err != nil {
		log.Error("migrate", "err", err)
		os.Exit(1)
	}

	repo := repository.New(pool)
	atClient := atrilinks.NewClient(
		cfg.AtrilinksBaseURL,
		cfg.AtrilinksUserID,
		cfg.AtrilinksPIN,
		cfg.AtrilinksPass,
		cfg.AddonAtrilinksCallbackURL,
	)
	webhook := digiflazz.NewWebhookClient(cfg.DigiflazzCallbackURL)
	svc := service.NewTopupService(cfg, repo, atClient, webhook, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	dfHandler := handler.NewDigiflazzHandler(svc)
	atCallback := handler.NewAtrilinksCallbackHandler(svc)
	mux.HandleFunc("/api/digiflazz/topup", dfHandler.Topup)
	mux.HandleFunc("/h2h/atrilinks/callback", atCallback.Callback)

	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func projectRoot() string {
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "migrations")); err == nil {
			return wd
		}
		if _, err := os.Stat(filepath.Join(wd, "..", "..", "migrations")); err == nil {
			return filepath.Join(wd, "..", "..")
		}
	}
	return "."
}
