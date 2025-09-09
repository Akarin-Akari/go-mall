# 实战篇第二章：GORM数据库操作与实践 🗄️

> **"数据是应用的灵魂，GORM是Go语言操作数据库的利器"** 💎

---

## 📖 章节导读

欢迎来到GORM数据库操作的实战世界！🌟 在上一章中，我们学习了Gin框架的Web开发技能。现在是时候深入数据持久化层，掌握Go语言中最流行的ORM框架——GORM。

GORM（Go Object Relational Mapping）是Go语言的一个功能强大的ORM库，它提供了简洁的API来操作数据库，支持多种数据库引擎，并且具有出色的性能表现。

### 🎯 学习目标

通过本章学习，你将掌握：

- **🏗️ GORM基础**：理解ORM概念和GORM的设计理念
- **📊 模型定义**：掌握数据模型的定义和关联关系
- **🔧 CRUD操作**：熟练进行增删改查操作
- **🔗 关联查询**：处理复杂的表关联和查询
- **⚡ 事务处理**：确保数据一致性和完整性
- **🚀 性能优化**：查询优化和连接池配置
- **🧪 测试技巧**：数据库操作的单元测试
- **🏢 企业实践**：结合mall-go项目的数据库设计

### 📋 章节大纲

```
02-gorm-database-operations.md
├── 🌟  GORM框架概述
├── 🚀  快速开始
├── 📊  模型定义与迁移
│   ├── 基础模型
│   ├── 字段标签
│   ├── 关联关系
│   └── 数据库迁移
├── 🔧  CRUD操作详解
│   ├── 创建记录
│   ├── 查询数据
│   ├── 更新记录
│   └── 删除数据
├── 🔗  高级查询技巧
│   ├── 条件查询
│   ├── 关联查询
│   ├── 聚合查询
│   └── 原生SQL
├── ⚡  事务与并发
├── 🚀  性能优化
├── 🏢  实战案例分析
├── 🎯  面试常考点
├── ⚠️   踩坑提醒
├── 📝  练习题
└── 📚  章节总结
```

---

## 🌟 GORM框架概述

### 什么是GORM？

GORM是Go语言的一个功能丰富的ORM（Object Relational Mapping）库，它提供了开发者友好的API来操作数据库。

```go
// 来自 mall-go/pkg/database/gorm.go
package database

import (
    "fmt"
    "time"
    
    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

/*
GORM的核心特性：

1. 全功能ORM：
   - 关联关系（Has One, Has Many, Many To Many, Polymorphism）
   - 钩子函数（Before/After Create/Save/Update/Delete/Find）
   - 预加载（Eager Loading）
   - 事务支持

2. 开发者友好：
   - 自动迁移
   - SQL构建器
   - 自动创建/更新时间戳
   - 软删除

3. 高性能：
   - 复合主键
   - 索引支持
   - 数据库连接池
   - 读写分离

4. 多数据库支持：
   - MySQL, PostgreSQL, SQLite, SQL Server
   - 插件化驱动架构
*/

// 数据库配置
type Config struct {
    Driver   string `yaml:"driver"`
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    Database string `yaml:"database"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    Charset  string `yaml:"charset"`
    
    // 连接池配置
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    MaxOpenConns    int           `yaml:"max_open_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
    
    // 日志配置
    LogLevel logger.LogLevel `yaml:"log_level"`
}

// 创建数据库连接
func NewConnection(config Config) (*gorm.DB, error) {
    var dialector gorm.Dialector
    
    switch config.Driver {
    case "mysql":
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
            config.Username, config.Password, config.Host, config.Port,
            config.Database, config.Charset)
        dialector = mysql.Open(dsn)
        
    case "postgres":
        dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
            config.Host, config.Username, config.Password, config.Database, config.Port)
        dialector = postgres.Open(dsn)
        
    case "sqlite":
        dialector = sqlite.Open(config.Database)
        
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
    }
    
    // GORM配置
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(config.LogLevel),
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "t_",           // 表名前缀
            SingularTable: true,           // 使用单数表名
        },
    }
    
    db, err := gorm.Open(dialector, gormConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // 获取底层sql.DB
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }
    
    // 配置连接池
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    return db, nil
}
```

### GORM vs 其他ORM框架对比

#### 与Java Hibernate对比

```java
// Java Hibernate
@Entity
@Table(name = "users")
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(name = "username", nullable = false, unique = true)
    private String username;
    
    @Column(name = "email", nullable = false)
    private String email;
    
    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL)
    private List<Order> orders;
    
    // getters and setters...
}

// 查询操作
@Repository
public class UserRepository {
    @Autowired
    private EntityManager entityManager;
    
    public List<User> findByUsername(String username) {
        return entityManager.createQuery(
            "SELECT u FROM User u WHERE u.username = :username", User.class)
            .setParameter("username", username)
            .getResultList();
    }
}
```

```go
// Go GORM等价实现
package model

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Username  string    `gorm:"uniqueIndex;not null;size:50" json:"username"`
    Email     string    `gorm:"not null;size:100" json:"email"`
    Orders    []Order   `gorm:"foreignKey:UserID" json:"orders,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 查询操作
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) ([]User, error) {
    var users []User
    err := r.db.Where("username = ?", username).Find(&users).Error
    return users, err
}

// 或者更简洁的方式
func (r *UserRepository) FindByUsernameSimple(username string) ([]User, error) {
    var users []User
    err := r.db.Find(&users, "username = ?", username).Error
    return users, err
}
```

#### 与Python SQLAlchemy对比

```python
# Python SQLAlchemy
from sqlalchemy import Column, Integer, String, DateTime, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship, sessionmaker

Base = declarative_base()

class User(Base):
    __tablename__ = 'users'
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    username = Column(String(50), nullable=False, unique=True)
    email = Column(String(100), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)
    
    orders = relationship("Order", back_populates="user")

# 查询操作
class UserRepository:
    def __init__(self, session):
        self.session = session
    
    def find_by_username(self, username):
        return self.session.query(User).filter(
            User.username == username
        ).all()
    
    def create_user(self, user_data):
        user = User(**user_data)
        self.session.add(user)
        self.session.commit()
        return user
```

```go
// Go GORM等价实现
type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    Username  string    `gorm:"uniqueIndex;not null;size:50"`
    Email     string    `gorm:"not null;size:100"`
    Orders    []Order   `gorm:"foreignKey:UserID"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) FindByUsername(username string) ([]User, error) {
    var users []User
    err := r.db.Where("username = ?", username).Find(&users).Error
    return users, err
}

func (r *UserRepository) CreateUser(userData User) (*User, error) {
    err := r.db.Create(&userData).Error
    if err != nil {
        return nil, err
    }
    return &userData, nil
}
```

### 框架对比总结

| 特性 | GORM (Go) | Hibernate (Java) | SQLAlchemy (Python) |
|------|-----------|------------------|-------------------|
| **性能** | 极高 | 中等 | 中等 |
| **内存占用** | 很低 | 较高 | 中等 |
| **学习曲线** | 平缓 | 陡峭 | 中等 |
| **类型安全** | 编译时检查 | 编译时检查 | 运行时检查 |
| **配置复杂度** | 简单 | 复杂 | 中等 |
| **关联查询** | 简洁 | 功能强大 | 灵活 |
| **迁移支持** | 自动迁移 | 需要工具 | 需要Alembic |

---

## 🚀 快速开始

让我们通过一个完整的示例来快速上手GORM。

### 项目初始化

```bash
# 创建项目目录
mkdir gorm-demo
cd gorm-demo

# 初始化Go模块
go mod init gorm-demo

# 安装GORM和数据库驱动
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get gorm.io/driver/postgres
go get gorm.io/driver/sqlite
```

### 第一个GORM应用

```go
// 来自 mall-go/cmd/migrate/main.go
package main

import (
    "fmt"
    "log"
    "time"
    
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// 用户模型
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:100;not null"`
    Email     string    `gorm:"uniqueIndex;size:100;not null"`
    Age       int       `gorm:"check:age > 0"`
    Birthday  time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 产品模型
