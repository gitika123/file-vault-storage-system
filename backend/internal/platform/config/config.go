package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv             string
	HTTPAddr           string
	DatabaseURL        string
	RedisURL           string
	FrontendOrigin     string
	SessionSecret      string
	CSRFSecret         string
	CookieSecure       bool
	APIRatePerSecond   int
	APIRateBurst       int
	UserQuotaBytes     int64
	MaxFileBytes       int64
	MaxUploadBytes     int64
	BlobStoreDriver    string
	BlobStorePath      string
	BlobStoreBucket    string
	BlobStoreRegion    string
	BlobStoreEndpoint  string
	BlobStoreAccessKey string
	BlobStoreSecretKey string
	PublicShareBaseURL string
	LogLevel           string
}

func Load() (Config, error) {
	c := Config{
		AppEnv:             envOr("APP_ENV", "development"),
		HTTPAddr:           envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisURL:           os.Getenv("REDIS_URL"),
		FrontendOrigin:     envOr("FRONTEND_ORIGIN", "http://localhost:5173"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		CSRFSecret:         os.Getenv("CSRF_SECRET"),
		BlobStoreDriver:    envOr("BLOB_STORE_DRIVER", "local"),
		BlobStorePath:      envOr("BLOB_STORE_PATH", "./data/blobs"),
		BlobStoreBucket:    os.Getenv("BLOB_STORE_BUCKET"),
		BlobStoreRegion:    envOr("BLOB_STORE_REGION", "us-east-1"),
		BlobStoreEndpoint:  os.Getenv("BLOB_STORE_ENDPOINT"),
		BlobStoreAccessKey: os.Getenv("BLOB_STORE_ACCESS_KEY"),
		BlobStoreSecretKey: os.Getenv("BLOB_STORE_SECRET_KEY"),
		PublicShareBaseURL: envOr("PUBLIC_SHARE_BASE_URL", "http://localhost:5173/share"),
		LogLevel:           envOr("LOG_LEVEL", "info"),
	}

	var err error
	if c.CookieSecure, err = boolEnv("COOKIE_SECURE", false); err != nil {
		return Config{}, err
	}
	if c.APIRatePerSecond, err = intEnv("API_RATE_PER_SECOND", 2); err != nil {
		return Config{}, err
	}
	if c.APIRateBurst, err = intEnv("API_RATE_BURST", 4); err != nil {
		return Config{}, err
	}
	if c.UserQuotaBytes, err = int64Env("USER_QUOTA_BYTES", 10*1024*1024); err != nil {
		return Config{}, err
	}
	if c.MaxFileBytes, err = int64Env("MAX_FILE_BYTES", 10*1024*1024); err != nil {
		return Config{}, err
	}
	if c.MaxUploadBytes, err = int64Env("MAX_UPLOAD_BYTES", 50*1024*1024); err != nil {
		return Config{}, err
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c Config) Validate() error {
	if c.AppEnv == "production" {
		for name, value := range map[string]string{
			"DATABASE_URL":   c.DatabaseURL,
			"REDIS_URL":      c.RedisURL,
			"SESSION_SECRET": c.SessionSecret,
			"CSRF_SECRET":    c.CSRFSecret,
		} {
			if len(value) < 32 && (name == "SESSION_SECRET" || name == "CSRF_SECRET") {
				return fmt.Errorf("%s must be at least 32 characters in production", name)
			}
			if value == "" {
				return fmt.Errorf("%s is required in production", name)
			}
		}
	}
	if c.APIRatePerSecond <= 0 || c.APIRateBurst <= 0 {
		return errors.New("API rate settings must be positive")
	}
	if c.UserQuotaBytes < 0 || c.MaxFileBytes <= 0 || c.MaxUploadBytes < c.MaxFileBytes {
		return errors.New("quota and upload size settings are invalid")
	}
	if c.BlobStoreDriver != "local" && c.BlobStoreDriver != "s3" {
		return fmt.Errorf("unsupported BLOB_STORE_DRIVER %q", c.BlobStoreDriver)
	}
	if c.BlobStoreDriver == "s3" && c.BlobStoreBucket == "" {
		return errors.New("BLOB_STORE_BUCKET is required when BLOB_STORE_DRIVER=s3")
	}
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func boolEnv(name string, fallback bool) (bool, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be boolean: %w", name, err)
	}
	return parsed, nil
}

func intEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}

func int64Env(name string, fallback int64) (int64, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return parsed, nil
}
