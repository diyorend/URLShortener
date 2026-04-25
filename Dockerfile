# Stage 1: Builder
# We use a full Go image to compile the binary.
# This image has the Go compiler — it's ~800MB.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go mod download (some modules need it)
RUN apk add --no-cache git ca-certificates

# Copy dependency files first.
# Docker caches each layer. If go.mod/go.sum haven't changed,
# Docker reuses the cached "go mod download" layer — faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a static binary.
# CGO_ENABLED=0: no C dependencies — the binary runs in Alpine (no glibc needed)
# -ldflags="-s -w": strip debug symbols — ~30% smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o url-shortener .

# Stage 2: Runner
# We use a tiny Alpine image — no Go compiler, no source code.
# Final image size: ~20MB vs ~800MB for the builder.
FROM alpine:latest

WORKDIR /app

# Copy ONLY the compiled binary and needed assets from the builder stage
COPY --from=builder /app/url-shortener .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/public ./public

EXPOSE 8080

CMD ["./url-shortener"]
