# Go语言错误处理最佳实践

> 🎯 **学习目标**: 掌握Go语言的错误处理机制，理解与传统异常处理的差异和优势
> 
> ⏱️ **预计学习时间**: 4-5小时
> 
> 📚 **前置知识**: 已完成基础篇全部章节和进阶篇结构体接口学习

## 📋 本章内容概览

- [Go错误处理哲学](#go错误处理哲学)
- [error接口深入解析](#error接口深入解析)
- [错误创建和包装模式](#错误创建和包装模式)
- [自定义错误类型设计](#自定义错误类型设计)
- [错误处理性能优化](#错误处理性能优化)
- [panic和recover机制](#panic和recover机制)
- [第三方错误处理库](#第三方错误处理库)
- [错误日志和监控实践](#错误日志和监控实践)
- [实战案例分析](#实战案例分析)

---

## 🤔 Go错误处理哲学

### "错误是值，而不是异常" 

这是Go语言错误处理的核心哲学，与Java/Python的异常机制有根本性的差异。

#### Java异常处理 (你熟悉的方式)

```java
// Java - 基于异常的错误处理
public class UserService {
    public User getUserById(Long id) throws UserNotFoundException, DatabaseException {
        try {
            // 数据库查询可能抛出SQLException
            User user = userRepository.findById(id);
            if (user == null) {
                throw new UserNotFoundException("用户不存在: " + id);
            }
            return user;
        } catch (SQLException e) {
            // 包装底层异常
            throw new DatabaseException("数据库查询失败", e);
        }
    }
    
    // 调用方需要处理异常
    public void processUser(Long userId) {
        try {
            User user = getUserById(userId);
            // 处理用户逻辑
            processUserData(user);
        } catch (UserNotFoundException e) {
            logger.warn("用户不存在: " + e.getMessage());
        } catch (DatabaseException e) {
            logger.error("数据库错误: " + e.getMessage(), e);
            throw new ServiceException("服务暂时不可用");
        }
    }
}
```

#### Python异常处理 (你熟悉的方式)

```python
# Python - 基于异常的错误处理
class UserService:
    def get_user_by_id(self, user_id):
        try:
            # 数据库查询可能抛出异常
            user = self.user_repository.find_by_id(user_id)
            if not user:
                raise UserNotFoundException(f"用户不存在: {user_id}")
            return user
        except DatabaseError as e:
            # 包装底层异常
            raise ServiceException(f"数据库查询失败: {str(e)}") from e
    
    def process_user(self, user_id):
        try:
            user = self.get_user_by_id(user_id)
            # 处理用户逻辑
            self.process_user_data(user)
        except UserNotFoundException as e:
            logger.warning(f"用户不存在: {e}")
        except ServiceException as e:
            logger.error(f"服务错误: {e}")
            raise RuntimeError("服务暂时不可用")
```

#### Go错误处理 (现代化的方式)

```go
// Go - 基于值的错误处理
package service

import (
    "fmt"
    "errors"
    
    "github.com/yourname/mall-go/internal/model"
    "github.com/yourname/mall-go/internal/repository"
)

// 自定义错误类型
var (
    ErrUserNotFound = errors.New("用户不存在")
    ErrDatabaseError = errors.New("数据库错误")
)

type UserService struct {
    userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
    return &UserService{userRepo: userRepo}
}

// 错误作为返回值，而不是异常
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        // 包装错误，添加上下文信息
        return nil, fmt.Errorf("获取用户失败 (ID: %d): %w", id, err)
    }
    
    if user == nil {
        return nil, fmt.Errorf("用户不存在 (ID: %d): %w", id, ErrUserNotFound)
    }
    
    return user, nil
}

// 调用方显式检查错误
func (s *UserService) ProcessUser(userID uint) error {
    user, err := s.GetUserByID(userID)
    if err != nil {
        // 根据错误类型进行不同处理
        if errors.Is(err, ErrUserNotFound) {
            // 用户不存在是预期的业务错误
            return fmt.Errorf("无法处理不存在的用户: %w", err)
        }
        // 其他错误可能是系统错误
        return fmt.Errorf("获取用户信息失败: %w", err)
    }
    
    // 处理用户逻辑
    if err := s.processUserData(user); err != nil {
        return fmt.Errorf("处理用户数据失败: %w", err)
    }
    
    return nil
}

func (s *UserService) processUserData(user *model.User) error {
    // 具体的业务逻辑
    return nil
}
```

### Go错误处理的优势

#### 1. 显式性和可预测性

```go
// Go - 错误处理是显式的，不会被忽略
func ReadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }
    
    return &config, nil
}

// 调用方必须处理错误
func main() {
    config, err := ReadConfig("config.json")
    if err != nil {
        log.Fatal("配置加载失败:", err)
    }
    
    // 确保config不为nil才能使用
    fmt.Printf("配置加载成功: %+v\n", config)
}
```

```java
// Java - 异常可能被忽略或意外传播
public Config readConfig(String filename) throws IOException {
    String content = Files.readString(Paths.get(filename));
    return objectMapper.readValue(content, Config.class);
}

// 调用方可能忘记处理异常
public void main() {
    try {
        Config config = readConfig("config.json");
        System.out.println("配置: " + config);
    } catch (IOException e) {
        // 异常处理可能被遗忘
        e.printStackTrace();
    }
}
```

#### 2. 性能优势

```go
// Go错误处理的性能测试
func BenchmarkGoErrorHandling(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result, err := divide(10, 2)
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}

func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("除数不能为零")
    }
    return a / b, nil
}

// 基准测试结果：
// BenchmarkGoErrorHandling-8    100000000    10.2 ns/op    0 B/op    0 allocs/op
```

```java
// Java异常处理的性能开销
public void benchmarkJavaExceptionHandling() {
    long start = System.nanoTime();
    for (int i = 0; i < 100000000; i++) {
        try {
            int result = divide(10, 2);
        } catch (ArithmeticException e) {
            // 异常处理
        }
    }
    long end = System.nanoTime();
    // 异常处理通常比Go错误处理慢10-100倍
}

public int divide(int a, int b) throws ArithmeticException {
    if (b == 0) {
        throw new ArithmeticException("除数不能为零");
    }
    return a / b;
}
```

#### 3. 错误恢复的灵活性

```go
// Go - 灵活的错误恢复策略
func ProcessWithRetry(operation func() error, maxRetries int) error {
    var lastErr error
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := operation()
        if err == nil {
            return nil // 成功
        }
        
        lastErr = err
        
        // 根据错误类型决定是否重试
        if isRetryableError(err) {
            log.Printf("第 %d 次尝试失败，将重试: %v", attempt, err)
            time.Sleep(time.Duration(attempt) * time.Second)
            continue
        }
        
        // 不可重试的错误，直接返回
        return fmt.Errorf("操作失败，不可重试: %w", err)
    }
    
    return fmt.Errorf("重试 %d 次后仍然失败: %w", maxRetries, lastErr)
}

func isRetryableError(err error) bool {
    // 检查是否是可重试的错误类型
    var netErr *net.OpError
    if errors.As(err, &netErr) {
        return netErr.Temporary() // 网络临时错误可重试
    }
    
    // 检查特定的错误类型
    return errors.Is(err, ErrTemporaryFailure)
}

// 使用示例
func main() {
    err := ProcessWithRetry(func() error {
        return callExternalAPI()
    }, 3)
    
    if err != nil {
        log.Printf("操作最终失败: %v", err)
    }
}
```

---

## 🔍 error接口深入解析

### error接口的设计哲学

Go的error接口是语言设计的杰作，体现了"简单而强大"的设计哲学。

#### error接口定义

```go
// Go标准库中error接口的定义
type error interface {
    Error() string
}

// 这个接口极其简单，但功能强大
// 任何实现了Error() string方法的类型都是error
```

#### 与Java/Python异常体系的对比

```java
// Java - 复杂的异常继承体系
public class Exception extends Throwable {
    // 大量的方法和字段
    public String getMessage() { ... }
    public String getLocalizedMessage() { ... }
    public Throwable getCause() { ... }
    public void printStackTrace() { ... }
    public StackTraceElement[] getStackTrace() { ... }
    // ... 还有很多方法
}

// 自定义异常需要继承
public class UserNotFoundException extends Exception {
    private final Long userId;
    
    public UserNotFoundException(String message, Long userId) {
        super(message);
        this.userId = userId;
    }
    
    public Long getUserId() {
        return userId;
    }
}
```

```python
# Python - 异常类继承体系
class Exception(BaseException):
    """Common base class for all non-exit exceptions."""
    def __init__(self, *args):
        self.args = args
    
    def __str__(self):
        return str(self.args[0]) if self.args else ''
    
    def __repr__(self):
        return f"{self.__class__.__name__}{self.args!r}"

# 自定义异常需要继承
class UserNotFoundException(Exception):
    def __init__(self, message, user_id):
        super().__init__(message)
        self.user_id = user_id
```

```go
// Go - 简单而灵活的error接口
type error interface {
    Error() string
}

// 任何类型都可以实现error接口
type UserNotFoundError struct {
    UserID uint
    Message string
}

func (e UserNotFoundError) Error() string {
    return fmt.Sprintf("用户不存在 (ID: %d): %s", e.UserID, e.Message)
}

// 甚至可以用简单的字符串类型
type SimpleError string

func (e SimpleError) Error() string {
    return string(e)
}

const (
    ErrInvalidInput SimpleError = "输入参数无效"
    ErrUnauthorized SimpleError = "未授权访问"
)
```

### 标准库中的错误创建

#### 1. errors.New() - 最简单的错误创建

```go
package main

import (
    "errors"
    "fmt"
)

// 创建简单的错误
var (
    ErrUserNotFound = errors.New("用户不存在")
    ErrInvalidEmail = errors.New("邮箱格式无效")
    ErrPasswordTooShort = errors.New("密码长度不足")
)

func validateUser(email, password string) error {
    if email == "" {
        return ErrInvalidEmail
    }
    
    if len(password) < 8 {
        return ErrPasswordTooShort
    }
    
    return nil
}

func main() {
    if err := validateUser("", "123"); err != nil {
        fmt.Printf("验证失败: %v\n", err)
        
        // 错误比较
        if errors.Is(err, ErrInvalidEmail) {
            fmt.Println("这是邮箱格式错误")
        }
    }
}
```

#### 2. fmt.Errorf() - 格式化错误创建

```go
package main

import (
    "fmt"
    "time"
)

// 来自 mall-go/internal/service/user.go
func (s *UserService) CreateUser(user *model.User) error {
    // 检查邮箱是否已存在
    existingUser, err := s.userRepo.GetByEmail(user.Email)
    if err != nil {
        return fmt.Errorf("检查邮箱唯一性失败: %w", err)
    }
    
    if existingUser != nil {
        return fmt.Errorf("邮箱 %s 已被使用", user.Email)
    }
    
    // 创建用户
    user.CreatedAt = time.Now()
    if err := s.userRepo.Create(user); err != nil {
        return fmt.Errorf("创建用户失败 (邮箱: %s): %w", user.Email, err)
    }
    
    return nil
}

// 使用示例
func main() {
    userService := service.NewUserService(userRepo)
    
    user := &model.User{
        Name:  "张三",
        Email: "zhangsan@example.com",
    }
    
    if err := userService.CreateUser(user); err != nil {
        fmt.Printf("用户创建失败: %v\n", err)
        // 输出: 用户创建失败: 创建用户失败 (邮箱: zhangsan@example.com): 数据库连接超时
    }
}
```

#### 3. 错误包装和解包 (Go 1.13+)

```go
package main

import (
    "errors"
    "fmt"
    "os"
)

// 来自 mall-go/pkg/config/loader.go
func LoadConfig(filename string) (*Config, error) {
    // 第一层：文件读取
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    // 第二层：JSON解析
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败 (文件: %s): %w", filename, err)
    }
    
    // 第三层：配置验证
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }
    
    return &config, nil
}

func validateConfig(config *Config) error {
    if config.Server.Port <= 0 {
        return fmt.Errorf("服务器端口无效: %d", config.Server.Port)
    }
    return nil
}

func main() {
    config, err := LoadConfig("nonexistent.json")
    if err != nil {
        fmt.Printf("配置加载失败: %v\n", err)
        
        // 错误解包 - 检查根本原因
        if errors.Is(err, os.ErrNotExist) {
            fmt.Println("配置文件不存在，使用默认配置")
        }
        
        // 错误类型断言
        var pathErr *os.PathError
        if errors.As(err, &pathErr) {
            fmt.Printf("路径错误: %s\n", pathErr.Path)
        }
    }
}
```

---

## 🎨 自定义错误类型设计

### 结构体错误类型

Go语言允许我们创建丰富的错误类型，携带更多上下文信息。

#### 1. 基础结构体错误类型

```go
// 来自 mall-go/pkg/errors/types.go
package errors

import (
    "fmt"
    "time"
)

// ValidationError 验证错误类型
type ValidationError struct {
    Field   string    `json:"field"`
    Value   interface{} `json:"value"`
    Message string    `json:"message"`
    Code    string    `json:"code"`
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("验证失败 [%s]: %s (值: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError 创建验证错误
func NewValidationError(field string, value interface{}, message, code string) *ValidationError {
    return &ValidationError{
        Field:   field,
        Value:   value,
        Message: message,
        Code:    code,
    }
}

// BusinessError 业务错误类型
type BusinessError struct {
    Code      string    `json:"code"`
    Message   string    `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

func (e BusinessError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewBusinessError(code, message string) *BusinessError {
    return &BusinessError{
        Code:      code,
        Message:   message,
        Details:   make(map[string]interface{}),
        Timestamp: time.Now(),
    }
}

func (e *BusinessError) WithDetail(key string, value interface{}) *BusinessError {
    e.Details[key] = value
    return e
}

// 与Java异常类的对比
/*
Java中需要继承Exception类：
public class ValidationException extends Exception {
    private String field;
    private Object value;
    private String code;

    public ValidationException(String field, Object value, String message, String code) {
        super(message);
        this.field = field;
        this.value = value;
        this.code = code;
    }

    // 需要大量的getter/setter方法
    public String getField() { return field; }
    public Object getValue() { return value; }
    public String getCode() { return code; }
}
*/
```

#### 2. 错误分类和错误码设计

```go
// 来自 mall-go/pkg/errors/codes.go
package errors

// 错误分类常量
const (
    // 系统级错误 (1000-1999)
    CodeSystemError     = "SYS_001"
    CodeDatabaseError   = "SYS_002"
    CodeNetworkError    = "SYS_003"
    CodeConfigError     = "SYS_004"

    // 认证授权错误 (2000-2999)
    CodeUnauthorized    = "AUTH_001"
    CodeForbidden       = "AUTH_002"
    CodeTokenExpired    = "AUTH_003"
    CodeInvalidToken    = "AUTH_004"

    // 业务逻辑错误 (3000-3999)
    CodeUserNotFound    = "BIZ_001"
    CodeEmailExists     = "BIZ_002"
    CodeInvalidPassword = "BIZ_003"
    CodeOrderNotFound   = "BIZ_004"

    // 验证错误 (4000-4999)
    CodeValidationFailed = "VAL_001"
    CodeInvalidEmail     = "VAL_002"
    CodeInvalidPhone     = "VAL_003"
    CodeRequiredField    = "VAL_004"
)

// 错误消息映射
var errorMessages = map[string]string{
    CodeSystemError:      "系统内部错误",
    CodeDatabaseError:    "数据库操作失败",
    CodeNetworkError:     "网络连接错误",
    CodeUnauthorized:     "未授权访问",
    CodeUserNotFound:     "用户不存在",
    CodeEmailExists:      "邮箱已存在",
    CodeValidationFailed: "数据验证失败",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code string) string {
    if msg, exists := errorMessages[code]; exists {
        return msg
    }
    return "未知错误"
}

// 分类错误构造函数
func NewSystemError(code, message string) *BusinessError {
    return &BusinessError{
        Code:      code,
        Message:   message,
        Timestamp: time.Now(),
    }
}

func NewAuthError(code string) *BusinessError {
    return &BusinessError{
        Code:      code,
        Message:   GetErrorMessage(code),
        Timestamp: time.Now(),
    }
}

func NewBusinessLogicError(code string, details map[string]interface{}) *BusinessError {
    return &BusinessError{
        Code:      code,
        Message:   GetErrorMessage(code),
        Details:   details,
        Timestamp: time.Now(),
    }
}
```

#### 3. 错误链和错误包装

```go
// 来自 mall-go/internal/service/user.go
package service

import (
    "fmt"
    "errors"

    "github.com/yourname/mall-go/pkg/errors"
    "github.com/yourname/mall-go/internal/model"
)

type UserService struct {
    userRepo repository.UserRepository
    logger   *zap.Logger
}

// 复杂的错误处理链
func (s *UserService) RegisterUser(req *RegisterRequest) (*model.User, error) {
    // 第一层：输入验证
    if err := s.validateRegisterRequest(req); err != nil {
        return nil, fmt.Errorf("用户注册验证失败: %w", err)
    }

    // 第二层：业务逻辑检查
    if err := s.checkEmailUniqueness(req.Email); err != nil {
        return nil, fmt.Errorf("邮箱唯一性检查失败: %w", err)
    }

    // 第三层：密码加密
    hashedPassword, err := s.hashPassword(req.Password)
    if err != nil {
        return nil, fmt.Errorf("密码加密失败: %w", err)
    }

    // 第四层：数据库操作
    user := &model.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashedPassword,
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, fmt.Errorf("创建用户记录失败: %w", err)
    }

    return user, nil
}

func (s *UserService) validateRegisterRequest(req *RegisterRequest) error {
    var validationErrors []error

    if req.Name == "" {
        validationErrors = append(validationErrors,
            errors.NewValidationError("name", req.Name, "用户名不能为空", errors.CodeRequiredField))
    }

    if !isValidEmail(req.Email) {
        validationErrors = append(validationErrors,
            errors.NewValidationError("email", req.Email, "邮箱格式无效", errors.CodeInvalidEmail))
    }

    if len(req.Password) < 8 {
        validationErrors = append(validationErrors,
            errors.NewValidationError("password", "***", "密码长度至少8位", errors.CodeValidationFailed))
    }

    if len(validationErrors) > 0 {
        return &MultiValidationError{Errors: validationErrors}
    }

    return nil
}

// 多重验证错误类型
type MultiValidationError struct {
    Errors []error
}

func (e MultiValidationError) Error() string {
    if len(e.Errors) == 1 {
        return e.Errors[0].Error()
    }

    return fmt.Sprintf("发现 %d 个验证错误: %v", len(e.Errors), e.Errors[0])
}

func (e MultiValidationError) Unwrap() []error {
    return e.Errors
}
```

#### 4. 错误类型判断和处理

```go
// 来自 mall-go/internal/handler/user.go
package handler

import (
    "net/http"
    "errors"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/yourname/mall-go/pkg/errors"
)

type UserHandler struct {
    userService *service.UserService
}

func (h *UserHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "请求参数格式错误",
            "details": err.Error(),
        })
        return
    }

    user, err := h.userService.RegisterUser(&req)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "用户注册成功",
        "user": user,
    })
}

// 统一错误处理
func (h *UserHandler) handleError(c *gin.Context, err error) {
    // 1. 检查验证错误
    var validationErr *errors.ValidationError
    if errors.As(err, &validationErr) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "数据验证失败",
            "field": validationErr.Field,
            "message": validationErr.Message,
            "code": validationErr.Code,
        })
        return
    }

    // 2. 检查多重验证错误
    var multiValidationErr *MultiValidationError
    if errors.As(err, &multiValidationErr) {
        var details []map[string]interface{}
        for _, e := range multiValidationErr.Errors {
            if ve, ok := e.(*errors.ValidationError); ok {
                details = append(details, map[string]interface{}{
                    "field": ve.Field,
                    "message": ve.Message,
                    "code": ve.Code,
                })
            }
        }

        c.JSON(http.StatusBadRequest, gin.H{
            "error": "数据验证失败",
            "details": details,
        })
        return
    }

    // 3. 检查业务错误
    var businessErr *errors.BusinessError
    if errors.As(err, &businessErr) {
        statusCode := getHTTPStatusFromErrorCode(businessErr.Code)
        c.JSON(statusCode, gin.H{
            "error": businessErr.Message,
            "code": businessErr.Code,
            "details": businessErr.Details,
            "timestamp": businessErr.Timestamp,
        })
        return
    }

    // 4. 默认系统错误
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "系统内部错误",
        "message": err.Error(),
    })
}

