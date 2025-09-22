@echo off
echo 🚀 启动Mall-Go后端服务并测试
echo ================================

cd /d E:\Workspace_Draft\GoLang\goProject\mall-go

echo 📡 启动后端服务...
start "Mall-Go Backend" cmd /k "go run cmd/server/main.go"

echo ⏳ 等待服务启动...
timeout /t 5 /nobreak > nul

echo 🧪 测试后端API...
cd /d E:\Workspace_Draft\GoLang\goProject

echo 测试健康检查...
curl -s http://localhost:8080/health
echo.

echo 测试登录API...
curl -X POST http://localhost:8080/api/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
echo.

echo ================================
echo 测试完成
pause
