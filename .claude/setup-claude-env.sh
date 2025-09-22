#!/bin/bash
# Claude Code 环境变量配置脚本
# 设置Claude Code的全局配置路径

GLOBAL_CLAUDE_PATH="$HOME/.claude"

set_claude_environment() {
    local permanent=${1:-false}
    
    # 定义环境变量
    declare -A env_vars=(
        ["CLAUDE_CONFIG_PATH"]="$GLOBAL_CLAUDE_PATH"
        ["CLAUDE_GLOBAL_RULES"]="$GLOBAL_CLAUDE_PATH/rules/global-claude-rules.md"
        ["CLAUDE_GLOBAL_MEMORIES"]="$GLOBAL_CLAUDE_PATH/memories/global-claude-memories.md"
        ["CLAUDE_MCP_CONFIG"]="$GLOBAL_CLAUDE_PATH/settings.local.json"
    )
    
    for var_name in "${!env_vars[@]}"; do
        local var_value="${env_vars[$var_name]}"
        
        if [ "$permanent" = "true" ]; then
            # 添加到shell配置文件
            local shell_config=""
            if [ -n "$BASH_VERSION" ]; then
                shell_config="$HOME/.bashrc"
            elif [ -n "$ZSH_VERSION" ]; then
                shell_config="$HOME/.zshrc"
            else
                shell_config="$HOME/.profile"
            fi
            
            if ! grep -q "export $var_name=" "$shell_config" 2>/dev/null; then
                echo "export $var_name=\"$var_value\"" >> "$shell_config"
                echo "已添加永久环境变量到 $shell_config: $var_name"
            else
                echo "环境变量已存在于 $shell_config: $var_name"
            fi
        else
            # 设置当前会话环境变量
            export "$var_name"="$var_value"
            echo "已设置会话环境变量: $var_name = $var_value"
        fi
    done
}

case "${1:-help}" in
    --permanent)
        echo "设置永久环境变量..."
        set_claude_environment true
        echo "请重新加载shell配置或重启终端以使环境变量生效"
        ;;
    --session)
        echo "设置当前会话环境变量..."
        set_claude_environment false
        ;;
    *)
        echo "Claude Code 环境变量配置脚本"
        echo "用法："
        echo "  $0 --permanent  # 设置永久环境变量"
        echo "  $0 --session    # 设置当前会话环境变量"
        ;;
esac
