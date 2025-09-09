# 进阶篇第四章：接口设计模式 🎨

> **"接口是Go语言的灵魂，设计模式是编程的艺术"** 💫

---

## 📖 章节导读

欢迎来到Go语言接口设计模式的世界！🌟 如果说并发是Go的心脏，那么接口就是Go的灵魂。在这一章中，我们将深入探索Go语言接口的设计哲学，学习如何用Go的方式实现经典设计模式，掌握企业级项目中的接口设计技巧。

### 🎯 学习目标

通过本章学习，你将掌握：

- **🏗️ Go接口设计哲学**：理解"接受接口，返回结构体"的设计原则
- **🎨 经典设计模式**：用Go语言实现工厂、策略、装饰器等模式
- **💉 依赖注入**：掌握Go语言中的依赖注入和控制反转
- **🔗 接口组合**：学会接口嵌入和组合的高级用法
- **🎭 类型系统**：深入理解空接口、类型断言和类型开关
- **🧪 测试技巧**：掌握接口测试和Mock的最佳实践
- **⚡ 性能优化**：了解接口使用中的性能考虑

### 📋 章节大纲

```
04-interface-design-patterns.md
├── 🏛️  Go接口设计哲学
├── 🏭  经典设计模式实现
│   ├── 工厂模式家族
│   ├── 策略模式
│   ├── 装饰器模式
│   ├── 适配器模式
│   ├── 观察者模式
│   └── 命令模式
├── 💉  依赖注入和控制反转
├── 🔗  接口组合和嵌入
├── 🎭  空接口和类型断言
├── 🧪  接口测试和Mock
├── 🏢  实战案例分析
├── 🎯  面试常考点
├── ⚠️   踩坑提醒
├── 📝  练习题
└── 📚  章节总结
```

---

## 🏛️ Go接口设计哲学

### Go接口的独特之处

Go语言的接口设计有着独特的哲学，与传统的面向对象语言有很大不同。

```go
// 来自 mall-go/pkg/design/philosophy.go
package design

import (
    "fmt"
    "io"
    "strings"
)

/*
Go接口设计的核心原则：

1. 接口隔离原则 (Interface Segregation Principle)
   - 接口应该小而专一
   - 客户端不应该依赖它不需要的接口

2. "接受接口，返回结构体" (Accept interfaces, return structs)
   - 函数参数使用接口类型，增加灵活性
   - 函数返回值使用具体类型，提供明确的API

3. 隐式实现 (Implicit Implementation)
   - 无需显式声明实现接口
   - 只要实现了接口的所有方法就自动满足接口

4. 组合优于继承 (Composition over Inheritance)
   - 通过接口组合实现复杂功能
   - 避免深层次的继承关系
*/

// 示例1：小接口设计
// ❌ 错误示例：接口过于庞大
type BadUserService interface {
    CreateUser(user User) error
    UpdateUser(user User) error
    DeleteUser(id int64) error
    GetUser(id int64) (*User, error)
    ListUsers(page, size int) ([]*User, error)
    AuthenticateUser(username, password string) (*User, error)
    ChangePassword(userID int64, oldPassword, newPassword string) error
    SendEmail(userID int64, subject, body string) error
    UploadAvatar(userID int64, avatar []byte) error
    GetUserStatistics(userID int64) (*UserStats, error)
}

// ✅ 正确示例：接口职责分离
type UserRepository interface {
    Create(user User) error
    Update(user User) error
    Delete(id int64) error
    GetByID(id int64) (*User, error)
    List(page, size int) ([]*User, error)
}

type UserAuthenticator interface {
    Authenticate(username, password string) (*User, error)
    ChangePassword(userID int64, oldPassword, newPassword string) error
}

type UserNotifier interface {
    SendEmail(userID int64, subject, body string) error
}

type UserAvatarManager interface {
    UploadAvatar(userID int64, avatar []byte) error
}

// 示例2："接受接口，返回结构体"原则
type User struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Avatar   string `json:"avatar"`
}

type UserStats struct {
    LoginCount    int64 `json:"login_count"`
    LastLoginTime int64 `json:"last_login_time"`
}

// ✅ 正确示例：接受接口参数
func ProcessUserData(reader io.Reader, writer io.Writer) error {
    // 接受接口，提供最大的灵活性
    data, err := io.ReadAll(reader)
    if err != nil {
        return err
    }
    
    // 处理数据
    processed := strings.ToUpper(string(data))
    
    _, err = writer.Write([]byte(processed))
    return err
}

// ✅ 正确示例：返回具体结构体
func NewUserService(repo UserRepository, auth UserAuthenticator) *UserService {
    // 返回具体类型，提供明确的API
    return &UserService{
        repo: repo,
        auth: auth,
    }
}

type UserService struct {
    repo UserRepository
    auth UserAuthenticator
}

func (s *UserService) CreateUser(user User) error {
    // 实现用户创建逻辑
    return s.repo.Create(user)
}

func (s *UserService) AuthenticateUser(username, password string) (*User, error) {
    return s.auth.Authenticate(username, password)
}
```

### 与Java/Python接口设计对比

```java
// Java接口设计
public interface UserService {
    // Java需要显式声明实现接口
    void createUser(User user) throws UserException;
    User getUser(Long id) throws UserException;
}

// 显式实现接口
public class UserServiceImpl implements UserService {
    @Override
    public void createUser(User user) throws UserException {
        // 实现逻辑
    }
    
    @Override
    public User getUser(Long id) throws UserException {
        // 实现逻辑
        return user;
    }
}

/*
Java vs Go接口对比：

1. 声明方式：
   - Java: 显式implements声明
   - Go: 隐式实现，duck typing

2. 接口大小：
   - Java: 倾向于大接口，如JPA的Repository
   - Go: 推崇小接口，单一职责

3. 继承vs组合：
   - Java: 支持接口继承，extends关键字
   - Go: 接口嵌入，组合优于继承

4. 泛型支持：
   - Java: 原生支持泛型接口
   - Go: 1.18+支持泛型，但设计更简洁
*/
```

```python
# Python接口设计（使用ABC抽象基类）
from abc import ABC, abstractmethod
from typing import List, Optional

class UserRepository(ABC):
    @abstractmethod
    def create(self, user: 'User') -> None:
        pass
    
    @abstractmethod
    def get_by_id(self, user_id: int) -> Optional['User']:
        pass

class DatabaseUserRepository(UserRepository):
    def create(self, user: 'User') -> None:
        # 实现数据库操作
        pass
    
    def get_by_id(self, user_id: int) -> Optional['User']:
        # 实现查询逻辑
        return user

"""
Python vs Go接口对比：

1. 抽象机制：
   - Python: ABC抽象基类，需要继承
   - Go: 接口类型，隐式实现

2. 类型检查：
   - Python: 运行时检查，typing提供静态检查
   - Go: 编译时检查，类型安全

3. 多重继承：
   - Python: 支持多重继承，但容易产生问题
   - Go: 接口组合，避免继承问题

4. 动态特性：
   - Python: 动态语言，运行时可修改
   - Go: 静态语言，编译时确定
"""
```

---

## 🏭 经典设计模式实现

### 工厂模式家族

工厂模式是创建型设计模式的代表，Go语言中的实现更加简洁和灵活。

#### 1. 简单工厂模式

```go
// 来自 mall-go/internal/factory/simple_factory.go
package factory

import (
    "fmt"
    "strings"
)

// 产品接口
type Product interface {
    GetName() string
    GetPrice() float64
    GetDescription() string
}

// 具体产品：电子产品
type ElectronicProduct struct {
    name        string
    price       float64
    description string
    warranty    int // 保修期（月）
}

func (e *ElectronicProduct) GetName() string {
    return e.name
}

func (e *ElectronicProduct) GetPrice() float64 {
    return e.price
}

func (e *ElectronicProduct) GetDescription() string {
    return fmt.Sprintf("%s (保修%d个月)", e.description, e.warranty)
}

// 具体产品：服装产品
type ClothingProduct struct {
    name        string
    price       float64
    description string
    size        string
    material    string
}

func (c *ClothingProduct) GetName() string {
    return c.name
}

func (c *ClothingProduct) GetPrice() float64 {
    return c.price
}

func (c *ClothingProduct) GetDescription() string {
    return fmt.Sprintf("%s (尺码:%s, 材质:%s)", c.description, c.size, c.material)
}

// 具体产品：图书产品
type BookProduct struct {
    name        string
    price       float64
    description string
    author      string
    isbn        string
}

func (b *BookProduct) GetName() string {
    return b.name
}

func (b *BookProduct) GetPrice() float64 {
    return b.price
}

func (b *BookProduct) GetDescription() string {
    return fmt.Sprintf("%s - 作者:%s (ISBN:%s)", b.description, b.author, b.isbn)
}

// 产品类型枚举
type ProductType string

const (
    ElectronicType ProductType = "electronic"
    ClothingType   ProductType = "clothing"
    BookType       ProductType = "book"
)

// 简单工厂
type SimpleProductFactory struct{}

func NewSimpleProductFactory() *SimpleProductFactory {
    return &SimpleProductFactory{}
}

func (f *SimpleProductFactory) CreateProduct(productType ProductType, params map[string]interface{}) (Product, error) {
    switch productType {
    case ElectronicType:
        return &ElectronicProduct{
            name:        params["name"].(string),
            price:       params["price"].(float64),
            description: params["description"].(string),
            warranty:    params["warranty"].(int),
        }, nil
        
    case ClothingType:
        return &ClothingProduct{
            name:        params["name"].(string),
            price:       params["price"].(float64),
            description: params["description"].(string),
            size:        params["size"].(string),
            material:    params["material"].(string),
        }, nil
        
    case BookType:
        return &BookProduct{
            name:        params["name"].(string),
            price:       params["price"].(float64),
            description: params["description"].(string),
            author:      params["author"].(string),
            isbn:        params["isbn"].(string),
        }, nil
        
    default:
        return nil, fmt.Errorf("不支持的产品类型: %s", productType)
    }
}

// 使用示例
func DemonstrateSimpleFactory() {
    fmt.Println("=== 简单工厂模式演示 ===")
    
    factory := NewSimpleProductFactory()
    
    // 创建电子产品
    electronic, err := factory.CreateProduct(ElectronicType, map[string]interface{}{
        "name":        "iPhone 15",
        "price":       7999.0,
        "description": "苹果最新款智能手机",
        "warranty":    12,
    })
    if err != nil {
        fmt.Printf("创建电子产品失败: %v\n", err)
        return
    }
    
    fmt.Printf("电子产品: %s - ¥%.2f\n", electronic.GetName(), electronic.GetPrice())
    fmt.Printf("描述: %s\n", electronic.GetDescription())
    
    // 创建服装产品
    clothing, err := factory.CreateProduct(ClothingType, map[string]interface{}{
        "name":        "优衣库T恤",
        "price":       99.0,
        "description": "纯棉舒适T恤",
        "size":        "L",
        "material":    "100%纯棉",
    })
    if err != nil {
        fmt.Printf("创建服装产品失败: %v\n", err)
        return
    }
    
    fmt.Printf("\n服装产品: %s - ¥%.2f\n", clothing.GetName(), clothing.GetPrice())
    fmt.Printf("描述: %s\n", clothing.GetDescription())
    
    // 创建图书产品
    book, err := factory.CreateProduct(BookType, map[string]interface{}{
        "name":        "Go语言实战",
        "price":       89.0,
        "description": "深入学习Go语言编程",
        "author":      "张三",
        "isbn":        "978-7-111-12345-6",
    })
    if err != nil {
        fmt.Printf("创建图书产品失败: %v\n", err)
        return
    }
    
    fmt.Printf("\n图书产品: %s - ¥%.2f\n", book.GetName(), book.GetPrice())
    fmt.Printf("描述: %s\n", book.GetDescription())
}
```

#### 2. 工厂方法模式

```go
// 来自 mall-go/internal/factory/factory_method.go
package factory

import (
    "fmt"
)

// 抽象工厂接口
type ProductFactory interface {
    CreateProduct(params map[string]interface{}) Product
    GetFactoryType() string
}

// 电子产品工厂
type ElectronicProductFactory struct{}

func NewElectronicProductFactory() *ElectronicProductFactory {
    return &ElectronicProductFactory{}
}

func (f *ElectronicProductFactory) CreateProduct(params map[string]interface{}) Product {
    return &ElectronicProduct{
        name:        params["name"].(string),
        price:       params["price"].(float64),
        description: params["description"].(string),
        warranty:    params["warranty"].(int),
    }
}

func (f *ElectronicProductFactory) GetFactoryType() string {
    return "电子产品工厂"
}

// 服装产品工厂
type ClothingProductFactory struct{}

func NewClothingProductFactory() *ClothingProductFactory {
    return &ClothingProductFactory{}
}

func (f *ClothingProductFactory) CreateProduct(params map[string]interface{}) Product {
    return &ClothingProduct{
        name:        params["name"].(string),
        price:       params["price"].(float64),
        description: params["description"].(string),
        size:        params["size"].(string),
        material:    params["material"].(string),
    }
}

func (f *ClothingProductFactory) GetFactoryType() string {
    return "服装产品工厂"
}

// 图书产品工厂
type BookProductFactory struct{}

func NewBookProductFactory() *BookProductFactory {
    return &BookProductFactory{}
}

func (f *BookProductFactory) CreateProduct(params map[string]interface{}) Product {
    return &BookProduct{
        name:        params["name"].(string),
        price:       params["price"].(float64),
        description: params["description"].(string),
        author:      params["author"].(string),
        isbn:        params["isbn"].(string),
    }
}

func (f *BookProductFactory) GetFactoryType() string {
    return "图书产品工厂"
}

// 工厂注册器
type FactoryRegistry struct {
    factories map[ProductType]ProductFactory
}

func NewFactoryRegistry() *FactoryRegistry {
    registry := &FactoryRegistry{
        factories: make(map[ProductType]ProductFactory),
    }

    // 注册工厂
    registry.RegisterFactory(ElectronicType, NewElectronicProductFactory())
    registry.RegisterFactory(ClothingType, NewClothingProductFactory())
    registry.RegisterFactory(BookType, NewBookProductFactory())

    return registry
}

func (r *FactoryRegistry) RegisterFactory(productType ProductType, factory ProductFactory) {
    r.factories[productType] = factory
    fmt.Printf("注册工厂: %s -> %s\n", productType, factory.GetFactoryType())
}

func (r *FactoryRegistry) GetFactory(productType ProductType) (ProductFactory, error) {
    factory, exists := r.factories[productType]
    if !exists {
        return nil, fmt.Errorf("未找到产品类型 %s 的工厂", productType)
    }
    return factory, nil
}

func (r *FactoryRegistry) CreateProduct(productType ProductType, params map[string]interface{}) (Product, error) {
    factory, err := r.GetFactory(productType)
    if err != nil {
        return nil, err
    }

    return factory.CreateProduct(params), nil
}

// 使用示例
func DemonstrateFactoryMethod() {
    fmt.Println("\n=== 工厂方法模式演示 ===")

    registry := NewFactoryRegistry()

    // 使用不同的工厂创建产品
    products := []struct {
        productType ProductType
        params      map[string]interface{}
    }{
        {
            ElectronicType,
            map[string]interface{}{
                "name":        "MacBook Pro",
                "price":       15999.0,
                "description": "苹果专业笔记本电脑",
                "warranty":    24,
            },
        },
        {
            ClothingType,
            map[string]interface{}{
                "name":        "Nike运动鞋",
                "price":       899.0,
                "description": "专业跑步鞋",
                "size":        "42",
                "material":    "合成材料",
            },
        },
        {
            BookType,
            map[string]interface{}{
                "name":        "设计模式",
                "price":       79.0,
                "description": "经典设计模式详解",
                "author":      "GoF",
                "isbn":        "978-7-111-98765-4",
            },
        },
    }

    fmt.Println("\n创建的产品:")
    for i, productInfo := range products {
        product, err := registry.CreateProduct(productInfo.productType, productInfo.params)
        if err != nil {
            fmt.Printf("创建产品失败: %v\n", err)
            continue
        }

        fmt.Printf("%d. %s - ¥%.2f\n", i+1, product.GetName(), product.GetPrice())
        fmt.Printf("   %s\n", product.GetDescription())
    }
}
```

#### 3. 抽象工厂模式

```go
// 来自 mall-go/internal/factory/abstract_factory.go
package factory

import (
    "fmt"
)

// 抽象产品族接口
type UIComponent interface {
    Render() string
}

type Button interface {
    UIComponent
    Click() string
}

type TextField interface {
    UIComponent
    SetText(text string)
    GetText() string
}

type Dialog interface {
    UIComponent
    Show() string
    Hide() string
}

// Windows风格组件
type WindowsButton struct {
    text string
}

func (b *WindowsButton) Render() string {
    return fmt.Sprintf("[Windows按钮: %s]", b.text)
}

func (b *WindowsButton) Click() string {
    return fmt.Sprintf("Windows按钮 '%s' 被点击", b.text)
}

type WindowsTextField struct {
    text string
}

func (t *WindowsTextField) Render() string {
    return fmt.Sprintf("[Windows文本框: %s]", t.text)
}

func (t *WindowsTextField) SetText(text string) {
    t.text = text
}

func (t *WindowsTextField) GetText() string {
    return t.text
}

type WindowsDialog struct {
    title string
}

func (d *WindowsDialog) Render() string {
    return fmt.Sprintf("[Windows对话框: %s]", d.title)
}

func (d *WindowsDialog) Show() string {
    return fmt.Sprintf("显示Windows对话框: %s", d.title)
}

func (d *WindowsDialog) Hide() string {
    return fmt.Sprintf("隐藏Windows对话框: %s", d.title)
}

// macOS风格组件
type MacOSButton struct {
    text string
}

func (b *MacOSButton) Render() string {
    return fmt.Sprintf("◉ macOS按钮: %s", b.text)
}

func (b *MacOSButton) Click() string {
    return fmt.Sprintf("macOS按钮 '%s' 被点击", b.text)
}

type MacOSTextField struct {
    text string
}

func (t *MacOSTextField) Render() string {
    return fmt.Sprintf("◉ macOS文本框: %s", t.text)
}

func (t *MacOSTextField) SetText(text string) {
    t.text = text
}

func (t *MacOSTextField) GetText() string {
    return t.text
}

type MacOSDialog struct {
    title string
}

func (d *MacOSDialog) Render() string {
    return fmt.Sprintf("◉ macOS对话框: %s", d.title)
}

func (d *MacOSDialog) Show() string {
    return fmt.Sprintf("显示macOS对话框: %s", d.title)
}

func (d *MacOSDialog) Hide() string {
    return fmt.Sprintf("隐藏macOS对话框: %s", d.title)
}

// 抽象工厂接口
type UIFactory interface {
    CreateButton(text string) Button
    CreateTextField(placeholder string) TextField
    CreateDialog(title string) Dialog
    GetTheme() string
}

// Windows UI工厂
type WindowsUIFactory struct{}

func NewWindowsUIFactory() *WindowsUIFactory {
    return &WindowsUIFactory{}
}

func (f *WindowsUIFactory) CreateButton(text string) Button {
    return &WindowsButton{text: text}
}

func (f *WindowsUIFactory) CreateTextField(placeholder string) TextField {
    return &WindowsTextField{text: placeholder}
}

func (f *WindowsUIFactory) CreateDialog(title string) Dialog {
    return &WindowsDialog{title: title}
}

func (f *WindowsUIFactory) GetTheme() string {
    return "Windows主题"
}

// macOS UI工厂
type MacOSUIFactory struct{}

func NewMacOSUIFactory() *MacOSUIFactory {
    return &MacOSUIFactory{}
}

func (f *MacOSUIFactory) CreateButton(text string) Button {
    return &MacOSButton{text: text}
}

func (f *MacOSUIFactory) CreateTextField(placeholder string) TextField {
    return &MacOSTextField{text: placeholder}
}

func (f *MacOSUIFactory) CreateDialog(title string) Dialog {
    return &MacOSDialog{title: title}
}

func (f *MacOSUIFactory) GetTheme() string {
    return "macOS主题"
}

// 应用程序类
type Application struct {
    factory UIFactory
    button  Button
    textField TextField
    dialog  Dialog
}

func NewApplication(factory UIFactory) *Application {
    app := &Application{factory: factory}
    app.createUI()
    return app
}

func (a *Application) createUI() {
    a.button = a.factory.CreateButton("确定")
    a.textField = a.factory.CreateTextField("请输入内容...")
    a.dialog = a.factory.CreateDialog("提示")
}

func (a *Application) Run() {
    fmt.Printf("应用程序使用: %s\n", a.factory.GetTheme())
    fmt.Printf("UI组件渲染:\n")
    fmt.Printf("  %s\n", a.button.Render())
    fmt.Printf("  %s\n", a.textField.Render())
    fmt.Printf("  %s\n", a.dialog.Render())

    fmt.Printf("\n交互演示:\n")
    fmt.Printf("  %s\n", a.button.Click())

    a.textField.SetText("Hello World")
    fmt.Printf("  设置文本: %s\n", a.textField.GetText())

    fmt.Printf("  %s\n", a.dialog.Show())
}

// 使用示例
func DemonstrateAbstractFactory() {
    fmt.Println("\n=== 抽象工厂模式演示 ===")

    // 创建Windows风格的应用
    fmt.Println("1. Windows风格应用:")
    windowsFactory := NewWindowsUIFactory()
    windowsApp := NewApplication(windowsFactory)
    windowsApp.Run()

    fmt.Println("\n" + strings.Repeat("-", 50))

    // 创建macOS风格的应用
    fmt.Println("2. macOS风格应用:")
    macosFactory := NewMacOSUIFactory()
    macosApp := NewApplication(macosFactory)
    macosApp.Run()
}
```

### 策略模式

策略模式定义了算法族，分别封装起来，让它们之间可以互相替换。

