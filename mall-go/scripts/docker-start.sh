#!/bin/bash

# GoMall Docker启动脚本

set -e

echo "=========================================="
echo "GoMall Docker启动脚本"
echo "=========================================="

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 检查Docker守护进程是否运行
if ! docker ps &> /dev/null; then
    echo "❌ Docker守护进程未运行，请启动Docker"
    exit 1
fi

echo "✅ Docker和Docker Compose已安装"

# 获取环境参数
ENV=${1:-dev}

if [ "$ENV" = "dev" ]; then
    echo "🚀 启动开发环境..."
    docker-compose build
    docker-compose up -d
    
    echo ""
    echo "✅ 开发环境启动成功！"
    echo ""
    echo "📋 服务信息："
    docker-compose ps
    
    echo ""
    echo "🌐 访问地址："
    echo "  API: http://localhost:8081"
    echo "  健康检查: http://localhost:8081/health"
    echo "  Metrics: http://localhost:8081/metrics"
    echo ""
    echo "💾 数据库信息："
    echo "  MySQL: localhost:3306 (mall_user / mall_password)"
    echo "  Redis: localhost:6379 (password: redis123456)"
    
elif [ "$ENV" = "prod" ]; then
    echo "🚀 启动生产环境..."
    
    # 检查.env.prod文件
    if [ ! -f ".env.prod" ]; then
        echo "❌ .env.prod文件不存在，请创建该文件"
        echo "参考: docs/docker-usage-guide.md"
        exit 1
    fi
    
    docker-compose -f docker-compose.prod.yml --env-file .env.prod build
    docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
    
    echo ""
    echo "✅ 生产环境启动成功！"
    echo ""
    echo "📋 服务信息："
    docker-compose -f docker-compose.prod.yml ps
    
else
    echo "❌ 无效的环境参数: $ENV"
    echo "用法: ./docker-start.sh [dev|prod]"
    exit 1
fi

echo ""
echo "📖 查看日志: docker-compose logs -f mall-go"
echo "🛑 停止服务: docker-compose down"
echo ""

