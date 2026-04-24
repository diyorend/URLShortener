package service

import (
	"testing"
)

// MockStorage is a "fake" database for testing
type MockStorage struct {
	data map[string]string
}

func (m *MockStorage) Save(short, long string) error {
	m.data[short] = long
	return nil
}

func (m *MockStorage) Get(short string) (string, error) {
	return m.data[short], nil
}

func TestShorten(t *testing.T) {
	mock := &MockStorage{data: make(map[string]string)}
	svc := NewURLService(mock)

	longURL := "https://archlinux.org"
	code, err := svc.Shorten(longURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(code) != 6 { // hex.EncodeToString(6 bytes) = 12 chars
		t.Errorf("Expected code length 6, got %d", len(code))
	}

	savedURL, _ := mock.Get(code)
	if savedURL != longURL {
		t.Errorf("Expected %s, got %s", longURL, savedURL)
	}
}
