package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"urlshortener/internal/storage"
)

// URLService contains the business logic: generate codes, validate URLs.
// It depends on the URLStorage interface — not on Postgres or Redis directly.
// This is dependency injection: the caller decides which storage to use.
type URLService struct {
	store storage.URLStorage
}

func NewURLService(s storage.URLStorage) *URLService {
	return &URLService{store: s}
}

// Shorten generates a random 6-character code and stores the mapping.
func (s *URLService) Shorten(ctx context.Context, longURL string) (string, error) {
	if longURL == "" {
		return "", fmt.Errorf("service.Shorten: url must not be empty")
	}
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}
	code := generateCode(6)

	if err := s.store.Save(ctx, code, longURL); err != nil {
		return "", fmt.Errorf("service.Shorten: %w", err)
	}

	return code, nil
}

// Get looks up the original URL for a short code.
func (s *URLService) Get(ctx context.Context, code string) (string, error) {
	url, err := s.store.Get(ctx, code)
	if err != nil {
		return "", fmt.Errorf("service.Get: %w", err)
	}
	return url, nil
}

// generateCode returns n random bytes as a hex string (2 hex chars per byte).
// 3 bytes → 6 hex chars, 62^6 ≈ 56 billion combinations.
func generateCode(n int) string {
	b := make([]byte, n/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
