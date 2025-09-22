# MCP服务器配置清单

## 已配置的MCP服务器（25+个）

### 核心服务
1. **context7** - 上下文增强引擎
2. **GitHub** - 代码仓库管理（已配置token）
3. **fetch** - 网络请求处理
4. **puppeteer** - 浏览器自动化

### 开发工具
5. **redis** - 数据库操作
6. **elasticsearch-mcp-server** - 搜索分析
7. **LeetCode** - 编程练习（CN站点）
8. **bingcn** - 中文搜索
9. **console-ninja** - 调试工具

### AI增强
10. **MiniMax** - AI模型服务
11. **magic** - AI魔法工具
12. **mem0-memory-mcp** - 记忆管理
13. **think-mcp-server** - 思维增强
14. **Sequential Thinking MCP** - 序列思维
15. **server-sequential-thinking** - 序列思维服务

### 设计与文档
16. **Framelink Figma MCP** - 设计工具（已配置API key）
17. **@master/mastergo-magic-mcp** - 设计协作
18. **firecrawl-mcp** - 网页抓取（已配置API key）
19. **mcp-server-firecrawl** - 网页抓取服务

### 系统控制
20. **desktop-commander** - 桌面命令
21. **ClaudeCommander MCP** - Claude命令控制
22. **mcp-browserbase** - 浏览器基础服务

### 数据服务
23. **neon** - 数据库服务
24. **exa** - 搜索增强
25. **servers** - 服务器管理

### 专业工具
26. **spec-workflow** - 规范工作流
27. **hackernews_composio** - 新闻聚合
28. **composio_composio** - 组合服务

## 配置状态检查

### ✅ 已正确配置
- GitHub（token已设置）
- Firecrawl（API key已设置）
- Figma（API key已设置）
- LeetCode（CN站点）
- Redis（本地连接）

### ⚠️ 需要配置API key
- MiniMax（需要设置MINIMAX_API_KEY）
- Elasticsearch（需要ES_API_KEY和ES_URL）
- Magic（需要API key）
- MasterGo（需要MG_MCP_TOKEN）

### 🔧 需要路径配置
- console-ninja（需要正确的mcp路径）
- debug（需要mcp-debug.js路径）
- spec-workflow（需要项目路径配置）

## 权限配置
所有必要的Bash命令权限已配置，包括：
- Maven编译命令
- Git操作
- 文件系统操作
- 网络请求
- 数据库连接
- 系统服务管理

## 注意事项
1. 某些MCP服务器需要有效的API密钥才能正常工作
2. 路径配置需要根据实际环境调整
3. 网络连接问题可能影响远程MCP服务
4. 定期检查MCP服务器的可用性和更新