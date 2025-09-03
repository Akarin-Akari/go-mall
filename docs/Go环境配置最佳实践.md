# Go 环境配置最佳实践

## 🎯 现代 Go 开发环境配置

### 1. Go 环境变量设置

#### Windows 系统

```
GOROOT = C:\Go                    # Go安装目录
GOPATH = C:\Users\YourName\go     # 全局工作空间（可选）
PATH   = %GOROOT%\bin;%GOPATH%\bin
```

#### Linux/Mac 系统

```bash
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

### 2. 项目结构（推荐）

```
# 你的工作目录
E:\Workspace_Draft\GoLang\goProject\
├── mall-go/                    # 商城项目
│   ├── go.mod
│   ├── cmd/
│   └── ...
├── blog-api/                   # 博客API项目
│   ├── go.mod
│   ├── cmd/
│   └── ...
├── microservice/               # 微服务项目
│   ├── go.mod
│   ├── cmd/
│   └── ...
└── ...
```

### 3. 每个项目独立管理

#### 项目 1：商城项目

```bash
cd E:\Workspace_Draft\GoLang\goProject\mall-go
go mod init mall-go
go mod tidy
go run cmd/server/main.go
```

#### 项目 2：博客项目

```bash
cd E:\Workspace_Draft\GoLang\goProject\blog-api
go mod init blog-api
go mod tidy
go run main.go
```

#### 项目 3：微服务项目

```bash
cd E:\Workspace_Draft\GoLang\goProject\microservice
go mod init microservice
go mod tidy
go run cmd/server/main.go
```

## 🔄 Go Modules 的优势

### 1. 项目独立性

- 每个项目有自己的 `go.mod` 文件
- 依赖版本独立管理
- 不需要共享 GOPATH

### 2. 版本控制

```go
// go.mod 示例
module mall-go

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    gorm.io/gorm v1.25.5
    // 每个项目可以指定不同的依赖版本
)
```

### 3. 依赖管理

```bash
# 添加依赖
go get github.com/gin-gonic/gin

# 更新依赖
go get -u github.com/gin-gonic/gin

# 清理未使用的依赖
go mod tidy
```

## 🛠️ 实际配置步骤

### 1. 设置环境变量

**Windows 系统：**

1. 右键"此电脑" → "属性" → "高级系统设置" → "环境变量"
2. 系统变量设置：
   ```
   GOROOT = C:\Go
   GOPATH = C:\Users\YourName\go
   ```
3. Path 变量添加：
   ```
   %GOROOT%\bin
   %GOPATH%\bin
   ```

### 2. 验证配置

```bash
go version
go env GOPATH
go env GOROOT
```

### 3. 项目开发流程

```bash
# 1. 创建新项目目录
mkdir my-new-project
cd my-new-project

# 2. 初始化Go模块
go mod init my-new-project

# 3. 添加依赖
go get github.com/gin-gonic/gin

# 4. 编写代码
# 5. 运行项目
go run main.go
```

## 📊 GOPATH vs Go Modules 对比

| 特性     | GOPATH 模式          | Go Modules 模式 |
| -------- | -------------------- | --------------- |
| 项目位置 | 必须在 GOPATH/src 下 | 任意目录        |
| 依赖管理 | 全局共享             | 项目独立        |
| 版本控制 | 困难                 | 简单            |
| 项目隔离 | 差                   | 好              |
| 现代化   | 过时                 | 推荐            |

## 🎯 最佳实践建议

### 1. 使用 Go Modules

- 所有新项目都使用 Go Modules
- 每个项目独立的 `go.mod` 文件
- 明确的依赖版本管理

### 2. GOPATH 用途

- 全局工具安装：`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- 老项目兼容性
- 一些第三方工具的缓存

### 3. 项目组织

```
工作目录/
├── 项目1/
│   ├── go.mod
│   └── ...
├── 项目2/
│   ├── go.mod
│   └── ...
└── 项目3/
    ├── go.mod
    └── ...
```

## 🚀 快速验证

配置完成后，你可以这样验证：

```bash
# 1. 检查Go环境
go version

# 2. 创建测试项目
mkdir test-project
cd test-project

# 3. 初始化模块
go mod init test-project

# 4. 添加依赖
go get github.com/gin-gonic/gin

# 5. 创建main.go
echo 'package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    r.Run(":8080")
}' > main.go

# 6. 运行项目
go run main.go
```

这样你就可以在任意目录下创建 Go 项目，而不需要修改 GOPATH 了！🎉