func getHTTPStatusFromErrorCode(code string) int {
    switch {
    case strings.HasPrefix(code, "AUTH_"):
        return http.StatusUnauthorized
    case strings.HasPrefix(code, "BIZ_"):
        return http.StatusBadRequest
    case strings.HasPrefix(code, "VAL_"):
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}
```

---

## ⚡ 错误处理性能优化

### 性能对比分析

Go的错误处理相比异常处理有显著的性能优势。

#### 1. 基准测试对比

```go
// 来自 mall-go/benchmark/error_test.go
package benchmark

import (
    "errors"
    "testing"
)

var (
    testError = errors.New("测试错误")
    result    int
    err       error
)

// Go错误处理基准测试
func BenchmarkGoErrorHandling(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result, err = divideWithError(100, 2)
        if err != nil {
            // 错误处理
            continue
        }
    }
}

// Go错误处理 - 有错误情况
func BenchmarkGoErrorHandlingWithError(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        result, err = divideWithError(100, 0)
        if err != nil {
            // 错误处理
            continue
        }
    }
}

func divideWithError(a, b int) (int, error) {
    if b == 0 {
        return 0, testError
    }
    return a / b, nil
}

// 模拟异常处理的开销
func BenchmarkSimulatedExceptionHandling(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    // 异常处理
                }
            }()

            if i%2 == 0 {
                panic("模拟异常")
            }
            result = 100 / 2
        }()
    }
}

/*
基准测试结果对比：
BenchmarkGoErrorHandling-8                100000000    10.2 ns/op    0 B/op    0 allocs/op
BenchmarkGoErrorHandlingWithError-8       100000000    11.5 ns/op    0 B/op    0 allocs/op
BenchmarkSimulatedExceptionHandling-8      10000000   150.3 ns/op   32 B/op    2 allocs/op

结论：Go错误处理比异常处理快约13倍，且无内存分配
*/
```

#### 2. 错误处理优化策略

```go
// 来自 mall-go/pkg/errors/pool.go
package errors

import (
    "sync"
)

// 错误对象池，减少内存分配
var validationErrorPool = sync.Pool{
    New: func() interface{} {
        return &ValidationError{}
    },
}

// 优化的验证错误创建
func NewValidationErrorFromPool(field string, value interface{}, message, code string) *ValidationError {
    err := validationErrorPool.Get().(*ValidationError)
    err.Field = field
    err.Value = value
    err.Message = message
    err.Code = code
    return err
}

// 释放错误对象回池
func ReleaseValidationError(err *ValidationError) {
    err.Field = ""
    err.Value = nil
    err.Message = ""
    err.Code = ""
    validationErrorPool.Put(err)
}

// 预定义错误，避免重复创建
var (
    ErrUserNotFound     = errors.New("用户不存在")
    ErrEmailExists      = errors.New("邮箱已存在")
    ErrInvalidPassword  = errors.New("密码无效")
    ErrUnauthorized     = errors.New("未授权访问")
)

// 错误包装优化
func WrapError(err error, message string) error {
    if err == nil {
        return nil
    }

    // 避免重复包装相同的错误
    if wrapped, ok := err.(*WrappedError); ok {
        if wrapped.Message == message {
            return err
        }
    }

    return &WrappedError{
        Message: message,
        Cause:   err,
    }
}

type WrappedError struct {
    Message string
    Cause   error
}

func (e *WrappedError) Error() string {
    return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *WrappedError) Unwrap() error {
    return e.Cause
}
```

#### 3. 高性能错误检查模式

```go
// 来自 mall-go/internal/service/optimized.go
package service

import (
    "errors"
    "fmt"
)

// 快速错误检查模式
type ErrorChecker struct {
    err error
}

func NewErrorChecker() *ErrorChecker {
    return &ErrorChecker{}
}

func (ec *ErrorChecker) Check(err error) *ErrorChecker {
    if ec.err == nil && err != nil {
        ec.err = err
    }
    return ec
}

func (ec *ErrorChecker) CheckWithContext(err error, context string) *ErrorChecker {
    if ec.err == nil && err != nil {
        ec.err = fmt.Errorf("%s: %w", context, err)
    }
    return ec
}

func (ec *ErrorChecker) Error() error {
    return ec.err
}

// 使用示例：链式错误检查
func (s *UserService) ProcessUserOptimized(userID uint) error {
    checker := NewErrorChecker()

    // 链式错误检查，只有第一个错误会被记录
    user, err := s.userRepo.GetByID(userID)
    checker.Check(err)

    err = s.validateUser(user)
    checker.CheckWithContext(err, "用户验证")

    err = s.updateUserStatus(user)
    checker.CheckWithContext(err, "状态更新")

    err = s.sendNotification(user)
    checker.CheckWithContext(err, "发送通知")

    return checker.Error()
}

// 批量操作的错误处理优化
func (s *UserService) BatchProcessUsers(userIDs []uint) ([]error, error) {
    if len(userIDs) == 0 {
        return nil, errors.New("用户ID列表不能为空")
    }

    // 预分配错误切片，避免动态扩容
    errs := make([]error, 0, len(userIDs))

    for _, userID := range userIDs {
        if err := s.ProcessUserOptimized(userID); err != nil {
            errs = append(errs, fmt.Errorf("处理用户 %d 失败: %w", userID, err))
        }
    }

    if len(errs) > 0 {
        return errs, fmt.Errorf("批量处理失败，%d/%d 个用户处理失败", len(errs), len(userIDs))
    }

    return nil, nil
}
```

---

## 😱 panic和recover机制

### panic vs 异常处理

Go的panic/recover机制类似于其他语言的异常处理，但使用场景更加严格。

#### 1. panic的使用场景

```go
// 来自 mall-go/pkg/database/connection.go
package database

import (
    "database/sql"
    "fmt"
    "log"
)

// ❌ 错误的panic使用 - 不要在业务逻辑中使用panic
func BadExample_GetUser(id uint) *User {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        panic(fmt.Sprintf("数据库查询失败: %v", err)) // 不好的做法
    }
    return user
}

// ✅ 正确的panic使用 - 程序初始化失败
func InitDatabase(dsn string) *sql.DB {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        panic(fmt.Sprintf("数据库连接失败: %v", err)) // 合理的panic使用
    }

    if err := db.Ping(); err != nil {
        panic(fmt.Sprintf("数据库连接测试失败: %v", err))
    }

    log.Println("数据库连接成功")
    return db
}

// ✅ 正确的panic使用 - 检测编程错误
func ProcessSlice(data []int, index int) int {
    if index < 0 || index >= len(data) {
        panic(fmt.Sprintf("索引越界: index=%d, len=%d", index, len(data)))
    }
    return data[index]
}

// ✅ 正确的panic使用 - 不可恢复的系统错误
func LoadCriticalConfig() *Config {
    config, err := loadConfigFromFile("critical.conf")
    if err != nil {
        panic(fmt.Sprintf("加载关键配置失败，程序无法继续运行: %v", err))
    }
    return config
}
```

#### 2. recover的使用模式

```go
// 来自 mall-go/pkg/middleware/recovery.go
package middleware

import (
    "fmt"
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// HTTP服务器的panic恢复中间件
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // 记录panic信息和堆栈
                stack := debug.Stack()
                logger.Error("HTTP请求panic",
                    zap.Any("panic", r),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("method", c.Request.Method),
                    zap.String("stack", string(stack)),
                )

                // 返回500错误，而不是让程序崩溃
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "服务器内部错误",
                    "code":  "INTERNAL_ERROR",
                })

                c.Abort()
            }
        }()

        c.Next()
    }
}

// 工作池中的panic恢复
func SafeWorker(taskChan <-chan Task, logger *zap.Logger) {
    for task := range taskChan {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    logger.Error("工作任务panic",
                        zap.Any("panic", r),
                        zap.String("task_id", task.ID),
                        zap.String("stack", string(debug.Stack())),
                    )
                }
            }()

            // 执行任务
            task.Execute()
        }()
    }
}

// 数据库事务中的panic处理
func (s *UserService) CreateUserWithTransaction(user *User) (err error) {
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("开始事务失败: %w", err)
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            err = fmt.Errorf("创建用户时发生panic: %v", r)
        } else if err != nil {
            tx.Rollback()
        } else {
            err = tx.Commit()
        }
    }()

    // 可能panic的操作
    if err = s.validateUser(user); err != nil {
        return err
    }

    if err = s.insertUser(tx, user); err != nil {
        return err
    }

    return nil
}
```

#### 3. 与Java/Python异常处理的对比

```java
// Java - 异常处理是常规控制流
public class UserService {
    public User createUser(User user) throws ValidationException, DatabaseException {
        try {
            validateUser(user);
            return userRepository.save(user);
        } catch (ValidationException e) {
            // 业务异常，正常的错误处理
            logger.warn("用户验证失败: " + e.getMessage());
            throw e;
        } catch (SQLException e) {
            // 系统异常，包装后抛出
            throw new DatabaseException("数据库操作失败", e);
        }
    }

    // 异常处理是预期的控制流
    public User findUserByEmail(String email) throws UserNotFoundException {
        User user = userRepository.findByEmail(email);
        if (user == null) {
            throw new UserNotFoundException("用户不存在: " + email);
        }
        return user;
    }
}
```

```python
# Python - 异常处理也是常规控制流
class UserService:
    def create_user(self, user):
        try:
            self.validate_user(user)
            return self.user_repository.save(user)
        except ValidationError as e:
            # 业务异常，正常处理
            logger.warning(f"用户验证失败: {e}")
            raise
        except DatabaseError as e:
            # 系统异常，包装后抛出
            raise ServiceError(f"数据库操作失败: {e}") from e

    def find_user_by_email(self, email):
        user = self.user_repository.find_by_email(email)
        if not user:
            raise UserNotFoundError(f"用户不存在: {email}")
        return user
