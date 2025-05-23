FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN make build

# Create a smaller final image
FROM alpine:3.21.3

# Install runtime dependencies
RUN apk add --no-cache ca-certificates openssh-client ansible

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/talis /app/talis

# Copy necessary files for runtime
COPY --from=builder /app/ansible /app/ansible
COPY --from=builder /app/.env.example /app/.env

# Expose the application port
EXPOSE 8080

# Set environment variables
ENV SERVER_PORT=8080

# Run the application
CMD ["/app/talis"] 
