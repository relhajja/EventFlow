FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build with executable permissions
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o eventflow-api . && chmod +x eventflow-api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder --chmod=755 /app/eventflow-api .

EXPOSE 8080

CMD ["./eventflow-api"]
