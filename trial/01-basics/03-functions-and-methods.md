# Go语言函数定义与方法详解

> 🎯 **学习目标**: 掌握Go语言的函数和方法设计，理解与传统面向对象语言的差异和优势
> 
> ⏱️ **预计学习时间**: 3-4小时
> 
> 📚 **前置知识**: 已完成变量类型和控制结构学习

## 📋 本章内容概览

- [Go函数的定义语法](#go函数的定义语法)
- [多返回值机制](#多返回值机制)
- [命名返回值和裸返回](#命名返回值和裸返回)
- [函数作为一等公民](#函数作为一等公民)
- [方法定义和接收者](#方法定义和接收者)
- [值接收者vs指针接收者](#值接收者vs指针接收者)
- [方法集和接口实现](#方法集和接口实现)
- [匿名函数和闭包](#匿名函数和闭包)
- [函数式编程实践](#函数式编程实践)
- [实战案例分析](#实战案例分析)

---

## 🔧 Go函数的定义语法

### Java vs Python vs Go 语法对比

**Java (你熟悉的方式):**
```java
// Java - 复杂的访问修饰符和类型声明
public class MathUtils {
    public static int add(int a, int b) {
        return a + b;
    }
    
    private static boolean isEven(int number) {
        return number % 2 == 0;
    }
    
    // 需要异常声明
    public static String readFile(String filename) throws IOException {
        // 实现...
        return content;
    }
}
```

**Python (你熟悉的方式):**
```python
# Python - 简洁但类型信息不明确
def add(a, b):
    return a + b

def is_even(number):
    return number % 2 == 0

# 类型提示（Python 3.5+）
def read_file(filename: str) -> str:
    # 实现...
    return content

# 可能抛出异常，但无需声明
def divide(a: float, b: float) -> float:
    if b == 0:
        raise ValueError("除数不能为零")
    return a / b
```

**Go (新的简洁方式):**
```go
// Go - 简洁且类型安全
package main

import "fmt"

// 基本函数定义
func add(a, b int) int {
    return a + b
}

// 布尔返回值
func isEven(number int) bool {
    return number%2 == 0
}

// 多返回值处理错误
func readFile(filename string) (string, error) {
    // 实现...
    if err != nil {
        return "", err
    }
    return content, nil
}

// 多个相同类型参数的简写
func multiply(a, b, c int) int {
    return a * b * c
}

// 可变参数
func sum(numbers ...int) int {
    total := 0
    for _, num := range numbers {
        total += num
    }
    return total
}
```

### Go函数的独特特性

#### 1. 无需访问修饰符

```go
// Go通过首字母大小写控制可见性
func PublicFunction() {    // 大写开头 = public
    fmt.Println("包外可见")
}

func privateFunction() {   // 小写开头 = private
    fmt.Println("包内可见")
}

// Java需要显式声明
// public static void publicMethod() {}
// private static void privateMethod() {}
```

#### 2. 参数类型的简洁写法

```go
// 相同类型的参数可以合并声明
func calculate(a, b, c int, x, y float64) float64 {
    return float64(a+b+c) + x + y
}

// 等价于
func calculateVerbose(a int, b int, c int, x float64, y float64) float64 {
    return float64(a+b+c) + x + y
}
```

#### 3. 可变参数的强大功能

```go
// 基本可变参数
func printf(format string, args ...interface{}) {
    fmt.Printf(format, args...)
}

// 类型安全的可变参数
func max(numbers ...int) int {
    if len(numbers) == 0 {
        return 0
    }
    
    maxNum := numbers[0]
    for _, num := range numbers[1:] {
        if num > maxNum {
            maxNum = num
        }
    }
    return maxNum
}

// 使用示例
result := max(1, 5, 3, 9, 2)  // 9

// 展开切片作为可变参数
nums := []int{1, 5, 3, 9, 2}
result = max(nums...)  // 9
```

#### 4. 实际项目中的函数应用

让我们看看mall-go项目中的真实例子：

```go
// 来自 mall-go/pkg/utils/validator.go
func ValidateEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

func ValidatePassword(password string) (bool, string) {
    if len(password) < 6 {
        return false, "密码长度不能少于6位"
    }
    
    if len(password) > 20 {
        return false, "密码长度不能超过20位"
    }
    
    // 检查是否包含数字
    hasNumber := regexp.MustCompile(`\d`).MatchString(password)
    if !hasNumber {
        return false, "密码必须包含至少一个数字"
    }
    
    // 检查是否包含字母
    hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
    if !hasLetter {
        return false, "密码必须包含至少一个字母"
    }
    
    return true, "密码格式正确"
}

// 来自 mall-go/pkg/utils/crypto.go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("密码加密失败: %w", err)
    }
    return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

---

## 🔄 多返回值机制

### Go的杀手级特性

多返回值是Go语言最具特色的功能之一，它让错误处理变得优雅而明确。

#### 1. 基本多返回值语法

```java
// Java - 需要创建对象或使用数组
public class Result {
    private String value;
    private Exception error;
    
    public Result(String value, Exception error) {
        this.value = value;
        this.error = error;
    }
    
    // getter方法...
}

public Result divide(int a, int b) {
    if (b == 0) {
        return new Result(null, new ArithmeticException("除数不能为零"));
    }
    return new Result(String.valueOf(a / b), null);
}
```

```python
# Python - 使用元组
def divide(a, b):
    if b == 0:
        return None, "除数不能为零"
    return a / b, None

# 或者抛出异常
def divide_with_exception(a, b):
    if b == 0:
        raise ValueError("除数不能为零")
    return a / b
```

```go
// Go - 天然支持多返回值
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为零")
    }
    return a / b, nil
}

// 使用多返回值
result, err := divide(10, 2)
if err != nil {
    log.Printf("计算失败: %v", err)
    return
}
fmt.Printf("结果: %.2f\n", result)
```

#### 2. 多返回值的常见模式

```go
// 1. 值 + 错误模式（最常见）
func getUserByID(id uint) (*User, error) {
    var user User
    err := db.First(&user, id).Error
    if err != nil {
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }
    return &user, nil
}

// 2. 值 + 布尔模式（检查存在性）
func getFromCache(key string) (string, bool) {
    value, exists := cache[key]
    return value, exists
}

// 3. 多个值 + 错误模式
func parseUserInfo(data string) (string, int, string, error) {
    parts := strings.Split(data, ",")
    if len(parts) != 3 {
        return "", 0, "", fmt.Errorf("数据格式错误")
    }
    
    name := strings.TrimSpace(parts[0])
    age, err := strconv.Atoi(strings.TrimSpace(parts[1]))
    if err != nil {
        return "", 0, "", fmt.Errorf("年龄格式错误: %w", err)
    }
    
    email := strings.TrimSpace(parts[2])
    return name, age, email, nil
}

// 4. 忽略不需要的返回值
_, err := doSomething()  // 只关心错误
value, _ := getFromCache("key")  // 忽略存在性检查
```

#### 3. 实际项目中的多返回值应用

```go
// 来自 mall-go/pkg/database/transaction.go
func WithTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
    tx := db.Begin()
    if tx.Error != nil {
        return fmt.Errorf("开始事务失败: %w", tx.Error)
    }
    
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    if err := fn(tx); err != nil {
        tx.Rollback()
        return fmt.Errorf("事务执行失败: %w", err)
    }
    
    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("提交事务失败: %w", err)
    }
    
    return nil
}

// 来自 mall-go/pkg/auth/jwt.go
func GenerateToken(userID uint, username string) (string, time.Time, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    
    claims := &Claims{
        UserID:   userID,
        Username: username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
            IssuedAt:  time.Now().Unix(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", time.Time{}, fmt.Errorf("生成token失败: %w", err)
    }
    
    return tokenString, expirationTime, nil
}

// 使用示例
token, expiry, err := GenerateToken(user.ID, user.Username)
if err != nil {
    return c.JSON(http.StatusInternalServerError, gin.H{
        "error": "生成令牌失败",
    })
}

c.JSON(http.StatusOK, gin.H{
    "token":  token,
    "expiry": expiry,
})
```

---

## 🏷️ 命名返回值和裸返回

### Go的语法糖特性

命名返回值是Go提供的一个语法糖，可以让代码更清晰，特别是在复杂的函数中。

#### 1. 基本命名返回值语法

```go
// 普通返回值
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为零")
    }
    return a / b, nil
}

// 命名返回值
func divideNamed(a, b float64) (result float64, err error) {
    if b == 0 {
        err = fmt.Errorf("除数不能为零")
        return  // 裸返回，等价于 return result, err
    }
    result = a / b
    return  // 裸返回
}

// 也可以显式返回
func divideExplicit(a, b float64) (result float64, err error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为零")  // 显式返回
    }
    result = a / b
    return result, err  // 显式返回
}
```

#### 2. 命名返回值的优势

```go
// 1. 提高代码可读性
func calculateStats(numbers []int) (sum, avg, min, max int) {
    if len(numbers) == 0 {
        return  // 所有返回值都是零值
    }
    
    sum = numbers[0]
    min = numbers[0]
    max = numbers[0]
    
    for _, num := range numbers {
        sum += num
        if num < min {
            min = num
        }
        if num > max {
            max = num
        }
    }
    
    avg = sum / len(numbers)
    return  // 裸返回，清晰明了
}

// 2. 与defer结合的高级用法
func processFile(filename string) (err error) {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("打开文件失败: %w", err)
    }
    
    defer func() {
        if closeErr := file.Close(); closeErr != nil {
            // 如果原本没有错误，设置关闭错误
            if err == nil {
                err = fmt.Errorf("关闭文件失败: %w", closeErr)
            }
        }
    }()
    
    // 处理文件...
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return fmt.Errorf("读取文件失败: %w", err)
    }
    
    // 验证数据...
    if len(data) == 0 {
        return fmt.Errorf("文件为空")
    }
    
    return nil  // 或者直接 return
}
```

#### 3. 实际项目中的命名返回值

```go
// 来自 mall-go/pkg/user/service.go
func (s *UserService) CreateUser(req *CreateUserRequest) (user *User, err error) {
    // 参数验证
    if req.Username == "" {
        err = fmt.Errorf("用户名不能为空")
        return
    }
    
    // 检查用户名是否已存在
    var existingUser User
    if err = s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
        err = fmt.Errorf("用户名已存在")
        return
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        err = fmt.Errorf("查询用户失败: %w", err)
        return
    }
    
    // 密码加密
    var hashedPassword string
    if hashedPassword, err = HashPassword(req.Password); err != nil {
        err = fmt.Errorf("密码加密失败: %w", err)
        return
    }
    
    // 创建用户
    user = &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
        Status:   "active",
    }
    
    if err = s.db.Create(user).Error; err != nil {
        err = fmt.Errorf("创建用户失败: %w", err)
        user = nil  // 确保返回nil
        return
    }
    
    return  // 成功时的裸返回
}
```

#### 4. 命名返回值的注意事项

```go
// ❌ 错误：命名返回值的遮蔽问题
func badExample() (result int, err error) {
    if result, err := doSomething(); err != nil {  // 遮蔽了命名返回值！
        return 0, err  // 这里的result是局部变量，不是返回值
    }
    return  // result仍然是零值！
}

