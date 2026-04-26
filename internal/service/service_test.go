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

// The main bug this fixes: user types "google.com" without https://
// Without the fix, redirect goes to /google.com — a relative path, 404.
// With the fix, redirect goes to https://google.com — correct.
func TestShorten_AddsHttpsIfMissing(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	code, err := svc.Shorten(ctx, "google.com")
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	stored, _ := svc.Get(ctx, code)
	if stored != "https://google.com" {
		t.Errorf("expected https://google.com, got %q", stored)
	}
}

func TestShorten_PreservesExistingHttps(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	code, _ := svc.Shorten(ctx, "https://archlinux.org")
	stored, _ := svc.Get(ctx, code)
	if stored != "https://archlinux.org" {
		t.Errorf("expected https://archlinux.org unchanged, got %q", stored)
	}
}

func TestShorten_PreservesExistingHttp(t *testing.T) {
	fake := storage.NewFakeStorage()
	svc := NewURLService(fake)
	ctx := context.Background()

	code, _ := svc.Shorten(ctx, "http://example.com")
	stored, _ := svc.Get(ctx, code)
	if stored != "http://example.com" {
		t.Errorf("expected http://example.com unchanged, got %q", stored)
	}
}
