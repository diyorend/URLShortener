package service

import (
	"testing"
)

// MockStorage is a "fake" database for testing
type MockStorage struct {
	data   map[string]string
	clicks map[string]int
}

func (m *MockStorage) Save(short, long string) error {
	m.data[short] = long
	return nil
}

func (m *MockStorage) Get(short string) (string, error) {
	return m.data[short], nil
}

func (m *MockStorage) IncrementClicks(short string) error {
	if m.clicks == nil {
		m.clicks = make(map[string]int)
	}
	m.clicks[short]++
	return nil
}

func TestShortenAndRedirect(t *testing.T) {
	mock := &MockStorage{
		data:   make(map[string]string),
		clicks: make(map[string]int),
	}
	svc := NewURLService(mock)

	longURL := "https://archlinux.org"
	code, _ := svc.Shorten(longURL)

	// Test Retrieval
	savedURL, err := svc.Get(code)
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}

	if savedURL != longURL {
		t.Errorf("Expected %s, got %s", longURL, savedURL)
	}

	if mock.clicks[code] != 1 {
		t.Errorf("Expected 1 click, got %d", mock.clicks[code])
	}
}
