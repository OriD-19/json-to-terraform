# Stage 1: build (Alpine for lightweight, fast build)
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build

# Copy module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary for Alpine; no CGO
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o /parser ./cmd/parser

# Stage 2: minimal runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /parser /usr/local/bin/parser

ENTRYPOINT ["/usr/local/bin/parser"]
CMD ["--help"]
