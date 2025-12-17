# ========================================
# Stage 1: Build Frontend with Node
# ========================================
FROM node:22-alpine3.22 AS frontend-builder

# Set work directory
WORKDIR /app/web

# Copy frontend source
COPY web/package*.json ./
RUN npm install --legacy-peer-deps

COPY web/ ./
RUN npm run build

# ========================================
# Stage 2: Build Backend with Go
# ========================================
FROM golang:1.24-alpine3.22 AS backend-builder

# Install dependencies
# RUN apk add --no-cache git

# debug
RUN apk add --no-cache git ca-certificates openssl && update-ca-certificates

WORKDIR /app

# Copy go mod and sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# debug
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go mod download

# Copy backend source code
COPY . .

# Copy built frontend into embed directory
# (your Go embed path: all:web/dist)
COPY --from=frontend-builder /app/web/dist ./web/dist

# debug
# COPY routes.json ./
# COPY .env ./

# Build binary
RUN go build -o /app/bin/proxify .

# ========================================
# Stage 3: Minimal Runtime Image
# ========================================
FROM alpine:3.20

WORKDIR /app

# debug
RUN apk add --no-cache tzdata

# Copy built binary
COPY --from=backend-builder /app/bin/proxify ./

# debug
# RUN mkdir -p /app/log && chown -R proxify:proxify /app

# Create non-root user
# RUN adduser -D proxify
# USER proxify

# debug
ENV TZ=Asia/Shanghai

# Expose port (same as your main.go)
EXPOSE 8080

# Run the app
ENTRYPOINT ["./proxify"]