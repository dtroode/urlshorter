FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /urlshorter ./cmd/shortener

# Use a smaller image for the final container
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache curl

# Copy the binary from builder
COPY --from=builder /urlshorter .

# Create directory for file storage
RUN mkdir -p /tmp

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./urlshorter"] 