```

```go
// Go - panic/recover只用于真正的异常情况
type UserService struct {
    userRepo repository.UserRepository
    logger   *zap.Logger
}

// 正常的错误处理，不使用panic
func (s *UserService) CreateUser(user *User) error {
    if err := s.validateUser(user); err != nil {
        return fmt.Errorf("用户验证失败: %w", err)
    }

    if err := s.userRepo.Save(user); err != nil {
        return fmt.Errorf("保存用户失败: %w", err)
    }

    return nil
}

// 正常的业务逻辑，返回错误而不是panic
func (s *UserService) FindUserByEmail(email string) (*User, error) {
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }

    if user == nil {
        return nil, fmt.Errorf("用户不存在: %s", email)
    }

    return user, nil
}

// panic只用于程序无法继续运行的情况
func (s *UserService) MustInitialize() {
    if s.userRepo == nil {
        panic("UserRepository未初始化，程序无法继续运行")
    }

    if s.logger == nil {
        panic("Logger未初始化，程序无法继续运行")
    }
}
```

---

## 📚 第三方错误处理库

### 流行的错误处理库

Go生态中有一些优秀的第三方错误处理库，可以增强标准库的功能。

#### 1. pkg/errors - 错误堆栈跟踪

```go
// go get github.com/pkg/errors

package main

import (
    "database/sql"
    "fmt"

    "github.com/pkg/errors"
)

// 使用pkg/errors增强错误信息
func connectDatabase(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        // 添加堆栈信息
        return nil, errors.Wrap(err, "打开数据库连接失败")
    }

    if err := db.Ping(); err != nil {
        // 添加更多上下文
        return nil, errors.Wrapf(err, "数据库连接测试失败 (DSN: %s)", dsn)
    }

    return db, nil
}

func getUserFromDB(db *sql.DB, userID int) (*User, error) {
    var user User
    err := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", userID).
        Scan(&user.ID, &user.Name, &user.Email)

    if err != nil {
        if err == sql.ErrNoRows {
            // 创建新的错误，带堆栈信息
            return nil, errors.Errorf("用户不存在 (ID: %d)", userID)
        }
        // 包装现有错误
        return nil, errors.Wrapf(err, "查询用户失败 (ID: %d)", userID)
    }

    return &user, nil
}

func main() {
    db, err := connectDatabase("invalid-dsn")
    if err != nil {
        // 打印详细的错误信息和堆栈
        fmt.Printf("错误: %+v\n", err)
        /*
        输出包含完整的调用堆栈：
        打开数据库连接失败: sql: unknown driver "mysql" (forgotten import?)
        main.connectDatabase
            /path/to/main.go:15
        main.main
            /path/to/main.go:45
        runtime.main
            /usr/local/go/src/runtime/proc.go:250
        */
        return
    }

    user, err := getUserFromDB(db, 123)
    if err != nil {
        // 检查错误原因
        cause := errors.Cause(err)
        if cause == sql.ErrNoRows {
            fmt.Println("用户不存在")
        } else {
            fmt.Printf("数据库错误: %+v\n", err)
        }
    }
}
```

#### 2. go-multierror - 多重错误处理

```go
// go get github.com/hashicorp/go-multierror

package main

import (
    "fmt"
    "time"

    "github.com/hashicorp/go-multierror"
)

// 来自 mall-go/internal/service/batch.go
func (s *UserService) BatchValidateUsers(users []*User) error {
    var result *multierror.Error

    for i, user := range users {
        if err := s.validateUser(user); err != nil {
            result = multierror.Append(result,
                fmt.Errorf("用户 %d 验证失败: %w", i, err))
        }
    }

    return result.ErrorOrNil()
}

// 并发处理多个任务
func (s *UserService) ProcessUsersParallel(userIDs []uint) error {
    var result *multierror.Error
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, userID := range userIDs {
        wg.Add(1)
        go func(id uint) {
            defer wg.Done()

            if err := s.ProcessUser(id); err != nil {
                mu.Lock()
                result = multierror.Append(result,
                    fmt.Errorf("处理用户 %d 失败: %w", id, err))
                mu.Unlock()
            }
        }(userID)
    }

    wg.Wait()
    return result.ErrorOrNil()
}

// 系统健康检查
func (s *SystemService) HealthCheck() error {
    var result *multierror.Error

    // 检查数据库连接
    if err := s.checkDatabase(); err != nil {
        result = multierror.Append(result,
            fmt.Errorf("数据库健康检查失败: %w", err))
    }

    // 检查Redis连接
    if err := s.checkRedis(); err != nil {
        result = multierror.Append(result,
            fmt.Errorf("Redis健康检查失败: %w", err))
    }

    // 检查外部API
    if err := s.checkExternalAPI(); err != nil {
        result = multierror.Append(result,
            fmt.Errorf("外部API健康检查失败: %w", err))
    }

    if result.ErrorOrNil() != nil {
        return fmt.Errorf("系统健康检查失败: %w", result)
    }

    return nil
}

func main() {
    service := NewUserService()

    users := []*User{
        {Name: "", Email: "invalid"},
        {Name: "张三", Email: "zhangsan@example.com"},
        {Name: "李四", Email: "invalid-email"},
    }

    if err := service.BatchValidateUsers(users); err != nil {
        fmt.Printf("批量验证失败:\n%v\n", err)
        /*
        输出：
        3 errors occurred:
            * 用户 0 验证失败: 用户名不能为空
            * 用户 0 验证失败: 邮箱格式无效
            * 用户 2 验证失败: 邮箱格式无效
        */
    }
}
```

#### 3. emperror - 企业级错误处理

```go
// go get emperror.dev/errors

package main

import (
    "context"
    "fmt"

    "emperror.dev/errors"
)

// 来自 mall-go/pkg/errors/enhanced.go
type ErrorHandler struct {
    logger Logger
}

func NewErrorHandler(logger Logger) *ErrorHandler {
    return &ErrorHandler{logger: logger}
}

// 增强的错误处理
func (h *ErrorHandler) Handle(ctx context.Context, err error) {
    if err == nil {
        return
    }

    // 提取错误详情
    details := errors.GetDetails(err)

    // 记录结构化日志
    h.logger.Error("处理错误",
        "error", err.Error(),
        "details", details,
        "stack", fmt.Sprintf("%+v", err),
    )

    // 根据错误类型进行不同处理
    switch {
    case errors.Is(err, context.Canceled):
        h.logger.Info("操作被取消")
    case errors.Is(err, context.DeadlineExceeded):
        h.logger.Warn("操作超时")
    default:
        // 发送告警
        h.sendAlert(ctx, err)
    }
}

func (h *ErrorHandler) sendAlert(ctx context.Context, err error) {
    // 发送错误告警到监控系统
    // 实现略...
}

// 使用示例
func (s *UserService) CreateUserWithEnhancedError(ctx context.Context, user *User) error {
    // 添加上下文信息
    err := s.userRepo.Create(user)
    if err != nil {
        return errors.WithDetails(err,
            "user_id", user.ID,
            "user_email", user.Email,
            "operation", "create_user",
        )
    }

    return nil
}
```

---

## 📊 错误日志和监控实践

### 结构化错误日志

良好的错误日志是问题诊断和系统监控的基础。

#### 1. 结构化日志记录

```go
// 来自 mall-go/pkg/logger/error.go
package logger

import (
    "context"
    "runtime"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type ErrorLogger struct {
    logger *zap.Logger
}

func NewErrorLogger() *ErrorLogger {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    logger, _ := config.Build()
    return &ErrorLogger{logger: logger}
}

// 记录错误的详细信息
func (el *ErrorLogger) LogError(ctx context.Context, err error, operation string, details map[string]interface{}) {
    if err == nil {
        return
    }

    // 获取调用者信息
    pc, file, line, _ := runtime.Caller(1)
    funcName := runtime.FuncForPC(pc).Name()

    // 构建日志字段
    fields := []zap.Field{
        zap.Error(err),
        zap.String("operation", operation),
        zap.String("caller", fmt.Sprintf("%s:%d", file, line)),
        zap.String("function", funcName),
        zap.Time("timestamp", time.Now()),
    }

    // 添加上下文信息
    if requestID := ctx.Value("request_id"); requestID != nil {
        fields = append(fields, zap.String("request_id", requestID.(string)))
    }

    if userID := ctx.Value("user_id"); userID != nil {
        fields = append(fields, zap.Uint("user_id", userID.(uint)))
    }

    // 添加自定义详情
    for key, value := range details {
        fields = append(fields, zap.Any(key, value))
    }

    // 根据错误类型选择日志级别
    switch {
    case isBusinessError(err):
        el.logger.Warn("业务错误", fields...)
    case isSystemError(err):
        el.logger.Error("系统错误", fields...)
    default:
        el.logger.Error("未知错误", fields...)
    }
}

func isBusinessError(err error) bool {
    var businessErr *BusinessError
    return errors.As(err, &businessErr)
}

func isSystemError(err error) bool {
    // 检查是否是系统级错误
    return strings.Contains(err.Error(), "database") ||
           strings.Contains(err.Error(), "network") ||
           strings.Contains(err.Error(), "timeout")
}
```

#### 2. 错误追踪和链路跟踪

```go
// 来自 mall-go/pkg/tracing/error.go
package tracing

import (
    "context"
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

// 在链路跟踪中记录错误
func RecordError(ctx context.Context, err error, operation string) {
    if err == nil {
        return
    }

    span := trace.SpanFromContext(ctx)
    if !span.IsRecording() {
        return
    }

    // 设置错误状态
    span.SetStatus(codes.Error, err.Error())

    // 记录错误事件
    span.AddEvent("error", trace.WithAttributes(
        attribute.String("error.type", fmt.Sprintf("%T", err)),
        attribute.String("error.message", err.Error()),
        attribute.String("operation", operation),
    ))

    // 添加错误标签
    span.SetAttributes(
        attribute.Bool("error", true),
        attribute.String("error.kind", getErrorKind(err)),
    )
}

func getErrorKind(err error) string {
    switch {
    case isValidationError(err):
        return "validation"
    case isBusinessError(err):
        return "business"
    case isSystemError(err):
        return "system"
    default:
        return "unknown"
    }
}

// 使用示例
func (s *UserService) CreateUserWithTracing(ctx context.Context, user *User) error {
    tracer := otel.Tracer("user-service")
    ctx, span := tracer.Start(ctx, "create_user")
    defer span.End()

    // 添加用户信息到span
    span.SetAttributes(
        attribute.String("user.email", user.Email),
        attribute.String("user.name", user.Name),
    )

    if err := s.validateUser(user); err != nil {
        RecordError(ctx, err, "validate_user")
        return fmt.Errorf("用户验证失败: %w", err)
    }

    if err := s.userRepo.Create(user); err != nil {
        RecordError(ctx, err, "create_user_record")
        return fmt.Errorf("创建用户记录失败: %w", err)
    }

    span.SetAttributes(attribute.Uint("user.id", user.ID))
    return nil
}
```

#### 3. 错误监控和告警

```go
// 来自 mall-go/pkg/monitoring/error.go
package monitoring

import (
    "context"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus指标
var (
    errorCounter = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "application_errors_total",
            Help: "应用程序错误总数",
        },
        []string{"service", "operation", "error_type"},
    )

    errorDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "error_handling_duration_seconds",
            Help: "错误处理耗时",
            Buckets: prometheus.DefBuckets,
        },
        []string{"service", "operation"},
    )
)

type ErrorMonitor struct {
    serviceName string
    alerter     Alerter
}

func NewErrorMonitor(serviceName string, alerter Alerter) *ErrorMonitor {
    return &ErrorMonitor{
        serviceName: serviceName,
        alerter:     alerter,
    }
}

// 监控错误
func (em *ErrorMonitor) RecordError(ctx context.Context, err error, operation string) {
    if err == nil {
        return
    }

    start := time.Now()
    defer func() {
        errorDuration.WithLabelValues(em.serviceName, operation).
            Observe(time.Since(start).Seconds())
    }()

    // 增加错误计数
    errorType := getErrorType(err)
    errorCounter.WithLabelValues(em.serviceName, operation, errorType).Inc()

    // 检查是否需要告警
    if shouldAlert(err) {
        em.sendAlert(ctx, err, operation)
    }
}

func getErrorType(err error) string {
    switch {
    case isValidationError(err):
        return "validation"
    case isBusinessError(err):
        return "business"
    case isSystemError(err):
        return "system"
    case isNetworkError(err):
        return "network"
    default:
        return "unknown"
    }
}

func shouldAlert(err error) bool {
    // 只对系统错误和网络错误发送告警
    return isSystemError(err) || isNetworkError(err)
}

func (em *ErrorMonitor) sendAlert(ctx context.Context, err error, operation string) {
    alert := Alert{
        Service:   em.serviceName,
        Operation: operation,
        Error:     err.Error(),
        Timestamp: time.Now(),
        Severity:  getSeverity(err),
    }

    em.alerter.Send(ctx, alert)
}

func getSeverity(err error) string {
    switch {
    case isSystemError(err):
        return "critical"
    case isNetworkError(err):
        return "warning"
    default:
        return "info"
    }
}

// 告警接口
type Alerter interface {
    Send(ctx context.Context, alert Alert) error
}

type Alert struct {
    Service   string    `json:"service"`
    Operation string    `json:"operation"`
    Error     string    `json:"error"`
    Timestamp time.Time `json:"timestamp"`
    Severity  string    `json:"severity"`
}

// 钉钉告警实现
type DingTalkAlerter struct {
    webhook string
}

func (d *DingTalkAlerter) Send(ctx context.Context, alert Alert) error {
    message := fmt.Sprintf("🚨 系统告警\n服务: %s\n操作: %s\n错误: %s\n时间: %s\n严重程度: %s",
        alert.Service, alert.Operation, alert.Error,
        alert.Timestamp.Format("2006-01-02 15:04:05"), alert.Severity)

    // 发送到钉钉群
    return sendToDingTalk(d.webhook, message)
}
```

---

## 🎯 面试常考点

### 1. Go错误处理vs异常处理的优缺点

**面试题**: "Go的错误处理相比Java/Python的异常处理有什么优缺点？"

**标准答案**:
```go
// Go错误处理的优点：
// 1. 显式性 - 错误必须被显式处理，不会被忽略
func ReadFile(filename string) ([]byte, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取文件失败: %w", err)
    }
    return data, nil
}

// 调用方必须处理错误
data, err := ReadFile("config.json")
if err != nil {
    log.Fatal(err) // 必须处理
}

// 2. 性能优势 - 无异常栈展开开销
// 3. 代码可读性 - 错误处理路径清晰可见
// 4. 类型安全 - 错误类型在编译时确定

// Go错误处理的缺点：
// 1. 代码冗长 - 需要大量的if err != nil检查
// 2. 错误传播繁琐 - 需要手动传播错误
// 3. 缺少堆栈信息 - 标准error不包含调用栈

