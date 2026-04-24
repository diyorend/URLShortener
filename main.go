package main

import (
	"context"
	"log"
	"urlshortener/internal/handler"
	"urlshortener/internal/service"
	"urlshortener/internal/storage"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()
	e := echo.New()

	// 1. Initialize Storage (Postgres)
	// Note: Use the connection string from docker-compose
	connStr := "postgres://user:password@localhost:5432/urlshortener?sslmode=disable"
	pgStore, err := storage.NewPostgresStorage(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize Service (Inject Storage)
	svc := service.NewURLService((pgStore))

	// 3. Initialize Handler (Inject Service)
	h := handler.NewURLHandler(svc)

	// 4. Routes
	e.POST("/shorten", h.Shorten)
	e.GET("/:code", h.Redirect)

	// 5. Start Server
	e.Logger.Fatal(e.Start(":8080"))
}