```go
// 来自 mall-go/internal/payment/strategy.go
package payment

import (
    "fmt"
    "math/rand"
    "time"
)

// 支付策略接口
type PaymentStrategy interface {
    Pay(amount float64, orderID string) (*PaymentResult, error)
    GetPaymentMethod() string
    ValidatePayment(amount float64) error
}

// 支付结果
type PaymentResult struct {
    TransactionID string    `json:"transaction_id"`
    Amount        float64   `json:"amount"`
    Method        string    `json:"method"`
    Status        string    `json:"status"`
    Message       string    `json:"message"`
    Timestamp     time.Time `json:"timestamp"`
}

// 支付宝支付策略
type AlipayStrategy struct {
    appID     string
    appSecret string
}

func NewAlipayStrategy(appID, appSecret string) *AlipayStrategy {
    return &AlipayStrategy{
        appID:     appID,
        appSecret: appSecret,
    }
}

func (a *AlipayStrategy) ValidatePayment(amount float64) error {
    if amount <= 0 {
        return fmt.Errorf("支付金额必须大于0")
    }
    if amount > 50000 {
        return fmt.Errorf("支付宝单笔支付金额不能超过50000元")
    }
    return nil
}

func (a *AlipayStrategy) Pay(amount float64, orderID string) (*PaymentResult, error) {
    if err := a.ValidatePayment(amount); err != nil {
        return nil, err
    }

    // 模拟支付宝支付流程
    fmt.Printf("正在通过支付宝支付 ¥%.2f (订单: %s)\n", amount, orderID)

    // 模拟网络请求延迟
    time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

    // 模拟90%成功率
    success := rand.Float32() < 0.9

    result := &PaymentResult{
        TransactionID: fmt.Sprintf("alipay_%d", time.Now().UnixNano()),
        Amount:        amount,
        Method:        a.GetPaymentMethod(),
        Timestamp:     time.Now(),
    }

    if success {
        result.Status = "success"
        result.Message = "支付宝支付成功"
    } else {
        result.Status = "failed"
        result.Message = "支付宝支付失败，请重试"
    }

    return result, nil
}

func (a *AlipayStrategy) GetPaymentMethod() string {
    return "支付宝"
}

// 微信支付策略
type WechatPayStrategy struct {
    merchantID string
    apiKey     string
}

func NewWechatPayStrategy(merchantID, apiKey string) *WechatPayStrategy {
    return &WechatPayStrategy{
        merchantID: merchantID,
        apiKey:     apiKey,
    }
}

func (w *WechatPayStrategy) ValidatePayment(amount float64) error {
    if amount <= 0 {
        return fmt.Errorf("支付金额必须大于0")
    }
    if amount > 20000 {
        return fmt.Errorf("微信支付单笔支付金额不能超过20000元")
    }
    return nil
}

func (w *WechatPayStrategy) Pay(amount float64, orderID string) (*PaymentResult, error) {
    if err := w.ValidatePayment(amount); err != nil {
        return nil, err
    }

    fmt.Printf("正在通过微信支付 ¥%.2f (订单: %s)\n", amount, orderID)

    time.Sleep(time.Duration(rand.Intn(800)) * time.Millisecond)

    success := rand.Float32() < 0.85

    result := &PaymentResult{
        TransactionID: fmt.Sprintf("wechat_%d", time.Now().UnixNano()),
        Amount:        amount,
        Method:        w.GetPaymentMethod(),
        Timestamp:     time.Now(),
    }

    if success {
        result.Status = "success"
        result.Message = "微信支付成功"
    } else {
        result.Status = "failed"
        result.Message = "微信支付失败，请重试"
    }

    return result, nil
}

func (w *WechatPayStrategy) GetPaymentMethod() string {
    return "微信支付"
}

// 银行卡支付策略
type BankCardStrategy struct {
    bankCode string
    gateway  string
}

func NewBankCardStrategy(bankCode, gateway string) *BankCardStrategy {
    return &BankCardStrategy{
        bankCode: bankCode,
        gateway:  gateway,
    }
}

func (b *BankCardStrategy) ValidatePayment(amount float64) error {
    if amount <= 0 {
        return fmt.Errorf("支付金额必须大于0")
    }
    if amount > 100000 {
        return fmt.Errorf("银行卡单笔支付金额不能超过100000元")
    }
    return nil
}

func (b *BankCardStrategy) Pay(amount float64, orderID string) (*PaymentResult, error) {
    if err := b.ValidatePayment(amount); err != nil {
        return nil, err
    }

    fmt.Printf("正在通过银行卡支付 ¥%.2f (订单: %s)\n", amount, orderID)

    time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond)

    success := rand.Float32() < 0.95

    result := &PaymentResult{
        TransactionID: fmt.Sprintf("bank_%d", time.Now().UnixNano()),
        Amount:        amount,
        Method:        b.GetPaymentMethod(),
        Timestamp:     time.Now(),
    }

    if success {
        result.Status = "success"
        result.Message = "银行卡支付成功"
    } else {
        result.Status = "failed"
        result.Message = "银行卡支付失败，请联系银行"
    }

    return result, nil
}

func (b *BankCardStrategy) GetPaymentMethod() string {
    return "银行卡"
}

// 支付上下文
type PaymentContext struct {
    strategy PaymentStrategy
}

func NewPaymentContext(strategy PaymentStrategy) *PaymentContext {
    return &PaymentContext{strategy: strategy}
}

func (p *PaymentContext) SetStrategy(strategy PaymentStrategy) {
    p.strategy = strategy
    fmt.Printf("切换支付方式为: %s\n", strategy.GetPaymentMethod())
}

func (p *PaymentContext) ExecutePayment(amount float64, orderID string) (*PaymentResult, error) {
    if p.strategy == nil {
        return nil, fmt.Errorf("未设置支付策略")
    }

    return p.strategy.Pay(amount, orderID)
}

// 支付服务
type PaymentService struct {
    strategies map[string]PaymentStrategy
    context    *PaymentContext
}

func NewPaymentService() *PaymentService {
    service := &PaymentService{
        strategies: make(map[string]PaymentStrategy),
    }

    // 注册支付策略
    service.RegisterStrategy("alipay", NewAlipayStrategy("app123", "secret456"))
    service.RegisterStrategy("wechat", NewWechatPayStrategy("mch789", "key012"))
    service.RegisterStrategy("bankcard", NewBankCardStrategy("ICBC", "gateway.icbc.com"))

    return service
}

func (s *PaymentService) RegisterStrategy(name string, strategy PaymentStrategy) {
    s.strategies[name] = strategy
    fmt.Printf("注册支付策略: %s -> %s\n", name, strategy.GetPaymentMethod())
}

func (s *PaymentService) Pay(method string, amount float64, orderID string) (*PaymentResult, error) {
    strategy, exists := s.strategies[method]
    if !exists {
        return nil, fmt.Errorf("不支持的支付方式: %s", method)
    }

    if s.context == nil {
        s.context = NewPaymentContext(strategy)
    } else {
        s.context.SetStrategy(strategy)
    }

    return s.context.ExecutePayment(amount, orderID)
}

func (s *PaymentService) GetSupportedMethods() []string {
    var methods []string
    for name := range s.strategies {
        methods = append(methods, name)
    }
    return methods
}

// 使用示例
func DemonstrateStrategyPattern() {
    fmt.Println("\n=== 策略模式演示 ===")

    paymentService := NewPaymentService()

    fmt.Printf("\n支持的支付方式: %v\n", paymentService.GetSupportedMethods())

    // 测试订单
    orders := []struct {
        orderID string
        amount  float64
        method  string
    }{
        {"ORDER001", 299.99, "alipay"},
        {"ORDER002", 1599.00, "wechat"},
        {"ORDER003", 5999.99, "bankcard"},
        {"ORDER004", 99.99, "alipay"},
    }

    fmt.Println("\n开始处理订单:")
    for _, order := range orders {
        fmt.Printf("\n处理订单 %s (¥%.2f):\n", order.orderID, order.amount)

        result, err := paymentService.Pay(order.method, order.amount, order.orderID)
        if err != nil {
            fmt.Printf("❌ 支付失败: %v\n", err)
            continue
        }

        if result.Status == "success" {
            fmt.Printf("✅ %s\n", result.Message)
            fmt.Printf("   交易ID: %s\n", result.TransactionID)
        } else {
            fmt.Printf("❌ %s\n", result.Message)
        }
    }
}
```

### 装饰器模式

装饰器模式允许向一个现有的对象添加新的功能，同时又不改变其结构。

```go
// 来自 mall-go/internal/middleware/decorator.go
package middleware

import (
    "fmt"
    "net/http"
    "time"
)

// HTTP处理器接口
type Handler interface {
    Handle(w http.ResponseWriter, r *http.Request)
}

// 基础处理器
type BaseHandler struct {
    name string
}

func NewBaseHandler(name string) *BaseHandler {
    return &BaseHandler{name: name}
}

func (h *BaseHandler) Handle(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("基础处理器 '%s' 处理请求: %s %s\n", h.name, r.Method, r.URL.Path)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf("Hello from %s", h.name)))
}

// 装饰器基类
type HandlerDecorator struct {
    handler Handler
}

// 日志装饰器
type LoggingDecorator struct {
    HandlerDecorator
    logger string
}

func NewLoggingDecorator(handler Handler, logger string) *LoggingDecorator {
    return &LoggingDecorator{
        HandlerDecorator: HandlerDecorator{handler: handler},
        logger:          logger,
    }
}

func (d *LoggingDecorator) Handle(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    fmt.Printf("[%s] 请求开始: %s %s\n", d.logger, r.Method, r.URL.Path)

    // 调用被装饰的处理器
    d.handler.Handle(w, r)

    duration := time.Since(start)
    fmt.Printf("[%s] 请求完成: 耗时 %v\n", d.logger, duration)
}

// 认证装饰器
type AuthDecorator struct {
    HandlerDecorator
    requiredRole string
}

func NewAuthDecorator(handler Handler, requiredRole string) *AuthDecorator {
    return &AuthDecorator{
        HandlerDecorator: HandlerDecorator{handler: handler},
        requiredRole:     requiredRole,
    }
}

func (d *AuthDecorator) Handle(w http.ResponseWriter, r *http.Request) {
    // 模拟认证检查
    userRole := r.Header.Get("X-User-Role")

    fmt.Printf("[认证] 检查用户权限: 需要 '%s', 当前 '%s'\n", d.requiredRole, userRole)

    if userRole != d.requiredRole {
        fmt.Printf("[认证] 权限不足，拒绝访问\n")
        w.WriteHeader(http.StatusForbidden)
        w.Write([]byte("权限不足"))
        return
    }

    fmt.Printf("[认证] 权限验证通过\n")
    d.handler.Handle(w, r)
}

// 限流装饰器
type RateLimitDecorator struct {
    HandlerDecorator
    maxRequests int
    window      time.Duration
    requests    map[string][]time.Time
}

func NewRateLimitDecorator(handler Handler, maxRequests int, window time.Duration) *RateLimitDecorator {
    return &RateLimitDecorator{
        HandlerDecorator: HandlerDecorator{handler: handler},
        maxRequests:      maxRequests,
        window:          window,
        requests:        make(map[string][]time.Time),
    }
}

func (d *RateLimitDecorator) Handle(w http.ResponseWriter, r *http.Request) {
    clientIP := r.RemoteAddr
    now := time.Now()

    // 清理过期的请求记录
    if requests, exists := d.requests[clientIP]; exists {
        var validRequests []time.Time
        for _, reqTime := range requests {
            if now.Sub(reqTime) < d.window {
                validRequests = append(validRequests, reqTime)
            }
        }
        d.requests[clientIP] = validRequests
    }

    // 检查是否超过限制
    if len(d.requests[clientIP]) >= d.maxRequests {
        fmt.Printf("[限流] 客户端 %s 请求过于频繁，拒绝服务\n", clientIP)
        w.WriteHeader(http.StatusTooManyRequests)
        w.Write([]byte("请求过于频繁，请稍后再试"))
        return
    }

    // 记录当前请求
    d.requests[clientIP] = append(d.requests[clientIP], now)
    fmt.Printf("[限流] 客户端 %s 请求通过 (%d/%d)\n", clientIP, len(d.requests[clientIP]), d.maxRequests)

    d.handler.Handle(w, r)
}

// 缓存装饰器
type CacheDecorator struct {
    HandlerDecorator
    cache map[string]CacheEntry
    ttl   time.Duration
}

type CacheEntry struct {
    data      []byte
    timestamp time.Time
}

func NewCacheDecorator(handler Handler, ttl time.Duration) *CacheDecorator {
    return &CacheDecorator{
        HandlerDecorator: HandlerDecorator{handler: handler},
        cache:           make(map[string]CacheEntry),
        ttl:             ttl,
    }
}

func (d *CacheDecorator) Handle(w http.ResponseWriter, r *http.Request) {
    cacheKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)

    // 检查缓存
    if entry, exists := d.cache[cacheKey]; exists {
        if time.Since(entry.timestamp) < d.ttl {
            fmt.Printf("[缓存] 缓存命中: %s\n", cacheKey)
            w.WriteHeader(http.StatusOK)
            w.Write(entry.data)
            return
        } else {
            fmt.Printf("[缓存] 缓存过期: %s\n", cacheKey)
            delete(d.cache, cacheKey)
        }
    }

    fmt.Printf("[缓存] 缓存未命中: %s\n", cacheKey)

    // 创建响应记录器来捕获响应
    recorder := &ResponseRecorder{
        ResponseWriter: w,
        body:          make([]byte, 0),
    }

    d.handler.Handle(recorder, r)

    // 缓存响应
    d.cache[cacheKey] = CacheEntry{
        data:      recorder.body,
        timestamp: time.Now(),
    }

    fmt.Printf("[缓存] 响应已缓存: %s\n", cacheKey)
}

// 响应记录器
type ResponseRecorder struct {
    http.ResponseWriter
    body []byte
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
    r.body = append(r.body, data...)
    return r.ResponseWriter.Write(data)
}

// 装饰器构建器
type DecoratorBuilder struct {
    handler Handler
}

func NewDecoratorBuilder(handler Handler) *DecoratorBuilder {
    return &DecoratorBuilder{handler: handler}
}

func (b *DecoratorBuilder) WithLogging(logger string) *DecoratorBuilder {
    b.handler = NewLoggingDecorator(b.handler, logger)
    return b
}

func (b *DecoratorBuilder) WithAuth(requiredRole string) *DecoratorBuilder {
    b.handler = NewAuthDecorator(b.handler, requiredRole)
    return b
}

func (b *DecoratorBuilder) WithRateLimit(maxRequests int, window time.Duration) *DecoratorBuilder {
    b.handler = NewRateLimitDecorator(b.handler, maxRequests, window)
    return b
}

func (b *DecoratorBuilder) WithCache(ttl time.Duration) *DecoratorBuilder {
    b.handler = NewCacheDecorator(b.handler, ttl)
    return b
}

func (b *DecoratorBuilder) Build() Handler {
    return b.handler
}

// 使用示例
func DemonstrateDecoratorPattern() {
    fmt.Println("\n=== 装饰器模式演示 ===")

    // 创建基础处理器
    baseHandler := NewBaseHandler("用户服务")

    // 使用构建器模式组合装饰器
    decoratedHandler := NewDecoratorBuilder(baseHandler).
        WithLogging("UserService").
        WithAuth("admin").
        WithRateLimit(3, time.Minute).
        WithCache(30 * time.Second).
        Build()

    // 模拟HTTP请求
    fmt.Println("\n模拟HTTP请求处理:")

    // 创建模拟请求
    requests := []*http.Request{
        {
            Method: "GET",
            URL:    &url.URL{Path: "/users"},
            Header: http.Header{"X-User-Role": []string{"admin"}},
            RemoteAddr: "192.168.1.100",
        },
        {
            Method: "GET",
            URL:    &url.URL{Path: "/users"},
            Header: http.Header{"X-User-Role": []string{"user"}},
            RemoteAddr: "192.168.1.101",
        },
        {
            Method: "GET",
            URL:    &url.URL{Path: "/users"},
            Header: http.Header{"X-User-Role": []string{"admin"}},
            RemoteAddr: "192.168.1.100",
        },
    }

    for i, req := range requests {
        fmt.Printf("\n--- 请求 %d ---\n", i+1)

        // 创建模拟响应写入器
        w := &MockResponseWriter{}

        decoratedHandler.Handle(w, req)

        fmt.Printf("响应状态: %d\n", w.statusCode)
        if len(w.body) > 0 {
            fmt.Printf("响应内容: %s\n", string(w.body))
        }
    }
}

// 模拟响应写入器
type MockResponseWriter struct {
    statusCode int
    body       []byte
    header     http.Header
}

func (w *MockResponseWriter) Header() http.Header {
    if w.header == nil {
        w.header = make(http.Header)
    }
    return w.header
}

func (w *MockResponseWriter) Write(data []byte) (int, error) {
    w.body = append(w.body, data...)
    return len(data), nil
}

func (w *MockResponseWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
}
```

### 适配器模式

适配器模式将一个类的接口转换成客户希望的另一个接口。

```go
// 来自 mall-go/internal/adapter/database_adapter.go
package adapter

import (
    "fmt"
    "time"
)

// 目标接口：统一的数据库操作接口
type DatabaseInterface interface {
    Connect() error
    Disconnect() error
    Query(sql string) ([]map[string]interface{}, error)
    Execute(sql string) error
    GetConnectionInfo() string
}

// 第三方MySQL库（被适配者）
type MySQLClient struct {
    host     string
    port     int
    username string
    password string
    database string
    connected bool
}

func NewMySQLClient(host string, port int, username, password, database string) *MySQLClient {
    return &MySQLClient{
        host:     host,
        port:     port,
        username: username,
        password: password,
        database: database,
    }
}

func (m *MySQLClient) OpenConnection() error {
    fmt.Printf("MySQL: 连接到 %s:%d/%s\n", m.host, m.port, m.database)
    m.connected = true
    return nil
}

func (m *MySQLClient) CloseConnection() error {
    fmt.Printf("MySQL: 关闭连接\n")
    m.connected = false
    return nil
}

func (m *MySQLClient) ExecuteQuery(query string) ([]map[string]interface{}, error) {
    if !m.connected {
        return nil, fmt.Errorf("MySQL连接未建立")
    }

    fmt.Printf("MySQL: 执行查询 - %s\n", query)

    // 模拟查询结果
    results := []map[string]interface{}{
        {"id": 1, "name": "张三", "email": "zhangsan@example.com"},
        {"id": 2, "name": "李四", "email": "lisi@example.com"},
    }

    return results, nil
}

func (m *MySQLClient) ExecuteCommand(command string) error {
    if !m.connected {
        return fmt.Errorf("MySQL连接未建立")
    }

    fmt.Printf("MySQL: 执行命令 - %s\n", command)
    return nil
}

// 第三方PostgreSQL库（被适配者）
type PostgreSQLClient struct {
    connectionString string
    isConnected     bool
}

func NewPostgreSQLClient(connectionString string) *PostgreSQLClient {
    return &PostgreSQLClient{
        connectionString: connectionString,
    }
}

func (p *PostgreSQLClient) EstablishConnection() error {
    fmt.Printf("PostgreSQL: 建立连接 - %s\n", p.connectionString)
    p.isConnected = true
    return nil
}

func (p *PostgreSQLClient) TerminateConnection() error {
    fmt.Printf("PostgreSQL: 终止连接\n")
    p.isConnected = false
    return nil
}

func (p *PostgreSQLClient) RunQuery(sql string) ([]map[string]interface{}, error) {
    if !p.isConnected {
        return nil, fmt.Errorf("PostgreSQL连接未建立")
    }

    fmt.Printf("PostgreSQL: 运行查询 - %s\n", sql)

    // 模拟查询结果
    results := []map[string]interface{}{
        {"user_id": 1, "username": "admin", "role": "administrator"},
        {"user_id": 2, "username": "user1", "role": "user"},
    }

    return results, nil
}

func (p *PostgreSQLClient) RunCommand(sql string) error {
    if !p.isConnected {
        return fmt.Errorf("PostgreSQL连接未建立")
    }

    fmt.Printf("PostgreSQL: 运行命令 - %s\n", sql)
    return nil
}

// MySQL适配器
type MySQLAdapter struct {
    client *MySQLClient
}

func NewMySQLAdapter(client *MySQLClient) *MySQLAdapter {
    return &MySQLAdapter{client: client}
}

func (a *MySQLAdapter) Connect() error {
    return a.client.OpenConnection()
}

func (a *MySQLAdapter) Disconnect() error {
    return a.client.CloseConnection()
}

func (a *MySQLAdapter) Query(sql string) ([]map[string]interface{}, error) {
    return a.client.ExecuteQuery(sql)
}

func (a *MySQLAdapter) Execute(sql string) error {
    return a.client.ExecuteCommand(sql)
}

func (a *MySQLAdapter) GetConnectionInfo() string {
    return fmt.Sprintf("MySQL - %s:%d/%s", a.client.host, a.client.port, a.client.database)
}

// PostgreSQL适配器
type PostgreSQLAdapter struct {
    client *PostgreSQLClient
}

func NewPostgreSQLAdapter(client *PostgreSQLClient) *PostgreSQLAdapter {
    return &PostgreSQLAdapter{client: client}
}

func (a *PostgreSQLAdapter) Connect() error {
    return a.client.EstablishConnection()
}

func (a *PostgreSQLAdapter) Disconnect() error {
    return a.client.TerminateConnection()
}

func (a *PostgreSQLAdapter) Query(sql string) ([]map[string]interface{}, error) {
    return a.client.RunQuery(sql)
}

func (a *PostgreSQLAdapter) Execute(sql string) error {
    return a.client.RunCommand(sql)
}

func (a *PostgreSQLAdapter) GetConnectionInfo() string {
    return fmt.Sprintf("PostgreSQL - %s", a.client.connectionString)
}

// 数据库管理器
type DatabaseManager struct {
    databases map[string]DatabaseInterface
}

func NewDatabaseManager() *DatabaseManager {
    return &DatabaseManager{
        databases: make(map[string]DatabaseInterface),
    }
}

func (dm *DatabaseManager) RegisterDatabase(name string, db DatabaseInterface) {
    dm.databases[name] = db
    fmt.Printf("注册数据库: %s -> %s\n", name, db.GetConnectionInfo())
}

func (dm *DatabaseManager) GetDatabase(name string) (DatabaseInterface, error) {
    db, exists := dm.databases[name]
    if !exists {
        return nil, fmt.Errorf("数据库 '%s' 未注册", name)
    }
    return db, nil
}

func (dm *DatabaseManager) ExecuteOnAll(operation func(DatabaseInterface) error) {
    for name, db := range dm.databases {
        fmt.Printf("\n在数据库 '%s' 上执行操作:\n", name)
        if err := operation(db); err != nil {
            fmt.Printf("操作失败: %v\n", err)
        }
    }
}

// 使用示例
func DemonstrateAdapterPattern() {
    fmt.Println("\n=== 适配器模式演示 ===")

    // 创建数据库管理器
    manager := NewDatabaseManager()

    // 创建MySQL客户端和适配器
    mysqlClient := NewMySQLClient("localhost", 3306, "root", "password", "mall")
    mysqlAdapter := NewMySQLAdapter(mysqlClient)

    // 创建PostgreSQL客户端和适配器
    pgClient := NewPostgreSQLClient("postgresql://user:pass@localhost:5432/mall")
    pgAdapter := NewPostgreSQLAdapter(pgClient)

    // 注册数据库
    manager.RegisterDatabase("mysql", mysqlAdapter)
    manager.RegisterDatabase("postgresql", pgAdapter)

    // 统一操作不同的数据库
    fmt.Println("\n=== 统一数据库操作演示 ===")

    // 连接所有数据库
    manager.ExecuteOnAll(func(db DatabaseInterface) error {
        return db.Connect()
    })

    // 在所有数据库上执行查询
    manager.ExecuteOnAll(func(db DatabaseInterface) error {
        results, err := db.Query("SELECT * FROM users")
        if err != nil {
            return err
        }

        fmt.Printf("查询结果 (%d 条记录):\n", len(results))
        for i, result := range results {
            fmt.Printf("  %d: %v\n", i+1, result)
        }

        return nil
    })

    // 在所有数据库上执行命令
    manager.ExecuteOnAll(func(db DatabaseInterface) error {
        return db.Execute("UPDATE users SET last_login = NOW()")
    })

    // 断开所有数据库连接
    manager.ExecuteOnAll(func(db DatabaseInterface) error {
        return db.Disconnect()
    })

    // 演示单独使用特定数据库
    fmt.Println("\n=== 单独数据库操作演示 ===")

    mysql, err := manager.GetDatabase("mysql")
    if err != nil {
        fmt.Printf("获取MySQL数据库失败: %v\n", err)
        return
    }

    fmt.Printf("使用数据库: %s\n", mysql.GetConnectionInfo())
    mysql.Connect()

    results, err := mysql.Query("SELECT COUNT(*) FROM orders")
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
    } else {
        fmt.Printf("订单统计: %v\n", results)
    }

    mysql.Disconnect()
}
```

---

## 💉 依赖注入和控制反转

依赖注入是一种实现控制反转的技术，Go语言中可以通过接口和构造函数实现。

