# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 creates a static binary that works across platforms
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nut2mqtt .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/nut2mqtt .

# Expose port if your app needs it (adjust as needed)
# EXPOSE 8080

CMD ["./nut2mqtt"]
