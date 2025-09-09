#!/bin/bash

# Mall-Go电商系统全面测试执行脚本
# 目标: 实现90%以上的代码测试覆盖率

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

# 全局变量
PROJECT_ROOT=$(pwd)
BACKEND_DIR="$PROJECT_ROOT/mall-go"
FRONTEND_DIR="$PROJECT_ROOT/mall-frontend"
REPORTS_DIR="$PROJECT_ROOT/test-reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# 创建报告目录
mkdir -p "$REPORTS_DIR"

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
BACKEND_COVERAGE=0
FRONTEND_COVERAGE=0

# 清理函数
cleanup() {
    log_info "清理测试环境..."
    
    # 停止可能运行的服务
    pkill -f "mall-go" || true
    pkill -f "next" || true
    
    # 清理临时文件
    rm -f "$BACKEND_DIR"/*.out
    rm -f "$FRONTEND_DIR"/coverage/*.json
    
    log_success "环境清理完成"
}

# 信号处理
trap cleanup EXIT

# 检查环境依赖
check_dependencies() {
    log_info "检查环境依赖..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        log_error "Go环境未安装"
        exit 1
    fi
    
    # 检查Node.js环境
    if ! command -v node &> /dev/null; then
        log_error "Node.js环境未安装"
        exit 1
    fi
    
    # 检查项目目录
    if [ ! -d "$BACKEND_DIR" ]; then
        log_error "后端项目目录不存在: $BACKEND_DIR"
        exit 1
    fi
    
    if [ ! -d "$FRONTEND_DIR" ]; then
        log_error "前端项目目录不存在: $FRONTEND_DIR"
        exit 1
    fi
    
    log_success "环境依赖检查通过"
}

# 后端测试
run_backend_tests() {
    log_info "开始后端测试..."
    
    cd "$BACKEND_DIR"
    
    # 安装依赖
    log_info "安装Go依赖..."
    go mod tidy
    
    # 运行单元测试
    log_info "运行后端单元测试..."
    go test -v -race -coverprofile=coverage.out ./... > "$REPORTS_DIR/backend-unit-test-$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "后端单元测试通过"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "后端单元测试失败"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # 生成覆盖率报告
    if [ -f coverage.out ]; then
        log_info "生成后端覆盖率报告..."
        go tool cover -html=coverage.out -o "$REPORTS_DIR/backend-coverage-$TIMESTAMP.html"
        BACKEND_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        log_success "后端代码覆盖率: ${BACKEND_COVERAGE}%"
    fi
    
    # 运行集成测试
    log_info "运行后端集成测试..."
    go test -v -tags=integration ./tests/integration/... > "$REPORTS_DIR/backend-integration-test-$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "后端集成测试通过"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "后端集成测试失败"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # 运行性能测试
    log_info "运行后端性能测试..."
    go test -bench=. -benchmem ./... > "$REPORTS_DIR/backend-benchmark-$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "后端性能测试完成"
    else
        log_warning "后端性能测试有警告"
    fi
    
    cd "$PROJECT_ROOT"
}

# 前端测试
run_frontend_tests() {
    log_info "开始前端测试..."
    
    cd "$FRONTEND_DIR"
    
    # 安装依赖
    log_info "安装前端依赖..."
    npm ci
    
    # 运行单元测试
    log_info "运行前端单元测试..."
    npm run test:coverage > "$REPORTS_DIR/frontend-unit-test-$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "前端单元测试通过"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "前端单元测试失败"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # 获取前端覆盖率
    if [ -f coverage/coverage-summary.json ]; then
        FRONTEND_COVERAGE=$(node -e "
            const coverage = require('./coverage/coverage-summary.json');
            console.log(coverage.total.lines.pct);
        ")
        log_success "前端代码覆盖率: ${FRONTEND_COVERAGE}%"
        
        # 复制覆盖率报告
        cp -r coverage "$REPORTS_DIR/frontend-coverage-$TIMESTAMP"
    fi
    
    # 构建项目
    log_info "构建前端项目..."
    npm run build > "$REPORTS_DIR/frontend-build-$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "前端项目构建成功"
    else
        log_error "前端项目构建失败"
        cd "$PROJECT_ROOT"
        return 1
    fi
    
    cd "$PROJECT_ROOT"
}

# E2E测试
run_e2e_tests() {
    log_info "开始端到端测试..."
    
    # 启动后端服务
    log_info "启动后端服务..."
    cd "$BACKEND_DIR"
    go run cmd/server/main.go &
    BACKEND_PID=$!
    
    # 等待后端服务启动
    sleep 10
    
    # 检查后端服务是否启动成功
    if ! curl -s http://localhost:8080/health > /dev/null; then
        log_error "后端服务启动失败"
        kill $BACKEND_PID || true
        return 1
    fi
    
    # 启动前端服务
    log_info "启动前端服务..."
    cd "$FRONTEND_DIR"
    npm start &
    FRONTEND_PID=$!
    
    # 等待前端服务启动
    sleep 15
    
    # 检查前端服务是否启动成功
    if ! curl -s http://localhost:3001 > /dev/null; then
        log_error "前端服务启动失败"
        kill $BACKEND_PID $FRONTEND_PID || true
        return 1
    fi
    
    # 运行Cypress E2E测试
    log_info "运行Cypress E2E测试..."
    npx cypress run --reporter json --reporter-options "output=$REPORTS_DIR/e2e-results-$TIMESTAMP.json" > "$REPORTS_DIR/e2e-test-$TIMESTAMP.log" 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "E2E测试通过"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "E2E测试失败"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # 停止服务
    log_info "停止测试服务..."
    kill $BACKEND_PID $FRONTEND_PID || true
    
    cd "$PROJECT_ROOT"
}

# 性能测试
run_performance_tests() {
    log_info "开始性能测试..."
    
    # 启动后端服务
    cd "$BACKEND_DIR"
    go run cmd/server/main.go &
    BACKEND_PID=$!
    
    # 等待服务启动
    sleep 10
    
    # 运行Apache Bench测试
    if command -v ab &> /dev/null; then
        log_info "运行Apache Bench性能测试..."
        
        # 测试健康检查接口
        ab -n 1000 -c 10 http://localhost:8080/health > "$REPORTS_DIR/performance-health-$TIMESTAMP.log" 2>&1
        
        # 测试商品列表接口
        ab -n 500 -c 5 http://localhost:8080/api/v1/products > "$REPORTS_DIR/performance-products-$TIMESTAMP.log" 2>&1
        
        log_success "性能测试完成"
    else
        log_warning "Apache Bench未安装，跳过性能测试"
    fi
    
    # 停止服务
    kill $BACKEND_PID || true
    
    cd "$PROJECT_ROOT"
}

# 生成测试报告
generate_report() {
    log_info "生成测试报告..."
    
    REPORT_FILE="$REPORTS_DIR/test-summary-$TIMESTAMP.md"
    
    cat > "$REPORT_FILE" << EOF
# Mall-Go电商系统测试报告

**测试时间**: $(date)
**测试版本**: $(git rev-parse --short HEAD)

## 📊 测试结果统计

- **总测试数**: $TOTAL_TESTS
- **通过测试**: $PASSED_TESTS
- **失败测试**: $FAILED_TESTS
- **通过率**: $(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)%

## 📈 代码覆盖率

- **后端覆盖率**: ${BACKEND_COVERAGE}%
- **前端覆盖率**: ${FRONTEND_COVERAGE}%
- **总体覆盖率**: $(echo "scale=2; ($BACKEND_COVERAGE + $FRONTEND_COVERAGE) / 2" | bc)%

## 📋 测试详情

### 后端测试
- 单元测试: $([ -f "$REPORTS_DIR/backend-unit-test-$TIMESTAMP.log" ] && echo "✅ 完成" || echo "❌ 失败")
- 集成测试: $([ -f "$REPORTS_DIR/backend-integration-test-$TIMESTAMP.log" ] && echo "✅ 完成" || echo "❌ 失败")
- 性能测试: $([ -f "$REPORTS_DIR/backend-benchmark-$TIMESTAMP.log" ] && echo "✅ 完成" || echo "❌ 失败")

### 前端测试
- 单元测试: $([ -f "$REPORTS_DIR/frontend-unit-test-$TIMESTAMP.log" ] && echo "✅ 完成" || echo "❌ 失败")
- 构建测试: $([ -f "$REPORTS_DIR/frontend-build-$TIMESTAMP.log" ] && echo "✅ 完成" || echo "❌ 失败")

### E2E测试
- 端到端测试: $([ -f "$REPORTS_DIR/e2e-test-$TIMESTAMP.log" ] && echo "✅ 完成" || echo "❌ 失败")

## 📁 报告文件

所有详细的测试日志和覆盖率报告已保存在: \`$REPORTS_DIR\`

EOF

    log_success "测试报告已生成: $REPORT_FILE"
}

# 主函数
main() {
    log_info "开始Mall-Go电商系统全面测试"
    log_info "目标: 实现90%以上的代码测试覆盖率"
    
    # 检查环境
    check_dependencies
    
    # 运行测试
    run_backend_tests
    run_frontend_tests
    run_e2e_tests
    run_performance_tests
    
    # 生成报告
    generate_report
    
    # 输出最终结果
    echo ""
    log_info "========== 测试完成 =========="
    log_info "总测试数: $TOTAL_TESTS"
    log_info "通过测试: $PASSED_TESTS"
    log_info "失败测试: $FAILED_TESTS"
    log_info "后端覆盖率: ${BACKEND_COVERAGE}%"
    log_info "前端覆盖率: ${FRONTEND_COVERAGE}%"
    
    # 检查是否达到目标覆盖率
    TOTAL_COVERAGE=$(echo "scale=2; ($BACKEND_COVERAGE + $FRONTEND_COVERAGE) / 2" | bc)
    if (( $(echo "$TOTAL_COVERAGE >= 90" | bc -l) )); then
        log_success "🎉 恭喜！达到90%以上代码覆盖率目标: ${TOTAL_COVERAGE}%"
        exit 0
    else
        log_warning "⚠️  未达到90%覆盖率目标，当前: ${TOTAL_COVERAGE}%"
        exit 1
    fi
}

# 执行主函数
main "$@"
