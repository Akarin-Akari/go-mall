# 🤖 Mall-Go 浏览器自动化测试完整指南

## 📌 快速概览

本指南展示如何使用 Playwright 浏览器自动化工具对 Mall-Go 项目进行端到端测试。

**服务配置：**
- 前端服务：`http://localhost:3000`（Next.js）
- 后端服务：`http://localhost:8081`（Go Gin）
- API基础URL：`http://localhost:8081/api/v1`

---

## ✅ 可用的自动化工具

| 工具 | 功能 | 用途 |
|------|------|------|
| `browser_navigate_Playwright` | 页面导航 | 访问指定URL |
| `browser_click_Playwright` | 元素点击 | 点击按钮、链接等 |
| `browser_type_Playwright` | 文本输入 | 填充表单字段 |
| `browser_snapshot_Playwright` | 页面快照 | 获取页面DOM结构 |
| `browser_take_screenshot_Playwright` | 截图 | 保存页面截图 |
| `browser_console_messages_Playwright` | 控制台日志 | 捕获JS错误和日志 |
| `browser_network_requests_Playwright` | 网络请求 | 监控API调用 |
| `browser_fill_form_Playwright` | 表单填充 | 批量填充表单 |
| `browser_select_option_Playwright` | 下拉选择 | 选择下拉菜单选项 |
| `browser_wait_for_Playwright` | 等待条件 | 等待文本出现/消失 |

---

## 🎯 测试场景

### 场景1：首页加载测试
```
1. 导航到 http://localhost:3000
2. 等待页面加载完成
3. 获取页面快照验证结构
4. 检查控制台是否有错误
5. 验证网络请求状态
```

### 场景2：用户注册流程
```
1. 导航到注册页面
2. 填充注册表单（邮箱、密码、确认密码）
3. 点击注册按钮
4. 验证API请求（POST /api/v1/users/register）
5. 检查响应状态和错误信息
```

### 场景3：用户登录流程
```
1. 导航到登录页面
2. 输入邮箱和密码
3. 点击登录按钮
4. 验证API请求（POST /api/v1/users/login）
5. 检查Token是否保存到localStorage
6. 验证重定向到首页
```

### 场景4：商品浏览
```
1. 导航到首页
2. 等待商品列表加载
3. 获取商品卡片元素
4. 点击商品卡片
5. 验证导航到商品详情页
6. 检查API请求（GET /api/v1/products/:id）
```

### 场景5：购物车操作
```
1. 登录用户
2. 浏览商品
3. 点击"添加到购物车"按钮
4. 验证API请求（POST /api/v1/cart/add）
5. 检查购物车数量更新
6. 验证购物车页面显示正确的商品
```

### 场景6：性能和错误检查
```
1. 监控所有网络请求
2. 计算页面加载时间
3. 检查API响应时间
4. 捕获所有控制台错误
5. 验证没有未捕获的异常
```

---

## 🔧 使用方法

### 方法1：单个工具测试

**示例：导航到首页**
```
调用 browser_navigate_Playwright
参数：url = "http://localhost:3000"
```

**示例：等待页面加载**
```
调用 browser_wait_for_Playwright
参数：text = "商品列表"
参数：time = 5
```

**示例：获取页面快照**
```
调用 browser_snapshot_Playwright
```

**示例：检查控制台错误**
```
调用 browser_console_messages_Playwright
参数：onlyErrors = true
```

**示例：获取网络请求**
```
调用 browser_network_requests_Playwright
```

### 方法2：组合工具进行完整流程

**完整的登录流程测试：**
```
1. browser_navigate_Playwright → 访问登录页
   url: "http://localhost:3000/login"

2. browser_snapshot_Playwright → 获取页面结构

3. browser_fill_form_Playwright → 填充登录表单
   fields: [
     { name: "邮箱", type: "textbox", ref: "email", value: "test@example.com" },
     { name: "密码", type: "textbox", ref: "password", value: "Test@123456" }
   ]

4. browser_click_Playwright → 点击登录按钮
   element: "登录按钮"
   ref: "login-button"

5. browser_wait_for_Playwright → 等待重定向
   text: "首页"
   time: 3

6. browser_console_messages_Playwright → 检查错误
   onlyErrors: true

7. browser_network_requests_Playwright → 验证API调用
```

---

## 📊 测试数据

### 测试账户
```
邮箱: test@example.com
密码: Test@123456
```

### 测试商品
```
商品1: ID=1, 名称="iPhone 15"
商品2: ID=2, 名称="MacBook Pro"
商品3: ID=3, 名称="iPad Air"
```

---

## 🐛 故障排查

### 问题1：无法连接到localhost:3000
**原因：** 前端服务未启动
**解决方案：**
```bash
cd mall-frontend
npm run dev
```

### 问题2：无法连接到localhost:8081
**原因：** 后端服务未启动
**解决方案：**
```bash
cd mall-go
go run cmd/server/main.go
```

### 问题3：API请求返回401
**原因：** Token过期或无效
**解决方案：**
- 重新登录获取新Token
- 检查Authorization请求头格式
- 验证Token是否正确保存

### 问题4：页面元素找不到
**原因：** 元素选择器不正确或页面未加载
**解决方案：**
- 使用 browser_snapshot_Playwright 获取页面结构
- 检查元素选择器是否正确
- 增加等待时间

---

## 📈 性能指标

### 关键指标
- 页面加载时间 < 3秒
- API响应时间 < 1秒
- 首屏渲染时间 < 2秒
- 无控制台错误

### 监控方法
```
1. 使用 browser_network_requests_Playwright 获取所有请求
2. 计算每个请求的响应时间
3. 统计错误请求数量
4. 使用 browser_console_messages_Playwright 检查错误
```

---

## 🎓 最佳实践

1. **按顺序执行测试** - 避免并发导致的状态混乱
2. **使用明确的等待条件** - 不要使用固定延迟
3. **验证关键数据** - 检查响应数据的完整性
4. **捕获失败信息** - 保存截图和日志用于调试
5. **清理测试数据** - 测试完成后清理创建的数据

---

## 📝 下一步

1. **启动服务**
   ```bash
   # 终端1：启动后端
   cd mall-go && go run cmd/server/main.go
   
   # 终端2：启动前端
   cd mall-frontend && npm run dev
   ```

2. **告诉我你想测试的功能**
   - 我会调用相应的工具进行自动化测试
   - 收集测试结果和错误信息
   - 生成完整的测试报告

3. **常见测试请求**
   - "测试首页加载"
   - "测试用户注册流程"
   - "测试登录功能"
   - "测试商品浏览"
   - "测试购物车操作"
   - "检查页面性能"
   - "捕获所有错误信息"