```go
// 来自 mall-go/internal/di/dependency_injection.go
package di

import (
    "fmt"
    "reflect"
    "sync"
)

// 服务接口定义
type UserRepository interface {
    GetUser(id int64) (*User, error)
    SaveUser(user *User) error
}

type EmailService interface {
    SendEmail(to, subject, body string) error
}

type LoggerService interface {
    Log(level string, message string)
    Error(message string)
    Info(message string)
}

// 数据模型
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 具体实现
type DatabaseUserRepository struct {
    connectionString string
    logger          LoggerService
}

func NewDatabaseUserRepository(connectionString string, logger LoggerService) *DatabaseUserRepository {
    return &DatabaseUserRepository{
        connectionString: connectionString,
        logger:          logger,
    }
}

func (r *DatabaseUserRepository) GetUser(id int64) (*User, error) {
    r.logger.Info(fmt.Sprintf("从数据库获取用户: %d", id))

    // 模拟数据库查询
    user := &User{
        ID:    id,
        Name:  fmt.Sprintf("用户%d", id),
        Email: fmt.Sprintf("user%d@example.com", id),
    }

    return user, nil
}

func (r *DatabaseUserRepository) SaveUser(user *User) error {
    r.logger.Info(fmt.Sprintf("保存用户到数据库: %s", user.Name))
    return nil
}

type SMTPEmailService struct {
    host   string
    port   int
    logger LoggerService
}

func NewSMTPEmailService(host string, port int, logger LoggerService) *SMTPEmailService {
    return &SMTPEmailService{
        host:   host,
        port:   port,
        logger: logger,
    }
}

func (s *SMTPEmailService) SendEmail(to, subject, body string) error {
    s.logger.Info(fmt.Sprintf("通过SMTP发送邮件: %s -> %s", subject, to))
    fmt.Printf("📧 邮件已发送: %s\n", subject)
    return nil
}

type ConsoleLogger struct {
    prefix string
}

func NewConsoleLogger(prefix string) *ConsoleLogger {
    return &ConsoleLogger{prefix: prefix}
}

func (l *ConsoleLogger) Log(level string, message string) {
    fmt.Printf("[%s] %s: %s\n", l.prefix, level, message)
}

func (l *ConsoleLogger) Error(message string) {
    l.Log("ERROR", message)
}

func (l *ConsoleLogger) Info(message string) {
    l.Log("INFO", message)
}

// 用户服务 - 依赖注入的目标
type UserService struct {
    userRepo     UserRepository
    emailService EmailService
    logger       LoggerService
}

// 构造函数注入
func NewUserService(userRepo UserRepository, emailService EmailService, logger LoggerService) *UserService {
    return &UserService{
        userRepo:     userRepo,
        emailService: emailService,
        logger:       logger,
    }
}

func (s *UserService) RegisterUser(name, email string) (*User, error) {
    s.logger.Info(fmt.Sprintf("注册新用户: %s", name))

    user := &User{
        ID:    time.Now().Unix(),
        Name:  name,
        Email: email,
    }

    // 保存用户
    if err := s.userRepo.SaveUser(user); err != nil {
        s.logger.Error(fmt.Sprintf("保存用户失败: %v", err))
        return nil, err
    }

    // 发送欢迎邮件
    welcomeSubject := "欢迎注册"
    welcomeBody := fmt.Sprintf("欢迎 %s 注册我们的服务！", name)

    if err := s.emailService.SendEmail(email, welcomeSubject, welcomeBody); err != nil {
        s.logger.Error(fmt.Sprintf("发送欢迎邮件失败: %v", err))
        // 邮件发送失败不影响用户注册
    }

    s.logger.Info(fmt.Sprintf("用户注册成功: %s", name))
    return user, nil
}

func (s *UserService) GetUserProfile(id int64) (*User, error) {
    s.logger.Info(fmt.Sprintf("获取用户资料: %d", id))
    return s.userRepo.GetUser(id)
}

// 简单的依赖注入容器
type DIContainer struct {
    services map[reflect.Type]interface{}
    mu       sync.RWMutex
}

func NewDIContainer() *DIContainer {
    return &DIContainer{
        services: make(map[reflect.Type]interface{}),
    }
}

// 注册服务
func (c *DIContainer) Register(serviceType interface{}, implementation interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()

    t := reflect.TypeOf(serviceType).Elem() // 获取接口类型
    c.services[t] = implementation

    fmt.Printf("注册服务: %s -> %T\n", t.Name(), implementation)
}

// 解析服务
func (c *DIContainer) Resolve(serviceType interface{}) (interface{}, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    t := reflect.TypeOf(serviceType).Elem()
    service, exists := c.services[t]
    if !exists {
        return nil, fmt.Errorf("服务 %s 未注册", t.Name())
    }

    return service, nil
}

// 自动装配
func (c *DIContainer) Autowire(target interface{}) error {
    targetValue := reflect.ValueOf(target)
    if targetValue.Kind() != reflect.Ptr {
        return fmt.Errorf("目标必须是指针类型")
    }

    targetValue = targetValue.Elem()
    targetType := targetValue.Type()

    for i := 0; i < targetValue.NumField(); i++ {
        field := targetValue.Field(i)
        fieldType := targetType.Field(i)

        // 检查是否有依赖注入标签
        if tag := fieldType.Tag.Get("inject"); tag == "true" {
            if !field.CanSet() {
                continue
            }

            // 解析依赖
            service, err := c.Resolve(reflect.New(field.Type()).Interface())
            if err != nil {
                return fmt.Errorf("无法解析字段 %s 的依赖: %v", fieldType.Name, err)
            }

            field.Set(reflect.ValueOf(service))
            fmt.Printf("自动装配: %s.%s\n", targetType.Name(), fieldType.Name)
        }
    }

    return nil
}

// 支持自动装配的用户服务
type AutowiredUserService struct {
    UserRepo     UserRepository `inject:"true"`
    EmailService EmailService   `inject:"true"`
    Logger       LoggerService  `inject:"true"`
}

func (s *AutowiredUserService) ProcessUser(name, email string) {
    s.Logger.Info(fmt.Sprintf("处理用户: %s", name))

    user := &User{
        ID:    time.Now().Unix(),
        Name:  name,
        Email: email,
    }

    s.UserRepo.SaveUser(user)
    s.EmailService.SendEmail(email, "处理完成", "您的用户信息已处理完成")
}

// 使用示例
func DemonstrateDependencyInjection() {
    fmt.Println("\n=== 依赖注入演示 ===")

    // 1. 手动依赖注入
    fmt.Println("1. 手动依赖注入:")

    logger := NewConsoleLogger("UserService")
    userRepo := NewDatabaseUserRepository("mysql://localhost:3306/mall", logger)
    emailService := NewSMTPEmailService("smtp.example.com", 587, logger)

    userService := NewUserService(userRepo, emailService, logger)

    user, err := userService.RegisterUser("张三", "zhangsan@example.com")
    if err != nil {
        fmt.Printf("用户注册失败: %v\n", err)
    } else {
        fmt.Printf("用户注册成功: %+v\n", user)
    }

    // 2. 使用DI容器
    fmt.Println("\n2. 使用DI容器:")

    container := NewDIContainer()

    // 注册服务
    container.Register((*LoggerService)(nil), logger)
    container.Register((*UserRepository)(nil), userRepo)
    container.Register((*EmailService)(nil), emailService)

    // 解析服务
    resolvedLogger, _ := container.Resolve((*LoggerService)(nil))
    resolvedUserRepo, _ := container.Resolve((*UserRepository)(nil))
    resolvedEmailService, _ := container.Resolve((*EmailService)(nil))

    containerUserService := NewUserService(
        resolvedUserRepo.(UserRepository),
        resolvedEmailService.(EmailService),
        resolvedLogger.(LoggerService),
    )

    profile, err := containerUserService.GetUserProfile(12345)
    if err != nil {
        fmt.Printf("获取用户资料失败: %v\n", err)
    } else {
        fmt.Printf("用户资料: %+v\n", profile)
    }

    // 3. 自动装配
    fmt.Println("\n3. 自动装配:")

    autowiredService := &AutowiredUserService{}

    if err := container.Autowire(autowiredService); err != nil {
        fmt.Printf("自动装配失败: %v\n", err)
    } else {
        autowiredService.ProcessUser("李四", "lisi@example.com")
    }
}
```

---

## 🔗 接口组合和嵌入

Go语言通过接口嵌入实现组合，这是一种强大的设计技术。

```go
// 来自 mall-go/internal/composition/interface_composition.go
package composition

import (
    "fmt"
    "time"
)

// 基础接口
type Reader interface {
    Read(data []byte) (int, error)
}

type Writer interface {
    Write(data []byte) (int, error)
}

type Closer interface {
    Close() error
}

type Seeker interface {
    Seek(offset int64, whence int) (int64, error)
}

// 组合接口
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

type ReadWriteSeeker interface {
    Reader
    Writer
    Seeker
}

type ReadWriteSeekCloser interface {
    Reader
    Writer
    Seeker
    Closer
}

// 文件操作接口
type FileOperations interface {
    ReadWriteSeekCloser

    // 扩展方法
    GetSize() int64
    GetPath() string
    IsOpen() bool
}

// 具体实现：内存文件
type MemoryFile struct {
    data     []byte
    position int64
    path     string
    isOpen   bool
}

func NewMemoryFile(path string, initialData []byte) *MemoryFile {
    data := make([]byte, len(initialData))
    copy(data, initialData)

    return &MemoryFile{
        data:   data,
        path:   path,
        isOpen: true,
    }
}

func (m *MemoryFile) Read(data []byte) (int, error) {
    if !m.isOpen {
        return 0, fmt.Errorf("文件已关闭")
    }

    if m.position >= int64(len(m.data)) {
        return 0, fmt.Errorf("已到达文件末尾")
    }

    n := copy(data, m.data[m.position:])
    m.position += int64(n)

    fmt.Printf("从内存文件读取 %d 字节\n", n)
    return n, nil
}

func (m *MemoryFile) Write(data []byte) (int, error) {
    if !m.isOpen {
        return 0, fmt.Errorf("文件已关闭")
    }

    // 扩展数据切片如果需要
    needed := m.position + int64(len(data))
    if needed > int64(len(m.data)) {
        newData := make([]byte, needed)
        copy(newData, m.data)
        m.data = newData
    }

    n := copy(m.data[m.position:], data)
    m.position += int64(n)

    fmt.Printf("向内存文件写入 %d 字节\n", n)
    return n, nil
}

func (m *MemoryFile) Seek(offset int64, whence int) (int64, error) {
    if !m.isOpen {
        return 0, fmt.Errorf("文件已关闭")
    }

    var newPosition int64

    switch whence {
    case 0: // 从文件开始
        newPosition = offset
    case 1: // 从当前位置
        newPosition = m.position + offset
    case 2: // 从文件末尾
        newPosition = int64(len(m.data)) + offset
    default:
        return 0, fmt.Errorf("无效的whence值")
    }

    if newPosition < 0 {
        return 0, fmt.Errorf("无效的文件位置")
    }

    m.position = newPosition
    fmt.Printf("文件位置移动到: %d\n", m.position)
    return m.position, nil
}

func (m *MemoryFile) Close() error {
    if !m.isOpen {
        return fmt.Errorf("文件已经关闭")
    }

    m.isOpen = false
    fmt.Printf("内存文件已关闭: %s\n", m.path)
    return nil
}

func (m *MemoryFile) GetSize() int64 {
    return int64(len(m.data))
}

func (m *MemoryFile) GetPath() string {
    return m.path
}

func (m *MemoryFile) IsOpen() bool {
    return m.isOpen
}

// 网络文件实现
type NetworkFile struct {
    url      string
    data     []byte
    position int64
    isOpen   bool
}

func NewNetworkFile(url string) *NetworkFile {
    // 模拟从网络加载数据
    data := []byte(fmt.Sprintf("网络文件内容来自: %s\n这是一些示例数据...", url))

    return &NetworkFile{
        url:    url,
        data:   data,
        isOpen: true,
    }
}

func (n *NetworkFile) Read(data []byte) (int, error) {
    if !n.isOpen {
        return 0, fmt.Errorf("网络文件连接已关闭")
    }

    if n.position >= int64(len(n.data)) {
        return 0, fmt.Errorf("已到达文件末尾")
    }

    copied := copy(data, n.data[n.position:])
    n.position += int64(copied)

    fmt.Printf("从网络文件读取 %d 字节\n", copied)
    return copied, nil
}

func (n *NetworkFile) Write(data []byte) (int, error) {
    return 0, fmt.Errorf("网络文件不支持写入操作")
}

func (n *NetworkFile) Seek(offset int64, whence int) (int64, error) {
    if !n.isOpen {
        return 0, fmt.Errorf("网络文件连接已关闭")
    }

    var newPosition int64

    switch whence {
    case 0:
        newPosition = offset
    case 1:
        newPosition = n.position + offset
    case 2:
        newPosition = int64(len(n.data)) + offset
    default:
        return 0, fmt.Errorf("无效的whence值")
    }

    if newPosition < 0 || newPosition > int64(len(n.data)) {
        return 0, fmt.Errorf("无效的文件位置")
    }

    n.position = newPosition
    fmt.Printf("网络文件位置移动到: %d\n", n.position)
    return n.position, nil
}

func (n *NetworkFile) Close() error {
    if !n.isOpen {
        return fmt.Errorf("网络文件连接已经关闭")
    }

    n.isOpen = false
    fmt.Printf("网络文件连接已关闭: %s\n", n.url)
    return nil
}

func (n *NetworkFile) GetSize() int64 {
    return int64(len(n.data))
}

func (n *NetworkFile) GetPath() string {
    return n.url
}

func (n *NetworkFile) IsOpen() bool {
    return n.isOpen
}

// 文件管理器
type FileManager struct {
    files map[string]FileOperations
}

func NewFileManager() *FileManager {
    return &FileManager{
        files: make(map[string]FileOperations),
    }
}

func (fm *FileManager) RegisterFile(name string, file FileOperations) {
    fm.files[name] = file
    fmt.Printf("注册文件: %s -> %s\n", name, file.GetPath())
}

func (fm *FileManager) GetFile(name string) (FileOperations, error) {
    file, exists := fm.files[name]
    if !exists {
        return nil, fmt.Errorf("文件 '%s' 未找到", name)
    }
    return file, nil
}

func (fm *FileManager) ProcessFile(name string, processor func(FileOperations) error) error {
    file, err := fm.GetFile(name)
    if err != nil {
        return err
    }

    return processor(file)
}

func (fm *FileManager) ProcessAllFiles(processor func(string, FileOperations) error) {
    for name, file := range fm.files {
        if err := processor(name, file); err != nil {
            fmt.Printf("处理文件 '%s' 失败: %v\n", name, err)
        }
    }
}

// 文件操作工具
func CopyData(src Reader, dst Writer) (int64, error) {
    buffer := make([]byte, 1024)
    var totalCopied int64

    for {
        n, err := src.Read(buffer)
        if n > 0 {
            written, writeErr := dst.Write(buffer[:n])
            totalCopied += int64(written)

            if writeErr != nil {
                return totalCopied, writeErr
            }
        }

        if err != nil {
            if err.Error() == "已到达文件末尾" {
                break
            }
            return totalCopied, err
        }
    }

    return totalCopied, nil
}

// 使用示例
func DemonstrateInterfaceComposition() {
    fmt.Println("\n=== 接口组合演示 ===")

    // 创建文件管理器
    manager := NewFileManager()

    // 创建不同类型的文件
    memFile := NewMemoryFile("/tmp/memory.txt", []byte("Hello, Memory File!"))
    netFile := NewNetworkFile("https://example.com/data.txt")

    // 注册文件
    manager.RegisterFile("memory", memFile)
    manager.RegisterFile("network", netFile)

    // 演示接口组合的使用
    fmt.Println("\n=== 文件信息 ===")
    manager.ProcessAllFiles(func(name string, file FileOperations) error {
        fmt.Printf("文件: %s\n", name)
        fmt.Printf("  路径: %s\n", file.GetPath())
        fmt.Printf("  大小: %d 字节\n", file.GetSize())
        fmt.Printf("  状态: %v\n", file.IsOpen())
        return nil
    })

    // 演示读取操作
    fmt.Println("\n=== 读取操作 ===")
    manager.ProcessFile("memory", func(file FileOperations) error {
        buffer := make([]byte, 50)
        n, err := file.Read(buffer)
        if err != nil {
            return err
        }

        fmt.Printf("读取内容: %s\n", string(buffer[:n]))
        return nil
    })

    // 演示写入操作
    fmt.Println("\n=== 写入操作 ===")
    manager.ProcessFile("memory", func(file FileOperations) error {
        data := []byte(" 追加的内容!")
        n, err := file.Write(data)
        if err != nil {
            return err
        }

        fmt.Printf("写入了 %d 字节\n", n)
        return nil
    })

    // 演示定位操作
    fmt.Println("\n=== 定位操作 ===")
    manager.ProcessFile("memory", func(file FileOperations) error {
        // 移动到文件开始
        pos, err := file.Seek(0, 0)
        if err != nil {
            return err
        }

        fmt.Printf("当前位置: %d\n", pos)

        // 读取完整内容
        buffer := make([]byte, file.GetSize())
        n, err := file.Read(buffer)
        if err != nil && err.Error() != "已到达文件末尾" {
            return err
        }

        fmt.Printf("完整内容: %s\n", string(buffer[:n]))
        return nil
    })

    // 演示文件复制
    fmt.Println("\n=== 文件复制 ===")
    srcFile, _ := manager.GetFile("network")
    dstFile := NewMemoryFile("/tmp/copy.txt", nil)

    // 重置源文件位置
    srcFile.Seek(0, 0)

    copied, err := CopyData(srcFile, dstFile)
    if err != nil {
        fmt.Printf("复制失败: %v\n", err)
    } else {
        fmt.Printf("复制了 %d 字节\n", copied)

        // 读取复制的内容
        dstFile.Seek(0, 0)
        buffer := make([]byte, dstFile.GetSize())
        n, _ := dstFile.Read(buffer)
        fmt.Printf("复制的内容: %s\n", string(buffer[:n]))
    }

    // 关闭所有文件
    fmt.Println("\n=== 关闭文件 ===")
    manager.ProcessAllFiles(func(name string, file FileOperations) error {
        return file.Close()
    })

    dstFile.Close()
}
```

---

## 🎭 空接口和类型断言

空接口`interface{}`是Go语言中的特殊接口，可以表示任何类型。

```go
// 来自 mall-go/internal/types/empty_interface.go
package types

import (
    "fmt"
    "reflect"
    "strconv"
)

// 通用数据容器
type DataContainer struct {
    data map[string]interface{}
}

func NewDataContainer() *DataContainer {
    return &DataContainer{
        data: make(map[string]interface{}),
    }
}

func (dc *DataContainer) Set(key string, value interface{}) {
    dc.data[key] = value
    fmt.Printf("设置 %s = %v (%T)\n", key, value, value)
}

func (dc *DataContainer) Get(key string) (interface{}, bool) {
    value, exists := dc.data[key]
    return value, exists
}

// 类型断言方法
func (dc *DataContainer) GetString(key string) (string, error) {
    value, exists := dc.data[key]
    if !exists {
        return "", fmt.Errorf("键 '%s' 不存在", key)
    }

    // 类型断言
    if str, ok := value.(string); ok {
        return str, nil
    }

    return "", fmt.Errorf("键 '%s' 的值不是字符串类型，实际类型: %T", key, value)
}

func (dc *DataContainer) GetInt(key string) (int, error) {
    value, exists := dc.data[key]
    if !exists {
        return 0, fmt.Errorf("键 '%s' 不存在", key)
    }

    // 支持多种数字类型的转换
    switch v := value.(type) {
    case int:
        return v, nil
    case int32:
        return int(v), nil
    case int64:
        return int(v), nil
    case float32:
        return int(v), nil
    case float64:
        return int(v), nil
    case string:
        // 尝试从字符串解析
        if i, err := strconv.Atoi(v); err == nil {
            return i, nil
        }
        return 0, fmt.Errorf("无法将字符串 '%s' 转换为整数", v)
    default:
        return 0, fmt.Errorf("键 '%s' 的值无法转换为整数，实际类型: %T", key, value)
    }
}

func (dc *DataContainer) GetFloat(key string) (float64, error) {
    value, exists := dc.data[key]
    if !exists {
        return 0, fmt.Errorf("键 '%s' 不存在", key)
    }

    switch v := value.(type) {
    case float64:
        return v, nil
    case float32:
        return float64(v), nil
    case int:
        return float64(v), nil
    case int32:
        return float64(v), nil
    case int64:
        return float64(v), nil
    case string:
        if f, err := strconv.ParseFloat(v, 64); err == nil {
            return f, nil
        }
        return 0, fmt.Errorf("无法将字符串 '%s' 转换为浮点数", v)
    default:
        return 0, fmt.Errorf("键 '%s' 的值无法转换为浮点数，实际类型: %T", key, value)
    }
}

func (dc *DataContainer) GetBool(key string) (bool, error) {
    value, exists := dc.data[key]
    if !exists {
        return false, fmt.Errorf("键 '%s' 不存在", key)
    }

    switch v := value.(type) {
    case bool:
        return v, nil
    case string:
        if b, err := strconv.ParseBool(v); err == nil {
            return b, nil
        }
        return false, fmt.Errorf("无法将字符串 '%s' 转换为布尔值", v)
    case int:
        return v != 0, nil
    default:
        return false, fmt.Errorf("键 '%s' 的值无法转换为布尔值，实际类型: %T", key, value)
    }
}

// 类型开关示例
func (dc *DataContainer) ProcessValue(key string) {
    value, exists := dc.data[key]
    if !exists {
        fmt.Printf("键 '%s' 不存在\n", key)
        return
    }

    fmt.Printf("处理键 '%s' 的值: ", key)

    // 使用类型开关
    switch v := value.(type) {
    case string:
        fmt.Printf("字符串: '%s' (长度: %d)\n", v, len(v))
    case int:
        fmt.Printf("整数: %d (二进制: %b)\n", v, v)
    case int32:
        fmt.Printf("32位整数: %d\n", v)
    case int64:
        fmt.Printf("64位整数: %d\n", v)
    case float32:
        fmt.Printf("32位浮点数: %.2f\n", v)
    case float64:
        fmt.Printf("64位浮点数: %.2f\n", v)
    case bool:
        fmt.Printf("布尔值: %t\n", v)
    case []interface{}:
        fmt.Printf("切片: %v (长度: %d)\n", v, len(v))
    case map[string]interface{}:
        fmt.Printf("映射: %v (键数量: %d)\n", v, len(v))
    case nil:
        fmt.Println("空值")
    default:
        fmt.Printf("未知类型: %T, 值: %v\n", v, v)
    }
}

// 反射辅助方法
func (dc *DataContainer) GetTypeInfo(key string) {
    value, exists := dc.data[key]
    if !exists {
        fmt.Printf("键 '%s' 不存在\n", key)
        return
    }

    t := reflect.TypeOf(value)
    v := reflect.ValueOf(value)

    fmt.Printf("键 '%s' 的类型信息:\n", key)
    fmt.Printf("  类型: %s\n", t)
    fmt.Printf("  种类: %s\n", t.Kind())
    fmt.Printf("  值: %v\n", v)

    if t != nil {
        fmt.Printf("  包路径: %s\n", t.PkgPath())
        fmt.Printf("  类型名: %s\n", t.Name())
    }
}

// 安全的类型断言函数
func SafeTypeAssertion[T any](value interface{}) (T, bool) {
    if v, ok := value.(T); ok {
        return v, true
    }
    var zero T
    return zero, false
}

// 通用转换器
type TypeConverter struct{}

func NewTypeConverter() *TypeConverter {
    return &TypeConverter{}
}

func (tc *TypeConverter) ToString(value interface{}) string {
    if value == nil {
        return ""
    }

    switch v := value.(type) {
    case string:
        return v
    case int, int32, int64:
        return fmt.Sprintf("%d", v)
    case float32, float64:
        return fmt.Sprintf("%.2f", v)
    case bool:
        return strconv.FormatBool(v)
    default:
        return fmt.Sprintf("%v", v)
    }
}

func (tc *TypeConverter) ToInt(value interface{}) (int, error) {
    if value == nil {
        return 0, fmt.Errorf("值为nil")
    }

    switch v := value.(type) {
    case int:
        return v, nil
    case int32:
        return int(v), nil
    case int64:
        return int(v), nil
    case float32:
        return int(v), nil
    case float64:
        return int(v), nil
    case string:
        return strconv.Atoi(v)
    case bool:
        if v {
            return 1, nil
        }
        return 0, nil
    default:
        return 0, fmt.Errorf("无法将 %T 转换为int", value)
    }
}

func (tc *TypeConverter) ToFloat64(value interface{}) (float64, error) {
    if value == nil {
        return 0, fmt.Errorf("值为nil")
    }

    switch v := value.(type) {
    case float64:
        return v, nil
    case float32:
        return float64(v), nil
    case int:
        return float64(v), nil
    case int32:
        return float64(v), nil
    case int64:
        return float64(v), nil
    case string:
        return strconv.ParseFloat(v, 64)
    default:
        return 0, fmt.Errorf("无法将 %T 转换为float64", value)
    }
}

// 使用示例
func DemonstrateEmptyInterface() {
    fmt.Println("\n=== 空接口和类型断言演示 ===")

    container := NewDataContainer()
    converter := NewTypeConverter()

    // 存储不同类型的数据
    fmt.Println("1. 存储不同类型的数据:")
    container.Set("name", "张三")
    container.Set("age", 25)
    container.Set("height", 175.5)
    container.Set("married", false)
    container.Set("hobbies", []interface{}{"读书", "游泳", "编程"})
    container.Set("address", map[string]interface{}{
        "city":    "北京",
        "district": "朝阳区",
        "zipcode":  100000,
    })
    container.Set("score", int64(95))
    container.Set("empty", nil)

    // 类型断言获取数据
    fmt.Println("\n2. 类型断言获取数据:")

    if name, err := container.GetString("name"); err == nil {
        fmt.Printf("姓名: %s\n", name)
    } else {
        fmt.Printf("获取姓名失败: %v\n", err)
    }

    if age, err := container.GetInt("age"); err == nil {
        fmt.Printf("年龄: %d\n", age)
    } else {
        fmt.Printf("获取年龄失败: %v\n", err)
    }

    if height, err := container.GetFloat("height"); err == nil {
        fmt.Printf("身高: %.1f cm\n", height)
    } else {
        fmt.Printf("获取身高失败: %v\n", err)
    }

    if married, err := container.GetBool("married"); err == nil {
        fmt.Printf("已婚: %t\n", married)
    } else {
        fmt.Printf("获取婚姻状态失败: %v\n", err)
    }

    // 类型开关处理
    fmt.Println("\n3. 类型开关处理:")
    keys := []string{"name", "age", "height", "married", "hobbies", "address", "score", "empty"}

    for _, key := range keys {
        container.ProcessValue(key)
    }

    // 反射类型信息
    fmt.Println("\n4. 反射类型信息:")
    container.GetTypeInfo("address")
    container.GetTypeInfo("hobbies")

    // 安全类型断言
    fmt.Println("\n5. 安全类型断言:")
    if value, exists := container.Get("age"); exists {
        if age, ok := SafeTypeAssertion[int](value); ok {
            fmt.Printf("安全获取年龄: %d\n", age)
        } else {
            fmt.Printf("年龄不是int类型\n")
        }
    }

    // 通用转换器
    fmt.Println("\n6. 通用转换器:")
    testValues := []interface{}{
        "123",
        456,
        78.9,
        true,
        false,
        nil,
    }

    for _, value := range testValues {
        fmt.Printf("原值: %v (%T)\n", value, value)
        fmt.Printf("  -> 字符串: %s\n", converter.ToString(value))

        if intVal, err := converter.ToInt(value); err == nil {
            fmt.Printf("  -> 整数: %d\n", intVal)
        } else {
            fmt.Printf("  -> 整数转换失败: %v\n", err)
        }

        if floatVal, err := converter.ToFloat64(value); err == nil {
            fmt.Printf("  -> 浮点数: %.2f\n", floatVal)
        } else {
            fmt.Printf("  -> 浮点数转换失败: %v\n", err)
        }

        fmt.Println()
    }
}
```

