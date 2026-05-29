# Augment MCP 修复指南

## 问题诊断

根据日志分析，问题在于：

- Cursor 尝试在 WSL2 环境中执行 `auggie --mcp -m default -w ""`
- `auggie` 脚本位于 Windows 路径：`/mnt/c/Users/Akari/AppData/Roaming/npm/auggie`
- 脚本需要 `node` 命令，但在 WSL2 中找不到 Node.js

## 解决方案

### 方案 1：在 WSL2 中安装 Node.js（推荐）

如果 WSL2 中没有安装 Node.js，需要先安装：

```bash
# 方法 1：使用 NodeSource 官方源（推荐）
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# 方法 2：使用 Ubuntu 官方源
sudo apt update
sudo apt install -y nodejs npm

# 验证安装
node --version
npm --version
```

### 方案 2：修改 MCP 配置使用 Windows 命令

如果不想在 WSL2 中安装 Node.js，可以修改 Cursor 的 MCP 配置，让它在 Windows 环境中执行 `auggie`。

需要找到 Cursor 的 MCP 配置文件（通常在 `%APPDATA%\Cursor\User\globalStorage\mcp.json` 或类似位置），将配置修改为：

```json
{
  "augment-context-engine": {
    "command": "cmd.exe",
    "args": ["/c", "auggie", "--mcp", "-m", "default", "-w", ""]
  }
}
```

### 方案 3：在 WSL2 中安装 auggie

如果 Node.js 已安装，可以在 WSL2 中安装 `auggie`：

```bash
npm install -g auggie
```

然后确保 Cursor 的 MCP 配置指向正确的 `auggie` 路径。

## 验证修复

修复后，重启 Cursor 并检查日志，应该不再出现 `exec: node: not found` 错误。

## 注意事项

- 如果使用方案 1，确保 WSL2 中的 Node.js 版本与 Windows 中的版本兼容
- 如果使用方案 2，确保 Windows 的 PATH 中包含 Node.js 和 auggie
- 修复后需要重启 Cursor 才能生效















