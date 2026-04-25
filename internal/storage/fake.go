package storage

import (
	"context"
	"sync"
)

// FakeStorage is an in-memory implementation of URLStorage for tests.
// Rule: if go test ./... requires Docker or a real database, your design is wrong.
// The FakeStorage lets unit tests run with zero infrastructure.
type FakeStorage struct {
	mu     sync.Mutex
	data   map[string]string
	clicks map[string]int
}

func NewFakeStorage() *FakeStorage {
	return &FakeStorage{
		data:   make(map[string]string),
		clicks: make(map[string]int),
	}
}

func (f *FakeStorage) Save(_ context.Context, short, long string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[short] = long
	return nil
}

func (f *FakeStorage) Get(_ context.Context, short string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[short]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (f *FakeStorage) IncrementClicks(_ context.Context, short string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.clicks[short]++
	return nil
}

// ClicksFor returns how many times a short code was accessed.
// Test-only helper — not part of the URLStorage interface.
func (f *FakeStorage) ClicksFor(short string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.clicks[short]
}
