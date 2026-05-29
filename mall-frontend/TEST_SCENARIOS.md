# 🧪 Mall-Go 自动化测试场景集合

## 📋 目录

1. [场景1：首页加载测试](#场景1首页加载测试)
2. [场景2：用户注册流程](#场景2用户注册流程)
3. [场景3：用户登录流程](#场景3用户登录流程)
4. [场景4：商品浏览](#场景4商品浏览)
5. [场景5：购物车操作](#场景5购物车操作)
6. [场景6：性能检查](#场景6性能检查)

---

## 场景1：首页加载测试

**目标：** 验证首页能否正常加载，检查页面结构和性能

**步骤：**

```
1️⃣ 导航到首页
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000"

2️⃣ 等待页面加载
   工具：browser_wait_for_Playwright
   参数：text = "商品" 或 "首页"
   参数：time = 5

3️⃣ 获取页面结构
   工具：browser_snapshot_Playwright

4️⃣ 检查控制台错误
   工具：browser_console_messages_Playwright
   参数：onlyErrors = true

5️⃣ 获取网络请求
   工具：browser_network_requests_Playwright

6️⃣ 保存截图
   工具：browser_take_screenshot_Playwright
   参数：filename = "homepage-load.png"
   参数：fullPage = true
```

**预期结果：**
- ✅ 页面加载成功
- ✅ 没有控制台错误
- ✅ 所有API请求返回200
- ✅ 页面加载时间 < 3秒

---

## 场景2：用户注册流程

**目标：** 测试用户注册功能，验证表单验证和API调用

**测试数据：**
```
邮箱: newuser@example.com
密码: Test@123456
确认密码: Test@123456
```

**步骤：**

```
1️⃣ 导航到注册页
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000/register"

2️⃣ 等待页面加载
   工具：browser_wait_for_Playwright
   参数：text = "注册" 或 "邮箱"
   参数：time = 3

3️⃣ 获取页面结构
   工具：browser_snapshot_Playwright

4️⃣ 填充注册表单
   工具：browser_fill_form_Playwright
   参数：fields = [
     { name: "邮箱", type: "textbox", ref: "email", value: "newuser@example.com" },
     { name: "密码", type: "textbox", ref: "password", value: "Test@123456" },
     { name: "确认密码", type: "textbox", ref: "confirm-password", value: "Test@123456" }
   ]

5️⃣ 点击注册按钮
   工具：browser_click_Playwright
   参数：element = "注册按钮"
   参数：ref = "register-button"

6️⃣ 等待响应
   工具：browser_wait_for_Playwright
   参数：text = "注册成功" 或 "登录"
   参数：time = 3

7️⃣ 检查网络请求
   工具：browser_network_requests_Playwright

8️⃣ 检查错误
   工具：browser_console_messages_Playwright
   参数：onlyErrors = true
```

**预期结果：**
- ✅ 表单验证通过
- ✅ API请求：POST /api/v1/users/register 返回200
- ✅ 显示成功提示
- ✅ 重定向到登录页或首页

---

## 场景3：用户登录流程

**目标：** 测试用户登录功能，验证Token保存和重定向

**测试数据：**
```
邮箱: test@example.com
密码: Test@123456
```

**步骤：**

```
1️⃣ 导航到登录页
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000/login"

2️⃣ 等待页面加载
   工具：browser_wait_for_Playwright
   参数：text = "登录" 或 "邮箱"
   参数：time = 3

3️⃣ 填充登录表单
   工具：browser_fill_form_Playwright
   参数：fields = [
     { name: "邮箱", type: "textbox", ref: "email", value: "test@example.com" },
     { name: "密码", type: "textbox", ref: "password", value: "Test@123456" }
   ]

4️⃣ 点击登录按钮
   工具：browser_click_Playwright
   参数：element = "登录按钮"
   参数：ref = "login-button"

5️⃣ 等待重定向
   工具：browser_wait_for_Playwright
   参数：text = "首页" 或 "商品"
   参数：time = 3

6️⃣ 验证网络请求
   工具：browser_network_requests_Playwright

7️⃣ 检查错误
   工具：browser_console_messages_Playwright
   参数：onlyErrors = true

8️⃣ 保存截图
   工具：browser_take_screenshot_Playwright
   参数：filename = "login-success.png"
```

**预期结果：**
- ✅ API请求：POST /api/v1/users/login 返回200
- ✅ Token保存到localStorage
- ✅ 重定向到首页
- ✅ 显示用户信息

---

## 场景4：商品浏览

**目标：** 测试商品列表和详情页面

**步骤：**

```
1️⃣ 导航到首页
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000"

2️⃣ 等待商品列表加载
   工具：browser_wait_for_Playwright
   参数：text = "商品" 或 "¥"
   参数：time = 5

3️⃣ 获取页面结构
   工具：browser_snapshot_Playwright

4️⃣ 点击第一个商品
   工具：browser_click_Playwright
   参数：element = "商品卡片"
   参数：ref = "product-card-1"

5️⃣ 等待详情页加载
   工具：browser_wait_for_Playwright
   参数：text = "商品详情" 或 "添加到购物车"
   参数：time = 3

6️⃣ 获取详情页结构
   工具：browser_snapshot_Playwright

7️⃣ 验证网络请求
   工具：browser_network_requests_Playwright

8️⃣ 保存截图
   工具：browser_take_screenshot_Playwright
   参数：filename = "product-detail.png"
```

**预期结果：**
- ✅ 商品列表加载成功
- ✅ API请求：GET /api/v1/products 返回200
- ✅ 点击商品后导航到详情页
- ✅ API请求：GET /api/v1/products/:id 返回200

---

## 场景5：购物车操作

**目标：** 测试添加商品到购物车

**前置条件：** 用户已登录

**步骤：**

```
1️⃣ 导航到商品详情页
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000/products/1"

2️⃣ 等待页面加载
   工具：browser_wait_for_Playwright
   参数：text = "添加到购物车"
   参数：time = 3

3️⃣ 输入购买数量
   工具：browser_type_Playwright
   参数：element = "数量输入框"
   参数：ref = "quantity-input"
   参数：text = "2"

4️⃣ 点击添加到购物车
   工具：browser_click_Playwright
   参数：element = "添加到购物车按钮"
   参数：ref = "add-to-cart-btn"

5️⃣ 等待成功提示
   工具：browser_wait_for_Playwright
   参数：text = "已添加到购物车" 或 "成功"
   参数：time = 2

6️⃣ 验证网络请求
   工具：browser_network_requests_Playwright

7️⃣ 导航到购物车页面
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000/cart"

8️⃣ 验证商品在购物车中
   工具：browser_snapshot_Playwright
```

**预期结果：**
- ✅ API请求：POST /api/v1/cart/add 返回200
- ✅ 显示成功提示
- ✅ 购物车页面显示添加的商品
- ✅ 数量和价格正确

---

## 场景6：性能检查

**目标：** 检查页面加载性能和API响应时间

**步骤：**

```
1️⃣ 导航到首页
   工具：browser_navigate_Playwright
   参数：url = "http://localhost:3000"

2️⃣ 等待页面加载
   工具：browser_wait_for_Playwright
   参数：text = "商品"
   参数：time = 5

3️⃣ 获取所有网络请求
   工具：browser_network_requests_Playwright

4️⃣ 获取控制台日志
   工具：browser_console_messages_Playwright

5️⃣ 分析性能指标
   - 计算总加载时间
   - 计算最慢的API请求
   - 统计错误请求数量
   - 检查是否有警告信息
```

**性能指标：**
```
✅ 页面加载时间 < 3秒
✅ API响应时间 < 1秒
✅ 首屏渲染时间 < 2秒
✅ 没有控制台错误
✅ 没有未捕获的异常
```

---

## 🚀 快速开始

### 1. 启动服务
```bash
# 终端1：启动后端
cd mall-go
go run cmd/server/main.go

# 终端2：启动前端
cd mall-frontend
npm run dev
```

### 2. 执行测试
告诉我你想执行的测试场景，例如：
- "执行场景1：首页加载测试"
- "执行场景3：用户登录流程"
- "执行场景5：购物车操作"

### 3. 查看结果
我会：
- 执行所有测试步骤
- 收集测试数据
- 生成测试报告
- 保存截图和日志

---

## 📊 测试报告模板

```
测试场景：[场景名称]
执行时间：[时间]
总耗时：[耗时]

✅ 通过的步骤：[数量]
❌ 失败的步骤：[数量]
⚠️ 警告：[数量]

详细结果：
[每个步骤的结果]

性能指标：
- 页面加载时间：[时间]
- API响应时间：[时间]
- 错误数量：[数量]

截图：
[保存的截图列表]
```

