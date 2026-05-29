# 修复 Cursor Augment MCP "No server info found" 错误

## 问题分析

错误信息 `No server info found` 表明 Cursor 无法找到 MCP 服务器的配置信息。

**重要发现**：`.claude/settings.local.json` 是 Claude Desktop 的配置文件，**不是 Cursor 的配置文件**。Cursor 使用不同的配置方式。

## 解决方案

### 方案 1：使用官方 CLI 命令配置（推荐）

根据 [Augment 官方文档](https://docs.augmentcode.com/context-services/mcp/quickstart-claude-code)，应该使用 `claude mcp add-json` 命令：

#### 步骤 1：安装 Auggie CLI

在 Windows PowerShell 或 CMD 中执行：

```powershell
npm install -g @augmentcode/auggie@prerelease
```

#### 步骤 2：登录 Augment

```powershell
auggie login
```

这会打开浏览器窗口进行身份验证。

#### 步骤 3：配置 MCP 服务器

**用户级别配置**（推荐，所有项目可用）：

```powershell
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["--mcp"]}'
```

**指定工作区路径**（如果需要）：

```powershell
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["-w","E:\\Workspace_Draft\\GoLang\\goProject","--mcp"]}'
```

**项目级别配置**（仅当前项目）：

```powershell
claude mcp add-json auggie-mcp --scope project '{"type":"stdio","command":"auggie","args":["--mcp"]}'
```

#### 步骤 4：重启 Cursor

配置完成后，**必须重启 Cursor** 才能生效。

### 方案 2：手动编辑 Cursor 配置文件

如果 `claude mcp add-json` 命令不可用，可以手动编辑配置文件。

#### 查找 Cursor 配置文件位置

Cursor 的 MCP 配置文件可能位于以下位置之一：

1. `%APPDATA%\Cursor\User\globalStorage\mcp.json`
2. `%APPDATA%\Cursor\User\settings.json`
3. 项目目录下的 `.cursor/mcp.json`

#### 配置文件格式

创建或编辑配置文件，添加以下内容：

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

**Windows 环境**（如果 auggie 不在 PATH 中）：

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

**指定工作区路径**：

```json
{
  "mcpServers": {
    "auggie-mcp": {
      "type": "stdio",
      "command": "auggie",
      "args": ["-w", "E:\\Workspace_Draft\\GoLang\\goProject", "--mcp"]
    }
  }
}
```

### 方案 3：使用自动化脚本

我已经创建了一个 PowerShell 脚本 `install-and-configure-augment-mcp.ps1`，可以自动完成安装和配置。

在 PowerShell 中执行：

```powershell
.\install-and-configure-augment-mcp.ps1
```

## 验证步骤

1. **检查 auggie 安装**：

   ```powershell
   auggie --version
   ```

2. **检查 auggie 登录状态**：

   ```powershell
   auggie token print
   ```

3. **重启 Cursor**：修改配置后必须重启

4. **检查 MCP 服务器状态**：
   - 在 Cursor 中打开命令面板（`Ctrl+Shift+P`）
   - 搜索 "MCP" 查看服务器状态
   - 或者查看日志，确认不再出现 "No server info found" 错误

## 常见问题

### Q1: `claude mcp add-json` 命令不存在？

**A**: 这个命令是 Claude CLI 的一部分。如果不可用，使用方案 2 手动编辑配置文件。

### Q2: auggie 命令找不到？

**A**: 确保 auggie 已全局安装，并且 npm 的全局 bin 目录在 PATH 中：

```powershell
npm config get prefix
# 将返回的路径下的 node_modules\.bin 添加到 PATH
```

### Q3: 配置后仍然显示 "No server info found"？

**A**:

1. 确认配置文件位置正确
2. 确认配置文件格式正确（JSON 格式）
3. 确认 auggie 命令可以执行
4. 重启 Cursor
5. 检查 Cursor 的日志文件获取详细错误信息

### Q4: 如何在非交互式环境中配置？

参考官方文档的"Advanced: Non-Interactive Setup"部分，使用环境变量：

```powershell
# 获取 token
auggie token print

# 配置环境变量
$env:AUGMENT_API_TOKEN = "your-access-token"
$env:AUGMENT_API_URL = "your-tenant-url"

# 配置 MCP
claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["--mcp"],"env":{"AUGMENT_API_TOKEN":"your-access-token","AUGMENT_API_URL":"your-tenant-url"}}'
```

## 相关链接

- [Augment MCP 官方文档](https://docs.augmentcode.com/context-services/mcp/quickstart-claude-code)
- [Auggie CLI 文档](https://docs.augmentcode.com/auggie-cli/overview)