// Java异常处理的优点：
// 1. 自动传播 - 异常会自动向上传播
// 2. 丰富的信息 - 包含完整的调用栈
// 3. 代码简洁 - 正常流程和异常处理分离

// Java异常处理的缺点：
// 1. 性能开销 - 异常创建和栈展开有开销
// 2. 隐式性 - 异常可能被忽略或意外传播
// 3. 控制流复杂 - 异常改变了正常的控制流
```

### 2. error接口的设计原理

**面试题**: "为什么Go的error接口只有一个Error()方法？"

**标准答案**:
```go
// error接口的简洁设计体现了Go的设计哲学：
type error interface {
    Error() string
}

// 1. 简单性 - 接口越小越容易实现和组合
// 2. 灵活性 - 任何类型都可以实现error接口
// 3. 组合性 - 可以通过组合实现复杂的错误类型

// 自定义错误类型示例
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("验证失败 [%s]: %s", e.Field, e.Message)
}

// 可以添加额外的方法
func (e ValidationError) GetField() string {
    return e.Field
}

// 错误类型断言
var validationErr ValidationError
if errors.As(err, &validationErr) {
    field := validationErr.GetField() // 使用额外的方法
}
```

### 3. 错误包装和解包的实现原理

**面试题**: "Go 1.13的错误包装是如何实现的？"

**标准答案**:
```go
// 错误包装通过Unwrap()方法实现
type wrapError struct {
    msg string
    err error
}

func (e *wrapError) Error() string {
    return e.msg + ": " + e.err.Error()
}

func (e *wrapError) Unwrap() error {
    return e.err
}

// fmt.Errorf使用%w动词创建包装错误
err := fmt.Errorf("操作失败: %w", originalErr)

// errors.Is递归检查错误链
func Is(err, target error) bool {
    if target == nil {
        return err == target
    }

    for {
        if err == target {
            return true
        }
        if x, ok := err.(interface{ Unwrap() error }); ok {
            err = x.Unwrap()
            if err == nil {
                return false
            }
        } else {
            return false
        }
    }
}

// errors.As递归查找特定类型的错误
func As(err error, target interface{}) bool {
    // 实现略...
    // 递归调用Unwrap()查找匹配的错误类型
}
```

### 4. panic和recover的使用场景

**面试题**: "什么时候应该使用panic，什么时候使用recover？"

**标准答案**:
```go
// panic的使用场景：
// 1. 程序初始化失败
func init() {
    if critical_resource == nil {
        panic("关键资源初始化失败")
    }
}

// 2. 编程错误（数组越界等）
func ProcessArray(arr []int, index int) {
    if index < 0 || index >= len(arr) {
        panic("索引越界")
    }
    // 处理逻辑
}

// 3. 不可恢复的错误
func LoadCriticalConfig() {
    config, err := loadConfig()
    if err != nil {
        panic("无法加载关键配置，程序无法继续")
    }
}

// recover的使用场景：
// 1. HTTP服务器中防止单个请求崩溃整个服务
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Panic: %v", r)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// 2. 工作池中防止单个任务崩溃
func worker(tasks <-chan Task) {
    for task := range tasks {
        func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Printf("Task panic: %v", r)
                }
            }()
            task.Execute()
        }()
    }
}

// ❌ 不要用panic/recover替代正常的错误处理
func BadExample() {
    defer func() {
        if r := recover(); r != nil {
            // 这是错误的用法
        }
    }()

    if someCondition {
        panic("业务错误") // 应该返回error
    }
}
```

### 5. 错误处理的性能考虑

**面试题**: "Go错误处理有哪些性能优化技巧？"

**标准答案**:
```go
// 1. 预定义错误，避免重复创建
var (
    ErrUserNotFound = errors.New("用户不存在")
    ErrInvalidInput = errors.New("输入无效")
)

// ❌ 每次都创建新错误
func BadExample(id int) error {
    return fmt.Errorf("用户不存在: %d", id) // 每次都分配内存
}

// ✅ 使用预定义错误
func GoodExample(id int) error {
    return ErrUserNotFound // 无内存分配
}

// 2. 错误对象池
var errorPool = sync.Pool{
    New: func() interface{} {
        return &CustomError{}
    },
}

func NewCustomErrorFromPool(msg string) *CustomError {
    err := errorPool.Get().(*CustomError)
    err.Message = msg
    return err
}

func ReleaseCustomError(err *CustomError) {
    err.Message = ""
    errorPool.Put(err)
}

// 3. 避免不必要的错误包装
func OptimizedFunction() error {
    err := someOperation()
    if err != nil {
        // 如果不需要添加上下文，直接返回
        return err
    }
    return nil
}

// 4. 批量错误处理
func BatchProcess(items []Item) []error {
    errors := make([]error, 0, len(items)) // 预分配容量
    for _, item := range items {
        if err := processItem(item); err != nil {
            errors = append(errors, err)
        }
    }
    return errors
}
```

---

## 💡 踩坑提醒

### 1. 错误检查的遗漏

```go
// ❌ 错误：忽略错误检查
func BadExample() {
    data, _ := os.ReadFile("config.json") // 忽略错误
    var config Config
    json.Unmarshal(data, &config) // 如果data为nil会panic

    // 使用config...
}

// ✅ 正确：始终检查错误
func GoodExample() error {
    data, err := os.ReadFile("config.json")
    if err != nil {
        return fmt.Errorf("读取配置文件失败: %w", err)
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return fmt.Errorf("解析配置文件失败: %w", err)
    }

    // 使用config...
    return nil
}

// ❌ 错误：部分错误检查
func PartialErrorCheck() {
    file, err := os.Open("data.txt")
    if err != nil {
        log.Println("打开文件失败")
        return
    }

    data := make([]byte, 1024)
    file.Read(data) // 忽略了Read的错误返回值
    file.Close()    // 忽略了Close的错误返回值
}

// ✅ 正确：完整的错误检查
func CompleteErrorCheck() error {
    file, err := os.Open("data.txt")
    if err != nil {
        return fmt.Errorf("打开文件失败: %w", err)
    }
    defer func() {
        if closeErr := file.Close(); closeErr != nil {
            log.Printf("关闭文件失败: %v", closeErr)
        }
    }()

    data := make([]byte, 1024)
    n, err := file.Read(data)
    if err != nil && err != io.EOF {
        return fmt.Errorf("读取文件失败: %w", err)
    }

    // 使用data[:n]...
    return nil
}
```

### 2. 错误信息的丢失

```go
// ❌ 错误：丢失原始错误信息
func BadErrorWrapping(userID uint) error {
    user, err := getUserFromDB(userID)
    if err != nil {
        return errors.New("获取用户失败") // 丢失了原始错误信息
    }
    return nil
}

// ✅ 正确：保留原始错误信息
func GoodErrorWrapping(userID uint) error {
    user, err := getUserFromDB(userID)
    if err != nil {
        return fmt.Errorf("获取用户失败 (ID: %d): %w", userID, err)
    }
    return nil
}

// ❌ 错误：错误信息不够详细
func VagueError() error {
    if someCondition {
        return errors.New("操作失败") // 信息太模糊
    }
    return nil
}

// ✅ 正确：提供详细的错误信息
func DetailedError(operation string, userID uint) error {
    if someCondition {
        return fmt.Errorf("用户操作失败: 操作=%s, 用户ID=%d, 原因=权限不足", operation, userID)
    }
    return nil
}
```

### 3. panic的滥用

```go
// ❌ 错误：在业务逻辑中使用panic
func BadPanicUsage(userID uint) *User {
    user, err := getUserFromDB(userID)
    if err != nil {
        panic(fmt.Sprintf("获取用户失败: %v", err)) // 不应该panic
    }

    if user == nil {
        panic("用户不存在") // 业务逻辑错误不应该panic
    }

    return user
}

// ✅ 正确：返回错误而不是panic
func GoodErrorHandling(userID uint) (*User, error) {
    user, err := getUserFromDB(userID)
    if err != nil {
        return nil, fmt.Errorf("获取用户失败: %w", err)
    }

    if user == nil {
        return nil, fmt.Errorf("用户不存在 (ID: %d)", userID)
    }

    return user, nil
}

// ❌ 错误：recover用于正常控制流
func BadRecoverUsage() string {
    defer func() {
        if r := recover(); r != nil {
            // 不应该用recover处理正常的业务逻辑
        }
    }()

    if someCondition {
        panic("业务错误") // 错误的用法
    }

    return "success"
}

// ✅ 正确：只在必要时使用panic/recover
func GoodRecoverUsage() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("意外的panic: %v", r)
            // 记录日志，但不用于正常业务逻辑
        }
    }()

    // 可能panic的第三方代码
    riskyThirdPartyFunction()
}
```

### 4. 错误类型断言的陷阱

```go
// ❌ 错误：直接类型断言
func BadTypeAssertion(err error) {
    if netErr, ok := err.(*net.OpError); ok {
        // 这只能检查直接类型，不能检查包装的错误
        fmt.Printf("网络错误: %v", netErr)
    }
}

// ✅ 正确：使用errors.As检查错误链
func GoodTypeAssertion(err error) {
    var netErr *net.OpError
    if errors.As(err, &netErr) {
        // 这会检查整个错误链
        fmt.Printf("网络错误: %v", netErr)
    }
}

// ❌ 错误：错误的错误比较
func BadErrorComparison(err error) {
    if err.Error() == "file not found" {
        // 字符串比较不可靠，错误消息可能变化
        fmt.Println("文件不存在")
    }
}

// ✅ 正确：使用errors.Is检查错误
func GoodErrorComparison(err error) {
    if errors.Is(err, os.ErrNotExist) {
        // 使用预定义的错误值进行比较
        fmt.Println("文件不存在")
    }
}
```

### 5. 并发环境下的错误处理

```go
// ❌ 错误：并发访问错误变量
func BadConcurrentError() error {
    var lastErr error
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if err := processTask(id); err != nil {
                lastErr = err // 竞态条件！
            }
        }(i)
    }

    wg.Wait()
    return lastErr
}

// ✅ 正确：使用互斥锁或通道保护错误
func GoodConcurrentError() error {
    var mu sync.Mutex
    var errors []error
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if err := processTask(id); err != nil {
                mu.Lock()
                errors = append(errors, err)
                mu.Unlock()
            }
        }(i)
    }

    wg.Wait()

    if len(errors) > 0 {
        return fmt.Errorf("处理失败，错误数量: %d", len(errors))
    }
    return nil
}

// 更好的方式：使用错误通道
func BetterConcurrentError() error {
    errChan := make(chan error, 10)
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            if err := processTask(id); err != nil {
                errChan <- err
            }
        }(i)
    }

    go func() {
        wg.Wait()
        close(errChan)
    }()

    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("处理失败，错误数量: %d", len(errors))
    }
    return nil
}
```

---

## 📝 本章练习题

### 基础练习

1. **基本错误创建和检查**
```go
// 实现一个文件处理函数，要求：
// 1. 读取文件内容
// 2. 验证文件格式（必须是JSON）
// 3. 解析JSON内容
// 4. 返回解析后的数据或详细的错误信息

// 参考答案：
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type Config struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Debug   bool   `json:"debug"`
}

func ProcessConfigFile(filename string) (*Config, error) {
    // 1. 验证文件扩展名
    if !strings.HasSuffix(strings.ToLower(filename), ".json") {
        return nil, fmt.Errorf("文件格式错误: 期望JSON文件，得到 %s", filepath.Ext(filename))
    }

    // 2. 读取文件内容
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取文件失败 (%s): %w", filename, err)
    }

    // 3. 检查文件是否为空
    if len(data) == 0 {
        return nil, fmt.Errorf("文件为空: %s", filename)
    }

    // 4. 解析JSON
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("JSON解析失败 (%s): %w", filename, err)
    }

    // 5. 验证必填字段
    if config.Name == "" {
        return nil, fmt.Errorf("配置验证失败: name字段不能为空")
    }

    if config.Version == "" {
        return nil, fmt.Errorf("配置验证失败: version字段不能为空")
    }

    return &config, nil
}

func main() {
    config, err := ProcessConfigFile("config.json")
    if err != nil {
        fmt.Printf("处理配置文件失败: %v\n", err)
        return
    }

    fmt.Printf("配置加载成功: %+v\n", config)
}
```

2. **错误包装和传播**
```go
// 实现一个用户服务，包含完整的错误处理链：
// 1. 数据验证错误
// 2. 数据库操作错误
// 3. 业务逻辑错误
// 4. 错误包装和传播

// 参考答案：
package main

import (
    "errors"
    "fmt"
    "regexp"
    "strings"
)

// 预定义错误
var (
    ErrUserNotFound    = errors.New("用户不存在")
    ErrEmailExists     = errors.New("邮箱已存在")
    ErrInvalidEmail    = errors.New("邮箱格式无效")
    ErrPasswordTooWeak = errors.New("密码强度不足")
)

type User struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"-"`
}

type UserService struct {
    users []User // 模拟数据库
}

func NewUserService() *UserService {
    return &UserService{
        users: make([]User, 0),
    }
}

// 创建用户 - 完整的错误处理链
func (s *UserService) CreateUser(name, email, password string) (*User, error) {
    // 第一层：输入验证
    if err := s.validateUserInput(name, email, password); err != nil {
        return nil, fmt.Errorf("用户输入验证失败: %w", err)
    }

    // 第二层：业务规则检查
    if err := s.checkBusinessRules(email); err != nil {
        return nil, fmt.Errorf("业务规则检查失败: %w", err)
    }

    // 第三层：数据库操作
    user := &User{
        ID:       uint(len(s.users) + 1),
        Name:     name,
        Email:    email,
        Password: password, // 实际应用中应该加密
    }

    if err := s.saveUser(user); err != nil {
        return nil, fmt.Errorf("保存用户失败: %w", err)
    }

    return user, nil
}

func (s *UserService) validateUserInput(name, email, password string) error {
    var validationErrors []string

    // 验证姓名
    if strings.TrimSpace(name) == "" {
        validationErrors = append(validationErrors, "姓名不能为空")
    }

    // 验证邮箱
    if !s.isValidEmail(email) {
        validationErrors = append(validationErrors, "邮箱格式无效")
    }

    // 验证密码
    if len(password) < 8 {
        validationErrors = append(validationErrors, "密码长度至少8位")
    }

    if !s.isStrongPassword(password) {
        validationErrors = append(validationErrors, "密码必须包含大小写字母和数字")
    }

    if len(validationErrors) > 0 {
        return fmt.Errorf("输入验证失败: %s", strings.Join(validationErrors, ", "))
    }

    return nil
}

func (s *UserService) checkBusinessRules(email string) error {
    // 检查邮箱是否已存在
    for _, user := range s.users {
        if user.Email == email {
            return fmt.Errorf("邮箱冲突 (%s): %w", email, ErrEmailExists)
        }
    }

    return nil
}

func (s *UserService) saveUser(user *User) error {
    // 模拟数据库保存可能失败
    if len(s.users) >= 100 {
        return fmt.Errorf("数据库容量已满，无法保存用户")
    }

    s.users = append(s.users, *user)
    return nil
}

func (s *UserService) isValidEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

func (s *UserService) isStrongPassword(password string) bool {
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasDigit := regexp.MustCompile(`\d`).MatchString(password)

    return hasUpper && hasLower && hasDigit
}

func main() {
    service := NewUserService()

    // 测试各种错误情况
    testCases := []struct {
        name     string
        email    string
        password string
        desc     string
    }{
        {"", "test@example.com", "Password123", "空姓名"},
        {"张三", "invalid-email", "Password123", "无效邮箱"},
        {"张三", "test@example.com", "123", "弱密码"},
        {"张三", "test@example.com", "Password123", "正常用户"},
        {"李四", "test@example.com", "Password123", "重复邮箱"},
    }

    for _, tc := range testCases {
        user, err := service.CreateUser(tc.name, tc.email, tc.password)
        if err != nil {
            fmt.Printf("❌ %s: %v\n", tc.desc, err)
        } else {
            fmt.Printf("✅ %s: 用户创建成功 (ID: %d)\n", tc.desc, user.ID)
        }
    }
}
```

