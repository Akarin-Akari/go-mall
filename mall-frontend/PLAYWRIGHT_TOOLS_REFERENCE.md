# 🛠️ Playwright 自动化工具参考手册

## 工具1：页面导航

### `browser_navigate_Playwright`
**功能：** 导航到指定URL

**参数：**
```json
{
  "url": "http://localhost:3000"
}
```

**示例：**
```
导航到首页
url: "http://localhost:3000"

导航到登录页
url: "http://localhost:3000/login"

导航到商品详情页
url: "http://localhost:3000/products/1"
```

**返回值：** 页面加载完成

---

## 工具2：元素点击

### `browser_click_Playwright`
**功能：** 点击页面上的元素

**参数：**
```json
{
  "element": "按钮描述",
  "ref": "元素引用",
  "button": "left|right|middle",
  "doubleClick": false,
  "modifiers": ["Shift", "Control"]
}
```

**示例：**
```
点击登录按钮
element: "登录按钮"
ref: "login-button"

点击添加到购物车
element: "添加到购物车按钮"
ref: "add-to-cart-btn"

双击商品卡片
element: "商品卡片"
ref: "product-card"
doubleClick: true
```

---

## 工具3：文本输入

### `browser_type_Playwright`
**功能：** 在输入框中输入文本

**参数：**
```json
{
  "element": "输入框描述",
  "ref": "元素引用",
  "text": "要输入的文本",
  "slowly": false,
  "submit": false
}
```

**示例：**
```
输入邮箱
element: "邮箱输入框"
ref: "email-input"
text: "test@example.com"

输入密码并提交
element: "密码输入框"
ref: "password-input"
text: "Test@123456"
submit: true
```

---

## 工具4：表单填充

### `browser_fill_form_Playwright`
**功能：** 批量填充表单字段

**参数：**
```json
{
  "fields": [
    {
      "name": "字段名称",
      "type": "textbox|checkbox|radio|combobox|slider",
      "ref": "元素引用",
      "value": "字段值"
    }
  ]
}
```

**示例：**
```
填充注册表单
fields: [
  { name: "邮箱", type: "textbox", ref: "email", value: "test@example.com" },
  { name: "密码", type: "textbox", ref: "password", value: "Test@123456" },
  { name: "确认密码", type: "textbox", ref: "confirm-password", value: "Test@123456" },
  { name: "同意条款", type: "checkbox", ref: "agree-terms", value: "true" }
]
```

---

## 工具5：下拉选择

### `browser_select_option_Playwright`
**功能：** 选择下拉菜单选项

**参数：**
```json
{
  "element": "下拉菜单描述",
  "ref": "元素引用",
  "values": ["选项值1", "选项值2"]
}
```

**示例：**
```
选择城市
element: "城市下拉菜单"
ref: "city-select"
values: ["北京"]

选择多个选项
element: "分类下拉菜单"
ref: "category-select"
values: ["电子产品", "图书"]
```

---

## 工具6：页面快照

### `browser_snapshot_Playwright`
**功能：** 获取页面的DOM结构和可访问性信息

**参数：** 无

**示例：**
```
获取当前页面的完整结构
调用 browser_snapshot_Playwright
```

**返回值：** 页面的DOM树和元素信息

---

## 工具7：截图

### `browser_take_screenshot_Playwright`
**功能：** 保存页面截图

**参数：**
```json
{
  "filename": "screenshot.png",
  "fullPage": false,
  "type": "png|jpeg"
}
```

**示例：**
```
保存首页截图
filename: "homepage.png"
fullPage: true
type: "png"

保存登录页截图
filename: "login-page.png"
fullPage: false
type: "jpeg"
```

---

## 工具8：控制台日志

### `browser_console_messages_Playwright`
**功能：** 获取浏览器控制台的日志和错误

**参数：**
```json
{
  "onlyErrors": false
}
```

**示例：**
```
获取所有控制台消息
调用 browser_console_messages_Playwright
onlyErrors: false

仅获取错误信息
调用 browser_console_messages_Playwright
onlyErrors: true
```

**返回值：** 控制台消息列表

---

## 工具9：网络请求

### `browser_network_requests_Playwright`
**功能：** 获取所有网络请求信息

**参数：** 无

**示例：**
```
获取所有网络请求
调用 browser_network_requests_Playwright
```

**返回值：** 请求列表，包含：
- 请求URL
- 请求方法（GET, POST等）
- 请求头
- 响应状态码
- 响应时间
- 响应体

---

## 工具10：等待条件

### `browser_wait_for_Playwright`
**功能：** 等待文本出现或消失

**参数：**
```json
{
  "text": "要等待的文本",
  "textGone": "要等待消失的文本",
  "time": 5
}
```

**示例：**
```
等待页面加载完成
text: "商品列表"
time: 5

等待加载动画消失
textGone: "加载中..."
time: 3

等待错误提示出现
text: "邮箱已被注册"
time: 2
```

---

## 完整测试流程示例

### 登录流程测试

```
1. 导航到登录页
   browser_navigate_Playwright
   url: "http://localhost:3000/login"

2. 获取页面结构
   browser_snapshot_Playwright

3. 填充登录表单
   browser_fill_form_Playwright
   fields: [
     { name: "邮箱", type: "textbox", ref: "email", value: "test@example.com" },
     { name: "密码", type: "textbox", ref: "password", value: "Test@123456" }
   ]

4. 点击登录按钮
   browser_click_Playwright
   element: "登录按钮"
   ref: "login-button"

5. 等待重定向
   browser_wait_for_Playwright
   text: "首页"
   time: 3

6. 检查控制台错误
   browser_console_messages_Playwright
   onlyErrors: true

7. 验证网络请求
   browser_network_requests_Playwright

8. 保存截图
   browser_take_screenshot_Playwright
   filename: "login-success.png"
```

---

## 常见问题

### Q1：如何等待元素加载？
**A：** 使用 `browser_wait_for_Playwright` 等待特定文本出现

### Q2：如何验证API请求？
**A：** 使用 `browser_network_requests_Playwright` 获取所有请求，检查URL、方法、状态码

### Q3：如何处理动态内容？
**A：** 使用 `browser_wait_for_Playwright` 等待内容加载，然后使用 `browser_snapshot_Playwright` 获取结构

### Q4：如何调试失败的测试？
**A：** 
1. 使用 `browser_take_screenshot_Playwright` 保存截图
2. 使用 `browser_console_messages_Playwright` 检查错误
3. 使用 `browser_snapshot_Playwright` 获取页面结构
4. 使用 `browser_network_requests_Playwright` 检查API调用

---

## 性能优化建议

1. **使用明确的等待条件** - 不要使用固定延迟
2. **并行执行独立测试** - 提高测试效率
3. **缓存测试数据** - 避免重复创建
4. **监控网络请求** - 识别性能瓶颈
5. **定期清理测试数据** - 保持数据库整洁