// ✅ 正确：避免遮蔽
func goodExample() (result int, err error) {
    result, err = doSomething()  // 直接赋值给命名返回值
    if err != nil {
        return 0, err
    }
    return
}

// ❌ 错误：裸返回在长函数中的可读性问题
func longFunction() (a, b, c, d int, err error) {
    // ... 100行代码 ...
    
    a = 1
    // ... 50行代码 ...
    
    b = 2
    // ... 30行代码 ...
    
    return  // 很难知道返回了什么值
}

// ✅ 正确：在长函数中使用显式返回
func longFunctionBetter() (a, b, c, d int, err error) {
    // ... 复杂逻辑 ...
    
    return a, b, c, d, err  // 明确显示返回值
}
```

---

## 🎭 函数作为一等公民

### Go中的函数式编程特性

在Go中，函数是一等公民，这意味着函数可以：
- 作为变量存储
- 作为参数传递
- 作为返回值返回
- 在运行时创建

#### 1. 函数类型和函数变量

```java
// Java - 需要使用接口或Lambda表达式
@FunctionalInterface
interface Calculator {
    int calculate(int a, int b);
}

Calculator add = (a, b) -> a + b;
Calculator multiply = (a, b) -> a * b;

public int compute(int x, int y, Calculator calc) {
    return calc.calculate(x, y);
}
```

```python
# Python - 函数天然是一等公民
def add(a, b):
    return a + b