type Product struct {
    ID          uint    `gorm:"primaryKey"`
    Code        string  `gorm:"uniqueIndex;size:50"`
    Name        string  `gorm:"size:100;not null"`
    Price       float64 `gorm:"not null"`
    Description string  `gorm:"type:text"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 订单模型
type Order struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null;index"`
    User      User      `gorm:"foreignKey:UserID"`
    ProductID uint      `gorm:"not null;index"`
    Product   Product   `gorm:"foreignKey:ProductID"`
    Quantity  int       `gorm:"not null;default:1"`
    Amount    float64   `gorm:"not null"`
    Status    string    `gorm:"size:20;default:'pending'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func main() {
    // 连接数据库
    db, err := gorm.Open(sqlite.Open("demo.db"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // 自动迁移模式
    err = db.AutoMigrate(&User{}, &Product{}, &Order{})
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    
    // 创建示例数据
    createSampleData(db)
    
    // 查询示例
    queryExamples(db)
    
    fmt.Println("GORM demo completed successfully!")
}

func createSampleData(db *gorm.DB) {
    fmt.Println("Creating sample data...")
    
    // 创建用户
    users := []User{
        {
            Name:     "张三",
            Email:    "zhangsan@example.com",
            Age:      25,
            Birthday: time.Date(1998, 5, 15, 0, 0, 0, 0, time.UTC),
        },
        {
            Name:     "李四",
            Email:    "lisi@example.com",
            Age:      30,
            Birthday: time.Date(1993, 8, 20, 0, 0, 0, 0, time.UTC),
        },
    }
    
    result := db.Create(&users)
    if result.Error != nil {
        log.Printf("Failed to create users: %v", result.Error)
        return
    }
    fmt.Printf("Created %d users\n", result.RowsAffected)
    
    // 创建产品
    products := []Product{
        {
            Code:        "BOOK001",
            Name:        "Go语言实战",
            Price:       89.00,
            Description: "深入学习Go语言的实战教程",
        },
        {
            Code:        "BOOK002",
            Name:        "数据结构与算法",
            Price:       79.00,
            Description: "计算机科学基础教程",
        },
    }
    
    result = db.Create(&products)
    if result.Error != nil {
        log.Printf("Failed to create products: %v", result.Error)
        return
    }
    fmt.Printf("Created %d products\n", result.RowsAffected)
    
    // 创建订单
    orders := []Order{
        {
            UserID:    users[0].ID,
            ProductID: products[0].ID,
            Quantity:  2,
            Amount:    178.00,
            Status:    "completed",
        },
        {
            UserID:    users[1].ID,
            ProductID: products[1].ID,
            Quantity:  1,
            Amount:    79.00,
            Status:    "pending",
        },
    }
    
    result = db.Create(&orders)
    if result.Error != nil {
        log.Printf("Failed to create orders: %v", result.Error)
        return
    }
    fmt.Printf("Created %d orders\n", result.RowsAffected)
}

func queryExamples(db *gorm.DB) {
    fmt.Println("\nQuery examples:")
    
    // 1. 查询所有用户
    var users []User
    db.Find(&users)
    fmt.Printf("Found %d users\n", len(users))
    
    // 2. 根据条件查询
    var user User
    db.Where("email = ?", "zhangsan@example.com").First(&user)
    fmt.Printf("Found user: %s\n", user.Name)
    
    // 3. 关联查询
    var orders []Order
    db.Preload("User").Preload("Product").Find(&orders)
    for _, order := range orders {
        fmt.Printf("Order: %s bought %s, quantity: %d\n",
            order.User.Name, order.Product.Name, order.Quantity)
    }
    
    // 4. 聚合查询
    var count int64
    db.Model(&Order{}).Where("status = ?", "completed").Count(&count)
    fmt.Printf("Completed orders count: %d\n", count)
    
    // 5. 原生SQL查询
    var result struct {
        UserName     string
        TotalAmount  float64
        OrderCount   int
    }
    
    db.Raw(`
        SELECT u.name as user_name, 
               SUM(o.amount) as total_amount,
               COUNT(o.id) as order_count
        FROM users u
        JOIN orders o ON u.id = o.user_id
        WHERE u.id = ?
        GROUP BY u.id, u.name
    `, user.ID).Scan(&result)
    
    fmt.Printf("User %s: Total amount: %.2f, Order count: %d\n",
        result.UserName, result.TotalAmount, result.OrderCount)
}
```

---

## 📊 模型定义与迁移

GORM的模型定义是数据库操作的基础，让我们深入了解如何定义高效的数据模型。

### 基础模型

```go
// 来自 mall-go/internal/model/base.go
package model

import (
    "time"
    "gorm.io/gorm"
)

// 基础模型 - 包含通用字段
type BaseModel struct {
    ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 用户模型
type User struct {
    BaseModel
    Username    string    `gorm:"uniqueIndex;size:50;not null" json:"username" binding:"required,min=3,max=20"`
    Email       string    `gorm:"uniqueIndex;size:100;not null" json:"email" binding:"required,email"`
    Password    string    `gorm:"size:255;not null" json:"-"`
    Nickname    string    `gorm:"size:50" json:"nickname"`
    Avatar      string    `gorm:"size:255" json:"avatar"`
    Phone       string    `gorm:"size:20" json:"phone"`
    Gender      int8      `gorm:"default:0;comment:0-未知,1-男,2-女" json:"gender"`
    Birthday    *time.Time `json:"birthday"`
    Status      int8      `gorm:"default:1;comment:0-禁用,1-启用" json:"status"`
    LastLoginAt *time.Time `json:"last_login_at"`

    // 关联关系
    Profile UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Orders  []Order     `gorm:"foreignKey:UserID" json:"orders,omitempty"`
    Reviews []Review    `gorm:"foreignKey:UserID" json:"reviews,omitempty"`
}

// 用户资料模型
type UserProfile struct {
    BaseModel
    UserID      uint   `gorm:"uniqueIndex;not null" json:"user_id"`
    RealName    string `gorm:"size:50" json:"real_name"`
    IDCard      string `gorm:"size:18" json:"id_card"`
    Address     string `gorm:"size:255" json:"address"`
    Company     string `gorm:"size:100" json:"company"`
    Position    string `gorm:"size:50" json:"position"`
    Bio         string `gorm:"type:text" json:"bio"`
    Website     string `gorm:"size:255" json:"website"`

    // 反向关联
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// 分类模型
type Category struct {
    BaseModel
    Name        string `gorm:"size:50;not null" json:"name" binding:"required"`
    Slug        string `gorm:"uniqueIndex;size:50;not null" json:"slug"`
    Description string `gorm:"type:text" json:"description"`
    Image       string `gorm:"size:255" json:"image"`
    Sort        int    `gorm:"default:0" json:"sort"`
    Status      int8   `gorm:"default:1" json:"status"`

    // 自关联 - 父子分类
    ParentID uint        `gorm:"default:0;index" json:"parent_id"`
    Parent   *Category   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children []Category  `gorm:"foreignKey:ParentID" json:"children,omitempty"`

    // 关联产品
    Products []Product `gorm:"many2many:product_categories;" json:"products,omitempty"`
}

// 产品模型
type Product struct {
    BaseModel
    Name        string  `gorm:"size:100;not null" json:"name" binding:"required"`
    Slug        string  `gorm:"uniqueIndex;size:100;not null" json:"slug"`
    Description string  `gorm:"type:text" json:"description"`
    Content     string  `gorm:"type:longtext" json:"content"`
    SKU         string  `gorm:"uniqueIndex;size:50;not null" json:"sku"`
    Price       float64 `gorm:"type:decimal(10,2);not null" json:"price" binding:"required,gt=0"`
    OriginalPrice float64 `gorm:"type:decimal(10,2)" json:"original_price"`
    Stock       int     `gorm:"not null;default:0" json:"stock"`
    Sales       int     `gorm:"default:0" json:"sales"`
    Views       int     `gorm:"default:0" json:"views"`
    Status      int8    `gorm:"default:1" json:"status"`

    // JSON字段
    Images      JSON `gorm:"type:json" json:"images"`
    Attributes  JSON `gorm:"type:json" json:"attributes"`
    Specs       JSON `gorm:"type:json" json:"specs"`

    // 关联关系
    Categories []Category `gorm:"many2many:product_categories;" json:"categories,omitempty"`
    Reviews    []Review   `gorm:"foreignKey:ProductID" json:"reviews,omitempty"`
    OrderItems []OrderItem `gorm:"foreignKey:ProductID" json:"order_items,omitempty"`
}

// 订单模型
type Order struct {
    BaseModel
    OrderNo     string  `gorm:"uniqueIndex;size:32;not null" json:"order_no"`
    UserID      uint    `gorm:"not null;index" json:"user_id"`
    TotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"total_amount"`
    PayAmount   float64 `gorm:"type:decimal(10,2);not null" json:"pay_amount"`
    Status      string  `gorm:"size:20;default:'pending'" json:"status"`
    PayStatus   string  `gorm:"size:20;default:'unpaid'" json:"pay_status"`
    PayMethod   string  `gorm:"size:20" json:"pay_method"`
    PayTime     *time.Time `json:"pay_time"`
    ShipTime    *time.Time `json:"ship_time"`
    FinishTime  *time.Time `json:"finish_time"`

    // 收货信息
    ReceiverName    string `gorm:"size:50;not null" json:"receiver_name"`
    ReceiverPhone   string `gorm:"size:20;not null" json:"receiver_phone"`
    ReceiverAddress string `gorm:"size:255;not null" json:"receiver_address"`

    // 备注信息
    Remark      string `gorm:"type:text" json:"remark"`
    AdminRemark string `gorm:"type:text" json:"admin_remark"`

    // 关联关系
    User       User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
    OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
}

// 订单项模型
type OrderItem struct {
    BaseModel
    OrderID   uint    `gorm:"not null;index" json:"order_id"`
    ProductID uint    `gorm:"not null;index" json:"product_id"`
    Quantity  int     `gorm:"not null" json:"quantity"`
    Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
    Amount    float64 `gorm:"type:decimal(10,2);not null" json:"amount"`

    // 快照信息（防止产品信息变更影响历史订单）
    ProductName  string `gorm:"size:100;not null" json:"product_name"`
    ProductImage string `gorm:"size:255" json:"product_image"`
    ProductSKU   string `gorm:"size:50" json:"product_sku"`

    // 关联关系
    Order   Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
    Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// 评论模型
type Review struct {
    BaseModel
    UserID    uint   `gorm:"not null;index" json:"user_id"`
    ProductID uint   `gorm:"not null;index" json:"product_id"`
    OrderID   uint   `gorm:"not null;index" json:"order_id"`
    Rating    int8   `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
    Content   string `gorm:"type:text" json:"content"`
    Images    JSON   `gorm:"type:json" json:"images"`
    Status    int8   `gorm:"default:1" json:"status"`

    // 管理员回复
    AdminReply   string     `gorm:"type:text" json:"admin_reply"`
    AdminReplyAt *time.Time `json:"admin_reply_at"`

    // 关联关系
    User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
    Order   Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// 自定义JSON类型
type JSON map[string]interface{}

// 实现Scanner接口
func (j *JSON) Scan(value interface{}) error {
    if value == nil {
        *j = make(JSON)
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("cannot scan %T into JSON", value)
    }

    return json.Unmarshal(bytes, j)
}

// 实现Valuer接口
func (j JSON) Value() (driver.Value, error) {
    if j == nil {
        return nil, nil
    }
    return json.Marshal(j)
}
```

### 字段标签详解

```go
// 来自 mall-go/internal/model/tags.go
package model

import (
    "time"
    "gorm.io/gorm"
)

// 字段标签示例
type TagExample struct {
    // 主键标签
    ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

    // 字段约束
    Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email    string `gorm:"index;size:100;not null" json:"email"`
    Age      int    `gorm:"check:age > 0" json:"age"`

    // 默认值
    Status   int8   `gorm:"default:1" json:"status"`
    IsActive bool   `gorm:"default:true" json:"is_active"`

    // 字段类型
    Price       float64 `gorm:"type:decimal(10,2)" json:"price"`
    Description string  `gorm:"type:text" json:"description"`
    Content     string  `gorm:"type:longtext" json:"content"`

    // 时间字段
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    // 忽略字段
    Password string `gorm:"-" json:"-"`

    // 只读字段
    CreatedBy uint `gorm:"<-:create" json:"created_by"`

    // 只写字段
    UpdatedBy uint `gorm:"<-:update" json:"updated_by"`

    // 序列化标签
    Metadata JSON `gorm:"type:json;serializer:json" json:"metadata"`
}

/*
常用GORM标签说明：

1. 主键和索引：
   - primaryKey: 主键
   - autoIncrement: 自增
   - uniqueIndex: 唯一索引
   - index: 普通索引

2. 字段约束：
   - not null: 非空
   - size:100: 字段长度
   - check:age > 0: 检查约束
   - default:1: 默认值

3. 字段类型：
   - type:varchar(100): 指定数据库类型
   - type:text: 文本类型
   - type:decimal(10,2): 精确小数

4. 关联标签：
   - foreignKey: 外键字段
   - references: 引用字段
   - constraint: 约束选项

5. 序列化：
   - serializer:json: JSON序列化
   - serializer:gob: GOB序列化

6. 权限控制：
   - <-:create: 只在创建时写入
   - <-:update: 只在更新时写入
   - <-:false: 禁止写入
   - ->:false: 禁止读取
   - -: 忽略字段
*/
```

### 关联关系

```go
// 来自 mall-go/internal/model/associations.go
package model

import (
    "time"
    "gorm.io/gorm"
)

// 1. Has One 关系 (一对一)
type User struct {
    ID      uint        `gorm:"primaryKey"`
    Name    string      `gorm:"size:100"`
    Profile UserProfile `gorm:"foreignKey:UserID"` // 用户有一个资料
}

type UserProfile struct {
    ID     uint   `gorm:"primaryKey"`
    UserID uint   `gorm:"uniqueIndex"` // 外键
    Bio    string `gorm:"type:text"`
    Avatar string `gorm:"size:255"`

    // 反向关联
    User User `gorm:"foreignKey:UserID"`
}

// 2. Has Many 关系 (一对多)
type Category struct {
    ID       uint      `gorm:"primaryKey"`
    Name     string    `gorm:"size:50"`
    Products []Product `gorm:"foreignKey:CategoryID"` // 分类有多个产品
}

type Product struct {
    ID         uint   `gorm:"primaryKey"`
    Name       string `gorm:"size:100"`
    CategoryID uint   `gorm:"index"` // 外键

    // 反向关联
    Category Category `gorm:"foreignKey:CategoryID"`
}

// 3. Many To Many 关系 (多对多)
type Product struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:100"`

    // 多对多关系
    Tags []Tag `gorm:"many2many:product_tags;"` // 产品有多个标签
}

type Tag struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:50"`

    // 反向关联
    Products []Product `gorm:"many2many:product_tags;"`
}

// 4. 自关联 (Self-Referencing)
type Category struct {
    ID       uint       `gorm:"primaryKey"`
    Name     string     `gorm:"size:50"`
    ParentID uint       `gorm:"default:0;index"`

    // 自关联
    Parent   *Category  `gorm:"foreignKey:ParentID"`
    Children []Category `gorm:"foreignKey:ParentID"`
}

// 5. 多态关联 (Polymorphic)
type Image struct {
    ID           uint   `gorm:"primaryKey"`
    URL          string `gorm:"size:255"`
    OwnerID      uint   `gorm:"index"`
    OwnerType    string `gorm:"size:50;index"`
}

type Product struct {
    ID     uint    `gorm:"primaryKey"`
    Name   string  `gorm:"size:100"`
    Images []Image `gorm:"polymorphic:Owner;"` // 多态关联
}

type User struct {
    ID     uint    `gorm:"primaryKey"`
    Name   string  `gorm:"size:100"`
    Images []Image `gorm:"polymorphic:Owner;"` // 多态关联
}

// 6. 复杂关联示例 - 电商订单系统
type Order struct {
    ID          uint      `gorm:"primaryKey"`
    OrderNo     string    `gorm:"uniqueIndex;size:32"`
    UserID      uint      `gorm:"not null;index"`
    TotalAmount float64   `gorm:"type:decimal(10,2)"`
    Status      string    `gorm:"size:20"`
    CreatedAt   time.Time

    // 关联关系
    User       User        `gorm:"foreignKey:UserID"`
    OrderItems []OrderItem `gorm:"foreignKey:OrderID"`

    // 通过中间表关联产品
    Products []Product `gorm:"many2many:order_products;"`
}

type OrderItem struct {
    ID        uint    `gorm:"primaryKey"`
    OrderID   uint    `gorm:"not null;index"`
    ProductID uint    `gorm:"not null;index"`
    Quantity  int     `gorm:"not null"`
    Price     float64 `gorm:"type:decimal(10,2)"`

    // 关联关系
    Order   Order   `gorm:"foreignKey:OrderID"`
    Product Product `gorm:"foreignKey:ProductID"`
}

