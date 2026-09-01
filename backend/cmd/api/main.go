package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/balkanid/file-vault/internal/auth"
	"github.com/balkanid/file-vault/internal/files"
	"github.com/balkanid/file-vault/internal/platform/config"
	"github.com/balkanid/file-vault/internal/platform/database"
	"github.com/balkanid/file-vault/internal/platform/health"
	"github.com/balkanid/file-vault/internal/platform/ratelimit"
	"github.com/balkanid/file-vault/internal/storage"
	"github.com/balkanid/file-vault/internal/upload"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := database.RunMigrations(ctx, db); err != nil {
		logger.Error("database migrations failed", "error", err)
		os.Exit(1)
	}

	healthHandler := health.Handler{DB: db}
	limiter := interface{ Allow(string) bool }(ratelimit.New(cfg.APIRatePerSecond, cfg.APIRateBurst))
	if cfg.RedisURL != "" {
		parsed, err := url.Parse(cfg.RedisURL)
		if err != nil {
			logger.Error("invalid Redis URL", "error", err)
			os.Exit(1)
		}
		options, err := redis.ParseURL(parsed.String())
		if err != nil {
			logger.Error("invalid Redis configuration", "error", err)
			os.Exit(1)
		}
		client := redis.NewClient(options)
		if err := client.Ping(ctx).Err(); err != nil {
			logger.Error("Redis unavailable", "error", err)
			os.Exit(1)
		}
		limiter = ratelimit.NewRedis(client, cfg.APIRatePerSecond, cfg.APIRateBurst)
	}
	authHandler := auth.HTTPHandler{Service: auth.Service{DB: db, CookieSecure: cfg.CookieSecure}, CookieSecure: cfg.CookieSecure, RateLimiter: limiter}
	var blobStore storage.Store = storage.LocalStore{Root: cfg.BlobStorePath}
	if cfg.BlobStoreDriver == "s3" {
		blobStore, err = storage.NewS3Store(ctx, cfg.BlobStoreBucket, cfg.BlobStoreRegion, cfg.BlobStoreEndpoint, cfg.BlobStoreAccessKey, cfg.BlobStoreSecretKey, cfg.BlobStorePath)
		if err != nil {
			logger.Error("object storage initialization failed", "error", err)
			os.Exit(1)
		}
	}
	uploadHandler := upload.HTTPHandler{Service: upload.Service{DB: db, Store: blobStore}, MaxFileBytes: cfg.MaxFileBytes, MaxUploadBytes: cfg.MaxUploadBytes, MaxFiles: 10}
	fileHandler := files.HTTPHandler{Service: files.Service{DB: db, Store: blobStore, Events: files.NewDownloadEventHub()}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", healthHandler.Live)
	mux.HandleFunc("GET /health/ready", healthHandler.Ready)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.Handle("POST /api/auth/logout", authHandler.RequireSession(true, http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /api/auth/me", authHandler.RequireSession(false, http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /api/auth/csrf", authHandler.RequireSession(false, http.HandlerFunc(authHandler.CSRF)))
	mux.Handle("POST /api/uploads", authHandler.RequireSession(true, http.HandlerFunc(uploadHandler.Upload)))
	mux.Handle("GET /api/stats/storage", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.Stats)))
	mux.Handle("GET /api/admin/stats", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.AdminStats)))
	mux.Handle("GET /api/admin/files", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.AdminFiles)))
	mux.Handle("GET /api/events/downloads", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.DownloadEvents)))
	mux.Handle("GET /api/files", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.List)))
	mux.Handle("GET /api/files/{id}", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.Detail)))
	mux.Handle("DELETE /api/files/{id}", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.Delete)))
	mux.Handle("PATCH /api/files/{id}", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.Rename)))
	mux.Handle("PATCH /api/files/{id}/folder", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.Move)))
	mux.Handle("PUT /api/files/{id}/tags", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.Tags)))
	mux.Handle("GET /api/folders", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.Folders)))
	mux.Handle("POST /api/folders", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.CreateFolder)))
	mux.Handle("PATCH /api/folders/{id}", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.RenameFolder)))
	mux.Handle("DELETE /api/folders/{id}", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.DeleteFolder)))
	mux.Handle("POST /api/shares/public", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.CreatePublicShare)))
	mux.Handle("DELETE /api/files/{id}/share", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.RevokePublicShare)))
	mux.Handle("DELETE /api/folders/{id}/share", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.RevokePublicFolderShare)))
	mux.Handle("POST /api/shares/direct", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.CreateDirectShare)))
	mux.Handle("GET /api/shares/{id}", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.ShareAccess)))
	mux.Handle("DELETE /api/shares/direct/{id}", authHandler.RequireSession(true, http.HandlerFunc(fileHandler.RevokeDirectShare)))
	mux.Handle("GET /api/files/{id}/content", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.Content)))
	mux.Handle("GET /api/files/{id}/preview", authHandler.RequireSession(false, http.HandlerFunc(fileHandler.Preview)))
	mux.Handle("GET /public/{token}/download", http.HandlerFunc(fileHandler.PublicContent))
	mux.Handle("GET /public/{token}", http.HandlerFunc(fileHandler.PublicLanding))

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