def multiply(a, b):
    return a * b

def compute(x, y, func):
    return func(x, y)

# 使用
result = compute(5, 3, add)  # 8
```

```go
// Go - 函数类型定义和使用
type Calculator func(int, int) int

// 定义具体的函数
func add(a, b int) int {
    return a + b
}

func multiply(a, b int) int {
    return a * b
}

// 函数作为参数
func compute(x, y int, calc Calculator) int {
    return calc(x, y)
}

// 使用示例
func main() {
    // 函数变量
    var operation Calculator
    operation = add
    result := operation(5, 3)  // 8
    
    // 直接传递函数
    result = compute(5, 3, multiply)  // 15
    
    // 匿名函数
    result = compute(5, 3, func(a, b int) int {
        return a - b
    })  // 2
}
```

#### 2. 高阶函数的实现

```go
// 返回函数的函数
func makeMultiplier(factor int) func(int) int {
    return func(x int) int {
        return x * factor
    }
}

// 使用示例
double := makeMultiplier(2)
triple := makeMultiplier(3)

fmt.Println(double(5))  // 10
fmt.Println(triple(5))  // 15

// 函数组合
func compose(f, g func(int) int) func(int) int {
    return func(x int) int {
        return f(g(x))
    }
}

// 使用函数组合
addOne := func(x int) int { return x + 1 }
square := func(x int) int { return x * x }

addThenSquare := compose(square, addOne)
fmt.Println(addThenSquare(3))  // (3+1)² = 16
```

#### 3. 实际项目中的函数式应用

```go
// 来自 mall-go/pkg/middleware/auth.go
type AuthMiddleware func(gin.HandlerFunc) gin.HandlerFunc

func RequireAuth(authService *AuthService) AuthMiddleware {
    return func(next gin.HandlerFunc) gin.HandlerFunc {
        return func(c *gin.Context) {
            token := c.GetHeader("Authorization")
            if token == "" {
                c.JSON(http.StatusUnauthorized, gin.H{
                    "error": "缺少认证令牌",
                })
                c.Abort()
                return
            }
            
            userID, err := authService.ValidateToken(token)
            if err != nil {
                c.JSON(http.StatusUnauthorized, gin.H{
                    "error": "无效的认证令牌",
                })
                c.Abort()
                return
            }
            
            c.Set("userID", userID)
            next(c)
        }
    }
}