// 关联查询示例
func AssociationExamples(db *gorm.DB) {
    // 1. 预加载 (Preload)
    var users []User
    db.Preload("Profile").Find(&users)

    // 2. 嵌套预加载
    var orders []Order
    db.Preload("User").Preload("OrderItems.Product").Find(&orders)

    // 3. 条件预加载
    db.Preload("OrderItems", "quantity > ?", 1).Find(&orders)

    // 4. 自定义预加载
    db.Preload("OrderItems", func(db *gorm.DB) *gorm.DB {
        return db.Order("price DESC")
    }).Find(&orders)

    // 5. 关联模式 (Association Mode)
    var user User
    db.First(&user, 1)

    // 添加关联
    var profile UserProfile
    db.Model(&user).Association("Profile").Append(&profile)

    // 替换关联
    db.Model(&user).Association("Profile").Replace(&profile)

    // 删除关联
    db.Model(&user).Association("Profile").Delete(&profile)

    // 清空关联
    db.Model(&user).Association("Profile").Clear()

    // 统计关联
    count := db.Model(&user).Association("Orders").Count()
    fmt.Printf("User has %d orders\n", count)
}
```

### 数据库迁移

```go
// 来自 mall-go/internal/database/migration.go
package database

import (
    "fmt"
    "log"

    "gorm.io/gorm"
    "mall-go/internal/model"
)

// 迁移管理器
type Migrator struct {
    db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
    return &Migrator{db: db}
}

// 自动迁移所有模型
func (m *Migrator) AutoMigrate() error {
    models := []interface{}{
        &model.User{},
        &model.UserProfile{},
        &model.Category{},
        &model.Product{},
        &model.Order{},
        &model.OrderItem{},
        &model.Review{},
    }

    for _, model := range models {
        if err := m.db.AutoMigrate(model); err != nil {
            return fmt.Errorf("failed to migrate %T: %w", model, err)
        }
        log.Printf("Migrated model: %T", model)
    }

    return nil
}

// 创建索引
func (m *Migrator) CreateIndexes() error {
    // 复合索引
    if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status)").Error; err != nil {
        return err
    }

    // 部分索引
    if err := m.db.Exec("CREATE INDEX IF NOT EXISTS idx_products_active ON products(status) WHERE status = 1").Error; err != nil {
        return err
    }

    // 全文索引 (MySQL)
    if err := m.db.Exec("CREATE FULLTEXT INDEX IF NOT EXISTS idx_products_search ON products(name, description)").Error; err != nil {
        return err
    }

    return nil
}

// 手动迁移示例
func (m *Migrator) ManualMigration() error {
    // 检查表是否存在
    if !m.db.Migrator().HasTable(&model.User{}) {
        // 创建表
        if err := m.db.Migrator().CreateTable(&model.User{}); err != nil {
            return err
        }
    }

    // 检查列是否存在
    if !m.db.Migrator().HasColumn(&model.User{}, "nickname") {
        // 添加列
        if err := m.db.Migrator().AddColumn(&model.User{}, "nickname"); err != nil {
            return err
        }
    }

    // 修改列类型
    if err := m.db.Migrator().AlterColumn(&model.User{}, "email"); err != nil {
        return err
    }

    // 重命名列
    if err := m.db.Migrator().RenameColumn(&model.User{}, "nick_name", "nickname"); err != nil {
        return err
    }

    // 删除列
    if err := m.db.Migrator().DropColumn(&model.User{}, "old_column"); err != nil {
        return err
    }

    // 创建索引
    if err := m.db.Migrator().CreateIndex(&model.User{}, "Email"); err != nil {
        return err
    }

    // 删除索引
    if err := m.db.Migrator().DropIndex(&model.User{}, "Email"); err != nil {
        return err
    }

    return nil
}

// 数据迁移 - 填充初始数据
func (m *Migrator) SeedData() error {
    // 创建默认分类
    categories := []model.Category{
        {Name: "电子产品", Slug: "electronics", Sort: 1},
        {Name: "图书", Slug: "books", Sort: 2},
        {Name: "服装", Slug: "clothing", Sort: 3},
    }

    for _, category := range categories {
        var count int64
        m.db.Model(&model.Category{}).Where("slug = ?", category.Slug).Count(&count)
        if count == 0 {
            if err := m.db.Create(&category).Error; err != nil {
                return err
            }
            log.Printf("Created category: %s", category.Name)
        }
    }

    // 创建管理员用户
    var adminCount int64
    m.db.Model(&model.User{}).Where("username = ?", "admin").Count(&adminCount)
    if adminCount == 0 {
        admin := model.User{
            Username: "admin",
            Email:    "admin@example.com",
            Password: "$2a$10$...", // 加密后的密码
            Nickname: "管理员",
            Status:   1,
        }

        if err := m.db.Create(&admin).Error; err != nil {
            return err
        }
        log.Println("Created admin user")
    }

    return nil
}

// 版本化迁移
type Migration struct {
    Version     string
    Description string
    Up          func(*gorm.DB) error
    Down        func(*gorm.DB) error
}

var migrations = []Migration{
    {
        Version:     "20240101_001",
        Description: "Create users table",
        Up: func(db *gorm.DB) error {
            return db.AutoMigrate(&model.User{})
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropTable(&model.User{})
        },
    },
    {
        Version:     "20240101_002",
        Description: "Add nickname column to users",
        Up: func(db *gorm.DB) error {
            return db.Migrator().AddColumn(&model.User{}, "nickname")
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropColumn(&model.User{}, "nickname")
        },
    },
}

// 执行迁移
func (m *Migrator) RunMigrations() error {
    // 创建迁移记录表
    if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
        return err
    }

    for _, migration := range migrations {
        var count int64
        m.db.Model(&MigrationRecord{}).Where("version = ?", migration.Version).Count(&count)

        if count == 0 {
            // 执行迁移
            if err := migration.Up(m.db); err != nil {
                return fmt.Errorf("migration %s failed: %w", migration.Version, err)
            }

            // 记录迁移
            record := MigrationRecord{
                Version:     migration.Version,
                Description: migration.Description,
                AppliedAt:   time.Now(),
            }

            if err := m.db.Create(&record).Error; err != nil {
                return err
            }

            log.Printf("Applied migration: %s", migration.Version)
        }
    }

    return nil
}

// 迁移记录模型
type MigrationRecord struct {
    ID          uint      `gorm:"primaryKey"`
    Version     string    `gorm:"uniqueIndex;size:50"`
    Description string    `gorm:"size:255"`
    AppliedAt   time.Time
}
```

---

## 🔧 CRUD操作详解

GORM提供了丰富的CRUD操作方法，让我们深入了解各种数据操作技巧。

### 创建记录

```go
// 来自 mall-go/internal/service/user_service.go
package service

import (
    "errors"
    "time"

    "gorm.io/gorm"
    "mall-go/internal/model"
)

type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{db: db}
}

// 1. 创建单条记录
func (s *UserService) CreateUser(user *model.User) error {
    // 基本创建
    result := s.db.Create(user)
    if result.Error != nil {
        return result.Error
    }

    // 检查影响行数
    if result.RowsAffected == 0 {
        return errors.New("no rows affected")
    }

    return nil
}

// 2. 批量创建
func (s *UserService) CreateUsers(users []model.User) error {
    // 批量创建 - 一次性插入
    result := s.db.Create(&users)
    return result.Error
}

// 3. 分批创建 - 避免单次插入过多数据
func (s *UserService) CreateUsersInBatches(users []model.User, batchSize int) error {
    result := s.db.CreateInBatches(&users, batchSize)
    return result.Error
}

// 4. 使用Map创建
func (s *UserService) CreateUserFromMap(userData map[string]interface{}) error {
    result := s.db.Model(&model.User{}).Create(userData)
    return result.Error
}

// 5. 创建时指定字段
func (s *UserService) CreateUserWithSelect(user *model.User) error {
    // 只创建指定字段
    result := s.db.Select("username", "email", "password").Create(user)
    return result.Error
}

// 6. 创建时忽略字段
func (s *UserService) CreateUserWithOmit(user *model.User) error {
    // 忽略指定字段
    result := s.db.Omit("created_at", "updated_at").Create(user)
    return result.Error
}

// 7. 冲突处理 - Upsert
func (s *UserService) UpsertUser(user *model.User) error {
    // MySQL: ON DUPLICATE KEY UPDATE
    result := s.db.Clauses(clause.OnConflict{
        Columns:   []clause.Column{{Name: "email"}},
        DoUpdates: clause.AssignmentColumns([]string{"username", "updated_at"}),
    }).Create(user)

    return result.Error
}

// 8. 创建关联数据
func (s *UserService) CreateUserWithProfile(user *model.User, profile *model.UserProfile) error {
    // 开启事务
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // 设置关联ID
        profile.UserID = user.ID

        // 创建用户资料
        if err := tx.Create(profile).Error; err != nil {
            return err
        }

        return nil
    })
}

// 9. 钩子函数示例
func (u *model.User) BeforeCreate(tx *gorm.DB) error {
    // 创建前的处理
    if u.Username == "" {
        return errors.New("username is required")
    }

    // 设置默认值
    if u.Status == 0 {
        u.Status = 1
    }

    return nil
}

func (u *model.User) AfterCreate(tx *gorm.DB) error {
    // 创建后的处理
    log.Printf("User created: %s", u.Username)

    // 创建默认用户资料
    profile := &model.UserProfile{
        UserID: u.ID,
        Bio:    "这个人很懒，什么都没留下",
    }

    return tx.Create(profile).Error
}
```

### 查询数据

```go
// 来自 mall-go/internal/service/product_service.go
package service

import (
    "gorm.io/gorm"
    "mall-go/internal/model"
)

type ProductService struct {
    db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
    return &ProductService{db: db}
}

// 1. 基础查询
func (s *ProductService) GetProduct(id uint) (*model.Product, error) {
    var product model.Product

    // 根据主键查询
    err := s.db.First(&product, id).Error
    if err != nil {
        return nil, err
    }

    return &product, nil
}

// 2. 条件查询
func (s *ProductService) FindProducts(conditions map[string]interface{}) ([]model.Product, error) {
    var products []model.Product

    query := s.db.Model(&model.Product{})

    // 动态添加条件
    for key, value := range conditions {
        switch key {
        case "name":
            query = query.Where("name LIKE ?", "%"+value.(string)+"%")
        case "category_id":
            query = query.Where("category_id = ?", value)
        case "min_price":
            query = query.Where("price >= ?", value)
        case "max_price":
            query = query.Where("price <= ?", value)
        case "status":
            query = query.Where("status = ?", value)
        }
    }

    err := query.Find(&products).Error
    return products, err
}

// 3. 复杂查询
func (s *ProductService) SearchProducts(keyword string, categoryID uint, minPrice, maxPrice float64, page, limit int) ([]model.Product, int64, error) {
    var products []model.Product
    var total int64

    query := s.db.Model(&model.Product{})

    // 关键词搜索
    if keyword != "" {
        query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
    }

    // 分类筛选
    if categoryID > 0 {
        query = query.Where("category_id = ?", categoryID)
    }

    // 价格范围
    if minPrice > 0 {
        query = query.Where("price >= ?", minPrice)
    }
    if maxPrice > 0 {
        query = query.Where("price <= ?", maxPrice)
    }

    // 只查询启用的产品
    query = query.Where("status = ?", 1)

    // 统计总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    offset := (page - 1) * limit
    err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&products).Error

    return products, total, err
}

// 4. 关联查询
func (s *ProductService) GetProductWithDetails(id uint) (*model.Product, error) {
    var product model.Product

    err := s.db.Preload("Categories").
        Preload("Reviews", func(db *gorm.DB) *gorm.DB {
            return db.Where("status = ?", 1).Order("created_at DESC").Limit(10)
        }).
        Preload("Reviews.User", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "username", "avatar")
        }).
        First(&product, id).Error

    return &product, err
}

// 5. 聚合查询
func (s *ProductService) GetProductStats() (map[string]interface{}, error) {
    var result struct {
        TotalProducts int64   `json:"total_products"`
        AvgPrice      float64 `json:"avg_price"`
        MaxPrice      float64 `json:"max_price"`
        MinPrice      float64 `json:"min_price"`
        TotalStock    int64   `json:"total_stock"`
    }

    err := s.db.Model(&model.Product{}).
        Select("COUNT(*) as total_products, AVG(price) as avg_price, MAX(price) as max_price, MIN(price) as min_price, SUM(stock) as total_stock").
        Where("status = ?", 1).
        Scan(&result).Error

    if err != nil {
        return nil, err
    }

    stats := map[string]interface{}{
        "total_products": result.TotalProducts,
        "avg_price":      result.AvgPrice,
        "max_price":      result.MaxPrice,
        "min_price":      result.MinPrice,
        "total_stock":    result.TotalStock,
    }

    return stats, nil
}

