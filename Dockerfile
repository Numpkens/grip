# STEP 1: Build the Binary (The "Builder" Stage)
FROM golang:1.25.2-alpine AS builder

# Install git and ca-certificates in case they're needed for fetching modules
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency files first to leverage Docker's layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project structure
COPY . .

# Build the Web UI binary (targeting your cmd/grip-web directory)
# CGO_ENABLED=0 ensures a static binary that works on any Linux distro
RUN CGO_ENABLED=0 GOOS=linux go build -o grip-server ./cmd/grip-web/main.go

# STEP 2: The Final Image (The "Production" Stage)
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/grip-server .

# IMPORTANT: Copy your static assets and templates
# Your Go code expects these to be in these relative paths
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose the port your web server listens on
EXPOSE 8080

# Run the web server by default
CMD ["./grip-server"]