### 中级练习

3. **自定义错误类型设计**
```go
// 设计一个电商系统的错误处理体系，包括：
// 1. 多种错误类型（验证、业务、系统）
// 2. 错误分类和错误码
// 3. 错误详情和上下文信息
// 4. 统一的错误处理接口

// 参考答案：
package main

import (
    "encoding/json"
    "fmt"
    "time"
)

// 错误分类常量
const (
    ErrorTypeValidation = "VALIDATION"
    ErrorTypeBusiness   = "BUSINESS"
    ErrorTypeSystem     = "SYSTEM"
    ErrorTypeNetwork    = "NETWORK"
)

// 错误码常量
const (
    CodeInvalidInput     = "E001"
    CodeUserNotFound     = "E002"
    CodeInsufficientFund = "E003"
    CodeProductOutStock  = "E004"
    CodeDatabaseError    = "E005"
    CodeNetworkTimeout   = "E006"
)

// 基础错误接口
type AppError interface {
    error
    GetCode() string
    GetType() string
    GetDetails() map[string]interface{}
    GetTimestamp() time.Time
    ToJSON() string
}

// 通用错误结构
type BaseError struct {
    Code      string                 `json:"code"`
    Type      string                 `json:"type"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Cause     error                  `json:"-"`
}

func (e *BaseError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *BaseError) GetCode() string {
    return e.Code
}

func (e *BaseError) GetType() string {
    return e.Type
}

func (e *BaseError) GetDetails() map[string]interface{} {
    return e.Details
}

func (e *BaseError) GetTimestamp() time.Time {
    return e.Timestamp
}

func (e *BaseError) ToJSON() string {
    data, _ := json.Marshal(e)
    return string(data)
}

func (e *BaseError) Unwrap() error {
    return e.Cause
}

// 验证错误
type ValidationError struct {
    *BaseError
    Field string `json:"field"`
    Value interface{} `json:"value"`
}

func NewValidationError(field string, value interface{}, message string) *ValidationError {
    return &ValidationError{
        BaseError: &BaseError{
            Code:      CodeInvalidInput,
            Type:      ErrorTypeValidation,
            Message:   message,
            Details:   map[string]interface{}{"field": field, "value": value},
            Timestamp: time.Now(),
        },
        Field: field,
        Value: value,
    }
}

// 业务错误
type BusinessError struct {
    *BaseError
    Operation string `json:"operation"`
}

func NewBusinessError(code, message, operation string) *BusinessError {
    return &BusinessError{
        BaseError: &BaseError{
            Code:      code,
            Type:      ErrorTypeBusiness,
            Message:   message,
            Details:   map[string]interface{}{"operation": operation},
            Timestamp: time.Now(),
        },
        Operation: operation,
    }
}

// 系统错误
type SystemError struct {
    *BaseError
    Component string `json:"component"`
}

func NewSystemError(code, message, component string, cause error) *SystemError {
    return &SystemError{
        BaseError: &BaseError{
            Code:      code,
            Type:      ErrorTypeSystem,
            Message:   message,
            Details:   map[string]interface{}{"component": component},
            Timestamp: time.Now(),
            Cause:     cause,
        },
        Component: component,
    }
}

// 电商服务示例
type OrderService struct {
    // 模拟依赖
}

func (s *OrderService) CreateOrder(userID uint, productID uint, quantity int) error {
    // 验证输入
    if userID == 0 {
        return NewValidationError("userID", userID, "用户ID不能为0")
    }

    if productID == 0 {
        return NewValidationError("productID", productID, "商品ID不能为0")
    }

    if quantity <= 0 {
        return NewValidationError("quantity", quantity, "数量必须大于0")
    }

    // 检查用户是否存在
    if !s.userExists(userID) {
        return NewBusinessError(CodeUserNotFound, "用户不存在", "create_order")
    }

    // 检查商品库存
    if !s.hasStock(productID, quantity) {
        err := NewBusinessError(CodeProductOutStock, "商品库存不足", "create_order")
        err.Details["product_id"] = productID
        err.Details["requested_quantity"] = quantity
        err.Details["available_stock"] = s.getStock(productID)
        return err
    }

    // 检查用户余额
    if !s.hasSufficientFund(userID, s.getProductPrice(productID)*float64(quantity)) {
        err := NewBusinessError(CodeInsufficientFund, "余额不足", "create_order")
        err.Details["user_id"] = userID
        err.Details["required_amount"] = s.getProductPrice(productID) * float64(quantity)
        err.Details["available_balance"] = s.getUserBalance(userID)
        return err
    }

    // 模拟数据库操作失败
    if err := s.saveOrder(userID, productID, quantity); err != nil {
        return NewSystemError(CodeDatabaseError, "订单保存失败", "database", err)
    }

    return nil
}

// 模拟方法
func (s *OrderService) userExists(userID uint) bool {
    return userID != 999 // 模拟用户999不存在
}

func (s *OrderService) hasStock(productID uint, quantity int) bool {
    return productID != 888 || quantity <= 5 // 模拟商品888库存只有5个
}

func (s *OrderService) getStock(productID uint) int {
    if productID == 888 {
        return 5
    }
    return 100
}

func (s *OrderService) hasSufficientFund(userID uint, amount float64) bool {
    return s.getUserBalance(userID) >= amount
}

func (s *OrderService) getUserBalance(userID uint) float64 {
    if userID == 777 {
        return 10.0 // 模拟用户777余额不足
    }
    return 1000.0
}

func (s *OrderService) getProductPrice(productID uint) float64 {
    return 99.99
}

func (s *OrderService) saveOrder(userID, productID uint, quantity int) error {
    if userID == 666 {
        return fmt.Errorf("数据库连接失败")
    }
    return nil
}

// 错误处理器
func HandleError(err error) {
    if err == nil {
        return
    }

    // 检查是否是应用错误
    var appErr AppError
    if errors.As(err, &appErr) {
        fmt.Printf("应用错误: %s\n", appErr.ToJSON())

        // 根据错误类型进行不同处理
        switch appErr.GetType() {
        case ErrorTypeValidation:
            fmt.Println("→ 返回400 Bad Request")
        case ErrorTypeBusiness:
            fmt.Println("→ 返回422 Unprocessable Entity")
        case ErrorTypeSystem:
            fmt.Println("→ 返回500 Internal Server Error")
            fmt.Println("→ 发送告警通知")
        }
    } else {
        // 未知错误
        fmt.Printf("未知错误: %v\n", err)
        fmt.Println("→ 返回500 Internal Server Error")
    }
}

func main() {
    service := &OrderService{}

    // 测试各种错误情况
    testCases := []struct {
        userID    uint
        productID uint
        quantity  int
        desc      string
    }{
        {0, 123, 1, "无效用户ID"},
        {123, 0, 1, "无效商品ID"},
        {123, 456, -1, "无效数量"},
        {999, 456, 1, "用户不存在"},
        {123, 888, 10, "库存不足"},
        {777, 456, 1, "余额不足"},
        {666, 456, 1, "数据库错误"},
        {123, 456, 1, "正常订单"},
    }

    for _, tc := range testCases {
        fmt.Printf("\n=== 测试: %s ===\n", tc.desc)
        err := service.CreateOrder(tc.userID, tc.productID, tc.quantity)
        if err != nil {
            HandleError(err)
        } else {
            fmt.Println("✅ 订单创建成功")
        }
    }
}
```

4. **并发错误处理**
```go
// 实现一个并发安全的批处理系统，要求：
// 1. 并发处理多个任务
// 2. 收集所有错误信息
// 3. 支持错误重试机制
// 4. 提供详细的处理报告

// 参考答案：
package main

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"
)

// 任务接口
type Task interface {
    GetID() string
    Execute() error
    GetRetryCount() int
    SetRetryCount(count int)
}

// 具体任务实现
type DataProcessTask struct {
    ID         string
    Data       string
    retryCount int
}

func (t *DataProcessTask) GetID() string {
    return t.ID
}

func (t *DataProcessTask) Execute() error {
    // 模拟任务执行，30%概率失败
    time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

    if rand.Float32() < 0.3 {
        return fmt.Errorf("任务 %s 处理失败: 数据格式错误", t.ID)
    }

    return nil
}

func (t *DataProcessTask) GetRetryCount() int {
    return t.retryCount
}

func (t *DataProcessTask) SetRetryCount(count int) {
    t.retryCount = count
}

// 任务结果
type TaskResult struct {
    TaskID    string
    Success   bool
    Error     error
    Attempts  int
    Duration  time.Duration
}

// 批处理报告
type BatchReport struct {
    TotalTasks    int
    SuccessTasks  int
    FailedTasks   int
    TotalDuration time.Duration
    Results       []TaskResult
    Errors        []error
}

func (r *BatchReport) GetSuccessRate() float64 {
    if r.TotalTasks == 0 {
        return 0
    }
    return float64(r.SuccessTasks) / float64(r.TotalTasks) * 100
}

// 批处理器
type BatchProcessor struct {
    maxWorkers   int
    maxRetries   int
    retryDelay   time.Duration
    timeout      time.Duration
}

func NewBatchProcessor(maxWorkers, maxRetries int, retryDelay, timeout time.Duration) *BatchProcessor {
    return &BatchProcessor{
        maxWorkers: maxWorkers,
        maxRetries: maxRetries,
        retryDelay: retryDelay,
        timeout:    timeout,
    }
}

func (bp *BatchProcessor) ProcessTasks(ctx context.Context, tasks []Task) *BatchReport {
    startTime := time.Now()

    // 创建任务通道
    taskChan := make(chan Task, len(tasks))
    resultChan := make(chan TaskResult, len(tasks))

    // 发送任务到通道
    for _, task := range tasks {
        taskChan <- task
    }
    close(taskChan)

    // 启动工作协程
    var wg sync.WaitGroup
    for i := 0; i < bp.maxWorkers; i++ {
        wg.Add(1)
        go bp.worker(ctx, &wg, taskChan, resultChan)
    }

    // 等待所有工作协程完成
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // 收集结果
    report := &BatchReport{
        TotalTasks: len(tasks),
        Results:    make([]TaskResult, 0, len(tasks)),
        Errors:     make([]error, 0),
    }

    for result := range resultChan {
        report.Results = append(report.Results, result)

        if result.Success {
            report.SuccessTasks++
        } else {
            report.FailedTasks++
            if result.Error != nil {
                report.Errors = append(report.Errors,
                    fmt.Errorf("任务 %s 失败: %w", result.TaskID, result.Error))
            }
        }
    }

    report.TotalDuration = time.Since(startTime)
    return report
}

func (bp *BatchProcessor) worker(ctx context.Context, wg *sync.WaitGroup, taskChan <-chan Task, resultChan chan<- TaskResult) {
    defer wg.Done()

    for task := range taskChan {
        select {
        case <-ctx.Done():
            // 上下文取消，停止处理
            resultChan <- TaskResult{
                TaskID:  task.GetID(),
                Success: false,
                Error:   ctx.Err(),
            }
            return
        default:
            result := bp.executeTaskWithRetry(ctx, task)
            resultChan <- result
        }
    }
}

func (bp *BatchProcessor) executeTaskWithRetry(ctx context.Context, task Task) TaskResult {
    startTime := time.Now()
    var lastErr error

    for attempt := 1; attempt <= bp.maxRetries+1; attempt++ {
        // 检查上下文是否取消
        select {
        case <-ctx.Done():
            return TaskResult{
                TaskID:   task.GetID(),
                Success:  false,
                Error:    ctx.Err(),
                Attempts: attempt - 1,
                Duration: time.Since(startTime),
            }
        default:
        }

        // 执行任务
        err := bp.executeWithTimeout(ctx, task)
        if err == nil {
            return TaskResult{
                TaskID:   task.GetID(),
                Success:  true,
                Attempts: attempt,
                Duration: time.Since(startTime),
            }
        }

        lastErr = err
        task.SetRetryCount(attempt)

        // 如果不是最后一次尝试，等待重试延迟
        if attempt <= bp.maxRetries {
            select {
            case <-ctx.Done():
                return TaskResult{
                    TaskID:   task.GetID(),
                    Success:  false,
                    Error:    ctx.Err(),
                    Attempts: attempt,
                    Duration: time.Since(startTime),
                }
            case <-time.After(bp.retryDelay):
                // 继续重试
            }
        }
    }

    return TaskResult{
        TaskID:   task.GetID(),
        Success:  false,
        Error:    fmt.Errorf("任务执行失败，已重试 %d 次: %w", bp.maxRetries, lastErr),
        Attempts: bp.maxRetries + 1,
        Duration: time.Since(startTime),
    }
}

func (bp *BatchProcessor) executeWithTimeout(ctx context.Context, task Task) error {
    // 创建带超时的上下文
    timeoutCtx, cancel := context.WithTimeout(ctx, bp.timeout)
    defer cancel()

    // 使用通道来处理超时
    errChan := make(chan error, 1)

    go func() {
        errChan <- task.Execute()
    }()

    select {
    case err := <-errChan:
        return err
    case <-timeoutCtx.Done():
        return fmt.Errorf("任务执行超时: %w", timeoutCtx.Err())
    }
}