// 6. 分组查询
func (s *ProductService) GetProductsByCategory() ([]map[string]interface{}, error) {
    var results []struct {
        CategoryID   uint    `json:"category_id"`
        CategoryName string  `json:"category_name"`
        ProductCount int64   `json:"product_count"`
        AvgPrice     float64 `json:"avg_price"`
    }

    err := s.db.Table("products p").
        Select("p.category_id, c.name as category_name, COUNT(p.id) as product_count, AVG(p.price) as avg_price").
        Joins("LEFT JOIN categories c ON p.category_id = c.id").
        Where("p.status = ?", 1).
        Group("p.category_id, c.name").
        Having("COUNT(p.id) > 0").
        Order("product_count DESC").
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    // 转换为map切片
    var data []map[string]interface{}
    for _, result := range results {
        data = append(data, map[string]interface{}{
            "category_id":    result.CategoryID,
            "category_name":  result.CategoryName,
            "product_count":  result.ProductCount,
            "avg_price":      result.AvgPrice,
        })
    }

    return data, nil
}

// 7. 子查询
func (s *ProductService) GetPopularProducts(limit int) ([]model.Product, error) {
    var products []model.Product

    // 查询销量前N的产品
    subQuery := s.db.Model(&model.OrderItem{}).
        Select("product_id, SUM(quantity) as total_sales").
        Group("product_id").
        Order("total_sales DESC").
        Limit(limit)

    err := s.db.Model(&model.Product{}).
        Where("id IN (?)", subQuery).
        Find(&products).Error

    return products, err
}

// 8. 原生SQL查询
func (s *ProductService) GetProductSalesReport(startDate, endDate time.Time) ([]map[string]interface{}, error) {
    var results []map[string]interface{}

    sql := `
        SELECT
            p.id,
            p.name,
            p.price,
            COALESCE(SUM(oi.quantity), 0) as total_sold,
            COALESCE(SUM(oi.amount), 0) as total_revenue
        FROM products p
        LEFT JOIN order_items oi ON p.id = oi.product_id
        LEFT JOIN orders o ON oi.order_id = o.id
        WHERE o.created_at BETWEEN ? AND ? OR o.created_at IS NULL
        GROUP BY p.id, p.name, p.price
        ORDER BY total_revenue DESC
    `

    err := s.db.Raw(sql, startDate, endDate).Scan(&results).Error
    return results, err
}
```

### 更新记录

```go
// 来自 mall-go/internal/service/user_service.go (更新部分)
package service

// 1. 更新单个字段
func (s *UserService) UpdateUserStatus(id uint, status int8) error {
    result := s.db.Model(&model.User{}).Where("id = ?", id).Update("status", status)
    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }

    return nil
}

// 2. 更新多个字段
func (s *UserService) UpdateUser(id uint, updates map[string]interface{}) error {
    result := s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates)
    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }

    return nil
}

// 3. 使用结构体更新
func (s *UserService) UpdateUserByStruct(id uint, user *model.User) error {
    // 注意：零值字段不会被更新
    result := s.db.Model(&model.User{}).Where("id = ?", id).Updates(user)
    return result.Error
}

// 4. 更新指定字段
func (s *UserService) UpdateUserSelect(id uint, user *model.User) error {
    // 只更新指定字段，包括零值
    result := s.db.Model(&model.User{}).Where("id = ?", id).
        Select("username", "email", "status").Updates(user)
    return result.Error
}

// 5. 忽略字段更新
func (s *UserService) UpdateUserOmit(id uint, user *model.User) error {
    // 忽略指定字段
    result := s.db.Model(&model.User{}).Where("id = ?", id).
        Omit("created_at", "password").Updates(user)
    return result.Error
}

// 6. 批量更新
func (s *UserService) BatchUpdateUserStatus(ids []uint, status int8) error {
    result := s.db.Model(&model.User{}).Where("id IN ?", ids).Update("status", status)
    return result.Error
}

// 7. 条件更新
func (s *UserService) UpdateInactiveUsers() error {
    // 更新30天未登录的用户状态
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

    result := s.db.Model(&model.User{}).
        Where("last_login_at < ? OR last_login_at IS NULL", thirtyDaysAgo).
        Update("status", 0)

    log.Printf("Updated %d inactive users", result.RowsAffected)
    return result.Error
}

// 8. 使用SQL表达式更新
func (s *UserService) IncrementUserLoginCount(id uint) error {
    result := s.db.Model(&model.User{}).Where("id = ?", id).
        Updates(map[string]interface{}{
            "login_count":   gorm.Expr("login_count + ?", 1),
            "last_login_at": time.Now(),
        })

    return result.Error
}

// 9. 更新关联数据
func (s *UserService) UpdateUserWithProfile(id uint, user *model.User, profile *model.UserProfile) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 更新用户信息
        if err := tx.Model(&model.User{}).Where("id = ?", id).Updates(user).Error; err != nil {
            return err
        }

        // 更新或创建用户资料
        var existingProfile model.UserProfile
        err := tx.Where("user_id = ?", id).First(&existingProfile).Error

        if err == gorm.ErrRecordNotFound {
            // 创建新资料
            profile.UserID = id
            return tx.Create(profile).Error
        } else if err != nil {
            return err
        } else {
            // 更新现有资料
            return tx.Model(&existingProfile).Updates(profile).Error
        }
    })
}

// 10. 钩子函数示例
func (u *model.User) BeforeUpdate(tx *gorm.DB) error {
    // 更新前的处理
    u.UpdatedAt = time.Now()

    // 验证邮箱格式
    if u.Email != "" {
        if !isValidEmail(u.Email) {
            return errors.New("invalid email format")
        }
    }

    return nil
}

func (u *model.User) AfterUpdate(tx *gorm.DB) error {
    // 更新后的处理
    log.Printf("User updated: %s", u.Username)

    // 清除缓存
    cache.Delete(fmt.Sprintf("user:%d", u.ID))

    return nil
}
```

### 删除数据

```go
// 来自 mall-go/internal/service/user_service.go (删除部分)
package service

// 1. 软删除（推荐）
func (s *UserService) DeleteUser(id uint) error {
    // GORM默认软删除，只是设置deleted_at字段
    result := s.db.Delete(&model.User{}, id)
    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }

    return nil
}

// 2. 批量软删除
func (s *UserService) DeleteUsers(ids []uint) error {
    result := s.db.Delete(&model.User{}, ids)
    return result.Error
}

// 3. 条件删除
func (s *UserService) DeleteInactiveUsers(days int) error {
    cutoffDate := time.Now().AddDate(0, 0, -days)

    result := s.db.Where("last_login_at < ? AND status = ?", cutoffDate, 0).
        Delete(&model.User{})

    log.Printf("Deleted %d inactive users", result.RowsAffected)
    return result.Error
}

// 4. 物理删除（永久删除）
func (s *UserService) PermanentDeleteUser(id uint) error {
    // 使用Unscoped()进行物理删除
    result := s.db.Unscoped().Delete(&model.User{}, id)
    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }

    return nil
}

// 5. 恢复软删除的记录
func (s *UserService) RestoreUser(id uint) error {
    // 查询包含软删除的记录
    var user model.User
    err := s.db.Unscoped().Where("id = ?", id).First(&user).Error
    if err != nil {
        return err
    }

    // 恢复记录
    result := s.db.Unscoped().Model(&user).Update("deleted_at", nil)
    return result.Error
}

// 6. 级联删除
func (s *UserService) DeleteUserWithRelations(id uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 删除用户资料
        if err := tx.Where("user_id = ?", id).Delete(&model.UserProfile{}).Error; err != nil {
            return err
        }

        // 删除用户评论
        if err := tx.Where("user_id = ?", id).Delete(&model.Review{}).Error; err != nil {
            return err
        }

        // 软删除用户订单（保留历史记录）
        if err := tx.Model(&model.Order{}).Where("user_id = ?", id).
            Update("status", "cancelled").Error; err != nil {
            return err
        }

        // 删除用户
        if err := tx.Delete(&model.User{}, id).Error; err != nil {
            return err
        }

        return nil
    })
}

// 7. 查询软删除的记录
func (s *UserService) GetDeletedUsers() ([]model.User, error) {
    var users []model.User
    err := s.db.Unscoped().Where("deleted_at IS NOT NULL").Find(&users).Error
    return users, err
}

// 8. 清理软删除记录
func (s *UserService) CleanupDeletedUsers(days int) error {
    cutoffDate := time.Now().AddDate(0, 0, -days)

    // 物理删除超过指定天数的软删除记录
    result := s.db.Unscoped().
        Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoffDate).
        Delete(&model.User{})

    log.Printf("Permanently deleted %d users", result.RowsAffected)
    return result.Error
}

// 9. 钩子函数示例
func (u *model.User) BeforeDelete(tx *gorm.DB) error {
    // 删除前的处理
    log.Printf("Deleting user: %s", u.Username)

    // 检查是否有未完成的订单
    var pendingOrders int64
    tx.Model(&model.Order{}).Where("user_id = ? AND status IN ?", u.ID, []string{"pending", "processing"}).Count(&pendingOrders)

    if pendingOrders > 0 {
        return errors.New("cannot delete user with pending orders")
    }

    return nil
}

func (u *model.User) AfterDelete(tx *gorm.DB) error {
    // 删除后的处理
    log.Printf("User deleted: %s", u.Username)

    // 清除相关缓存
    cache.Delete(fmt.Sprintf("user:%d", u.ID))
    cache.Delete(fmt.Sprintf("user:profile:%d", u.ID))

    // 发送通知
    notification.Send("user_deleted", map[string]interface{}{
        "user_id":  u.ID,
        "username": u.Username,
    })

    return nil
}

