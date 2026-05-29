# Augment MCP 修复指南

## 问题诊断

根据日志和官方文档分析，问题在于：

1. **配置错误**：当前 `augment-context-engine` 配置使用了错误的包 `@upstash/context7-mcp@latest`
2. **环境问题**：Cursor 尝试在 WSL2 中执行 `auggie`，但 WSL2 中找不到 `node`
3. **误解**：Augment MCP **不需要**在 WSL2 中运行，可以在 Windows 环境中运行

## 官方文档参考

根据 [Augment 官方文档](https://docs.augmentcode.com/context-services/mcp/quickstart-claude-code)，正确的配置步骤是：

### 1. 安装 Auggie CLI

```bash
npm install -g @augmentcode/auggie@prerelease
```

### 2. 登录 Augment

```bash
auggie login
```

这会打开浏览器窗口进行身份验证。

### 3. 配置 MCP 服务器

**用户级别配置**（所有项目可用）：

```bash
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["--mcp"]}'
```

**项目级别配置**（仅当前项目）：

```bash
claude mcp add-json auggie-mcp --scope project '{"type":"stdio","command":"auggie","args":["--mcp"]}'
```

### 4. 指定工作区（可选）

如果需要指定工作区路径：

```bash
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["-w","/path/to/project","--mcp"]}'
```

## 修复方案

### 方案 1：使用 Windows 命令（已修复）

我已经修改了 `.claude/settings.local.json` 文件，将配置改为：

```json
"augment-context-engine": {
  "command": "cmd",
  "args": [
    "/c",
    "auggie",
    "--mcp"
  ]
}
```

这样会在 Windows 环境中执行 `auggie`，而不是在 WSL2 中。

### 方案 2：使用 npx（如果 auggie 未全局安装）

如果 `auggie` 没有全局安装，可以使用 `npx`：

```json
"augment-context-engine": {
  "command": "cmd",
  "args": [
    "/c",
    "npx",
    "-y",
    "@augmentcode/auggie@prerelease",
    "--mcp"
  ]
}
```

### 方案 3：在 WSL2 中安装 Node.js（不推荐）

如果确实需要在 WSL2 中运行，需要先安装 Node.js：

```bash
# 使用 NodeSource 官方源
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# 然后安装 auggie
npm install -g @augmentcode/auggie@prerelease
```

## 验证修复

修复后，需要：

1. **重启 Cursor**：让新的 MCP 配置生效
2. **检查日志**：应该不再出现 `exec: node: not found` 错误
3. **测试功能**：尝试使用 Augment 的代码检索功能

## 注意事项

- **不需要 WSL2**：Augment MCP 可以在 Windows 环境中正常运行
- **需要 Node.js**：确保 Windows 环境中已安装 Node.js 和 npm
- **需要登录**：首次使用前需要运行 `auggie login` 进行身份验证
- **工作区路径**：如果需要指定工作区，使用 `-w` 参数，Windows 路径格式如 `E:\\Workspace_Draft\\GoLang\\goProject`

## 相关链接

- [Augment MCP 官方文档](https://docs.augmentcode.com/context-services/mcp/quickstart-claude-code)
- [Auggie CLI 文档](https://docs.augmentcode.com/auggie-cli/overview)














