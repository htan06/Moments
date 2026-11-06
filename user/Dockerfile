
# Stage 1: Build
FROM golang:1.26.2-bookworm AS builder

RUN apt-get update

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV GOOS=linux
ENV GOARCH=amd64

RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/user ./cmd

# Stage 2: Runtime
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        libvips42 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/user ./user

CMD ["./user"]