// 辅助函数
func isValidEmail(email string) bool {
    // 简单的邮箱验证
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

---

## 🔗 高级查询技巧

GORM提供了强大的查询构建器，让我们掌握各种高级查询技巧。

### 条件查询

```go
// 来自 mall-go/internal/service/query_service.go
package service

import (
    "fmt"
    "strings"
    "time"

    "gorm.io/gorm"
    "mall-go/internal/model"
)

type QueryService struct {
    db *gorm.DB
}

func NewQueryService(db *gorm.DB) *QueryService {
    return &QueryService{db: db}
}

// 1. 基础条件查询
func (s *QueryService) BasicConditions() {
    var users []model.User

    // 等于条件
    s.db.Where("username = ?", "admin").Find(&users)

    // 不等于条件
    s.db.Where("username <> ?", "admin").Find(&users)

    // IN条件
    s.db.Where("id IN ?", []uint{1, 2, 3}).Find(&users)

    // LIKE条件
    s.db.Where("username LIKE ?", "%admin%").Find(&users)

    // 范围条件
    s.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&users)

    // 空值条件
    s.db.Where("deleted_at IS NULL").Find(&users)
    s.db.Where("phone IS NOT NULL").Find(&users)
}

// 2. 复合条件查询
func (s *QueryService) ComplexConditions() {
    var users []model.User

    // AND条件
    s.db.Where("status = ? AND gender = ?", 1, 1).Find(&users)

    // OR条件
    s.db.Where("username = ? OR email = ?", "admin", "admin@example.com").Find(&users)

    // 混合条件
    s.db.Where("(username = ? OR email = ?) AND status = ?", "admin", "admin@example.com", 1).Find(&users)

    // 使用结构体条件
    s.db.Where(&model.User{Status: 1, Gender: 1}).Find(&users)

    // 使用Map条件
    s.db.Where(map[string]interface{}{
        "status": 1,
        "gender": 1,
    }).Find(&users)
}

// 3. 动态查询构建器
func (s *QueryService) DynamicQuery(filters map[string]interface{}) ([]model.User, error) {
    var users []model.User

    query := s.db.Model(&model.User{})

    // 动态添加条件
    if username, ok := filters["username"]; ok && username != "" {
        query = query.Where("username LIKE ?", "%"+username.(string)+"%")
    }

    if email, ok := filters["email"]; ok && email != "" {
        query = query.Where("email = ?", email)
    }

    if status, ok := filters["status"]; ok {
        query = query.Where("status = ?", status)
    }

    if ageRange, ok := filters["age_range"]; ok {
        ages := ageRange.([]int)
        if len(ages) == 2 {
            query = query.Where("age BETWEEN ? AND ?", ages[0], ages[1])
        }
    }

    if createdAfter, ok := filters["created_after"]; ok {
        query = query.Where("created_at >= ?", createdAfter)
    }

    // 排序
    if sortBy, ok := filters["sort_by"]; ok {
        sortField := sortBy.(string)
        order := "ASC"
        if orderBy, ok := filters["order"]; ok {
            order = orderBy.(string)
        }
        query = query.Order(fmt.Sprintf("%s %s", sortField, order))
    } else {
        query = query.Order("created_at DESC")
    }

    // 分页
    if page, ok := filters["page"]; ok {
        if limit, ok := filters["limit"]; ok {
            offset := (page.(int) - 1) * limit.(int)
            query = query.Offset(offset).Limit(limit.(int))
        }
    }

    err := query.Find(&users).Error
    return users, err
}

// 4. 子查询
func (s *QueryService) SubQueries() {
    var users []model.User

    // 简单子查询
    s.db.Where("id IN (?)", s.db.Table("orders").Select("user_id").Where("status = ?", "completed")).Find(&users)

    // EXISTS子查询
    s.db.Where("EXISTS (?)", s.db.Table("orders").Select("1").Where("orders.user_id = users.id AND status = ?", "completed")).Find(&users)

    // 复杂子查询
    subQuery := s.db.Table("orders").
        Select("user_id, COUNT(*) as order_count").
        Where("status = ?", "completed").
        Group("user_id").
        Having("COUNT(*) > ?", 5)

    s.db.Where("id IN (?)", subQuery).Find(&users)
}

// 5. 窗口函数查询
func (s *QueryService) WindowFunctions() {
    var results []struct {
        UserID      uint    `json:"user_id"`
        Username    string  `json:"username"`
        OrderCount  int     `json:"order_count"`
        TotalAmount float64 `json:"total_amount"`
        Rank        int     `json:"rank"`
    }

    sql := `
        SELECT
            u.id as user_id,
            u.username,
            COUNT(o.id) as order_count,
            COALESCE(SUM(o.total_amount), 0) as total_amount,
            ROW_NUMBER() OVER (ORDER BY COALESCE(SUM(o.total_amount), 0) DESC) as rank
        FROM users u
        LEFT JOIN orders o ON u.id = o.user_id AND o.status = 'completed'
        GROUP BY u.id, u.username
        ORDER BY total_amount DESC
        LIMIT 10
    `

    s.db.Raw(sql).Scan(&results)
}
```

### 关联查询

```go
// 关联查询高级技巧
func (s *QueryService) AdvancedAssociations() {
    // 1. 预加载优化
    var orders []model.Order

    // 基础预加载
    s.db.Preload("User").Preload("OrderItems").Find(&orders)

    // 嵌套预加载
    s.db.Preload("User.Profile").Preload("OrderItems.Product").Find(&orders)

    // 条件预加载
    s.db.Preload("OrderItems", "quantity > ?", 1).Find(&orders)

    // 自定义预加载
    s.db.Preload("OrderItems", func(db *gorm.DB) *gorm.DB {
        return db.Order("price DESC").Limit(5)
    }).Find(&orders)

    // 2. 连接查询
    var results []struct {
        OrderID     uint    `json:"order_id"`
        OrderNo     string  `json:"order_no"`
        Username    string  `json:"username"`
        TotalAmount float64 `json:"total_amount"`
    }

    s.db.Table("orders o").
        Select("o.id as order_id, o.order_no, u.username, o.total_amount").
        Joins("LEFT JOIN users u ON o.user_id = u.id").
        Where("o.status = ?", "completed").
        Scan(&results)

    // 3. 关联统计
    var users []model.User
    s.db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
        return db.Select("user_id, COUNT(*) as order_count, SUM(total_amount) as total_spent").
            Group("user_id")
    }).Find(&users)

    // 4. 多对多关联
    var products []model.Product
    s.db.Preload("Categories", func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", 1)
    }).Find(&products)

    // 5. 关联模式操作
    var user model.User
    s.db.First(&user, 1)

    // 添加关联
    var categories []model.Category
    s.db.Find(&categories, []uint{1, 2, 3})
    s.db.Model(&user).Association("Categories").Append(&categories)

    // 替换关联
    s.db.Model(&user).Association("Categories").Replace(&categories)

    // 删除关联
    s.db.Model(&user).Association("Categories").Delete(&categories[0])

    // 清空关联
    s.db.Model(&user).Association("Categories").Clear()

    // 统计关联
    count := s.db.Model(&user).Association("Orders").Count()
    fmt.Printf("User has %d orders\n", count)
}
```

---

## ⚡ 事务与并发

事务是保证数据一致性的重要机制，GORM提供了多种事务处理方式。

### 事务处理

```go
// 来自 mall-go/internal/service/transaction_service.go
package service

import (
    "errors"
    "fmt"

    "gorm.io/gorm"
    "mall-go/internal/model"
)

type TransactionService struct {
    db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
    return &TransactionService{db: db}
}

// 1. 自动事务（推荐）
func (s *TransactionService) CreateOrderWithItems(order *model.Order, items []model.OrderItem) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建订单
        if err := tx.Create(order).Error; err != nil {
            return err
        }

        // 设置订单项的订单ID
        for i := range items {
            items[i].OrderID = order.ID
        }

        // 创建订单项
        if err := tx.Create(&items).Error; err != nil {
            return err
        }

        // 更新产品库存
        for _, item := range items {
            result := tx.Model(&model.Product{}).
                Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                Update("stock", gorm.Expr("stock - ?", item.Quantity))

            if result.Error != nil {
                return result.Error
            }

            if result.RowsAffected == 0 {
                return fmt.Errorf("insufficient stock for product %d", item.ProductID)
            }
        }

        return nil
    })
}

// 2. 手动事务
func (s *TransactionService) ManualTransaction(order *model.Order, items []model.OrderItem) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if tx.Error != nil {
        return tx.Error
    }

    // 创建订单
    if err := tx.Create(order).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 设置订单项的订单ID
    for i := range items {
        items[i].OrderID = order.ID
    }

    // 创建订单项
    if err := tx.Create(&items).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 更新库存
    for _, item := range items {
        result := tx.Model(&model.Product{}).
            Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
            Update("stock", gorm.Expr("stock - ?", item.Quantity))

        if result.Error != nil {
            tx.Rollback()
            return result.Error
        }

        if result.RowsAffected == 0 {
            tx.Rollback()
            return fmt.Errorf("insufficient stock for product %d", item.ProductID)
        }
    }

    return tx.Commit().Error
}

// 3. 嵌套事务
func (s *TransactionService) NestedTransaction() error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 外层事务操作
        user := &model.User{Username: "test", Email: "test@example.com"}
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // 嵌套事务
        return tx.Transaction(func(tx2 *gorm.DB) error {
            profile := &model.UserProfile{UserID: user.ID, Bio: "Test user"}
            return tx2.Create(profile).Error
        })
    })
}

// 4. 保存点（Savepoint）
func (s *TransactionService) SavepointTransaction() error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 创建用户
    user := &model.User{Username: "test", Email: "test@example.com"}
    if err := tx.Create(user).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 创建保存点
    sp := tx.SavePoint("sp1")
    if sp.Error != nil {
        tx.Rollback()
        return sp.Error
    }

    // 尝试创建资料
    profile := &model.UserProfile{UserID: user.ID, Bio: "Test user"}
    if err := tx.Create(profile).Error; err != nil {
        // 回滚到保存点
        tx.RollbackTo("sp1")
        // 继续其他操作...
    }

    return tx.Commit().Error
}

// 5. 分布式事务示例
func (s *TransactionService) DistributedTransaction(orderData *model.Order, paymentData map[string]interface{}) error {
    // 本地事务
    localTx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            localTx.Rollback()
        }
    }()

    // 创建订单
    if err := localTx.Create(orderData).Error; err != nil {
        localTx.Rollback()
        return err
    }

    // 调用支付服务
    paymentResult, err := s.callPaymentService(paymentData)
    if err != nil {
        localTx.Rollback()
        return err
    }

    // 更新订单支付状态
    if err := localTx.Model(orderData).Update("pay_status", "paid").Error; err != nil {
        localTx.Rollback()
        // 补偿：取消支付
        s.cancelPayment(paymentResult.TransactionID)
        return err
    }

    return localTx.Commit().Error
}

// 6. 事务钩子
func (s *TransactionService) TransactionHooks() error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 事务开始钩子
        tx.Callback().Create().Before("gorm:create").Register("before_create_in_tx", func(db *gorm.DB) {
            fmt.Println("Before create in transaction")
        })

        // 事务提交钩子
        tx.Callback().Create().After("gorm:create").Register("after_create_in_tx", func(db *gorm.DB) {
            fmt.Println("After create in transaction")
        })

        user := &model.User{Username: "test", Email: "test@example.com"}
        return tx.Create(user).Error
    })
}

// 辅助方法
func (s *TransactionService) callPaymentService(data map[string]interface{}) (*PaymentResult, error) {
    // 模拟调用外部支付服务
    return &PaymentResult{TransactionID: "tx_123456"}, nil
}

func (s *TransactionService) cancelPayment(transactionID string) error {
    // 模拟取消支付
    fmt.Printf("Cancelling payment: %s\n", transactionID)
    return nil
}

type PaymentResult struct {
    TransactionID string
}
```

---

## 🚀 性能优化

数据库性能优化是企业级应用的关键，让我们掌握GORM的性能优化技巧。

### 查询优化

```go
// 来自 mall-go/internal/service/optimization_service.go
package service

import (
    "context"
    "time"

    "gorm.io/gorm"
    "mall-go/internal/model"
)

type OptimizationService struct {
    db *gorm.DB
}

func NewOptimizationService(db *gorm.DB) *OptimizationService {
    return &OptimizationService{db: db}
}

// 1. 索引优化
func (s *OptimizationService) IndexOptimization() {
    // 创建复合索引
    s.db.Exec("CREATE INDEX idx_orders_user_status ON orders(user_id, status)")

    // 创建部分索引
    s.db.Exec("CREATE INDEX idx_products_active ON products(name) WHERE status = 1")

    // 创建表达式索引
    s.db.Exec("CREATE INDEX idx_users_lower_email ON users(LOWER(email))")

    // 查看索引使用情况
    var result []map[string]interface{}
    s.db.Raw("EXPLAIN SELECT * FROM orders WHERE user_id = ? AND status = ?", 1, "completed").Scan(&result)
}

// 2. 查询优化
func (s *OptimizationService) QueryOptimization() {
    // ❌ 错误：N+1查询问题
    var orders []model.Order
    s.db.Find(&orders)
    for _, order := range orders {
        var user model.User
        s.db.First(&user, order.UserID) // 每个订单都查询一次用户
    }

    // ✅ 正确：使用预加载
    s.db.Preload("User").Find(&orders)

    // ❌ 错误：查询不必要的字段
    s.db.Find(&orders)

    // ✅ 正确：只查询需要的字段
    s.db.Select("id", "order_no", "user_id", "total_amount").Find(&orders)

    // ❌ 错误：使用OFFSET进行深度分页
    s.db.Offset(10000).Limit(20).Find(&orders)

    // ✅ 正确：使用游标分页
    var lastID uint = 10000
    s.db.Where("id > ?", lastID).Limit(20).Order("id ASC").Find(&orders)
}

// 3. 批量操作优化
func (s *OptimizationService) BatchOptimization() {
    // ❌ 错误：逐条插入
    users := make([]model.User, 1000)
    for i, user := range users {
        s.db.Create(&user) // 1000次数据库调用
    }

    // ✅ 正确：批量插入
    s.db.CreateInBatches(&users, 100) // 10次数据库调用

    // ❌ 错误：逐条更新
    var userIDs []uint
    s.db.Model(&model.User{}).Pluck("id", &userIDs)
    for _, id := range userIDs {
        s.db.Model(&model.User{}).Where("id = ?", id).Update("status", 1)
    }

    // ✅ 正确：批量更新
    s.db.Model(&model.User{}).Where("id IN ?", userIDs).Update("status", 1)
}

// 4. 连接池优化
func (s *OptimizationService) ConnectionPoolOptimization() {
    sqlDB, err := s.db.DB()
    if err != nil {
        return
    }

    // 设置最大空闲连接数
    sqlDB.SetMaxIdleConns(10)

    // 设置最大打开连接数
    sqlDB.SetMaxOpenConns(100)

    // 设置连接最大生存时间
    sqlDB.SetConnMaxLifetime(time.Hour)

    // 设置连接最大空闲时间
    sqlDB.SetConnMaxIdleTime(time.Minute * 30)
}

// 5. 缓存策略
func (s *OptimizationService) CacheStrategy() {
    // 查询缓存示例
    var user model.User
    cacheKey := fmt.Sprintf("user:%d", 1)

    // 先从缓存获取
    if cached := cache.Get(cacheKey); cached != nil {
        json.Unmarshal(cached.([]byte), &user)
        return
    }

    // 缓存未命中，查询数据库
    if err := s.db.First(&user, 1).Error; err != nil {
        return
    }

    // 写入缓存
    if data, err := json.Marshal(user); err == nil {
        cache.Set(cacheKey, data, time.Hour)
    }
}

// 6. 读写分离
func (s *OptimizationService) ReadWriteSeparation() {
    // 配置读写分离
    masterDB, _ := gorm.Open(mysql.Open(masterDSN), &gorm.Config{})
    slaveDB, _ := gorm.Open(mysql.Open(slaveDSN), &gorm.Config{})

    // 写操作使用主库
    masterDB.Create(&model.User{Username: "test"})

    // 读操作使用从库
    var users []model.User
    slaveDB.Find(&users)
}

// 7. 分表分库
func (s *OptimizationService) Sharding() {
    // 根据用户ID分表
    userID := uint(12345)
    tableIndex := userID % 10
    tableName := fmt.Sprintf("orders_%d", tableIndex)

    var orders []model.Order
    s.db.Table(tableName).Where("user_id = ?", userID).Find(&orders)
}

// 8. 慢查询监控
func (s *OptimizationService) SlowQueryMonitoring() {
    // 自定义Logger监控慢查询
    slowLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold: time.Second,   // 慢查询阈值
            LogLevel:      logger.Warn,   // 日志级别
            Colorful:      true,          // 彩色输出
        },
    )

    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: slowLogger,
    })

    // 使用中间件记录查询时间
    db.Callback().Query().Before("gorm:query").Register("query_time_start", func(db *gorm.DB) {
        db.Set("query_start_time", time.Now())
    })

    db.Callback().Query().After("gorm:query").Register("query_time_end", func(db *gorm.DB) {
        if startTime, ok := db.Get("query_start_time"); ok {
            duration := time.Since(startTime.(time.Time))
            if duration > time.Second {
                log.Printf("Slow query detected: %v, SQL: %s", duration, db.Statement.SQL.String())
            }
        }
    })
}
```

### 内存优化

```go
// 内存优化技巧
func (s *OptimizationService) MemoryOptimization() {
    // 1. 使用Scan避免创建完整模型
    var results []struct {
        ID   uint   `json:"id"`
        Name string `json:"name"`
    }
    s.db.Model(&model.User{}).Select("id", "username as name").Scan(&results)

    // 2. 流式查询处理大数据集
    rows, err := s.db.Model(&model.User{}).Rows()
    if err != nil {
        return
    }
    defer rows.Close()

    for rows.Next() {
        var user model.User
        s.db.ScanRows(rows, &user)
        // 处理单条记录
        processUser(&user)
    }

    // 3. 分页处理大数据集
    pageSize := 1000
    var lastID uint = 0

    for {
        var users []model.User
        err := s.db.Where("id > ?", lastID).Limit(pageSize).Order("id ASC").Find(&users).Error
        if err != nil {
            break
        }

        if len(users) == 0 {
            break
        }

        // 处理当前批次
        for _, user := range users {
            processUser(&user)
        }

        lastID = users[len(users)-1].ID
    }

    // 4. 使用Context控制查询超时
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    var users []model.User
    s.db.WithContext(ctx).Find(&users)
}

func processUser(user *model.User) {
    // 处理用户数据
    fmt.Printf("Processing user: %s\n", user.Username)
}
```

---

## 🏢 实战案例分析

让我们通过mall-go项目的真实案例，看看如何在企业级项目中应用GORM。

### 电商订单系统设计

```go
// 来自 mall-go/internal/service/order_service.go
package service

import (
    "errors"
    "fmt"
    "time"

    "gorm.io/gorm"
    "mall-go/internal/model"
)

type OrderService struct {
    db             *gorm.DB
    productService *ProductService
    userService    *UserService
}

func NewOrderService(db *gorm.DB, productService *ProductService, userService *UserService) *OrderService {
    return &OrderService{
        db:             db,
        productService: productService,
        userService:    userService,
    }
}

// 创建订单 - 复杂业务逻辑
func (s *OrderService) CreateOrder(userID uint, items []CreateOrderItem, address ShippingAddress) (*model.Order, error) {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 验证用户
        user, err := s.userService.GetUser(userID)
        if err != nil {
            return fmt.Errorf("user not found: %w", err)
        }

        if user.Status != 1 {
            return errors.New("user account is disabled")
        }

        // 2. 验证商品和库存
        var orderItems []model.OrderItem
        var totalAmount float64

        for _, item := range items {
            product, err := s.productService.GetProduct(item.ProductID)
            if err != nil {
                return fmt.Errorf("product %d not found: %w", item.ProductID, err)
            }

            if product.Status != 1 {
                return fmt.Errorf("product %s is not available", product.Name)
            }

            if product.Stock < item.Quantity {
                return fmt.Errorf("insufficient stock for product %s", product.Name)
            }

            // 计算金额
            amount := product.Price * float64(item.Quantity)
            totalAmount += amount

            orderItem := model.OrderItem{
                ProductID:    product.ID,
                Quantity:     item.Quantity,
                Price:        product.Price,
                Amount:       amount,
                ProductName:  product.Name,
                ProductImage: product.Images["thumbnail"].(string),
                ProductSKU:   product.SKU,
            }

            orderItems = append(orderItems, orderItem)
        }

        // 3. 创建订单
        order := &model.Order{
            OrderNo:         generateOrderNo(),
            UserID:          userID,
            TotalAmount:     totalAmount,
            PayAmount:       totalAmount, // 简化处理，实际可能有优惠
            Status:          "pending",
            PayStatus:       "unpaid",
            ReceiverName:    address.Name,
            ReceiverPhone:   address.Phone,
            ReceiverAddress: address.Address,
        }

        if err := tx.Create(order).Error; err != nil {
            return fmt.Errorf("failed to create order: %w", err)
        }

        // 4. 创建订单项
        for i := range orderItems {
            orderItems[i].OrderID = order.ID
        }

        if err := tx.Create(&orderItems).Error; err != nil {
            return fmt.Errorf("failed to create order items: %w", err)
        }

        // 5. 扣减库存
        for _, item := range items {
            result := tx.Model(&model.Product{}).
                Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                Update("stock", gorm.Expr("stock - ?", item.Quantity))

            if result.Error != nil {
                return fmt.Errorf("failed to update stock: %w", result.Error)
            }

            if result.RowsAffected == 0 {
                return fmt.Errorf("concurrent stock update conflict for product %d", item.ProductID)
            }
        }

        // 6. 记录库存变更日志
        for _, item := range orderItems {
            stockLog := &StockLog{
                ProductID:   item.ProductID,
                OrderID:     order.ID,
                Type:        "order",
                Quantity:    -item.Quantity,
                Reason:      "订单扣减",
                OperatorID:  userID,
            }

            if err := tx.Create(stockLog).Error; err != nil {
                return fmt.Errorf("failed to create stock log: %w", err)
            }
        }

        return nil
    })
}

// 订单支付
func (s *OrderService) PayOrder(orderID uint, payMethod string) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 获取订单
        var order model.Order
        if err := tx.Where("id = ? AND pay_status = ?", orderID, "unpaid").First(&order).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                return errors.New("order not found or already paid")
            }
            return err
        }

        // 2. 检查订单状态
        if order.Status == "cancelled" {
            return errors.New("cannot pay cancelled order")
        }

        // 3. 调用支付服务（模拟）
        paymentResult, err := s.processPayment(order.PayAmount, payMethod)
        if err != nil {
            return fmt.Errorf("payment failed: %w", err)
        }

        // 4. 更新订单状态
        now := time.Now()
        updates := map[string]interface{}{
            "pay_status": "paid",
            "pay_method": payMethod,
            "pay_time":   &now,
            "status":     "paid",
        }

        if err := tx.Model(&order).Updates(updates).Error; err != nil {
            // 支付成功但更新失败，需要补偿
            s.refundPayment(paymentResult.TransactionID)
            return fmt.Errorf("failed to update order status: %w", err)
        }

        // 5. 增加商品销量
        var orderItems []model.OrderItem
        if err := tx.Where("order_id = ?", orderID).Find(&orderItems).Error; err != nil {
            return err
        }

        for _, item := range orderItems {
            tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
                Update("sales", gorm.Expr("sales + ?", item.Quantity))
        }

        return nil
    })
}

// 订单取消
func (s *OrderService) CancelOrder(orderID uint, reason string) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 获取订单
        var order model.Order
        if err := tx.Preload("OrderItems").First(&order, orderID).Error; err != nil {
            return err
        }

        // 2. 检查是否可以取消
        if !s.canCancelOrder(&order) {
            return errors.New("order cannot be cancelled")
        }

        // 3. 更新订单状态
        updates := map[string]interface{}{
            "status":       "cancelled",
            "admin_remark": reason,
        }

        if err := tx.Model(&order).Updates(updates).Error; err != nil {
            return err
        }

        // 4. 恢复库存
        for _, item := range order.OrderItems {
            tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
                Update("stock", gorm.Expr("stock + ?", item.Quantity))

            // 记录库存变更
            stockLog := &StockLog{
                ProductID:  item.ProductID,
                OrderID:    order.ID,
                Type:       "cancel",
                Quantity:   item.Quantity,
                Reason:     "订单取消",
                OperatorID: order.UserID,
            }
            tx.Create(stockLog)
        }

        // 5. 处理退款
        if order.PayStatus == "paid" {
            if err := s.processRefund(&order); err != nil {
                return fmt.Errorf("refund failed: %w", err)
            }
        }

        return nil
    })
}

// 辅助方法
func (s *OrderService) canCancelOrder(order *model.Order) bool {
    // 已发货的订单不能取消
    if order.Status == "shipped" || order.Status == "completed" {
        return false
    }

    // 已取消的订单不能再次取消
    if order.Status == "cancelled" {
        return false
    }

    return true
}

func (s *OrderService) processPayment(amount float64, method string) (*PaymentResult, error) {
    // 模拟支付处理
    return &PaymentResult{
        TransactionID: fmt.Sprintf("tx_%d", time.Now().Unix()),
        Amount:        amount,
        Status:        "success",
    }, nil
}

func (s *OrderService) refundPayment(transactionID string) error {
    // 模拟退款处理
    fmt.Printf("Processing refund for transaction: %s\n", transactionID)
    return nil
}

func (s *OrderService) processRefund(order *model.Order) error {
    // 模拟退款处理
    fmt.Printf("Processing refund for order: %s, amount: %.2f\n", order.OrderNo, order.PayAmount)
    return nil
}

func generateOrderNo() string {
    return fmt.Sprintf("ORD%d", time.Now().UnixNano())
}

// 相关结构体
type CreateOrderItem struct {
    ProductID uint `json:"product_id"`
    Quantity  int  `json:"quantity"`
}

type ShippingAddress struct {
    Name    string `json:"name"`
    Phone   string `json:"phone"`
    Address string `json:"address"`
}

type StockLog struct {
    ID         uint      `gorm:"primaryKey"`
    ProductID  uint      `gorm:"not null;index"`
    OrderID    uint      `gorm:"index"`
    Type       string    `gorm:"size:20;not null"`
    Quantity   int       `gorm:"not null"`
    Reason     string    `gorm:"size:100"`
    OperatorID uint      `gorm:"not null"`
    CreatedAt  time.Time
}

type PaymentResult struct {
    TransactionID string
    Amount        float64
    Status        string
}
```

---

## 🎯 面试常考点

### 1. GORM的核心特性和优势

**问题：** GORM相比其他Go ORM框架有什么优势？

**答案：**
```go
/*
GORM的核心优势：

1. 功能完整：
   - 完整的ORM功能：CRUD、关联、事务、迁移
   - 丰富的查询方法：链式调用、条件构建
   - 自动化功能：时间戳、软删除、钩子函数

2. 性能优异：
   - 预编译语句缓存
   - 连接池管理
   - 批量操作优化
   - 懒加载和预加载

3. 开发友好：
   - 简洁的API设计
   - 强类型支持
   - 自动迁移
   - 丰富的标签支持

4. 生态完善：
   - 多数据库支持
   - 插件系统
   - 社区活跃
   - 文档完善
*/

// 性能对比示例
func BenchmarkGORMvsSQL(b *testing.B) {
    // GORM查询
    b.Run("GORM", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var user model.User
            db.First(&user, 1)
        }
    })

    // 原生SQL查询
    b.Run("SQL", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var user model.User
            db.Raw("SELECT * FROM users WHERE id = ?", 1).Scan(&user)
        }
    })
}
```

### 2. 关联关系的实现原理

**问题：** GORM中的关联关系是如何实现的？

**答案：**
```go
// 关联关系实现原理

// 1. Has One - 一对一关系
type User struct {
    ID      uint        `gorm:"primaryKey"`
    Profile UserProfile `gorm:"foreignKey:UserID"` // 外键在Profile表
}

type UserProfile struct {
    ID     uint `gorm:"primaryKey"`
    UserID uint `gorm:"uniqueIndex"` // 外键字段
}

// 2. Has Many - 一对多关系
type User struct {
    ID     uint    `gorm:"primaryKey"`
    Orders []Order `gorm:"foreignKey:UserID"` // 外键在Order表
}

type Order struct {
    ID     uint `gorm:"primaryKey"`
    UserID uint `gorm:"index"` // 外键字段
}

// 3. Many To Many - 多对多关系
type Product struct {
    ID   uint `gorm:"primaryKey"`
    Tags []Tag `gorm:"many2many:product_tags;"` // 中间表
}

type Tag struct {
    ID       uint      `gorm:"primaryKey"`
    Products []Product `gorm:"many2many:product_tags;"`
}

/*
实现原理：
1. 外键约束：通过外键字段建立表间关系
2. 中间表：多对多关系通过中间表实现
3. 预加载：通过JOIN或子查询获取关联数据
4. 懒加载：按需查询关联数据
*/
```

### 3. 事务的ACID特性

**问题：** 如何在GORM中保证事务的ACID特性？

**答案：**
```go
// ACID特性保证

// 1. 原子性（Atomicity）- 事务要么全部成功，要么全部失败
func AtomicityExample(db *gorm.DB) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // 操作1：创建用户
        user := &User{Username: "test"}
        if err := tx.Create(user).Error; err != nil {
            return err // 自动回滚
        }

        // 操作2：创建订单
        order := &Order{UserID: user.ID}
        if err := tx.Create(order).Error; err != nil {
            return err // 自动回滚，用户创建也会被撤销
        }

        return nil // 提交事务
    })
}

// 2. 一致性（Consistency）- 数据库约束保证
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex;not null"` // 唯一约束
    Age   int    `gorm:"check:age >= 0"`       // 检查约束
}

