package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

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
}

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
		return Config{}, fmt.Errorf("ATRILINKS_USER_ID dan ATRILINKS_PIN wajib diisi")
	}
	if cfg.AddonAtrilinksCallbackURL == "" {
		return Config{}, fmt.Errorf("ADDON_ATRILINKS_CALLBACK_URL wajib diisi")
	}

	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

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
