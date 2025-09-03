#!/bin/bash

# Mall-Go电商系统快速启动脚本
# 适用于Linux/macOS系统

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装，请先安装 $1"
        exit 1
    fi
}

# 检查端口是否被占用
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null ; then
        log_warning "端口 $1 已被占用"
        return 1
    fi
    return 0
}

# 主函数
main() {
    log_info "开始Mall-Go电商系统快速部署..."
    
    # 1. 环境检查
    log_info "检查环境依赖..."
    check_command "go"
    check_command "node"
    check_command "npm"
    check_command "git"
    
    # 检查版本
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    NODE_VERSION=$(node --version | sed 's/v//')
    
    log_info "Go版本: $GO_VERSION"
    log_info "Node.js版本: $NODE_VERSION"
    
    # 2. 检查端口
    log_info "检查端口占用情况..."
    if ! check_port 8080; then
        log_error "后端端口8080被占用，请释放端口后重试"
        exit 1
    fi
    
    if ! check_port 3001; then
        log_warning "前端端口3001被占用，Next.js将自动使用其他端口"
    fi
    
    # 3. 启动后端服务
    log_info "启动后端服务..."
    cd mall-go
    
    # 安装Go依赖
    log_info "安装Go依赖..."
    go mod tidy
    go mod download
    
    # 后台启动后端服务
    log_info "启动后端API服务..."
    nohup go run cmd/server/main.go > backend.log 2>&1 &
    BACKEND_PID=$!
    echo $BACKEND_PID > backend.pid
    
    # 等待后端启动
    log_info "等待后端服务启动..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null; then
            log_success "后端服务启动成功！"
            break
        fi
        sleep 1
        if [ $i -eq 30 ]; then
            log_error "后端服务启动超时"
            exit 1
        fi
    done
    
    # 4. 启动前端服务
    log_info "启动前端服务..."
    cd ../mall-frontend
    
    # 安装npm依赖
    log_info "安装npm依赖..."
    npm install
    
    # 创建环境变量文件
    log_info "配置环境变量..."
    cat > .env.local << EOF
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_API_TIMEOUT=10000
NEXT_PUBLIC_APP_NAME=Mall-Go
NEXT_PUBLIC_APP_VERSION=1.0.0
NEXT_PUBLIC_DEBUG=true
EOF
    
    # 启动前端服务
    log_info "启动前端Web服务..."
    npm run dev &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > frontend.pid
    
    # 等待前端启动
    log_info "等待前端服务启动..."
    sleep 10
    
    # 5. 验证服务
    log_info "验证服务状态..."
    
    # 验证后端
    if curl -s http://localhost:8080/health | grep -q "ok"; then
        log_success "后端API服务正常"
    else
        log_error "后端API服务异常"
    fi
    
    # 验证前端（检查端口）
    if lsof -Pi :3001 -sTCP:LISTEN -t >/dev/null; then
        log_success "前端Web服务正常 (端口3001)"
        FRONTEND_URL="http://localhost:3001"
    elif lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null; then
        log_success "前端Web服务正常 (端口3000)"
        FRONTEND_URL="http://localhost:3000"
    else
        log_warning "前端服务端口检测失败，请手动检查"
        FRONTEND_URL="http://localhost:3001"
    fi
    
    # 6. 显示启动信息
    echo ""
    log_success "🎉 Mall-Go电商系统启动成功！"
    echo ""
    echo "📋 服务信息:"
    echo "  后端API: http://localhost:8080"
    echo "  前端Web: $FRONTEND_URL"
    echo "  健康检查: http://localhost:8080/health"
    echo ""
    echo "👤 测试账号:"
    echo "  用户名: newuser2024"
    echo "  密码: 123456789"
    echo ""
    echo "🔧 管理命令:"
    echo "  查看后端日志: tail -f mall-go/backend.log"
    echo "  停止后端服务: kill \$(cat mall-go/backend.pid)"
    echo "  停止前端服务: kill \$(cat mall-frontend/frontend.pid)"
    echo ""
    echo "🌐 立即体验: $FRONTEND_URL"
    
    # 7. 自动打开浏览器（可选）
    if command -v xdg-open &> /dev/null; then
        log_info "正在打开浏览器..."
        xdg-open $FRONTEND_URL
    elif command -v open &> /dev/null; then
        log_info "正在打开浏览器..."
        open $FRONTEND_URL
    fi
}

# 清理函数
cleanup() {
    log_info "正在清理进程..."
    if [ -f mall-go/backend.pid ]; then
        kill $(cat mall-go/backend.pid) 2>/dev/null || true
        rm mall-go/backend.pid
    fi
    if [ -f mall-frontend/frontend.pid ]; then
        kill $(cat mall-frontend/frontend.pid) 2>/dev/null || true
        rm mall-frontend/frontend.pid
    fi
}

# 捕获退出信号
trap cleanup EXIT

# 执行主函数
main "$@"
