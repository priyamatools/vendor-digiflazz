package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config menyimpan semua pengaturan aplikasi dari environment.
type Config struct {
	HTTPPort int
	Development bool

	DatabaseURL string

	DigiflazzUsername   string
	DigiflazzAPIKey     string
	DigiflazzCallbackURL string

	AtrilinksBaseURL          string
	AtrilinksUserID           string
	AtrilinksPIN              string
	AtrilinksPass             string
	AddonAtrilinksCallbackURL string

	// Rekonsiliasi pending (checktoday) saat Digiflazz replay & worker latar belakang.
	PendingWorkerEnabled   bool
	PendingWorkerInterval  time.Duration
	PendingWorkerBatchSize int
	PendingReconcileMinAge time.Duration
	PendingReconcileThrottle time.Duration
	ReconcileTimezone      string
}

// Load membaca .env dan memvalidasi variabel wajib.
func Load() (Config, error) {
	_ = godotenv.Load()

	port := 8080
	if v := os.Getenv("HTTP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("HTTP_PORT: %w", err)
		}
		port = p
	}

	cfg := Config{
		HTTPPort:                  port,
		Development:               isDevelopment(),
		DatabaseURL:               os.Getenv("DATABASE_URL"),
		DigiflazzUsername:         os.Getenv("DIGIFLAZZ_API_USERNAME"),
		DigiflazzAPIKey:           os.Getenv("DIGIFLAZZ_API_KEY"),
		DigiflazzCallbackURL:      os.Getenv("DIGIFLAZZ_CALLBACK_URL"),
		AtrilinksBaseURL:          envOr("ATRILINKS_BASE_URL", "http://api.atrilinks.com"),
		AtrilinksUserID:           os.Getenv("ATRILINKS_USER_ID"),
		AtrilinksPIN:              os.Getenv("ATRILINKS_PIN"),
		AtrilinksPass:             os.Getenv("ATRILINKS_PASS"),
		AddonAtrilinksCallbackURL: os.Getenv("ADDON_ATRILINKS_CALLBACK_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL wajib diisi")
	}
	if cfg.DigiflazzUsername == "" || cfg.DigiflazzAPIKey == "" {
		return Config{}, fmt.Errorf("DIGIFLAZZ_API_USERNAME dan DIGIFLAZZ_API_KEY wajib diisi")
	}
	if cfg.DigiflazzCallbackURL == "" {
		return Config{}, fmt.Errorf("DIGIFLAZZ_CALLBACK_URL wajib diisi")
	}
	if cfg.AtrilinksUserID == "" || cfg.AtrilinksPIN == "" {
		return Config{}, fmt.Errorf("ATRILINKS_USER_ID dan ATRILINKS_PIN wajib diisi (untuk BLAKE3 sign)")
	}
	// Password boleh kosong jika akun tidak memakai pass; tetap masuk ke string combined.
	if cfg.AddonAtrilinksCallbackURL == "" {
		return Config{}, fmt.Errorf("ADDON_ATRILINKS_CALLBACK_URL wajib diisi")
	}

	cfg.PendingWorkerEnabled = envBoolOr("PENDING_WORKER_ENABLED", true)
	cfg.PendingWorkerInterval = envDurationSecOr("PENDING_WORKER_INTERVAL_SEC", 120)
	cfg.PendingWorkerBatchSize = envIntOr("PENDING_WORKER_BATCH_SIZE", 50)
	cfg.PendingReconcileMinAge = envDurationSecOr("PENDING_RECONCILE_MIN_AGE_SEC", 30)
	cfg.PendingReconcileThrottle = envDurationSecOr("PENDING_RECONCILE_THROTTLE_SEC", 60)
	cfg.ReconcileTimezone = envOr("RECONCILE_TIMEZONE", "Asia/Jakarta")

	if cfg.PendingWorkerBatchSize < 1 {
		cfg.PendingWorkerBatchSize = 1
	}
	if cfg.PendingWorkerInterval < 10*time.Second {
		cfg.PendingWorkerInterval = 10 * time.Second
	}
	if cfg.PendingReconcileThrottle < 5*time.Second {
		cfg.PendingReconcileThrottle = 5 * time.Second
	}

	return cfg, nil
}

// envOr mengambil nilai env atau default jika kosong.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// envIntOr membaca integer dari env atau default.
func envIntOr(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// envBoolOr membaca boolean dari env (true/1/yes) atau default.
func envBoolOr(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return def
	}
}

// envDurationSecOr membaca durasi dalam detik dari env.
func envDurationSecOr(key string, defSec int) time.Duration {
	sec := envIntOr(key, defSec)
	if sec < 0 {
		sec = defSec
	}
	return time.Duration(sec) * time.Second
}

// isDevelopment true jika mode debug/logging pengembangan aktif.
func isDevelopment() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) {
	case "development", "dev", "local":
		return true
	}
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("DEV_MODE"))); v == "true" || v == "1" {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LOG_LEVEL")), "debug") {
		return true
	}
	return false
}
