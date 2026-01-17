# Build stage for Go
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o dumper ./cmd/dumper

# Build stage for Mini App
FROM oven/bun:1 AS frontend

WORKDIR /app
COPY mini-app/package.json mini-app/bun.lock ./
RUN bun install --frozen-lockfile
COPY mini-app/ .
RUN bun run build

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/dumper .
COPY --from=frontend /app/dist ./mini-app/dist

RUN mkdir -p /app/data

EXPOSE 8080

ENV DATA_DIR=/app/data
ENV HTTP_PORT=8080

CMD ["./dumper"]
