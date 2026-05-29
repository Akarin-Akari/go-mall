# CI/CD流水线使用指南

## 概述

GoMall项目使用GitHub Actions实现完整的CI/CD流水线，包括：
- ✅ 自动化测试
- ✅ 代码质量检查
- ✅ 安全扫描
- ✅ Docker镜像构建
- ✅ 自动化部署

---

## 工作流说明

### 1. test.yml - 自动化测试

**触发条件**: 
- Push到 main, master, aitest, develop 分支
- Pull Request到上述分支

**执行步骤**:
1. 启动MySQL和Redis服务
2. 下载Go依赖
3. 运行单元测试（带代码覆盖率）
4. 上传覆盖率到Codecov
5. 运行集成测试
6. 生成测试报告

**环境变量**:
```bash
ENV=test
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=test_user
DB_PASSWORD=test_password
DB_NAME=mall_go_test
REDIS_HOST=localhost
REDIS_PORT=6379
```

### 2. lint.yml - 代码质量检查

**触发条件**: 
- Push到 main, master, aitest, develop 分支
- Pull Request到上述分支

**执行步骤**:
1. 运行golangci-lint
2. 检查代码格式（gofmt）
3. 运行go vet
4. 安全扫描（gosec）
5. 静态检查（staticcheck）
6. 检查未使用代码

**检查项**:
- ✅ 代码风格
- ✅ 代码格式
- ✅ 潜在bug
- ✅ 安全问题
- ✅ 性能问题
- ✅ 未使用代码

### 3. build.yml - Docker镜像构建

**触发条件**:
- Push到 main, master, aitest 分支
- 创建版本标签（v*）
- Pull Request（仅构建，不推送）

**执行步骤**:
1. 设置Docker Buildx
2. 登录容器镜像仓库
3. 提取元数据（版本、标签）
4. 构建并推送Docker镜像
5. 缓存构建层

**镜像标签**:
- `branch-name` - 分支名称
- `v1.0.0` - 版本号
- `sha-xxxxx` - 提交哈希
- `latest` - 最新版本（仅main分支）

### 4. deploy.yml - 自动化部署

**触发条件**:
- Push到 main/master 分支（部署到staging）
- 创建版本标签（部署到production）
- 手动触发（workflow_dispatch）

**部署环境**:
- **Staging**: 自动部署到测试环境
- **Production**: 手动或通过版本标签部署

**执行步骤**:
1. 提取版本号
2. 部署到目标环境
3. 运行烟雾测试
4. 运行健康检查
5. 发送通知

### 5. ci.yml - 综合CI工作流

**触发条件**: 
- Push到 main, master, aitest, develop 分支
- Pull Request到上述分支

**执行步骤**:
1. 代码质量检查（quality）
2. 运行测试（test）- 依赖quality通过
3. 构建Docker镜像（build）- 依赖test通过
4. 生成报告（report）

**依赖关系**:
```
quality
   ↓
test
   ↓
build
   ↓
report
```

---

## 快速开始

### 查看工作流状态

```bash
# 在GitHub仓库中查看Actions标签
# https://github.com/your-org/mall-go/actions
```

### 手动触发工作流

```bash
# 使用GitHub CLI
gh workflow run deploy.yml -f environment=staging

# 或在GitHub Web界面中点击"Run workflow"
```

### 查看工作流日志

```bash
# 使用GitHub CLI
gh run list
gh run view <run-id>
gh run view <run-id> --log

# 或在GitHub Web界面中查看
```

---

## 配置说明

### 环境变量

工作流使用以下环境变量：

```yaml
# 测试环境
ENV: test
DB_DRIVER: mysql
DB_HOST: localhost
DB_PORT: 3306
DB_USERNAME: test_user
DB_PASSWORD: test_password
DB_NAME: mall_go_test
REDIS_HOST: localhost
REDIS_PORT: 6379
```

### Secrets（敏感信息）

需要在GitHub仓库设置中配置以下Secrets：