---

## 🧪 接口测试和Mock

接口测试是Go语言测试的重要组成部分，Mock技术让单元测试更加灵活。

```go
// 来自 mall-go/internal/testing/interface_testing.go
package testing

import (
    "errors"
    "fmt"
    "testing"
    "time"
)

// 业务接口定义
type OrderRepository interface {
    GetOrder(id string) (*Order, error)
    SaveOrder(order *Order) error
    DeleteOrder(id string) error
    ListOrders(userID string) ([]*Order, error)
}

type PaymentService interface {
    ProcessPayment(orderID string, amount float64) (*PaymentResult, error)
    RefundPayment(transactionID string) error
    GetPaymentStatus(transactionID string) (string, error)
}

type NotificationService interface {
    SendOrderConfirmation(userID, orderID string) error
    SendPaymentNotification(userID string, amount float64) error
}

// 数据模型
type Order struct {
    ID       string    `json:"id"`
    UserID   string    `json:"user_id"`
    Amount   float64   `json:"amount"`
    Status   string    `json:"status"`
    CreateAt time.Time `json:"create_at"`
}

type PaymentResult struct {
    TransactionID string  `json:"transaction_id"`
    Status        string  `json:"status"`
    Amount        float64 `json:"amount"`
}

// 订单服务（被测试的业务逻辑）
type OrderService struct {
    orderRepo    OrderRepository
    paymentSvc   PaymentService
    notificationSvc NotificationService
}

func NewOrderService(orderRepo OrderRepository, paymentSvc PaymentService, notificationSvc NotificationService) *OrderService {
    return &OrderService{
        orderRepo:       orderRepo,
        paymentSvc:      paymentSvc,
        notificationSvc: notificationSvc,
    }
}

func (s *OrderService) CreateOrder(userID string, amount float64) (*Order, error) {
    if amount <= 0 {
        return nil, errors.New("订单金额必须大于0")
    }

    order := &Order{
        ID:       fmt.Sprintf("ORDER_%d", time.Now().UnixNano()),
        UserID:   userID,
        Amount:   amount,
        Status:   "pending",
        CreateAt: time.Now(),
    }

    // 保存订单
    if err := s.orderRepo.SaveOrder(order); err != nil {
        return nil, fmt.Errorf("保存订单失败: %w", err)
    }

    // 发送订单确认通知
    if err := s.notificationSvc.SendOrderConfirmation(userID, order.ID); err != nil {
        // 通知失败不影响订单创建
        fmt.Printf("发送订单确认通知失败: %v\n", err)
    }

    return order, nil
}

func (s *OrderService) PayOrder(orderID string) error {
    // 获取订单
    order, err := s.orderRepo.GetOrder(orderID)
    if err != nil {
        return fmt.Errorf("获取订单失败: %w", err)
    }

    if order.Status != "pending" {
        return fmt.Errorf("订单状态不正确: %s", order.Status)
    }

    // 处理支付
    paymentResult, err := s.paymentSvc.ProcessPayment(orderID, order.Amount)
    if err != nil {
        return fmt.Errorf("支付处理失败: %w", err)
    }

    // 更新订单状态
    if paymentResult.Status == "success" {
        order.Status = "paid"
    } else {
        order.Status = "payment_failed"
    }

    if err := s.orderRepo.SaveOrder(order); err != nil {
        return fmt.Errorf("更新订单状态失败: %w", err)
    }

    // 发送支付通知
    if paymentResult.Status == "success" {
        if err := s.notificationSvc.SendPaymentNotification(order.UserID, order.Amount); err != nil {
            fmt.Printf("发送支付通知失败: %v\n", err)
        }
    }

    return nil
}

// Mock实现
type MockOrderRepository struct {
    orders map[string]*Order
    saveError error
    getError  error
}

func NewMockOrderRepository() *MockOrderRepository {
    return &MockOrderRepository{
        orders: make(map[string]*Order),
    }
}

func (m *MockOrderRepository) GetOrder(id string) (*Order, error) {
    if m.getError != nil {
        return nil, m.getError
    }

    order, exists := m.orders[id]
    if !exists {
        return nil, errors.New("订单不存在")
    }

    return order, nil
}

func (m *MockOrderRepository) SaveOrder(order *Order) error {
    if m.saveError != nil {
        return m.saveError
    }

    m.orders[order.ID] = order
    return nil
}

func (m *MockOrderRepository) DeleteOrder(id string) error {
    delete(m.orders, id)
    return nil
}

func (m *MockOrderRepository) ListOrders(userID string) ([]*Order, error) {
    var orders []*Order
    for _, order := range m.orders {
        if order.UserID == userID {
            orders = append(orders, order)
        }
    }
    return orders, nil
}

// Mock设置方法
func (m *MockOrderRepository) SetSaveError(err error) {
    m.saveError = err
}

func (m *MockOrderRepository) SetGetError(err error) {
    m.getError = err
}

func (m *MockOrderRepository) GetOrderCount() int {
    return len(m.orders)
}

type MockPaymentService struct {
    processError error
    refundError  error
    paymentResult *PaymentResult
}

func NewMockPaymentService() *MockPaymentService {
    return &MockPaymentService{
        paymentResult: &PaymentResult{
            TransactionID: "TXN_123456",
            Status:        "success",
            Amount:        0,
        },
    }
}

func (m *MockPaymentService) ProcessPayment(orderID string, amount float64) (*PaymentResult, error) {
    if m.processError != nil {
        return nil, m.processError
    }

    result := *m.paymentResult
    result.Amount = amount
    return &result, nil
}

func (m *MockPaymentService) RefundPayment(transactionID string) error {
    return m.refundError
}

func (m *MockPaymentService) GetPaymentStatus(transactionID string) (string, error) {
    return m.paymentResult.Status, nil
}

// Mock设置方法
func (m *MockPaymentService) SetProcessError(err error) {
    m.processError = err
}

func (m *MockPaymentService) SetPaymentResult(result *PaymentResult) {
    m.paymentResult = result
}

type MockNotificationService struct {
    confirmationError error
    paymentError      error
    sentNotifications []string
}

func NewMockNotificationService() *MockNotificationService {
    return &MockNotificationService{
        sentNotifications: make([]string, 0),
    }
}

func (m *MockNotificationService) SendOrderConfirmation(userID, orderID string) error {
    if m.confirmationError != nil {
        return m.confirmationError
    }

    m.sentNotifications = append(m.sentNotifications, fmt.Sprintf("order_confirmation_%s_%s", userID, orderID))
    return nil
}

func (m *MockNotificationService) SendPaymentNotification(userID string, amount float64) error {
    if m.paymentError != nil {
        return m.paymentError
    }

    m.sentNotifications = append(m.sentNotifications, fmt.Sprintf("payment_notification_%s_%.2f", userID, amount))
    return nil
}

// Mock设置方法
func (m *MockNotificationService) SetConfirmationError(err error) {
    m.confirmationError = err
}

func (m *MockNotificationService) SetPaymentError(err error) {
    m.paymentError = err
}

func (m *MockNotificationService) GetSentNotifications() []string {
    return m.sentNotifications
}

// 测试用例
func TestOrderService_CreateOrder(t *testing.T) {
    // 准备Mock对象
    mockRepo := NewMockOrderRepository()
    mockPayment := NewMockPaymentService()
    mockNotification := NewMockNotificationService()

    service := NewOrderService(mockRepo, mockPayment, mockNotification)

    // 测试用例1：正常创建订单
    t.Run("正常创建订单", func(t *testing.T) {
        order, err := service.CreateOrder("user123", 99.99)

        if err != nil {
            t.Errorf("创建订单失败: %v", err)
        }

        if order == nil {
            t.Error("订单不应该为nil")
        }

        if order.UserID != "user123" {
            t.Errorf("用户ID不匹配: 期望 user123, 实际 %s", order.UserID)
        }

        if order.Amount != 99.99 {
            t.Errorf("订单金额不匹配: 期望 99.99, 实际 %.2f", order.Amount)
        }

        if order.Status != "pending" {
            t.Errorf("订单状态不匹配: 期望 pending, 实际 %s", order.Status)
        }

        // 验证订单已保存
        if mockRepo.GetOrderCount() != 1 {
            t.Errorf("订单数量不匹配: 期望 1, 实际 %d", mockRepo.GetOrderCount())
        }

        // 验证通知已发送
        notifications := mockNotification.GetSentNotifications()
        if len(notifications) != 1 {
            t.Errorf("通知数量不匹配: 期望 1, 实际 %d", len(notifications))
        }
    })

    // 测试用例2：无效金额
    t.Run("无效金额", func(t *testing.T) {
        order, err := service.CreateOrder("user123", -10)

        if err == nil {
            t.Error("应该返回错误")
        }

        if order != nil {
            t.Error("订单应该为nil")
        }
    })

    // 测试用例3：保存订单失败
    t.Run("保存订单失败", func(t *testing.T) {
        mockRepo.SetSaveError(errors.New("数据库连接失败"))

        order, err := service.CreateOrder("user123", 50.0)

        if err == nil {
            t.Error("应该返回错误")
        }

        if order != nil {
            t.Error("订单应该为nil")
        }

        // 重置错误
        mockRepo.SetSaveError(nil)
    })
}

func TestOrderService_PayOrder(t *testing.T) {
    // 准备Mock对象
    mockRepo := NewMockOrderRepository()
    mockPayment := NewMockPaymentService()
    mockNotification := NewMockNotificationService()

    service := NewOrderService(mockRepo, mockPayment, mockNotification)

    // 准备测试数据
    testOrder := &Order{
        ID:     "ORDER_123",
        UserID: "user123",
        Amount: 100.0,
        Status: "pending",
    }
    mockRepo.SaveOrder(testOrder)

    // 测试用例1：正常支付
    t.Run("正常支付", func(t *testing.T) {
        err := service.PayOrder("ORDER_123")

        if err != nil {
            t.Errorf("支付失败: %v", err)
        }

        // 验证订单状态已更新
        order, _ := mockRepo.GetOrder("ORDER_123")
        if order.Status != "paid" {
            t.Errorf("订单状态不匹配: 期望 paid, 实际 %s", order.Status)
        }

        // 验证支付通知已发送
        notifications := mockNotification.GetSentNotifications()
        found := false
        for _, notification := range notifications {
            if notification == "payment_notification_user123_100.00" {
                found = true
                break
            }
        }
        if !found {
            t.Error("支付通知未发送")
        }
    })

    // 测试用例2：订单不存在
    t.Run("订单不存在", func(t *testing.T) {
        err := service.PayOrder("NONEXISTENT")

        if err == nil {
            t.Error("应该返回错误")
        }
    })

    // 测试用例3：支付失败
    t.Run("支付失败", func(t *testing.T) {
        // 重置订单状态
        testOrder.Status = "pending"
        mockRepo.SaveOrder(testOrder)

        // 设置支付失败
        mockPayment.SetPaymentResult(&PaymentResult{
            TransactionID: "TXN_FAILED",
            Status:        "failed",
            Amount:        100.0,
        })

        err := service.PayOrder("ORDER_123")

        if err != nil {
            t.Errorf("支付处理不应该返回错误: %v", err)
        }

        // 验证订单状态
        order, _ := mockRepo.GetOrder("ORDER_123")
        if order.Status != "payment_failed" {
            t.Errorf("订单状态不匹配: 期望 payment_failed, 实际 %s", order.Status)
        }
    })
}

// 基准测试
func BenchmarkOrderService_CreateOrder(b *testing.B) {
    mockRepo := NewMockOrderRepository()
    mockPayment := NewMockPaymentService()
    mockNotification := NewMockNotificationService()

    service := NewOrderService(mockRepo, mockPayment, mockNotification)

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        userID := fmt.Sprintf("user%d", i)
        amount := float64(i%1000 + 1)

        _, err := service.CreateOrder(userID, amount)
        if err != nil {
            b.Errorf("创建订单失败: %v", err)
        }
    }
}

// 使用示例
func DemonstrateInterfaceTesting() {
    fmt.Println("\n=== 接口测试和Mock演示 ===")

    // 创建Mock对象
    mockRepo := NewMockOrderRepository()
    mockPayment := NewMockPaymentService()
    mockNotification := NewMockNotificationService()

    // 创建服务
    service := NewOrderService(mockRepo, mockPayment, mockNotification)

    fmt.Println("1. 正常业务流程测试:")

    // 创建订单
    order, err := service.CreateOrder("user123", 199.99)
    if err != nil {
        fmt.Printf("❌ 创建订单失败: %v\n", err)
    } else {
        fmt.Printf("✅ 订单创建成功: %s\n", order.ID)
    }

    // 支付订单
    if order != nil {
        err = service.PayOrder(order.ID)
        if err != nil {
            fmt.Printf("❌ 支付失败: %v\n", err)
        } else {
            fmt.Printf("✅ 支付成功\n")
        }
    }

    fmt.Println("\n2. 异常情况测试:")

    // 测试保存失败
    mockRepo.SetSaveError(errors.New("数据库连接失败"))
    _, err = service.CreateOrder("user456", 99.99)
    if err != nil {
        fmt.Printf("✅ 正确处理保存失败: %v\n", err)
    }

    // 重置错误
    mockRepo.SetSaveError(nil)

    // 测试支付失败
    mockPayment.SetProcessError(errors.New("支付网关不可用"))
    if order != nil {
        // 重置订单状态
        order.Status = "pending"
        mockRepo.SaveOrder(order)

        err = service.PayOrder(order.ID)
        if err != nil {
            fmt.Printf("✅ 正确处理支付失败: %v\n", err)
        }
    }

    fmt.Println("\n3. Mock状态检查:")
    fmt.Printf("订单数量: %d\n", mockRepo.GetOrderCount())
    fmt.Printf("发送的通知: %v\n", mockNotification.GetSentNotifications())
}
```

---

## 🏢 实战案例分析

让我们通过mall-go项目中的真实案例，深入理解企业级Go项目中的接口设计实践。

### 案例1：商品服务的接口设计

```go
// 来自 mall-go/internal/service/product_service.go
package service

import (
    "context"
    "fmt"
    "time"
)

// 商品核心接口定义
type ProductRepository interface {
    // 基础CRUD操作
    Create(ctx context.Context, product *Product) error
    GetByID(ctx context.Context, id int64) (*Product, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id int64) error

    // 查询操作
    List(ctx context.Context, filter *ProductFilter) ([]*Product, error)
    Search(ctx context.Context, keyword string, limit int) ([]*Product, error)
    GetByCategory(ctx context.Context, categoryID int64) ([]*Product, error)
}

// 缓存接口
type ProductCache interface {
    Get(ctx context.Context, key string) (*Product, error)
    Set(ctx context.Context, key string, product *Product, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context, pattern string) error
}

// 搜索引擎接口
type ProductSearchEngine interface {
    Index(ctx context.Context, product *Product) error
    Search(ctx context.Context, query *SearchQuery) (*SearchResult, error)
    Delete(ctx context.Context, productID int64) error
    BulkIndex(ctx context.Context, products []*Product) error
}

// 库存管理接口
type InventoryManager interface {
    CheckStock(ctx context.Context, productID int64, quantity int) (bool, error)
    ReserveStock(ctx context.Context, productID int64, quantity int) (*Reservation, error)
    ReleaseStock(ctx context.Context, reservationID string) error
    UpdateStock(ctx context.Context, productID int64, quantity int) error
}

// 价格计算接口
type PriceCalculator interface {
    CalculatePrice(ctx context.Context, productID int64, userID int64) (*PriceInfo, error)
    CalculateDiscount(ctx context.Context, productID int64, couponCode string) (*DiscountInfo, error)
    GetPriceHistory(ctx context.Context, productID int64, days int) ([]*PriceHistory, error)
}

// 数据模型
type Product struct {
    ID          int64     `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    CategoryID  int64     `json:"category_id" db:"category_id"`
    Price       float64   `json:"price" db:"price"`
    Stock       int       `json:"stock" db:"stock"`
    Status      string    `json:"status" db:"status"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ProductFilter struct {
    CategoryID *int64   `json:"category_id,omitempty"`
    PriceMin   *float64 `json:"price_min,omitempty"`
    PriceMax   *float64 `json:"price_max,omitempty"`
    Status     *string  `json:"status,omitempty"`
    Keyword    *string  `json:"keyword,omitempty"`
    Page       int      `json:"page"`
    PageSize   int      `json:"page_size"`
}

type SearchQuery struct {
    Keyword    string            `json:"keyword"`
    Filters    map[string]string `json:"filters"`
    SortBy     string            `json:"sort_by"`
    SortOrder  string            `json:"sort_order"`
    Page       int               `json:"page"`
    PageSize   int               `json:"page_size"`
}

type SearchResult struct {
    Products   []*Product `json:"products"`
    Total      int64      `json:"total"`
    Page       int        `json:"page"`
    PageSize   int        `json:"page_size"`
    Took       int64      `json:"took"` // 搜索耗时(ms)
}

type Reservation struct {
    ID        string    `json:"id"`
    ProductID int64     `json:"product_id"`
    Quantity  int       `json:"quantity"`
    ExpiresAt time.Time `json:"expires_at"`
}

type PriceInfo struct {
    OriginalPrice float64 `json:"original_price"`
    CurrentPrice  float64 `json:"current_price"`
    Discount      float64 `json:"discount"`
    Currency      string  `json:"currency"`
}

type DiscountInfo struct {
    CouponCode   string  `json:"coupon_code"`
    DiscountType string  `json:"discount_type"` // percentage, fixed
    DiscountValue float64 `json:"discount_value"`
    FinalPrice   float64 `json:"final_price"`
}

type PriceHistory struct {
    Date  time.Time `json:"date"`
    Price float64   `json:"price"`
}

// 商品服务实现
type ProductService struct {
    repo          ProductRepository
    cache         ProductCache
    searchEngine  ProductSearchEngine
    inventory     InventoryManager
    priceCalc     PriceCalculator
}

func NewProductService(
    repo ProductRepository,
    cache ProductCache,
    searchEngine ProductSearchEngine,
    inventory InventoryManager,
    priceCalc PriceCalculator,
) *ProductService {
    return &ProductService{
        repo:         repo,
        cache:        cache,
        searchEngine: searchEngine,
        inventory:    inventory,
        priceCalc:    priceCalc,
    }
}

// 创建商品 - 展示接口协作
func (s *ProductService) CreateProduct(ctx context.Context, product *Product) (*Product, error) {
    // 1. 数据验证
    if err := s.validateProduct(product); err != nil {
        return nil, fmt.Errorf("商品数据验证失败: %w", err)
    }

    // 2. 保存到数据库
    if err := s.repo.Create(ctx, product); err != nil {
        return nil, fmt.Errorf("保存商品失败: %w", err)
    }

    // 3. 添加到搜索引擎
    if err := s.searchEngine.Index(ctx, product); err != nil {
        // 搜索引擎失败不影响商品创建，记录日志
        fmt.Printf("添加商品到搜索引擎失败: %v\n", err)
    }

    // 4. 初始化库存
    if product.Stock > 0 {
        if err := s.inventory.UpdateStock(ctx, product.ID, product.Stock); err != nil {
            fmt.Printf("初始化库存失败: %v\n", err)
        }
    }

    // 5. 缓存商品信息
    cacheKey := fmt.Sprintf("product:%d", product.ID)
    if err := s.cache.Set(ctx, cacheKey, product, 30*time.Minute); err != nil {
        fmt.Printf("缓存商品信息失败: %v\n", err)
    }

    return product, nil
}

// 获取商品 - 展示缓存策略
func (s *ProductService) GetProduct(ctx context.Context, id int64, userID int64) (*ProductDetail, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("product:%d", id)
    if product, err := s.cache.Get(ctx, cacheKey); err == nil {
        return s.buildProductDetail(ctx, product, userID)
    }

    // 2. 从数据库获取
    product, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("获取商品失败: %w", err)
    }

    // 3. 更新缓存
    if err := s.cache.Set(ctx, cacheKey, product, 30*time.Minute); err != nil {
        fmt.Printf("更新缓存失败: %v\n", err)
    }

    return s.buildProductDetail(ctx, product, userID)
}

// 构建商品详情 - 展示接口组合使用
func (s *ProductService) buildProductDetail(ctx context.Context, product *Product, userID int64) (*ProductDetail, error) {
    detail := &ProductDetail{
        Product: product,
    }

    // 获取价格信息
    if priceInfo, err := s.priceCalc.CalculatePrice(ctx, product.ID, userID); err == nil {
        detail.PriceInfo = priceInfo
    }

    // 检查库存
    if available, err := s.inventory.CheckStock(ctx, product.ID, 1); err == nil {
        detail.Available = available
    }

    return detail, nil
}

// 搜索商品 - 展示搜索引擎集成
func (s *ProductService) SearchProducts(ctx context.Context, query *SearchQuery) (*SearchResult, error) {
    // 使用搜索引擎
    result, err := s.searchEngine.Search(ctx, query)
    if err != nil {
        // 搜索引擎失败，降级到数据库查询
        fmt.Printf("搜索引擎查询失败，降级到数据库: %v\n", err)
        return s.fallbackSearch(ctx, query)
    }

    return result, nil
}

// 降级搜索
func (s *ProductService) fallbackSearch(ctx context.Context, query *SearchQuery) (*SearchResult, error) {
    filter := &ProductFilter{
        Keyword:  &query.Keyword,
        Page:     query.Page,
        PageSize: query.PageSize,
    }

    products, err := s.repo.List(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("数据库查询失败: %w", err)
    }

    return &SearchResult{
        Products: products,
        Total:    int64(len(products)),
        Page:     query.Page,
        PageSize: query.PageSize,
        Took:     0,
    }, nil
}

func (s *ProductService) validateProduct(product *Product) error {
    if product.Name == "" {
        return fmt.Errorf("商品名称不能为空")
    }
    if product.Price < 0 {
        return fmt.Errorf("商品价格不能为负数")
    }
    if product.Stock < 0 {
        return fmt.Errorf("商品库存不能为负数")
    }
    return nil
}

type ProductDetail struct {
    *Product
    PriceInfo *PriceInfo `json:"price_info,omitempty"`
    Available bool       `json:"available"`
}
```

