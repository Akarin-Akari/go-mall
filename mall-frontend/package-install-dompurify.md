# DOMPurify 安装说明

## 安装命令

```bash
# 安装DOMPurify
npm install dompurify

# 安装TypeScript类型定义
npm install --save-dev @types/dompurify
```

## 安装后需要执行的步骤

1. 在项目根目录执行上述安装命令
2. 更新 `src/utils/xssProtection.ts` 文件，替换自实现的HTML清理器
3. 更新 `src/components/common/SecureInput.tsx` 文件，使用新的清理器
4. 测试所有输入组件的功能

## 注意事项

- DOMPurify 是一个成熟的HTML清理库，比自实现的清理器更安全
- 需要配置适当的白名单标签和属性
- 在生产环境中测试所有输入场景
