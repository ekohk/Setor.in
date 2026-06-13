# Multi-stage build for Setor.in backend (Go 1.24)
FROM golang:1.24-bookworm AS builder
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN apt-get update \
	&& apt-get install -y --no-install-recommends git ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . ./
RUN go build -ldflags "-s -w" -o /app/bin/api ./cmd/api

FROM gcr.io/distroless/static-debian12 AS runtime
COPY --from=builder /app/bin/api /usr/local/bin/api
# Copy CA certs from builder so TLS works in a minimal runtime image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/api"]