#!/bin/bash
# Claude Code 配置同步脚本
# 在全局配置和项目配置之间同步

DIRECTION="${1:-ToGlobal}"  # ToGlobal, ToProject, Bidirectional
PROJECT_PATH="${2:-.}"
FORCE="${3:-false}"

GLOBAL_PATH="$HOME/.claude"
PROJECT_CLAUDE_PATH="$PROJECT_PATH/.claude"

sync_config_to_global() {
    local project_path="$1"
    local force_overwrite="$2"
    
    echo "同步项目配置到全局..."
    
    declare -A files_to_sync=(
        ["settings.local.json"]="settings.local.json"
        ["claude-rules.md"]="rules/project-claude-rules.md"
        ["claude-memories.md"]="memories/project-claude-memories.md"
    )
    
    for source_file in "${!files_to_sync[@]}"; do
        local source_path="$PROJECT_CLAUDE_PATH/$source_file"
        local target_path="$GLOBAL_PATH/${files_to_sync[$source_file]}"
        local target_dir=$(dirname "$target_path")
        
        if [ -f "$source_path" ]; then
            mkdir -p "$target_dir"
            
            if [ ! -f "$target_path" ] || [ "$force_overwrite" = "true" ]; then
                cp "$source_path" "$target_path"
                echo "✅ 已同步: $source_file"
            else
                echo "⚠️  目标文件已存在，跳过: $source_file (使用 --force 强制覆盖)"
            fi
        fi
    done
}

sync_config_to_project() {
    local project_path="$1"
    local force_overwrite="$2"
    
    echo "同步全局配置到项目..."
    
    mkdir -p "$PROJECT_CLAUDE_PATH"
    
    declare -A files_to_sync=(
        ["settings.local.json"]="settings.local.json"
        ["rules/global-claude-rules.md"]="claude-rules.md"
        ["memories/global-claude-memories.md"]="claude-memories.md"
    )
    
    for source_file in "${!files_to_sync[@]}"; do
        local source_path="$GLOBAL_PATH/$source_file"
        local target_path="$PROJECT_CLAUDE_PATH/${files_to_sync[$source_file]}"
        
        if [ -f "$source_path" ]; then
            if [ ! -f "$target_path" ] || [ "$force_overwrite" = "true" ]; then
                cp "$source_path" "$target_path"
                echo "✅ 已同步: ${files_to_sync[$source_file]}"
            else
                echo "⚠️  目标文件已存在，跳过: ${files_to_sync[$source_file]} (使用 --force 强制覆盖)"
            fi
        fi
    done
}

# 主同步逻辑
case "$DIRECTION" in
    ToGlobal)
        sync_config_to_global "$PROJECT_PATH" "$FORCE"
        ;;
    ToProject)
        sync_config_to_project "$PROJECT_PATH" "$FORCE"
        ;;
    Bidirectional)
        echo "双向同步..."
        sync_config_to_global "$PROJECT_PATH" "$FORCE"
        sync_config_to_project "$PROJECT_PATH" "$FORCE"
        ;;
    *)
        echo "未知同步方向: $DIRECTION"
        echo "可用选项: ToGlobal, ToProject, Bidirectional"
        exit 1
        ;;
esac