func main() {
    // 创建测试任务
    tasks := make([]Task, 20)
    for i := 0; i < 20; i++ {
        tasks[i] = &DataProcessTask{
            ID:   fmt.Sprintf("task-%02d", i+1),
            Data: fmt.Sprintf("data-%d", i+1),
        }
    }

    // 创建批处理器
    processor := NewBatchProcessor(
        5,                      // 5个工作协程
        2,                      // 最多重试2次
        100*time.Millisecond,   // 重试延迟100ms
        500*time.Millisecond,   // 任务超时500ms
    )

    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    fmt.Println("开始批处理任务...")
    report := processor.ProcessTasks(ctx, tasks)

    // 打印报告
    fmt.Printf("\n=== 批处理报告 ===\n")
    fmt.Printf("总任务数: %d\n", report.TotalTasks)
    fmt.Printf("成功任务: %d\n", report.SuccessTasks)
    fmt.Printf("失败任务: %d\n", report.FailedTasks)
    fmt.Printf("成功率: %.2f%%\n", report.GetSuccessRate())
    fmt.Printf("总耗时: %v\n", report.TotalDuration)

    if len(report.Errors) > 0 {
        fmt.Printf("\n=== 错误详情 ===\n")
        for i, err := range report.Errors {
            fmt.Printf("%d. %v\n", i+1, err)
        }
    }

    fmt.Printf("\n=== 任务详情 ===\n")
    for _, result := range report.Results {
        status := "✅"
        if !result.Success {
            status = "❌"
        }
        fmt.Printf("%s %s - 尝试次数: %d, 耗时: %v\n",
            status, result.TaskID, result.Attempts, result.Duration)
    }
}
```

### 高级练习

5. **企业级错误处理系统设计**
```go
// 设计一个完整的企业级错误处理系统，包括：
// 1. 分布式错误追踪
// 2. 错误聚合和分析
// 3. 自动告警和恢复
// 4. 错误指标监控
// 5. 多语言错误消息支持

// 参考答案：
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/google/uuid"
)

// 错误严重程度
type Severity int

const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityError
    SeverityCritical
)

func (s Severity) String() string {
    switch s {
    case SeverityInfo:
        return "INFO"
    case SeverityWarning:
        return "WARNING"
    case SeverityError:
        return "ERROR"
    case SeverityCritical:
        return "CRITICAL"
    default:
        return "UNKNOWN"
    }
}

// 错误事件
type ErrorEvent struct {
    ID          string                 `json:"id"`
    TraceID     string                 `json:"trace_id"`
    SpanID      string                 `json:"span_id"`
    Service     string                 `json:"service"`
    Operation   string                 `json:"operation"`
    ErrorCode   string                 `json:"error_code"`
    ErrorType   string                 `json:"error_type"`
    Message     string                 `json:"message"`
    Severity    Severity               `json:"severity"`
    Timestamp   time.Time              `json:"timestamp"`
    UserID      string                 `json:"user_id,omitempty"`
    SessionID   string                 `json:"session_id,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    StackTrace  string                 `json:"stack_trace,omitempty"`
    Context     map[string]interface{} `json:"context,omitempty"`
    Tags        map[string]string      `json:"tags,omitempty"`
    Fingerprint string                 `json:"fingerprint"` // 用于错误聚合
}

// 错误聚合信息
type ErrorAggregate struct {
    Fingerprint   string    `json:"fingerprint"`
    Count         int       `json:"count"`
    FirstSeen     time.Time `json:"first_seen"`
    LastSeen      time.Time `json:"last_seen"`
    Service       string    `json:"service"`
    ErrorCode     string    `json:"error_code"`
    Message       string    `json:"message"`
    Severity      Severity  `json:"severity"`
    AffectedUsers []string  `json:"affected_users"`
}

// 告警规则
type AlertRule struct {
    ID          string        `json:"id"`
    Name        string        `json:"name"`
    Service     string        `json:"service"`
    ErrorCode   string        `json:"error_code"`
    Severity    Severity      `json:"severity"`
    Threshold   int           `json:"threshold"`
    TimeWindow  time.Duration `json:"time_window"`
    Enabled     bool          `json:"enabled"`
}

// 告警通知
type AlertNotification struct {
    RuleID      string         `json:"rule_id"`
    RuleName    string         `json:"rule_name"`
    Aggregate   ErrorAggregate `json:"aggregate"`
    Timestamp   time.Time      `json:"timestamp"`
    Message     string         `json:"message"`
}

// 错误处理器接口
type ErrorHandler interface {
    HandleError(ctx context.Context, event *ErrorEvent) error
}

// 错误存储接口
type ErrorStorage interface {
    Store(ctx context.Context, event *ErrorEvent) error
    GetAggregates(ctx context.Context, filters map[string]interface{}) ([]ErrorAggregate, error)
    UpdateAggregate(ctx context.Context, aggregate *ErrorAggregate) error
}

// 告警器接口
type Alerter interface {
    SendAlert(ctx context.Context, notification *AlertNotification) error
}

// 内存存储实现（生产环境应使用数据库）
type MemoryStorage struct {
    events     []ErrorEvent
    aggregates map[string]*ErrorAggregate
    mutex      sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        events:     make([]ErrorEvent, 0),
        aggregates: make(map[string]*ErrorAggregate),
    }
}

func (ms *MemoryStorage) Store(ctx context.Context, event *ErrorEvent) error {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    // 存储事件
    ms.events = append(ms.events, *event)

    // 更新聚合信息
    if aggregate, exists := ms.aggregates[event.Fingerprint]; exists {
        aggregate.Count++
        aggregate.LastSeen = event.Timestamp

        // 添加受影响的用户
        if event.UserID != "" {
            found := false
            for _, userID := range aggregate.AffectedUsers {
                if userID == event.UserID {
                    found = true
                    break
                }
            }
            if !found {
                aggregate.AffectedUsers = append(aggregate.AffectedUsers, event.UserID)
            }
        }
    } else {
        // 创建新的聚合
        aggregate := &ErrorAggregate{
            Fingerprint: event.Fingerprint,
            Count:       1,
            FirstSeen:   event.Timestamp,
            LastSeen:    event.Timestamp,
            Service:     event.Service,
            ErrorCode:   event.ErrorCode,
            Message:     event.Message,
            Severity:    event.Severity,
            AffectedUsers: make([]string, 0),
        }

        if event.UserID != "" {
            aggregate.AffectedUsers = append(aggregate.AffectedUsers, event.UserID)
        }

        ms.aggregates[event.Fingerprint] = aggregate
    }

    return nil
}

func (ms *MemoryStorage) GetAggregates(ctx context.Context, filters map[string]interface{}) ([]ErrorAggregate, error) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    result := make([]ErrorAggregate, 0, len(ms.aggregates))
    for _, aggregate := range ms.aggregates {
        result = append(result, *aggregate)
    }

    return result, nil
}

func (ms *MemoryStorage) UpdateAggregate(ctx context.Context, aggregate *ErrorAggregate) error {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    ms.aggregates[aggregate.Fingerprint] = aggregate
    return nil
}

// 控制台告警器
type ConsoleAlerter struct{}

func (ca *ConsoleAlerter) SendAlert(ctx context.Context, notification *AlertNotification) error {
    fmt.Printf("🚨 告警通知 🚨\n")
    fmt.Printf("规则: %s\n", notification.RuleName)
    fmt.Printf("服务: %s\n", notification.Aggregate.Service)
    fmt.Printf("错误: %s\n", notification.Aggregate.Message)
    fmt.Printf("发生次数: %d\n", notification.Aggregate.Count)
    fmt.Printf("严重程度: %s\n", notification.Aggregate.Severity)
    fmt.Printf("受影响用户: %d\n", len(notification.Aggregate.AffectedUsers))
    fmt.Printf("时间: %s\n", notification.Timestamp.Format("2006-01-02 15:04:05"))
    fmt.Println("=" * 50)
    return nil
}

// 企业级错误处理系统
type EnterpriseErrorSystem struct {
    storage    ErrorStorage
    alerter    Alerter
    rules      []AlertRule
    handlers   []ErrorHandler

    // 错误指标
    metrics    *ErrorMetrics

    // 多语言支持
    i18n       *I18nManager

    // 后台处理
    alertChan  chan *ErrorEvent
    stopChan   chan struct{}
    wg         sync.WaitGroup
}

// 错误指标
type ErrorMetrics struct {
    TotalErrors    int64
    ErrorsByCode   map[string]int64
    ErrorsByService map[string]int64
    mutex          sync.RWMutex
}

func NewErrorMetrics() *ErrorMetrics {
    return &ErrorMetrics{
        ErrorsByCode:    make(map[string]int64),
        ErrorsByService: make(map[string]int64),
    }
}

func (em *ErrorMetrics) IncrementError(service, errorCode string) {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    em.TotalErrors++
    em.ErrorsByCode[errorCode]++
    em.ErrorsByService[service]++
}

func (em *ErrorMetrics) GetMetrics() map[string]interface{} {
    em.mutex.RLock()
    defer em.mutex.RUnlock()

    return map[string]interface{}{
        "total_errors":      em.TotalErrors,
        "errors_by_code":    em.ErrorsByCode,
        "errors_by_service": em.ErrorsByService,
    }
}

// 国际化管理器
type I18nManager struct {
    messages map[string]map[string]string // [language][key]message
}

func NewI18nManager() *I18nManager {
    return &I18nManager{
        messages: map[string]map[string]string{
            "zh": {
                "user_not_found":     "用户不存在",
                "invalid_input":      "输入参数无效",
                "database_error":     "数据库操作失败",
                "network_timeout":    "网络连接超时",
                "insufficient_funds": "账户余额不足",
            },
            "en": {
                "user_not_found":     "User not found",
                "invalid_input":      "Invalid input parameters",
                "database_error":     "Database operation failed",
                "network_timeout":    "Network connection timeout",
                "insufficient_funds": "Insufficient account balance",
            },
        },
    }
}

func (i18n *I18nManager) GetMessage(language, key string) string {
    if langMessages, exists := i18n.messages[language]; exists {
        if message, exists := langMessages[key]; exists {
            return message
        }
    }

    // 回退到英文
    if langMessages, exists := i18n.messages["en"]; exists {
        if message, exists := langMessages[key]; exists {
            return message
        }
    }

    return key // 如果都找不到，返回key本身
}

func NewEnterpriseErrorSystem(storage ErrorStorage, alerter Alerter) *EnterpriseErrorSystem {
    system := &EnterpriseErrorSystem{
        storage:   storage,
        alerter:   alerter,
        rules:     make([]AlertRule, 0),
        handlers:  make([]ErrorHandler, 0),
        metrics:   NewErrorMetrics(),
        i18n:      NewI18nManager(),
        alertChan: make(chan *ErrorEvent, 1000),
        stopChan:  make(chan struct{}),
    }

    // 启动后台处理协程
    system.wg.Add(1)
    go system.processAlerts()

    return system
}

func (ees *EnterpriseErrorSystem) AddAlertRule(rule AlertRule) {
    ees.rules = append(ees.rules, rule)
}

func (ees *EnterpriseErrorSystem) AddHandler(handler ErrorHandler) {
    ees.handlers = append(ees.handlers, handler)
}

func (ees *EnterpriseErrorSystem) ReportError(ctx context.Context, service, operation, errorCode, message string, severity Severity) error {
    // 生成错误指纹用于聚合
    fingerprint := fmt.Sprintf("%s:%s:%s", service, errorCode, message)

    // 创建错误事件
    event := &ErrorEvent{
        ID:          uuid.New().String(),
        TraceID:     getTraceID(ctx),
        SpanID:      getSpanID(ctx),
        Service:     service,
        Operation:   operation,
        ErrorCode:   errorCode,
        ErrorType:   getErrorType(errorCode),
        Message:     message,
        Severity:    severity,
        Timestamp:   time.Now(),
        UserID:      getUserID(ctx),
        SessionID:   getSessionID(ctx),
        RequestID:   getRequestID(ctx),
        Context:     getContextData(ctx),
        Tags:        getContextTags(ctx),
        Fingerprint: fingerprint,
    }

    // 更新指标
    ees.metrics.IncrementError(service, errorCode)

    // 存储错误事件
    if err := ees.storage.Store(ctx, event); err != nil {
        log.Printf("存储错误事件失败: %v", err)
    }

    // 调用处理器
    for _, handler := range ees.handlers {
        if err := handler.HandleError(ctx, event); err != nil {
            log.Printf("错误处理器执行失败: %v", err)
        }
    }

    // 发送到告警处理通道
    select {
    case ees.alertChan <- event:
    default:
        log.Println("告警通道已满，丢弃错误事件")
    }

    return nil
}

func (ees *EnterpriseErrorSystem) processAlerts() {
    defer ees.wg.Done()

    ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次告警
    defer ticker.Stop()

    for {
        select {
        case <-ees.stopChan:
            return
        case <-ticker.C:
            ees.checkAlertRules()
        }
    }
}

func (ees *EnterpriseErrorSystem) checkAlertRules() {
    ctx := context.Background()

    // 获取聚合数据
    aggregates, err := ees.storage.GetAggregates(ctx, nil)
    if err != nil {
        log.Printf("获取聚合数据失败: %v", err)
        return
    }

    // 检查每个告警规则
    for _, rule := range ees.rules {
        if !rule.Enabled {
            continue
        }

        for _, aggregate := range aggregates {
            if ees.shouldTriggerAlert(rule, aggregate) {
                notification := &AlertNotification{
                    RuleID:    rule.ID,
                    RuleName:  rule.Name,
                    Aggregate: aggregate,
                    Timestamp: time.Now(),
                    Message:   fmt.Sprintf("错误 %s 在 %v 内发生了 %d 次，超过阈值 %d",
                        aggregate.ErrorCode, rule.TimeWindow, aggregate.Count, rule.Threshold),
                }

                if err := ees.alerter.SendAlert(ctx, notification); err != nil {
                    log.Printf("发送告警失败: %v", err)
                }
            }
        }
    }
}

func (ees *EnterpriseErrorSystem) shouldTriggerAlert(rule AlertRule, aggregate ErrorAggregate) bool {
    // 检查服务匹配
    if rule.Service != "" && rule.Service != aggregate.Service {
        return false
    }

    // 检查错误码匹配
    if rule.ErrorCode != "" && rule.ErrorCode != aggregate.ErrorCode {
        return false
    }

    // 检查严重程度
    if aggregate.Severity < rule.Severity {
        return false
    }

    // 检查时间窗口内的错误次数
    if time.Since(aggregate.LastSeen) > rule.TimeWindow {
        return false
    }

    // 检查是否超过阈值
    return aggregate.Count >= rule.Threshold
}

func (ees *EnterpriseErrorSystem) GetMetrics() map[string]interface{} {
    return ees.metrics.GetMetrics()
}

func (ees *EnterpriseErrorSystem) GetLocalizedMessage(language, errorCode string) string {
    return ees.i18n.GetMessage(language, errorCode)
}

func (ees *EnterpriseErrorSystem) Stop() {
    close(ees.stopChan)
    ees.wg.Wait()
}

// 上下文辅助函数
func getTraceID(ctx context.Context) string {
    if traceID := ctx.Value("trace_id"); traceID != nil {
        return traceID.(string)
    }
    return uuid.New().String()
}

func getSpanID(ctx context.Context) string {
    if spanID := ctx.Value("span_id"); spanID != nil {
        return spanID.(string)
    }
    return uuid.New().String()
}

func getUserID(ctx context.Context) string {
    if userID := ctx.Value("user_id"); userID != nil {
        return fmt.Sprintf("%v", userID)
    }
    return ""
}

func getSessionID(ctx context.Context) string {
    if sessionID := ctx.Value("session_id"); sessionID != nil {
        return sessionID.(string)
    }
    return ""
}

func getRequestID(ctx context.Context) string {
    if requestID := ctx.Value("request_id"); requestID != nil {
        return requestID.(string)
    }
    return ""
}

func getContextData(ctx context.Context) map[string]interface{} {
    if data := ctx.Value("context_data"); data != nil {
        return data.(map[string]interface{})
    }
    return make(map[string]interface{})
}

func getContextTags(ctx context.Context) map[string]string {
    if tags := ctx.Value("context_tags"); tags != nil {
        return tags.(map[string]string)
    }
    return make(map[string]string)
}

func getErrorType(errorCode string) string {
    switch {
    case errorCode[0] == '1':
        return "VALIDATION"
    case errorCode[0] == '2':
        return "BUSINESS"
    case errorCode[0] == '3':
        return "SYSTEM"
    case errorCode[0] == '4':
        return "NETWORK"
    default:
        return "UNKNOWN"
    }
}

// 自定义错误处理器示例
type LoggingErrorHandler struct {
    logger *log.Logger
}

func NewLoggingErrorHandler() *LoggingErrorHandler {
    return &LoggingErrorHandler{
        logger: log.New(os.Stdout, "[ERROR] ", log.LstdFlags),
    }
}

func (leh *LoggingErrorHandler) HandleError(ctx context.Context, event *ErrorEvent) error {
    leh.logger.Printf("Service: %s, Operation: %s, Error: %s, Severity: %s",
        event.Service, event.Operation, event.Message, event.Severity)
    return nil
}

// 使用示例
func main() {
    // 创建存储和告警器
    storage := NewMemoryStorage()
    alerter := &ConsoleAlerter{}

    // 创建企业级错误系统
    errorSystem := NewEnterpriseErrorSystem(storage, alerter)
    defer errorSystem.Stop()

    // 添加错误处理器
    errorSystem.AddHandler(NewLoggingErrorHandler())

    // 添加告警规则
    errorSystem.AddAlertRule(AlertRule{
        ID:         "rule-001",
        Name:       "用户服务高错误率",
        Service:    "user-service",
        ErrorCode:  "2001",
        Severity:   SeverityError,
        Threshold:  5,
        TimeWindow: 5 * time.Minute,
        Enabled:    true,
    })

    errorSystem.AddAlertRule(AlertRule{
        ID:         "rule-002",
        Name:       "系统级严重错误",
        Severity:   SeverityCritical,
        Threshold:  1,
        TimeWindow: 1 * time.Minute,
        Enabled:    true,
    })

    // 模拟错误报告
    ctx := context.Background()
    ctx = context.WithValue(ctx, "user_id", "user123")
    ctx = context.WithValue(ctx, "request_id", "req-456")

    // 报告各种错误
    errorSystem.ReportError(ctx, "user-service", "create_user", "2001", "用户创建失败", SeverityError)
    errorSystem.ReportError(ctx, "user-service", "create_user", "2001", "用户创建失败", SeverityError)
    errorSystem.ReportError(ctx, "user-service", "create_user", "2001", "用户创建失败", SeverityError)
    errorSystem.ReportError(ctx, "order-service", "create_order", "3001", "数据库连接失败", SeverityCritical)

    // 等待一段时间让后台处理完成
    time.Sleep(2 * time.Second)

    // 打印指标
    fmt.Println("\n=== 错误指标 ===")
    metrics := errorSystem.GetMetrics()
    metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
    fmt.Println(string(metricsJSON))

    // 测试多语言支持
    fmt.Println("\n=== 多语言错误消息 ===")
    fmt.Printf("中文: %s\n", errorSystem.GetLocalizedMessage("zh", "user_not_found"))
    fmt.Printf("英文: %s\n", errorSystem.GetLocalizedMessage("en", "user_not_found"))

    fmt.Println("\n企业级错误处理系统演示完成！")
}
```

---

## 🏢 实战案例分析

### Mall-Go项目错误处理架构

让我们深入分析一个真实的Go电商项目中的错误处理实现。

#### 1. 项目结构和错误处理分层

```go
// 来自 mall-go/pkg/errors/errors.go
package errors

import (
    "fmt"
    "net/http"
    "time"
)

// 错误接口定义
type AppError interface {
    error
    Code() string
    HTTPStatus() int
    Details() map[string]interface{}
    WithDetail(key string, value interface{}) AppError
    WithContext(ctx map[string]interface{}) AppError
}

// 基础错误实现
type BaseError struct {
    ErrCode    string                 `json:"code"`
    ErrMessage string                 `json:"message"`
    ErrDetails map[string]interface{} `json:"details,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    cause      error                  // 原始错误，不序列化
}

