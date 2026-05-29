# Augment MCP 安装和配置脚本
# 在 Windows PowerShell 中执行此脚本

Write-Host "=== Augment MCP 安装和配置脚本 ===" -ForegroundColor Green

# 步骤 1: 检查 Node.js
Write-Host "`n[1/4] 检查 Node.js 安装..." -ForegroundColor Yellow
$nodeVersion = node --version 2>$null
if ($nodeVersion) {
    Write-Host "✓ Node.js 已安装: $nodeVersion" -ForegroundColor Green
} else {
    Write-Host "✗ Node.js 未安装，请先安装 Node.js" -ForegroundColor Red
    exit 1
}

# 步骤 2: 安装 Auggie CLI
Write-Host "`n[2/4] 安装 Auggie CLI..." -ForegroundColor Yellow
npm install -g @augmentcode/auggie@prerelease
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Auggie CLI 安装成功" -ForegroundColor Green
} else {
    Write-Host "✗ Auggie CLI 安装失败" -ForegroundColor Red
    exit 1
}

# 步骤 3: 验证安装
Write-Host "`n[3/4] 验证 Auggie 安装..." -ForegroundColor Yellow
$auggieVersion = auggie --version 2>$null
if ($auggieVersion) {
    Write-Host "✓ Auggie 已安装: $auggieVersion" -ForegroundColor Green
} else {
    Write-Host "✗ Auggie 未找到，请检查 PATH 环境变量" -ForegroundColor Red
    exit 1
}

# 步骤 4: 配置 MCP 服务器
Write-Host "`n[4/4] 配置 MCP 服务器..." -ForegroundColor Yellow
Write-Host "请选择配置方式:" -ForegroundColor Cyan
Write-Host "1. 用户级别配置（推荐，所有项目可用）" -ForegroundColor White
Write-Host "2. 项目级别配置（仅当前项目）" -ForegroundColor White
Write-Host "3. 跳过配置（手动配置）" -ForegroundColor White

$choice = Read-Host "请输入选项 (1-3)"

if ($choice -eq "1") {
    Write-Host "`n执行用户级别配置..." -ForegroundColor Yellow
    $workspace = Read-Host "请输入工作区路径（留空使用默认）"
    if ($workspace) {
        $workspaceEscaped = $workspace -replace '\\', '\\'
        claude mcp add-json auggie-mcp --scope user "{\"type\":\"stdio\",\"command\":\"auggie\",\"args\":[\"-w\",\"$workspaceEscaped\",\"--mcp\"]}"
    } else {
        claude mcp add-json auggie-mcp --scope user '{"type":"stdio","command":"auggie","args":["--mcp"]}'
    }
    Write-Host "✓ 用户级别配置完成" -ForegroundColor Green
} elseif ($choice -eq "2") {
    Write-Host "`n执行项目级别配置..." -ForegroundColor Yellow
    $workspace = Read-Host "请输入工作区路径（留空使用默认）"
    if ($workspace) {
        $workspaceEscaped = $workspace -replace '\\', '\\'
        claude mcp add-json auggie-mcp --scope project "{\"type\":\"stdio\",\"command\":\"auggie\",\"args\":[\"-w\",\"$workspaceEscaped\",\"--mcp\"]}"
    } else {
        claude mcp add-json auggie-mcp --scope project '{"type":"stdio","command":"auggie","args":["--mcp"]}'
    }
    Write-Host "✓ 项目级别配置完成" -ForegroundColor Green
} else {
    Write-Host "跳过自动配置，请手动配置" -ForegroundColor Yellow
}

Write-Host "`n=== 配置完成 ===" -ForegroundColor Green
Write-Host "下一步操作:" -ForegroundColor Cyan
Write-Host "1. 运行 'auggie login' 进行身份验证" -ForegroundColor White
Write-Host "2. 重启 Cursor IDE" -ForegroundColor White
Write-Host "3. 检查 MCP 服务器是否正常加载" -ForegroundColor White
















