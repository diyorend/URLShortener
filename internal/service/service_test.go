package service

import (
	"context"
	"errors"
	"testing"
	"urlshortener/internal/storage"
)

func TestShorten_StoresURL(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	code, err := svc.Shorten(ctx, "https://archlinux.org")
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-char code, got %q (len %d)", code, len(code))
	}
}

func TestGet_ReturnsOriginalURL(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	longURL := "https://archlinux.org"
	code, _ := svc.Shorten(ctx, longURL)

	got, err := svc.Get(ctx, code)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != longURL {
		t.Errorf("expected %q, got %q", longURL, got)
	}
}

func TestGet_NotFound(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	_, err := svc.Get(ctx, "nosuchcode")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// errors.Is works through wrapped errors — service wraps storage errors with fmt.Errorf + %w
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestShorten_EmptyURL(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	_, err := svc.Shorten(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}
