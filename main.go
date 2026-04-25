package main

import (
	"context"
	"os"
	"urlshortener/internal/handler"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()
	e := echo.New()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@postgres:5432/urlshortener?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis:6379"
	}
	pgStore, err := storage.NewPostgresStorage(ctx, dbURL)
	if err != nil {
		e.Logger.Fatalf("Failed to connect to Postgres: %v", err)
	}
	rdbStore, _ := storage.NewRedisStorage(ctx, redisURL)

	combinedStore := storage.NewCachedStorage(pgStore, rdbStore)

	svc := service.NewURLService((combinedStore))

	h := handler.NewURLHandler(svc)

	e.Static("/", "public")
	e.File("/", "public/index.html")

	e.POST("/shorten", h.Shorten)
	e.GET("/:code", h.Redirect)

	e.Logger.Fatal(e.Start(":8080"))
}
