package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"urlshortener/internal/handler"
	"urlshortener/internal/metrics"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()

	// --- Configuration from environment ---
	// We read from env so the same binary works in Docker, Kubernetes, and locally.
	dbURL := envOrDefault("DATABASE_URL",
		"postgres://user:password@postgres:5432/urlshortener?sslmode=disable")
	redisURL := envOrDefault("REDIS_URL", "redis:6379")

	// --- Storage layer ---
	pgStore, err := storage.NewPostgresStorage(ctx, dbURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	rdbStore, err := storage.NewRedisStorage(ctx, redisURL)
	if err != nil {
		// Redis failure is non-fatal: the app still works without caching.
		// In production you'd alert on this but not crash.
		log.Printf("WARNING: redis unavailable, running without cache: %v", err)
	}

	var store storage.URLStorage
	if rdbStore != nil {
		store = storage.NewCachedStorage(pgStore, rdbStore)
	} else {
		store = pgStore
	}

	// --- Business logic layer ---
	svc := service.NewURLService(store)

	// --- Prometheus metrics ---
	// We create them here and pass them where needed.
	// Every company running Kubernetes scrapes /metrics — this is not optional.
	_ = metrics.New() // metrics register themselves via promauto

	// --- HTTP server ---
	e := echo.New()
	e.HideBanner = true

	// Middleware runs on every request, in order.
	e.Use(middleware.RequestID())     // adds X-Request-ID to every request/response
	e.Use(middleware.Recover())       // catches panics → 500 instead of crashing the server
	e.Use(middleware.RequestLogger()) // logs method, path, status, latency for every request

	// --- Routes ---
	h := handler.NewURLHandler(svc)

	e.Static("/", "public")
	e.File("/", "public/index.html")
	e.POST("/shorten", h.Shorten)
	e.GET("/:code", h.Redirect)

	// Prometheus scrape endpoint — monitoring systems poll this.
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Health check — used by Docker/k8s to know when the app is ready.
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// --- Graceful shutdown ---
	// We start the server in a goroutine so we can block on the signal below.
	go func() {
		port := envOrDefault("PORT", "8080")
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Block until SIGINT (Ctrl+C) or SIGTERM (docker stop / k8s pod termination).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("received signal %v — shutting down gracefully", sig)

	// Give active requests up to 10 seconds to finish.
	// After that, force-close remaining connections.
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