// 来自 mall-go/pkg/utils/functional.go
// 函数式编程工具函数
func Map[T, R any](slice []T, fn func(T) R) []R {
    result := make([]R, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

func Reduce[T, R any](slice []T, initial R, fn func(R, T) R) R {
    result := initial
    for _, v := range slice {
        result = fn(result, v)
    }
    return result
}

// 使用示例
numbers := []int{1, 2, 3, 4, 5}

// 映射：每个数字乘以2
doubled := Map(numbers, func(x int) int { return x * 2 })
// [2, 4, 6, 8, 10]

// 过滤：只保留偶数
evens := Filter(numbers, func(x int) bool { return x%2 == 0 })
// [2, 4]

// 归约：计算总和
sum := Reduce(numbers, 0, func(acc, x int) int { return acc + x })
// 15
```

---

## 🏗️ 方法定义和接收者

### Go的"面向对象"实现方式

Go没有传统的类，但通过方法和接收者实现了面向对象的特性。

#### 1. 方法vs函数的区别

```java
// Java - 类中的方法
public class User {
    private String name;
    private int age;

    public User(String name, int age) {
        this.name = name;
        this.age = age;
    }

    public String getName() {
        return this.name;
    }

    public void setAge(int age) {
        this.age = age;
    }

    public String getInfo() {
        return "Name: " + this.name + ", Age: " + this.age;
    }
}
```

```python
# Python - 类中的方法
class User:
    def __init__(self, name, age):
        self.name = name
        self.age = age

    def get_name(self):
        return self.name

    def set_age(self, age):
        self.age = age

    def get_info(self):
        return f"Name: {self.name}, Age: {self.age}"
```

```go
// Go - 结构体 + 方法
type User struct {
    Name string
    Age  int
}

// 值接收者方法
func (u User) GetName() string {
    return u.Name
}

// 指针接收者方法
func (u *User) SetAge(age int) {
    u.Age = age
}

// 值接收者方法
func (u User) GetInfo() string {
    return fmt.Sprintf("Name: %s, Age: %d", u.Name, u.Age)
}

// 使用示例
user := User{Name: "张三", Age: 25}
fmt.Println(user.GetName())  // 张三

user.SetAge(26)
fmt.Println(user.GetInfo())  // Name: 张三, Age: 26
```

#### 2. 为任何类型定义方法

Go的强大之处在于可以为任何类型定义方法，不仅仅是结构体：

```go
// 为基本类型定义方法
type MyInt int

func (m MyInt) IsEven() bool {
    return int(m)%2 == 0
}

func (m MyInt) Double() MyInt {
    return m * 2
}

// 使用示例
var num MyInt = 5
fmt.Println(num.IsEven())  // false
fmt.Println(num.Double())  // 10

// 为切片类型定义方法
type IntSlice []int

func (is IntSlice) Sum() int {
    total := 0
    for _, v := range is {
        total += v
    }
    return total
}

func (is IntSlice) Average() float64 {
    if len(is) == 0 {
        return 0
    }
    return float64(is.Sum()) / float64(len(is))
}

// 使用示例
numbers := IntSlice{1, 2, 3, 4, 5}
fmt.Println(numbers.Sum())      // 15
fmt.Println(numbers.Average())  // 3.0

// 为映射类型定义方法
type StringCounter map[string]int

func (sc StringCounter) Add(key string) {
    sc[key]++
}

func (sc StringCounter) Count(key string) int {
    return sc[key]
}

func (sc StringCounter) Total() int {
    total := 0
    for _, count := range sc {
        total += count
    }
    return total
}

// 使用示例
counter := make(StringCounter)
counter.Add("apple")
counter.Add("banana")
counter.Add("apple")
fmt.Println(counter.Count("apple"))  // 2
fmt.Println(counter.Total())         // 3
```

#### 3. 实际项目中的方法设计

```go
// 来自 mall-go/internal/model/user.go
type User struct {
    ID        uint      `json:"id" gorm:"primarykey"`
    Username  string    `json:"username" gorm:"uniqueIndex;not null"`
    Email     string    `json:"email" gorm:"uniqueIndex;not null"`
    Password  string    `json:"-" gorm:"not null"`
    Status    string    `json:"status" gorm:"default:active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 值接收者方法 - 不修改原对象
func (u User) IsActive() bool {
    return u.Status == "active"
}

func (u User) GetDisplayName() string {
    if u.Username != "" {
        return u.Username
    }
    return u.Email
}

func (u User) MaskEmail() string {
    parts := strings.Split(u.Email, "@")
    if len(parts) != 2 {
        return u.Email
    }

    username := parts[0]
    domain := parts[1]

    if len(username) <= 2 {
        return u.Email
    }

    masked := username[:1] + strings.Repeat("*", len(username)-2) + username[len(username)-1:]
    return masked + "@" + domain
}

// 指针接收者方法 - 修改原对象
func (u *User) Activate() {
    u.Status = "active"
    u.UpdatedAt = time.Now()
}

func (u *User) Deactivate() {
    u.Status = "inactive"
    u.UpdatedAt = time.Now()
}

func (u *User) UpdatePassword(newPassword string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("密码加密失败: %w", err)
    }

    u.Password = string(hashedPassword)
    u.UpdatedAt = time.Now()
    return nil
}

// 验证密码
func (u User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

---

## ⚖️ 值接收者vs指针接收者

### 选择的黄金法则

这是Go语言中最重要的设计决策之一，直接影响性能、语义和接口实现。

#### 1. 基本概念对比

```go
type Counter struct {
    value int
}

// 值接收者 - 接收结构体的副本
func (c Counter) GetValue() int {
    return c.value  // 读取操作，不修改原对象
}

func (c Counter) AddValue(n int) {
    c.value += n  // ❌ 修改的是副本，原对象不变！
}

// 指针接收者 - 接收结构体的指针
func (c *Counter) SetValue(n int) {
    c.value = n  // ✅ 修改原对象
}

func (c *Counter) Increment() {
    c.value++  // ✅ 修改原对象
}

// 使用示例
func main() {
    counter := Counter{value: 0}

    fmt.Println(counter.GetValue())  // 0

    counter.AddValue(5)  // 值接收者，不会修改原对象
    fmt.Println(counter.GetValue())  // 仍然是 0！

    counter.SetValue(10)  // 指针接收者，修改原对象
    fmt.Println(counter.GetValue())  // 10

    counter.Increment()  // 指针接收者，修改原对象
    fmt.Println(counter.GetValue())  // 11
}
```

#### 2. 选择接收者类型的原则

```go
// 1. 需要修改接收者时，必须使用指针接收者
type BankAccount struct {
    balance decimal.Decimal
}

func (ba *BankAccount) Deposit(amount decimal.Decimal) {
    ba.balance = ba.balance.Add(amount)  // 修改余额
}

func (ba *BankAccount) Withdraw(amount decimal.Decimal) error {
    if ba.balance.LessThan(amount) {
        return fmt.Errorf("余额不足")
    }
    ba.balance = ba.balance.Sub(amount)  // 修改余额
    return nil
}

func (ba BankAccount) GetBalance() decimal.Decimal {
    return ba.balance  // 只读操作，可以使用值接收者
}

// 2. 大型结构体使用指针接收者避免复制开销
type LargeStruct struct {
    data [1000000]int  // 大型数组
    name string
}

// ❌ 错误：值接收者会复制整个大型结构体
func (ls LargeStruct) ProcessData() {
    // 每次调用都会复制4MB的数据！
    for i := range ls.data {
        ls.data[i] *= 2  // 而且修改的是副本
    }
}

// ✅ 正确：指针接收者避免复制
func (ls *LargeStruct) ProcessData() {
    // 只传递指针，高效且能修改原数据
    for i := range ls.data {
        ls.data[i] *= 2
    }
}

func (ls *LargeStruct) GetName() string {
    return ls.name  // 即使是读操作，为了一致性也使用指针接收者
}

// 3. 保持一致性：如果有指针接收者方法，建议都用指针接收者
type User struct {
    name string
    age  int
}

func (u *User) SetName(name string) {
    u.name = name  // 指针接收者
}

func (u *User) GetName() string {
    return u.name  // 为了一致性，也使用指针接收者
}

func (u *User) GetAge() int {
    return u.age  // 保持一致性
}

// 4. 小型结构体的只读操作可以使用值接收者
type Point struct {
    X, Y float64
}

func (p Point) Distance() float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)  // 只读操作，结构体小
}

func (p Point) Add(other Point) Point {
    return Point{X: p.X + other.X, Y: p.Y + other.Y}  // 返回新值，不修改原值
}

func (p *Point) Move(dx, dy float64) {
    p.X += dx  // 修改原对象
    p.Y += dy
}
```

#### 3. 接收者类型对方法集的影响

```go
type MyType struct {
    value int
}

// 值接收者方法
func (m MyType) ValueMethod() {
    fmt.Println("值接收者方法")
}

// 指针接收者方法
func (m *MyType) PointerMethod() {
    fmt.Println("指针接收者方法")
}

func main() {
    // 值类型变量
    var value MyType
    value.ValueMethod()    // ✅ 直接调用
    value.PointerMethod()  // ✅ Go自动取地址：(&value).PointerMethod()

    // 指针类型变量
    var pointer *MyType = &MyType{}
    pointer.ValueMethod()    // ✅ Go自动解引用：(*pointer).ValueMethod()
    pointer.PointerMethod()  // ✅ 直接调用

    // 接口实现的差异
    var iface interface{} = value
    if method, ok := iface.(interface{ ValueMethod() }); ok {
        method.ValueMethod()  // ✅ 值类型实现了值接收者方法
    }

    if method, ok := iface.(interface{ PointerMethod() }); ok {
        method.PointerMethod()  // ❌ 值类型没有实现指针接收者方法
    } else {
        fmt.Println("值类型没有实现指针接收者方法")
    }

    // 指针类型实现了所有方法
    var ifacePtr interface{} = &value
    if method, ok := ifacePtr.(interface{ ValueMethod() }); ok {
        method.ValueMethod()  // ✅ 指针类型实现了值接收者方法
    }

    if method, ok := ifacePtr.(interface{ PointerMethod() }); ok {
        method.PointerMethod()  // ✅ 指针类型实现了指针接收者方法
    }
}
```

---

## 🎯 面试常考点

### 1. 函数和方法的区别

**面试题**: "Go语言中函数和方法有什么区别？"

**标准答案**:
```go
// 函数 - 独立存在，不属于任何类型
func Add(a, b int) int {
    return a + b
}

// 方法 - 属于特定类型，有接收者
type Calculator struct{}

func (c Calculator) Add(a, b int) int {
    return a + b
}

// 区别：
// 1. 方法有接收者，函数没有
// 2. 方法可以访问接收者的字段和其他方法
// 3. 方法参与接口实现，函数不参与
// 4. 方法调用语法：receiver.Method()，函数调用：Function()
```

### 2. 值接收者vs指针接收者的选择

**面试题**: "什么时候使用值接收者，什么时候使用指针接收者？"

**标准答案**:
```go
// 使用指针接收者的情况：
// 1. 需要修改接收者
func (u *User) SetName(name string) {
    u.Name = name
}

// 2. 接收者是大型结构体（避免复制）
type LargeStruct struct {
    data [1000000]int
}
func (ls *LargeStruct) Process() { /* 避免复制 */ }

// 3. 保持一致性（如果有指针接收者方法，建议都用指针）
func (u *User) GetName() string { return u.Name }

// 使用值接收者的情况：
// 1. 只读操作且结构体较小
type Point struct { X, Y float64 }
func (p Point) Distance() float64 { return math.Sqrt(p.X*p.X + p.Y*p.Y) }

// 2. 基本类型的别名
type Counter int
func (c Counter) String() string { return fmt.Sprintf("%d", c) }
```

### 3. 多返回值的最佳实践

**面试题**: "Go语言多返回值有什么最佳实践？"

**标准答案**:
```go
// 1. 错误处理模式（最常见）
func ReadFile(filename string) ([]byte, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("读取文件失败: %w", err)
    }
    return data, nil
}

