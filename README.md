# URL Shortener

A production-grade URL shortener built with Go, demonstrating the core patterns used in mid-level backend engineering.

## What This Teaches

| Concept                           | Where It Appears                                           |
| --------------------------------- | ---------------------------------------------------------- |
| Interfaces & dependency injection | `storage.URLStorage`, `handler.URLShortener`               |
| Context propagation               | Every function signature: `func(ctx context.Context, ...)` |
| Connection pool                   | `pgxpool.Pool` in `postgres.go`                            |
| Cache-aside pattern               | `CachedStorage.Get()` in `storage.go`                      |
| Sentinel errors                   | `storage.ErrNotFound`, `errors.Is()` in handler            |
| Goroutines (fire & forget)        | Click counter increment after cache hit                    |
| Graceful shutdown                 | `signal.Notify` + `e.Shutdown()` in `main.go`              |
| Prometheus metrics                | `/metrics` endpoint, counter + histogram                   |
| Multi-stage Docker build          | `Dockerfile`                                               |
| Health checks                     | `/health` endpoint, `docker-compose` healthcheck           |
| Unit tests (no Docker)            | `service_test.go` with `FakeStorage`                       |
| GitHub Actions CI                 | `.github/workflows/ci.yml`                                 |

## Architecture

```
HTTP Request
     │
     ▼
  Echo Router
     │
     ▼
  URLHandler          ← depends on URLShortener interface
     │
     ▼
  URLService          ← business logic: generate codes, validate
     │
     ▼
  CachedStorage       ← composite: Redis on top, Postgres underneath
   /        \
Redis       Postgres   ← both implement URLStorage interface
```

The key insight: each layer only knows about the **interface** of the layer below it. The handler doesn't import the `service` package — it uses the `URLShortener` interface. This means you can test each layer in total isolation.

## Quick Start

**Requirements:** Docker and Docker Compose.

```bash
# 1. Clone the project
git clone <your-repo>
cd urlshortener

# 2. Start everything
docker-compose up --build
```

Open http://localhost:8080 in your browser.

## Running Tests

```bash
# These require no Docker — they use FakeStorage
go test ./... -race -v
```

The `-race` flag enables Go's data race detector. It catches concurrent access bugs that only appear under load. Always run this in CI.

## API

### Shorten a URL

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://archlinux.org"}'
```

Response:

```json
{ "short_url": "http://localhost:8080/a3f9c2" }
```

### Follow a Short URL

```bash
curl -L http://localhost:8080/a3f9c2
# → redirects to https://archlinux.org
```

### Health Check

```bash
curl http://localhost:8080/health
# → {"status": "ok"}
```

### Prometheus Metrics

```bash
curl http://localhost:8080/metrics
# → Prometheus text format with urlshortener_* metrics
```

## Project Structure

```
urlshortener/
├── main.go                         # Entry point: wires everything together
├── Dockerfile                      # Multi-stage build
├── docker-compose.yml              # Full stack: app + postgres + redis
├── Makefile                        # Common commands
├── migrations/
│   └── 001_init.sql               # Database schema
├── internal/
│   ├── handler/
│   │   └── handler.go             # HTTP handlers (depends on interface)
│   ├── service/
│   │   ├── service.go             # Business logic
│   │   └── service_test.go        # Unit tests (no Docker required)
│   ├── storage/
│   │   ├── storage.go             # URLStorage interface + CachedStorage
│   │   ├── postgres.go            # Postgres implementation
│   │   ├── redis.go               # Redis implementation
│   │   └── fake.go                # In-memory fake for tests
│   └── metrics/
│       └── metrics.go             # Prometheus metric definitions
├── public/
│   └── index.html                 # Simple frontend
├── api/proto/
│   └── url.proto                  # gRPC service definition (future use)
├── .env.example                   # Configuration template
├── .gitignore
└── .github/workflows/
    └── ci.yml                     # GitHub Actions: test + build on every push
```

## Key Design Decisions

### Why interfaces everywhere?

```go
// handler.go depends on this interface — not on *service.URLService
type URLShortener interface {
    Shorten(ctx context.Context, longURL string) (string, error)
    Get(ctx context.Context, code string) (string, error)
}
```

Because `go test ./...` runs in milliseconds with no infrastructure. If the handler imported `*service.URLService` directly, you'd need a real service. If the service imported `*PostgresStorage`, you'd need a real database. Interfaces break these dependencies.

### Why context as first argument?

```go
func (s *URLService) Shorten(ctx context.Context, longURL string) (string, error)
```

Context is a cancellation signal. When a user closes their browser mid-request, the context cancels. Without it, the database query keeps running, the goroutine keeps running, and under load you run out of resources. With context, the query stops when the client disconnects.

### Why CachedStorage?

The service doesn't know whether it's talking to Redis or Postgres. CachedStorage is transparent to everything above it. This pattern is called the **decorator pattern**: wrapping an implementation to add behavior (caching) without changing the interface.

### Why pgxpool instead of pgx.Conn?

A single connection serializes all requests — only one query runs at a time. A pool maintains multiple connections and serves them concurrently. Under any real load, a single connection is the bottleneck.

### Why graceful shutdown?

Without it: `Ctrl+C` or `docker stop` kills the process instantly. Active HTTP requests return connection errors. Users see failures. With graceful shutdown, the server stops accepting new connections and waits up to 10 seconds for active requests to finish cleanly.
