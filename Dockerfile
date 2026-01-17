# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o dumper ./cmd/dumper

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/dumper .

# Create directories for data and mini-app
RUN mkdir -p /app/data /app/mini-app/dist

EXPOSE 8080

ENV DATA_DIR=/app/data
ENV HTTP_PORT=8080

CMD ["./dumper"]