// 2. 值 + 布尔模式（检查存在性）
func GetFromMap(m map[string]int, key string) (int, bool) {
    value, exists := m[key]
    return value, exists
}

// 3. 命名返回值提高可读性
func Divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = fmt.Errorf("除数不能为零")
        return
    }
    result = a / b
    return
}

// 4. 忽略不需要的返回值
_, err := doSomething()  // 只关心错误
value, _ := getFromCache("key")  // 忽略存在性检查
```

### 4. 闭包的内存管理

**面试题**: "Go语言中闭包是如何管理内存的？"

**标准答案**:
```go
// 闭包会捕获外部变量，延长其生命周期
func createCounter() func() int {
    count := 0  // 这个变量会被闭包捕获
    return func() int {
        count++  // 闭包引用外部变量
        return count
    }
}

// 内存管理要点：
// 1. 被闭包引用的变量会逃逸到堆上
// 2. 只要闭包存在，被引用的变量就不会被GC回收
// 3. 每个闭包都有自己的变量副本
// 4. 避免在循环中创建大量闭包导致内存泄漏

// 常见陷阱：
funcs := make([]func(), 0)
for i := 0; i < 3; i++ {
    // ❌ 错误：所有闭包都引用同一个i
    funcs = append(funcs, func() {
        fmt.Println(i)  // 都会打印3
    })
}