func (e *BaseError) Error() string {
    if e.cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.ErrCode, e.ErrMessage, e.cause)
    }
    return fmt.Sprintf("[%s] %s", e.ErrCode, e.ErrMessage)
}

func (e *BaseError) Code() string {
    return e.ErrCode
}

func (e *BaseError) HTTPStatus() int {
    // 根据错误码前缀确定HTTP状态码
    switch e.ErrCode[:1] {
    case "1": // 1xxx - 客户端错误
        return http.StatusBadRequest
    case "2": // 2xxx - 认证授权错误
        return http.StatusUnauthorized
    case "3": // 3xxx - 业务逻辑错误
        return http.StatusUnprocessableEntity
    case "4": // 4xxx - 资源不存在
        return http.StatusNotFound
    case "5": // 5xxx - 服务器错误
        return http.StatusInternalServerError
    default:
        return http.StatusInternalServerError
    }
}

func (e *BaseError) Details() map[string]interface{} {
    return e.ErrDetails
}

func (e *BaseError) WithDetail(key string, value interface{}) AppError {
    if e.ErrDetails == nil {
        e.ErrDetails = make(map[string]interface{})
    }
    e.ErrDetails[key] = value
    return e
}

func (e *BaseError) WithContext(ctx map[string]interface{}) AppError {
    if e.ErrDetails == nil {
        e.ErrDetails = make(map[string]interface{})
    }
    for k, v := range ctx {
        e.ErrDetails[k] = v
    }
    return e
}

func (e *BaseError) Unwrap() error {
    return e.cause
}

// 错误构造函数
func New(code, message string) AppError {
    return &BaseError{
        ErrCode:    code,
        ErrMessage: message,
        ErrDetails: make(map[string]interface{}),
        Timestamp:  time.Now(),
    }
}

func Wrap(err error, code, message string) AppError {
    return &BaseError{
        ErrCode:    code,
        ErrMessage: message,
        ErrDetails: make(map[string]interface{}),
        Timestamp:  time.Now(),
        cause:      err,
    }
}

// 预定义错误常量
const (
    // 用户相关错误 (1xxx)
    CodeInvalidInput     = "1001"
    CodeMissingParameter = "1002"
    CodeInvalidFormat    = "1003"

    // 认证授权错误 (2xxx)
    CodeUnauthorized     = "2001"
    CodeTokenExpired     = "2002"
    CodeInsufficientPerm = "2003"

    // 业务逻辑错误 (3xxx)
    CodeUserNotFound     = "3001"
    CodeEmailExists      = "3002"
    CodeInsufficientFund = "3003"
    CodeOrderNotFound    = "3004"

    // 资源错误 (4xxx)
    CodeResourceNotFound = "4001"
    CodeResourceConflict = "4002"

    // 系统错误 (5xxx)
    CodeDatabaseError    = "5001"
    CodeNetworkError     = "5002"
    CodeInternalError    = "5003"
)

// 便捷构造函数
func InvalidInput(message string) AppError {
    return New(CodeInvalidInput, message)
}

func Unauthorized(message string) AppError {
    return New(CodeUnauthorized, message)
}

func UserNotFound(userID interface{}) AppError {
    return New(CodeUserNotFound, "用户不存在").
        WithDetail("user_id", userID)
}

func DatabaseError(err error) AppError {
    return Wrap(err, CodeDatabaseError, "数据库操作失败")
}
```

#### 2. 服务层错误处理实现

```go
// 来自 mall-go/internal/service/user_service.go
package service

import (
    "context"
    "fmt"
    "regexp"

    "github.com/yourname/mall-go/pkg/errors"
    "github.com/yourname/mall-go/internal/model"
    "github.com/yourname/mall-go/internal/repository"
    "go.uber.org/zap"
)

type UserService struct {
    userRepo repository.UserRepository
    logger   *zap.Logger
    emailRegex *regexp.Regexp
}

func NewUserService(userRepo repository.UserRepository, logger *zap.Logger) *UserService {
    return &UserService{
        userRepo:   userRepo,
        logger:     logger,
        emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    }
}

// 用户注册 - 完整的错误处理链
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
    // 第一层：输入验证
    if err := s.validateRegisterRequest(req); err != nil {
        s.logger.Warn("用户注册输入验证失败",
            zap.String("email", req.Email),
            zap.Error(err),
        )
        return nil, err
    }

    // 第二层：业务规则检查
    if err := s.checkBusinessRules(ctx, req); err != nil {
        s.logger.Warn("用户注册业务规则检查失败",
            zap.String("email", req.Email),
            zap.Error(err),
        )
        return nil, err
    }

    // 第三层：数据处理
    user, err := s.createUserRecord(ctx, req)
    if err != nil {
        s.logger.Error("创建用户记录失败",
            zap.String("email", req.Email),
            zap.Error(err),
        )
        return nil, err
    }

    s.logger.Info("用户注册成功",
        zap.Uint("user_id", user.ID),
        zap.String("email", user.Email),
    )

    return user, nil
}

func (s *UserService) validateRegisterRequest(req *RegisterRequest) error {
    var validationErrors []string

    // 验证邮箱
    if req.Email == "" {
        validationErrors = append(validationErrors, "邮箱不能为空")
    } else if !s.emailRegex.MatchString(req.Email) {
        validationErrors = append(validationErrors, "邮箱格式无效")
    }

    // 验证密码
    if req.Password == "" {
        validationErrors = append(validationErrors, "密码不能为空")
    } else if len(req.Password) < 8 {
        validationErrors = append(validationErrors, "密码长度至少8位")
    }

    // 验证姓名
    if req.Name == "" {
        validationErrors = append(validationErrors, "姓名不能为空")
    }

    if len(validationErrors) > 0 {
        return errors.InvalidInput(fmt.Sprintf("输入验证失败: %v", validationErrors)).
            WithDetail("validation_errors", validationErrors).
            WithDetail("email", req.Email)
    }

    return nil
}

func (s *UserService) checkBusinessRules(ctx context.Context, req *RegisterRequest) error {
    // 检查邮箱是否已存在
    existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
    if err != nil {
        // 数据库查询错误
        return errors.DatabaseError(err).
            WithDetail("operation", "check_email_exists").
            WithDetail("email", req.Email)
    }

    if existingUser != nil {
        return errors.New(errors.CodeEmailExists, "邮箱已被注册").
            WithDetail("email", req.Email).
            WithDetail("existing_user_id", existingUser.ID)
    }

    return nil
}

func (s *UserService) createUserRecord(ctx context.Context, req *RegisterRequest) (*model.User, error) {
    // 密码加密
    hashedPassword, err := s.hashPassword(req.Password)
    if err != nil {
        return nil, errors.Wrap(err, errors.CodeInternalError, "密码加密失败")
    }

    // 创建用户对象
    user := &model.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashedPassword,
        Status:   model.UserStatusActive,
    }

    // 保存到数据库
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, errors.DatabaseError(err).
            WithDetail("operation", "create_user").
            WithDetail("email", req.Email)
    }

    return user, nil
}

// 用户登录
func (s *UserService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
    // 输入验证
    if email == "" || password == "" {
        return nil, "", errors.InvalidInput("邮箱和密码不能为空")
    }

    // 查找用户
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        s.logger.Error("查询用户失败", zap.String("email", email), zap.Error(err))
        return nil, "", errors.DatabaseError(err).WithDetail("email", email)
    }

    if user == nil {
        s.logger.Warn("用户不存在", zap.String("email", email))
        return nil, "", errors.UserNotFound(email)
    }

    // 验证密码
    if !s.verifyPassword(password, user.Password) {
        s.logger.Warn("密码验证失败", zap.String("email", email))
        return nil, "", errors.New(errors.CodeUnauthorized, "邮箱或密码错误").
            WithDetail("email", email)
    }

    // 检查用户状态
    if user.Status != model.UserStatusActive {
        return nil, "", errors.New(errors.CodeUnauthorized, "用户账户已被禁用").
            WithDetail("user_id", user.ID).
            WithDetail("status", user.Status)
    }

    // 生成JWT令牌
    token, err := s.generateJWT(user)
    if err != nil {
        s.logger.Error("生成JWT失败", zap.Uint("user_id", user.ID), zap.Error(err))
        return nil, "", errors.Wrap(err, errors.CodeInternalError, "生成访问令牌失败")
    }

    s.logger.Info("用户登录成功", zap.Uint("user_id", user.ID), zap.String("email", email))
    return user, token, nil
}

// 获取用户信息
func (s *UserService) GetUser(ctx context.Context, userID uint) (*model.User, error) {
    if userID == 0 {
        return nil, errors.InvalidInput("用户ID不能为0")
    }

    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        s.logger.Error("查询用户失败", zap.Uint("user_id", userID), zap.Error(err))
        return nil, errors.DatabaseError(err).WithDetail("user_id", userID)
    }

    if user == nil {
        return nil, errors.UserNotFound(userID)
    }

    return user, nil
}

// 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID uint, req *UpdateUserRequest) (*model.User, error) {
    // 获取现有用户
    user, err := s.GetUser(ctx, userID)
    if err != nil {
        return nil, err // 错误已经被包装过了
    }

    // 验证更新请求
    if err := s.validateUpdateRequest(req); err != nil {
        return nil, err
    }

    // 检查邮箱唯一性（如果邮箱有变化）
    if req.Email != "" && req.Email != user.Email {
        if err := s.checkEmailUniqueness(ctx, req.Email, userID); err != nil {
            return nil, err
        }
        user.Email = req.Email
    }

    // 更新字段
    if req.Name != "" {
        user.Name = req.Name
    }

    // 保存更新
    if err := s.userRepo.Update(ctx, user); err != nil {
        s.logger.Error("更新用户失败", zap.Uint("user_id", userID), zap.Error(err))
        return nil, errors.DatabaseError(err).
            WithDetail("user_id", userID).
            WithDetail("operation", "update_user")
    }

    s.logger.Info("用户信息更新成功", zap.Uint("user_id", userID))
    return user, nil
}

func (s *UserService) validateUpdateRequest(req *UpdateUserRequest) error {
    if req.Email != "" && !s.emailRegex.MatchString(req.Email) {
        return errors.InvalidInput("邮箱格式无效").WithDetail("email", req.Email)
    }

    if req.Name != "" && len(req.Name) < 2 {
        return errors.InvalidInput("姓名长度至少2位").WithDetail("name", req.Name)
    }

    return nil
}

func (s *UserService) checkEmailUniqueness(ctx context.Context, email string, excludeUserID uint) error {
    existingUser, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return errors.DatabaseError(err).WithDetail("email", email)
    }

    if existingUser != nil && existingUser.ID != excludeUserID {
        return errors.New(errors.CodeEmailExists, "邮箱已被其他用户使用").
            WithDetail("email", email).
            WithDetail("existing_user_id", existingUser.ID)
    }

    return nil
}

