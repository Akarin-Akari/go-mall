# Go语言变量声明与数据类型详解

> 🎯 **学习目标**: 掌握Go语言的变量声明方式和数据类型系统，理解与Java/Python的核心差异
> 
> ⏱️ **预计学习时间**: 2-3小时
> 
> 📚 **前置知识**: Java或Python基础语法

## 📋 本章内容概览

- [变量声明的多种方式](#变量声明的多种方式)
- [Go的数据类型系统](#go的数据类型系统)
- [零值概念与初始化](#零值概念与初始化)
- [类型转换与类型推断](#类型转换与类型推断)
- [常量定义与枚举](#常量定义与枚举)
- [指针基础概念](#指针基础概念)
- [实战案例分析](#实战案例分析)
- [面试常考点](#面试常考点)

---

## 🚀 变量声明的多种方式

### Java vs Python vs Go 对比

让我们先看看三种语言的变量声明差异：

**Java (你熟悉的方式):**
```java
// Java - 类型在前，变量名在后
String username = "admin";
int age = 25;
List<String> users = new ArrayList<>();
boolean isActive = true;

// 类型推断 (Java 10+)
var name = "张三";  // 编译器推断为String
```

**Python (你熟悉的方式):**
```python
# Python - 动态类型，无需声明类型
username = "admin"
age = 25
users = []
is_active = True

# 类型注解 (Python 3.5+)
name: str = "张三"
```

**Go (新的方式):**
```go
// Go - 多种声明方式
var username string = "admin"    // 完整声明
var age int = 25                 // 完整声明
var users []string               // 声明但不初始化
var isActive bool = true         // 完整声明

// 类型推断
var name = "张三"                // 编译器推断为string
username := "admin"              // 短变量声明 (最常用)
```

### Go变量声明的四种方式

#### 1. 完整声明 (var 关键字)

```go
// 语法: var 变量名 类型 = 值
var username string = "admin"
var age int = 25
var salary float64 = 8500.50
var isActive bool = true
```

**适用场景:**
- 包级别变量声明
- 需要明确指定类型的场合
- 初始值为零值的情况

#### 2. 类型推断声明

```go
// 语法: var 变量名 = 值 (编译器自动推断类型)
var username = "admin"     // 推断为 string
var age = 25              // 推断为 int
var salary = 8500.50      // 推断为 float64
var isActive = true       // 推断为 bool
```

#### 3. 短变量声明 (最常用)

```go
// 语法: 变量名 := 值 (只能在函数内使用)
username := "admin"
age := 25
salary := 8500.50
isActive := true
```

**⚠️ 重要限制:**
- 只能在函数内部使用
- 不能用于包级别变量
- 左侧必须至少有一个新变量

#### 4. 批量声明

```go
// 方式1: 分组声明
var (
    username string = "admin"
    age      int    = 25
    salary   float64 = 8500.50
    isActive bool   = true
)

// 方式2: 多变量同时声明
var username, email, phone string
var x, y, z int = 1, 2, 3
a, b, c := 10, 20, 30
```

### 实际项目中的应用

让我们看看mall-go项目中的真实例子：

```go
// 来自 mall-go/internal/model/user.go
type User struct {
    ID       uint   `gorm:"primarykey" json:"id"`
    Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
    Password string `gorm:"not null;size:255" json:"-"`
    
    // 时间戳字段
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 来自 mall-go/pkg/database/database.go
func Init() *gorm.DB {
    var err error                    // 声明错误变量
    cfg := config.GlobalConfig      // 短变量声明
    
    // 根据驱动类型连接数据库
    if cfg.Database.Driver == "sqlite" {
        DB, err = gorm.Open(sqlite.Open(cfg.Database.DBName), gormConfig)
    } else {
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            cfg.Database.Username,
            cfg.Database.Password,
            cfg.Database.Host,
            cfg.Database.Port,
            cfg.Database.DBName,
        )
        DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
    }
    
    if err != nil {
        log.Fatalf("数据库连接失败: %v", err)
    }
    
    return DB
}
```

---

## 🔢 Go的数据类型系统

### 基本数据类型对比

| 类型分类 | Java | Python | Go |
|---------|------|--------|-----|
| 整数 | `int`, `long` | `int` | `int`, `int8`, `int16`, `int32`, `int64` |
| 无符号整数 | 无 | 无 | `uint`, `uint8`, `uint16`, `uint32`, `uint64` |
| 浮点数 | `float`, `double` | `float` | `float32`, `float64` |
| 布尔值 | `boolean` | `bool` | `bool` |
| 字符串 | `String` | `str` | `string` |
| 字符 | `char` | 无 | `rune` (int32别名) |
| 字节 | `byte` | `bytes` | `byte` (uint8别名) |

### 整数类型详解

```go
// 有符号整数
var a int8 = 127        // -128 到 127
var b int16 = 32767     // -32768 到 32767  
var c int32 = 2147483647 // -2^31 到 2^31-1
var d int64 = 9223372036854775807 // -2^63 到 2^63-1
var e int = 100         // 平台相关 (32位或64位)

// 无符号整数
var ua uint8 = 255      // 0 到 255
var ub uint16 = 65535   // 0 到 65535
var uc uint32 = 4294967295 // 0 到 2^32-1
var ud uint64 = 18446744073709551615 // 0 到 2^64-1
var ue uint = 100       // 平台相关

// 特殊类型
var f byte = 255        // uint8 的别名
var g rune = '中'       // int32 的别名，用于Unicode字符
```

**与Java对比:**
```java
// Java
byte b = 127;           // -128 到 127
short s = 32767;        // -32768 到 32767
int i = 2147483647;     // -2^31 到 2^31-1
long l = 9223372036854775807L; // -2^63 到 2^63-1

// Java没有无符号整数类型！
```

### 浮点数类型

```go
// Go的浮点数
var price float32 = 99.99    // 单精度，7位有效数字
var salary float64 = 8500.50 // 双精度，15位有效数字 (推荐)

// 科学计数法
var bigNumber = 1.23e9       // 1.23 * 10^9
var smallNumber = 1.23e-9    // 1.23 * 10^-9
```

**💡 最佳实践:**
- 优先使用 `float64`，精度更高
- 金融计算建议使用 `decimal.Decimal` (第三方库)

### 字符串类型

```go
// 字符串声明
var name string = "张三"
var message = "Hello, World!"
var empty string            // 空字符串 ""

// 多行字符串 (反引号)
var sql = `
    SELECT id, username, email 
    FROM users 
    WHERE status = 'active'
    ORDER BY created_at DESC
`

// 字符串拼接
var fullName = "姓名: " + name
var greeting = fmt.Sprintf("Hello, %s!", name)
```

**与Java/Python对比:**
```java
// Java
String name = "张三";
String message = "Hello, World!";
String sql = """
    SELECT id, username, email 
    FROM users 
    WHERE status = 'active'
    ORDER BY created_at DESC
    """;
```

```python
# Python
name = "张三"
message = "Hello, World!"
sql = """
    SELECT id, username, email 
    FROM users 
    WHERE status = 'active'
    ORDER BY created_at DESC
"""
```

### 布尔类型

```go
// 布尔值
var isActive bool = true
var isDeleted bool = false
var isEmpty bool            // 默认为 false

// 布尔运算
var result = isActive && !isDeleted
var canAccess = isActive || isAdmin
```

**⚠️ 注意差异:**
- Go: `true`/`false` (小写)
- Java: `true`/`false` (小写)  
- Python: `True`/`False` (首字母大写)

---

## 🔄 零值概念与初始化

### 什么是零值？

Go语言中，所有变量都有一个**零值**(zero value)，这是Go的一个重要特性：

```go
var i int        // 0
var f float64    // 0.0
var b bool       // false
var s string     // ""
var p *int       // nil
var slice []int  // nil
var m map[string]int // nil
var ch chan int  // nil
var fn func()    // nil
```

**与Java/Python对比:**
```java
// Java - 需要显式初始化，否则编译错误
int i;           // 编译错误！必须初始化
int i = 0;       // 正确

String s;        // 编译错误！必须初始化  
String s = null; // 正确，但可能导致NullPointerException
```

```python
# Python - 未初始化变量会导致运行时错误
print(i)  # NameError: name 'i' is not defined
```

### 零值的优势

```go
// Go - 零值是安全的
var users []User
if users == nil {
    users = make([]User, 0)  // 安全检查
}

var counter int
counter++  // 安全，从0开始计数

var buffer strings.Builder
buffer.WriteString("Hello")  // 零值可以直接使用
```

### 实际项目中的零值应用

```go
// 来自 mall-go/internal/model/user.go
type User struct {
    ID       uint   `gorm:"primarykey" json:"id"`
    Username string `gorm:"not null" json:"username"`
    Email    string `gorm:"not null" json:"email"`
    
    // 这些字段如果不设置，会使用零值
    LoginCount     int `gorm:"default:0" json:"login_count"`     // 零值: 0
    PostCount      int `gorm:"default:0" json:"post_count"`      // 零值: 0
    FollowerCount  int `gorm:"default:0" json:"follower_count"`  // 零值: 0
    FollowingCount int `gorm:"default:0" json:"following_count"` // 零值: 0
}

// 创建用户时，未设置的字段自动使用零值
func CreateUser(username, email string) *User {
    return &User{
        Username: username,
        Email:    email,
        // LoginCount, PostCount 等会自动设为 0
    }
}
```

---

## 🔄 类型转换与类型推断

### 显式类型转换

Go要求显式类型转换，不允许隐式转换：

```go
// Go - 必须显式转换
var i int = 42
var f float64 = float64(i)  // 必须显式转换
var u uint = uint(i)        // 必须显式转换

// 字符串转换
var s string = string(i)    // 错误！不能直接转换
var s string = fmt.Sprintf("%d", i)  // 正确方式
var s string = strconv.Itoa(i)       // 更高效的方式
```

**与Java对比:**
```java
// Java - 有隐式转换
int i = 42;
double d = i;        // 隐式转换，允许
float f = i;         // 隐式转换，允许

// 但缩小转换需要显式
double d = 42.5;
int i = (int) d;     // 显式转换
```

### 类型推断规则

```go
// 整数字面量推断
var a = 42          // int
var b = int64(42)   // int64

// 浮点数字面量推断  
var c = 3.14        // float64
var d = float32(3.14) // float32

// 字符串字面量推断
var e = "hello"     // string

// 布尔字面量推断
var f = true        // bool

// 复杂类型推断
var users = []User{}           // []User
var scores = map[string]int{}  // map[string]int
```

### 常见类型转换函数

```go
import (
    "strconv"
    "fmt"
)

// 字符串与数字互转
var str = "123"
var num, err = strconv.Atoi(str)        // string -> int
var str2 = strconv.Itoa(num)            // int -> string
var float, err2 = strconv.ParseFloat(str, 64) // string -> float64

// 使用fmt包
var str3 = fmt.Sprintf("%d", num)       // 任意类型 -> string
var str4 = fmt.Sprintf("%.2f", 3.14159) // "3.14"
```

---

## 📌 常量定义与枚举

### 常量声明

```go
// 单个常量
const PI = 3.14159
const AppName = "Mall-Go"
const MaxRetries = 3

// 批量常量声明
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
    StatusDeleted  = "deleted"
)

// 类型化常量
const (
    MaxUsers    int     = 1000
    DefaultRate float64 = 0.05
    AppVersion  string  = "1.0.0"
)
```

### iota 枚举生成器

```go
// 基础用法
const (
    Sunday = iota    // 0
    Monday           // 1
    Tuesday          // 2
    Wednesday        // 3
    Thursday         // 4
    Friday           // 5
    Saturday         // 6
)

// 跳过某些值
const (
    _ = iota         // 跳过 0
    KB = 1 << (10 * iota) // 1024
    MB               // 1048576
    GB               // 1073741824
)

// 实际项目中的枚举
const (
    OrderStatusPending = iota + 1  // 1
    OrderStatusPaid                // 2
    OrderStatusShipped             // 3
    OrderStatusDelivered           // 4
    OrderStatusCancelled           // 5
)
```

### 实际项目中的常量应用

```go
// 来自 mall-go/pkg/response/response.go
const (
    CodeSuccess      = 200  // 成功
    CodeError        = 500  // 服务器错误
    CodeInvalidParam = 400  // 参数错误
    CodeUnauthorized = 401  // 未授权
    CodeForbidden    = 403  // 禁止访问
    CodeNotFound     = 404  // 资源不存在
    CodeConflict     = 409  // 资源冲突
    CodeTooManyReq   = 429  // 请求过多
)

// 来自 mall-go/internal/model/order.go
const (
    OrderStatusPending   = "pending"    // 待支付
    OrderStatusPaid      = "paid"       // 已支付
    OrderStatusShipped   = "shipped"    // 已发货
    OrderStatusDelivered = "delivered"  // 已送达
    OrderStatusCancelled = "cancelled"  // 已取消
    OrderStatusRefunded  = "refunded"   // 已退款
)
```

---

## 👉 指针基础概念

### 指针 vs 引用对比

**Java (引用):**
```java
// Java - 对象引用
User user1 = new User("张三");
User user2 = user1;  // user2 引用同一个对象
user2.setName("李四");
System.out.println(user1.getName()); // "李四"
```

**Go (指针):**
```go
// Go - 指针
var x int = 42
var p *int = &x    // p 是指向 x 的指针
fmt.Println(*p)    // 42 (解引用)
*p = 100           // 通过指针修改值
fmt.Println(x)     // 100

// 结构体指针
user1 := &User{Name: "张三"}
user2 := user1     // user2 指向同一个对象
user2.Name = "李四"
fmt.Println(user1.Name) // "李四"
```

### 指针的基本操作

```go
// 声明指针
var p *int        // 指向 int 的指针，初始值为 nil

// 获取地址
var x int = 42
p = &x            // p 指向 x 的地址

// 解引用
fmt.Println(*p)   // 42，获取指针指向的值
*p = 100          // 修改指针指向的值

// 检查空指针
if p != nil {
    fmt.Println(*p)
}
```

### 函数参数：值传递 vs 指针传递

```go
// 值传递 - 不会修改原变量
func updateValueByValue(x int) {
    x = 100
}

// 指针传递 - 会修改原变量
func updateValueByPointer(x *int) {
    *x = 100
}

func main() {
    var num = 42
    
    updateValueByValue(num)
    fmt.Println(num)  // 42，未改变
    
    updateValueByPointer(&num)
    fmt.Println(num)  // 100，已改变
}
```

### 实际项目中的指针应用

```go
// 来自 mall-go/internal/model/user.go
func (u *User) SetPassword(password string) error {
    // 使用指针接收者，可以修改结构体
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)  // 修改原结构体
    return nil
}

// 来自 mall-go/pkg/database/database.go
func InitDB() (*gorm.DB, error) {
    // 返回指针，避免复制大对象
    return Init(), nil
}
```

---

## 💼 实战案例分析

让我们通过mall-go项目中的真实代码来理解变量和类型的应用：

### 案例1: 用户模型定义

```go
// mall-go/internal/model/user.go
type User struct {
    // 基本字段 - 使用了不同的数据类型
    ID       uint   `gorm:"primarykey" json:"id"`                    // 无符号整数
    Username string `gorm:"uniqueIndex;not null;size:50" json:"username"` // 字符串
    Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`   // 字符串
    Password string `gorm:"not null;size:255" json:"-"`              // 字符串，JSON忽略
    
    // 个人信息
    RealName    string `gorm:"size:50" json:"real_name"`            // 可选字符串
    Phone       string `gorm:"size:20" json:"phone"`                // 可选字符串
    Avatar      string `gorm:"size:255" json:"avatar"`              // 可选字符串
    Gender      int    `gorm:"default:0" json:"gender"`             // 整数枚举
    Birthday    *time.Time `json:"birthday"`                        // 指针类型，可为nil
    
    // 状态字段
    Status      string `gorm:"size:20;default:'active'" json:"status"` // 字符串枚举
    IsVerified  bool   `gorm:"default:false" json:"is_verified"`    // 布尔值
    
    // 统计信息 - 使用零值初始化
    LoginCount     int `gorm:"default:0" json:"login_count"`        // 整数，默认0
    PostCount      int `gorm:"default:0" json:"post_count"`         // 整数，默认0
    FollowerCount  int `gorm:"default:0" json:"follower_count"`     // 整数，默认0
    FollowingCount int `gorm:"default:0" json:"following_count"`    // 整数，默认0
    
    // 时间戳
    CreatedAt time.Time      `json:"created_at"`                    // 时间类型
    UpdatedAt time.Time      `json:"updated_at"`                    // 时间类型
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                // 软删除
}
```

**关键知识点:**
1. **结构体标签**: `gorm` 和 `json` 标签用于ORM映射和JSON序列化
2. **指针字段**: `Birthday *time.Time` 使用指针，可以表示"未设置"状态
3. **零值应用**: 统计字段使用零值初始化，安全且合理
4. **类型选择**: 不同字段选择合适的数据类型

### 案例2: 数据库连接配置

```go
// mall-go/pkg/database/database.go
func Init() *gorm.DB {
    var err error                    // 错误变量，零值为 nil
    
    // 获取配置 - 短变量声明
    cfg := config.GlobalConfig
    
    // 配置GORM - 结构体字面量
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    }
    
    // 条件分支中的变量声明
    if cfg.Database.Driver == "sqlite" {
        // 在条件分支中声明和赋值
        DB, err = gorm.Open(sqlite.Open(cfg.Database.DBName), gormConfig)
    } else if cfg.Database.Driver == "memory" {
        DB, err = gorm.Open(sqlite.Open(":memory:"), gormConfig)
        log.Println("使用内存数据库模式（仅用于测试）")
    } else {
        // 字符串格式化 - 类型转换
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
            cfg.Database.Username,    // string
            cfg.Database.Password,    // string  
            cfg.Database.Host,        // string
            cfg.Database.Port,        // int -> 格式化为字符串
            cfg.Database.DBName,      // string
        )
        DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
    }
    
    // 错误处理
    if err != nil {
        log.Fatalf("数据库连接失败: %v", err)
    }
    
    return DB
}
```

**关键知识点:**
1. **变量作用域**: `err` 在函数开始声明，在多个分支中使用
2. **短变量声明**: `cfg := config.GlobalConfig` 简洁明了
3. **字符串格式化**: `fmt.Sprintf` 进行类型转换和格式化
4. **指针返回**: 返回 `*gorm.DB` 指针，避免复制大对象

### 案例3: 响应结构定义

```go
// mall-go/pkg/response/response.go
type Response struct {
    Code    int         `json:"code"`    // 状态码
    Message string      `json:"message"` // 消息
    Data    interface{} `json:"data"`    // 数据 - 空接口类型
}

// 分页响应结构
type PageResult struct {
    List     interface{} `json:"list"`      // 数据列表 - 空接口
    Total    int64       `json:"total"`     // 总数 - 64位整数
    Page     int         `json:"page"`      // 当前页 - 32位整数
    PageSize int         `json:"page_size"` // 每页大小 - 32位整数
}

// 状态码常量
const (
    CodeSuccess      = 200  // 成功
    CodeError        = 500  // 服务器错误
    CodeInvalidParam = 400  // 参数错误
    CodeUnauthorized = 401  // 未授权
    CodeForbidden    = 403  // 禁止访问
    CodeNotFound     = 404  // 资源不存在
    CodeConflict     = 409  // 资源冲突
    CodeTooManyReq   = 429  // 请求过多
)
```

**关键知识点:**
1. **空接口**: `interface{}` 可以接受任何类型的值
2. **整数类型选择**: `Total` 使用 `int64`，`Page` 使用 `int`
3. **常量定义**: 使用常量定义状态码，提高代码可读性

---

## 🎯 面试常考点

### 1. 变量声明方式的区别

**面试题**: "Go语言有几种变量声明方式？它们的区别是什么？"

**标准答案**:
```go
// 1. 完整声明 - 可用于包级别和函数内
var username string = "admin"

// 2. 类型推断 - 可用于包级别和函数内  
var age = 25

// 3. 短变量声明 - 只能用于函数内
email := "user@example.com"

// 4. 批量声明 - 可用于包级别和函数内
var (
    host string = "localhost"
    port int    = 8080
)
```

**关键区别**:
- 短变量声明 `:=` 只能在函数内使用
- `var` 声明可以在包级别使用
- 短变量声明左侧必须至少有一个新变量

### 2. 零值概念

**面试题**: "什么是Go的零值？为什么要设计零值？"

**标准答案**:
```go
// Go的零值设计让变量始终有一个安全的初始状态
var i int        // 0
var s string     // ""
var b bool       // false
var p *int       // nil
var slice []int  // nil

// 零值的优势：避免未初始化变量的问题
var counter int
counter++  // 安全，从0开始

// 对比其他语言的问题
// Java: int i; i++; // 编译错误
// C++: int i; i++; // 未定义行为
```

### 3. 类型转换

**面试题**: "Go语言的类型转换有什么特点？"

**标准答案**:
- Go不允许隐式类型转换，必须显式转换
- 即使是兼容类型也需要显式转换
- 字符串和数字转换需要使用 `strconv` 包

```go
var i int = 42
var f float64 = float64(i)  // 必须显式转换
var s string = strconv.Itoa(i)  // 字符串转换
```

### 4. 指针 vs 引用

**面试题**: "Go的指针和Java的引用有什么区别？"

**标准答案**:
- Go有真正的指针，可以进行指针运算（受限）
- Java只有对象引用，没有指针概念
- Go的指针可以为 `nil`，需要检查
- Go支持指针传递和值传递

```go
// Go指针
var x int = 42
var p *int = &x
*p = 100  // 通过指针修改值

// Java引用
User user = new User();  // user是引用，不是指针
```

### 5. 常量和iota

**面试题**: "解释Go语言中的iota关键字"

**标准答案**:
```go
// iota是常量生成器，在const块中自动递增
const (
    Sunday = iota    // 0
    Monday           // 1
    Tuesday          // 2
)

// 可以用于位运算
const (
    Read = 1 << iota  // 1
    Write             // 2  
    Execute           // 4
)
```

### 💡 踩坑提醒

#### 1. 短变量声明的常见错误
```go
// ❌ 错误：在包级别使用短变量声明
package main
username := "admin"  // 编译错误！

// ✅ 正确：在函数内使用
func main() {
    username := "admin"  // 正确
}

// ❌ 错误：重复声明所有变量
func main() {
    username := "admin"
    username := "user"   // 编译错误！左侧没有新变量
}

// ✅ 正确：至少有一个新变量
func main() {
    username := "admin"
    username, email := "user", "user@example.com"  // 正确，email是新变量
}
```

#### 2. 类型转换的陷阱
```go
// ❌ 错误：直接转换可能溢出
var bigNum int64 = 9223372036854775807
var smallNum int32 = int32(bigNum)  // 数据溢出！

// ✅ 正确：检查范围
if bigNum <= math.MaxInt32 {
    smallNum := int32(bigNum)
    fmt.Println(smallNum)
}

// ❌ 错误：字符串转换
var num int = 65
var char string = string(num)  // 得到 "A"，不是 "65"！

// ✅ 正确：数字转字符串
var numStr string = strconv.Itoa(num)  // "65"
```

#### 3. 指针的常见问题
```go
// ❌ 错误：空指针解引用
var p *int
fmt.Println(*p)  // panic: runtime error

// ✅ 正确：检查空指针
if p != nil {
    fmt.Println(*p)
}

// ❌ 错误：返回局部变量地址（在某些情况下）
func getPointer() *int {
    x := 42
    return &x  // 在Go中这是安全的，但在C/C++中是危险的
}
```

#### 4. 零值的误解
```go
// ❌ 误解：认为零值是"未初始化"
var users []User
if users == nil {
    // 这不是错误状态，而是正常的零值状态
    users = make([]User, 0)
}

// ✅ 正确理解：零值是安全的初始状态
var counter int  // 0，可以直接使用
counter++        // 安全，现在是1
```

### 🎯 面试技巧

#### 1. 展示对比思维
当面试官问Go语法时，主动对比Java：
```
"Go的变量声明和Java不同，Go支持类型推断，比如 username := 'admin'，
编译器会自动推断为string类型，这比Java的var关键字更简洁。"
```

#### 2. 强调零值优势
```
"Go的零值设计很巧妙，避免了Java中NullPointerException的问题，
比如int的零值是0，可以直接使用，不需要担心未初始化。"
```

#### 3. 提及性能考虑
```
"Go的指针比Java的引用更直接，可以减少内存分配，
在高性能场景下，指针传递比值传递更高效。"
```

#### 4. 展示实际经验
```
"在我使用Go开发商城项目时，经常用短变量声明，比如 user := &User{}，
这比Java的 User user = new User() 更简洁。"
```

---

## 📝 本章练习题

### 基础练习

1. **变量声明练习**
```go
// 请用四种不同方式声明以下变量：
// - 用户名: "admin"
// - 年龄: 25
// - 薪水: 8500.50
// - 是否激活: true

// 参考答案：
// 方式1：完整声明
var username string = "admin"
var age int = 25
var salary float64 = 8500.50
var isActive bool = true

// 方式2：类型推断
var username2 = "admin"
var age2 = 25
var salary2 = 8500.50
var isActive2 = true

// 方式3：短变量声明（函数内）
func example() {
    username3 := "admin"
    age3 := 25
    salary3 := 8500.50
    isActive3 := true
}

// 方式4：批量声明
var (
    username4 string  = "admin"
    age4      int     = 25
    salary4   float64 = 8500.50
    isActive4 bool    = true
)
```

2. **类型转换练习**
```go
// 完成以下类型转换，处理可能的错误：
func convertString(s string) (int, float64, error) {
    // 参考答案：
    intVal, err := strconv.Atoi(s)
    if err != nil {
        return 0, 0, fmt.Errorf("转换为int失败: %v", err)
    }

    floatVal, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0, 0, fmt.Errorf("转换为float64失败: %v", err)
    }

    return intVal, floatVal, nil
}

// 使用示例：
func main() {
    i, f, err := convertString("123")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("int: %d, float: %.2f\n", i, f)
}
```

3. **指针练习**
```go
// 编写一个函数，交换两个整数的值
func swap(a, b *int) {
    // 参考答案：
    *a, *b = *b, *a
}

// 使用示例：
func main() {
    x, y := 10, 20
    fmt.Printf("交换前: x=%d, y=%d\n", x, y)
    swap(&x, &y)
    fmt.Printf("交换后: x=%d, y=%d\n", x, y)
}
```

### 进阶练习

4. **结构体设计**
```go
// 设计一个商品结构体
type Product struct {
    ID          uint           `json:"id" gorm:"primarykey"`
    Name        string         `json:"name" gorm:"not null;size:200"`
    Price       decimal.Decimal `json:"price" gorm:"type:decimal(10,2);not null"`
    Stock       int            `json:"stock" gorm:"default:0"`
    IsActive    bool           `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time      `json:"created_at"`
    Description *string        `json:"description,omitempty" gorm:"type:text"`
}

// 工厂函数
func NewProduct(name string, price decimal.Decimal, stock int) *Product {
    return &Product{
        Name:      name,
        Price:     price,
        Stock:     stock,
        IsActive:  true,
        CreatedAt: time.Now(),
    }
}
```

5. **常量枚举**
```go
// 使用iota定义订单状态枚举
const (
    OrderStatusPending = iota + 1  // 1 - 待支付
    OrderStatusPaid                // 2 - 已支付
    OrderStatusShipped             // 3 - 已发货
    OrderStatusCompleted           // 4 - 已完成
    OrderStatusCancelled           // 5 - 已取消
)

// 状态名称映射
var OrderStatusNames = map[int]string{
    OrderStatusPending:   "待支付",
    OrderStatusPaid:      "已支付",
    OrderStatusShipped:   "已发货",
    OrderStatusCompleted: "已完成",
    OrderStatusCancelled: "已取消",
}

// 获取状态名称
func GetOrderStatusName(status int) string {
    if name, ok := OrderStatusNames[status]; ok {
        return name
    }
    return "未知状态"
}
```

### 实战练习

6. **配置结构体**
```go
// 参考mall-go项目的数据库配置结构体
type DatabaseConfig struct {
    Driver          string `yaml:"driver" json:"driver"`                    // 数据库驱动
    Host            string `yaml:"host" json:"host"`                        // 主机地址
    Port            int    `yaml:"port" json:"port"`                        // 端口号
    Username        string `yaml:"username" json:"username"`                // 用户名
    Password        string `yaml:"password" json:"password"`                // 密码
    DBName          string `yaml:"dbname" json:"dbname"`                    // 数据库名
    MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`    // 最大空闲连接
    MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`    // 最大打开连接
    ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // 连接最大生命周期
}

// 默认配置
func DefaultDatabaseConfig() *DatabaseConfig {
    return &DatabaseConfig{
        Driver:          "mysql",
        Host:            "localhost",
        Port:            3306,
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: 3600,
    }
}

// 构建DSN连接字符串
func (cfg *DatabaseConfig) BuildDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.Username,
        cfg.Password,
        cfg.Host,
        cfg.Port,
        cfg.DBName,
    )
}
```

### 🏆 挑战练习

7. **类型安全的枚举**
```go
// 创建一个类型安全的用户状态枚举
type UserStatus int

const (
    UserStatusActive UserStatus = iota + 1
    UserStatusInactive
    UserStatusSuspended
    UserStatusDeleted
)

// 实现String方法
func (s UserStatus) String() string {
    switch s {
    case UserStatusActive:
        return "active"
    case UserStatusInactive:
        return "inactive"
    case UserStatusSuspended:
        return "suspended"
    case UserStatusDeleted:
        return "deleted"
    default:
        return "unknown"
    }
}

// 从字符串解析
func ParseUserStatus(s string) (UserStatus, error) {
    switch s {
    case "active":
        return UserStatusActive, nil
    case "inactive":
        return UserStatusInactive, nil
    case "suspended":
        return UserStatusSuspended, nil
    case "deleted":
        return UserStatusDeleted, nil
    default:
        return 0, fmt.Errorf("invalid user status: %s", s)
    }
}
```

8. **泛型变量容器（Go 1.18+）**
```go
// 创建一个泛型的可选值容器
type Optional[T any] struct {
    value *T
}

// 创建有值的Optional
func Some[T any](value T) Optional[T] {
    return Optional[T]{value: &value}
}

// 创建空的Optional
func None[T any]() Optional[T] {
    return Optional[T]{value: nil}
}

// 检查是否有值
func (o Optional[T]) IsPresent() bool {
    return o.value != nil
}

// 获取值
func (o Optional[T]) Get() (T, bool) {
    if o.value != nil {
        return *o.value, true
    }
    var zero T
    return zero, false
}

// 使用示例
func main() {
    // 有值的情况
    name := Some("张三")
    if name.IsPresent() {
        if val, ok := name.Get(); ok {
            fmt.Println("姓名:", val)
        }
    }

    // 空值的情况
    age := None[int]()
    if !age.IsPresent() {
        fmt.Println("年龄未设置")
    }
}
```

---

## 🎉 本章总结

通过本章学习，你应该掌握了：

### ✅ 核心概念
- [x] Go语言的四种变量声明方式
- [x] Go的数据类型系统和零值概念  
- [x] 类型转换和类型推断规则
- [x] 常量定义和iota枚举生成器
- [x] 指针的基本概念和应用

### ✅ 与Java/Python的差异
- [x] 变量声明语法的差异
- [x] 类型系统的严格性差异
- [x] 零值 vs 未初始化变量
- [x] 指针 vs 引用的概念差异

### ✅ 实际应用
- [x] 在mall-go项目中的实际应用
- [x] 结构体字段的类型选择
- [x] 配置和响应结构的设计
- [x] 错误处理中的变量使用

### 🚀 下一步学习

恭喜完成第一章！接下来我们将学习：
- **[控制结构与流程控制](./02-control-structures.md)** - if/for/switch语句
- **[函数定义与方法](./03-functions-and-methods.md)** - 函数声明、参数传递、方法定义

---

> 💡 **学习提示**: 
> 1. 多动手练习，每个概念都要写代码验证
> 2. 对比Java/Python加深理解差异
> 3. 参考mall-go项目中的实际应用
> 4. 重点理解零值和指针概念，这是Go的特色

**继续加油！Go语言的简洁之美正在向你展现！** 🎯
