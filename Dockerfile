FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files (if you have them)
# COPY go.mod go.sum ./
# RUN go mod download

# Copy source code
COPY server/ ./

RUN go mod init mail-notification
RUN go mod tidy

# Build the application
RUN go build -o mail-notification-server .

# Create a lightweight production image
FROM alpine:latest

WORKDIR /app

# Install timezone data
RUN apk add --no-cache tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/mail-notification-server .

# Expose the default port
EXPOSE 3333

# Run the server
CMD ["./mail-notification-server"]