# Go语言控制结构与流程控制详解

> 🎯 **学习目标**: 掌握Go语言的控制结构语法，理解与传统语言的差异和优势
> 
> ⏱️ **预计学习时间**: 2-3小时
> 
> 📚 **前置知识**: 已完成变量和类型学习

## 📋 本章内容概览

- [if/else条件语句](#ifelse条件语句)
- [for循环的多种形态](#for循环的多种形态)
- [switch语句的强大功能](#switch语句的强大功能)
- [goto和标签控制](#goto和标签控制)
- [defer延迟执行机制](#defer延迟执行机制)
- [错误处理中的控制流](#错误处理中的控制流)
- [实战案例分析](#实战案例分析)
- [面试常考点](#面试常考点)

---

## 🔀 if/else条件语句

### Java vs Python vs Go 语法对比

**Java (你熟悉的方式):**
```java
// Java - 必须有括号，支持三元运算符
int age = 25;
if (age >= 18) {
    System.out.println("成年人");
} else if (age >= 13) {
    System.out.println("青少年");
} else {
    System.out.println("儿童");
}

// 三元运算符
String status = (age >= 18) ? "成年" : "未成年";

// 条件可以是复杂表达式
if ((age >= 18 && hasLicense) || isEmergency) {
    // 允许驾驶
}
```

**Python (你熟悉的方式):**
```python
# Python - 使用冒号和缩进
age = 25
if age >= 18:
    print("成年人")
elif age >= 13:
    print("青少年")
else:
    print("儿童")

# 三元表达式
status = "成年" if age >= 18 else "未成年"

# 链式比较
if 13 <= age < 18:
    print("青少年")
```

**Go (新的简洁方式):**
```go
// Go - 无需括号，无三元运算符，但有初始化语句
age := 25
if age >= 18 {
    fmt.Println("成年人")
} else if age >= 13 {
    fmt.Println("青少年")
} else {
    fmt.Println("儿童")
}

// Go的独特特性：if语句中的初始化
if status := getStatus(age); status == "adult" {
    fmt.Println("成年人")
    // status变量只在if块中可见
}

// 错误处理的常见模式
if err := doSomething(); err != nil {
    log.Fatal(err)
    return
}
// err变量在这里不可见
```

### Go if语句的独特特性

#### 1. 无需括号，但必须有大括号

```go
// ✅ 正确：无括号，有大括号
if x > 0 {
    fmt.Println("正数")
}

// ❌ 错误：有括号（虽然能编译，但不符合Go风格）
if (x > 0) {
    fmt.Println("正数")
}

// ❌ 错误：无大括号（编译错误）
if x > 0
    fmt.Println("正数")  // 编译错误！

// ❌ 错误：大括号不能换行（编译错误）
if x > 0 
{  // 编译错误！Go要求大括号在同一行
    fmt.Println("正数")
}
```

#### 2. if语句中的初始化语句

这是Go的一个强大特性，可以在条件判断前进行变量初始化：

```go
// 基本用法
if num := rand.Intn(100); num > 50 {
    fmt.Printf("大数字: %d\n", num)
} else {
    fmt.Printf("小数字: %d\n", num)
}
// num变量在这里不可见

// 错误处理的经典模式
if file, err := os.Open("config.txt"); err != nil {
    log.Printf("打开文件失败: %v", err)
} else {
    defer file.Close()  // 确保文件关闭
    // 处理文件...
}

// 类型断言
if str, ok := value.(string); ok {
    fmt.Printf("字符串值: %s\n", str)
} else {
    fmt.Println("不是字符串类型")
}
```

#### 3. 实际项目中的if语句应用

让我们看看mall-go项目中的真实例子：

```go
// 来自 mall-go/pkg/database/database.go
func Init() *gorm.DB {
    var err error
    cfg := config.GlobalConfig
    
    // 根据驱动类型连接数据库
    if cfg.Database.Driver == "sqlite" {
        DB, err = gorm.Open(sqlite.Open(cfg.Database.DBName), gormConfig)
    } else if cfg.Database.Driver == "memory" {
        DB, err = gorm.Open(sqlite.Open(":memory:"), gormConfig)
        log.Println("使用内存数据库模式（仅用于测试）")
    } else {
        // MySQL连接 - 复杂的条件判断
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            cfg.Database.Username,
            cfg.Database.Password,
            cfg.Database.Host,
            cfg.Database.Port,
            cfg.Database.DBName,
        )
        DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
        
        // 如果数据库不存在，尝试创建数据库
        if err != nil && (err.Error() == "Error 1049: Unknown database '"+cfg.Database.DBName+"'" ||
            fmt.Sprintf("Error 1049 (42000): Unknown database '%s'", cfg.Database.DBName) == err.Error()) {
            log.Printf("数据库 %s 不存在，尝试创建...", cfg.Database.DBName)
            
            // 嵌套的if语句处理数据库创建
            if rootDB, rootErr := gorm.Open(mysql.Open(rootDSN), gormConfig); rootErr != nil {
                log.Printf("连接MySQL服务器失败: %v", rootErr)
            } else {
                createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET UTF8MB4 COLLATE UTF8MB4_UNICODE_CI", cfg.Database.DBName)
                if createErr := rootDB.Exec(createSQL).Error; createErr != nil {
                    log.Printf("创建数据库失败: %v", createErr)
                } else {
                    log.Printf("数据库 %s 创建成功", cfg.Database.DBName)
                    DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
                }
            }
        }
    }
    
    if err != nil {
        log.Fatalf("数据库连接失败: %v", err)
    }
    
    return DB
}
```

---

## 🔄 for循环的多种形态

### Go只有for循环！

与Java和Python不同，Go语言只有一种循环结构：`for`循环。但它非常灵活，可以实现其他语言中while、do-while等循环的功能。

#### 1. 传统的三段式for循环

```java
// Java
for (int i = 0; i < 10; i++) {
    System.out.println(i);
}
```

```python
# Python
for i in range(10):
    print(i)
```

```go
// Go - 类似Java的语法，但无需括号
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// 可以省略初始化语句
i := 0
for ; i < 10; i++ {
    fmt.Println(i)
}

// 可以省略后置语句
for i := 0; i < 10; {
    fmt.Println(i)
    i++
}
```

#### 2. while循环的实现

```java
// Java while循环
int i = 0;
while (i < 10) {
    System.out.println(i);
    i++;
}
```

```python
# Python while循环
i = 0
while i < 10:
    print(i)
    i += 1
```

```go
// Go实现while循环 - 只保留条件
i := 0
for i < 10 {
    fmt.Println(i)
    i++
}

// 无限循环（相当于while(true)）
for {
    fmt.Println("无限循环")
    if someCondition {
        break
    }
}
```

#### 3. range循环 - Go的强大特性

```go
// 遍历切片
numbers := []int{1, 2, 3, 4, 5}

// 获取索引和值
for index, value := range numbers {
    fmt.Printf("索引: %d, 值: %d\n", index, value)
}

// 只要值，忽略索引
for _, value := range numbers {
    fmt.Printf("值: %d\n", value)
}

// 只要索引，忽略值
for index := range numbers {
    fmt.Printf("索引: %d\n", index)
}

// 遍历字符串（按rune遍历，支持Unicode）
text := "Hello, 世界"
for index, char := range text {
    fmt.Printf("位置: %d, 字符: %c\n", index, char)
}

// 遍历map
userScores := map[string]int{
    "Alice": 95,
    "Bob":   87,
    "Carol": 92,
}

for name, score := range userScores {
    fmt.Printf("%s: %d分\n", name, score)
}

// 遍历channel
ch := make(chan int, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch)

for value := range ch {
    fmt.Printf("从channel接收: %d\n", value)
}
```

#### 4. 实际项目中的for循环应用

```go
// 来自 mall-go/pkg/cart/calculation_service.go
func (cs *CalculationService) CalculateCart(cart *model.Cart, userID uint, region string) (*CartCalculation, error) {
    calculation := &CartCalculation{
        SubtotalAmount: decimal.Zero,
        // ... 其他字段初始化
    }
    
    // 遍历购物车商品
    for _, item := range cart.Items {
        // 获取商品信息
        var product model.Product
        if err := cs.db.First(&product, item.ProductID).Error; err != nil {
            return nil, fmt.Errorf("商品不存在: %v", err)
        }
        
        // 计算商品小计
        itemTotal := product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
        calculation.SubtotalAmount = calculation.SubtotalAmount.Add(itemTotal)
        
        // 如果商品被选中，加入选中金额
        if item.Selected {
            calculation.SelectedAmount = calculation.SelectedAmount.Add(itemTotal)
        }
        
        // 计算重量
        if !product.Weight.IsZero() {
            itemWeight := product.Weight.Mul(decimal.NewFromInt(int64(item.Quantity)))
            calculation.TotalWeight = calculation.TotalWeight.Add(itemWeight)
        }
    }
    
    // 计算优惠券折扣
    if cart.CouponID != nil {
        for _, coupon := range cs.getAvailableCoupons(userID) {
            if coupon.ID == *cart.CouponID {
                discount := cs.calculateCouponDiscount(coupon, calculation.SelectedAmount)
                calculation.CouponDiscount = discount
                break  // 找到匹配的优惠券后退出循环
            }
        }
    }
    
    return calculation, nil
}
```

#### 5. 循环控制语句

```go
// break和continue
for i := 0; i < 10; i++ {
    if i == 3 {
        continue  // 跳过当前迭代
    }
    if i == 7 {
        break     // 退出循环
    }
    fmt.Println(i)
}

// 标签break和continue（用于嵌套循环）
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if i == 1 && j == 1 {
            break outer  // 跳出外层循环
        }
        fmt.Printf("i=%d, j=%d\n", i, j)
    }
}

// 在range循环中使用
numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
for _, num := range numbers {
    if num%2 == 0 {
        continue  // 跳过偶数
    }
    if num > 7 {
        break     // 大于7时退出
    }
    fmt.Printf("奇数: %d\n", num)
}
```

---

## 🔀 switch语句的强大功能

### Go的switch比Java/Python更强大

**Java switch (传统方式):**
```java
// Java - 只能用于整数、字符、字符串、枚举
int day = 3;
switch (day) {
    case 1:
        System.out.println("Monday");
        break;  // 必须有break，否则会fall through
    case 2:
        System.out.println("Tuesday");
        break;
    case 3:
        System.out.println("Wednesday");
        break;
    default:
        System.out.println("Other day");
}

// Java 14+ 新语法
String result = switch (day) {
    case 1 -> "Monday";
    case 2 -> "Tuesday";
    case 3 -> "Wednesday";
    default -> "Other day";
};
```

**Python (使用if-elif):**
```python
# Python 3.10之前没有switch，使用if-elif
day = 3
if day == 1:
    print("Monday")
elif day == 2:
    print("Tuesday")
elif day == 3:
    print("Wednesday")
else:
    print("Other day")

# Python 3.10+ match语句
match day:
    case 1:
        print("Monday")
    case 2:
        print("Tuesday")
    case 3:
        print("Wednesday")
    case _:
        print("Other day")
```

**Go switch (更强大):**
```go
// Go - 支持任何类型，自动break，无需显式break
day := 3
switch day {
case 1:
    fmt.Println("Monday")
case 2:
    fmt.Println("Tuesday")
case 3:
    fmt.Println("Wednesday")  // 自动break，不会继续执行
default:
    fmt.Println("Other day")
}

// 多个值匹配
switch day {
case 1, 2, 3, 4, 5:
    fmt.Println("工作日")
case 6, 7:
    fmt.Println("周末")
}

// 表达式switch
switch {
case day >= 1 && day <= 5:
    fmt.Println("工作日")
case day == 6 || day == 7:
    fmt.Println("周末")
default:
    fmt.Println("无效日期")
}
```

### Go switch的高级特性

#### 1. 类型switch

```go
// 类型断言switch
func processValue(v interface{}) {
    switch val := v.(type) {
    case int:
        fmt.Printf("整数: %d\n", val)
    case string:
        fmt.Printf("字符串: %s\n", val)
    case bool:
        fmt.Printf("布尔值: %t\n", val)
    case []int:
        fmt.Printf("整数切片: %v\n", val)
    case map[string]int:
        fmt.Printf("字符串到整数的映射: %v\n", val)
    case nil:
        fmt.Println("空值")
    default:
        fmt.Printf("未知类型: %T\n", val)
    }
}

// 使用示例
processValue(42)                    // 整数: 42
processValue("hello")               // 字符串: hello
processValue([]int{1, 2, 3})       // 整数切片: [1 2 3]
processValue(map[string]int{"a": 1}) // 字符串到整数的映射: map[a:1]
```

#### 2. 带初始化的switch

```go
// switch语句中的初始化
switch status := getStatus(); status {
case "active":
    fmt.Println("用户活跃")
case "inactive":
    fmt.Println("用户不活跃")
case "suspended":
    fmt.Println("用户被暂停")
default:
    fmt.Printf("未知状态: %s\n", status)
}
// status变量在这里不可见

// 错误处理中的应用
switch err := doSomething(); {
case err == nil:
    fmt.Println("操作成功")
case errors.Is(err, ErrNotFound):
    fmt.Println("资源不存在")
case errors.Is(err, ErrPermissionDenied):
    fmt.Println("权限不足")
default:
    fmt.Printf("其他错误: %v\n", err)
}
```

#### 3. fallthrough关键字

```go
// 显式的fall through（很少使用）
grade := 'B'
switch grade {
case 'A':
    fmt.Println("优秀")
    fallthrough  // 继续执行下一个case
case 'B':
    fmt.Println("良好")
    fallthrough
case 'C':
    fmt.Println("及格")
default:
    fmt.Println("需要努力")
}
// 输出：良好 及格
```

#### 4. 实际项目中的switch应用

```go
// 来自 mall-go/pkg/response/response.go
func (r *Response) SetCode(code int) *Response {
    r.Code = code
    
    // 根据状态码设置默认消息
    switch code {
    case CodeSuccess:
        r.Message = "操作成功"
    case CodeInvalidParam:
        r.Message = "参数错误"
    case CodeUnauthorized:
        r.Message = "未授权访问"
    case CodeForbidden:
        r.Message = "禁止访问"
    case CodeNotFound:
        r.Message = "资源不存在"
    case CodeConflict:
        r.Message = "资源冲突"
    case CodeTooManyReq:
        r.Message = "请求过多"
    case CodeError:
        r.Message = "服务器内部错误"
    default:
        r.Message = "未知错误"
    }
    
    return r
}

// 来自 mall-go/internal/model/order.go - 订单状态处理
func (o *Order) CanCancel() bool {
    switch o.Status {
    case OrderStatusPending:
        return true  // 待支付状态可以取消
    case OrderStatusPaid:
        // 已支付但未发货可以取消
        return o.ShippedAt == nil
    case OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled, OrderStatusRefunded:
        return false  // 这些状态不能取消
    default:
        return false
    }
}

func (o *Order) GetStatusText() string {
    switch o.Status {
    case OrderStatusPending:
        return "待支付"
    case OrderStatusPaid:
        return "已支付"
    case OrderStatusShipped:
        return "已发货"
    case OrderStatusDelivered:
        return "已送达"
    case OrderStatusCancelled:
        return "已取消"
    case OrderStatusRefunded:
        return "已退款"
    default:
        return "未知状态"
    }
}
```

---

## 🏷️ goto和标签控制

### goto的谨慎使用

虽然Go支持goto语句，但在现代编程中应该谨慎使用。不过在某些特定场景下，goto可以让代码更清晰。

```go
// goto的基本用法
func processData() error {
    // 初始化资源
    file, err := os.Open("data.txt")
    if err != nil {
        goto cleanup
    }
    
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        goto cleanup
    }
    
    // 处理数据...
    if someError := processFile(file); someError != nil {
        goto cleanup
    }
    
    // 正常完成
    file.Close()
    conn.Close()
    return nil
    
cleanup:
    // 清理资源
    if file != nil {
        file.Close()
    }
    if conn != nil {
        conn.Close()
    }
    return err
}

// 更好的方式：使用defer
func processDataBetter() error {
    file, err := os.Open("data.txt")
    if err != nil {
        return err
    }
    defer file.Close()  // 自动清理
    
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        return err
    }
    defer conn.Close()  // 自动清理
    
    // 处理数据...
    return processFile(file)
}
```

### 标签在循环中的应用

```go
// 跳出嵌套循环
func findElement(matrix [][]int, target int) (int, int) {
outer:
    for i, row := range matrix {
        for j, val := range row {
            if val == target {
                return i, j
            }
            if val > target {
                continue outer  // 跳到外层循环的下一次迭代
            }
        }
    }
    return -1, -1  // 未找到
}

// 错误处理中的goto应用
func complexOperation() error {
    step1 := func() error { /* ... */ return nil }
    step2 := func() error { /* ... */ return nil }
    step3 := func() error { /* ... */ return nil }
    
    if err := step1(); err != nil {
        goto handleError
    }
    
    if err := step2(); err != nil {
        goto handleError
    }
    
    if err := step3(); err != nil {
        goto handleError
    }
    
    return nil
    
handleError:
    // 统一的错误处理逻辑
    log.Printf("操作失败: %v", err)
    // 清理工作...
    return err
}
```

---

## ⏰ defer延迟执行机制

### defer的独特魅力

defer是Go语言的一个独特特性，用于延迟函数调用的执行，通常用于资源清理。

#### 1. defer的基本用法

```go
// 基本用法
func readFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()  // 函数返回前自动执行
    
    // 读取文件内容...
    data, err := ioutil.ReadAll(file)
    if err != nil {
        return err  // file.Close()会在return前执行
    }
    
    fmt.Printf("文件内容: %s\n", data)
    return nil  // file.Close()会在return前执行
}

// 多个defer的执行顺序（LIFO - 后进先出）
func deferOrder() {
    defer fmt.Println("第一个defer")
    defer fmt.Println("第二个defer")
    defer fmt.Println("第三个defer")
    fmt.Println("正常执行")
}
// 输出：
// 正常执行
// 第三个defer
// 第二个defer
// 第一个defer
```

#### 2. defer的高级用法

```go
// defer中的变量捕获
func deferCapture() {
    x := 1
    defer fmt.Printf("defer中的x: %d\n", x)  // 捕获当前值：1
    
    x = 2
    fmt.Printf("函数中的x: %d\n", x)  // 输出：2
}
// 输出：
// 函数中的x: 2
// defer中的x: 1

// defer函数的返回值修改
func deferReturn() (result int) {
    defer func() {
        result++  // 修改命名返回值
    }()
    return 5  // 实际返回6
}

// defer中的错误处理
func deferErrorHandling() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recovered: %v", r)
        }
    }()
    
    // 可能发生panic的代码
    panic("something went wrong")
}
```

#### 3. 实际项目中的defer应用

```go
// 来自 mall-go/pkg/database/database.go
func Transaction(fn func(*gorm.DB) error) error {
    if DB == nil {
        return fmt.Errorf("数据库未初始化")
    }
    
    tx := DB.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()  // panic时回滚事务
            panic(r)       // 重新抛出panic
        }
    }()
    
    if err := fn(tx); err != nil {
        tx.Rollback()  // 错误时回滚事务
        return err
    }
    
    return tx.Commit().Error  // 提交事务
}

// 来自 mall-go/pkg/logger/logger.go
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > time.Second {
            l.logger.Warn("慢操作检测", 
                zap.Duration("duration", duration),
                zap.Any("fields", fields))
        }
    }()
    
    // 执行日志操作...
    return l
}

// HTTP请求的defer应用
func makeHTTPRequest(url string) (*http.Response, error) {
    client := &http.Client{Timeout: 30 * time.Second}
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    
    // 注意：这里不能defer resp.Body.Close()
    // 因为调用者需要读取响应体
    // 应该由调用者负责关闭
    
    return resp, nil
}

// 正确的HTTP客户端使用方式
func fetchData(url string) ([]byte, error) {
    resp, err := makeHTTPRequest(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()  // 确保响应体被关闭
    
    return ioutil.ReadAll(resp.Body)
}
```

#### 4. defer的性能考虑

```go
// defer有轻微的性能开销
func withoutDefer() {
    // 直接调用，性能最好
    cleanup()
}

func withDefer() {
    defer cleanup()  // 有轻微开销，但通常可以忽略
}

// 在循环中使用defer要小心
func badDeferUsage() {
    for i := 0; i < 1000; i++ {
        file, _ := os.Open(fmt.Sprintf("file%d.txt", i))
        defer file.Close()  // ❌ 错误：所有文件会在函数结束时才关闭
    }
}

func goodDeferUsage() {
    for i := 0; i < 1000; i++ {
        func() {
            file, _ := os.Open(fmt.Sprintf("file%d.txt", i))
            defer file.Close()  // ✅ 正确：每次迭代结束时关闭文件
            // 处理文件...
        }()
    }
}
```

---

## 🚨 错误处理中的控制流

### Go的错误处理哲学

Go语言的错误处理与Java的异常机制完全不同，它鼓励显式的错误检查。

#### 1. 基本错误处理模式

```java
// Java - 异常处理
try {
    String content = readFile("config.txt");
    User user = parseUser(content);
    saveUser(user);
} catch (IOException e) {
    logger.error("文件操作失败", e);
} catch (ParseException e) {
    logger.error("解析失败", e);
} catch (DatabaseException e) {
    logger.error("数据库操作失败", e);
}
```

```go
// Go - 显式错误检查
func processUser() error {
    content, err := readFile("config.txt")
    if err != nil {
        return fmt.Errorf("读取文件失败: %v", err)
    }
    
    user, err := parseUser(content)
    if err != nil {
        return fmt.Errorf("解析用户失败: %v", err)
    }
    
    if err := saveUser(user); err != nil {
        return fmt.Errorf("保存用户失败: %v", err)
    }
    
    return nil
}
```

#### 2. 错误包装和链式处理

```go
import (
    "errors"
    "fmt"
)

// 自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("字段 %s 验证失败: %s", e.Field, e.Message)
}

// 错误包装
func validateUser(user *User) error {
    if user.Name == "" {
        return ValidationError{Field: "name", Message: "姓名不能为空"}
    }
    
    if user.Age < 0 {
        return ValidationError{Field: "age", Message: "年龄不能为负数"}
    }
    
    if !strings.Contains(user.Email, "@") {
        return ValidationError{Field: "email", Message: "邮箱格式不正确"}
    }
    
    return nil
}

// 错误处理的控制流
func createUser(userData map[string]interface{}) (*User, error) {
    user := &User{}
    
    // 数据转换
    if name, ok := userData["name"].(string); ok {
        user.Name = name
    } else {
        return nil, errors.New("姓名字段类型错误")
    }
    
    if age, ok := userData["age"].(float64); ok {
        user.Age = int(age)
    } else {
        return nil, errors.New("年龄字段类型错误")
    }
    
    if email, ok := userData["email"].(string); ok {
        user.Email = email
    } else {
        return nil, errors.New("邮箱字段类型错误")
    }
    
    // 验证用户数据
    if err := validateUser(user); err != nil {
        return nil, fmt.Errorf("用户数据验证失败: %w", err)  // 错误包装
    }
    
    return user, nil
}

// 错误处理的不同策略
func handleUserCreation(userData map[string]interface{}) {
    user, err := createUser(userData)
    if err != nil {
        // 类型断言检查特定错误
        var validationErr ValidationError
        if errors.As(err, &validationErr) {
            fmt.Printf("验证错误 - 字段: %s, 消息: %s\n", 
                validationErr.Field, validationErr.Message)
            return
        }
        
        // 检查是否是包装的错误
        if errors.Is(err, ValidationError{}) {
            fmt.Println("这是一个验证错误")
        }
        
        // 通用错误处理
        fmt.Printf("创建用户失败: %v\n", err)
        return
    }
    
    fmt.Printf("用户创建成功: %+v\n", user)
}
```

#### 3. 实际项目中的错误处理

```go
// 来自 mall-go/pkg/user/service.go
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // 参数验证
    if req.Username == "" {
        return nil, errors.New("用户名不能为空")
    }
    
    if len(req.Password) < 6 {
        return nil, errors.New("密码长度不能少于6位")
    }
    
    // 检查用户名是否已存在
    var existingUser User
    err := s.db.Where("username = ?", req.Username).First(&existingUser).Error
    if err == nil {
        return nil, errors.New("用户名已存在")
    } else if !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, fmt.Errorf("查询用户失败: %v", err)
    }
    
    // 密码加密
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("密码加密失败: %v", err)
    }
    
    // 创建用户
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Status:   "active",
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return nil, fmt.Errorf("创建用户失败: %v", err)
    }
    
    return user, nil
}

// 错误处理的中间件模式
func ErrorHandlingMiddleware() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        if err, ok := recovered.(string); ok {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": fmt.Sprintf("服务器内部错误: %s", err),
            })
        }
        c.AbortWithStatus(http.StatusInternalServerError)
    })
}

// 统一的错误响应处理
func HandleError(c *gin.Context, err error) {
    var validationErr ValidationError
    if errors.As(err, &validationErr) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "参数验证失败",
            "field":   validationErr.Field,
            "message": validationErr.Message,
        })
        return
    }
    
    if errors.Is(err, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "资源不存在",
        })
        return
    }
    
    // 默认错误处理
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "服务器内部错误",
    })
}
```

---

## 💼 实战案例分析

### 案例1: 用户登录验证流程

让我们通过一个完整的用户登录验证流程来看看Go控制结构的实际应用：

```go
// 来自 mall-go/internal/handler/user/auth.go
func (h *UserHandler) Login(c *gin.Context) {
    var req LoginRequest

    // 参数绑定和验证
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "参数格式错误",
            "details": err.Error(),
        })
        return
    }

    // 输入验证
    switch {
    case req.Username == "":
        c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
        return
    case req.Password == "":
        c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能为空"})
        return
    case len(req.Password) < 6:
        c.JSON(http.StatusBadRequest, gin.H{"error": "密码长度不能少于6位"})
        return
    }

    // 查找用户
    var user model.User
    if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
        }
        return
    }

    // 检查用户状态
    switch user.Status {
    case "inactive":
        c.JSON(http.StatusForbidden, gin.H{"error": "账户未激活"})
        return
    case "suspended":
        c.JSON(http.StatusForbidden, gin.H{"error": "账户已被暂停"})
        return
    case "deleted":
        c.JSON(http.StatusForbidden, gin.H{"error": "账户已被删除"})
        return
    }

    // 验证密码
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        // 记录失败尝试
        go func() {
            h.recordLoginAttempt(user.ID, false, c.ClientIP())
        }()

        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
        return
    }

    // 生成JWT令牌
    token, err := h.generateJWT(user.ID, user.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
        return
    }

    // 更新登录信息
    go func() {
        h.updateLoginInfo(user.ID, c.ClientIP())
        h.recordLoginAttempt(user.ID, true, c.ClientIP())
    }()

    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "message": "登录成功",
        "token":   token,
        "user": gin.H{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
        },
    })
}

// 辅助函数：记录登录尝试
func (h *UserHandler) recordLoginAttempt(userID uint, success bool, ip string) {
    attempt := &model.LoginAttempt{
        UserID:    userID,
        Success:   success,
        IP:        ip,
        UserAgent: "", // 可以从context获取
        CreatedAt: time.Now(),
    }

    if err := h.db.Create(attempt).Error; err != nil {
        log.Printf("记录登录尝试失败: %v", err)
    }

    // 如果登录失败，检查是否需要锁定账户
    if !success {
        h.checkAccountLockout(userID, ip)
    }
}

// 检查账户锁定
func (h *UserHandler) checkAccountLockout(userID uint, ip string) {
    // 查询最近15分钟的失败尝试次数
    var failedCount int64
    h.db.Model(&model.LoginAttempt{}).
        Where("user_id = ? AND success = false AND created_at > ?",
              userID, time.Now().Add(-15*time.Minute)).
        Count(&failedCount)

    // 如果失败次数超过5次，锁定账户30分钟
    if failedCount >= 5 {
        h.db.Model(&model.User{}).
            Where("id = ?", userID).
            Update("locked_until", time.Now().Add(30*time.Minute))

        log.Printf("用户 %d 因多次登录失败被锁定30分钟", userID)
    }
}
```

### 案例2: 订单状态机处理

```go
// 来自 mall-go/pkg/order/state_machine.go
type OrderStateMachine struct {
    db *gorm.DB
}

func (osm *OrderStateMachine) TransitionState(orderID uint, newStatus string, reason string) error {
    var order model.Order
    if err := osm.db.First(&order, orderID).Error; err != nil {
        return fmt.Errorf("订单不存在: %v", err)
    }

    // 验证状态转换是否合法
    if !osm.isValidTransition(order.Status, newStatus) {
        return fmt.Errorf("不能从状态 %s 转换到 %s", order.Status, newStatus)
    }

    // 开始事务
    tx := osm.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    // 根据新状态执行相应的业务逻辑
    switch newStatus {
    case "paid":
        if err := osm.handlePaidTransition(tx, &order); err != nil {
            tx.Rollback()
            return err
        }

    case "shipped":
        if err := osm.handleShippedTransition(tx, &order); err != nil {
            tx.Rollback()
            return err
        }

    case "delivered":
        if err := osm.handleDeliveredTransition(tx, &order); err != nil {
            tx.Rollback()
            return err
        }

    case "cancelled":
        if err := osm.handleCancelledTransition(tx, &order, reason); err != nil {
            tx.Rollback()
            return err
        }

    case "refunded":
        if err := osm.handleRefundedTransition(tx, &order); err != nil {
            tx.Rollback()
            return err
        }
    }

    // 更新订单状态
    now := time.Now()
    updates := map[string]interface{}{
        "status":     newStatus,
        "updated_at": now,
    }

    // 根据状态设置特定的时间戳
    switch newStatus {
    case "paid":
        updates["paid_at"] = &now
    case "shipped":
        updates["shipped_at"] = &now
    case "delivered":
        updates["delivered_at"] = &now
    case "cancelled":
        updates["cancelled_at"] = &now
    case "refunded":
        updates["refunded_at"] = &now
    }

    if err := tx.Model(&order).Updates(updates).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("更新订单状态失败: %v", err)
    }

    // 记录状态变更历史
    history := &model.OrderStatusHistory{
        OrderID:   orderID,
        FromStatus: order.Status,
        ToStatus:   newStatus,
        Reason:     reason,
        CreatedAt:  now,
    }

    if err := tx.Create(history).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("记录状态历史失败: %v", err)
    }

    // 提交事务
    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("提交事务失败: %v", err)
    }

    // 异步发送通知
    go osm.sendStatusChangeNotification(orderID, order.Status, newStatus)

    return nil
}

// 验证状态转换是否合法
func (osm *OrderStateMachine) isValidTransition(from, to string) bool {
    validTransitions := map[string][]string{
        "pending":   {"paid", "cancelled"},
        "paid":      {"shipped", "cancelled", "refunded"},
        "shipped":   {"delivered", "cancelled"},
        "delivered": {"refunded"},
        "cancelled": {},  // 取消状态不能转换到其他状态
        "refunded":  {},  // 退款状态不能转换到其他状态
    }

    allowedStates, exists := validTransitions[from]
    if !exists {
        return false
    }

    for _, allowedState := range allowedStates {
        if allowedState == to {
            return true
        }
    }

    return false
}

// 处理支付完成状态转换
func (osm *OrderStateMachine) handlePaidTransition(tx *gorm.DB, order *model.Order) error {
    // 扣减库存
    var orderItems []model.OrderItem
    if err := tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
        return fmt.Errorf("查询订单项失败: %v", err)
    }

    for _, item := range orderItems {
        // 使用乐观锁扣减库存
        result := tx.Model(&model.Product{}).
            Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
            Update("stock", gorm.Expr("stock - ?", item.Quantity))

        if result.Error != nil {
            return fmt.Errorf("扣减库存失败: %v", result.Error)
        }

        if result.RowsAffected == 0 {
            return fmt.Errorf("商品 %d 库存不足", item.ProductID)
        }
    }

    return nil
}

// 处理发货状态转换
func (osm *OrderStateMachine) handleShippedTransition(tx *gorm.DB, order *model.Order) error {
    // 生成物流单号
    trackingNumber := osm.generateTrackingNumber()

    // 创建物流记录
    shipping := &model.OrderShipping{
        OrderID:        order.ID,
        TrackingNumber: trackingNumber,
        Carrier:        "默认快递",
        Status:         "shipped",
        CreatedAt:      time.Now(),
    }

    if err := tx.Create(shipping).Error; err != nil {
        return fmt.Errorf("创建物流记录失败: %v", err)
    }

    return nil
}

// 处理取消状态转换
func (osm *OrderStateMachine) handleCancelledTransition(tx *gorm.DB, order *model.Order, reason string) error {
    // 如果已支付，需要退款
    if order.Status == "paid" {
        refund := &model.OrderRefund{
            OrderID:     order.ID,
            Amount:      order.TotalAmount,
            Reason:      reason,
            Status:      "processing",
            CreatedAt:   time.Now(),
        }

        if err := tx.Create(refund).Error; err != nil {
            return fmt.Errorf("创建退款记录失败: %v", err)
        }

        // 异步处理退款
        go osm.processRefund(refund.ID)
    }

    // 恢复库存
    var orderItems []model.OrderItem
    if err := tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
        return fmt.Errorf("查询订单项失败: %v", err)
    }

    for _, item := range orderItems {
        if err := tx.Model(&model.Product{}).
            Where("id = ?", item.ProductID).
            Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
            return fmt.Errorf("恢复库存失败: %v", err)
        }
    }

    return nil
}
```

### 案例3: 批量数据处理

```go
// 来自 mall-go/pkg/batch/processor.go
type BatchProcessor struct {
    db        *gorm.DB
    batchSize int
    workers   int
}

func (bp *BatchProcessor) ProcessUsers(userIDs []uint, operation string) error {
    if len(userIDs) == 0 {
        return nil
    }

    // 创建工作通道
    jobs := make(chan []uint, bp.workers)
    results := make(chan error, bp.workers)

    // 启动工作协程
    for i := 0; i < bp.workers; i++ {
        go bp.worker(jobs, results, operation)
    }

    // 分批发送任务
    go func() {
        defer close(jobs)

        for i := 0; i < len(userIDs); i += bp.batchSize {
            end := i + bp.batchSize
            if end > len(userIDs) {
                end = len(userIDs)
            }

            batch := userIDs[i:end]
            jobs <- batch
        }
    }()

    // 收集结果
    batches := (len(userIDs) + bp.batchSize - 1) / bp.batchSize
    var errors []error

    for i := 0; i < batches; i++ {
        if err := <-results; err != nil {
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("批量处理失败，错误数量: %d", len(errors))
    }

    return nil
}

func (bp *BatchProcessor) worker(jobs <-chan []uint, results chan<- error, operation string) {
    for batch := range jobs {
        err := bp.processBatch(batch, operation)
        results <- err
    }
}

func (bp *BatchProcessor) processBatch(userIDs []uint, operation string) error {
    tx := bp.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    switch operation {
    case "activate":
        err := bp.activateUsers(tx, userIDs)
        if err != nil {
            tx.Rollback()
            return err
        }

    case "deactivate":
        err := bp.deactivateUsers(tx, userIDs)
        if err != nil {
            tx.Rollback()
            return err
        }

    case "delete":
        err := bp.deleteUsers(tx, userIDs)
        if err != nil {
            tx.Rollback()
            return err
        }

    default:
        tx.Rollback()
        return fmt.Errorf("不支持的操作: %s", operation)
    }

    return tx.Commit().Error
}

func (bp *BatchProcessor) activateUsers(tx *gorm.DB, userIDs []uint) error {
    // 批量更新用户状态
    result := tx.Model(&model.User{}).
        Where("id IN ?", userIDs).
        Updates(map[string]interface{}{
            "status":     "active",
            "updated_at": time.Now(),
        })

    if result.Error != nil {
        return fmt.Errorf("激活用户失败: %v", result.Error)
    }

    // 记录操作日志
    for _, userID := range userIDs {
        log := &model.UserOperationLog{
            UserID:    userID,
            Operation: "activate",
            CreatedAt: time.Now(),
        }

        if err := tx.Create(log).Error; err != nil {
            return fmt.Errorf("记录操作日志失败: %v", err)
        }
    }

    return nil
}
```

---

## 🎯 面试常考点

### 1. Go语言的控制结构特点

**面试题**: "Go语言的控制结构有什么特点？与Java有什么区别？"

**标准答案**:
```go
// Go的特点：
// 1. if语句无需括号，但必须有大括号
if x > 0 {  // ✅ 正确
    fmt.Println("正数")
}

// 2. 支持初始化语句
if err := doSomething(); err != nil {
    return err
}

// 3. 只有for循环，没有while
for i := 0; i < 10; i++ {  // 传统for
    fmt.Println(i)
}

for condition {  // while循环的实现
    // ...
}

for {  // 无限循环
    // ...
}

// 4. switch自动break，支持多值匹配
switch day {
case 1, 2, 3, 4, 5:
    fmt.Println("工作日")
case 6, 7:
    fmt.Println("周末")
}
```

### 2. defer的执行顺序和用途

**面试题**: "解释defer的执行顺序和主要用途"

**标准答案**:
```go
// defer执行顺序：LIFO（后进先出）
func deferOrder() {
    defer fmt.Println("1")  // 最后执行
    defer fmt.Println("2")  // 第二执行
    defer fmt.Println("3")  // 最先执行
    fmt.Println("正常执行")
}
// 输出：正常执行 3 2 1

// 主要用途：
// 1. 资源清理
func readFile() error {
    file, err := os.Open("test.txt")
    if err != nil {
        return err
    }
    defer file.Close()  // 确保文件关闭

    // 处理文件...
    return nil
}

// 2. 事务处理
func transaction() error {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 业务逻辑...
    return tx.Commit().Error
}
```

### 3. range循环的特性

**面试题**: "range循环有什么需要注意的地方？"

**标准答案**:
```go
// 1. range会复制值
slice := []int{1, 2, 3}
for _, v := range slice {
    v = v * 2  // 不会修改原slice
}
fmt.Println(slice)  // [1, 2, 3]

// 正确的修改方式
for i := range slice {
    slice[i] = slice[i] * 2
}

// 2. range遍历map的顺序是随机的
m := map[string]int{"a": 1, "b": 2, "c": 3}
for k, v := range m {
    fmt.Printf("%s: %d\n", k, v)  // 顺序不确定
}

// 3. range遍历字符串按rune遍历
s := "Hello, 世界"
for i, r := range s {
    fmt.Printf("位置: %d, 字符: %c\n", i, r)
}
```

### 4. 错误处理最佳实践

**面试题**: "Go语言的错误处理有什么最佳实践？"

**标准答案**:
```go
// 1. 及早返回
func processData() error {
    data, err := readData()
    if err != nil {
        return fmt.Errorf("读取数据失败: %w", err)  // 错误包装
    }

    if err := validateData(data); err != nil {
        return fmt.Errorf("数据验证失败: %w", err)
    }

    return saveData(data)
}

// 2. 错误类型检查
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    // 处理特定类型的错误
}

if errors.Is(err, ErrNotFound) {
    // 处理特定的错误值
}

// 3. 自定义错误类型
type APIError struct {
    Code    int
    Message string
}

func (e APIError) Error() string {
    return fmt.Sprintf("API错误 %d: %s", e.Code, e.Message)
}
```

### 5. switch语句的高级用法

**面试题**: "Go的switch语句有哪些高级用法？"

**标准答案**:
```go
// 1. 类型switch
func processValue(v interface{}) {
    switch val := v.(type) {
    case int:
        fmt.Printf("整数: %d\n", val)
    case string:
        fmt.Printf("字符串: %s\n", val)
    case []int:
        fmt.Printf("整数切片: %v\n", val)
    default:
        fmt.Printf("未知类型: %T\n", val)
    }
}

// 2. 表达式switch
func getGrade(score int) string {
    switch {
    case score >= 90:
        return "A"
    case score >= 80:
        return "B"
    case score >= 70:
        return "C"
    case score >= 60:
        return "D"
    default:
        return "F"
    }
}

// 3. 带初始化的switch
switch status := getStatus(); status {
case "active":
    return "用户活跃"
case "inactive":
    return "用户不活跃"
default:
    return "未知状态"
}
```

---

## 💡 踩坑提醒

### 1. if语句的常见错误

```go
// ❌ 错误：大括号换行
if condition
{  // 编译错误！
    // ...
}

// ❌ 错误：缺少大括号
if condition
    fmt.Println("hello")  // 编译错误！

// ❌ 错误：在if外使用初始化的变量
if err := doSomething(); err != nil {
    return err
}
fmt.Println(err)  // 编译错误！err不在作用域内

// ✅ 正确：扩大变量作用域
var err error
if err = doSomething(); err != nil {
    return err
}
fmt.Println(err)  // 正确
```

### 2. for循环的陷阱

```go
// ❌ 错误：闭包中的循环变量
var funcs []func()
for i := 0; i < 3; i++ {
    funcs = append(funcs, func() {
        fmt.Println(i)  // 都会打印3！
    })
}

// ✅ 正确：捕获循环变量
var funcs []func()
for i := 0; i < 3; i++ {
    i := i  // 创建新变量
    funcs = append(funcs, func() {
        fmt.Println(i)  // 正确打印0, 1, 2
    })
}

// ❌ 错误：修改range的值
slice := []int{1, 2, 3}
for _, v := range slice {
    v = v * 2  // 不会修改原slice
}

// ✅ 正确：使用索引修改
for i := range slice {
    slice[i] = slice[i] * 2
}
```

### 3. defer的陷阱

```go
// ❌ 错误：defer在循环中
func badDefer() {
    for i := 0; i < 5; i++ {
        file, _ := os.Open(fmt.Sprintf("file%d.txt", i))
        defer file.Close()  // 所有文件在函数结束时才关闭！
    }
}

// ✅ 正确：使用匿名函数
func goodDefer() {
    for i := 0; i < 5; i++ {
        func() {
            file, _ := os.Open(fmt.Sprintf("file%d.txt", i))
            defer file.Close()  // 每次迭代结束时关闭
            // 处理文件...
        }()
    }
}

// ❌ 错误：defer参数的求值时机
func deferTrap() {
    x := 1
    defer fmt.Println(x)  // 捕获当前值1
    x = 2
    // 输出：1（不是2）
}

// ✅ 正确：使用闭包
func deferCorrect() {
    x := 1
    defer func() {
        fmt.Println(x)  // 使用最新值
    }()
    x = 2
    // 输出：2
}
```

### 4. switch的陷阱

```go
// ❌ 错误：期望fall through但没有使用fallthrough
func badSwitch(x int) {
    switch x {
    case 1:
        fmt.Println("one")
    case 2:
        fmt.Println("two")  // 不会继续执行case 3
    case 3:
        fmt.Println("three")
    }
}

// ✅ 正确：使用fallthrough或多值匹配
func goodSwitch(x int) {
    switch x {
    case 1:
        fmt.Println("one")
        fallthrough  // 显式继续
    case 2, 3:  // 或者使用多值匹配
        fmt.Println("two or three")
    }
}
```

---

## 📝 本章练习题

### 基础练习

1. **条件语句练习**
```go
// 编写一个函数，根据分数返回等级
// 90-100: A, 80-89: B, 70-79: C, 60-69: D, <60: F
// 要求使用Go的if语句特性

func getGrade(score int) (string, error) {
    // 参考答案：
    if score < 0 || score > 100 {
        return "", fmt.Errorf("分数必须在0-100之间")
    }

    switch {
    case score >= 90:
        return "A", nil
    case score >= 80:
        return "B", nil
    case score >= 70:
        return "C", nil
    case score >= 60:
        return "D", nil
    default:
        return "F", nil
    }
}

// 使用if语句的版本
func getGradeWithIf(score int) (string, error) {
    if score < 0 || score > 100 {
        return "", fmt.Errorf("分数必须在0-100之间")
    }

    if grade := func() string {
        if score >= 90 {
            return "A"
        } else if score >= 80 {
            return "B"
        } else if score >= 70 {
            return "C"
        } else if score >= 60 {
            return "D"
        }
        return "F"
    }(); grade != "" {
        return grade, nil
    }

    return "F", nil
}
```

2. **循环练习**
```go
// 编写函数实现以下功能：
// 1. 计算1到n的和
// 2. 找出数组中的最大值和最小值
// 3. 统计字符串中每个字符的出现次数

// 参考答案：
func sumToN(n int) int {
    sum := 0
    for i := 1; i <= n; i++ {
        sum += i
    }
    return sum
}

func findMinMax(numbers []int) (min, max int, err error) {
    if len(numbers) == 0 {
        return 0, 0, fmt.Errorf("数组不能为空")
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

func countChars(s string) map[rune]int {
    counts := make(map[rune]int)
    for _, char := range s {
        counts[char]++
    }
    return counts
}
```

### 进阶练习

3. **错误处理练习**
```go
// 设计一个文件处理函数，要求：
// 1. 读取文件内容
// 2. 验证内容格式
// 3. 处理各种可能的错误
// 4. 使用defer确保资源清理

type FileProcessor struct {
    maxSize int64
}

func (fp *FileProcessor) ProcessFile(filename string) ([]byte, error) {
    // 检查文件是否存在
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return nil, fmt.Errorf("文件不存在: %s", filename)
    }

    // 打开文件
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("打开文件失败: %w", err)
    }
    defer file.Close()  // 确保文件关闭

    // 检查文件大小
    if info, err := file.Stat(); err != nil {
        return nil, fmt.Errorf("获取文件信息失败: %w", err)
    } else if info.Size() > fp.maxSize {
        return nil, fmt.Errorf("文件太大: %d bytes, 最大允许: %d bytes",
            info.Size(), fp.maxSize)
    }

    // 读取文件内容
    content, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("读取文件失败: %w", err)
    }

    // 验证内容格式（假设是JSON）
    if !json.Valid(content) {
        return nil, fmt.Errorf("文件内容不是有效的JSON格式")
    }

    return content, nil
}

// 使用示例
func main() {
    processor := &FileProcessor{maxSize: 1024 * 1024} // 1MB

    content, err := processor.ProcessFile("config.json")
    if err != nil {
        log.Printf("处理文件失败: %v", err)
        return
    }

    fmt.Printf("文件内容: %s\n", content)
}
```

4. **状态机练习**
```go
// 实现一个简单的状态机，模拟订单处理流程
// 状态：pending -> paid -> shipped -> delivered
// 或者：pending -> cancelled

type OrderState string

const (
    StatePending   OrderState = "pending"
    StatePaid      OrderState = "paid"
    StateShipped   OrderState = "shipped"
    StateDelivered OrderState = "delivered"
    StateCancelled OrderState = "cancelled"
)

type Order struct {
    ID     string
    State  OrderState
    Amount decimal.Decimal
}

type OrderStateMachine struct {
    validTransitions map[OrderState][]OrderState
}

func NewOrderStateMachine() *OrderStateMachine {
    return &OrderStateMachine{
        validTransitions: map[OrderState][]OrderState{
            StatePending:   {StatePaid, StateCancelled},
            StatePaid:      {StateShipped, StateCancelled},
            StateShipped:   {StateDelivered},
            StateDelivered: {},
            StateCancelled: {},
        },
    }
}

func (osm *OrderStateMachine) CanTransition(from, to OrderState) bool {
    allowedStates, exists := osm.validTransitions[from]
    if !exists {
        return false
    }

    for _, state := range allowedStates {
        if state == to {
            return true
        }
    }
    return false
}

func (osm *OrderStateMachine) Transition(order *Order, newState OrderState) error {
    if !osm.CanTransition(order.State, newState) {
        return fmt.Errorf("不能从状态 %s 转换到 %s", order.State, newState)
    }

    oldState := order.State

    // 执行状态转换的业务逻辑
    switch newState {
    case StatePaid:
        if err := osm.handlePayment(order); err != nil {
            return fmt.Errorf("处理支付失败: %w", err)
        }

    case StateShipped:
        if err := osm.handleShipping(order); err != nil {
            return fmt.Errorf("处理发货失败: %w", err)
        }

    case StateDelivered:
        if err := osm.handleDelivery(order); err != nil {
            return fmt.Errorf("处理送达失败: %w", err)
        }

    case StateCancelled:
        if err := osm.handleCancellation(order); err != nil {
            return fmt.Errorf("处理取消失败: %w", err)
        }
    }

    order.State = newState
    log.Printf("订单 %s 状态从 %s 转换到 %s", order.ID, oldState, newState)

    return nil
}

func (osm *OrderStateMachine) handlePayment(order *Order) error {
    // 模拟支付处理
    log.Printf("处理订单 %s 的支付，金额: %s", order.ID, order.Amount.String())
    return nil
}

func (osm *OrderStateMachine) handleShipping(order *Order) error {
    // 模拟发货处理
    log.Printf("订单 %s 开始发货", order.ID)
    return nil
}

func (osm *OrderStateMachine) handleDelivery(order *Order) error {
    // 模拟送达处理
    log.Printf("订单 %s 已送达", order.ID)
    return nil
}

func (osm *OrderStateMachine) handleCancellation(order *Order) error {
    // 模拟取消处理
    log.Printf("订单 %s 已取消", order.ID)
    return nil
}
```

### 高级练习

5. **并发控制练习**
```go
// 实现一个工作池，用于并发处理任务
// 要求：
// 1. 限制并发数量
// 2. 优雅关闭
// 3. 错误处理
// 4. 进度监控

type Task struct {
    ID   int
    Data interface{}
}

type Result struct {
    TaskID int
    Data   interface{}
    Error  error
}

type WorkerPool struct {
    workerCount int
    taskChan    chan Task
    resultChan  chan Result
    doneChan    chan struct{}
    wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int) *WorkerPool {
    return &WorkerPool{
        workerCount: workerCount,
        taskChan:    make(chan Task, workerCount*2),
        resultChan:  make(chan Result, workerCount*2),
        doneChan:    make(chan struct{}),
    }
}

func (wp *WorkerPool) Start(processor func(Task) (interface{}, error)) {
    // 启动工作协程
    for i := 0; i < wp.workerCount; i++ {
        wp.wg.Add(1)
        go wp.worker(i, processor)
    }

    // 启动结果收集协程
    go wp.resultCollector()
}

func (wp *WorkerPool) worker(id int, processor func(Task) (interface{}, error)) {
    defer wp.wg.Done()

    for {
        select {
        case task, ok := <-wp.taskChan:
            if !ok {
                log.Printf("Worker %d 停止工作", id)
                return
            }

            log.Printf("Worker %d 处理任务 %d", id, task.ID)
            data, err := processor(task)

            result := Result{
                TaskID: task.ID,
                Data:   data,
                Error:  err,
            }

            select {
            case wp.resultChan <- result:
            case <-wp.doneChan:
                return
            }

        case <-wp.doneChan:
            log.Printf("Worker %d 收到停止信号", id)
            return
        }
    }
}

func (wp *WorkerPool) resultCollector() {
    for {
        select {
        case result := <-wp.resultChan:
            if result.Error != nil {
                log.Printf("任务 %d 处理失败: %v", result.TaskID, result.Error)
            } else {
                log.Printf("任务 %d 处理成功: %v", result.TaskID, result.Data)
            }

        case <-wp.doneChan:
            log.Println("结果收集器停止")
            return
        }
    }
}

func (wp *WorkerPool) Submit(task Task) {
    select {
    case wp.taskChan <- task:
    case <-wp.doneChan:
        log.Printf("工作池已关闭，任务 %d 被丢弃", task.ID)
    }
}

func (wp *WorkerPool) Stop() {
    log.Println("开始关闭工作池...")

    // 关闭任务通道
    close(wp.taskChan)

    // 等待所有工作协程完成
    wp.wg.Wait()

    // 发送停止信号
    close(wp.doneChan)

    log.Println("工作池已关闭")
}

// 使用示例
func main() {
    pool := NewWorkerPool(3)

    // 定义任务处理器
    processor := func(task Task) (interface{}, error) {
        // 模拟处理时间
        time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

        // 模拟随机错误
        if rand.Float32() < 0.1 {
            return nil, fmt.Errorf("随机错误")
        }

        return fmt.Sprintf("处理结果: %v", task.Data), nil
    }

    // 启动工作池
    pool.Start(processor)

    // 提交任务
    for i := 0; i < 10; i++ {
        task := Task{
            ID:   i,
            Data: fmt.Sprintf("任务数据 %d", i),
        }
        pool.Submit(task)
    }

    // 等待一段时间让任务完成
    time.Sleep(5 * time.Second)

    // 关闭工作池
    pool.Stop()
}
```

---

## 🎉 本章总结

通过本章学习，你应该掌握了：

### ✅ 核心概念
- [x] Go语言控制结构的语法特点和与Java/Python的差异
- [x] if语句的初始化语法和作用域控制
- [x] for循环的多种形态和range的强大功能
- [x] switch语句的高级特性和类型switch
- [x] defer延迟执行机制的原理和应用场景
- [x] Go语言错误处理的哲学和最佳实践

### ✅ 实际应用
- [x] 用户登录验证流程的完整实现
- [x] 订单状态机的设计和实现
- [x] 批量数据处理的并发控制
- [x] 文件处理中的资源管理
- [x] 工作池模式的实现

### ✅ 最佳实践
- [x] 控制结构的正确使用方式
- [x] 错误处理的标准模式
- [x] defer的使用技巧和陷阱避免
- [x] 并发控制的设计原则

### 🚀 下一步学习

恭喜完成基础篇第二章！接下来我们将学习：
- **[函数定义与方法](./03-functions-and-methods.md)** - Go函数的强大特性
- **[包管理与模块系统](./04-packages-and-imports.md)** - Go的模块化设计

---

> 💡 **学习提示**:
> 1. 多练习控制结构的组合使用
> 2. 重点理解defer的执行时机和用途
> 3. 掌握Go的错误处理模式，这是面试重点
> 4. 结合实际项目理解控制流的设计思想

**继续加油！Go语言的控制结构正在让你的代码更加简洁和高效！** 🎯
```
```
```
```