### 案例2：订单处理的接口设计

```go
// 来自 mall-go/internal/service/order_service.go
package service

import (
    "context"
    "fmt"
    "time"
)

// 订单处理相关接口
type OrderProcessor interface {
    ProcessOrder(ctx context.Context, order *Order) error
    CancelOrder(ctx context.Context, orderID string) error
    RefundOrder(ctx context.Context, orderID string, reason string) error
}

type PaymentProcessor interface {
    ProcessPayment(ctx context.Context, payment *PaymentRequest) (*PaymentResponse, error)
    RefundPayment(ctx context.Context, transactionID string, amount float64) error
    QueryPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatus, error)
}

type ShippingService interface {
    CreateShipment(ctx context.Context, order *Order) (*Shipment, error)
    TrackShipment(ctx context.Context, trackingNumber string) (*TrackingInfo, error)
    CancelShipment(ctx context.Context, shipmentID string) error
}

type NotificationSender interface {
    SendOrderConfirmation(ctx context.Context, order *Order) error
    SendPaymentNotification(ctx context.Context, payment *PaymentResponse) error
    SendShippingNotification(ctx context.Context, shipment *Shipment) error
}

// 事件发布接口
type EventPublisher interface {
    PublishOrderCreated(ctx context.Context, order *Order) error
    PublishOrderPaid(ctx context.Context, order *Order) error
    PublishOrderShipped(ctx context.Context, order *Order) error
    PublishOrderCompleted(ctx context.Context, order *Order) error
}

// 订单服务实现 - 展示复杂业务流程的接口协作
type OrderService struct {
    orderRepo    OrderRepository
    productSvc   *ProductService
    paymentProc  PaymentProcessor
    shippingSvc  ShippingService
    notifier     NotificationSender
    eventPub     EventPublisher
    inventory    InventoryManager
}

func NewOrderService(
    orderRepo OrderRepository,
    productSvc *ProductService,
    paymentProc PaymentProcessor,
    shippingSvc ShippingService,
    notifier NotificationSender,
    eventPub EventPublisher,
    inventory InventoryManager,
) *OrderService {
    return &OrderService{
        orderRepo:   orderRepo,
        productSvc:  productSvc,
        paymentProc: paymentProc,
        shippingSvc: shippingSvc,
        notifier:    notifier,
        eventPub:    eventPub,
        inventory:   inventory,
    }
}

// 创建订单 - 展示事务性操作的接口设计
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    // 1. 验证商品和库存
    var totalAmount float64
    var reservations []*Reservation

    for _, item := range req.Items {
        // 检查商品存在性
        product, err := s.productSvc.GetProduct(ctx, item.ProductID, req.UserID)
        if err != nil {
            return nil, fmt.Errorf("商品不存在: %w", err)
        }

        // 检查库存
        available, err := s.inventory.CheckStock(ctx, item.ProductID, item.Quantity)
        if err != nil || !available {
            return nil, fmt.Errorf("商品库存不足: 商品ID %d", item.ProductID)
        }

        // 预留库存
        reservation, err := s.inventory.ReserveStock(ctx, item.ProductID, item.Quantity)
        if err != nil {
            // 释放已预留的库存
            s.releaseReservations(ctx, reservations)
            return nil, fmt.Errorf("预留库存失败: %w", err)
        }
        reservations = append(reservations, reservation)

        totalAmount += product.Price * float64(item.Quantity)
    }

    // 2. 创建订单
    order := &Order{
        ID:          generateOrderID(),
        UserID:      req.UserID,
        Items:       req.Items,
        TotalAmount: totalAmount,
        Status:      "pending",
        CreatedAt:   time.Now(),
    }

    if err := s.orderRepo.Create(ctx, order); err != nil {
        s.releaseReservations(ctx, reservations)
        return nil, fmt.Errorf("创建订单失败: %w", err)
    }

    // 3. 发送订单确认通知
    if err := s.notifier.SendOrderConfirmation(ctx, order); err != nil {
        fmt.Printf("发送订单确认通知失败: %v\n", err)
    }

    // 4. 发布订单创建事件
    if err := s.eventPub.PublishOrderCreated(ctx, order); err != nil {
        fmt.Printf("发布订单创建事件失败: %v\n", err)
    }

    return order, nil
}

// 处理支付 - 展示状态机模式的接口设计
func (s *OrderService) ProcessPayment(ctx context.Context, orderID string, paymentMethod string) error {
    // 1. 获取订单
    order, err := s.orderRepo.GetByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("获取订单失败: %w", err)
    }

    if order.Status != "pending" {
        return fmt.Errorf("订单状态不正确: %s", order.Status)
    }

    // 2. 处理支付
    paymentReq := &PaymentRequest{
        OrderID:       orderID,
        Amount:        order.TotalAmount,
        PaymentMethod: paymentMethod,
        UserID:        order.UserID,
    }

    paymentResp, err := s.paymentProc.ProcessPayment(ctx, paymentReq)
    if err != nil {
        return fmt.Errorf("支付处理失败: %w", err)
    }

    // 3. 更新订单状态
    if paymentResp.Status == "success" {
        order.Status = "paid"
        order.PaymentID = paymentResp.TransactionID
        order.PaidAt = time.Now()

        if err := s.orderRepo.Update(ctx, order); err != nil {
            return fmt.Errorf("更新订单状态失败: %w", err)
        }

        // 4. 发送支付通知
        if err := s.notifier.SendPaymentNotification(ctx, paymentResp); err != nil {
            fmt.Printf("发送支付通知失败: %v\n", err)
        }

        // 5. 发布支付成功事件
        if err := s.eventPub.PublishOrderPaid(ctx, order); err != nil {
            fmt.Printf("发布支付成功事件失败: %v\n", err)
        }

        // 6. 自动创建发货单
        go s.createShipmentAsync(context.Background(), order)

    } else {
        order.Status = "payment_failed"
        s.orderRepo.Update(ctx, order)
    }

    return nil
}

// 异步创建发货单
func (s *OrderService) createShipmentAsync(ctx context.Context, order *Order) {
    shipment, err := s.shippingSvc.CreateShipment(ctx, order)
    if err != nil {
        fmt.Printf("创建发货单失败: %v\n", err)
        return
    }

    // 更新订单状态
    order.Status = "shipped"
    order.ShippingID = shipment.ID
    order.ShippedAt = time.Now()

    if err := s.orderRepo.Update(ctx, order); err != nil {
        fmt.Printf("更新订单发货状态失败: %v\n", err)
        return
    }

    // 发送发货通知
    if err := s.notifier.SendShippingNotification(ctx, shipment); err != nil {
        fmt.Printf("发送发货通知失败: %v\n", err)
    }

    // 发布发货事件
    if err := s.eventPub.PublishOrderShipped(ctx, order); err != nil {
        fmt.Printf("发布发货事件失败: %v\n", err)
    }
}

func (s *OrderService) releaseReservations(ctx context.Context, reservations []*Reservation) {
    for _, reservation := range reservations {
        if err := s.inventory.ReleaseStock(ctx, reservation.ID); err != nil {
            fmt.Printf("释放库存预留失败: %v\n", err)
        }
    }
}

func generateOrderID() string {
    return fmt.Sprintf("ORDER_%d", time.Now().UnixNano())
}

// 相关数据结构
type Order struct {
    ID          string       `json:"id"`
    UserID      int64        `json:"user_id"`
    Items       []*OrderItem `json:"items"`
    TotalAmount float64      `json:"total_amount"`
    Status      string       `json:"status"`
    PaymentID   string       `json:"payment_id,omitempty"`
    ShippingID  string       `json:"shipping_id,omitempty"`
    CreatedAt   time.Time    `json:"created_at"`
    PaidAt      time.Time    `json:"paid_at,omitempty"`
    ShippedAt   time.Time    `json:"shipped_at,omitempty"`
}

type OrderItem struct {
    ProductID int64 `json:"product_id"`
    Quantity  int   `json:"quantity"`
    Price     float64 `json:"price"`
}

type CreateOrderRequest struct {
    UserID int64        `json:"user_id"`
    Items  []*OrderItem `json:"items"`
}

type PaymentRequest struct {
    OrderID       string  `json:"order_id"`
    Amount        float64 `json:"amount"`
    PaymentMethod string  `json:"payment_method"`
    UserID        int64   `json:"user_id"`
}

type PaymentResponse struct {
    TransactionID string  `json:"transaction_id"`
    Status        string  `json:"status"`
    Amount        float64 `json:"amount"`
    Message       string  `json:"message"`
}

type PaymentStatus struct {
    TransactionID string `json:"transaction_id"`
    Status        string `json:"status"`
    Amount        float64 `json:"amount"`
}

type Shipment struct {
    ID             string `json:"id"`
    OrderID        string `json:"order_id"`
    TrackingNumber string `json:"tracking_number"`
    Carrier        string `json:"carrier"`
    Status         string `json:"status"`
}

type TrackingInfo struct {
    TrackingNumber string           `json:"tracking_number"`
    Status         string           `json:"status"`
    Events         []*TrackingEvent `json:"events"`
}

type TrackingEvent struct {
    Time        time.Time `json:"time"`
    Location    string    `json:"location"`
    Description string    `json:"description"`
}

type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
    Update(ctx context.Context, order *Order) error
    GetByUserID(ctx context.Context, userID int64) ([]*Order, error)
}
```

---

## 🎯 面试常考点

Go语言接口是面试中的高频考点，让我们深入分析常见问题和标准答案。

### 1. 接口的底层实现原理

**面试题：请解释Go语言接口的底层实现原理，包括iface和eface的区别。**

```go
// 来自 mall-go/docs/interview/interface_internals.go
package interview

import (
    "fmt"
    "unsafe"
)

/*
Go接口底层实现原理详解：

1. eface (empty interface)
   - 用于表示空接口 interface{}
   - 结构：type + data

2. iface (interface with methods)
   - 用于表示有方法的接口
   - 结构：tab + data

3. itab (interface table)
   - 包含接口类型信息和方法表
   - 用于动态分发方法调用
*/

// 模拟Go运行时的接口结构
type eface struct {
    _type *_type
    data  unsafe.Pointer
}

type iface struct {
    tab  *itab
    data unsafe.Pointer
}

type itab struct {
    inter *interfacetype
    _type *_type
    hash  uint32
    _     [4]byte
    fun   [1]uintptr // 方法表
}

type _type struct {
    size       uintptr
    ptrdata    uintptr
    hash       uint32
    tflag      uint8
    align      uint8
    fieldAlign uint8
    kind       uint8
    equal      func(unsafe.Pointer, unsafe.Pointer) bool
    gcdata     *byte
    str        int32
    ptrToThis  int32
}

type interfacetype struct {
    typ     _type
    pkgpath string
    mhdr    []imethod
}

type imethod struct {
    name int32
    ityp int32
}

// 演示接口赋值的内部过程
func DemonstrateInterfaceInternals() {
    fmt.Println("=== Go接口底层实现演示 ===")

    // 1. 空接口赋值
    fmt.Println("1. 空接口 (eface) 演示:")
    var empty interface{}

    // 赋值整数
    empty = 42
    fmt.Printf("empty = 42: type=%T, value=%v\n", empty, empty)

    // 赋值字符串
    empty = "hello"
    fmt.Printf("empty = \"hello\": type=%T, value=%v\n", empty, empty)

    // 赋值结构体
    type Person struct {
        Name string
        Age  int
    }
    empty = Person{Name: "张三", Age: 25}
    fmt.Printf("empty = Person{}: type=%T, value=%v\n", empty, empty)

    // 2. 有方法接口赋值
    fmt.Println("\n2. 有方法接口 (iface) 演示:")

    type Speaker interface {
        Speak() string
    }

    type Dog struct {
        Name string
    }

    func (d Dog) Speak() string {
        return fmt.Sprintf("%s: 汪汪!", d.Name)
    }

    type Cat struct {
        Name string
    }

    func (c Cat) Speak() string {
        return fmt.Sprintf("%s: 喵喵!", c.Name)
    }

    var speaker Speaker

    // 赋值Dog
    speaker = Dog{Name: "旺财"}
    fmt.Printf("speaker = Dog: %s\n", speaker.Speak())

    // 赋值Cat
    speaker = Cat{Name: "咪咪"}
    fmt.Printf("speaker = Cat: %s\n", speaker.Speak())

    // 3. 接口方法调用的动态分发
    fmt.Println("\n3. 动态分发演示:")
    animals := []Speaker{
        Dog{Name: "小黄"},
        Cat{Name: "小白"},
        Dog{Name: "大黑"},
    }

    for i, animal := range animals {
        fmt.Printf("动物%d: %s\n", i+1, animal.Speak())
    }
}

// 标准答案总结
func InterfaceInternalsAnswer() {
    fmt.Println(`
Go接口底层实现原理标准答案：

1. 数据结构：
   - eface：用于空接口，包含 _type 和 data
   - iface：用于有方法接口，包含 itab 和 data
   - itab：接口表，包含类型信息和方法表

2. 方法分发：
   - 编译时确定方法在itab中的位置
   - 运行时通过itab.fun[index]调用具体方法
   - 实现了多态和动态分发

3. 性能考虑：
   - 接口调用比直接调用慢（需要查表）
   - 空接口类型断言需要类型比较
   - 接口赋值可能涉及内存分配

4. 与其他语言对比：
   - Java：基于虚函数表，需要显式实现
   - Python：基于duck typing，运行时检查
   - Go：编译时检查+运行时分发，隐式实现
    `)
}
```

### 2. 接口与nil的关系

**面试题：解释Go语言中接口与nil的关系，为什么有时候接口不等于nil？**

```go
// 接口nil判断的陷阱
func DemonstrateInterfaceNil() {
    fmt.Println("=== 接口与nil关系演示 ===")

    // 1. 基本nil判断
    fmt.Println("1. 基本nil判断:")
    var i interface{}
    fmt.Printf("var i interface{}: i == nil? %t\n", i == nil)

    // 2. 接口赋值nil指针的陷阱
    fmt.Println("\n2. nil指针陷阱:")

    type MyError struct {
        msg string
    }

    func (e *MyError) Error() string {
        if e == nil {
            return "no error"
        }
        return e.msg
    }

    // 函数返回nil指针
    func getError() error {
        var err *MyError = nil
        return err // 这里返回的不是nil接口！
    }

    err := getError()
    fmt.Printf("err := getError(): err == nil? %t\n", err == nil)
    fmt.Printf("err的类型: %T\n", err)
    fmt.Printf("err的值: %v\n", err)

    // 正确的做法
    func getErrorCorrect() error {
        var err *MyError = nil
        if err != nil {
            return err
        }
        return nil // 返回真正的nil接口
    }

    err2 := getErrorCorrect()
    fmt.Printf("err2 := getErrorCorrect(): err2 == nil? %t\n", err2 == nil)

    // 3. 接口内部结构分析
    fmt.Println("\n3. 接口内部结构分析:")

    // nil接口：type=nil, data=nil
    var nilInterface interface{}
    fmt.Printf("nil接口: %+v\n", nilInterface)

    // 非nil接口：type=*MyError, data=nil
    var nilPointer *MyError = nil
    var nonNilInterface interface{} = nilPointer
    fmt.Printf("非nil接口(nil指针): %+v\n", nonNilInterface)
    fmt.Printf("类型检查: %T\n", nonNilInterface)

    // 4. 安全的nil检查方法
    fmt.Println("\n4. 安全的nil检查:")

    func isNil(i interface{}) bool {
        if i == nil {
            return true
        }

        // 使用反射检查
        v := reflect.ValueOf(i)
        switch v.Kind() {
        case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
            return v.IsNil()
        default:
            return false
        }
    }

    fmt.Printf("isNil(nilInterface): %t\n", isNil(nilInterface))
    fmt.Printf("isNil(nonNilInterface): %t\n", isNil(nonNilInterface))
    fmt.Printf("isNil(err): %t\n", isNil(err))
}

// 标准答案
func InterfaceNilAnswer() {
    fmt.Println(`
接口与nil关系标准答案：

1. 接口的nil判断：
   - 接口包含type和data两部分
   - 只有当type和data都为nil时，接口才等于nil
   - type不为nil但data为nil的接口不等于nil

2. 常见陷阱：
   - 返回nil指针给接口类型
   - 接口类型的零值判断
   - 类型断言时的nil检查

3. 解决方案：
   - 返回明确的nil而不是nil指针
   - 使用反射进行安全的nil检查
   - 理解接口的内部结构

4. 最佳实践：
   - 函数返回error时，成功时返回nil而不是nil指针
   - 使用类型断言前先检查接口是否为nil
   - 避免将nil指针赋值给接口类型
    `)
}
```

### 3. 接口断言的性能考虑

**面试题：类型断言和类型开关的性能如何？什么时候使用哪种方式？**

```go
// 性能测试和对比
func DemonstrateTypeAssertionPerformance() {
    fmt.Println("=== 类型断言性能演示 ===")

    // 准备测试数据
    values := []interface{}{
        "string1", "string2", "string3",
        123, 456, 789,
        12.34, 56.78, 90.12,
        true, false, true,
    }

    // 1. 类型断言方式
    fmt.Println("1. 类型断言方式:")
    start := time.Now()
    var stringCount, intCount, floatCount, boolCount int

    for i := 0; i < 1000000; i++ {
        for _, v := range values {
            if _, ok := v.(string); ok {
                stringCount++
            } else if _, ok := v.(int); ok {
                intCount++
            } else if _, ok := v.(float64); ok {
                floatCount++
            } else if _, ok := v.(bool); ok {
                boolCount++
            }
        }
    }

    duration1 := time.Since(start)
    fmt.Printf("类型断言耗时: %v\n", duration1)
    fmt.Printf("统计: string=%d, int=%d, float=%d, bool=%d\n",
        stringCount, intCount, floatCount, boolCount)

    // 2. 类型开关方式
    fmt.Println("\n2. 类型开关方式:")
    start = time.Now()
    stringCount, intCount, floatCount, boolCount = 0, 0, 0, 0

    for i := 0; i < 1000000; i++ {
        for _, v := range values {
            switch v.(type) {
            case string:
                stringCount++
            case int:
                intCount++
            case float64:
                floatCount++
            case bool:
                boolCount++
            }
        }
    }

    duration2 := time.Since(start)
    fmt.Printf("类型开关耗时: %v\n", duration2)
    fmt.Printf("统计: string=%d, int=%d, float=%d, bool=%d\n",
        stringCount, intCount, floatCount, boolCount)

    // 3. 反射方式
    fmt.Println("\n3. 反射方式:")
    start = time.Now()
    stringCount, intCount, floatCount, boolCount = 0, 0, 0, 0

    for i := 0; i < 1000000; i++ {
        for _, v := range values {
            t := reflect.TypeOf(v)
            switch t.Kind() {
            case reflect.String:
                stringCount++
            case reflect.Int:
                intCount++
            case reflect.Float64:
                floatCount++
            case reflect.Bool:
                boolCount++
            }
        }
    }

    duration3 := time.Since(start)
    fmt.Printf("反射方式耗时: %v\n", duration3)
    fmt.Printf("统计: string=%d, int=%d, float=%d, bool=%d\n",
        stringCount, intCount, floatCount, boolCount)

    // 性能对比
    fmt.Printf("\n性能对比:\n")
    fmt.Printf("类型断言: %v (基准)\n", duration1)
    fmt.Printf("类型开关: %v (%.2fx)\n", duration2, float64(duration2)/float64(duration1))
    fmt.Printf("反射方式: %v (%.2fx)\n", duration3, float64(duration3)/float64(duration1))
}

// 最佳实践示例
func TypeAssertionBestPractices() {
    fmt.Println("\n=== 类型断言最佳实践 ===")

    // 1. 安全的类型断言
    fmt.Println("1. 安全的类型断言:")

    var value interface{} = "hello world"

    // ❌ 不安全的断言（可能panic）
    // str := value.(string)

    // ✅ 安全的断言
    if str, ok := value.(string); ok {
        fmt.Printf("安全断言成功: %s\n", str)
    } else {
        fmt.Printf("断言失败，不是string类型\n")
    }

    // 2. 类型开关的优雅使用
    fmt.Println("\n2. 类型开关的优雅使用:")

    func processValue(v interface{}) string {
        switch val := v.(type) {
        case string:
            return fmt.Sprintf("字符串: %s (长度: %d)", val, len(val))
        case int:
            return fmt.Sprintf("整数: %d (二进制: %b)", val, val)
        case float64:
            return fmt.Sprintf("浮点数: %.2f", val)
        case bool:
            return fmt.Sprintf("布尔值: %t", val)
        case nil:
            return "空值"
        default:
            return fmt.Sprintf("未知类型: %T", val)
        }
    }

    testValues := []interface{}{
        "Go语言",
        42,
        3.14159,
        true,
        nil,
        []int{1, 2, 3},
    }

    for _, v := range testValues {
        fmt.Printf("处理结果: %s\n", processValue(v))
    }

    // 3. 泛型时代的类型断言
    fmt.Println("\n3. 泛型时代的类型断言:")

    // Go 1.18+ 泛型版本
    func SafeCast[T any](value interface{}) (T, bool) {
        if v, ok := value.(T); ok {
            return v, true
        }
        var zero T
        return zero, false
    }

    // 使用泛型版本
    if str, ok := SafeCast[string](value); ok {
        fmt.Printf("泛型断言成功: %s\n", str)
    }

    if num, ok := SafeCast[int](42); ok {
        fmt.Printf("泛型断言成功: %d\n", num)
    }
}

// 标准答案
func TypeAssertionPerformanceAnswer() {
    fmt.Println(`