// ✅ 正确：每个闭包捕获不同的变量
for i := 0; i < 3; i++ {
    i := i  // 创建新变量
    funcs = append(funcs, func() {
        fmt.Println(i)  // 正确打印0, 1, 2
    })
}
```

### 5. 接口实现的隐式特性

**面试题**: "Go接口的隐式实现有什么优势和注意事项？"

**标准答案**:
```go
// 优势：
// 1. 解耦 - 实现者不需要知道接口的存在
// 2. 灵活 - 可以为第三方类型实现接口
// 3. 测试友好 - 容易创建mock对象
// 4. 渐进式设计 - 可以后续抽象出接口

type Writer interface {
    Write([]byte) (int, error)
}

// 任何有Write方法的类型都自动实现了Writer接口
type MyWriter struct{}

func (mw MyWriter) Write(data []byte) (int, error) {
    // 实现写入逻辑
    return len(data), nil
}

// MyWriter自动实现了Writer接口，无需显式声明

// 注意事项：
// 1. 方法签名必须完全匹配
// 2. 值类型和指针类型的方法集不同
// 3. 接口应该小而专注（接口隔离原则）
// 4. 在使用方定义接口，而不是实现方
```

---

## 💡 踩坑提醒

### 1. 值接收者和指针接收者的混用陷阱

```go
// ❌ 错误：混用值接收者和指针接收者
type User struct {
    Name string
}

func (u User) SetName(name string) {    // 值接收者，不会修改原对象
    u.Name = name
}

func (u *User) GetName() string {       // 指针接收者
    return u.Name
}

// 使用时的困惑
user := User{Name: "张三"}
user.SetName("李四")
fmt.Println(user.GetName())  // 仍然是"张三"！

// ✅ 正确：保持一致性
func (u *User) SetName(name string) {   // 都用指针接收者
    u.Name = name
}

func (u *User) GetName() string {
    return u.Name
}
```

### 2. 闭包中的循环变量陷阱

```go
// ❌ 错误：闭包捕获循环变量
var funcs []func()
for i := 0; i < 3; i++ {
    funcs = append(funcs, func() {
        fmt.Println(i)  // 都会打印3！
    })
}

// ✅ 正确：捕获循环变量的副本
var funcs []func()
for i := 0; i < 3; i++ {
    i := i  // 创建新变量
    funcs = append(funcs, func() {
        fmt.Println(i)  // 正确打印0, 1, 2
    })
}

// 或者使用参数传递
for i := 0; i < 3; i++ {
    funcs = append(funcs, func(n int) func() {
        return func() {
            fmt.Println(n)
        }
    }(i))
}
```

### 3. 命名返回值的遮蔽问题

```go
// ❌ 错误：命名返回值被遮蔽
func badExample() (result int, err error) {
    if result, err := doSomething(); err != nil {  // 遮蔽了命名返回值！
        return 0, err  // 这里的result是局部变量
    }
    return  // result仍然是零值！
}

// ✅ 正确：避免遮蔽
func goodExample() (result int, err error) {
    result, err = doSomething()  // 直接赋值给命名返回值
    if err != nil {
        return 0, err
    }
    return
}
```

### 4. 方法集和接口实现的陷阱

```go
type MyInterface interface {
    Method()
}

