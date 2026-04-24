package service

import (
	"crypto/rand"
	"encoding/hex"
	"urlshortener/internal/storage"
)

type URLService struct {
	store storage.URLStorage
}

func NewURLService(s storage.URLStorage) *URLService {
	return &URLService{store: s}
}

func (s *URLService) Shorten(longURL string) (string, error) {
	// 1. Generate a random short code
	code := generateCode(6)

	// 2. Save it using the interface
	err := s.store.Save(code, longURL)
	if err != nil {
		return "", err
	}

	return code, nil
}

func generateCode(n int) string {
	b := make([]byte, n/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}