类型断言性能标准答案：

1. 性能排序（从快到慢）：
   - 类型开关 (type switch) - 最快
   - 类型断言 (type assertion) - 中等
   - 反射 (reflection) - 最慢

2. 选择原则：
   - 多个类型判断：使用类型开关
   - 单个类型判断：使用类型断言
   - 复杂类型操作：考虑反射
   - 性能敏感场景：避免反射

3. 最佳实践：
   - 总是使用安全的类型断言 (value, ok := v.(Type))
   - 类型开关中使用变量接收值
   - 避免在热点代码中使用反射
   - 考虑使用泛型减少类型断言

4. 与其他语言对比：
   - Java：instanceof + 强制转换
   - Python：isinstance() + 鸭子类型
   - Go：编译时类型检查 + 运行时断言
    `)
}
```

---

## ⚠️ 踩坑提醒

在Go接口编程中，有一些常见的陷阱需要特别注意。

### 1. 接口nil判断陷阱

```go
// 来自 mall-go/docs/pitfalls/interface_pitfalls.go
package pitfalls

import (
    "fmt"
    "reflect"
)

// 陷阱1：接口nil判断
func PitfallInterfaceNil() {
    fmt.Println("=== 陷阱1：接口nil判断 ===")

    // ❌ 错误示例
    type MyError struct {
        Code    int
        Message string
    }

    func (e *MyError) Error() string {
        return fmt.Sprintf("错误%d: %s", e.Code, e.Message)
    }

    // 这个函数有问题！
    func badFunction() error {
        var err *MyError // nil指针
        // 一些逻辑...
        if someCondition := false; someCondition {
            err = &MyError{Code: 500, Message: "服务器错误"}
        }
        return err // 返回nil指针，但接口不为nil！
    }

    err := badFunction()
    fmt.Printf("err == nil? %t\n", err == nil) // false！
    fmt.Printf("err的类型: %T\n", err)
    fmt.Printf("err的值: %v\n", err)

    // 这会导致问题
    if err != nil {
        fmt.Printf("发生错误: %v\n", err) // 会执行，但实际没有错误
    }

    // ✅ 正确示例
    func goodFunction() error {
        var err *MyError
        // 一些逻辑...
        if someCondition := false; someCondition {
            err = &MyError{Code: 500, Message: "服务器错误"}
        }

        if err != nil {
            return err
        }
        return nil // 返回真正的nil接口
    }

    err2 := goodFunction()
    fmt.Printf("err2 == nil? %t\n", err2 == nil) // true

    // 解决方案：安全的nil检查
    func isReallyNil(i interface{}) bool {
        if i == nil {
            return true
        }

        v := reflect.ValueOf(i)
        switch v.Kind() {
        case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
            return v.IsNil()
        default:
            return false
        }
    }

    fmt.Printf("isReallyNil(err): %t\n", isReallyNil(err))
    fmt.Printf("isReallyNil(err2): %t\n", isReallyNil(err2))
}

// 陷阱2：空接口滥用
func PitfallEmptyInterfaceAbuse() {
    fmt.Println("\n=== 陷阱2：空接口滥用 ===")

    // ❌ 错误示例：过度使用空接口
    type BadConfig struct {
        Settings map[string]interface{} // 类型不明确
    }

    func (c *BadConfig) Get(key string) interface{} {
        return c.Settings[key] // 返回类型不明确
    }

    func (c *BadConfig) Set(key string, value interface{}) {
        c.Settings[key] = value // 可以设置任何类型
    }

    badConfig := &BadConfig{
        Settings: make(map[string]interface{}),
    }

    badConfig.Set("port", 8080)
    badConfig.Set("host", "localhost")
    badConfig.Set("debug", true)
    badConfig.Set("timeout", "30s") // 应该是time.Duration，但设置成了字符串

    // 使用时需要大量类型断言，容易出错
    if port, ok := badConfig.Get("port").(int); ok {
        fmt.Printf("端口: %d\n", port)
    }

    // 这里会出错，因为timeout是字符串而不是int
    if timeout, ok := badConfig.Get("timeout").(int); ok {
        fmt.Printf("超时: %d\n", timeout)
    } else {
        fmt.Printf("超时类型错误: %T\n", badConfig.Get("timeout"))
    }

    // ✅ 正确示例：使用具体类型
    type GoodConfig struct {
        Port    int           `json:"port"`
        Host    string        `json:"host"`
        Debug   bool          `json:"debug"`
        Timeout time.Duration `json:"timeout"`
    }

    func (c *GoodConfig) GetPort() int {
        return c.Port
    }

    func (c *GoodConfig) SetPort(port int) {
        c.Port = port
    }

    goodConfig := &GoodConfig{
        Port:    8080,
        Host:    "localhost",
        Debug:   true,
        Timeout: 30 * time.Second,
    }

    fmt.Printf("好的配置 - 端口: %d, 主机: %s\n", goodConfig.Port, goodConfig.Host)

    // 更好的方案：使用泛型（Go 1.18+）
    type TypedConfig[T any] struct {
        value T
    }

    func NewTypedConfig[T any](value T) *TypedConfig[T] {
        return &TypedConfig[T]{value: value}
    }

    func (c *TypedConfig[T]) Get() T {
        return c.value
    }

    func (c *TypedConfig[T]) Set(value T) {
        c.value = value
    }

    portConfig := NewTypedConfig(8080)
    fmt.Printf("类型安全配置 - 端口: %d\n", portConfig.Get())
}

// 陷阱3：接口设计过度抽象
func PitfallOverAbstraction() {
    fmt.Println("\n=== 陷阱3：接口设计过度抽象 ===")

    // ❌ 错误示例：过度抽象的接口
    type BadDataProcessor interface {
        Process(data interface{}) interface{}
        Transform(data interface{}, rules interface{}) interface{}
        Validate(data interface{}, schema interface{}) (bool, error)
        Serialize(data interface{}, format string) ([]byte, error)
        Deserialize(data []byte, format string) (interface{}, error)
        Cache(key string, data interface{}, ttl time.Duration) error
        GetFromCache(key string) (interface{}, error)
        Log(level string, message string, data interface{})
        Metrics(name string, value interface{})
        Notify(event string, data interface{}) error
    }

    // 这个接口违反了单一职责原则，太过庞大

    // ✅ 正确示例：职责分离的接口设计
    type DataProcessor interface {
        Process(data []byte) ([]byte, error)
    }

    type DataValidator interface {
        Validate(data []byte) error
    }

    type DataSerializer interface {
        Serialize(data interface{}) ([]byte, error)
        Deserialize(data []byte, target interface{}) error
    }

    type CacheManager interface {
        Set(key string, value []byte, ttl time.Duration) error
        Get(key string) ([]byte, error)
        Delete(key string) error
    }

    // 组合使用多个小接口
    type DataService struct {
        processor  DataProcessor
        validator  DataValidator
        serializer DataSerializer
        cache      CacheManager
    }

    func NewDataService(
        processor DataProcessor,
        validator DataValidator,
        serializer DataSerializer,
        cache CacheManager,
    ) *DataService {
        return &DataService{
            processor:  processor,
            validator:  validator,
            serializer: serializer,
            cache:      cache,
        }
    }

    func (s *DataService) ProcessData(data []byte) ([]byte, error) {
        // 验证数据
        if err := s.validator.Validate(data); err != nil {
            return nil, fmt.Errorf("数据验证失败: %w", err)
        }

        // 处理数据
        result, err := s.processor.Process(data)
        if err != nil {
            return nil, fmt.Errorf("数据处理失败: %w", err)
        }

        return result, nil
    }

    fmt.Println("接口职责分离设计完成")
}

// 陷阱4：类型断言性能问题
func PitfallTypeAssertionPerformance() {
    fmt.Println("\n=== 陷阱4：类型断言性能问题 ===")

    // ❌ 错误示例：频繁的类型断言
    func badProcessValues(values []interface{}) {
        for _, v := range values {
            // 每次都进行多次类型断言
            if str, ok := v.(string); ok {
                fmt.Printf("字符串: %s\n", str)
            } else if num, ok := v.(int); ok {
                fmt.Printf("整数: %d\n", num)
            } else if f, ok := v.(float64); ok {
                fmt.Printf("浮点数: %.2f\n", f)
            } else if b, ok := v.(bool); ok {
                fmt.Printf("布尔值: %t\n", b)
            }
        }
    }

    // ✅ 正确示例：使用类型开关
    func goodProcessValues(values []interface{}) {
        for _, v := range values {
            switch val := v.(type) {
            case string:
                fmt.Printf("字符串: %s\n", val)
            case int:
                fmt.Printf("整数: %d\n", val)
            case float64:
                fmt.Printf("浮点数: %.2f\n", val)
            case bool:
                fmt.Printf("布尔值: %t\n", val)
            default:
                fmt.Printf("未知类型: %T\n", val)
            }
        }
    }

    // 更好的方案：避免使用interface{}
    type Value struct {
        Type  string
        Data  interface{}
    }

    func (v Value) String() string {
        switch v.Type {
        case "string":
            return v.Data.(string)
        case "int":
            return fmt.Sprintf("%d", v.Data.(int))
        case "float":
            return fmt.Sprintf("%.2f", v.Data.(float64))
        case "bool":
            return fmt.Sprintf("%t", v.Data.(bool))
        default:
            return fmt.Sprintf("%v", v.Data)
        }
    }

    // 或者使用泛型（Go 1.18+）
    type TypedValue[T any] struct {
        Value T
    }

    func (tv TypedValue[T]) String() string {
        return fmt.Sprintf("%v", tv.Value)
    }

    // 测试数据
    testValues := []interface{}{"hello", 42, 3.14, true}

    fmt.Println("使用类型开关处理:")
    goodProcessValues(testValues)
}

// 陷阱5：接口方法签名设计问题
func PitfallInterfaceMethodSignature() {
    fmt.Println("\n=== 陷阱5：接口方法签名设计问题 ===")

    // ❌ 错误示例：方法签名不一致
    type BadReader interface {
        Read() ([]byte, error)        // 没有参数
        ReadN(n int) ([]byte, error)  // 有参数
        ReadAll() []byte              // 没有错误返回
    }

    type BadWriter interface {
        Write(data []byte) error      // 没有返回写入字节数
        WriteString(s string) (int, error) // 返回字节数
    }

    // ✅ 正确示例：一致的方法签名
    type GoodReader interface {
        Read(p []byte) (n int, err error) // 标准io.Reader签名
    }

    type GoodWriter interface {
        Write(p []byte) (n int, err error) // 标准io.Writer签名
    }

    // 组合接口
    type ReadWriter interface {
        GoodReader
        GoodWriter
    }

    // 实现示例
    type Buffer struct {
        data []byte
        pos  int
    }

    func (b *Buffer) Read(p []byte) (n int, err error) {
        if b.pos >= len(b.data) {
            return 0, fmt.Errorf("EOF")
        }

        n = copy(p, b.data[b.pos:])
        b.pos += n
        return n, nil
    }

    func (b *Buffer) Write(p []byte) (n int, err error) {
        b.data = append(b.data, p...)
        return len(p), nil
    }

    // 使用标准接口的好处
    func processData(rw ReadWriter) error {
        // 可以与任何实现了ReadWriter的类型一起工作
        buffer := make([]byte, 1024)
        n, err := rw.Read(buffer)
        if err != nil {
            return err
        }

        _, err = rw.Write(buffer[:n])
        return err
    }

    buf := &Buffer{data: []byte("hello world")}
    if err := processData(buf); err != nil {
        fmt.Printf("处理数据失败: %v\n", err)
    } else {
        fmt.Println("数据处理成功")
    }
}

// 踩坑总结
func InterfacePitfallsSummary() {
    fmt.Println(`
=== Go接口编程踩坑总结 ===

1. 接口nil判断陷阱：
   ❌ 返回nil指针给接口类型
   ✅ 返回明确的nil或使用反射检查

2. 空接口滥用：
   ❌ 过度使用interface{}导致类型不安全
   ✅ 使用具体类型或泛型

3. 接口设计过度抽象：
   ❌ 单个接口包含太多方法
   ✅ 遵循单一职责原则，设计小接口

4. 类型断言性能问题：
   ❌ 频繁的类型断言
   ✅ 使用类型开关或避免interface{}

5. 方法签名不一致：
   ❌ 接口方法签名不统一
   ✅ 遵循Go标准库的设计模式

最佳实践：
- 接受接口，返回结构体
- 接口应该小而专一
- 优先使用标准库接口
- 避免过度抽象
- 注意接口的nil语义
    `)
}
```

---

## 📝 练习题

通过5道从基础到高级的练习题，巩固Go接口设计模式的知识。

### 练习题1：基础接口设计（⭐）

**题目描述：**
设计一个图形计算系统，包含不同的几何图形（圆形、矩形、三角形），每个图形都能计算面积和周长。要求使用接口设计，并实现一个图形管理器来统一处理所有图形。

```go
// 练习题1：基础接口设计
package exercises

import (
    "fmt"
    "math"
)

// 解答：
// 1. 定义图形接口
type Shape interface {
    Area() float64
    Perimeter() float64
    Name() string
}

// 2. 实现具体图形
type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

func (c Circle) Name() string {
    return "圆形"
}

type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

func (r Rectangle) Name() string {
    return "矩形"
}

type Triangle struct {
    A, B, C float64 // 三边长
}

func (t Triangle) Area() float64 {
    // 使用海伦公式
    s := (t.A + t.B + t.C) / 2
    return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Perimeter() float64 {
    return t.A + t.B + t.C
}

func (t Triangle) Name() string {
    return "三角形"
}

// 3. 图形管理器
type ShapeManager struct {
    shapes []Shape
}

func NewShapeManager() *ShapeManager {
    return &ShapeManager{
        shapes: make([]Shape, 0),
    }
}

func (sm *ShapeManager) AddShape(shape Shape) {
    sm.shapes = append(sm.shapes, shape)
    fmt.Printf("添加图形: %s\n", shape.Name())
}

func (sm *ShapeManager) CalculateTotalArea() float64 {
    var total float64
    for _, shape := range sm.shapes {
        total += shape.Area()
    }
    return total
}

func (sm *ShapeManager) CalculateTotalPerimeter() float64 {
    var total float64
    for _, shape := range sm.shapes {
        total += shape.Perimeter()
    }
    return total
}

func (sm *ShapeManager) PrintShapeInfo() {
    fmt.Println("=== 图形信息 ===")
    for i, shape := range sm.shapes {
        fmt.Printf("%d. %s - 面积: %.2f, 周长: %.2f\n",
            i+1, shape.Name(), shape.Area(), shape.Perimeter())
    }
    fmt.Printf("总面积: %.2f\n", sm.CalculateTotalArea())
    fmt.Printf("总周长: %.2f\n", sm.CalculateTotalPerimeter())
}

// 测试函数
func TestExercise1() {
    fmt.Println("=== 练习题1：基础接口设计 ===")

    manager := NewShapeManager()

    // 添加不同图形
    manager.AddShape(Circle{Radius: 5})
    manager.AddShape(Rectangle{Width: 4, Height: 6})
    manager.AddShape(Triangle{A: 3, B: 4, C: 5})

    // 打印信息
    manager.PrintShapeInfo()
}

/*
解析说明：
1. 接口设计：定义了Shape接口，包含面积、周长和名称方法
2. 多态实现：不同图形实现相同接口，但具体计算方式不同
3. 统一管理：ShapeManager通过接口统一处理所有图形
4. 扩展性：新增图形类型只需实现Shape接口即可

扩展思考：
- 如何添加3D图形（体积计算）？
- 如何实现图形的序列化和反序列化？
- 如何添加图形的绘制功能？
*/
```

### 练习题2：接口组合设计（⭐⭐）

**题目描述：**
设计一个文件处理系统，支持不同的文件操作（读取、写入、压缩、加密）。使用接口组合的方式，让不同的文件处理器可以灵活组合这些功能。

```go
// 练习题2：接口组合设计
package exercises

import (
    "fmt"
    "strings"
    "time"
)

// 解答：
// 1. 定义基础接口
type Reader interface {
    Read() ([]byte, error)
}

type Writer interface {
    Write(data []byte) error
}

type Compressor interface {
    Compress(data []byte) ([]byte, error)
    Decompress(data []byte) ([]byte, error)
}

type Encryptor interface {
    Encrypt(data []byte) ([]byte, error)
    Decrypt(data []byte) ([]byte, error)
}

// 2. 组合接口
type ReadWriter interface {
    Reader
    Writer
}

type CompressedReader interface {
    Reader
    Compressor
}

type EncryptedWriter interface {
    Writer
    Encryptor
}

type SecureFileProcessor interface {
    Reader
    Writer
    Compressor
    Encryptor
}

// 3. 具体实现
type TextFile struct {
    content string
    path    string
}

func NewTextFile(path string, content string) *TextFile {
    return &TextFile{
        content: content,
        path:    path,
    }
}

func (tf *TextFile) Read() ([]byte, error) {
    fmt.Printf("读取文件: %s\n", tf.path)
    return []byte(tf.content), nil
}

func (tf *TextFile) Write(data []byte) error {
    fmt.Printf("写入文件: %s\n", tf.path)
    tf.content = string(data)
    return nil
}

// 压缩器实现
type GzipCompressor struct{}

func (gc *GzipCompressor) Compress(data []byte) ([]byte, error) {
    fmt.Printf("GZIP压缩数据，原始大小: %d 字节\n", len(data))
    // 模拟压缩（实际应该使用gzip库）
    compressed := fmt.Sprintf("GZIP[%s]", string(data))
    return []byte(compressed), nil
}

func (gc *GzipCompressor) Decompress(data []byte) ([]byte, error) {
    fmt.Printf("GZIP解压数据\n")
    // 模拟解压
    content := string(data)
    if strings.HasPrefix(content, "GZIP[") && strings.HasSuffix(content, "]") {
        original := content[5 : len(content)-1]
        return []byte(original), nil
    }
    return data, nil
}

// 加密器实现
type AESEncryptor struct {
    key string
}

func NewAESEncryptor(key string) *AESEncryptor {
    return &AESEncryptor{key: key}
}

func (ae *AESEncryptor) Encrypt(data []byte) ([]byte, error) {
    fmt.Printf("AES加密数据，密钥: %s\n", ae.key)
    // 模拟加密（实际应该使用crypto/aes）
    encrypted := fmt.Sprintf("AES[%s]", string(data))
    return []byte(encrypted), nil
}

func (ae *AESEncryptor) Decrypt(data []byte) ([]byte, error) {
    fmt.Printf("AES解密数据\n")
    // 模拟解密
    content := string(data)
    if strings.HasPrefix(content, "AES[") && strings.HasSuffix(content, "]") {
        original := content[4 : len(content)-1]
        return []byte(original), nil
    }
    return data, nil
}

// 4. 组合文件处理器
type CompositeFileProcessor struct {
    file       ReadWriter
    compressor Compressor
    encryptor  Encryptor
}

func NewCompositeFileProcessor(file ReadWriter, compressor Compressor, encryptor Encryptor) *CompositeFileProcessor {
    return &CompositeFileProcessor{
        file:       file,
        compressor: compressor,
        encryptor:  encryptor,
    }
}

func (cfp *CompositeFileProcessor) Read() ([]byte, error) {
    // 读取 -> 解密 -> 解压
    data, err := cfp.file.Read()
    if err != nil {
        return nil, err
    }

    if cfp.encryptor != nil {
        data, err = cfp.encryptor.Decrypt(data)
        if err != nil {
            return nil, err
        }
    }

    if cfp.compressor != nil {
        data, err = cfp.compressor.Decompress(data)
        if err != nil {
            return nil, err
        }
    }

    return data, nil
}

func (cfp *CompositeFileProcessor) Write(data []byte) error {
    // 压缩 -> 加密 -> 写入
    processedData := data
    var err error

    if cfp.compressor != nil {
        processedData, err = cfp.compressor.Compress(processedData)
        if err != nil {
            return err
        }
    }

    if cfp.encryptor != nil {
        processedData, err = cfp.encryptor.Encrypt(processedData)
        if err != nil {
            return err
        }
    }

    return cfp.file.Write(processedData)
}

func (cfp *CompositeFileProcessor) Compress(data []byte) ([]byte, error) {
    if cfp.compressor != nil {
        return cfp.compressor.Compress(data)
    }
    return data, nil
}

func (cfp *CompositeFileProcessor) Decompress(data []byte) ([]byte, error) {
    if cfp.compressor != nil {
        return cfp.compressor.Decompress(data)
    }
    return data, nil
}

func (cfp *CompositeFileProcessor) Encrypt(data []byte) ([]byte, error) {
    if cfp.encryptor != nil {
        return cfp.encryptor.Encrypt(data)
    }
    return data, nil
}

func (cfp *CompositeFileProcessor) Decrypt(data []byte) ([]byte, error) {
    if cfp.encryptor != nil {
        return cfp.encryptor.Decrypt(data)
    }
    return data, nil
}

// 5. 文件处理工厂
type FileProcessorFactory struct{}

func (fpf *FileProcessorFactory) CreateSimpleProcessor(path string) ReadWriter {
    return NewTextFile(path, "")
}

func (fpf *FileProcessorFactory) CreateCompressedProcessor(path string) *CompositeFileProcessor {
    file := NewTextFile(path, "")
    compressor := &GzipCompressor{}
    return NewCompositeFileProcessor(file, compressor, nil)
}

func (fpf *FileProcessorFactory) CreateEncryptedProcessor(path string, key string) *CompositeFileProcessor {
    file := NewTextFile(path, "")
    encryptor := NewAESEncryptor(key)
    return NewCompositeFileProcessor(file, nil, encryptor)
}

func (fpf *FileProcessorFactory) CreateSecureProcessor(path string, key string) *CompositeFileProcessor {
    file := NewTextFile(path, "")
    compressor := &GzipCompressor{}
    encryptor := NewAESEncryptor(key)
    return NewCompositeFileProcessor(file, compressor, encryptor)
}

// 测试函数
func TestExercise2() {
    fmt.Println("=== 练习题2：接口组合设计 ===")

    factory := &FileProcessorFactory{}

    // 测试不同类型的处理器
    testData := []byte("这是一个测试文件的内容，包含一些重要的数据。")

    fmt.Println("1. 简单文件处理器:")
    simple := factory.CreateSimpleProcessor("simple.txt")
    simple.Write(testData)
    data, _ := simple.Read()
    fmt.Printf("读取结果: %s\n", string(data))

    fmt.Println("\n2. 压缩文件处理器:")
    compressed := factory.CreateCompressedProcessor("compressed.txt.gz")
    compressed.Write(testData)
    data, _ = compressed.Read()
    fmt.Printf("读取结果: %s\n", string(data))

    fmt.Println("\n3. 加密文件处理器:")
    encrypted := factory.CreateEncryptedProcessor("encrypted.txt", "mykey123")
    encrypted.Write(testData)
    data, _ = encrypted.Read()
    fmt.Printf("读取结果: %s\n", string(data))

    fmt.Println("\n4. 安全文件处理器（压缩+加密）:")
    secure := factory.CreateSecureProcessor("secure.txt", "secretkey")
    secure.Write(testData)
    data, _ = secure.Read()
    fmt.Printf("读取结果: %s\n", string(data))
}

/*
解析说明：
1. 接口分离：将文件操作分解为独立的小接口
2. 接口组合：通过嵌入创建复合接口
3. 装饰器模式：CompositeFileProcessor装饰基础文件操作
4. 工厂模式：FileProcessorFactory创建不同配置的处理器

扩展思考：
- 如何添加更多压缩算法（ZIP、LZ4等）？
- 如何实现流式处理大文件？
- 如何添加文件完整性校验？
- 如何支持异步文件操作？
*/
```