type MyStruct struct{}

func (m *MyStruct) Method() {  // 只有指针接收者方法
    fmt.Println("方法调用")
}

func main() {
    var s MyStruct
    var p *MyStruct = &s

    // 方法调用没问题
    s.Method()  // Go自动取地址
    p.Method()  // 直接调用

    // 接口实现有问题
    var iface MyInterface
    // iface = s  // ❌ 编译错误：MyStruct没有实现Method()
    iface = p     // ✅ 正确：*MyStruct实现了Method()
    iface.Method()
}
```

---

## 📝 本章练习题

### 基础练习

1. **函数定义和多返回值练习**
```go
// 编写以下函数：
// 1. 计算两个数的四则运算，返回结果和可能的错误
// 2. 解析字符串为整数，返回结果和是否成功
// 3. 查找切片中的最大值和最小值

// 参考答案：
func Calculate(a, b float64, op string) (float64, error) {
    switch op {
    case "+":
        return a + b, nil
    case "-":
        return a - b, nil
    case "*":
        return a * b, nil
    case "/":
        if b == 0 {
            return 0, fmt.Errorf("除数不能为零")
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("不支持的操作符: %s", op)
    }
}

func ParseInt(s string) (int, bool) {
    if num, err := strconv.Atoi(s); err == nil {
        return num, true
    }
    return 0, false
}

func FindMinMax(numbers []int) (min, max int, err error) {
    if len(numbers) == 0 {
        return 0, 0, fmt.Errorf("切片不能为空")
    }

    min, max = numbers[0], numbers[0]
    for _, num := range numbers[1:] {
        if num < min {
            min = num
        }
        if num > max {
            max = num
        }
    }
    return min, max, nil
}
```

2. **方法定义练习**
```go
// 为以下类型定义方法：
// 1. Rectangle结构体：计算面积和周长
// 2. StringSlice类型：排序、去重、查找

// 参考答案：
type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor
    r.Height *= factor
}

type StringSlice []string

func (ss StringSlice) Len() int           { return len(ss) }
func (ss StringSlice) Less(i, j int) bool { return ss[i] < ss[j] }
func (ss StringSlice) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }

func (ss *StringSlice) Sort() {
    sort.Sort(ss)
}

func (ss StringSlice) Contains(target string) bool {
    for _, s := range ss {
        if s == target {
            return true
        }
    }
    return false
}

func (ss StringSlice) Unique() StringSlice {
    seen := make(map[string]bool)
    var result StringSlice

    for _, s := range ss {
        if !seen[s] {
            seen[s] = true
            result = append(result, s)
        }
    }
    return result
}
```

### 进阶练习

3. **函数式编程练习**
```go
// 实现通用的函数式编程工具：
// 1. Map函数：对切片每个元素应用函数
// 2. Filter函数：过滤切片元素
// 3. Reduce函数：归约切片为单个值

// 参考答案：
func Map[T, R any](slice []T, fn func(T) R) []R {
    result := make([]R, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

func Reduce[T, R any](slice []T, initial R, fn func(R, T) R) R {
    result := initial
    for _, v := range slice {
        result = fn(result, v)
    }
    return result
}

// 使用示例
func main() {
    numbers := []int{1, 2, 3, 4, 5}

    // 映射：每个数字平方
    squares := Map(numbers, func(x int) int { return x * x })
    fmt.Println(squares)  // [1, 4, 9, 16, 25]

    // 过滤：只保留偶数
    evens := Filter(numbers, func(x int) bool { return x%2 == 0 })
    fmt.Println(evens)  // [2, 4]

    // 归约：计算总和
    sum := Reduce(numbers, 0, func(acc, x int) int { return acc + x })
    fmt.Println(sum)  // 15
}
```

4. **闭包和装饰器练习**
```go
// 实现以下装饰器：
// 1. 计时装饰器：测量函数执行时间
// 2. 重试装饰器：失败时自动重试
// 3. 缓存装饰器：缓存函数结果

// 参考答案：
func WithTiming[T any](fn func() T) func() T {
    return func() T {
        start := time.Now()
        result := fn()
        duration := time.Since(start)
        fmt.Printf("函数执行时间: %v\n", duration)
        return result
    }
}

func WithRetry[T any](maxAttempts int, delay time.Duration, fn func() (T, error)) func() (T, error) {
    return func() (T, error) {
        var result T
        var lastErr error

        for attempt := 1; attempt <= maxAttempts; attempt++ {
            if result, err := fn(); err != nil {
                lastErr = err
                if attempt < maxAttempts {
                    fmt.Printf("第 %d 次尝试失败: %v，%v 后重试\n", attempt, err, delay)
                    time.Sleep(delay)
                    continue
                }
            } else {
                return result, nil
            }
        }

        return result, fmt.Errorf("重试 %d 次后仍然失败: %w", maxAttempts, lastErr)
    }
}

func WithCache[K comparable, V any](fn func(K) V) func(K) V {
    cache := make(map[K]V)
    mutex := sync.RWMutex{}

    return func(key K) V {
        // 先尝试从缓存读取
        mutex.RLock()
        if value, exists := cache[key]; exists {
            mutex.RUnlock()
            fmt.Printf("缓存命中: %v\n", key)
            return value
        }
        mutex.RUnlock()

        // 缓存未命中，计算结果
        fmt.Printf("缓存未命中，计算结果: %v\n", key)
        result := fn(key)

        // 存入缓存
        mutex.Lock()
        cache[key] = result
        mutex.Unlock()

        return result
    }
}

// 使用示例
func expensiveCalculation(n int) int {
    time.Sleep(100 * time.Millisecond)  // 模拟耗时计算
    return n * n
}

func main() {
    // 使用缓存装饰器
    cachedCalc := WithCache(expensiveCalculation)

    fmt.Println(cachedCalc(5))  // 缓存未命中，计算结果
    fmt.Println(cachedCalc(5))  // 缓存命中

    // 使用计时装饰器
    timedCalc := WithTiming(func() int {
        return expensiveCalculation(3)
    })

    result := timedCalc()
    fmt.Printf("结果: %d\n", result)
}
```

### 高级练习

5. **中间件模式练习**
```go
// 实现HTTP中间件系统：
// 1. 日志中间件：记录请求信息
// 2. 认证中间件：验证用户身份
// 3. 限流中间件：控制请求频率

// 参考答案：
type HandlerFunc func(http.ResponseWriter, *http.Request)
type Middleware func(HandlerFunc) HandlerFunc

// 日志中间件
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // 包装ResponseWriter以捕获状态码
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        next(wrapped, r)

        duration := time.Since(start)
        fmt.Printf("[%s] %s %s %d %v\n",
            time.Now().Format("2006-01-02 15:04:05"),
            r.Method,
            r.URL.Path,
            wrapped.statusCode,
            duration,
        )
    }
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// 认证中间件
func AuthMiddleware(next HandlerFunc) HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "缺少认证令牌", http.StatusUnauthorized)
            return
        }

        // 简单的令牌验证（实际项目中应该更复杂）
        if !strings.HasPrefix(token, "Bearer ") {
            http.Error(w, "无效的认证令牌格式", http.StatusUnauthorized)
            return
        }

        // 验证通过，继续处理
        next(w, r)
    }
}

