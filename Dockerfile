# STAGE 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git/ca-certificates in case you have private modules or need SSL
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary with optimizations:
# CGO_ENABLED=0 makes it a static binary (works in scratch/alpine)
# -ldflags="-s -w" strips debug info to make it smaller
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o url-shortener .

# STAGE 2: Runner
FROM alpine:latest

WORKDIR /root/

# Copy only the compiled binary from the builder stage 
COPY --from=builder /app/url-shortener .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/public ./public

EXPOSE 8080
EXPOSE 50051

CMD ["./url-shortener"]