// 3. 隔离性（Isolation）- 事务隔离级别
func IsolationExample(db *gorm.DB) {
    // 设置事务隔离级别
    db.Exec("SET TRANSACTION ISOLATION LEVEL READ COMMITTED")

    // 使用行锁防止并发修改
    var product Product
    db.Set("gorm:query_option", "FOR UPDATE").First(&product, 1)
}

// 4. 持久性（Durability）- 数据持久化保证
func DurabilityExample(db *gorm.DB) {
    // 同步写入，确保数据持久化
    db.Exec("SET sync_binlog = 1")
    db.Exec("SET innodb_flush_log_at_trx_commit = 1")
}
```

### 4. 性能优化策略

**问题：** GORM应用中常见的性能问题及解决方案？

**答案：**
```go
// 常见性能问题及解决方案

// 1. N+1查询问题
// ❌ 问题代码
func BadQuery(db *gorm.DB) {
    var orders []Order
    db.Find(&orders)
    for _, order := range orders {
        var user User
        db.First(&user, order.UserID) // N次查询
    }
}

// ✅ 解决方案
func GoodQuery(db *gorm.DB) {
    var orders []Order
    db.Preload("User").Find(&orders) // 1次查询
}

// 2. 索引优化
func IndexOptimization(db *gorm.DB) {
    // 创建复合索引
    db.Exec("CREATE INDEX idx_orders_user_status ON orders(user_id, status)")

    // 使用索引提示
    var orders []Order
    db.Set("gorm:query_option", "USE INDEX (idx_orders_user_status)").
        Where("user_id = ? AND status = ?", 1, "completed").Find(&orders)
}