// 限流中间件
func RateLimitMiddleware(maxRequests int, window time.Duration) Middleware {
    requests := make(map[string][]time.Time)
    mutex := sync.RWMutex{}

    return func(next HandlerFunc) HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            clientIP := r.RemoteAddr
            now := time.Now()

            mutex.Lock()
            defer mutex.Unlock()

            // 清理过期的请求记录
            if times, exists := requests[clientIP]; exists {
                var validTimes []time.Time
                for _, t := range times {
                    if now.Sub(t) < window {
                        validTimes = append(validTimes, t)
                    }
                }
                requests[clientIP] = validTimes
            }

            // 检查是否超过限制
            if len(requests[clientIP]) >= maxRequests {
                http.Error(w, "请求过于频繁", http.StatusTooManyRequests)
                return
            }

            // 记录当前请求
            requests[clientIP] = append(requests[clientIP], now)

            next(w, r)
        }
    }
}

// 中间件链
func Chain(middlewares ...Middleware) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        for i := len(middlewares) - 1; i >= 0; i-- {
            next = middlewares[i](next)
        }
        return next
    }
}

// 使用示例
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
    // 组合中间件
    handler := Chain(
        LoggingMiddleware,
        AuthMiddleware,
        RateLimitMiddleware(10, time.Minute),
    )(helloHandler)

    http.HandleFunc("/hello", handler)

    fmt.Println("服务器启动在 :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## 🎉 本章总结

通过本章学习，你应该掌握了：

### ✅ 核心概念
- [x] Go函数的定义语法和与Java/Python的差异
- [x] 多返回值机制的强大功能和最佳实践
- [x] 命名返回值和裸返回的使用技巧
- [x] 函数作为一等公民的函数式编程特性
- [x] 方法定义和接收者类型的选择原则
- [x] 值接收者vs指针接收者的性能和语义差异
- [x] 方法集和接口实现的隐式机制
- [x] 匿名函数和闭包的高级应用

### ✅ 实际应用
- [x] 用户模型的方法设计和最佳实践
- [x] 中间件模式的函数式实现
- [x] 装饰器模式的闭包应用
- [x] 函数式编程工具的通用实现
- [x] 接口设计的解耦和抽象

### ✅ 最佳实践
- [x] 函数和方法的设计原则
- [x] 接收者类型的选择策略
- [x] 多返回值的错误处理模式
- [x] 闭包的内存管理和性能考虑
- [x] 接口实现的隐式特性和方法集规则

### 🚀 下一步学习

恭喜完成基础篇第三章！接下来我们将学习：
- **[包管理与模块系统](./04-packages-and-imports.md)** - Go的模块化设计
- **[进阶篇：错误处理最佳实践](../02-advanced/02-error-handling.md)** - 深入的错误处理

---

> 💡 **学习提示**:
> 1. 多练习方法和接收者的设计，这是Go面向对象的核心
> 2. 重点理解值接收者vs指针接收者的选择，这是面试重点
> 3. 掌握函数式编程的思想，提升代码的抽象能力
> 4. 结合实际项目理解接口设计的解耦价值

**继续加油！Go语言的函数和方法正在让你的代码更加优雅和高效！** 🎯
```
```
```
```
