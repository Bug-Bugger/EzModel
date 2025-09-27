# Frontend build stage
FROM node:20-alpine AS frontend-builder

# Install pnpm
RUN npm install -g pnpm

# Set working directory for frontend
WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install frontend dependencies
RUN pnpm install --frozen-lockfile

# Copy frontend source code
COPY frontend/ .

# Build frontend
RUN pnpm build

# Backend build stage
FROM golang:1.24.1-alpine AS backend-builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files from backend directory
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy backend source code
COPY backend/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ezmodel \
    cmd/api/main.go

# Final stage
FROM scratch

# Copy ca-certificates for HTTPS requests
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=backend-builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the backend binary
COPY --from=backend-builder /app/ezmodel /ezmodel

# Copy the frontend build
COPY --from=frontend-builder /app/frontend/build /static

# Expose port
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/ezmodel", "health"]

# Run the application
ENTRYPOINT ["/ezmodel"]