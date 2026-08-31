package config

import "testing"

func TestConfigValidateDefaults(t *testing.T) {
	c := Config{
		AppEnv:           "development",
		APIRatePerSecond: 2,
		APIRateBurst:     4,
		UserQuotaBytes:   10 * 1024 * 1024,
		MaxFileBytes:     10 * 1024 * 1024,
		MaxUploadBytes:   50 * 1024 * 1024,
		BlobStoreDriver:  "local",
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected defaults to validate: %v", err)
	}
}

func TestConfigValidateProductionSecrets(t *testing.T) {
	c := Config{
		AppEnv:           "production",
		DatabaseURL:      "postgres://db",
		RedisURL:         "redis://cache",
		SessionSecret:    "short",
		CSRFSecret:       "short",
		APIRatePerSecond: 2,
		APIRateBurst:     4,
		MaxFileBytes:     1,
		MaxUploadBytes:   1,
		BlobStoreDriver:  "local",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected short production secrets to fail validation")
	}
}
