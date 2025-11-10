# Function Builder Worker
# Builds Docker images from user source code via NATS events

FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install Docker CLI and buildkit
RUN apk add --no-cache git docker-cli

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build the worker binary
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates docker-cli

WORKDIR /app

COPY --from=builder /app/worker .

CMD ["./worker"]
