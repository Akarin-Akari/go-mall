# Mall-Go部署工具使用说明

## 📋 工具集概览

本工具集为Mall-Go电商系统提供了完整的部署、管理和监控解决方案，包含以下文件：

### 📚 文档文件
- **`Mall-Go部署与启动指南.md`** - 详细的手动部署指南
- **`README-部署工具使用说明.md`** - 本文件，工具使用说明

### 🚀 自动化脚本
- **`quick-start.bat`** - Windows快速启动脚本
- **`quick-start.sh`** - Linux/macOS快速启动脚本  
- **`stop-services.bat`** - Windows服务停止脚本
- **`check-status.bat`** - Windows系统状态检查脚本

---

## 🎯 使用场景

### 场景1: 新手快速体验
**目标**: 零基础用户快速启动系统进行体验
```bash
# Windows用户
双击运行: quick-start.bat

# Linux/macOS用户  
chmod +x quick-start.sh
./quick-start.sh
```

### 场景2: 开发环境搭建
**目标**: 开发人员搭建完整的开发环境
```bash
1. 阅读: Mall-Go部署与启动指南.md
2. 手动执行各个步骤
3. 使用: check-status.bat 验证环境
```

### 场景3: 测试环境部署
**目标**: 测试人员部署稳定的测试环境
```bash
1. 使用: quick-start.bat 快速启动
2. 使用: check-status.bat 验证状态
3. 执行黑盒测试流程
```

### 场景4: 生产环境部署
**目标**: 运维人员部署生产环境
```bash
1. 参考: Mall-Go部署与启动指南.md
2. 根据生产环境需求调整配置
3. 使用监控脚本定期检查状态
```

---

## 🛠️ 工具详细说明

### 1. quick-start.bat (Windows快速启动)

**功能**: 一键启动前后端服务
**适用**: Windows 10/11系统
**执行时间**: 约2-5分钟

**执行流程**:
1. ✅ 环境依赖检查 (Go, Node.js, npm)
2. ✅ 端口占用检查 (8080, 3001)
3. ✅ 后端服务启动 (Go API)
4. ✅ 前端服务启动 (Next.js)
5. ✅ 服务状态验证
6. ✅ 自动打开浏览器

**输出示例**:
```
========================================
  🎉 Mall-Go电商系统启动成功！
========================================

📋 服务信息:
  后端API: http://localhost:8080
  前端Web: http://localhost:3001
  健康检查: http://localhost:8080/health

👤 测试账号:
  用户名: newuser2024
  密码: 123456789
```

### 2. quick-start.sh (Linux/macOS快速启动)

**功能**: 一键启动前后端服务 (Unix系统)
**适用**: Linux/macOS系统
**执行时间**: 约2-5分钟

**使用方法**:
```bash
# 添加执行权限
chmod +x quick-start.sh

# 执行脚本
./quick-start.sh
```

**特色功能**:
- 🎨 彩色日志输出
- 🔍 智能端口检测
- 🌐 自动打开浏览器
- 🛡️ 信号处理和清理

### 3. stop-services.bat (服务停止脚本)

**功能**: 安全停止所有Mall-Go相关服务
**适用**: Windows系统

**停止范围**:
- 🔴 后端API服务 (端口8080)
- 🔴 前端Web服务 (端口3001/3000)
- 🔴 相关Node.js进程
- 🧹 可选清理日志文件

**使用场景**:
- 开发完成后清理环境
- 解决端口冲突问题
- 系统维护前停止服务

### 4. check-status.bat (状态检查脚本)

**功能**: 全面检查系统状态和健康度
**适用**: Windows系统

**检查项目**:
- ✅ 环境依赖 (Go, Node.js, npm, Git)
- ✅ 项目结构 (目录、关键文件)
- ✅ 服务运行状态 (端口占用)
- ✅ API端点测试 (健康检查、核心API)
- ✅ 系统资源 (内存、磁盘)
- 💡 智能建议 (问题诊断)

**输出示例**:
```
✅ Go: go1.19.5
✅ Node.js: v18.17.0
✅ 后端服务: 正在运行 (端口8080)
✅ 前端服务: 正在运行 (端口3001)
✅ GET /health - 健康检查
✅ GET /api/v1/products - 商品列表
```

---

## 📖 使用指南

### 第一次使用 (推荐流程)

#### 步骤1: 环境准备
```bash
# 确保已安装必要软件
- Go 1.19+
- Node.js 18+
- Git
```