// 3. 批量操作优化
func BatchOptimization(db *gorm.DB) {
    users := make([]User, 1000)

    // 批量插入
    db.CreateInBatches(&users, 100)

    // 批量更新
    db.Model(&User{}).Where("status = ?", 0).Update("status", 1)
}

// 4. 查询字段优化
func FieldOptimization(db *gorm.DB) {
    // 只查询需要的字段
    var users []User
    db.Select("id", "username", "email").Find(&users)

    // 使用Scan避免完整模型
    var results []struct {
        ID   uint   `json:"id"`
        Name string `json:"name"`
    }
    db.Model(&User{}).Select("id", "username as name").Scan(&results)
}
```

---

## ⚠️ 踩坑提醒

### 1. 模型定义陷阱

```go
// ❌ 错误：忘记设置主键
type BadModel struct {
    Name string
}

// ✅ 正确：明确设置主键
type GoodModel struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
}

// ❌ 错误：外键字段类型不匹配
type User struct {
    ID uint `gorm:"primaryKey"`
}

type BadOrder struct {
    ID     uint `gorm:"primaryKey"`
    UserID int  `gorm:"index"` // 类型不匹配
}

// ✅ 正确：外键字段类型匹配
type GoodOrder struct {
    ID     uint `gorm:"primaryKey"`
    UserID uint `gorm:"index"` // 类型匹配
}
```

### 2. 查询陷阱

```go
// ❌ 错误：零值查询问题
func BadZeroValueQuery(db *gorm.DB) {
    user := User{Status: 0} // 零值
    var users []User
    db.Where(&user).Find(&users) // 零值字段被忽略
}

// ✅ 正确：使用Map或指定字段
func GoodZeroValueQuery(db *gorm.DB) {
    var users []User
    db.Where(map[string]interface{}{"status": 0}).Find(&users)
    // 或者
    db.Where("status = ?", 0).Find(&users)
}

// ❌ 错误：预加载条件错误
func BadPreload(db *gorm.DB) {
    var users []User
    db.Preload("Orders", "status = completed").Find(&users) // 缺少引号
}

// ✅ 正确：预加载条件正确
func GoodPreload(db *gorm.DB) {
    var users []User
    db.Preload("Orders", "status = ?", "completed").Find(&users)
}
```

### 3. 事务陷阱

```go
// ❌ 错误：忘记处理事务回滚
func BadTransaction(db *gorm.DB) {
    tx := db.Begin()

    if err := tx.Create(&user).Error; err != nil {
        return err // 忘记回滚
    }

    tx.Commit()
}

// ✅ 正确：正确处理事务
func GoodTransaction(db *gorm.DB) error {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}

// ✅ 更好：使用自动事务
func BestTransaction(db *gorm.DB) error {
    return db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&user).Error; err != nil {
            return err // 自动回滚
        }
        return nil // 自动提交
    })
}
```

### 4. 性能陷阱

```go
// ❌ 错误：在循环中执行数据库操作
func BadLoop(db *gorm.DB) {
    userIDs := []uint{1, 2, 3, 4, 5}
    for _, id := range userIDs {
        var user User
        db.First(&user, id) // N次查询
    }
}

// ✅ 正确：批量查询
func GoodBatch(db *gorm.DB) {
    userIDs := []uint{1, 2, 3, 4, 5}
    var users []User
    db.Where("id IN ?", userIDs).Find(&users) // 1次查询
}

// ❌ 错误：查询所有字段
func BadSelect(db *gorm.DB) {
    var users []User
    db.Find(&users) // 查询所有字段
}

// ✅ 正确：只查询需要的字段
func GoodSelect(db *gorm.DB) {
    var users []User
    db.Select("id", "username", "email").Find(&users)
}
```

---

## 📝 练习题

### 练习题1：用户管理系统（⭐⭐）

**题目描述：**
设计一个用户管理系统，包含用户、角色、权限的多对多关系，实现完整的CRUD操作。

```go
// 练习题1：用户管理系统
package main

import (
    "time"
    "gorm.io/gorm"
)

// 解答：
// 1. 模型定义
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex;size:50;not null"`
    Email     string    `gorm:"uniqueIndex;size:100;not null"`
    Password  string    `gorm:"size:255;not null"`
    Status    int8      `gorm:"default:1"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`

    // 多对多关系
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"uniqueIndex;size:50;not null"`
    Description string `gorm:"size:255"`
    Status      int8   `gorm:"default:1"`
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // 多对多关系
    Users       []User       `gorm:"many2many:user_roles;"`
    Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type Permission struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"uniqueIndex;size:50;not null"`
    Code        string `gorm:"uniqueIndex;size:100;not null"`
    Description string `gorm:"size:255"`
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // 多对多关系
    Roles []Role `gorm:"many2many:role_permissions;"`
}

// 2. 服务层实现
type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{db: db}
}

// 创建用户并分配角色
func (s *UserService) CreateUserWithRoles(user *User, roleIDs []uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // 查询角色
        var roles []Role
        if err := tx.Where("id IN ? AND status = ?", roleIDs, 1).Find(&roles).Error; err != nil {
            return err
        }

        // 分配角色
        if err := tx.Model(user).Association("Roles").Append(&roles); err != nil {
            return err
        }

        return nil
    })
}

// 获取用户及其角色权限
func (s *UserService) GetUserWithPermissions(userID uint) (*User, error) {
    var user User
    err := s.db.Preload("Roles.Permissions").First(&user, userID).Error
    return &user, err
}

// 检查用户权限
func (s *UserService) CheckUserPermission(userID uint, permissionCode string) (bool, error) {
    var count int64
    err := s.db.Table("users u").
        Joins("JOIN user_roles ur ON u.id = ur.user_id").
        Joins("JOIN roles r ON ur.role_id = r.id").
        Joins("JOIN role_permissions rp ON r.id = rp.role_id").
        Joins("JOIN permissions p ON rp.permission_id = p.id").
        Where("u.id = ? AND p.code = ? AND u.status = ? AND r.status = ?",
              userID, permissionCode, 1, 1).
        Count(&count).Error

    return count > 0, err
}

// 更新用户角色
func (s *UserService) UpdateUserRoles(userID uint, roleIDs []uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var user User
        if err := tx.First(&user, userID).Error; err != nil {
            return err
        }

        // 查询新角色
        var roles []Role
        if err := tx.Where("id IN ? AND status = ?", roleIDs, 1).Find(&roles).Error; err != nil {
            return err
        }

        // 替换角色
        if err := tx.Model(&user).Association("Roles").Replace(&roles); err != nil {
            return err
        }

        return nil
    })
}

// 删除用户（软删除）
func (s *UserService) DeleteUser(userID uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 清除用户角色关联
        if err := tx.Model(&User{}).Where("id = ?", userID).Association("Roles").Clear(); err != nil {
            return err
        }

        // 软删除用户
        if err := tx.Delete(&User{}, userID).Error; err != nil {
            return err
        }

        return nil
    })
}

// 3. 测试函数
func TestUserManagement(db *gorm.DB) {
    // 自动迁移
    db.AutoMigrate(&User{}, &Role{}, &Permission{})

    service := NewUserService(db)

    // 创建权限
    permissions := []Permission{
        {Name: "用户查看", Code: "user:view", Description: "查看用户信息"},
        {Name: "用户创建", Code: "user:create", Description: "创建用户"},
        {Name: "用户编辑", Code: "user:edit", Description: "编辑用户信息"},
        {Name: "用户删除", Code: "user:delete", Description: "删除用户"},
    }
    db.Create(&permissions)

    // 创建角色
    adminRole := Role{Name: "管理员", Description: "系统管理员"}
    db.Create(&adminRole)

    // 分配权限给角色
    db.Model(&adminRole).Association("Permissions").Append(&permissions)

    // 创建用户并分配角色
    user := &User{
        Username: "admin",
        Email:    "admin@example.com",
        Password: "hashed_password",
    }

    service.CreateUserWithRoles(user, []uint{adminRole.ID})

    // 检查权限
    hasPermission, _ := service.CheckUserPermission(user.ID, "user:create")
    fmt.Printf("User has create permission: %v\n", hasPermission)
}

/*
解析说明：
1. 多对多关系：用户-角色-权限的三层关系设计
2. 事务处理：确保关联操作的原子性
3. 预加载：高效获取关联数据
4. 权限检查：通过JOIN查询实现权限验证
5. 软删除：保留数据完整性

扩展思考：
- 如何实现角色继承？
- 如何添加数据权限控制？
- 如何实现权限缓存？
- 如何处理权限变更的实时性？
*/
```

### 练习题2：电商库存管理系统（⭐⭐⭐）

**题目描述：**
设计一个电商库存管理系统，支持并发扣减库存、库存预占、库存回滚等复杂业务场景。

```go
// 练习题2：电商库存管理系统
package main

import (
    "errors"
    "fmt"
    "time"

    "gorm.io/gorm"
)

// 解答：
// 1. 模型定义
type Product struct {
    ID          uint    `gorm:"primaryKey"`
    SKU         string  `gorm:"uniqueIndex;size:50;not null"`
    Name        string  `gorm:"size:100;not null"`
    Stock       int     `gorm:"not null;default:0"`        // 实际库存
    ReservedStock int   `gorm:"not null;default:0"`        // 预占库存
    AvailableStock int `gorm:"not null;default:0"`         // 可用库存
    Version     int     `gorm:"not null;default:0"`        // 乐观锁版本号
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type StockRecord struct {
    ID          uint      `gorm:"primaryKey"`
    ProductID   uint      `gorm:"not null;index"`
    OrderID     uint      `gorm:"index"`
    Type        string    `gorm:"size:20;not null"` // in, out, reserve, release
    Quantity    int       `gorm:"not null"`
    BeforeStock int       `gorm:"not null"`
    AfterStock  int       `gorm:"not null"`
    Reason      string    `gorm:"size:100"`
    OperatorID  uint      `gorm:"not null"`
    CreatedAt   time.Time

    Product Product `gorm:"foreignKey:ProductID"`
}

type StockReservation struct {
    ID        uint      `gorm:"primaryKey"`
    ProductID uint      `gorm:"not null;index"`
    OrderID   uint      `gorm:"uniqueIndex;not null"`
    Quantity  int       `gorm:"not null"`
    Status    string    `gorm:"size:20;not null"` // reserved, confirmed, cancelled
    ExpiresAt time.Time `gorm:"not null;index"`
    CreatedAt time.Time
    UpdatedAt time.Time

    Product Product `gorm:"foreignKey:ProductID"`
}

// 2. 库存服务实现
type StockService struct {
    db *gorm.DB
}

func NewStockService(db *gorm.DB) *StockService {
    return &StockService{db: db}
}

// 入库操作
func (s *StockService) StockIn(productID uint, quantity int, reason string, operatorID uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var product Product
        if err := tx.First(&product, productID).Error; err != nil {
            return err
        }

        beforeStock := product.Stock

        // 更新库存
        result := tx.Model(&product).Updates(map[string]interface{}{
            "stock":           product.Stock + quantity,
            "available_stock": product.AvailableStock + quantity,
            "version":         product.Version + 1,
        })

        if result.Error != nil {
            return result.Error
        }

        // 记录库存变更
        record := StockRecord{
            ProductID:   productID,
            Type:        "in",
            Quantity:    quantity,
            BeforeStock: beforeStock,
            AfterStock:  beforeStock + quantity,
            Reason:      reason,
            OperatorID:  operatorID,
        }

        return tx.Create(&record).Error
    })
}

