# CI/CD快速参考卡片

## 工作流概览

| 工作流 | 文件 | 触发条件 | 功能 |
|------|------|---------|------|
| Tests | test.yml | Push/PR | 运行单元测试和集成测试 |
| Lint | lint.yml | Push/PR | 代码质量检查和安全扫描 |
| Build | build.yml | Push/Tag | Docker镜像构建和推送 |
| Deploy | deploy.yml | Push/Tag | 自动化部署到staging/production |
| CI | ci.yml | Push/PR | 综合检查（质量→测试→构建） |

---

## 快速命令

### 查看工作流

```bash
# 列出所有工作流
gh workflow list

# 查看最近的运行
gh run list

# 查看特定工作流的运行
gh run list --workflow=test.yml

# 查看运行详情
gh run view <run-id>

# 查看运行日志
gh run view <run-id> --log
```

### 手动触发工作流

```bash
# 触发测试
gh workflow run test.yml

# 触发代码检查
gh workflow run lint.yml

# 触发构建
gh workflow run build.yml

# 触发部署
gh workflow run deploy.yml -f environment=staging
```

### 下载产物

```bash
# 列出产物
gh run download <run-id>

# 下载特定产物
gh run download <run-id> --name test-report
```

---

## 本地验证

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包的测试
go test -v ./internal/handler/...

# 运行测试并生成覆盖率
go test -v -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -html=coverage.out
```

### 代码检查

```bash
# 运行golangci-lint
golangci-lint run

# 检查代码格式
gofmt -s -l .

# 自动修复格式
gofmt -s -w .

# 运行go vet
go vet ./...

# 安全扫描
gosec ./...
```

### Docker构建

```bash
# 构建镜像
docker build -t mall-go:latest .

# 构建并指定标签
docker build -t mall-go:v1.0.0 .

# 查看构建日志
docker build --progress=plain -t mall-go:latest .
```

---

## 分支和标签

### 分支命名规范

```bash
# 功能分支
git checkout -b feature/add-sqlite-support

# 修复分支
git checkout -b fix/database-connection-issue

# 文档分支
git checkout -b docs/update-readme
```

### 版本标签

```bash
# 创建版本标签
git tag -a v1.0.0 -m "Release version 1.0.0"

# 推送标签
git push origin v1.0.0

# 列出标签
git tag -l

# 删除标签
git tag -d v1.0.0
git push origin --delete v1.0.0
```

---

## 常见问题

### Q: 如何跳过CI检查？

**A**: 在提交信息中添加 `[skip ci]`：
```bash
git commit -m "docs: update README [skip ci]"
```

### Q: 如何查看工作流失败的原因？

**A**: 
```bash
# 查看详细日志
gh run view <run-id> --log

# 或在GitHub Web界面中查看
```

### Q: 如何重新运行失败的工作流？

**A**:
```bash
# 使用GitHub CLI
gh run rerun <run-id>

# 或在GitHub Web界面中点击"Re-run jobs"
```

### Q: 如何配置Secrets？

**A**: 在GitHub仓库设置中：
1. Settings → Secrets and variables → Actions
2. 点击"New repository secret"
3. 输入名称和值

### Q: 如何查看工作流的执行时间？

**A**:
```bash
# 查看运行详情
gh run view <run-id>

# 查看特定job的时间
gh run view <run-id> --json jobs
```

---

## 工作流状态

### 状态含义

| 状态 | 含义 |
|------|------|
| ✅ completed | 成功完成 |
| ❌ failure | 失败 |
| ⏳ in_progress | 进行中 |
| ⏸️ queued | 排队中 |
| ⚠️ skipped | 已跳过 |

### 查看状态

```bash
# 查看最近运行的状态
gh run list --limit 10

# 查看特定工作流的状态
gh run list --workflow=test.yml --limit 10
```

---

## 部署流程

### Staging部署

```bash
# 自动触发（Push到main分支）
git push origin main

# 或手动触发
gh workflow run deploy.yml -f environment=staging
```

### Production部署

```bash
# 创建版本标签（自动触发）
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 或手动触发
gh workflow run deploy.yml -f environment=production
```

---

## 监控和告警

### 查看工作流统计

```bash
# 查看所有工作流
gh workflow list

# 查看工作流的最近运行
gh run list --workflow=test.yml

# 查看工作流的成功率
gh run list --workflow=test.yml --json conclusion
```

### 设置通知

在GitHub仓库设置中配置：
- 邮件通知
- Slack集成
- 自定义Webhook

---

## 最佳实践

### ✅ 推荐做法

- ✅ 为每个功能创建单独的分支
- ✅ 在合并前等待CI/CD检查通过
- ✅ 定期更新依赖
- ✅ 编写有意义的提交信息
- ✅ 使用版本标签标记发布

### ❌ 避免做法

- ❌ 直接推送到main分支
- ❌ 忽略CI/CD检查失败
- ❌ 提交大量未测试的代码
- ❌ 使用模糊的提交信息
- ❌ 跳过代码审查

---

## 相关文档

- [CI/CD完整指南](./cicd-guide.md)
- [Docker使用指南](./docker-usage-guide.md)
- [GitHub Actions官方文档](https://docs.github.com/en/actions)

---

**最后更新**: 2025-01-31

