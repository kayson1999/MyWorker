## ============================================
## 阶段 1：构建前端（Vite + Vue 3）
## ============================================
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# 接收构建参数：子路径前缀（默认 '/'，部署在子路径时传入如 '/Worker/'）
ARG APP_BASE_PATH=/
ENV APP_BASE_PATH=${APP_BASE_PATH}

# 先复制依赖描述文件，利用 Docker 缓存
COPY package.json package-lock.json ./
RUN npm ci

# 复制前端源码并构建
COPY index.html vite.config.js ./
COPY src/ ./src/
COPY public/ ./public/

RUN npm run build

## ============================================
## 阶段 2：构建 Go 后端
## ============================================
FROM golang:1.21-alpine AS backend-builder

# 设置 Go 模块代理为国内镜像，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 安装 CGO 依赖（sqlite3 需要）
RUN apk add --no-cache gcc musl-dev

# 复制 Go 模块文件并下载依赖
COPY server/go.mod server/go.sum ./
RUN go mod download

# 复制后端源码并构建
COPY server/ ./
RUN CGO_ENABLED=1 go build -o myworker-server .

## ============================================
## 阶段 3：运行（最小化镜像）
## ============================================
FROM alpine:3.19 AS production

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=backend-builder /app/myworker-server ./


# 从第一阶段复制前端构建产物
COPY --from=frontend-builder /app/dist ./dist/

# 创建数据目录和日志目录
RUN mkdir -p /app/db /app/logs

# 默认环境变量
ENV PORT=8008

EXPOSE 8008

# 启动后端服务
CMD ["./myworker-server"]
