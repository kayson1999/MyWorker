#!/bin/bash
# ============================================
# 打工人打卡（MyWorker）— 一键启动脚本
# ============================================

set -e

# 项目根目录（脚本所在目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # 无颜色

# 打印带颜色的信息
info()  { echo -e "${CYAN}[INFO]${NC}  $1"; }
ok()    { echo -e "${GREEN}[OK]${NC}    $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 使用说明
usage() {
    echo ""
    echo "👷 打工人打卡（MyWorker）启动脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  start       启动本地开发环境（前端 + Go 后端）"
    echo "  start-prod  本地生产模式（构建前端 + 编译并启动 Go 后端）"
    echo "  docker      Docker Compose 构建并启动"
    echo "  docker-stop Docker Compose 停止服务"
    echo "  docker-logs Docker Compose 查看日志"
    echo "  install     安装前端依赖 + 下载 Go 模块"
    echo "  build       构建前端 + 编译 Go 后端"
    echo "  help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 start          # 本地开发"
    echo "  $0 docker         # Docker 一键部署"
    echo ""
}

# 检查命令是否存在
check_cmd() {
    if ! command -v "$1" &> /dev/null; then
        error "$1 未安装，请先安装后重试"
        exit 1
    fi
}

# 安装依赖
do_install() {
    info "安装前端依赖..."
    npm install
    ok "前端依赖安装完成"

    info "下载 Go 模块依赖..."
    (cd server && go mod tidy)
    ok "Go 模块依赖下载完成"
}

# 本地开发模式
do_dev() {
    check_cmd node
    check_cmd npm
    check_cmd go

    # 检查前端依赖是否已安装
    if [ ! -d "node_modules" ]; then
        warn "检测到前端依赖未安装，正在自动安装..."
        npm install
    fi

    info "启动开发环境..."
    echo ""
    info "前端开发服务器: http://localhost:3002"
    info "后端 API 服务:  http://localhost:8008"
    echo ""

    # 启动 Go 后端（开发模式，直接 go run）
    (cd server && go run .) &
    SERVER_PID=$!

    # 启动前端开发服务器
    npm run dev &
    DEV_PID=$!

    # 捕获退出信号，优雅关闭
    trap "info '正在停止服务...'; kill $SERVER_PID $DEV_PID 2>/dev/null; exit 0" SIGINT SIGTERM

    ok "开发环境已启动，按 Ctrl+C 停止"
    wait
}

# 本地生产模式
do_prod() {
    check_cmd node
    check_cmd npm
    check_cmd go

    # 检查前端依赖是否已安装
    if [ ! -d "node_modules" ]; then
        warn "检测到前端依赖未安装，正在自动安装..."
        npm install
    fi

    info "构建前端..."
    npm run build
    ok "前端构建完成"

    info "编译 Go 后端..."
    (cd server && CGO_ENABLED=1 go build -o ../myworker-server .)
    ok "Go 后端编译完成"

    # 复制 .env 到可执行文件同级目录
    cp server/.env .env.runtime 2>/dev/null || true

    info "启动生产服务..."
    echo ""
    info "服务地址: http://localhost:8008"
    echo ""

    ./myworker-server
}

# Docker 部署
do_docker() {
    check_cmd docker

    # 检查 docker compose 命令
    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        error "未找到 docker compose 或 docker-compose 命令"
        exit 1
    fi

    info "使用 Docker Compose 构建并启动..."
    $COMPOSE_CMD up -d --build

    echo ""
    ok "🚀 服务已启动！"
    info "访问地址: http://localhost:8008"
    info "查看日志: $0 docker-logs"
    info "停止服务: $0 docker-stop"
    echo ""
}

# Docker 停止
do_docker_stop() {
    check_cmd docker

    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        error "未找到 docker compose 或 docker-compose 命令"
        exit 1
    fi

    info "停止 Docker 服务..."
    $COMPOSE_CMD down
    ok "服务已停止"
}

# Docker 日志
do_docker_logs() {
    check_cmd docker

    if docker compose version &> /dev/null; then
        COMPOSE_CMD="docker compose"
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        error "未找到 docker compose 或 docker-compose 命令"
        exit 1
    fi

    $COMPOSE_CMD logs -f
}

# 构建
do_build() {
    check_cmd node
    check_cmd npm
    check_cmd go

    if [ ! -d "node_modules" ]; then
        warn "检测到前端依赖未安装，正在自动安装..."
        npm install
    fi

    info "构建前端..."
    npm run build
    ok "前端构建完成，产物目录: dist/"

    info "编译 Go 后端..."
    (cd server && CGO_ENABLED=1 go build -o ../myworker-server .)
    ok "Go 后端编译完成，二进制文件: myworker-server"
}

# ============================================
# 主入口
# ============================================
case "${1:-help}" in
    start)
        do_dev
        ;;
    start-prod)
        do_prod
        ;;
    docker)
        do_docker
        ;;
    docker-stop)
        do_docker_stop
        ;;
    docker-logs)
        do_docker_logs
        ;;
    install)
        do_install
        ;;
    build)
        do_build
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        error "未知命令: $1"
        usage
        exit 1
        ;;
esac
