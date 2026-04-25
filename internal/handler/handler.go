package handler

import (
	"fmt"
	"net/http"
	"os"
	"urlshortener/internal/service"

	"github.com/labstack/echo/v4"
)

type URLHandler struct {
	svc     *service.URLService
	baseURL string
}

func NewURLHandler(svc *service.URLService) *URLHandler {
	base := os.Getenv("BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}

	return &URLHandler{
		svc:     svc,
		baseURL: base,
	}
}

// Request struct for JSON binding
type shortenRequest struct {
	URL string `json:"url"`
}

func (h *URLHandler) Shorten(c echo.Context) error {
	req := new(shortenRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	code, err := h.svc.Shorten(req.URL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"short_url": fmt.Sprintf("%s/%s", h.baseURL, code)})
}

func (h *URLHandler) Redirect(c echo.Context) error {
	// 1. Get the code from the URL path parameter
	code := c.Param("code")

	// 2. Look it up in the storage
	longURL, err := h.svc.Get(code)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "URL not found"})
	}

	// 3. Redirect the user's browser to the original site
	return c.Redirect(http.StatusMovedPermanently, longURL)
}
