package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"urlshortener/internal/storage"

	"github.com/labstack/echo/v4"
)

// URLShortener is the interface our handler needs.
// The handler doesn't import the service package — it depends on the interface.
// This is important: interfaces belong to the caller, not the implementor.
type URLShortener interface {
	Shorten(ctx context.Context, longURL string) (string, error)
	Get(ctx context.Context, code string) (string, error)
}

// URLHandler holds our dependencies as interfaces.
type URLHandler struct {
	svc     URLShortener
	baseURL string
}

func NewURLHandler(svc URLShortener) *URLHandler {
	base := os.Getenv("BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return &URLHandler{svc: svc, baseURL: base}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

// Shorten handles POST /shorten
// It reads a JSON body, delegates to the service, and returns the short URL.
func (h *URLHandler) Shorten(c echo.Context) error {
	req := new(shortenRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
	}
	if req.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "url field is required")
	}

	// c.Request().Context() carries the request deadline set by Echo.
	// If the client disconnects, this context cancels — which cancels DB queries too.
	ctx := c.Request().Context()

	code, err := h.svc.Shorten(ctx, req.URL)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to shorten URL")
	}

	return c.JSON(http.StatusCreated, shortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", h.baseURL, code),
	})
}

// Redirect handles GET /:code
// It looks up the code and does a 301 redirect to the original URL.
func (h *URLHandler) Redirect(c echo.Context) error {
	code := c.Param("code")
	ctx := c.Request().Context()

	longURL, err := h.svc.Get(ctx, code)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "short URL not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "lookup failed")
	}

	// 301 = permanent redirect — browsers will cache it.
	// Use 302 if you might change destinations later.
	return c.Redirect(http.StatusMovedPermanently, longURL)
}
