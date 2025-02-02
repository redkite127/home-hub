# Build stage
FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# Copy only the files needed for downloading dependencies first
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached)
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o home-hub .

# Final stage
FROM alpine:3

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /app/home-hub .

# Use nobody user for better security
USER nobody

# Command to run
CMD ["./home-hub"]