// 预占库存（下单时）
func (s *StockService) ReserveStock(productID, orderID uint, quantity int) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var product Product
        if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, productID).Error; err != nil {
            return err
        }

        // 检查可用库存
        if product.AvailableStock < quantity {
            return errors.New("insufficient available stock")
        }

        // 使用乐观锁更新
        result := tx.Model(&Product{}).
            Where("id = ? AND version = ?", productID, product.Version).
            Updates(map[string]interface{}{
                "reserved_stock":  product.ReservedStock + quantity,
                "available_stock": product.AvailableStock - quantity,
                "version":         product.Version + 1,
            })

        if result.Error != nil {
            return result.Error
        }

        if result.RowsAffected == 0 {
            return errors.New("concurrent update detected, please retry")
        }

        // 创建预占记录
        reservation := StockReservation{
            ProductID: productID,
            OrderID:   orderID,
            Quantity:  quantity,
            Status:    "reserved",
            ExpiresAt: time.Now().Add(30 * time.Minute), // 30分钟后过期
        }

        if err := tx.Create(&reservation).Error; err != nil {
            return err
        }

        // 记录库存变更
        record := StockRecord{
            ProductID:   productID,
            OrderID:     orderID,
            Type:        "reserve",
            Quantity:    quantity,
            BeforeStock: product.AvailableStock,
            AfterStock:  product.AvailableStock - quantity,
            Reason:      "订单预占",
            OperatorID:  1, // 系统操作
        }

        return tx.Create(&record).Error
    })
}

// 确认扣减库存（支付成功后）
func (s *StockService) ConfirmStock(orderID uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var reservation StockReservation
        if err := tx.Where("order_id = ? AND status = ?", orderID, "reserved").
            First(&reservation).Error; err != nil {
            return err
        }

        var product Product
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            First(&product, reservation.ProductID).Error; err != nil {
            return err
        }

        // 确认扣减：从预占库存转为实际扣减
        result := tx.Model(&Product{}).
            Where("id = ? AND version = ?", product.ID, product.Version).
            Updates(map[string]interface{}{
                "stock":          product.Stock - reservation.Quantity,
                "reserved_stock": product.ReservedStock - reservation.Quantity,
                "version":        product.Version + 1,
            })

        if result.Error != nil {
            return result.Error
        }

        if result.RowsAffected == 0 {
            return errors.New("concurrent update detected")
        }

        // 更新预占状态
        if err := tx.Model(&reservation).Update("status", "confirmed").Error; err != nil {
            return err
        }

        // 记录库存变更
        record := StockRecord{
            ProductID:   reservation.ProductID,
            OrderID:     orderID,
            Type:        "out",
            Quantity:    reservation.Quantity,
            BeforeStock: product.Stock,
            AfterStock:  product.Stock - reservation.Quantity,
            Reason:      "订单确认",
            OperatorID:  1,
        }

        return tx.Create(&record).Error
    })
}

// 释放预占库存（订单取消或超时）
func (s *StockService) ReleaseStock(orderID uint) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var reservation StockReservation
        if err := tx.Where("order_id = ? AND status = ?", orderID, "reserved").
            First(&reservation).Error; err != nil {
            return err
        }

        var product Product
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            First(&product, reservation.ProductID).Error; err != nil {
            return err
        }

        // 释放预占库存
        result := tx.Model(&Product{}).
            Where("id = ? AND version = ?", product.ID, product.Version).
            Updates(map[string]interface{}{
                "reserved_stock":  product.ReservedStock - reservation.Quantity,
                "available_stock": product.AvailableStock + reservation.Quantity,
                "version":         product.Version + 1,
            })

        if result.Error != nil {
            return result.Error
        }

        if result.RowsAffected == 0 {
            return errors.New("concurrent update detected")
        }

        // 更新预占状态
        if err := tx.Model(&reservation).Update("status", "cancelled").Error; err != nil {
            return err
        }

        // 记录库存变更
        record := StockRecord{
            ProductID:   reservation.ProductID,
            OrderID:     orderID,
            Type:        "release",
            Quantity:    reservation.Quantity,
            BeforeStock: product.AvailableStock - reservation.Quantity,
            AfterStock:  product.AvailableStock,
            Reason:      "订单取消",
            OperatorID:  1,
        }

        return tx.Create(&record).Error
    })
}

// 清理过期预占
func (s *StockService) CleanExpiredReservations() error {
    var expiredReservations []StockReservation
    if err := s.db.Where("status = ? AND expires_at < ?", "reserved", time.Now()).
        Find(&expiredReservations).Error; err != nil {
        return err
    }

    for _, reservation := range expiredReservations {
        if err := s.ReleaseStock(reservation.OrderID); err != nil {
            fmt.Printf("Failed to release expired reservation for order %d: %v\n",
                      reservation.OrderID, err)
        }
    }

    return nil
}

// 获取库存统计
func (s *StockService) GetStockStats(productID uint) (map[string]interface{}, error) {
    var product Product
    if err := s.db.First(&product, productID).Error; err != nil {
        return nil, err
    }

    var totalIn, totalOut int
    s.db.Model(&StockRecord{}).Where("product_id = ? AND type = ?", productID, "in").
        Select("COALESCE(SUM(quantity), 0)").Scan(&totalIn)

    s.db.Model(&StockRecord{}).Where("product_id = ? AND type = ?", productID, "out").
        Select("COALESCE(SUM(quantity), 0)").Scan(&totalOut)

    return map[string]interface{}{
        "product_id":       product.ID,
        "sku":             product.SKU,
        "current_stock":   product.Stock,
        "reserved_stock":  product.ReservedStock,
        "available_stock": product.AvailableStock,
        "total_in":        totalIn,
        "total_out":       totalOut,
        "version":         product.Version,
    }, nil
}

// 3. 测试函数
func TestStockManagement(db *gorm.DB) {
    // 自动迁移
    db.AutoMigrate(&Product{}, &StockRecord{}, &StockReservation{})

    service := NewStockService(db)

    // 创建商品
    product := Product{
        SKU:  "TEST001",
        Name: "测试商品",
    }
    db.Create(&product)

    // 入库
    service.StockIn(product.ID, 100, "初始入库", 1)

    // 预占库存
    service.ReserveStock(product.ID, 1001, 5)
    service.ReserveStock(product.ID, 1002, 3)

    // 确认一个订单
    service.ConfirmStock(1001)

    // 取消一个订单
    service.ReleaseStock(1002)

    // 查看库存统计
    stats, _ := service.GetStockStats(product.ID)
    fmt.Printf("Stock stats: %+v\n", stats)
}

/*
解析说明：
1. 三级库存：实际库存、预占库存、可用库存
2. 乐观锁：使用version字段防止并发冲突
3. 悲观锁：使用FOR UPDATE防止并发读取
4. 预占机制：下单时预占，支付时确认，取消时释放
5. 过期处理：自动清理过期的预占记录
6. 操作记录：完整的库存变更日志

扩展思考：
- 如何实现分布式库存管理？
- 如何处理库存超卖问题？
- 如何实现库存预警？
- 如何优化高并发场景下的性能？
*/
```

---

## 📚 章节总结

### 🎯 本章学习成果

通过本章的学习，你已经掌握了：

#### 📖 理论知识
- **GORM框架核心概念**：ORM原理、模型映射、标签系统
- **数据库设计原则**：表结构设计、索引优化、关联关系
- **事务ACID特性**：原子性、一致性、隔离性、持久性
- **性能优化理论**：查询优化、缓存策略、连接池管理

#### 🛠️ 实践技能
- **模型定义与迁移**：结构体标签、自动迁移、手动迁移
- **CRUD操作精通**：创建、查询、更新、删除的各种技巧
- **高级查询技巧**：条件查询、关联查询、子查询、聚合查询
- **事务处理能力**：自动事务、手动事务、嵌套事务、分布式事务
- **性能优化实践**：索引优化、批量操作、连接池配置、缓存应用

#### 🏗️ 架构能力
- **数据库架构设计**：表结构设计、关联关系设计、索引策略
- **业务逻辑抽象**：服务层设计、事务边界划分、错误处理
- **性能监控与优化**：慢查询监控、性能分析、瓶颈识别
- **企业级开发实践**：代码规范、最佳实践、踩坑避免

### 🆚 与其他语言框架对比总结

| 特性 | GORM (Go) | Hibernate (Java) | SQLAlchemy (Python) |
|------|-----------|------------------|---------------------|
| **学习曲线** | 简单易学 | 复杂，配置繁琐 | 中等，概念较多 |
| **性能表现** | 高性能，低延迟 | 中等，JVM优化后较好 | 中等，解释型语言限制 |
| **类型安全** | 编译时检查 | 编译时检查 | 运行时检查 |
| **并发处理** | 原生协程支持 | 线程池模型 | GIL限制 |
| **内存占用** | 低内存占用 | 较高内存占用 | 中等内存占用 |
| **部署复杂度** | 单文件部署 | 需要JVM环境 | 需要Python环境 |

### 🎯 面试准备要点

#### 核心概念掌握
- GORM的设计理念和核心特性
- ORM与原生SQL的优缺点对比
- 数据库连接池的工作原理
- 事务隔离级别的区别和应用场景

#### 实践经验展示
- 复杂业务场景的数据库设计经验
- 高并发场景下的性能优化实践
- 数据一致性保证的解决方案
- 分布式事务的处理经验

#### 问题解决能力
- 常见性能问题的诊断和解决
- 数据库死锁的预防和处理
- 大数据量场景的优化策略
- 数据迁移和版本管理经验

### 🚀 下一步学习建议

#### 深入学习方向
1. **高级特性探索**
   - 自定义数据类型
   - 插件开发
   - 钩子函数高级应用
   - 数据库驱动定制

2. **性能优化进阶**
   - 分库分表实践
   - 读写分离架构
   - 缓存策略设计
   - 监控体系建设

3. **企业级应用**
   - 微服务数据管理
   - 分布式事务处理
   - 数据一致性保证
   - 灾备和恢复策略

#### 实践项目建议
1. **个人项目**：使用GORM构建一个完整的博客系统
2. **开源贡献**：参与GORM社区，提交bug修复或功能改进
3. **企业实践**：在实际项目中应用所学知识，积累实战经验

### 💡 学习心得

GORM作为Go语言生态中最优秀的ORM框架，不仅提供了强大的功能，更重要的是它体现了Go语言简洁、高效的设计哲学。通过本章的学习，我们不仅掌握了GORM的使用技巧，更重要的是培养了数据库设计和优化的思维能力。

在实际开发中，要始终记住：
- **简单优于复杂**：优先选择简单直接的解决方案
- **性能优于功能**：在保证功能的前提下，优先考虑性能
- **可维护性优于技巧性**：代码要易于理解和维护
- **实践优于理论**：通过实际项目验证所学知识

### 🎉 恭喜完成

恭喜你完成了GORM数据库操作与实践的学习！你现在已经具备了：

✅ **扎实的理论基础** - 深入理解ORM原理和数据库设计
✅ **丰富的实践经验** - 掌握各种复杂场景的解决方案
✅ **优秀的架构能力** - 能够设计高性能、可扩展的数据层
✅ **完善的面试准备** - 具备回答各种GORM相关问题的能力

继续保持学习的热情，在Go语言的道路上不断前进！下一章我们将学习Redis缓存应用，进一步提升系统的性能和可扩展性。

---

*"数据是应用的核心，而GORM是连接Go应用与数据库的优雅桥梁。掌握了GORM，你就掌握了Go Web开发的核心技能！"* 🚀✨
```
```
```
```
```
```
```
```