#### 步骤2: 获取项目
```bash
# 克隆或下载项目到本地
# 确保目录结构正确:
# ├── mall-go/          (后端项目)
# ├── mall-frontend/    (前端项目)
# ├── quick-start.bat   (启动脚本)
# └── 其他工具文件...
```

#### 步骤3: 快速启动
```bash
# Windows用户
双击 quick-start.bat

# 等待启动完成，自动打开浏览器
# 访问 http://localhost:3001
```

#### 步骤4: 功能测试
```bash
# 使用测试账号登录
用户名: newuser2024
密码: 123456789

# 体验完整购物流程
1. 浏览商品 → 2. 加入购物车 → 3. 确认订单 → 4. 查看订单
```

### 日常开发使用

#### 启动开发环境
```bash
# 快速启动
quick-start.bat

# 检查状态
check-status.bat
```

#### 停止开发环境
```bash
# 停止所有服务
stop-services.bat
```

#### 问题排查
```bash
# 1. 检查系统状态
check-status.bat

# 2. 查看详细指南
阅读 Mall-Go部署与启动指南.md

# 3. 重新启动服务
stop-services.bat
quick-start.bat
```

---

## 🔧 高级配置

### 自定义端口配置

#### 后端端口修改
```yaml
# 编辑 mall-go/config/config.yaml
server:
  port: 8081  # 修改为其他端口
```

#### 前端端口修改
```bash
# 编辑 mall-frontend/package.json
"scripts": {
  "dev": "next dev -p 3002"  # 修改为其他端口
}
```

### 环境变量配置

#### 前端环境变量
```bash
# 编辑 mall-frontend/.env.local
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081  # 对应后端端口
NEXT_PUBLIC_API_TIMEOUT=15000                   # 增加超时时间
NEXT_PUBLIC_DEBUG=false                         # 生产环境关闭调试
```

### 数据库配置

#### SQLite配置 (默认)
```yaml
# mall-go/config/config.yaml
database:
  type: sqlite
  dsn: mall_go.db
```

#### MySQL配置 (可选)
```yaml
# mall-go/config/config.yaml
database:
  type: mysql
  dsn: "user:password@tcp(localhost:3306)/mall_go?charset=utf8mb4&parseTime=True&loc=Local"
```

---

## 🚨 常见问题解决

### 问题1: 端口被占用
```bash
# 现象: 启动失败，提示端口占用
# 解决: 
1. 运行 stop-services.bat
2. 或手动释放端口
3. 重新运行 quick-start.bat
```

### 问题2: 依赖安装失败
```bash
# 现象: npm install 或 go mod download 失败
# 解决:
1. 检查网络连接
2. 配置代理 (如需要)
3. 清理缓存重试
```

### 问题3: 服务启动超时
```bash
# 现象: 等待服务启动超时
# 解决:
1. 检查系统资源 (内存、CPU)
2. 关闭不必要的程序
3. 增加启动等待时间
```

### 问题4: API连接失败
```bash
# 现象: 前端无法连接后端API
# 解决:
1. 运行 check-status.bat 检查状态
2. 确认后端服务正常运行
3. 检查防火墙设置
4. 验证环境变量配置
```

---

## 📞 技术支持

### 获取帮助
1. **查看状态**: 运行 `check-status.bat` 获取详细状态信息
2. **阅读文档**: 参考 `Mall-Go部署与启动指南.md`
3. **重置环境**: 运行 `stop-services.bat` 后重新启动

### 报告问题时请提供
1. 操作系统版本
2. Go和Node.js版本信息
3. 错误信息截图
4. `check-status.bat` 的输出结果

---

## 🎉 成功标志

当看到以下信息时，表示部署成功：

### 控制台输出
```
🎉 Mall-Go电商系统启动成功！
📋 服务信息:
  后端API: http://localhost:8080 ✅
  前端Web: http://localhost:3001 ✅
```

### 浏览器访问
- ✅ 首页正常显示轮播图和商品
- ✅ 用户可以正常注册和登录
- ✅ 商品列表显示15个测试商品
- ✅ 购物车和订单功能正常

### API测试
```bash
curl http://localhost:8080/health
# 返回: {"message":"Mall Go API is running","status":"ok"}
```

---

**Mall-Go部署工具集 v1.0**  
**更新时间**: 2024年1月  
**维护状态**: 活跃维护  
**兼容性**: Windows 10/11, Linux, macOS
