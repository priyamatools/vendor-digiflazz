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
	"github.com/addon-vendor-digiflazz/internal/database"
	"github.com/addon-vendor-digiflazz/internal/digiflazz"
	"github.com/addon-vendor-digiflazz/internal/handler"
	"github.com/addon-vendor-digiflazz/internal/migrate"
	"github.com/addon-vendor-digiflazz/internal/repository"
	"github.com/addon-vendor-digiflazz/internal/service"
	"github.com/addon-vendor-digiflazz/internal/worker"
)

// main memuat konfigurasi, migrasi DB, dan menjalankan server HTTP.
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	if cfg.Development {
		logLevel = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	if cfg.Development {
		log.Info("development mode enabled", "http_debug", true)
	}

	ctx := context.Background()
	pool, err := database.ConnectPool(ctx, cfg.DatabaseURL, log)
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
		cfg.AtrilinksPass,
		cfg.AtrilinksPIN,
		cfg.AddonAtrilinksCallbackURL,
		log,
		cfg.Development,
	)
	webhook := digiflazz.NewWebhookClient(cfg.DigiflazzCallbackURL, log, cfg.Development)
	svc := service.NewTopupService(cfg, repo, atClient, webhook, log)

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	go worker.RunPendingReconcile(workerCtx, cfg, svc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	dfHandler := handler.NewDigiflazzHandler(svc, log, cfg.Development)
	atCallback := handler.NewAtrilinksCallbackHandler(svc, log, cfg.Development)
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

	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

// projectRoot menemukan direktori yang berisi folder migrations.
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