### 练习题3：观察者模式实现（⭐⭐⭐）

**题目描述：**
实现一个事件系统，支持订阅和发布机制。要求支持不同类型的事件，多个观察者可以订阅同一事件，并且支持异步通知。

```go
// 练习题3：观察者模式实现
package exercises

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// 解答：
// 1. 定义事件接口
type Event interface {
    GetType() string
    GetData() interface{}
    GetTimestamp() time.Time
}

// 2. 观察者接口
type Observer interface {
    OnEvent(ctx context.Context, event Event) error
    GetID() string
}

// 3. 事件发布者接口
type EventPublisher interface {
    Subscribe(eventType string, observer Observer) error
    Unsubscribe(eventType string, observerID string) error
    Publish(ctx context.Context, event Event) error
    PublishAsync(ctx context.Context, event Event) error
}

// 4. 具体事件实现
type BaseEvent struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
}

func NewBaseEvent(eventType string, data interface{}) *BaseEvent {
    return &BaseEvent{
        Type:      eventType,
        Data:      data,
        Timestamp: time.Now(),
    }
}

func (e *BaseEvent) GetType() string {
    return e.Type
}

func (e *BaseEvent) GetData() interface{} {
    return e.Data
}

func (e *BaseEvent) GetTimestamp() time.Time {
    return e.Timestamp
}

// 用户事件
type UserEvent struct {
    *BaseEvent
    UserID int64  `json:"user_id"`
    Action string `json:"action"`
}

func NewUserEvent(userID int64, action string, data interface{}) *UserEvent {
    return &UserEvent{
        BaseEvent: NewBaseEvent("user", data),
        UserID:    userID,
        Action:    action,
    }
}

// 订单事件
type OrderEvent struct {
    *BaseEvent
    OrderID string  `json:"order_id"`
    Status  string  `json:"status"`
    Amount  float64 `json:"amount"`
}

func NewOrderEvent(orderID, status string, amount float64, data interface{}) *OrderEvent {
    return &OrderEvent{
        BaseEvent: NewBaseEvent("order", data),
        OrderID:   orderID,
        Status:    status,
        Amount:    amount,
    }
}

// 5. 具体观察者实现
type EmailNotifier struct {
    ID       string
    Email    string
    Template string
}

func NewEmailNotifier(id, email, template string) *EmailNotifier {
    return &EmailNotifier{
        ID:       id,
        Email:    email,
        Template: template,
    }
}

func (en *EmailNotifier) OnEvent(ctx context.Context, event Event) error {
    fmt.Printf("[邮件通知] 发送到 %s: 事件类型=%s, 时间=%s\n",
        en.Email, event.GetType(), event.GetTimestamp().Format("15:04:05"))

    // 模拟邮件发送延迟
    time.Sleep(100 * time.Millisecond)

    switch e := event.(type) {
    case *UserEvent:
        fmt.Printf("  用户事件: 用户ID=%d, 动作=%s\n", e.UserID, e.Action)
    case *OrderEvent:
        fmt.Printf("  订单事件: 订单ID=%s, 状态=%s, 金额=%.2f\n", e.OrderID, e.Status, e.Amount)
    default:
        fmt.Printf("  通用事件: 数据=%v\n", e.GetData())
    }

    return nil
}

func (en *EmailNotifier) GetID() string {
    return en.ID
}

type SMSNotifier struct {
    ID    string
    Phone string
}

func NewSMSNotifier(id, phone string) *SMSNotifier {
    return &SMSNotifier{
        ID:    id,
        Phone: phone,
    }
}

func (sn *SMSNotifier) OnEvent(ctx context.Context, event Event) error {
    fmt.Printf("[短信通知] 发送到 %s: 事件类型=%s\n", sn.Phone, event.GetType())

    // 模拟短信发送延迟
    time.Sleep(50 * time.Millisecond)

    return nil
}

func (sn *SMSNotifier) GetID() string {
    return sn.ID
}

type LogObserver struct {
    ID     string
    Logger string
}

func NewLogObserver(id, logger string) *LogObserver {
    return &LogObserver{
        ID:     id,
        Logger: logger,
    }
}

func (lo *LogObserver) OnEvent(ctx context.Context, event Event) error {
    fmt.Printf("[日志记录] %s: [%s] 事件=%s, 数据=%v\n",
        lo.Logger, event.GetTimestamp().Format("2006-01-02 15:04:05"),
        event.GetType(), event.GetData())
    return nil
}

func (lo *LogObserver) GetID() string {
    return lo.ID
}

// 6. 事件总线实现
type EventBus struct {
    subscribers map[string][]Observer
    mutex       sync.RWMutex
    workerPool  chan struct{} // 限制并发数
}

func NewEventBus(maxWorkers int) *EventBus {
    return &EventBus{
        subscribers: make(map[string][]Observer),
        workerPool:  make(chan struct{}, maxWorkers),
    }
}

func (eb *EventBus) Subscribe(eventType string, observer Observer) error {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()

    // 检查是否已经订阅
    for _, obs := range eb.subscribers[eventType] {
        if obs.GetID() == observer.GetID() {
            return fmt.Errorf("观察者 %s 已经订阅了事件 %s", observer.GetID(), eventType)
        }
    }

    eb.subscribers[eventType] = append(eb.subscribers[eventType], observer)
    fmt.Printf("观察者 %s 订阅了事件 %s\n", observer.GetID(), eventType)
    return nil
}

func (eb *EventBus) Unsubscribe(eventType string, observerID string) error {
    eb.mutex.Lock()
    defer eb.mutex.Unlock()

    observers := eb.subscribers[eventType]
    for i, obs := range observers {
        if obs.GetID() == observerID {
            // 删除观察者
            eb.subscribers[eventType] = append(observers[:i], observers[i+1:]...)
            fmt.Printf("观察者 %s 取消订阅事件 %s\n", observerID, eventType)
            return nil
        }
    }

    return fmt.Errorf("观察者 %s 没有订阅事件 %s", observerID, eventType)
}

func (eb *EventBus) Publish(ctx context.Context, event Event) error {
    eb.mutex.RLock()
    observers := make([]Observer, len(eb.subscribers[event.GetType()]))
    copy(observers, eb.subscribers[event.GetType()])
    eb.mutex.RUnlock()

    fmt.Printf("发布事件: %s, 观察者数量: %d\n", event.GetType(), len(observers))

    // 同步通知所有观察者
    for _, observer := range observers {
        if err := observer.OnEvent(ctx, event); err != nil {
            fmt.Printf("观察者 %s 处理事件失败: %v\n", observer.GetID(), err)
        }
    }

    return nil
}

func (eb *EventBus) PublishAsync(ctx context.Context, event Event) error {
    eb.mutex.RLock()
    observers := make([]Observer, len(eb.subscribers[event.GetType()]))
    copy(observers, eb.subscribers[event.GetType()])
    eb.mutex.RUnlock()

    fmt.Printf("异步发布事件: %s, 观察者数量: %d\n", event.GetType(), len(observers))

    // 异步通知所有观察者
    var wg sync.WaitGroup
    for _, observer := range observers {
        wg.Add(1)
        go func(obs Observer) {
            defer wg.Done()

            // 获取工作池令牌
            eb.workerPool <- struct{}{}
            defer func() { <-eb.workerPool }()

            if err := obs.OnEvent(ctx, event); err != nil {
                fmt.Printf("观察者 %s 异步处理事件失败: %v\n", obs.GetID(), err)
            }
        }(observer)
    }

    // 可以选择等待所有观察者完成，或者立即返回
    go func() {
        wg.Wait()
        fmt.Printf("事件 %s 的所有异步处理完成\n", event.GetType())
    }()

    return nil
}

func (eb *EventBus) GetSubscriberCount(eventType string) int {
    eb.mutex.RLock()
    defer eb.mutex.RUnlock()
    return len(eb.subscribers[eventType])
}

// 7. 事件管理器
type EventManager struct {
    eventBus EventPublisher
    stats    map[string]int
    mutex    sync.Mutex
}

func NewEventManager(eventBus EventPublisher) *EventManager {
    return &EventManager{
        eventBus: eventBus,
        stats:    make(map[string]int),
    }
}

func (em *EventManager) RegisterObserver(eventType string, observer Observer) error {
    return em.eventBus.Subscribe(eventType, observer)
}

func (em *EventManager) UnregisterObserver(eventType string, observerID string) error {
    return em.eventBus.Unsubscribe(eventType, observerID)
}

func (em *EventManager) EmitEvent(ctx context.Context, event Event, async bool) error {
    em.mutex.Lock()
    em.stats[event.GetType()]++
    em.mutex.Unlock()

    if async {
        return em.eventBus.PublishAsync(ctx, event)
    }
    return em.eventBus.Publish(ctx, event)
}

func (em *EventManager) GetStats() map[string]int {
    em.mutex.Lock()
    defer em.mutex.Unlock()

    stats := make(map[string]int)
    for k, v := range em.stats {
        stats[k] = v
    }
    return stats
}

// 测试函数
func TestExercise3() {
    fmt.Println("=== 练习题3：观察者模式实现 ===")

    // 创建事件总线
    eventBus := NewEventBus(5) // 最多5个并发工作者
    manager := NewEventManager(eventBus)

    // 创建观察者
    emailNotifier := NewEmailNotifier("email1", "user@example.com", "default")
    smsNotifier := NewSMSNotifier("sms1", "13800138000")
    logObserver := NewLogObserver("log1", "SystemLogger")

    // 订阅事件
    manager.RegisterObserver("user", emailNotifier)
    manager.RegisterObserver("user", logObserver)
    manager.RegisterObserver("order", emailNotifier)
    manager.RegisterObserver("order", smsNotifier)
    manager.RegisterObserver("order", logObserver)

    ctx := context.Background()

    // 发布用户事件
    fmt.Println("\n1. 发布用户事件（同步）:")
    userEvent := NewUserEvent(12345, "login", map[string]string{"ip": "192.168.1.100"})
    manager.EmitEvent(ctx, userEvent, false)

    // 发布订单事件
    fmt.Println("\n2. 发布订单事件（异步）:")
    orderEvent := NewOrderEvent("ORDER_001", "created", 299.99, map[string]string{"payment": "pending"})
    manager.EmitEvent(ctx, orderEvent, true)

    // 等待异步处理完成
    time.Sleep(200 * time.Millisecond)

    // 取消订阅
    fmt.Println("\n3. 取消订阅:")
    manager.UnregisterObserver("user", "email1")

    // 再次发布用户事件
    fmt.Println("\n4. 再次发布用户事件:")
    userEvent2 := NewUserEvent(12346, "logout", map[string]string{"session": "expired"})
    manager.EmitEvent(ctx, userEvent2, false)

    // 显示统计信息
    fmt.Println("\n5. 事件统计:")
    stats := manager.GetStats()
    for eventType, count := range stats {
        fmt.Printf("事件类型 %s: 发布次数 %d\n", eventType, count)
    }
}

/*
解析说明：
1. 接口设计：定义了Event、Observer、EventPublisher接口
2. 事件类型：支持不同类型的事件，可扩展
3. 异步处理：支持同步和异步事件通知
4. 并发控制：使用工作池限制并发数量
5. 统计功能：记录事件发布统计信息

扩展思考：
- 如何实现事件持久化？
- 如何添加事件过滤功能？
- 如何实现事件重试机制？
- 如何支持分布式事件系统？
*/
```

### 练习题4：依赖注入容器（⭐⭐⭐⭐）

**题目描述：**
实现一个完整的依赖注入容器，支持构造函数注入、接口注入、单例模式、生命周期管理等功能。要求支持循环依赖检测和自动装配。

```go
// 练习题4：依赖注入容器
package exercises

import (
    "fmt"
    "reflect"
    "sync"
)

// 解答：
// 1. 定义依赖注入相关接口
type ServiceContainer interface {
    Register(name string, factory ServiceFactory, options ...ServiceOption) error
    RegisterSingleton(name string, factory ServiceFactory) error
    RegisterTransient(name string, factory ServiceFactory) error
    Resolve(name string) (interface{}, error)
    ResolveByType(serviceType reflect.Type) (interface{}, error)
    Build() error
}

type ServiceFactory func(container ServiceContainer) (interface{}, error)

type ServiceLifetime int

const (
    Transient ServiceLifetime = iota // 每次创建新实例
    Singleton                        // 单例模式
    Scoped                          // 作用域内单例
)

type ServiceDescriptor struct {
    Name        string
    ServiceType reflect.Type
    Factory     ServiceFactory
    Lifetime    ServiceLifetime
    Instance    interface{}
    Dependencies []string
}

type ServiceOption func(*ServiceDescriptor)

// 2. 服务选项
func WithLifetime(lifetime ServiceLifetime) ServiceOption {
    return func(sd *ServiceDescriptor) {
        sd.Lifetime = lifetime
    }
}

func WithDependencies(deps ...string) ServiceOption {
    return func(sd *ServiceDescriptor) {
        sd.Dependencies = deps
    }
}

func WithServiceType(serviceType reflect.Type) ServiceOption {
    return func(sd *ServiceDescriptor) {
        sd.ServiceType = serviceType
    }
}

// 3. 依赖注入容器实现
type DIContainer struct {
    services    map[string]*ServiceDescriptor
    instances   map[string]interface{}
    building    map[string]bool // 循环依赖检测
    mutex       sync.RWMutex
    built       bool
}

func NewDIContainer() *DIContainer {
    return &DIContainer{
        services:  make(map[string]*ServiceDescriptor),
        instances: make(map[string]interface{}),
        building:  make(map[string]bool),
    }
}

func (c *DIContainer) Register(name string, factory ServiceFactory, options ...ServiceOption) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    if c.built {
        return fmt.Errorf("容器已构建，无法注册新服务")
    }

    descriptor := &ServiceDescriptor{
        Name:     name,
        Factory:  factory,
        Lifetime: Transient, // 默认为瞬态
    }

    // 应用选项
    for _, option := range options {
        option(descriptor)
    }

    c.services[name] = descriptor
    fmt.Printf("注册服务: %s (生命周期: %v)\n", name, descriptor.Lifetime)
    return nil
}

func (c *DIContainer) RegisterSingleton(name string, factory ServiceFactory) error {
    return c.Register(name, factory, WithLifetime(Singleton))
}

func (c *DIContainer) RegisterTransient(name string, factory ServiceFactory) error {
    return c.Register(name, factory, WithLifetime(Transient))
}

func (c *DIContainer) Resolve(name string) (interface{}, error) {
    c.mutex.RLock()
    descriptor, exists := c.services[name]
    c.mutex.RUnlock()

    if !exists {
        return nil, fmt.Errorf("服务 '%s' 未注册", name)
    }

    return c.createInstance(descriptor)
}

func (c *DIContainer) ResolveByType(serviceType reflect.Type) (interface{}, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    for _, descriptor := range c.services {
        if descriptor.ServiceType == serviceType {
            return c.createInstance(descriptor)
        }
    }

    return nil, fmt.Errorf("未找到类型 %s 的服务", serviceType.String())
}

func (c *DIContainer) createInstance(descriptor *ServiceDescriptor) (interface{}, error) {
    // 检查循环依赖
    if c.building[descriptor.Name] {
        return nil, fmt.Errorf("检测到循环依赖: %s", descriptor.Name)
    }

    // 单例模式检查
    if descriptor.Lifetime == Singleton {
        c.mutex.RLock()
        if instance, exists := c.instances[descriptor.Name]; exists {
            c.mutex.RUnlock()
            return instance, nil
        }
        c.mutex.RUnlock()
    }

    // 标记正在构建
    c.building[descriptor.Name] = true
    defer delete(c.building, descriptor.Name)

    // 创建实例
    instance, err := descriptor.Factory(c)
    if err != nil {
        return nil, fmt.Errorf("创建服务 '%s' 失败: %w", descriptor.Name, err)
    }

    // 单例模式缓存
    if descriptor.Lifetime == Singleton {
        c.mutex.Lock()
        c.instances[descriptor.Name] = instance
        c.mutex.Unlock()
    }

    return instance, nil
}

func (c *DIContainer) Build() error {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    // 验证所有依赖
    for name, descriptor := range c.services {
        for _, dep := range descriptor.Dependencies {
            if _, exists := c.services[dep]; !exists {
                return fmt.Errorf("服务 '%s' 的依赖 '%s' 未注册", name, dep)
            }
        }
    }

    c.built = true
    fmt.Println("依赖注入容器构建完成")
    return nil
}

// 4. 示例服务接口和实现
type Logger interface {
    Log(message string)
    Error(message string)
}

type Database interface {
    Connect() error
    Query(sql string) ([]map[string]interface{}, error)
    Close() error
}

type UserRepository interface {
    GetUser(id int64) (*User, error)
    SaveUser(user *User) error
}

type UserService interface {
    CreateUser(name, email string) (*User, error)
    GetUser(id int64) (*User, error)
}

// 具体实现
type ConsoleLogger struct {
    prefix string
}

func (cl *ConsoleLogger) Log(message string) {
    fmt.Printf("[%s] INFO: %s\n", cl.prefix, message)
}

func (cl *ConsoleLogger) Error(message string) {
    fmt.Printf("[%s] ERROR: %s\n", cl.prefix, message)
}

type MySQLDatabase struct {
    connectionString string
    logger          Logger
}

func (db *MySQLDatabase) Connect() error {
    db.logger.Log("连接到MySQL数据库: " + db.connectionString)
    return nil
}

func (db *MySQLDatabase) Query(sql string) ([]map[string]interface{}, error) {
    db.logger.Log("执行SQL查询: " + sql)
    return []map[string]interface{}{
        {"id": 1, "name": "张三", "email": "zhangsan@example.com"},
    }, nil
}

func (db *MySQLDatabase) Close() error {
    db.logger.Log("关闭数据库连接")
    return nil
}

type DatabaseUserRepository struct {
    db     Database
    logger Logger
}

func (repo *DatabaseUserRepository) GetUser(id int64) (*User, error) {
    repo.logger.Log(fmt.Sprintf("获取用户: %d", id))

    results, err := repo.db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %d", id))
    if err != nil {
        return nil, err
    }

    if len(results) == 0 {
        return nil, fmt.Errorf("用户不存在")
    }

    result := results[0]
    return &User{
        ID:    int64(result["id"].(int)),
        Name:  result["name"].(string),
        Email: result["email"].(string),
    }, nil
}

func (repo *DatabaseUserRepository) SaveUser(user *User) error {
    repo.logger.Log(fmt.Sprintf("保存用户: %s", user.Name))

    sql := fmt.Sprintf("INSERT INTO users (name, email) VALUES ('%s', '%s')", user.Name, user.Email)
    _, err := repo.db.Query(sql)
    return err
}

type DefaultUserService struct {
    repo   UserRepository
    logger Logger
}

func (svc *DefaultUserService) CreateUser(name, email string) (*User, error) {
    svc.logger.Log(fmt.Sprintf("创建用户: %s", name))

    user := &User{
        ID:    time.Now().Unix(),
        Name:  name,
        Email: email,
    }

    if err := svc.repo.SaveUser(user); err != nil {
        svc.logger.Error(fmt.Sprintf("保存用户失败: %v", err))
        return nil, err
    }

    return user, nil
}

func (svc *DefaultUserService) GetUser(id int64) (*User, error) {
    svc.logger.Log(fmt.Sprintf("获取用户服务: %d", id))
    return svc.repo.GetUser(id)
}

type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 5. 服务配置
func ConfigureServices(container ServiceContainer) error {
    // 注册Logger（单例）
    err := container.RegisterSingleton("logger", func(c ServiceContainer) (interface{}, error) {
        return &ConsoleLogger{prefix: "APP"}, nil
    })
    if err != nil {
        return err
    }

    // 注册Database（单例）
    err = container.RegisterSingleton("database", func(c ServiceContainer) (interface{}, error) {
        logger, err := c.Resolve("logger")
        if err != nil {
            return nil, err
        }

        db := &MySQLDatabase{
            connectionString: "mysql://localhost:3306/mall",
            logger:          logger.(Logger),
        }

        if err := db.Connect(); err != nil {
            return nil, err
        }

        return db, nil
    })
    if err != nil {
        return err
    }

    // 注册UserRepository（单例）
    err = container.RegisterSingleton("userRepository", func(c ServiceContainer) (interface{}, error) {
        db, err := c.Resolve("database")
        if err != nil {
            return nil, err
        }

        logger, err := c.Resolve("logger")
        if err != nil {
            return nil, err
        }

        return &DatabaseUserRepository{
            db:     db.(Database),
            logger: logger.(Logger),
        }, nil
    })
    if err != nil {
        return err
    }

    // 注册UserService（瞬态）
    err = container.RegisterTransient("userService", func(c ServiceContainer) (interface{}, error) {
        repo, err := c.Resolve("userRepository")
        if err != nil {
            return nil, err
        }

        logger, err := c.Resolve("logger")
        if err != nil {
            return nil, err
        }

        return &DefaultUserService{
            repo:   repo.(UserRepository),
            logger: logger.(Logger),
        }, nil
    })
    if err != nil {
        return err
    }

    return container.Build()
}

// 测试函数
func TestExercise4() {
    fmt.Println("=== 练习题4：依赖注入容器 ===")

    container := NewDIContainer()

    // 配置服务
    if err := ConfigureServices(container); err != nil {
        fmt.Printf("配置服务失败: %v\n", err)
        return
    }

    fmt.Println("\n1. 解析用户服务:")
    userService, err := container.Resolve("userService")
    if err != nil {
        fmt.Printf("解析用户服务失败: %v\n", err)
        return
    }

    svc := userService.(UserService)

    // 创建用户
    fmt.Println("\n2. 创建用户:")
    user, err := svc.CreateUser("李四", "lisi@example.com")
    if err != nil {
        fmt.Printf("创建用户失败: %v\n", err)
        return
    }

    fmt.Printf("用户创建成功: %+v\n", user)

    // 获取用户
    fmt.Println("\n3. 获取用户:")
    retrievedUser, err := svc.GetUser(1)
    if err != nil {
        fmt.Printf("获取用户失败: %v\n", err)
        return
    }

    fmt.Printf("获取用户成功: %+v\n", retrievedUser)

    // 测试单例模式
    fmt.Println("\n4. 测试单例模式:")
    logger1, _ := container.Resolve("logger")
    logger2, _ := container.Resolve("logger")

    fmt.Printf("Logger实例相同: %t\n", logger1 == logger2)

    // 测试瞬态模式
    fmt.Println("\n5. 测试瞬态模式:")
    userService1, _ := container.Resolve("userService")
    userService2, _ := container.Resolve("userService")

    fmt.Printf("UserService实例相同: %t\n", userService1 == userService2)
}

/*
解析说明：
1. 容器设计：实现了完整的依赖注入容器
2. 生命周期管理：支持单例、瞬态等生命周期
3. 循环依赖检测：防止无限递归创建
4. 自动装配：通过工厂函数自动解析依赖
5. 类型安全：使用接口确保类型安全

扩展思考：
- 如何实现作用域生命周期？
- 如何支持泛型依赖注入？
- 如何实现配置文件驱动的服务注册？
- 如何添加服务健康检查？
*/
```

### 练习题5：插件系统架构（⭐⭐⭐⭐⭐）

**题目描述：**
设计一个完整的插件系统，支持动态加载插件、插件生命周期管理、插件间通信、配置管理等功能。要求支持热插拔和版本管理。

