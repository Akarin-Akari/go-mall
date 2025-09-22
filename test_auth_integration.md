# Mall-Go 用户认证功能联调测试指南

## 🚀 测试环境准备

### 1. 启动后端Go服务器
```bash
# 在 mall-go 目录下执行
cd mall-go
go run test_simple_server.go
```

### 2. 启动前端开发服务器
```bash
# 在 mall-frontend 目录下执行
cd mall-frontend
npm run dev
```

## 🔧 后端API测试

### 1. 健康检查接口测试
```bash
curl http://localhost:8081/health
```
**期望响应：**
```json
{
  "code": 200,
  "message": "Mall Go API is running",
  "data": {
    "status": "ok",
    "time": "2025-01-18 15:30:45"
  }
}
```

### 2. 用户注册接口测试
```bash
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "123456",
    "nickname": "测试用户"
  }'
```
**期望响应：**
```json
{
  "code": 200,
  "message": "用户注册成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "测试用户"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 3. 用户登录接口测试
```bash
curl -X POST http://localhost:8081/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "123456"
  }'
```
**期望响应：**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "测试用户"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 4. 受保护接口测试（需要JWT Token）
```bash
# 使用从登录接口获得的token
curl -X GET http://localhost:8081/api/v1/users/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 🎨 前端功能测试

### 1. 访问前端应用
打开浏览器访问：`http://localhost:3000`

### 2. 用户注册流程测试
1. 点击右上角的"注册"按钮
2. 填写注册表单：
   - 用户名：testuser2
   - 邮箱：test2@example.com
   - 密码：123456
   - 确认密码：123456
   - 昵称：测试用户2
3. 点击"注册"按钮
4. 验证注册成功后自动跳转到首页
5. 验证右上角显示用户头像和昵称

### 3. 用户登录流程测试
1. 如果已登录，先点击"退出登录"
2. 点击右上角的"登录"按钮
3. 填写登录表单：
   - 邮箱：test@example.com
   - 密码：123456
4. 勾选"记住我"选项
5. 点击"登录"按钮
6. 验证登录成功后自动跳转到首页
7. 验证右上角显示用户头像和昵称

### 4. JWT Token自动管理测试
1. 打开浏览器开发者工具（F12）
2. 切换到Network标签页
3. 在已登录状态下，刷新页面或访问其他页面
4. 观察网络请求，验证：
   - 所有API请求都自动添加了`Authorization: Bearer <token>`请求头
   - Token格式正确
   - 请求成功返回

### 5. Token过期自动刷新测试
1. 在浏览器控制台中手动修改localStorage中的token为过期token
2. 发起任何API请求
3. 验证系统自动检测到401错误
4. 验证系统自动尝试刷新token
5. 验证刷新失败后自动跳转到登录页

### 6. 用户登出功能测试
1. 在已登录状态下，点击用户头像下拉菜单
2. 点击"退出登录"
3. 验证：
   - localStorage中的token被清除
   - 页面跳转到首页
   - 右上角显示"登录"和"注册"按钮
   - 用户状态重置

## 🛡️ 权限控制测试

### 1. 受保护路由测试
1. 在未登录状态下，尝试访问需要登录的页面（如用户中心）
2. 验证自动跳转到登录页
3. 登录成功后验证自动跳转回原页面

### 2. 组件级权限测试
1. 验证未登录时某些功能按钮不显示
2. 验证登录后相应功能按钮正常显示
3. 验证不同权限用户看到的功能不同

## 🔍 错误处理测试

### 1. 网络错误测试
1. 断开网络连接
2. 尝试登录或注册
3. 验证显示友好的网络错误提示
4. 恢复网络连接后验证功能正常

### 2. 无效凭据测试
1. 使用错误的邮箱或密码尝试登录
2. 验证显示"邮箱或密码错误"提示
3. 验证不会泄露具体是哪个字段错误

### 3. 表单验证测试
1. 测试各种无效输入：
   - 空字段
   - 无效邮箱格式
   - 密码长度不足
   - 密码确认不匹配
2. 验证实时表单验证提示
3. 验证提交前的完整性检查

## 📱 多标签页同步测试

### 1. 登录状态同步
1. 打开两个标签页访问应用
2. 在一个标签页中登录
3. 验证另一个标签页的登录状态自动更新

### 2. 登出状态同步
1. 在一个标签页中登出
2. 验证另一个标签页的登录状态自动更新

## ✅ 测试检查清单

- [ ] 后端服务器成功启动在8081端口
- [ ] 前端服务器成功启动在3000端口
- [ ] 健康检查接口正常响应
- [ ] 用户注册接口正常工作
- [ ] 用户登录接口正常工作
- [ ] JWT Token正确生成和返回
- [ ] 前端注册流程完整可用
- [ ] 前端登录流程完整可用
- [ ] JWT Token自动添加到请求头
- [ ] Token过期自动处理
- [ ] 用户登出功能正常
- [ ] 受保护路由权限控制
- [ ] 组件级权限控制
- [ ] 网络错误友好提示
- [ ] 无效凭据错误提示
- [ ] 表单验证完整
- [ ] 多标签页状态同步
- [ ] 页面刷新状态持久化
- [ ] CORS配置正确
- [ ] 所有API响应格式统一

## 🐛 常见问题排查

### 1. 后端服务器启动失败
- 检查8081端口是否被占用
- 检查Go环境是否正确安装
- 检查依赖包是否完整

### 2. 前端无法连接后端
- 检查后端服务器是否正常运行
- 检查CORS配置是否正确
- 检查API baseURL配置是否正确

### 3. JWT Token问题
- 检查token格式是否正确
- 检查token是否过期
- 检查Authorization头是否正确添加

### 4. 登录状态不持久
- 检查localStorage是否正常工作
- 检查token存储和读取逻辑
- 检查页面刷新时的状态恢复逻辑