```bash
# 容器镜像仓库凭证
REGISTRY_USERNAME
REGISTRY_PASSWORD

# 部署凭证
DEPLOY_SSH_KEY
DEPLOY_HOST
DEPLOY_USER

# 通知凭证
SLACK_WEBHOOK_URL
```

### 分支保护规则

建议为main和master分支配置以下保护规则：

```yaml
# 需要通过以下检查才能合并PR
- CI (test.yml)
- Lint (lint.yml)
- Build (build.yml)

# 需要至少1个批准
- Require pull request reviews

# 禁止强制推送
- Dismiss stale pull request approvals
```

---

## 常见任务

### 运行特定工作流

```bash
# 运行测试
gh workflow run test.yml

# 运行代码检查
gh workflow run lint.yml

# 构建Docker镜像
gh workflow run build.yml

# 部署到staging
gh workflow run deploy.yml -f environment=staging
```

### 查看工作流结果

```bash
# 列出最近的运行
gh run list --workflow=test.yml

# 查看特定运行的详情
gh run view <run-id> --log

# 下载工作流产物
gh run download <run-id> --name test-report
```

### 调试工作流

```bash
# 启用调试日志
gh workflow run test.yml --debug

# 查看详细日志
gh run view <run-id> --log --verbose
```

---

## 故障排查

### 测试失败

**症状**: test.yml工作流失败

**排查步骤**:
1. 查看工作流日志
2. 检查MySQL和Redis服务是否启动
3. 检查环境变量配置
4. 本地运行测试验证

```bash
# 本地运行测试
go test -v ./...

# 查看工作流日志
gh run view <run-id> --log
```

### 代码检查失败

**症状**: lint.yml工作流失败

**排查步骤**:
1. 查看具体的检查失败信息
2. 本地运行golangci-lint
3. 修复代码问题
4. 重新提交

```bash
# 本地运行lint检查
golangci-lint run

# 自动修复格式问题
gofmt -s -w .
```

### Docker构建失败

**症状**: build.yml工作流失败

**排查步骤**:
1. 查看Docker构建日志
2. 本地构建Docker镜像
3. 检查Dockerfile语法
4. 检查依赖是否可用

```bash
# 本地构建Docker镜像
docker build -t mall-go:test .

# 查看构建日志
docker build --progress=plain -t mall-go:test .
```

### 部署失败

**症状**: deploy.yml工作流失败

**排查步骤**:
1. 查看部署日志
2. 检查部署凭证是否正确
3. 检查目标环境是否可访问
4. 检查健康检查端点

```bash
# 查看部署日志
gh run view <run-id> --log

# 手动测试健康检查
curl -f https://staging.example.com/health
```

---

## 最佳实践

### 1. 分支策略

```
main/master (生产)
    ↑
    ├─ aitest (测试)
    │   ↑
    │   └─ feature/* (功能分支)
    │
    └─ develop (开发)
        ↑
        └─ feature/* (功能分支)
```

### 2. 提交规范

```bash
# 使用有意义的提交信息
git commit -m "feat: add SQLite support to Docker"
git commit -m "fix: resolve database connection issue"
git commit -m "docs: update CI/CD guide"
```

### 3. Pull Request流程

1. 创建功能分支
2. 提交代码
3. 创建Pull Request
4. 等待CI/CD检查通过
5. 获得代码审查批准
6. 合并到目标分支

### 4. 版本发布

```bash
# 创建版本标签
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# 自动触发部署到生产环境
```

---

## 监控和告警

### GitHub Actions监控

```bash
# 查看工作流统计
gh workflow list

# 查看最近的运行
gh run list

# 查看特定工作流的运行
gh run list --workflow=test.yml
```

### 集成通知

可以配置以下通知方式：

- ✅ GitHub通知
- ✅ Slack通知
- ✅ 邮件通知
- ✅ 自定义Webhook

---

## 相关文档

- [Docker使用指南](./docker-usage-guide.md)
- [本地开发环境搭建](./local-development-setup.md)
- [GitHub Actions官方文档](https://docs.github.com/en/actions)

---

**最后更新**: 2025-01-31

