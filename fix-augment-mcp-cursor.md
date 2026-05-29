# 修复 Cursor Augment MCP "No server info found" 错误

## 问题分析

错误信息 `No server info found` 表明 Cursor 无法找到 MCP 服务器的配置信息。这可能是因为：

1. **配置方式不正确**：直接在 `.claude/settings.local.json` 中配置可能不是 Cursor 识别 MCP 服务器的正确方式
2. **auggie 未安装**：需要先安装 `@augmentcode/auggie@prerelease`
3. **未使用官方配置命令**：应该使用 `claude mcp add-json` 命令来注册 MCP 服务器

## 解决方案

### 步骤 1：安装 Auggie CLI

在 Windows PowerShell 或 CMD 中执行：

```powershell
npm install -g @augmentcode/auggie@prerelease
```

### 步骤 2：登录 Augment

```powershell
auggie login
```

这会打开浏览器窗口进行身份验证。

### 步骤 3：使用官方命令配置 MCP 服务器

根据 [Augment 官方文档](https://docs.augmentcode.com/context-services/mcp/quickstart-claude-code)，应该使用以下命令：

**用户级别配置**（推荐，所有项目可用）：

```powershell
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["--mcp"]}'
```

**项目级别配置**（仅当前项目）：

```powershell
claude mcp add-json auggie-mcp --scope project '{"type":"stdio","command":"auggie","args":["--mcp"]}'
```

**指定工作区路径**（如果需要）：

```powershell
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["-w","E:\\Workspace_Draft\\GoLang\\goProject","--mcp"]}'
```

### 步骤 4：验证配置

重启 Cursor 后，检查 MCP 服务器是否正常加载。可以在 Cursor 的命令面板中搜索 "MCP" 查看服务器状态。

## 替代方案：手动编辑配置文件

如果 `claude mcp add-json` 命令不可用，可以尝试手动编辑 Cursor 的 MCP 配置文件。

Cursor 的 MCP 配置文件通常位于：

- Windows: `%APPDATA%\Cursor\User\globalStorage\mcp.json`
- 或者: `%APPDATA%\Cursor\User\settings.json`

配置文件格式应该是：

```json
{
  "mcpServers": {
    "auggie-mcp": {
      "type": "stdio",
      "command": "auggie",
      "args": ["--mcp"]
    }
  }
}
```

或者使用 Windows 命令：

```json
{
  "mcpServers": {
    "auggie-mcp": {
      "type": "stdio",
      "command": "cmd",
      "args": ["/c", "auggie", "--mcp"]
    }
  }
}
```

## 注意事项

1. **确保 auggie 在 PATH 中**：安装后确保 `auggie` 命令可以在命令行中直接执行
2. **重启 Cursor**：修改配置后必须重启 Cursor 才能生效
3. **检查日志**：如果仍有问题，查看 Cursor 的日志文件获取更详细的错误信息

## 验证步骤

1. 在命令行中执行 `auggie --version` 确认安装成功
2. 执行 `auggie login` 完成身份验证
3. 使用 `claude mcp add-json` 命令添加服务器配置
4. 重启 Cursor
5. 检查是否还有 "No server info found" 错误














