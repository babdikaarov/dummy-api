# Stage 1: Development with Air for hot reload
FROM golang:1.24-alpine AS dev

WORKDIR /app

# Install Air for hot reload (v1.61.5 is the last version compatible with Go 1.24)
RUN go install github.com/air-verse/air@v1.61.5

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Air will be started via docker-compose command
CMD ["air", "-c", ".air.toml"]

# Stage 2: Builder for production
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ololo-gate ./cmd/main.go

# Stage 3: Production with minimal image
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copy the binary from builder
COPY --from=builder /app/ololo-gate /ololo-gate

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/ololo-gate"]
