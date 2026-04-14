## ============================================
## Stage 1: Build frontend (Vite + Vue 3)
## ============================================

FROM node:18-alpine AS frontend-builder

# 使用更快的 Alpine 镜像源
RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories

WORKDIR /app

# Accept build argument: subpath prefix (default /, deploy to subpath like /Worker/)
ARG APP_BASE_PATH=/
ENV APP_BASE_PATH=${APP_BASE_PATH}

# Copy dependency files first to leverage Docker cache
COPY package.json package-lock.json ./
RUN npm config set registry https://registry.npmmirror.com \
    && npm ci --no-audit --no-fund

# Copy frontend source code and build
COPY index.html vite.config.js ./
COPY src/ ./src/
COPY public/ ./public/
RUN npm run build

## ============================================
## Stage 2: Build Go backend
## ============================================

FROM golang:1.21 AS backend-builder

# Configure Go proxy for faster dependency downloads
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# Copy Go module files and download dependencies
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy backend source code and build
COPY server/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o myworker-server .

## ============================================
## Stage 3: Runtime (minimal image)
## ============================================

FROM debian:bookworm-slim AS production

WORKDIR /app

# Install runtime dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

# Set timezone
ENV TZ=Asia/Shanghai

# Copy binary file from build stage
COPY --from=backend-builder /app/myworker-server ./

# Copy frontend build artifacts from first stage
COPY --from=frontend-builder /app/dist ./dist/

# Create data directory and logs directory
RUN mkdir -p /app/data /app/logs

# Default environment variables
ENV PORT=8008

EXPOSE 8008

# Start backend service
CMD ["./myworker-server"]
