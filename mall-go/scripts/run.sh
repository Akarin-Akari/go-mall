#!/bin/bash

# Mall Go 项目运行脚本

echo "🚀 启动 Mall Go 商城后端项目..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go"
    exit 1
fi

echo "✅ Go环境检查通过"

# 检查配置文件
if [ ! -f "configs/config.yaml" ]; then
    echo "❌ 错误: 配置文件 configs/config.yaml 不存在"
    exit 1
fi

echo "✅ 配置文件检查通过"

# 安装依赖
echo "📦 安装项目依赖..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ 依赖安装失败"
    exit 1
fi

echo "✅ 依赖安装完成"

# 检查数据库连接
echo "🔍 检查数据库连接..."
# 这里可以添加数据库连接检查逻辑

# 运行项目
echo "🎯 启动服务器..."
go run cmd/server/main.go