```go
// 练习题5：插件系统架构
package exercises

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// 解答：
// 1. 插件核心接口
type Plugin interface {
    GetInfo() *PluginInfo
    Initialize(ctx context.Context, config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Destroy(ctx context.Context) error
    GetStatus() PluginStatus
}

type PluginInfo struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Author      string            `json:"author"`
    Dependencies []string         `json:"dependencies"`
    Capabilities []string         `json:"capabilities"`
    Config      map[string]interface{} `json:"config"`
}

type PluginStatus int

const (
    StatusUnknown PluginStatus = iota
    StatusInitialized
    StatusStarted
    StatusStopped
    StatusError
)

func (s PluginStatus) String() string {
    switch s {
    case StatusInitialized:
        return "Initialized"
    case StatusStarted:
        return "Started"
    case StatusStopped:
        return "Stopped"
    case StatusError:
        return "Error"
    default:
        return "Unknown"
    }
}

// 2. 插件管理器接口
type PluginManager interface {
    RegisterPlugin(plugin Plugin) error
    UnregisterPlugin(name string) error
    StartPlugin(name string) error
    StopPlugin(name string) error
    GetPlugin(name string) (Plugin, error)
    ListPlugins() []*PluginInfo
    GetPluginStatus(name string) (PluginStatus, error)
}

// 3. 插件通信接口
type MessageBus interface {
    Subscribe(topic string, handler MessageHandler) error
    Unsubscribe(topic string, handlerID string) error
    Publish(topic string, message *Message) error
    PublishAsync(topic string, message *Message) error
}

type MessageHandler interface {
    GetID() string
    Handle(ctx context.Context, message *Message) error
}

type Message struct {
    ID        string                 `json:"id"`
    Topic     string                 `json:"topic"`
    Source    string                 `json:"source"`
    Target    string                 `json:"target"`
    Type      string                 `json:"type"`
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
}

// 4. 插件注册表
type PluginRegistry struct {
    plugins map[string]Plugin
    status  map[string]PluginStatus
    configs map[string]map[string]interface{}
    mutex   sync.RWMutex
}

func NewPluginRegistry() *PluginRegistry {
    return &PluginRegistry{
        plugins: make(map[string]Plugin),
        status:  make(map[string]PluginStatus),
        configs: make(map[string]map[string]interface{}),
    }
}

func (pr *PluginRegistry) RegisterPlugin(plugin Plugin) error {
    pr.mutex.Lock()
    defer pr.mutex.Unlock()

    info := plugin.GetInfo()

    // 检查插件是否已存在
    if _, exists := pr.plugins[info.Name]; exists {
        return fmt.Errorf("插件 %s 已经注册", info.Name)
    }

    // 检查依赖
    for _, dep := range info.Dependencies {
        if _, exists := pr.plugins[dep]; !exists {
            return fmt.Errorf("插件 %s 的依赖 %s 未找到", info.Name, dep)
        }
    }

    pr.plugins[info.Name] = plugin
    pr.status[info.Name] = StatusUnknown
    pr.configs[info.Name] = info.Config

    fmt.Printf("注册插件: %s v%s\n", info.Name, info.Version)
    return nil
}

func (pr *PluginRegistry) UnregisterPlugin(name string) error {
    pr.mutex.Lock()
    defer pr.mutex.Unlock()

    plugin, exists := pr.plugins[name]
    if !exists {
        return fmt.Errorf("插件 %s 未找到", name)
    }

    // 检查是否有其他插件依赖此插件
    for pluginName, p := range pr.plugins {
        if pluginName == name {
            continue
        }

        info := p.GetInfo()
        for _, dep := range info.Dependencies {
            if dep == name {
                return fmt.Errorf("插件 %s 被 %s 依赖，无法卸载", name, pluginName)
            }
        }
    }

    // 停止插件
    if pr.status[name] == StatusStarted {
        plugin.Stop(context.Background())
    }

    // 销毁插件
    plugin.Destroy(context.Background())

    delete(pr.plugins, name)
    delete(pr.status, name)
    delete(pr.configs, name)

    fmt.Printf("卸载插件: %s\n", name)
    return nil
}

func (pr *PluginRegistry) StartPlugin(name string) error {
    pr.mutex.Lock()
    defer pr.mutex.Unlock()

    plugin, exists := pr.plugins[name]
    if !exists {
        return fmt.Errorf("插件 %s 未找到", name)
    }

    if pr.status[name] == StatusStarted {
        return fmt.Errorf("插件 %s 已经启动", name)
    }

    ctx := context.Background()

    // 初始化插件
    if pr.status[name] == StatusUnknown {
        if err := plugin.Initialize(ctx, pr.configs[name]); err != nil {
            pr.status[name] = StatusError
            return fmt.Errorf("初始化插件 %s 失败: %w", name, err)
        }
        pr.status[name] = StatusInitialized
    }

    // 启动插件
    if err := plugin.Start(ctx); err != nil {
        pr.status[name] = StatusError
        return fmt.Errorf("启动插件 %s 失败: %w", name, err)
    }

    pr.status[name] = StatusStarted
    fmt.Printf("启动插件: %s\n", name)
    return nil
}

func (pr *PluginRegistry) StopPlugin(name string) error {
    pr.mutex.Lock()
    defer pr.mutex.Unlock()

    plugin, exists := pr.plugins[name]
    if !exists {
        return fmt.Errorf("插件 %s 未找到", name)
    }

    if pr.status[name] != StatusStarted {
        return fmt.Errorf("插件 %s 未启动", name)
    }

    if err := plugin.Stop(context.Background()); err != nil {
        pr.status[name] = StatusError
        return fmt.Errorf("停止插件 %s 失败: %w", name, err)
    }

    pr.status[name] = StatusStopped
    fmt.Printf("停止插件: %s\n", name)
    return nil
}

func (pr *PluginRegistry) GetPlugin(name string) (Plugin, error) {
    pr.mutex.RLock()
    defer pr.mutex.RUnlock()

    plugin, exists := pr.plugins[name]
    if !exists {
        return nil, fmt.Errorf("插件 %s 未找到", name)
    }

    return plugin, nil
}

func (pr *PluginRegistry) ListPlugins() []*PluginInfo {
    pr.mutex.RLock()
    defer pr.mutex.RUnlock()

    var infos []*PluginInfo
    for _, plugin := range pr.plugins {
        infos = append(infos, plugin.GetInfo())
    }

    return infos
}

func (pr *PluginRegistry) GetPluginStatus(name string) (PluginStatus, error) {
    pr.mutex.RLock()
    defer pr.mutex.RUnlock()

    status, exists := pr.status[name]
    if !exists {
        return StatusUnknown, fmt.Errorf("插件 %s 未找到", name)
    }

    return status, nil
}

// 5. 消息总线实现
type SimpleMsgBus struct {
    handlers map[string][]MessageHandler
    mutex    sync.RWMutex
}

func NewSimpleMsgBus() *SimpleMsgBus {
    return &SimpleMsgBus{
        handlers: make(map[string][]MessageHandler),
    }
}

func (mb *SimpleMsgBus) Subscribe(topic string, handler MessageHandler) error {
    mb.mutex.Lock()
    defer mb.mutex.Unlock()

    mb.handlers[topic] = append(mb.handlers[topic], handler)
    fmt.Printf("订阅主题: %s, 处理器: %s\n", topic, handler.GetID())
    return nil
}

func (mb *SimpleMsgBus) Unsubscribe(topic string, handlerID string) error {
    mb.mutex.Lock()
    defer mb.mutex.Unlock()

    handlers := mb.handlers[topic]
    for i, handler := range handlers {
        if handler.GetID() == handlerID {
            mb.handlers[topic] = append(handlers[:i], handlers[i+1:]...)
            fmt.Printf("取消订阅主题: %s, 处理器: %s\n", topic, handlerID)
            return nil
        }
    }

    return fmt.Errorf("处理器 %s 未订阅主题 %s", handlerID, topic)
}

func (mb *SimpleMsgBus) Publish(topic string, message *Message) error {
    mb.mutex.RLock()
    handlers := make([]MessageHandler, len(mb.handlers[topic]))
    copy(handlers, mb.handlers[topic])
    mb.mutex.RUnlock()

    fmt.Printf("发布消息到主题: %s, 处理器数量: %d\n", topic, len(handlers))

    for _, handler := range handlers {
        if err := handler.Handle(context.Background(), message); err != nil {
            fmt.Printf("处理器 %s 处理消息失败: %v\n", handler.GetID(), err)
        }
    }

    return nil
}

func (mb *SimpleMsgBus) PublishAsync(topic string, message *Message) error {
    mb.mutex.RLock()
    handlers := make([]MessageHandler, len(mb.handlers[topic]))
    copy(handlers, mb.handlers[topic])
    mb.mutex.RUnlock()

    fmt.Printf("异步发布消息到主题: %s, 处理器数量: %d\n", topic, len(handlers))

    for _, handler := range handlers {
        go func(h MessageHandler) {
            if err := h.Handle(context.Background(), message); err != nil {
                fmt.Printf("处理器 %s 异步处理消息失败: %v\n", h.GetID(), err)
            }
        }(handler)
    }

    return nil
}

// 6. 示例插件实现
type LoggerPlugin struct {
    info   *PluginInfo
    status PluginStatus
    config map[string]interface{}
    msgBus MessageBus
}

func NewLoggerPlugin(msgBus MessageBus) *LoggerPlugin {
    return &LoggerPlugin{
        info: &PluginInfo{
            Name:        "logger",
            Version:     "1.0.0",
            Description: "日志记录插件",
            Author:      "System",
            Capabilities: []string{"logging", "file-output"},
            Config: map[string]interface{}{
                "level":  "info",
                "output": "console",
            },
        },
        status: StatusUnknown,
        msgBus: msgBus,
    }
}

func (lp *LoggerPlugin) GetInfo() *PluginInfo {
    return lp.info
}

func (lp *LoggerPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    lp.config = config
    fmt.Printf("初始化日志插件，配置: %v\n", config)

    // 订阅日志消息
    handler := &LogMessageHandler{id: "logger-handler"}
    lp.msgBus.Subscribe("log", handler)

    lp.status = StatusInitialized
    return nil
}

func (lp *LoggerPlugin) Start(ctx context.Context) error {
    fmt.Println("启动日志插件")
    lp.status = StatusStarted
    return nil
}

func (lp *LoggerPlugin) Stop(ctx context.Context) error {
    fmt.Println("停止日志插件")
    lp.status = StatusStopped
    return nil
}

func (lp *LoggerPlugin) Destroy(ctx context.Context) error {
    fmt.Println("销毁日志插件")
    lp.msgBus.Unsubscribe("log", "logger-handler")
    return nil
}

func (lp *LoggerPlugin) GetStatus() PluginStatus {
    return lp.status
}

type LogMessageHandler struct {
    id string
}

func (lmh *LogMessageHandler) GetID() string {
    return lmh.id
}

func (lmh *LogMessageHandler) Handle(ctx context.Context, message *Message) error {
    level := message.Data["level"].(string)
    msg := message.Data["message"].(string)
    fmt.Printf("[%s] %s: %s\n", message.Timestamp.Format("15:04:05"), level, msg)
    return nil
}

// 数据库插件
type DatabasePlugin struct {
    info   *PluginInfo
    status PluginStatus
    config map[string]interface{}
    msgBus MessageBus
}

func NewDatabasePlugin(msgBus MessageBus) *DatabasePlugin {
    return &DatabasePlugin{
        info: &PluginInfo{
            Name:        "database",
            Version:     "1.0.0",
            Description: "数据库连接插件",
            Author:      "System",
            Dependencies: []string{"logger"},
            Capabilities: []string{"mysql", "postgresql"},
            Config: map[string]interface{}{
                "driver": "mysql",
                "host":   "localhost",
                "port":   3306,
            },
        },
        status: StatusUnknown,
        msgBus: msgBus,
    }
}

func (dp *DatabasePlugin) GetInfo() *PluginInfo {
    return dp.info
}

func (dp *DatabasePlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    dp.config = config

    // 发送日志消息
    logMsg := &Message{
        ID:     "db-init",
        Topic:  "log",
        Source: "database",
        Type:   "info",
        Data: map[string]interface{}{
            "level":   "INFO",
            "message": "初始化数据库插件",
        },
        Timestamp: time.Now(),
    }
    dp.msgBus.Publish("log", logMsg)

    dp.status = StatusInitialized
    return nil
}

func (dp *DatabasePlugin) Start(ctx context.Context) error {
    logMsg := &Message{
        ID:     "db-start",
        Topic:  "log",
        Source: "database",
        Type:   "info",
        Data: map[string]interface{}{
            "level":   "INFO",
            "message": "启动数据库插件",
        },
        Timestamp: time.Now(),
    }
    dp.msgBus.Publish("log", logMsg)

    dp.status = StatusStarted
    return nil
}

func (dp *DatabasePlugin) Stop(ctx context.Context) error {
    logMsg := &Message{
        ID:     "db-stop",
        Topic:  "log",
        Source: "database",
        Type:   "info",
        Data: map[string]interface{}{
            "level":   "INFO",
            "message": "停止数据库插件",
        },
        Timestamp: time.Now(),
    }
    dp.msgBus.Publish("log", logMsg)

    dp.status = StatusStopped
    return nil
}

func (dp *DatabasePlugin) Destroy(ctx context.Context) error {
    fmt.Println("销毁数据库插件")
    return nil
}

func (dp *DatabasePlugin) GetStatus() PluginStatus {
    return dp.status
}

// 测试函数
func TestExercise5() {
    fmt.Println("=== 练习题5：插件系统架构 ===")

    // 创建插件系统组件
    registry := NewPluginRegistry()
    msgBus := NewSimpleMsgBus()

    // 创建插件
    loggerPlugin := NewLoggerPlugin(msgBus)
    dbPlugin := NewDatabasePlugin(msgBus)

    // 注册插件
    fmt.Println("1. 注册插件:")
    registry.RegisterPlugin(loggerPlugin)
    registry.RegisterPlugin(dbPlugin)

    // 列出插件
    fmt.Println("\n2. 插件列表:")
    plugins := registry.ListPlugins()
    for _, info := range plugins {
        fmt.Printf("- %s v%s: %s\n", info.Name, info.Version, info.Description)
        if len(info.Dependencies) > 0 {
            fmt.Printf("  依赖: %v\n", info.Dependencies)
        }
        if len(info.Capabilities) > 0 {
            fmt.Printf("  能力: %v\n", info.Capabilities)
        }
    }

    // 启动插件
    fmt.Println("\n3. 启动插件:")
    registry.StartPlugin("logger")
    registry.StartPlugin("database")

    // 检查状态
    fmt.Println("\n4. 插件状态:")
    for _, info := range plugins {
        status, _ := registry.GetPluginStatus(info.Name)
        fmt.Printf("- %s: %s\n", info.Name, status)
    }

    // 测试插件通信
    fmt.Println("\n5. 测试插件通信:")
    testMsg := &Message{
        ID:     "test-msg",
        Topic:  "log",
        Source: "system",
        Type:   "info",
        Data: map[string]interface{}{
            "level":   "INFO",
            "message": "这是一条测试消息",
        },
        Timestamp: time.Now(),
    }
    msgBus.Publish("log", testMsg)

    // 停止插件
    fmt.Println("\n6. 停止插件:")
    registry.StopPlugin("database")
    registry.StopPlugin("logger")

    // 卸载插件
    fmt.Println("\n7. 卸载插件:")
    registry.UnregisterPlugin("database")
    registry.UnregisterPlugin("logger")
}

/*
解析说明：
1. 插件架构：定义了完整的插件生命周期管理
2. 依赖管理：支持插件间依赖关系检查
3. 消息通信：实现了插件间的消息总线通信
4. 状态管理：跟踪插件的运行状态
5. 热插拔：支持动态加载和卸载插件

扩展思考：
- 如何实现插件的版本兼容性检查？
- 如何支持插件的配置热更新？
- 如何实现插件的资源隔离？
- 如何添加插件的性能监控？
- 如何支持分布式插件系统？
*/
```

---

## 📚 章节总结

恭喜你完成了Go语言接口设计模式的学习！🎉 让我们来总结一下这一章的核心内容。

### 🎯 核心知识点回顾

#### 1. Go接口设计哲学 🏛️

```go
// Go接口设计的核心原则
/*
1. "接受接口，返回结构体" (Accept interfaces, return structs)
   - 函数参数使用接口类型，增加灵活性
   - 函数返回值使用具体类型，提供明确的API

2. 接口隔离原则 (Interface Segregation Principle)
   - 接口应该小而专一
   - 客户端不应该依赖它不需要的接口

3. 隐式实现 (Implicit Implementation)
   - 无需显式声明实现接口
   - 只要实现了接口的所有方法就自动满足接口

4. 组合优于继承 (Composition over Inheritance)
   - 通过接口组合实现复杂功能
   - 避免深层次的继承关系
*/

// 示例：优秀的接口设计
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type ReadWriter interface {
    Reader
    Writer
}
```

#### 2. 经典设计模式在Go中的实现 🎨

| 设计模式 | Go实现特点 | 适用场景 |
|---------|-----------|---------|
| **工厂模式** | 使用函数返回接口类型 | 对象创建复杂时 |
| **策略模式** | 接口+结构体实现 | 算法可替换时 |
| **装饰器模式** | 接口嵌入+组合 | 功能动态扩展 |
| **适配器模式** | 接口转换 | 系统集成时 |
| **观察者模式** | 接口+通道 | 事件驱动系统 |

#### 3. 依赖注入和控制反转 💉

```go
// 依赖注入的三种方式
type UserService struct {
    // 1. 构造函数注入（推荐）
    repo   UserRepository
    logger Logger
}

func NewUserService(repo UserRepository, logger Logger) *UserService {
    return &UserService{repo: repo, logger: logger}
}

// 2. 接口注入
type ServiceInjector interface {
    InjectUserRepository(repo UserRepository)
    InjectLogger(logger Logger)
}

// 3. 属性注入（使用标签）
type AutowiredService struct {
    Repo   UserRepository `inject:"true"`
    Logger Logger         `inject:"true"`
}
```

#### 4. 接口组合和嵌入 🔗

```go
// 接口组合的强大之处
type ReadWriteCloser interface {
    io.Reader    // 嵌入标准接口
    io.Writer    // 嵌入标准接口
    io.Closer    // 嵌入标准接口
}

// 实现时只需实现各个小接口
type File struct {
    // ...
}

func (f *File) Read(p []byte) (n int, err error) { /* ... */ }
func (f *File) Write(p []byte) (n int, err error) { /* ... */ }
func (f *File) Close() error { /* ... */ }
// File 自动实现了 ReadWriteCloser 接口
```

### 🚀 Go接口设计最佳实践清单

#### ✅ 应该做的事情

1. **设计小接口**
   ```go
   // ✅ 好的设计
   type Reader interface {
       Read([]byte) (int, error)
   }

   type Writer interface {
       Write([]byte) (int, error)
   }
   ```

2. **使用标准库接口**
   ```go
   // ✅ 优先使用标准接口
   func ProcessData(r io.Reader, w io.Writer) error {
       // 与标准库兼容
   }
   ```

3. **接受接口，返回结构体**
   ```go
   // ✅ 参数使用接口
   func NewService(repo Repository) *Service {
       return &Service{repo: repo}  // 返回具体类型
   }
   ```

4. **使用接口组合**
   ```go
   // ✅ 通过组合创建复杂接口
   type ReadWriteSeeker interface {
       io.Reader
       io.Writer
       io.Seeker
   }
   ```

#### ❌ 不应该做的事情

1. **避免大接口**
   ```go
   // ❌ 违反接口隔离原则
   type BadService interface {
       CreateUser(user User) error
       UpdateUser(user User) error
       DeleteUser(id int) error
       SendEmail(to, subject, body string) error
       ProcessPayment(amount float64) error
       GenerateReport() ([]byte, error)
   }
   ```

2. **避免滥用空接口**
   ```go
   // ❌ 类型不安全
   func BadFunction(data interface{}) interface{} {
       // 需要大量类型断言
   }

   // ✅ 使用具体类型或泛型
   func GoodFunction[T any](data T) T {
       return data
   }
   ```

3. **避免接口nil陷阱**
   ```go
   // ❌ 返回nil指针给接口
   func BadGetError() error {
       var err *MyError = nil
       return err  // 接口不为nil！
   }

   // ✅ 返回明确的nil
   func GoodGetError() error {
       var err *MyError = nil
       if err != nil {
           return err
       }
       return nil  // 真正的nil接口
   }
   ```

### 🔗 与前面章节的知识连接

#### 与结构体和接口章节的联系
- **结构体实现接口**：通过方法集实现接口契约
- **接口嵌入**：扩展了结构体嵌入的概念到接口层面
- **多态实现**：同一接口的不同实现提供不同行为

#### 与错误处理章节的联系
- **错误接口**：`error` 是Go中最重要的接口
- **自定义错误类型**：实现error接口创建丰富的错误信息
- **错误包装**：通过接口实现错误链

#### 与并发编程章节的联系
- **并发安全接口**：接口方法的并发安全性设计
- **通道与接口**：结合使用实现异步通信
- **Context接口**：贯穿整个请求生命周期的上下文管理

### 🎓 下一步学习建议

#### 立即实践 (本周内)
1. **重构现有代码**：将大接口拆分为小接口
2. **实现设计模式**：选择1-2个设计模式在项目中应用
3. **编写接口测试**：为关键接口编写Mock测试

#### 深入学习 (本月内)
1. **研究标准库**：深入学习Go标准库的接口设计
2. **性能优化**：分析接口调用的性能影响
3. **架构设计**：在系统架构中合理使用接口

#### 高级进阶 (长期目标)
1. **插件系统**：设计可扩展的插件架构
2. **微服务接口**：设计服务间的接口契约
3. **开源贡献**：参与Go开源项目，学习最佳实践

### 🌟 Go接口设计的独特优势

#### 相比Java的优势
```java
// Java需要显式实现
public class MyWriter implements Writer {
    @Override
    public void write(String data) { /* ... */ }
}
```

```go
// Go隐式实现，更灵活
type MyWriter struct{}

func (w MyWriter) Write(p []byte) (int, error) {
    // 自动实现了io.Writer接口
}
```

#### 相比Python的优势
```python
# Python运行时检查
def process_data(writer):
    if hasattr(writer, 'write'):
        writer.write(data)  # 运行时可能出错
```

```go
// Go编译时检查
func ProcessData(w io.Writer) error {
    _, err := w.Write(data)  // 编译时保证类型安全
    return err
}
```

### 🎯 总结感言

Go语言的接口设计体现了"简单而强大"的哲学。通过小接口、隐式实现和组合设计，Go提供了一种既灵活又类型安全的抽象机制。掌握接口设计模式，不仅能写出更优雅的代码，更能设计出可扩展、可维护的系统架构。

记住Go接口设计的黄金法则：
- **Keep interfaces small** - 保持接口小巧
- **Accept interfaces, return structs** - 接受接口，返回结构体
- **Composition over inheritance** - 组合优于继承
- **Design for testability** - 为可测试性而设计

继续你的Go语言学习之旅，下一章我们将探索更高级的主题。加油！🚀

---

**本章完成标志：**
- ✅ 掌握Go接口设计哲学
- ✅ 实现经典设计模式
- ✅ 理解依赖注入原理
- ✅ 熟练使用接口组合
- ✅ 避免常见接口陷阱
- ✅ 完成5道综合练习

**学习成果：**
- 📖 **理论知识**：接口设计原理和最佳实践
- 🛠️ **实践技能**：设计模式实现和接口测试
- 🏗️ **架构能力**：系统接口设计和插件架构
- 🎯 **面试准备**：接口相关面试题和标准答案

恭喜你已经成为Go接口设计的专家！🎉
```
```
```
```
```
```
```
```
```
```
```
```
```
```
```
```