// 辅助方法
func (s *UserService) hashPassword(password string) (string, error) {
    // 实际实现中使用bcrypt等安全的哈希算法
    return fmt.Sprintf("hashed_%s", password), nil
}

func (s *UserService) verifyPassword(password, hashedPassword string) bool {
    // 实际实现中使用bcrypt.CompareHashAndPassword
    return fmt.Sprintf("hashed_%s", password) == hashedPassword
}

func (s *UserService) generateJWT(user *model.User) (string, error) {
    // 实际实现中使用JWT库生成令牌
    return fmt.Sprintf("jwt_token_for_user_%d", user.ID), nil
}

// 请求结构体
type RegisterRequest struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
    Name  string `json:"name,omitempty"`
    Email string `json:"email,omitempty"`
}
```

#### 3. HTTP处理层错误处理

```go
// 来自 mall-go/internal/handler/user_handler.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/yourname/mall-go/pkg/errors"
    "github.com/yourname/mall-go/internal/service"
    "go.uber.org/zap"
)

type UserHandler struct {
    userService *service.UserService
    logger      *zap.Logger
}

func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{
        userService: userService,
        logger:      logger,
    }
}

// 用户注册
func (h *UserHandler) Register(c *gin.Context) {
    var req service.RegisterRequest

    // 绑定请求参数
    if err := c.ShouldBindJSON(&req); err != nil {
        h.handleError(c, errors.InvalidInput("请求参数格式错误").
            WithDetail("bind_error", err.Error()))
        return
    }

    // 调用服务层
    user, err := h.userService.Register(c.Request.Context(), &req)
    if err != nil {
        h.handleError(c, err)
        return
    }

    // 返回成功响应
    c.JSON(http.StatusCreated, gin.H{
        "code":    "0000",
        "message": "用户注册成功",
        "data": gin.H{
            "user_id": user.ID,
            "name":    user.Name,
            "email":   user.Email,
        },
    })
}

// 用户登录
func (h *UserHandler) Login(c *gin.Context) {
    var req struct {
        Email    string `json:"email" binding:"required"`
        Password string `json:"password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        h.handleError(c, errors.InvalidInput("请求参数格式错误").
            WithDetail("bind_error", err.Error()))
        return
    }

    user, token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    "0000",
        "message": "登录成功",
        "data": gin.H{
            "user": gin.H{
                "id":    user.ID,
                "name":  user.Name,
                "email": user.Email,
            },
            "token": token,
        },
    })
}

// 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := strconv.ParseUint(userIDStr, 10, 32)
    if err != nil {
        h.handleError(c, errors.InvalidInput("用户ID格式无效").
            WithDetail("user_id", userIDStr))
        return
    }

    user, err := h.userService.GetUser(c.Request.Context(), uint(userID))
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    "0000",
        "message": "获取用户信息成功",
        "data": gin.H{
            "id":     user.ID,
            "name":   user.Name,
            "email":  user.Email,
            "status": user.Status,
        },
    })
}

// 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
    userIDStr := c.Param("id")
    userID, err := strconv.ParseUint(userIDStr, 10, 32)
    if err != nil {
        h.handleError(c, errors.InvalidInput("用户ID格式无效").
            WithDetail("user_id", userIDStr))
        return
    }

    var req service.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.handleError(c, errors.InvalidInput("请求参数格式错误").
            WithDetail("bind_error", err.Error()))
        return
    }

    user, err := h.userService.UpdateUser(c.Request.Context(), uint(userID), &req)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    "0000",
        "message": "用户信息更新成功",
        "data": gin.H{
            "id":    user.ID,
            "name":  user.Name,
            "email": user.Email,
        },
    })
}

// 统一错误处理
func (h *UserHandler) handleError(c *gin.Context, err error) {
    // 记录错误日志
    h.logger.Error("处理请求时发生错误",
        zap.String("path", c.Request.URL.Path),
        zap.String("method", c.Request.Method),
        zap.Error(err),
    )

    // 检查是否是应用错误
    var appErr errors.AppError
    if errors.As(err, &appErr) {
        // 构建错误响应
        response := gin.H{
            "code":      appErr.Code(),
            "message":   appErr.Error(),
            "timestamp": appErr.Details()["timestamp"],
        }

        // 添加详细信息（开发环境）
        if gin.Mode() == gin.DebugMode {
            response["details"] = appErr.Details()
        }

        c.JSON(appErr.HTTPStatus(), response)
        return
    }

    // 未知错误，返回通用错误响应
    c.JSON(http.StatusInternalServerError, gin.H{
        "code":    "5000",
        "message": "服务器内部错误",
    })
}
```

#### 4. 中间件错误处理

```go
// 来自 mall-go/internal/middleware/error.go
package middleware

import (
    "fmt"
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// 全局错误恢复中间件
func ErrorRecovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // 记录panic信息
                stack := debug.Stack()
                logger.Error("HTTP请求发生panic",
                    zap.Any("panic", r),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("method", c.Request.Method),
                    zap.String("user_agent", c.Request.UserAgent()),
                    zap.String("stack", string(stack)),
                )

                // 返回500错误
                c.JSON(http.StatusInternalServerError, gin.H{
                    "code":    "5000",
                    "message": "服务器内部错误",
                })

                c.Abort()
            }
        }()

        c.Next()
    }
}

// 请求日志中间件
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        logger.Info("HTTP请求",
            zap.String("method", param.Method),
            zap.String("path", param.Path),
            zap.Int("status", param.StatusCode),
            zap.Duration("latency", param.Latency),
            zap.String("client_ip", param.ClientIP),
            zap.String("user_agent", param.Request.UserAgent()),
        )
        return ""
    })
}

// 错误统计中间件
func ErrorMetrics() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 统计错误响应
        if c.Writer.Status() >= 400 {
            // 这里可以集成Prometheus等监控系统
            fmt.Printf("错误响应: %s %s -> %d\n",
                c.Request.Method, c.Request.URL.Path, c.Writer.Status())
        }
    }
}
```

#### 5. 与Java Spring Boot的对比

```java
// Java Spring Boot 错误处理对比
@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(ValidationException.class)
    public ResponseEntity<ErrorResponse> handleValidation(ValidationException e) {
        return ResponseEntity.badRequest()
            .body(new ErrorResponse("VALIDATION_ERROR", e.getMessage()));
    }

    @ExceptionHandler(BusinessException.class)
    public ResponseEntity<ErrorResponse> handleBusiness(BusinessException e) {
        return ResponseEntity.unprocessableEntity()
            .body(new ErrorResponse(e.getCode(), e.getMessage()));
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGeneral(Exception e) {
        log.error("Unexpected error", e);
        return ResponseEntity.internalServerError()
            .body(new ErrorResponse("INTERNAL_ERROR", "服务器内部错误"));
    }
}

@Service
public class UserService {

    public User register(RegisterRequest request) throws BusinessException {
        // 输入验证
        if (!isValidEmail(request.getEmail())) {
            throw new ValidationException("邮箱格式无效");
        }

        // 业务检查
        if (userRepository.existsByEmail(request.getEmail())) {
            throw new BusinessException("EMAIL_EXISTS", "邮箱已存在");
        }

        try {
            return userRepository.save(new User(request));
        } catch (DataAccessException e) {
            throw new BusinessException("DATABASE_ERROR", "数据库操作失败", e);
        }
    }
}

/*
Java vs Go 错误处理对比：

1. 异常传播：
   - Java: 异常自动向上传播，可能被忽略
   - Go: 错误必须显式检查和处理

2. 性能影响：
   - Java: 异常创建和栈展开有性能开销
   - Go: 错误处理几乎无性能开销

3. 代码可读性：
   - Java: 正常流程和异常处理分离，代码简洁
   - Go: 错误检查代码较多，但错误处理路径清晰

4. 类型安全：
   - Java: 编译时检查受检异常
   - Go: 错误类型在运行时确定

5. 调试信息：
   - Java: 异常包含完整调用栈
   - Go: 需要手动添加上下文信息
*/
```

#### 6. 与Python Flask的对比

```python
# Python Flask 错误处理对比
from flask import Flask, jsonify, request
import logging

app = Flask(__name__)

class AppError(Exception):
    def __init__(self, code, message, status_code=500):
        self.code = code
        self.message = message
        self.status_code = status_code

class ValidationError(AppError):
    def __init__(self, message):
        super().__init__("VALIDATION_ERROR", message, 400)

class BusinessError(AppError):
    def __init__(self, code, message):
        super().__init__(code, message, 422)

@app.errorhandler(AppError)
def handle_app_error(error):
    return jsonify({
        'code': error.code,
        'message': error.message
    }), error.status_code

@app.errorhandler(Exception)
def handle_general_error(error):
    logging.error(f"Unexpected error: {error}", exc_info=True)
    return jsonify({
        'code': 'INTERNAL_ERROR',
        'message': '服务器内部错误'
    }), 500

class UserService:
    def register(self, request_data):
        # 输入验证
        if not self.is_valid_email(request_data.get('email')):
            raise ValidationError("邮箱格式无效")

        # 业务检查
        if self.user_exists(request_data['email']):
            raise BusinessError("EMAIL_EXISTS", "邮箱已存在")

        try:
            return self.create_user(request_data)
        except DatabaseError as e:
            raise BusinessError("DATABASE_ERROR", "数据库操作失败") from e

@app.route('/users', methods=['POST'])
def register_user():
    try:
        user_service = UserService()
        user = user_service.register(request.json)
        return jsonify({'user': user}), 201
    except AppError:
        raise  # 让错误处理器处理
    except Exception as e:
        logging.error(f"Registration failed: {e}")
        raise AppError("INTERNAL_ERROR", "注册失败", 500)

"""
Python vs Go 错误处理对比：

1. 异常机制：
   - Python: 使用异常处理，支持异常链
   - Go: 使用返回值，支持错误包装

2. 错误信息：
   - Python: 异常包含traceback信息
   - Go: 需要手动添加上下文

3. 代码风格：
   - Python: try/except块处理异常
   - Go: if err != nil 显式检查

4. 性能：
   - Python: 异常处理有一定开销
   - Go: 错误处理开销极小

5. 调试：
   - Python: 异常自动包含调用栈
   - Go: 需要使用第三方库或手动添加
"""
```

---

## 📚 章节总结

### 🎯 核心知识点回顾

通过本章的学习，我们深入掌握了Go语言错误处理的精髓：

#### 1. **Go错误处理哲学** 🧠
- **"Errors are values, not exceptions"** - 错误是值，不是异常
- **显式优于隐式** - 必须显式检查和处理错误
- **简单而强大** - 通过简单的error接口实现强大的错误处理

#### 2. **错误处理核心技术** ⚙️
- **error接口设计** - 只有一个Error()方法的简洁接口
- **错误创建模式** - errors.New()、fmt.Errorf()、自定义错误类型
- **错误包装和解包** - Go 1.13的%w动词和errors.Is/As函数
- **错误传播链** - 通过包装保持错误上下文

#### 3. **企业级错误处理实践** 🏢
- **分层错误处理** - 不同层次的错误处理策略
- **错误分类和编码** - 系统化的错误分类体系
- **结构化错误信息** - 包含丰富上下文的错误对象
- **统一错误响应** - 标准化的API错误响应格式

#### 4. **性能和监控** 📊
- **高性能错误处理** - 相比异常处理的性能优势
- **错误指标监控** - 错误统计和趋势分析
- **分布式错误追踪** - 跨服务的错误链路追踪
- **自动告警机制** - 基于规则的错误告警

### 🆚 与其他语言的关键差异

| 特性 | Go | Java | Python |
|------|----|----- |--------|
| **错误表示** | 返回值 | 异常对象 | 异常对象 |
| **错误传播** | 显式传播 | 自动传播 | 自动传播 |
| **性能开销** | 极低 | 中等 | 中等 |
| **类型安全** | 运行时 | 编译时 | 运行时 |
| **调用栈** | 需手动添加 | 自动包含 | 自动包含 |
| **代码风格** | if err != nil | try/catch | try/except |

### 💡 最佳实践总结

#### ✅ **应该做的**
1. **始终检查错误** - 永远不要忽略error返回值
2. **提供上下文** - 使用fmt.Errorf添加错误上下文
3. **分层处理** - 不同层次采用不同的错误处理策略
4. **预定义错误** - 使用var定义可比较的错误值
5. **结构化日志** - 记录结构化的错误信息

#### ❌ **不应该做的**
1. **忽略错误** - 使用_丢弃错误返回值
2. **滥用panic** - 在业务逻辑中使用panic
3. **丢失上下文** - 直接返回底层错误而不添加上下文
4. **字符串比较** - 通过错误消息字符串判断错误类型
5. **过度包装** - 不必要的多层错误包装

### 🚀 进阶学习方向

#### 1. **并发错误处理** 🔄
- 学习在goroutine中的错误处理模式
- 掌握errgroup包的使用
- 理解context在错误处理中的作用

#### 2. **微服务错误处理** 🌐
- 分布式系统中的错误传播
- 服务间错误码标准化
- 熔断器和重试机制

#### 3. **监控和可观测性** 📈
- 集成Prometheus进行错误监控
- 使用Jaeger进行分布式追踪
- 构建错误仪表板和告警系统

#### 4. **测试驱动的错误处理** 🧪
- 错误场景的单元测试
- 错误处理的集成测试
- 混沌工程和故障注入

### 📖 推荐阅读资源

#### **官方文档**
- [Go Blog: Error handling and Go](https://blog.golang.org/error-handling-and-go)
- [Go Blog: Working with Errors in Go 1.13](https://blog.golang.org/go1.13-errors)

#### **开源项目**
- [pkg/errors](https://github.com/pkg/errors) - 错误处理增强库
- [go-multierror](https://github.com/hashicorp/go-multierror) - 多重错误处理
- [emperror](https://github.com/emperror/errors) - 企业级错误处理

#### **实战项目**
- 研究知名Go项目的错误处理实现
- 参与开源项目，学习最佳实践
- 构建自己的错误处理框架

### 🎓 下一步学习建议

恭喜你完成了Go语言错误处理的深度学习！🎉

**接下来建议学习：**

1. **进阶篇第三章：并发编程基础**
   - Goroutine和Channel的使用
   - 并发安全和同步原语
   - 并发模式和最佳实践

2. **进阶篇第四章：接口设计模式**
   - 接口的高级用法
   - 设计模式在Go中的实现
   - 依赖注入和控制反转

3. **实战篇：构建完整的Web应用**
   - 将错误处理应用到实际项目
   - 集成数据库、缓存、消息队列
   - 部署和监控

记住，**优秀的错误处理是优秀软件的基石**！继续保持学习的热情，在实践中不断完善你的错误处理技能。

---

*"程序员的三大美德：懒惰、急躁和傲慢。但在错误处理上，我们要勤奋、耐心和谦逊。"* 😄

**Happy Coding! 🚀**
