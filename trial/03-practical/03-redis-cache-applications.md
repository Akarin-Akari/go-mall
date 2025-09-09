# 实战篇第三章：Redis缓存应用与实践 🚀

> *"缓存是性能优化的银弹，而Redis是缓存界的王者！"* 💎

## 📖 章节概述

欢迎来到Go语言Redis缓存应用的实战世界！🎉 本章将带你深入掌握Redis在Go项目中的各种应用场景，从基础的键值存储到复杂的分布式锁、消息队列等高级特性。

### 🎯 学习目标

通过本章学习，你将掌握：

- **Redis基础操作** 🔧 - 连接配置、基本数据类型操作
- **缓存策略设计** 📊 - 缓存穿透、击穿、雪崩的解决方案
- **分布式锁实现** 🔒 - 基于Redis的分布式锁机制
- **消息队列应用** 📨 - 发布订阅、Stream消息队列
- **性能优化技巧** ⚡ - 连接池、管道、集群配置
- **企业级实践** 🏢 - 结合mall-go项目的真实应用场景

### 🆚 Redis vs 其他缓存方案

| 特性 | Redis (Go) | Memcached | Ehcache (Java) | Django Cache (Python) |
|------|------------|-----------|----------------|------------------------|
| **数据类型** | 丰富(String/Hash/List/Set/ZSet) | 仅String | 对象缓存 | 多种后端支持 |
| **持久化** | RDB + AOF双重保障 | 无持久化 | 可选持久化 | 依赖后端 |
| **分布式** | 原生集群支持 | 客户端分片 | 分布式缓存 | 分布式支持 |
| **性能** | 极高性能，单线程模型 | 高性能多线程 | JVM内存缓存 | 中等性能 |
| **功能丰富度** | 发布订阅、事务、Lua脚本 | 基础缓存 | 丰富的缓存策略 | 基础缓存功能 |
| **运维复杂度** | 中等，配置灵活 | 简单 | 复杂，JVM调优 | 简单 |

### 🏗️ 技术栈对比

#### Go + Redis 的优势
```go
// Go的并发优势 + Redis的高性能 = 完美组合
func main() {
    // 1. 轻量级协程处理大量并发
    for i := 0; i < 10000; i++ {
        go handleRequest(i) // 每个请求一个协程
    }
    
    // 2. Redis连接池复用
    rdb := redis.NewClient(&redis.Options{
        PoolSize: 100, // 连接池大小
    })
    
    // 3. 类型安全的操作
    val, err := rdb.Get(ctx, "key").Result()
    if err == redis.Nil {
        // 键不存在的优雅处理
    }
}
```

#### Java + Redis 对比
```java
// Java版本 - 更重量级但功能丰富
@Service
public class RedisService {
    @Autowired
    private RedisTemplate<String, Object> redisTemplate;
    
    public void setCache(String key, Object value) {
        redisTemplate.opsForValue().set(key, value, Duration.ofHours(1));
    }
}
```

#### Python + Redis 对比
```python
# Python版本 - 简洁但性能有限
import redis

r = redis.Redis(host='localhost', port=6379, db=0)
r.set('key', 'value', ex=3600)  # 设置1小时过期
```

---

## 🔧 Redis基础配置与连接

### Go-Redis客户端介绍

go-redis是Go语言中最流行的Redis客户端库，具有以下特点：

- **类型安全** ✅ - 编译时类型检查，避免运行时错误
- **高性能** ⚡ - 连接池管理，支持管道操作
- **功能完整** 🎯 - 支持Redis所有命令和数据类型
- **易于使用** 😊 - 简洁的API设计，链式调用
- **生产就绪** 🏭 - 支持集群、哨兵、分片等企业级特性

### 基础连接配置

```go
// 来自 mall-go/internal/config/redis.go
package config

import (
    "context"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// Redis配置结构
type RedisConfig struct {
    Host         string `yaml:"host" json:"host"`
    Port         int    `yaml:"port" json:"port"`
    Password     string `yaml:"password" json:"password"`
    DB           int    `yaml:"db" json:"db"`
    PoolSize     int    `yaml:"pool_size" json:"pool_size"`
    MinIdleConns int    `yaml:"min_idle_conns" json:"min_idle_conns"`
    MaxRetries   int    `yaml:"max_retries" json:"max_retries"`
    DialTimeout  int    `yaml:"dial_timeout" json:"dial_timeout"`
    ReadTimeout  int    `yaml:"read_timeout" json:"read_timeout"`
    WriteTimeout int    `yaml:"write_timeout" json:"write_timeout"`
    PoolTimeout  int    `yaml:"pool_timeout" json:"pool_timeout"`
}

// 创建Redis客户端
func NewRedisClient(cfg *RedisConfig) *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password:     cfg.Password,
        DB:           cfg.DB,
        PoolSize:     cfg.PoolSize,     // 连接池大小
        MinIdleConns: cfg.MinIdleConns, // 最小空闲连接数
        MaxRetries:   cfg.MaxRetries,   // 最大重试次数
        DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
        ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
        PoolTimeout:  time.Duration(cfg.PoolTimeout) * time.Second,
    })
    
    return rdb
}

// 测试连接
func TestRedisConnection(rdb *redis.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Ping测试
    pong, err := rdb.Ping(ctx).Result()
    if err != nil {
        return fmt.Errorf("redis connection failed: %w", err)
    }
    
    fmt.Printf("Redis connected successfully: %s\n", pong)
    return nil
}

// 获取连接池状态
func GetPoolStats(rdb *redis.Client) *redis.PoolStats {
    return rdb.PoolStats()
}

// 打印连接池统计信息
func PrintPoolStats(rdb *redis.Client) {
    stats := rdb.PoolStats()
    fmt.Printf("Redis Pool Stats:\n")
    fmt.Printf("  Hits: %d\n", stats.Hits)
    fmt.Printf("  Misses: %d\n", stats.Misses)
    fmt.Printf("  Timeouts: %d\n", stats.Timeouts)
    fmt.Printf("  TotalConns: %d\n", stats.TotalConns)
    fmt.Printf("  IdleConns: %d\n", stats.IdleConns)
    fmt.Printf("  StaleConns: %d\n", stats.StaleConns)
}
```

### 高级连接配置

```go
// 来自 mall-go/internal/config/redis_advanced.go
package config

import (
    "context"
    "crypto/tls"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// 集群配置
func NewRedisClusterClient(addrs []string, password string) *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs:    addrs,
        Password: password,
        
        // 集群特定配置
        MaxRedirects:   8,  // 最大重定向次数
        ReadOnly:       false, // 是否只读
        RouteByLatency: true,  // 按延迟路由
        RouteRandomly:  false, // 随机路由
        
        // 连接池配置
        PoolSize:     100,
        MinIdleConns: 10,
        MaxRetries:   3,
        
        // 超时配置
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolTimeout:  4 * time.Second,
    })
}

// 哨兵配置
func NewRedisSentinelClient(masterName string, sentinelAddrs []string, password string) *redis.Client {
    return redis.NewFailoverClient(&redis.FailoverOptions{
        MasterName:    masterName,
        SentinelAddrs: sentinelAddrs,
        Password:      password,
        DB:            0,
        
        // 哨兵特定配置
        SentinelPassword: "", // 哨兵密码
        MaxRetryBackoff:  512 * time.Millisecond,
        
        // 连接池配置
        PoolSize:     50,
        MinIdleConns: 5,
        MaxRetries:   3,
    })
}

// TLS安全连接
func NewRedisClientWithTLS(addr, password string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       0,
        
        // TLS配置
        TLSConfig: &tls.Config{
            ServerName: "redis.example.com",
            MinVersion: tls.VersionTLS12,
        },
        
        // 连接池配置
        PoolSize:     20,
        MinIdleConns: 2,
        MaxRetries:   3,
    })
}

// 动态凭证提供者
func NewRedisClientWithCredentialsProvider(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr: addr,
        
        // 动态凭证提供者
        CredentialsProvider: func() (username, password string) {
            // 从配置中心、环境变量或其他安全存储获取凭证
            return getCredentialsFromVault()
        },
        
        // 或者使用上下文相关的凭证提供者
        CredentialsProviderContext: func(ctx context.Context) (username, password string, err error) {
            // 根据上下文获取凭证
            return getCredentialsFromContext(ctx)
        },
    })
}

// 辅助函数
func getCredentialsFromVault() (string, string) {
    // 模拟从密钥管理系统获取凭证
    return "redis_user", "secure_password"
}

func getCredentialsFromContext(ctx context.Context) (string, string, error) {
    // 从上下文中获取用户信息，返回对应的Redis凭证
    userID := ctx.Value("user_id")
    if userID == nil {
        return "", "", fmt.Errorf("user not authenticated")
    }
    
    return fmt.Sprintf("user_%v", userID), "dynamic_password", nil
}
```

---

## 📊 Redis数据类型操作

Redis支持多种数据类型，每种类型都有其特定的使用场景。让我们深入了解各种数据类型的操作。

### String类型操作

```go
// 来自 mall-go/internal/service/redis_string_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisStringService struct {
    rdb *redis.Client
}

func NewRedisStringService(rdb *redis.Client) *RedisStringService {
    return &RedisStringService{rdb: rdb}
}

// 1. 基础字符串操作
func (s *RedisStringService) BasicStringOperations(ctx context.Context) {
    // 设置键值
    err := s.rdb.Set(ctx, "user:1:name", "张三", time.Hour).Err()
    if err != nil {
        panic(err)
    }

    // 获取值
    val, err := s.rdb.Get(ctx, "user:1:name").Result()
    if err == redis.Nil {
        fmt.Println("键不存在")
    } else if err != nil {
        panic(err)
    } else {
        fmt.Printf("用户名: %s\n", val)
    }

    // 设置过期时间
    s.rdb.Expire(ctx, "user:1:name", 30*time.Minute)

    // 获取剩余过期时间
    ttl, err := s.rdb.TTL(ctx, "user:1:name").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("剩余过期时间: %v\n", ttl)
}

// 2. 原子操作
func (s *RedisStringService) AtomicOperations(ctx context.Context) {
    // 原子递增
    count, err := s.rdb.Incr(ctx, "page:views").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("页面访问次数: %d\n", count)

    // 原子递增指定值
    s.rdb.IncrBy(ctx, "user:1:score", 10)

    // 浮点数递增
    s.rdb.IncrByFloat(ctx, "user:1:balance", 99.99)

    // 原子递减
    s.rdb.Decr(ctx, "inventory:product:1")
    s.rdb.DecrBy(ctx, "inventory:product:1", 5)
}

// 3. 条件设置
func (s *RedisStringService) ConditionalSet(ctx context.Context) {
    // 仅当键不存在时设置
    success, err := s.rdb.SetNX(ctx, "lock:order:1", "processing", 10*time.Minute).Result()
    if err != nil {
        panic(err)
    }

    if success {
        fmt.Println("获取锁成功")
        // 执行业务逻辑
        defer s.rdb.Del(ctx, "lock:order:1") // 释放锁
    } else {
        fmt.Println("获取锁失败，订单正在处理中")
    }

    // 仅当键存在时设置
    s.rdb.SetXX(ctx, "user:1:status", "active", time.Hour)
}

// 4. 批量操作
func (s *RedisStringService) BatchOperations(ctx context.Context) {
    // 批量设置
    err := s.rdb.MSet(ctx,
        "user:1:name", "张三",
        "user:1:age", "25",
        "user:1:city", "北京",
    ).Err()
    if err != nil {
        panic(err)
    }

    // 批量获取
    vals, err := s.rdb.MGet(ctx, "user:1:name", "user:1:age", "user:1:city").Result()
    if err != nil {
        panic(err)
    }

    for i, val := range vals {
        if val != nil {
            fmt.Printf("值 %d: %s\n", i, val)
        }
    }

    // 批量设置（仅当所有键都不存在时）
    success, err := s.rdb.MSetNX(ctx,
        "config:app:version", "1.0.0",
        "config:app:env", "production",
    ).Result()

    if success {
        fmt.Println("配置初始化成功")
    }
}

// 5. JSON对象存储
func (s *RedisStringService) JSONOperations(ctx context.Context) {
    // 用户信息结构
    type User struct {
        ID       int    `json:"id"`
        Name     string `json:"name"`
        Email    string `json:"email"`
        Age      int    `json:"age"`
        IsActive bool   `json:"is_active"`
    }

    user := User{
        ID:       1,
        Name:     "张三",
        Email:    "zhangsan@example.com",
        Age:      25,
        IsActive: true,
    }

    // 序列化并存储
    userJSON, err := json.Marshal(user)
    if err != nil {
        panic(err)
    }

    err = s.rdb.Set(ctx, "user:1:profile", userJSON, time.Hour).Err()
    if err != nil {
        panic(err)
    }

    // 获取并反序列化
    userJSONStr, err := s.rdb.Get(ctx, "user:1:profile").Result()
    if err != nil {
        panic(err)
    }

    var retrievedUser User
    err = json.Unmarshal([]byte(userJSONStr), &retrievedUser)
    if err != nil {
        panic(err)
    }

    fmt.Printf("用户信息: %+v\n", retrievedUser)
}

// 6. 位操作
func (s *RedisStringService) BitOperations(ctx context.Context) {
    // 设置位
    s.rdb.SetBit(ctx, "user:online:2024-01-01", 1001, 1) // 用户1001在线
    s.rdb.SetBit(ctx, "user:online:2024-01-01", 1002, 1) // 用户1002在线

    // 获取位
    bit, err := s.rdb.GetBit(ctx, "user:online:2024-01-01", 1001).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户1001是否在线: %d\n", bit)

    // 统计位数
    count, err := s.rdb.BitCount(ctx, "user:online:2024-01-01", nil).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("今日在线用户数: %d\n", count)

    // 位运算
    s.rdb.BitOpAnd(ctx, "user:active:both", "user:online:2024-01-01", "user:online:2024-01-02")
}
```

### Hash类型操作

```go
// Hash类型操作
func (s *RedisStringService) HashOperations(ctx context.Context) {
    // 设置Hash字段
    err := s.rdb.HSet(ctx, "user:1", map[string]interface{}{
        "name":     "张三",
        "age":      25,
        "city":     "北京",
        "is_vip":   true,
        "balance":  1000.50,
    }).Err()
    if err != nil {
        panic(err)
    }

    // 获取单个字段
    name, err := s.rdb.HGet(ctx, "user:1", "name").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户姓名: %s\n", name)

    // 获取多个字段
    vals, err := s.rdb.HMGet(ctx, "user:1", "name", "age", "city").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户信息: %v\n", vals)

    // 获取所有字段
    userMap, err := s.rdb.HGetAll(ctx, "user:1").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("完整用户信息: %v\n", userMap)

    // 原子递增Hash字段
    newBalance, err := s.rdb.HIncrByFloat(ctx, "user:1", "balance", 99.50).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("新余额: %.2f\n", newBalance)

    // 仅当字段不存在时设置
    success, err := s.rdb.HSetNX(ctx, "user:1", "created_at", time.Now().Unix()).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("创建时间设置成功: %v\n", success)

    // 删除Hash字段
    s.rdb.HDel(ctx, "user:1", "temp_field")

    // 检查字段是否存在
    exists, err := s.rdb.HExists(ctx, "user:1", "name").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("name字段存在: %v\n", exists)

    // 获取Hash长度
    length, err := s.rdb.HLen(ctx, "user:1").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Hash字段数量: %d\n", length)

    // 获取所有字段名
    keys, err := s.rdb.HKeys(ctx, "user:1").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("所有字段名: %v\n", keys)

    // 获取所有字段值
    values, err := s.rdb.HVals(ctx, "user:1").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("所有字段值: %v\n", values)
}
```

### List类型操作

```go
// List类型操作 - 适用于消息队列、最新动态等场景
func (s *RedisStringService) ListOperations(ctx context.Context) {
    // 左侧推入元素
    s.rdb.LPush(ctx, "user:1:notifications",
        "您有新的订单",
        "您的订单已发货",
        "系统维护通知",
    )

    // 右侧推入元素
    s.rdb.RPush(ctx, "user:1:browse_history",
        "商品A",
        "商品B",
        "商品C",
    )

    // 获取列表长度
    length, err := s.rdb.LLen(ctx, "user:1:notifications").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("通知数量: %d\n", length)

    // 获取指定范围的元素
    notifications, err := s.rdb.LRange(ctx, "user:1:notifications", 0, 2).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("最新3条通知: %v\n", notifications)

    // 获取指定索引的元素
    firstNotification, err := s.rdb.LIndex(ctx, "user:1:notifications", 0).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("最新通知: %s\n", firstNotification)

    // 左侧弹出元素
    notification, err := s.rdb.LPop(ctx, "user:1:notifications").Result()
    if err != nil && err != redis.Nil {
        panic(err)
    }
    if err != redis.Nil {
        fmt.Printf("处理通知: %s\n", notification)
    }

    // 右侧弹出元素
    lastItem, err := s.rdb.RPop(ctx, "user:1:browse_history").Result()
    if err != nil && err != redis.Nil {
        panic(err)
    }
    if err != redis.Nil {
        fmt.Printf("最后浏览: %s\n", lastItem)
    }

    // 阻塞式弹出（用于消息队列）
    go func() {
        result, err := s.rdb.BLPop(ctx, 30*time.Second, "task:queue").Result()
        if err != nil {
            fmt.Printf("队列超时或错误: %v\n", err)
            return
        }
        fmt.Printf("处理任务: %s\n", result[1]) // result[0]是键名，result[1]是值
    }()

    // 修剪列表（保留指定范围）
    s.rdb.LTrim(ctx, "user:1:browse_history", 0, 9) // 只保留最新10条

    // 在指定元素前/后插入
    s.rdb.LInsertBefore(ctx, "user:1:notifications", "系统维护通知", "紧急通知")
    s.rdb.LInsertAfter(ctx, "user:1:notifications", "系统维护通知", "维护完成通知")

    // 设置指定索引的值
    s.rdb.LSet(ctx, "user:1:notifications", 0, "已读通知")

    // 删除指定值的元素
    s.rdb.LRem(ctx, "user:1:notifications", 1, "已读通知") // 删除1个"已读通知"
}
```

### Set类型操作

```go
// Set类型操作 - 适用于标签、关注关系等场景
func (s *RedisStringService) SetOperations(ctx context.Context) {
    // 添加成员
    s.rdb.SAdd(ctx, "user:1:tags", "技术", "编程", "Go语言", "Redis")
    s.rdb.SAdd(ctx, "user:2:tags", "技术", "Java", "Spring", "MySQL")

    // 获取所有成员
    tags, err := s.rdb.SMembers(ctx, "user:1:tags").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户1的标签: %v\n", tags)

    // 检查成员是否存在
    exists, err := s.rdb.SIsMember(ctx, "user:1:tags", "Go语言").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户1是否有Go语言标签: %v\n", exists)

    // 获取集合大小
    size, err := s.rdb.SCard(ctx, "user:1:tags").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户1标签数量: %d\n", size)

    // 随机获取成员
    randomTag, err := s.rdb.SRandMember(ctx, "user:1:tags").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("随机标签: %s\n", randomTag)

    // 随机获取多个成员
    randomTags, err := s.rdb.SRandMemberN(ctx, "user:1:tags", 2).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("随机2个标签: %v\n", randomTags)

    // 弹出成员
    poppedTag, err := s.rdb.SPop(ctx, "user:1:tags").Result()
    if err != nil && err != redis.Nil {
        panic(err)
    }
    if err != redis.Nil {
        fmt.Printf("弹出的标签: %s\n", poppedTag)
    }

    // 移除指定成员
    s.rdb.SRem(ctx, "user:1:tags", "过时标签")

    // 集合运算
    // 交集 - 共同标签
    commonTags, err := s.rdb.SInter(ctx, "user:1:tags", "user:2:tags").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("共同标签: %v\n", commonTags)

    // 并集 - 所有标签
    allTags, err := s.rdb.SUnion(ctx, "user:1:tags", "user:2:tags").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("所有标签: %v\n", allTags)

    // 差集 - 用户1独有的标签
    uniqueTags, err := s.rdb.SDiff(ctx, "user:1:tags", "user:2:tags").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("用户1独有标签: %v\n", uniqueTags)

    // 将交集结果存储到新集合
    s.rdb.SInterStore(ctx, "common:tags", "user:1:tags", "user:2:tags")

    // 将并集结果存储到新集合
    s.rdb.SUnionStore(ctx, "all:tags", "user:1:tags", "user:2:tags")

    // 将差集结果存储到新集合
    s.rdb.SDiffStore(ctx, "user:1:unique:tags", "user:1:tags", "user:2:tags")
}
```

### ZSet类型操作

```go
// ZSet类型操作 - 适用于排行榜、优先级队列等场景
func (s *RedisStringService) ZSetOperations(ctx context.Context) {
    // 添加成员和分数
    s.rdb.ZAdd(ctx, "game:leaderboard",
        redis.Z{Score: 1000, Member: "player1"},
        redis.Z{Score: 1500, Member: "player2"},
        redis.Z{Score: 800, Member: "player3"},
        redis.Z{Score: 2000, Member: "player4"},
        redis.Z{Score: 1200, Member: "player5"},
    )

    // 获取指定范围的成员（按分数升序）
    players, err := s.rdb.ZRange(ctx, "game:leaderboard", 0, 2).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("分数最低的3名玩家: %v\n", players)

    // 获取指定范围的成员（按分数降序）
    topPlayers, err := s.rdb.ZRevRange(ctx, "game:leaderboard", 0, 2).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("排行榜前3名: %v\n", topPlayers)

    // 获取成员和分数
    playersWithScores, err := s.rdb.ZRevRangeWithScores(ctx, "game:leaderboard", 0, 2).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("前3名详细信息:\n")
    for i, player := range playersWithScores {
        fmt.Printf("  第%d名: %s (分数: %.0f)\n", i+1, player.Member, player.Score)
    }

    // 获取成员的分数
    score, err := s.rdb.ZScore(ctx, "game:leaderboard", "player2").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("player2的分数: %.0f\n", score)

    // 获取成员的排名（从0开始，分数升序）
    rank, err := s.rdb.ZRank(ctx, "game:leaderboard", "player2").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("player2的排名（升序）: %d\n", rank)

    // 获取成员的排名（从0开始，分数降序）
    revRank, err := s.rdb.ZRevRank(ctx, "game:leaderboard", "player2").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("player2的排名（降序）: %d\n", revRank)

    // 增加成员的分数
    newScore, err := s.rdb.ZIncrBy(ctx, "game:leaderboard", 100, "player3").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("player3新分数: %.0f\n", newScore)

    // 获取集合大小
    count, err := s.rdb.ZCard(ctx, "game:leaderboard").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("排行榜玩家数量: %d\n", count)

    // 获取指定分数范围的成员数量
    countInRange, err := s.rdb.ZCount(ctx, "game:leaderboard", "1000", "1500").Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("分数在1000-1500之间的玩家数量: %d\n", countInRange)

    // 按分数范围获取成员
    playersInRange, err := s.rdb.ZRangeByScore(ctx, "game:leaderboard", &redis.ZRangeBy{
        Min: "1000",
        Max: "1500",
    }).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("分数在1000-1500之间的玩家: %v\n", playersInRange)

    // 按分数范围获取成员和分数（带限制）
    playersWithLimit, err := s.rdb.ZRangeByScoreWithScores(ctx, "game:leaderboard", &redis.ZRangeBy{
        Min:    "1000",
        Max:    "+inf",
        Offset: 0,
        Count:  3,
    }).Result()
    if err != nil {
        panic(err)
    }
    fmt.Printf("分数>=1000的前3名玩家:\n")
    for _, player := range playersWithLimit {
        fmt.Printf("  %s: %.0f\n", player.Member, player.Score)
    }

    // 移除成员
    s.rdb.ZRem(ctx, "game:leaderboard", "inactive_player")

    // 按排名范围移除成员
    s.rdb.ZRemRangeByRank(ctx, "game:leaderboard", 0, 0) // 移除排名最低的1个

    // 按分数范围移除成员
    s.rdb.ZRemRangeByScore(ctx, "game:leaderboard", "-inf", "500") // 移除分数<=500的成员
}
```

---

## 🎯 缓存策略设计

缓存策略是Redis应用的核心，合理的缓存策略能够显著提升系统性能。让我们深入了解各种缓存模式和最佳实践。

### 缓存模式对比

| 缓存模式 | 适用场景 | 优点 | 缺点 | 一致性 |
|----------|----------|------|------|--------|
| **Cache-Aside** | 读多写少 | 简单可控 | 代码复杂 | 最终一致性 |
| **Read-Through** | 读密集型 | 代码简洁 | 缓存依赖强 | 强一致性 |
| **Write-Through** | 写一致性要求高 | 数据一致 | 写性能差 | 强一致性 |
| **Write-Behind** | 写密集型 | 写性能好 | 数据丢失风险 | 最终一致性 |

### Cache-Aside模式实现

```go
// 来自 mall-go/internal/service/cache_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
    "mall-go/internal/model"
)

type CacheService struct {
    rdb *redis.Client
    db  *gorm.DB
}

func NewCacheService(rdb *redis.Client, db *gorm.DB) *CacheService {
    return &CacheService{
        rdb: rdb,
        db:  db,
    }
}

// Cache-Aside模式 - 用户信息缓存
func (s *CacheService) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)

    // 1. 先从缓存获取
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中
        var user model.User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            fmt.Printf("缓存命中: user:%d\n", userID)
            return &user, nil
        }
    }

    // 2. 缓存未命中，从数据库获取
    var user model.User
    if err := s.db.First(&user, userID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            // 防止缓存穿透：缓存空值
            s.rdb.Set(ctx, cacheKey, "null", 5*time.Minute)
        }
        return nil, err
    }

    // 3. 写入缓存
    userJSON, err := json.Marshal(user)
    if err == nil {
        s.rdb.Set(ctx, cacheKey, userJSON, time.Hour)
        fmt.Printf("缓存更新: user:%d\n", userID)
    }

    return &user, nil
}

// 更新用户信息
func (s *CacheService) UpdateUser(ctx context.Context, userID uint, updates map[string]interface{}) error {
    cacheKey := fmt.Sprintf("user:%d", userID)

    // 1. 更新数据库
    if err := s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
        return err
    }

    // 2. 删除缓存（让下次读取时重新加载）
    s.rdb.Del(ctx, cacheKey)
    fmt.Printf("缓存删除: user:%d\n", userID)

    return nil
}

// 批量获取用户信息（解决N+1问题）
func (s *CacheService) GetUsersByIDs(ctx context.Context, userIDs []uint) (map[uint]*model.User, error) {
    if len(userIDs) == 0 {
        return make(map[uint]*model.User), nil
    }

    // 1. 构建缓存键
    cacheKeys := make([]string, len(userIDs))
    keyToID := make(map[string]uint)
    for i, id := range userIDs {
        key := fmt.Sprintf("user:%d", id)
        cacheKeys[i] = key
        keyToID[key] = id
    }

    // 2. 批量获取缓存
    cached, err := s.rdb.MGet(ctx, cacheKeys...).Result()
    if err != nil {
        return nil, err
    }

    result := make(map[uint]*model.User)
    var missedIDs []uint

    // 3. 处理缓存结果
    for i, val := range cached {
        userID := keyToID[cacheKeys[i]]

        if val == nil || val == "null" {
            missedIDs = append(missedIDs, userID)
            continue
        }

        var user model.User
        if err := json.Unmarshal([]byte(val.(string)), &user); err == nil {
            result[userID] = &user
        } else {
            missedIDs = append(missedIDs, userID)
        }
    }

    // 4. 从数据库获取缓存未命中的数据
    if len(missedIDs) > 0 {
        var users []model.User
        if err := s.db.Where("id IN ?", missedIDs).Find(&users).Error; err != nil {
            return nil, err
        }

        // 5. 更新缓存并添加到结果
        for _, user := range users {
            result[user.ID] = &user

            // 异步更新缓存
            go func(u model.User) {
                userJSON, err := json.Marshal(u)
                if err == nil {
                    cacheKey := fmt.Sprintf("user:%d", u.ID)
                    s.rdb.Set(context.Background(), cacheKey, userJSON, time.Hour)
                }
            }(user)
        }

        // 6. 对于不存在的用户，缓存空值防止穿透
        existingIDs := make(map[uint]bool)
        for _, user := range users {
            existingIDs[user.ID] = true
        }

        for _, id := range missedIDs {
            if !existingIDs[id] {
                cacheKey := fmt.Sprintf("user:%d", id)
                s.rdb.Set(ctx, cacheKey, "null", 5*time.Minute)
            }
        }
    }

    fmt.Printf("批量获取用户: 总数%d, 缓存命中%d, 数据库查询%d\n",
               len(userIDs), len(result)-len(missedIDs), len(missedIDs))

    return result, nil
}
```

### 缓存问题解决方案

```go
// 缓存问题解决方案
type CacheProblemSolver struct {
    rdb *redis.Client
    db  *gorm.DB
}

func NewCacheProblemSolver(rdb *redis.Client, db *gorm.DB) *CacheProblemSolver {
    return &CacheProblemSolver{rdb: rdb, db: db}
}

// 1. 缓存穿透解决方案
func (s *CacheProblemSolver) GetProductWithBloomFilter(ctx context.Context, productID uint) (*model.Product, error) {
    cacheKey := fmt.Sprintf("product:%d", productID)
    bloomKey := "product:bloom"

    // 1. 检查布隆过滤器
    exists, err := s.rdb.BFExists(ctx, bloomKey, fmt.Sprintf("%d", productID)).Result()
    if err != nil {
        // 如果布隆过滤器不可用，降级到普通查询
        return s.getProductNormal(ctx, productID)
    }

    if !exists {
        // 布隆过滤器表示不存在，直接返回
        fmt.Printf("布隆过滤器拦截: product:%d\n", productID)
        return nil, gorm.ErrRecordNotFound
    }

    // 2. 检查缓存
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil && cached != "null" {
        var product model.Product
        if err := json.Unmarshal([]byte(cached), &product); err == nil {
            return &product, nil
        }
    }

    // 3. 查询数据库
    var product model.Product
    if err := s.db.First(&product, productID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            // 缓存空值，防止频繁查询
            s.rdb.Set(ctx, cacheKey, "null", 5*time.Minute)
        }
        return nil, err
    }

    // 4. 更新缓存
    productJSON, _ := json.Marshal(product)
    s.rdb.Set(ctx, cacheKey, productJSON, time.Hour)

    return &product, nil
}

// 2. 缓存击穿解决方案（分布式锁）
func (s *CacheProblemSolver) GetHotProductWithLock(ctx context.Context, productID uint) (*model.Product, error) {
    cacheKey := fmt.Sprintf("product:%d", productID)
    lockKey := fmt.Sprintf("lock:product:%d", productID)

    // 1. 检查缓存
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil && cached != "null" {
        var product model.Product
        if err := json.Unmarshal([]byte(cached), &product); err == nil {
            return &product, nil
        }
    }

    // 2. 获取分布式锁
    lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
    locked, err := s.rdb.SetNX(ctx, lockKey, lockValue, 10*time.Second).Result()
    if err != nil {
        return nil, err
    }

    if !locked {
        // 未获取到锁，等待一段时间后重试
        time.Sleep(50 * time.Millisecond)
        return s.GetHotProductWithLock(ctx, productID) // 递归重试
    }

    // 3. 获取锁成功，再次检查缓存（双重检查）
    defer func() {
        // 释放锁（使用Lua脚本保证原子性）
        luaScript := `
            if redis.call("get", KEYS[1]) == ARGV[1] then
                return redis.call("del", KEYS[1])
            else
                return 0
            end
        `
        s.rdb.Eval(ctx, luaScript, []string{lockKey}, lockValue)
    }()

    cached, err = s.rdb.Get(ctx, cacheKey).Result()
    if err == nil && cached != "null" {
        var product model.Product
        if err := json.Unmarshal([]byte(cached), &product); err == nil {
            return &product, nil
        }
    }

    // 4. 查询数据库
    var product model.Product
    if err := s.db.First(&product, productID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            s.rdb.Set(ctx, cacheKey, "null", 5*time.Minute)
        }
        return nil, err
    }

    // 5. 更新缓存
    productJSON, _ := json.Marshal(product)
    s.rdb.Set(ctx, cacheKey, productJSON, time.Hour)

    return &product, nil
}

// 3. 缓存雪崩解决方案（随机过期时间）
func (s *CacheProblemSolver) SetCacheWithRandomExpire(ctx context.Context, key string, value interface{}, baseExpire time.Duration) error {
    // 添加随机时间，避免同时过期
    randomExpire := baseExpire + time.Duration(rand.Intn(300))*time.Second // 0-5分钟随机

    valueJSON, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return s.rdb.Set(ctx, key, valueJSON, randomExpire).Err()
}

// 4. 缓存预热
func (s *CacheProblemSolver) WarmUpCache(ctx context.Context) error {
    // 预热热门商品
    var hotProducts []model.Product
    if err := s.db.Where("is_hot = ?", true).Limit(100).Find(&hotProducts).Error; err != nil {
        return err
    }

    // 批量设置缓存
    pipe := s.rdb.Pipeline()
    for _, product := range hotProducts {
        cacheKey := fmt.Sprintf("product:%d", product.ID)
        productJSON, _ := json.Marshal(product)

        // 使用随机过期时间
        expire := time.Hour + time.Duration(rand.Intn(1800))*time.Second
        pipe.Set(ctx, cacheKey, productJSON, expire)
    }

    _, err := pipe.Exec(ctx)
    if err != nil {
        return err
    }

    fmt.Printf("缓存预热完成: %d个热门商品\n", len(hotProducts))
    return nil
}

// 辅助方法
func (s *CacheProblemSolver) getProductNormal(ctx context.Context, productID uint) (*model.Product, error) {
    var product model.Product
    if err := s.db.First(&product, productID).Error; err != nil {
        return nil, err
    }
    return &product, nil
}
```

---

## 🔒 分布式锁实现

分布式锁是分布式系统中保证数据一致性的重要机制。Redis提供了多种实现分布式锁的方案。

### 基础分布式锁实现

```go
// 来自 mall-go/internal/service/distributed_lock_service.go
package service

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

var (
    ErrLockNotAcquired = errors.New("lock not acquired")
    ErrLockNotHeld     = errors.New("lock not held")
)

// 分布式锁结构
type DistributedLock struct {
    rdb        *redis.Client
    key        string
    value      string
    expiration time.Duration
}

// 创建分布式锁
func NewDistributedLock(rdb *redis.Client, key string, expiration time.Duration) *DistributedLock {
    return &DistributedLock{
        rdb:        rdb,
        key:        key,
        value:      generateLockValue(),
        expiration: expiration,
    }
}

// 生成锁的唯一值
func generateLockValue() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

// 获取锁
func (l *DistributedLock) Acquire(ctx context.Context) error {
    // 使用SET命令的NX和EX选项实现原子操作
    success, err := l.rdb.SetNX(ctx, l.key, l.value, l.expiration).Result()
    if err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }

    if !success {
        return ErrLockNotAcquired
    }

    fmt.Printf("锁获取成功: %s\n", l.key)
    return nil
}

// 释放锁（使用Lua脚本保证原子性）
func (l *DistributedLock) Release(ctx context.Context) error {
    // Lua脚本：只有持有锁的客户端才能释放锁
    luaScript := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `

    result, err := l.rdb.Eval(ctx, luaScript, []string{l.key}, l.value).Result()
    if err != nil {
        return fmt.Errorf("failed to release lock: %w", err)
    }

    if result.(int64) == 0 {
        return ErrLockNotHeld
    }

    fmt.Printf("锁释放成功: %s\n", l.key)
    return nil
}

// 续期锁（防止业务执行时间过长导致锁过期）
func (l *DistributedLock) Renew(ctx context.Context) error {
    luaScript := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("expire", KEYS[1], ARGV[2])
        else
            return 0
        end
    `

    result, err := l.rdb.Eval(ctx, luaScript, []string{l.key}, l.value, int(l.expiration.Seconds())).Result()
    if err != nil {
        return fmt.Errorf("failed to renew lock: %w", err)
    }

    if result.(int64) == 0 {
        return ErrLockNotHeld
    }

    fmt.Printf("锁续期成功: %s\n", l.key)
    return nil
}

// 尝试获取锁（带重试）
func (l *DistributedLock) TryAcquire(ctx context.Context, retryInterval time.Duration, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := l.Acquire(ctx)
        if err == nil {
            return nil
        }

        if err != ErrLockNotAcquired {
            return err
        }

        if i < maxRetries-1 {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(retryInterval):
                // 继续重试
            }
        }
    }

    return ErrLockNotAcquired
}

// 自动续期锁
func (l *DistributedLock) AcquireWithAutoRenew(ctx context.Context) (func(), error) {
    // 获取锁
    if err := l.Acquire(ctx); err != nil {
        return nil, err
    }

    // 创建取消上下文
    renewCtx, cancel := context.WithCancel(ctx)

    // 启动自动续期协程
    go func() {
        ticker := time.NewTicker(l.expiration / 3) // 每1/3过期时间续期一次
        defer ticker.Stop()

        for {
            select {
            case <-renewCtx.Done():
                return
            case <-ticker.C:
                if err := l.Renew(renewCtx); err != nil {
                    fmt.Printf("自动续期失败: %v\n", err)
                    return
                }
            }
        }
    }()

    // 返回释放函数
    releaseFunc := func() {
        cancel() // 停止自动续期
        if err := l.Release(context.Background()); err != nil {
            fmt.Printf("释放锁失败: %v\n", err)
        }
    }

    return releaseFunc, nil
}
```

### 高级分布式锁应用

```go
// 分布式锁服务
type LockService struct {
    rdb *redis.Client
}

func NewLockService(rdb *redis.Client) *LockService {
    return &LockService{rdb: rdb}
}

// 1. 订单处理锁（防止重复下单）
func (s *LockService) ProcessOrderWithLock(ctx context.Context, userID uint, orderData *CreateOrderRequest) (*model.Order, error) {
    lockKey := fmt.Sprintf("lock:order:user:%d", userID)
    lock := NewDistributedLock(s.rdb, lockKey, 30*time.Second)

    // 获取锁
    if err := lock.TryAcquire(ctx, 100*time.Millisecond, 10); err != nil {
        return nil, fmt.Errorf("获取订单锁失败: %w", err)
    }
    defer lock.Release(ctx)

    // 检查是否有重复订单
    var existingOrder model.Order
    err := s.db.Where("user_id = ? AND status = ? AND created_at > ?",
                      userID, "pending", time.Now().Add(-5*time.Minute)).
              First(&existingOrder).Error

    if err == nil {
        return nil, errors.New("5分钟内已有待处理订单，请勿重复下单")
    }

    // 创建订单
    order := &model.Order{
        UserID:      userID,
        TotalAmount: orderData.TotalAmount,
        Status:      "pending",
    }

    if err := s.db.Create(order).Error; err != nil {
        return nil, err
    }

    fmt.Printf("订单创建成功: %d\n", order.ID)
    return order, nil
}

// 2. 库存扣减锁（防止超卖）
func (s *LockService) DeductInventoryWithLock(ctx context.Context, productID uint, quantity int) error {
    lockKey := fmt.Sprintf("lock:inventory:%d", productID)
    lock := NewDistributedLock(s.rdb, lockKey, 10*time.Second)

    if err := lock.Acquire(ctx); err != nil {
        return fmt.Errorf("获取库存锁失败: %w", err)
    }
    defer lock.Release(ctx)

    // 检查库存
    var product model.Product
    if err := s.db.First(&product, productID).Error; err != nil {
        return err
    }

    if product.Stock < quantity {
        return errors.New("库存不足")
    }

    // 扣减库存
    result := s.db.Model(&product).Where("id = ? AND stock >= ?", productID, quantity).
              Update("stock", gorm.Expr("stock - ?", quantity))

    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return errors.New("库存扣减失败，可能库存不足")
    }

    fmt.Printf("库存扣减成功: 商品%d, 数量%d\n", productID, quantity)
    return nil
}

// 3. 缓存更新锁（防止缓存击穿）
func (s *LockService) UpdateCacheWithLock(ctx context.Context, cacheKey string, loader func() (interface{}, error)) (interface{}, error) {
    // 先检查缓存
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }

    // 缓存未命中，获取更新锁
    lockKey := fmt.Sprintf("lock:cache:%s", cacheKey)
    lock := NewDistributedLock(s.rdb, lockKey, 30*time.Second)

    if err := lock.TryAcquire(ctx, 50*time.Millisecond, 5); err != nil {
        // 获取锁失败，可能其他进程正在更新，等待一下再次尝试获取缓存
        time.Sleep(100 * time.Millisecond)
        cached, err := s.rdb.Get(ctx, cacheKey).Result()
        if err == nil {
            var result interface{}
            if err := json.Unmarshal([]byte(cached), &result); err == nil {
                return result, nil
            }
        }
        return nil, fmt.Errorf("获取缓存更新锁失败: %w", err)
    }
    defer lock.Release(ctx)

    // 双重检查缓存
    cached, err = s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }

    // 加载数据
    data, err := loader()
    if err != nil {
        return nil, err
    }

    // 更新缓存
    dataJSON, _ := json.Marshal(data)
    s.rdb.Set(ctx, cacheKey, dataJSON, time.Hour)

    return data, nil
}

// 4. 分布式任务锁（防止重复执行）
func (s *LockService) ExecuteTaskWithLock(ctx context.Context, taskID string, task func() error) error {
    lockKey := fmt.Sprintf("lock:task:%s", taskID)
    lock := NewDistributedLock(s.rdb, lockKey, 5*time.Minute)

    // 获取带自动续期的锁
    releaseFunc, err := lock.AcquireWithAutoRenew(ctx)
    if err != nil {
        return fmt.Errorf("获取任务锁失败: %w", err)
    }
    defer releaseFunc()

    // 执行任务
    fmt.Printf("开始执行任务: %s\n", taskID)
    if err := task(); err != nil {
        return fmt.Errorf("任务执行失败: %w", err)
    }

    fmt.Printf("任务执行完成: %s\n", taskID)
    return nil
}

// 5. 限流锁（基于滑动窗口）
func (s *LockService) RateLimitWithLock(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    lockKey := fmt.Sprintf("lock:ratelimit:%s", key)
    lock := NewDistributedLock(s.rdb, lockKey, 1*time.Second)

    if err := lock.Acquire(ctx); err != nil {
        return false, err
    }
    defer lock.Release(ctx)

    now := time.Now()
    windowStart := now.Add(-window)

    // 清理过期记录
    s.rdb.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart.UnixNano()))

    // 检查当前窗口内的请求数
    count, err := s.rdb.ZCard(ctx, key).Result()
    if err != nil {
        return false, err
    }

    if count >= int64(limit) {
        return false, nil // 超过限制
    }

    // 添加当前请求
    s.rdb.ZAdd(ctx, key, redis.Z{
        Score:  float64(now.UnixNano()),
        Member: fmt.Sprintf("%d", now.UnixNano()),
    })

    // 设置过期时间
    s.rdb.Expire(ctx, key, window)

    return true, nil
}

// 辅助结构体
type CreateOrderRequest struct {
    TotalAmount float64 `json:"total_amount"`
    Items       []OrderItem `json:"items"`
}

type OrderItem struct {
    ProductID uint `json:"product_id"`
    Quantity  int  `json:"quantity"`
}
```

---

## 📨 消息队列应用

Redis提供了多种消息队列实现方式，包括List、Pub/Sub、Stream等，适用于不同的业务场景。

### 基于List的简单队列

```go
// 来自 mall-go/internal/service/message_queue_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type MessageQueueService struct {
    rdb *redis.Client
}

func NewMessageQueueService(rdb *redis.Client) *MessageQueueService {
    return &MessageQueueService{rdb: rdb}
}

// 消息结构
type Message struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Timestamp int64                  `json:"timestamp"`
    Retry     int                    `json:"retry"`
}

// 1. 简单队列生产者
func (s *MessageQueueService) ProduceMessage(ctx context.Context, queueName string, msgType string, payload map[string]interface{}) error {
    message := Message{
        ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
        Type:      msgType,
        Payload:   payload,
        Timestamp: time.Now().Unix(),
        Retry:     0,
    }

    messageJSON, err := json.Marshal(message)
    if err != nil {
        return err
    }

    // 推入队列右端
    err = s.rdb.RPush(ctx, queueName, messageJSON).Err()
    if err != nil {
        return err
    }

    fmt.Printf("消息发送成功: %s -> %s\n", msgType, queueName)
    return nil
}

// 2. 简单队列消费者
func (s *MessageQueueService) ConsumeMessages(ctx context.Context, queueName string, handler func(Message) error) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("消费者停止: %s\n", queueName)
            return
        default:
            // 阻塞式弹出消息
            result, err := s.rdb.BLPop(ctx, 5*time.Second, queueName).Result()
            if err == redis.Nil {
                // 超时，继续循环
                continue
            }
            if err != nil {
                fmt.Printf("消费消息错误: %v\n", err)
                time.Sleep(time.Second)
                continue
            }

            // 解析消息
            var message Message
            if err := json.Unmarshal([]byte(result[1]), &message); err != nil {
                fmt.Printf("消息解析错误: %v\n", err)
                continue
            }

            // 处理消息
            if err := handler(message); err != nil {
                fmt.Printf("消息处理失败: %v\n", err)
                // 重试逻辑
                s.retryMessage(ctx, queueName, message)
            } else {
                fmt.Printf("消息处理成功: %s\n", message.ID)
            }
        }
    }
}

// 3. 延迟队列实现
func (s *MessageQueueService) ProduceDelayedMessage(ctx context.Context, queueName string, message Message, delay time.Duration) error {
    delayQueueName := fmt.Sprintf("%s:delayed", queueName)
    executeTime := time.Now().Add(delay).Unix()

    messageJSON, err := json.Marshal(message)
    if err != nil {
        return err
    }

    // 使用有序集合实现延迟队列
    err = s.rdb.ZAdd(ctx, delayQueueName, redis.Z{
        Score:  float64(executeTime),
        Member: messageJSON,
    }).Err()

    if err != nil {
        return err
    }

    fmt.Printf("延迟消息发送成功: %s, 延迟%v\n", message.ID, delay)
    return nil
}

// 4. 延迟队列处理器
func (s *MessageQueueService) ProcessDelayedMessages(ctx context.Context, queueName string) {
    delayQueueName := fmt.Sprintf("%s:delayed", queueName)
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            now := time.Now().Unix()

            // 获取到期的消息
            messages, err := s.rdb.ZRangeByScoreWithScores(ctx, delayQueueName, &redis.ZRangeBy{
                Min: "-inf",
                Max: fmt.Sprintf("%d", now),
            }).Result()

            if err != nil {
                fmt.Printf("获取延迟消息错误: %v\n", err)
                continue
            }

            for _, msg := range messages {
                // 移动到正常队列
                messageJSON := msg.Member.(string)

                // 使用事务确保原子性
                pipe := s.rdb.TxPipeline()
                pipe.ZRem(ctx, delayQueueName, messageJSON)
                pipe.RPush(ctx, queueName, messageJSON)

                _, err := pipe.Exec(ctx)
                if err != nil {
                    fmt.Printf("移动延迟消息错误: %v\n", err)
                    continue
                }

                fmt.Printf("延迟消息已到期，移动到队列: %s\n", queueName)
            }
        }
    }
}

// 5. 优先级队列
func (s *MessageQueueService) ProducePriorityMessage(ctx context.Context, queueName string, message Message, priority int) error {
    priorityQueueName := fmt.Sprintf("%s:priority", queueName)

    messageJSON, err := json.Marshal(message)
    if err != nil {
        return err
    }

    // 使用有序集合，分数越高优先级越高
    err = s.rdb.ZAdd(ctx, priorityQueueName, redis.Z{
        Score:  float64(priority),
        Member: messageJSON,
    }).Err()

    if err != nil {
        return err
    }

    fmt.Printf("优先级消息发送成功: %s, 优先级%d\n", message.ID, priority)
    return nil
}

// 6. 优先级队列消费者
func (s *MessageQueueService) ConsumePriorityMessages(ctx context.Context, queueName string, handler func(Message) error) {
    priorityQueueName := fmt.Sprintf("%s:priority", queueName)

    for {
        select {
        case <-ctx.Done():
            return
        default:
            // 获取最高优先级的消息
            messages, err := s.rdb.ZRevRangeWithScores(ctx, priorityQueueName, 0, 0).Result()
            if err != nil {
                fmt.Printf("获取优先级消息错误: %v\n", err)
                time.Sleep(time.Second)
                continue
            }

            if len(messages) == 0 {
                time.Sleep(100 * time.Millisecond)
                continue
            }

            messageJSON := messages[0].Member.(string)

            // 原子性移除消息
            removed, err := s.rdb.ZRem(ctx, priorityQueueName, messageJSON).Result()
            if err != nil || removed == 0 {
                continue // 可能被其他消费者处理了
            }

            // 解析并处理消息
            var message Message
            if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
                fmt.Printf("消息解析错误: %v\n", err)
                continue
            }

            if err := handler(message); err != nil {
                fmt.Printf("消息处理失败: %v\n", err)
                s.retryMessage(ctx, queueName, message)
            } else {
                fmt.Printf("优先级消息处理成功: %s\n", message.ID)
            }
        }
    }
}

// 重试消息处理
func (s *MessageQueueService) retryMessage(ctx context.Context, queueName string, message Message) {
    maxRetries := 3
    if message.Retry >= maxRetries {
        // 超过最大重试次数，移入死信队列
        deadLetterQueue := fmt.Sprintf("%s:dead", queueName)
        messageJSON, _ := json.Marshal(message)
        s.rdb.RPush(ctx, deadLetterQueue, messageJSON)
        fmt.Printf("消息移入死信队列: %s\n", message.ID)
        return
    }

    // 增加重试次数
    message.Retry++

    // 延迟重试（指数退避）
    delay := time.Duration(message.Retry*message.Retry) * time.Second
    s.ProduceDelayedMessage(ctx, queueName, message, delay)

    fmt.Printf("消息重试: %s, 第%d次重试\n", message.ID, message.Retry)
}
```

### 发布订阅模式

```go
// 发布订阅服务
type PubSubService struct {
    rdb *redis.Client
}

func NewPubSubService(rdb *redis.Client) *PubSubService {
    return &PubSubService{rdb: rdb}
}

// 1. 发布消息
func (s *PubSubService) Publish(ctx context.Context, channel string, message interface{}) error {
    messageJSON, err := json.Marshal(message)
    if err != nil {
        return err
    }

    subscribers, err := s.rdb.Publish(ctx, channel, messageJSON).Result()
    if err != nil {
        return err
    }

    fmt.Printf("消息发布成功: %s, 订阅者数量: %d\n", channel, subscribers)
    return nil
}

// 2. 订阅消息
func (s *PubSubService) Subscribe(ctx context.Context, channels []string, handler func(string, []byte) error) error {
    pubsub := s.rdb.Subscribe(ctx, channels...)
    defer pubsub.Close()

    // 等待订阅确认
    _, err := pubsub.Receive(ctx)
    if err != nil {
        return err
    }

    fmt.Printf("订阅成功: %v\n", channels)

    // 处理消息
    ch := pubsub.Channel()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-ch:
            if msg == nil {
                return nil
            }

            if err := handler(msg.Channel, []byte(msg.Payload)); err != nil {
                fmt.Printf("消息处理失败: %v\n", err)
            }
        }
    }
}

// 3. 模式订阅
func (s *PubSubService) PatternSubscribe(ctx context.Context, patterns []string, handler func(string, string, []byte) error) error {
    pubsub := s.rdb.PSubscribe(ctx, patterns...)
    defer pubsub.Close()

    // 等待订阅确认
    _, err := pubsub.Receive(ctx)
    if err != nil {
        return err
    }

    fmt.Printf("模式订阅成功: %v\n", patterns)

    // 处理消息
    ch := pubsub.Channel()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-ch:
            if msg == nil {
                return nil
            }

            if err := handler(msg.Pattern, msg.Channel, []byte(msg.Payload)); err != nil {
                fmt.Printf("模式消息处理失败: %v\n", err)
            }
        }
    }
}

// 4. 实时通知系统示例
func (s *PubSubService) StartNotificationSystem(ctx context.Context) {
    // 订阅用户通知
    go func() {
        s.Subscribe(ctx, []string{"user:notifications"}, func(channel string, data []byte) error {
            var notification struct {
                UserID  uint   `json:"user_id"`
                Type    string `json:"type"`
                Message string `json:"message"`
            }

            if err := json.Unmarshal(data, &notification); err != nil {
                return err
            }

            fmt.Printf("用户通知: 用户%d, 类型%s, 消息%s\n",
                      notification.UserID, notification.Type, notification.Message)

            // 这里可以推送到WebSocket、邮件、短信等
            return nil
        })
    }()

    // 订阅系统事件
    go func() {
        s.PatternSubscribe(ctx, []string{"system:*"}, func(pattern, channel string, data []byte) error {
            fmt.Printf("系统事件: 模式%s, 频道%s, 数据%s\n", pattern, channel, string(data))
            return nil
        })
    }()

    // 模拟发送通知
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                // 发送用户通知
                s.Publish(ctx, "user:notifications", map[string]interface{}{
                    "user_id": 1,
                    "type":    "order",
                    "message": "您的订单已发货",
                })

                // 发送系统事件
                s.Publish(ctx, "system:health", map[string]interface{}{
                    "status":    "healthy",
                    "timestamp": time.Now().Unix(),
                })
            }
        }
    }()
}
```

---

## ⚡ 性能优化技巧

Redis性能优化是企业级应用的关键，让我们掌握各种优化技巧。

### 连接池优化

```go
// 来自 mall-go/internal/config/redis_performance.go
package config

import (
    "context"
    "fmt"
    "runtime"
    "time"

    "github.com/redis/go-redis/v9"
)

// 性能优化配置
type RedisPerformanceConfig struct {
    // 连接池配置
    PoolSize        int           `yaml:"pool_size"`        // 连接池大小
    MinIdleConns    int           `yaml:"min_idle_conns"`   // 最小空闲连接数
    MaxConnAge      time.Duration `yaml:"max_conn_age"`     // 连接最大生存时间
    PoolTimeout     time.Duration `yaml:"pool_timeout"`     // 获取连接超时时间
    IdleTimeout     time.Duration `yaml:"idle_timeout"`     // 空闲连接超时时间
    IdleCheckFreq   time.Duration `yaml:"idle_check_freq"`  // 空闲连接检查频率

    // 网络配置
    DialTimeout  time.Duration `yaml:"dial_timeout"`  // 连接超时
    ReadTimeout  time.Duration `yaml:"read_timeout"`  // 读取超时
    WriteTimeout time.Duration `yaml:"write_timeout"` // 写入超时

    // 重试配置
    MaxRetries      int           `yaml:"max_retries"`       // 最大重试次数
    MinRetryBackoff time.Duration `yaml:"min_retry_backoff"` // 最小重试间隔
    MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"` // 最大重试间隔
}

// 创建高性能Redis客户端
func NewHighPerformanceRedisClient(cfg *RedisPerformanceConfig) *redis.Client {
    // 根据CPU核心数动态调整连接池大小
    if cfg.PoolSize == 0 {
        cfg.PoolSize = runtime.NumCPU() * 10
    }

    if cfg.MinIdleConns == 0 {
        cfg.MinIdleConns = cfg.PoolSize / 4
    }

    return redis.NewClient(&redis.Options{
        Addr: "localhost:6379",

        // 连接池优化
        PoolSize:      cfg.PoolSize,
        MinIdleConns:  cfg.MinIdleConns,
        MaxConnAge:    cfg.MaxConnAge,
        PoolTimeout:   cfg.PoolTimeout,
        IdleTimeout:   cfg.IdleTimeout,
        IdleCheckFreq: cfg.IdleCheckFreq,

        // 网络优化
        DialTimeout:  cfg.DialTimeout,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,

        // 重试优化
        MaxRetries:      cfg.MaxRetries,
        MinRetryBackoff: cfg.MinRetryBackoff,
        MaxRetryBackoff: cfg.MaxRetryBackoff,

        // 其他优化
        ReadOnly:       false,
        RouteByLatency: true,  // 按延迟路由
        RouteRandomly:  false, // 不随机路由
    })
}

// 默认高性能配置
func DefaultHighPerformanceConfig() *RedisPerformanceConfig {
    return &RedisPerformanceConfig{
        PoolSize:        100,
        MinIdleConns:    25,
        MaxConnAge:      30 * time.Minute,
        PoolTimeout:     4 * time.Second,
        IdleTimeout:     5 * time.Minute,
        IdleCheckFreq:   time.Minute,

        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,

        MaxRetries:      3,
        MinRetryBackoff: 8 * time.Millisecond,
        MaxRetryBackoff: 512 * time.Millisecond,
    }
}
```

### 管道操作优化

```go
// 管道操作服务
type PipelineService struct {
    rdb *redis.Client
}

func NewPipelineService(rdb *redis.Client) *PipelineService {
    return &PipelineService{rdb: rdb}
}

// 1. 批量操作优化
func (s *PipelineService) BatchSetUsers(ctx context.Context, users []model.User) error {
    pipe := s.rdb.Pipeline()

    for _, user := range users {
        cacheKey := fmt.Sprintf("user:%d", user.ID)
        userJSON, _ := json.Marshal(user)

        // 添加到管道
        pipe.Set(ctx, cacheKey, userJSON, time.Hour)
        pipe.SAdd(ctx, "users:active", user.ID)

        if user.IsVIP {
            pipe.SAdd(ctx, "users:vip", user.ID)
        }
    }

    // 执行管道
    start := time.Now()
    _, err := pipe.Exec(ctx)
    duration := time.Since(start)

    if err != nil {
        return err
    }

    fmt.Printf("批量设置%d个用户缓存，耗时: %v\n", len(users), duration)
    return nil
}

// 2. 批量获取优化
func (s *PipelineService) BatchGetUsers(ctx context.Context, userIDs []uint) (map[uint]*model.User, error) {
    if len(userIDs) == 0 {
        return make(map[uint]*model.User), nil
    }

    pipe := s.rdb.Pipeline()

    // 批量添加GET命令
    for _, userID := range userIDs {
        cacheKey := fmt.Sprintf("user:%d", userID)
        pipe.Get(ctx, cacheKey)
    }

    // 执行管道
    start := time.Now()
    cmds, err := pipe.Exec(ctx)
    duration := time.Since(start)

    if err != nil {
        return nil, err
    }

    // 处理结果
    result := make(map[uint]*model.User)
    for i, cmd := range cmds {
        userID := userIDs[i]
        getCmd := cmd.(*redis.StringCmd)

        val, err := getCmd.Result()
        if err == redis.Nil {
            continue // 键不存在
        }
        if err != nil {
            continue // 其他错误
        }

        var user model.User
        if err := json.Unmarshal([]byte(val), &user); err == nil {
            result[userID] = &user
        }
    }

    fmt.Printf("批量获取%d个用户缓存，命中%d个，耗时: %v\n",
               len(userIDs), len(result), duration)

    return result, nil
}

// 3. 事务管道优化
func (s *PipelineService) TransferPoints(ctx context.Context, fromUserID, toUserID uint, points int) error {
    // 使用事务管道确保原子性
    txPipe := s.rdb.TxPipeline()

    fromKey := fmt.Sprintf("user:%d:points", fromUserID)
    toKey := fmt.Sprintf("user:%d:points", toUserID)

    // 检查余额
    fromPoints, err := s.rdb.Get(ctx, fromKey).Int()
    if err != nil {
        return fmt.Errorf("获取用户%d积分失败: %w", fromUserID, err)
    }

    if fromPoints < points {
        return fmt.Errorf("用户%d积分不足", fromUserID)
    }

    // 添加事务命令
    txPipe.DecrBy(ctx, fromKey, int64(points))
    txPipe.IncrBy(ctx, toKey, int64(points))

    // 记录转账日志
    logKey := fmt.Sprintf("transfer:log:%d", time.Now().Unix())
    logData := map[string]interface{}{
        "from":   fromUserID,
        "to":     toUserID,
        "points": points,
        "time":   time.Now().Unix(),
    }
    logJSON, _ := json.Marshal(logData)
    txPipe.Set(ctx, logKey, logJSON, 24*time.Hour)

    // 执行事务
    _, err = txPipe.Exec(ctx)
    if err != nil {
        return fmt.Errorf("积分转账失败: %w", err)
    }

    fmt.Printf("积分转账成功: %d -> %d, 积分: %d\n", fromUserID, toUserID, points)
    return nil
}

// 4. 统计信息批量更新
func (s *PipelineService) UpdateStatistics(ctx context.Context, stats map[string]int64) error {
    pipe := s.rdb.Pipeline()

    for key, value := range stats {
        pipe.IncrBy(ctx, fmt.Sprintf("stats:%s", key), value)
        pipe.Expire(ctx, fmt.Sprintf("stats:%s", key), 24*time.Hour)
    }

    // 更新最后更新时间
    pipe.Set(ctx, "stats:last_update", time.Now().Unix(), 24*time.Hour)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return err
    }

    fmt.Printf("统计信息更新成功: %d项\n", len(stats))
    return nil
}
```

### 内存优化策略

```go
// 内存优化服务
type MemoryOptimizationService struct {
    rdb *redis.Client
}

func NewMemoryOptimizationService(rdb *redis.Client) *MemoryOptimizationService {
    return &MemoryOptimizationService{rdb: rdb}
}

// 1. 内存使用分析
func (s *MemoryOptimizationService) AnalyzeMemoryUsage(ctx context.Context) error {
    // 获取内存信息
    info, err := s.rdb.Info(ctx, "memory").Result()
    if err != nil {
        return err
    }

    fmt.Println("Redis内存使用情况:")
    fmt.Println(info)

    // 获取键空间信息
    info, err = s.rdb.Info(ctx, "keyspace").Result()
    if err != nil {
        return err
    }

    fmt.Println("Redis键空间信息:")
    fmt.Println(info)

    // 分析大键
    return s.findLargeKeys(ctx)
}

// 2. 查找大键
func (s *MemoryOptimizationService) findLargeKeys(ctx context.Context) error {
    var cursor uint64
    largeKeys := make(map[string]int64)

    for {
        keys, nextCursor, err := s.rdb.Scan(ctx, cursor, "*", 1000).Result()
        if err != nil {
            return err
        }

        // 检查每个键的内存使用
        for _, key := range keys {
            memUsage, err := s.rdb.MemoryUsage(ctx, key).Result()
            if err != nil {
                continue
            }

            // 大于1MB的键
            if memUsage > 1024*1024 {
                largeKeys[key] = memUsage
            }
        }

        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }

    fmt.Printf("发现%d个大键:\n", len(largeKeys))
    for key, size := range largeKeys {
        fmt.Printf("  %s: %d bytes (%.2f MB)\n", key, size, float64(size)/(1024*1024))
    }

    return nil
}

// 3. 过期键清理
func (s *MemoryOptimizationService) CleanupExpiredKeys(ctx context.Context) error {
    var cursor uint64
    cleanedCount := 0

    for {
        keys, nextCursor, err := s.rdb.Scan(ctx, cursor, "*", 1000).Result()
        if err != nil {
            return err
        }

        for _, key := range keys {
            ttl, err := s.rdb.TTL(ctx, key).Result()
            if err != nil {
                continue
            }

            // 已过期但未被清理的键
            if ttl == -2 {
                s.rdb.Del(ctx, key)
                cleanedCount++
            }
        }

        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }

    fmt.Printf("清理过期键: %d个\n", cleanedCount)
    return nil
}

// 4. 内存碎片整理
func (s *MemoryOptimizationService) DefragmentMemory(ctx context.Context) error {
    // 获取内存碎片率
    info, err := s.rdb.Info(ctx, "memory").Result()
    if err != nil {
        return err
    }

    fmt.Println("内存碎片整理前:")
    fmt.Println(info)

    // 执行内存整理（Redis 4.0+）
    result, err := s.rdb.Do(ctx, "MEMORY", "PURGE").Result()
    if err != nil {
        return err
    }

    fmt.Printf("内存整理结果: %v\n", result)

    // 获取整理后的内存信息
    info, err = s.rdb.Info(ctx, "memory").Result()
    if err != nil {
        return err
    }

    fmt.Println("内存碎片整理后:")
    fmt.Println(info)

    return nil
}

// 5. 键过期策略优化
func (s *MemoryOptimizationService) OptimizeExpirationStrategy(ctx context.Context) error {
    // 设置不同类型数据的过期策略
    strategies := map[string]time.Duration{
        "user:*":         time.Hour,        // 用户信息1小时
        "product:*":      30 * time.Minute, // 商品信息30分钟
        "session:*":      15 * time.Minute, // 会话信息15分钟
        "cache:*":        5 * time.Minute,  // 临时缓存5分钟
        "stats:*":        24 * time.Hour,   // 统计信息24小时
        "config:*":       7 * 24 * time.Hour, // 配置信息7天
    }

    var cursor uint64
    updatedCount := 0

    for {
        keys, nextCursor, err := s.rdb.Scan(ctx, cursor, "*", 1000).Result()
        if err != nil {
            return err
        }

        for _, key := range keys {
            // 检查是否已有过期时间
            ttl, err := s.rdb.TTL(ctx, key).Result()
            if err != nil {
                continue
            }

            if ttl == -1 { // 没有过期时间
                // 根据键模式设置过期时间
                for pattern, duration := range strategies {
                    matched, _ := filepath.Match(pattern, key)
                    if matched {
                        s.rdb.Expire(ctx, key, duration)
                        updatedCount++
                        break
                    }
                }
            }
        }

        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }

    fmt.Printf("优化过期策略: %d个键\n", updatedCount)
    return nil
}
```

---

## 🏢 实战案例分析

让我们通过mall-go项目的真实案例，看看如何在企业级项目中应用Redis。

### 电商系统缓存架构

```go
// 来自 mall-go/internal/service/mall_cache_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"
    "time"

    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
    "mall-go/internal/model"
)

// 商城缓存服务
type MallCacheService struct {
    rdb         *redis.Client
    db          *gorm.DB
    lockService *LockService
}

func NewMallCacheService(rdb *redis.Client, db *gorm.DB) *MallCacheService {
    return &MallCacheService{
        rdb:         rdb,
        db:          db,
        lockService: NewLockService(rdb),
    }
}

// 1. 商品详情页缓存策略
func (s *MallCacheService) GetProductDetail(ctx context.Context, productID uint) (*ProductDetailVO, error) {
    cacheKey := fmt.Sprintf("product:detail:%d", productID)

    // 1. 尝试从缓存获取
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var detail ProductDetailVO
        if err := json.Unmarshal([]byte(cached), &detail); err == nil {
            // 异步更新访问统计
            go s.incrementProductViews(context.Background(), productID)
            return &detail, nil
        }
    }

    // 2. 缓存未命中，使用分布式锁防止缓存击穿
    lockKey := fmt.Sprintf("lock:product:detail:%d", productID)
    lock := NewDistributedLock(s.rdb, lockKey, 10*time.Second)

    if err := lock.TryAcquire(ctx, 50*time.Millisecond, 5); err != nil {
        // 获取锁失败，等待后重试获取缓存
        time.Sleep(100 * time.Millisecond)
        cached, err := s.rdb.Get(ctx, cacheKey).Result()
        if err == nil {
            var detail ProductDetailVO
            if err := json.Unmarshal([]byte(cached), &detail); err == nil {
                return &detail, nil
            }
        }
        return nil, fmt.Errorf("获取商品详情失败: %w", err)
    }
    defer lock.Release(ctx)

    // 3. 双重检查缓存
    cached, err = s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var detail ProductDetailVO
        if err := json.Unmarshal([]byte(cached), &detail); err == nil {
            return &detail, nil
        }
    }

    // 4. 从数据库加载数据
    detail, err := s.loadProductDetailFromDB(ctx, productID)
    if err != nil {
        return nil, err
    }

    // 5. 写入缓存（随机过期时间防止雪崩）
    detailJSON, _ := json.Marshal(detail)
    expireTime := time.Hour + time.Duration(rand.Intn(1800))*time.Second // 1-1.5小时
    s.rdb.Set(ctx, cacheKey, detailJSON, expireTime)

    // 6. 异步更新相关缓存
    go s.updateRelatedCache(context.Background(), productID, detail)

    return detail, nil
}

// 2. 购物车缓存实现
func (s *MallCacheService) AddToCart(ctx context.Context, userID, productID uint, quantity int) error {
    cartKey := fmt.Sprintf("cart:user:%d", userID)
    productKey := fmt.Sprintf("product:%d", productID)

    // 使用Hash存储购物车
    pipe := s.rdb.Pipeline()

    // 增加商品数量
    pipe.HIncrBy(ctx, cartKey, productKey, int64(quantity))

    // 设置购物车过期时间（7天）
    pipe.Expire(ctx, cartKey, 7*24*time.Hour)

    // 更新购物车商品总数缓存
    pipe.HLen(ctx, cartKey)

    results, err := pipe.Exec(ctx)
    if err != nil {
        return err
    }

    // 获取购物车商品总数
    totalItems := results[2].(*redis.IntCmd).Val()

    // 更新用户购物车计数器
    counterKey := fmt.Sprintf("cart:count:user:%d", userID)
    s.rdb.Set(ctx, counterKey, totalItems, 7*24*time.Hour)

    // 发布购物车更新事件
    s.publishCartUpdateEvent(ctx, userID, "add", productID, quantity)

    fmt.Printf("商品添加到购物车: 用户%d, 商品%d, 数量%d\n", userID, productID, quantity)
    return nil
}

// 3. 秒杀活动缓存策略
func (s *MallCacheService) SeckillProduct(ctx context.Context, userID, productID uint) error {
    seckillKey := fmt.Sprintf("seckill:product:%d", productID)
    userKey := fmt.Sprintf("seckill:user:%d:product:%d", userID, productID)

    // 1. 检查用户是否已参与
    participated, err := s.rdb.Exists(ctx, userKey).Result()
    if err != nil {
        return err
    }
    if participated > 0 {
        return errors.New("您已参与过此次秒杀")
    }

    // 2. 使用Lua脚本实现原子性秒杀
    luaScript := `
        local seckill_key = KEYS[1]
        local user_key = KEYS[2]
        local user_id = ARGV[1]

        -- 检查库存
        local stock = redis.call('get', seckill_key)
        if not stock or tonumber(stock) <= 0 then
            return 0  -- 库存不足
        end

        -- 扣减库存
        local new_stock = redis.call('decr', seckill_key)
        if new_stock < 0 then
            redis.call('incr', seckill_key)  -- 回滚
            return 0  -- 库存不足
        end

        -- 记录用户参与
        redis.call('setex', user_key, 3600, user_id)

        return 1  -- 秒杀成功
    `

    result, err := s.rdb.Eval(ctx, luaScript, []string{seckillKey, userKey}, userID).Result()
    if err != nil {
        return err
    }

    if result.(int64) == 0 {
        return errors.New("秒杀失败，商品已售罄")
    }

    // 3. 异步处理订单创建
    go s.createSeckillOrder(context.Background(), userID, productID)

    fmt.Printf("秒杀成功: 用户%d, 商品%d\n", userID, productID)
    return nil
}

// 4. 用户会话管理
func (s *MallCacheService) CreateUserSession(ctx context.Context, userID uint, deviceInfo string) (string, error) {
    sessionID := generateSessionID()
    sessionKey := fmt.Sprintf("session:%s", sessionID)

    sessionData := map[string]interface{}{
        "user_id":     userID,
        "device_info": deviceInfo,
        "login_time":  time.Now().Unix(),
        "last_active": time.Now().Unix(),
    }

    // 使用Hash存储会话数据
    pipe := s.rdb.Pipeline()
    pipe.HMSet(ctx, sessionKey, sessionData)
    pipe.Expire(ctx, sessionKey, 30*time.Minute) // 30分钟过期

    // 维护用户活跃会话列表
    userSessionsKey := fmt.Sprintf("user:sessions:%d", userID)
    pipe.SAdd(ctx, userSessionsKey, sessionID)
    pipe.Expire(ctx, userSessionsKey, 24*time.Hour)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return "", err
    }

    fmt.Printf("用户会话创建: 用户%d, 会话%s\n", userID, sessionID)
    return sessionID, nil
}

// 5. 实时统计数据
func (s *MallCacheService) UpdateRealTimeStats(ctx context.Context, event string, value int64) error {
    now := time.Now()

    // 按不同时间维度统计
    timeFormats := map[string]string{
        "minute": now.Format("2006-01-02 15:04"),
        "hour":   now.Format("2006-01-02 15"),
        "day":    now.Format("2006-01-02"),
        "month":  now.Format("2006-01"),
    }

    pipe := s.rdb.Pipeline()

    for dimension, timeKey := range timeFormats {
        statsKey := fmt.Sprintf("stats:%s:%s:%s", event, dimension, timeKey)
        pipe.IncrBy(ctx, statsKey, value)

        // 设置不同的过期时间
        var expireTime time.Duration
        switch dimension {
        case "minute":
            expireTime = 2 * time.Hour
        case "hour":
            expireTime = 7 * 24 * time.Hour
        case "day":
            expireTime = 30 * 24 * time.Hour
        case "month":
            expireTime = 365 * 24 * time.Hour
        }

        pipe.Expire(ctx, statsKey, expireTime)
    }

    _, err := pipe.Exec(ctx)
    if err != nil {
        return err
    }

    // 更新实时排行榜
    if event == "product_view" || event == "product_purchase" {
        s.updateProductRanking(ctx, event, value)
    }

    return nil
}

// 辅助方法
func (s *MallCacheService) loadProductDetailFromDB(ctx context.Context, productID uint) (*ProductDetailVO, error) {
    // 模拟从数据库加载商品详情
    var product model.Product
    if err := s.db.Preload("Category").Preload("Images").First(&product, productID).Error; err != nil {
        return nil, err
    }

    // 转换为VO
    detail := &ProductDetailVO{
        ID:          product.ID,
        Name:        product.Name,
        Price:       product.Price,
        Description: product.Description,
        Stock:       product.Stock,
        CategoryID:  product.CategoryID,
        Images:      product.Images,
        CreatedAt:   product.CreatedAt,
    }

    return detail, nil
}

func (s *MallCacheService) incrementProductViews(ctx context.Context, productID uint) {
    viewKey := fmt.Sprintf("product:views:%d", productID)
    s.rdb.Incr(ctx, viewKey)

    // 更新今日浏览统计
    s.UpdateRealTimeStats(ctx, "product_view", 1)
}

func (s *MallCacheService) updateRelatedCache(ctx context.Context, productID uint, detail *ProductDetailVO) {
    // 更新分类商品列表缓存
    categoryKey := fmt.Sprintf("category:products:%d", detail.CategoryID)
    s.rdb.Del(ctx, categoryKey) // 删除缓存，下次访问时重新加载

    // 更新搜索索引
    searchKey := fmt.Sprintf("search:product:%d", productID)
    searchData := map[string]interface{}{
        "name":        detail.Name,
        "category_id": detail.CategoryID,
        "price":       detail.Price,
    }
    searchJSON, _ := json.Marshal(searchData)
    s.rdb.Set(ctx, searchKey, searchJSON, 24*time.Hour)
}

func (s *MallCacheService) publishCartUpdateEvent(ctx context.Context, userID uint, action string, productID uint, quantity int) {
    event := map[string]interface{}{
        "user_id":    userID,
        "action":     action,
        "product_id": productID,
        "quantity":   quantity,
        "timestamp":  time.Now().Unix(),
    }

    eventJSON, _ := json.Marshal(event)
    s.rdb.Publish(ctx, "cart:updates", eventJSON)
}

func (s *MallCacheService) createSeckillOrder(ctx context.Context, userID, productID uint) {
    // 模拟异步创建秒杀订单
    fmt.Printf("异步创建秒杀订单: 用户%d, 商品%d\n", userID, productID)
}

func (s *MallCacheService) updateProductRanking(ctx context.Context, event string, productID int64) {
    rankingKey := fmt.Sprintf("ranking:%s:daily", event)
    s.rdb.ZIncrBy(ctx, rankingKey, 1, fmt.Sprintf("%d", productID))
    s.rdb.Expire(ctx, rankingKey, 24*time.Hour)
}

func generateSessionID() string {
    return fmt.Sprintf("sess_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

// 相关结构体
type ProductDetailVO struct {
    ID          uint                   `json:"id"`
    Name        string                 `json:"name"`
    Price       float64                `json:"price"`
    Description string                 `json:"description"`
    Stock       int                    `json:"stock"`
    CategoryID  uint                   `json:"category_id"`
    Images      []model.ProductImage   `json:"images"`
    CreatedAt   time.Time              `json:"created_at"`
}
```

---

## 🎯 面试常考点

### 1. Redis数据类型和应用场景

**问题：** Redis有哪些数据类型？分别适用于什么场景？

**答案：**
```go
/*
Redis五大基础数据类型及应用场景：

1. String（字符串）
   - 应用场景：缓存、计数器、分布式锁、会话存储
   - 示例：用户信息缓存、页面访问计数、验证码存储

2. Hash（哈希）
   - 应用场景：对象存储、用户属性、配置信息
   - 示例：用户资料、商品属性、系统配置

3. List（列表）
   - 应用场景：消息队列、最新动态、栈和队列
   - 示例：消息队列、用户动态、浏览历史

4. Set（集合）
   - 应用场景：标签、关注关系、去重、交并差集运算
   - 示例：用户标签、共同好友、抽奖去重

5. ZSet（有序集合）
   - 应用场景：排行榜、优先级队列、范围查询
   - 示例：游戏排行榜、热搜榜、延迟队列

高级数据类型：
- Bitmap：用户签到、在线状态统计
- HyperLogLog：UV统计、基数估算
- Geo：地理位置、附近的人
- Stream：消息流、事件溯源
*/

// 实际应用示例
func DataTypeExamples(rdb *redis.Client, ctx context.Context) {
    // String - 计数器
    rdb.Incr(ctx, "page:views")

    // Hash - 用户信息
    rdb.HMSet(ctx, "user:1", map[string]interface{}{
        "name": "张三",
        "age":  25,
        "city": "北京",
    })

    // List - 消息队列
    rdb.LPush(ctx, "task:queue", "task1", "task2")

    // Set - 标签系统
    rdb.SAdd(ctx, "user:1:tags", "技术", "编程", "Go")

    // ZSet - 排行榜
    rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 1000, Member: "player1"})
}
```

### 2. 缓存穿透、击穿、雪崩

**问题：** 什么是缓存穿透、击穿、雪崩？如何解决？

**答案：**
```go
/*
三大缓存问题及解决方案：

1. 缓存穿透（Cache Penetration）
   - 问题：查询不存在的数据，缓存和数据库都没有，导致每次都查数据库
   - 解决方案：
     a) 缓存空值（设置较短过期时间）
     b) 布隆过滤器预先过滤
     c) 参数校验

2. 缓存击穿（Cache Breakdown）
   - 问题：热点数据过期，大量并发请求同时访问数据库
   - 解决方案：
     a) 分布式锁（只允许一个请求查数据库）
     b) 热点数据永不过期
     c) 异步刷新缓存

3. 缓存雪崩（Cache Avalanche）
   - 问题：大量缓存同时过期，数据库压力激增
   - 解决方案：
     a) 随机过期时间
     b) 缓存预热
     c) 多级缓存
     d) 限流降级
*/

// 解决方案实现
func CacheProblemSolutions(rdb *redis.Client, ctx context.Context) {
    // 1. 缓存穿透 - 布隆过滤器
    exists, _ := rdb.BFExists(ctx, "user:bloom", "user123").Result()
    if !exists {
        return // 直接返回，不查数据库
    }

    // 2. 缓存击穿 - 分布式锁
    lockKey := "lock:user:123"
    locked, _ := rdb.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
    if locked {
        // 获取锁成功，查询数据库并更新缓存
        defer rdb.Del(ctx, lockKey)
        // ... 查询数据库逻辑
    } else {
        // 等待其他线程更新缓存
        time.Sleep(50 * time.Millisecond)
        // 重试获取缓存
    }

    // 3. 缓存雪崩 - 随机过期时间
    baseExpire := time.Hour
    randomExpire := baseExpire + time.Duration(rand.Intn(1800))*time.Second
    rdb.Set(ctx, "key", "value", randomExpire)
}
```

### 3. Redis持久化机制

**问题：** Redis的持久化机制有哪些？各有什么优缺点？

**答案：**
```go
/*
Redis持久化机制对比：

1. RDB（Redis Database）
   - 原理：定期生成数据快照保存到磁盘
   - 优点：
     * 文件紧凑，适合备份
     * 恢复速度快
     * 对性能影响小
   - 缺点：
     * 可能丢失最后一次快照后的数据
     * fork子进程时会阻塞主进程
   - 配置：save 900 1（900秒内至少1个key变化）

2. AOF（Append Only File）
   - 原理：记录每个写操作命令到日志文件
   - 优点：
     * 数据安全性高，最多丢失1秒数据
     * 日志文件可读性好
     * 支持自动重写压缩
   - 缺点：
     * 文件体积大
     * 恢复速度慢
     * 对性能影响较大
   - 配置：appendfsync everysec（每秒同步）

3. 混合持久化（Redis 4.0+）
   - 原理：AOF重写时使用RDB格式，增量使用AOF格式
   - 优点：结合两者优势，快速恢复且数据安全
   - 配置：aof-use-rdb-preamble yes

推荐配置：
- 生产环境：开启AOF + 定期RDB备份
- 缓存场景：仅RDB即可
- 高可用场景：混合持久化
*/

// 持久化配置示例
func PersistenceConfig() {
    /*
    # redis.conf 配置示例

    # RDB配置
    save 900 1      # 900秒内至少1个key变化
    save 300 10     # 300秒内至少10个key变化
    save 60 10000   # 60秒内至少10000个key变化

    # AOF配置
    appendonly yes
    appendfsync everysec
    auto-aof-rewrite-percentage 100
    auto-aof-rewrite-min-size 64mb

    # 混合持久化
    aof-use-rdb-preamble yes
    */
}
```

### 4. Redis集群和高可用

**问题：** Redis如何实现高可用？集群模式有什么特点？

**答案：**
```go
/*
Redis高可用方案：

1. 主从复制（Master-Slave）
   - 特点：一主多从，读写分离
   - 优点：提高读性能，数据备份
   - 缺点：主节点故障需手动切换
   - 配置：slaveof <masterip> <masterport>

2. 哨兵模式（Sentinel）
   - 特点：自动故障转移，监控主从状态
   - 优点：自动切换，高可用
   - 缺点：不能解决写性能瓶颈
   - 配置：sentinel monitor mymaster 127.0.0.1 6379 2

3. 集群模式（Cluster）
   - 特点：分布式存储，数据分片
   - 优点：水平扩展，高性能
   - 缺点：运维复杂，不支持多数据库
   - 配置：cluster-enabled yes

集群特性：
- 16384个槽位分配给不同节点
- 支持自动故障转移
- 最少需要3个主节点
- 每个主节点建议至少1个从节点
*/

// Go客户端集群配置
func ClusterConfig() *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{
            "127.0.0.1:7000",
            "127.0.0.1:7001",
            "127.0.0.1:7002",
        },
        Password: "",

        // 集群配置
        MaxRedirects:   8,
        ReadOnly:       false,
        RouteByLatency: true,
        RouteRandomly:  false,

        // 连接池配置
        PoolSize:     100,
        MinIdleConns: 10,
    })
}
```

---

## ⚠️ 踩坑提醒

### 1. 连接池配置陷阱

```go
// ❌ 错误：连接池配置不当
func BadPoolConfig() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        PoolSize: 1000,      // 过大的连接池
        PoolTimeout: 1*time.Second, // 过短的超时时间
    })
}

// ✅ 正确：合理的连接池配置
func GoodPoolConfig() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:         "localhost:6379",
        PoolSize:     runtime.NumCPU() * 10, // 根据CPU核心数调整
        MinIdleConns: 5,                     // 保持最小空闲连接
        PoolTimeout:  4 * time.Second,       // 合理的超时时间
        IdleTimeout:  5 * time.Minute,       // 空闲连接超时
        MaxConnAge:   30 * time.Minute,      // 连接最大生存时间
    })
}
```

### 2. 键命名陷阱

```go
// ❌ 错误：键命名不规范
func BadKeyNaming(rdb *redis.Client, ctx context.Context) {
    // 没有命名空间
    rdb.Set(ctx, "user", "data", time.Hour)

    // 使用特殊字符
    rdb.Set(ctx, "user:name with spaces", "data", time.Hour)

    // 键名过长
    rdb.Set(ctx, "this_is_a_very_long_key_name_that_should_be_avoided", "data", time.Hour)
}

// ✅ 正确：规范的键命名
func GoodKeyNaming(rdb *redis.Client, ctx context.Context) {
    // 使用命名空间和层级结构
    rdb.Set(ctx, "mall:user:1:profile", "data", time.Hour)

    // 使用冒号分隔，避免特殊字符
    rdb.Set(ctx, "mall:product:123:detail", "data", time.Hour)

    // 简洁明了的键名
    rdb.Set(ctx, "mall:cache:hot_products", "data", time.Hour)
}
```

### 3. 内存泄漏陷阱

```go
// ❌ 错误：没有设置过期时间
func MemoryLeak(rdb *redis.Client, ctx context.Context) {
    // 永不过期的键会导致内存泄漏
    rdb.Set(ctx, "temp:data", "value", 0)

    // 大量临时数据没有清理
    for i := 0; i < 10000; i++ {
        rdb.Set(ctx, fmt.Sprintf("temp:%d", i), "data", 0)
    }
}

// ✅ 正确：合理设置过期时间
func MemoryManagement(rdb *redis.Client, ctx context.Context) {
    // 设置合理的过期时间
    rdb.Set(ctx, "temp:data", "value", 10*time.Minute)

    // 批量设置临时数据的过期时间
    pipe := rdb.Pipeline()
    for i := 0; i < 10000; i++ {
        key := fmt.Sprintf("temp:%d", i)
        pipe.Set(ctx, key, "data", time.Hour)
    }
    pipe.Exec(ctx)
}
```

### 4. 事务使用陷阱

```go
// ❌ 错误：事务使用不当
func BadTransaction(rdb *redis.Client, ctx context.Context) {
    // 在事务中使用阻塞命令
    pipe := rdb.TxPipeline()
    pipe.BLPop(ctx, 0, "queue") // 阻塞命令不应在事务中使用
    pipe.Exec(ctx)

    // 事务中使用随机命令
    pipe = rdb.TxPipeline()
    pipe.Randomkey(ctx) // 随机命令结果不确定
    pipe.Exec(ctx)
}

// ✅ 正确：正确使用事务
func GoodTransaction(rdb *redis.Client, ctx context.Context) {
    // 使用WATCH实现乐观锁
    err := rdb.Watch(ctx, func(tx *redis.Tx) error {
        // 检查条件
        val, err := tx.Get(ctx, "balance").Int()
        if err != nil {
            return err
        }

        if val < 100 {
            return errors.New("余额不足")
        }

        // 执行事务
        _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
            pipe.DecrBy(ctx, "balance", 100)
            pipe.IncrBy(ctx, "spent", 100)
            return nil
        })

        return err
    }, "balance")

    if err != nil {
        fmt.Printf("事务执行失败: %v\n", err)
    }
}
```

### 5. 大键问题陷阱

```go
// ❌ 错误：创建大键
func CreateBigKey(rdb *redis.Client, ctx context.Context) {
    // 单个键存储大量数据
    largeData := make([]string, 1000000)
    for i := range largeData {
        largeData[i] = fmt.Sprintf("data_%d", i)
    }

    // 这会创建一个非常大的键
    rdb.SAdd(ctx, "big_set", largeData)
}

// ✅ 正确：拆分大键
func SplitBigKey(rdb *redis.Client, ctx context.Context) {
    // 将大键拆分成多个小键
    batchSize := 1000
    totalData := 1000000

    for batch := 0; batch < totalData/batchSize; batch++ {
        key := fmt.Sprintf("data_set:%d", batch)

        batchData := make([]interface{}, batchSize)
        for i := 0; i < batchSize; i++ {
            batchData[i] = fmt.Sprintf("data_%d", batch*batchSize+i)
        }

        rdb.SAdd(ctx, key, batchData...)
        rdb.Expire(ctx, key, time.Hour) // 设置过期时间
    }
}
```

---

## 📝 练习题

### 练习题1：分布式限流器实现（⭐⭐）

**题目描述：**
实现一个基于Redis的分布式限流器，支持固定窗口和滑动窗口两种算法，要求线程安全且高性能。

```go
// 练习题1：分布式限流器实现
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// 解答：
// 限流器接口
type RateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
    AllowN(ctx context.Context, key string, n int) (bool, error)
}

// 1. 固定窗口限流器
type FixedWindowLimiter struct {
    rdb    *redis.Client
    limit  int           // 限制次数
    window time.Duration // 时间窗口
}

func NewFixedWindowLimiter(rdb *redis.Client, limit int, window time.Duration) *FixedWindowLimiter {
    return &FixedWindowLimiter{
        rdb:    rdb,
        limit:  limit,
        window: window,
    }
}

func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
    return l.AllowN(ctx, key, 1)
}

func (l *FixedWindowLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
    now := time.Now()
    windowStart := now.Truncate(l.window)
    windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

    // 使用Lua脚本保证原子性
    luaScript := `
        local key = KEYS[1]
        local limit = tonumber(ARGV[1])
        local n = tonumber(ARGV[2])
        local window = tonumber(ARGV[3])

        local current = redis.call('GET', key)
        if current == false then
            current = 0
        else
            current = tonumber(current)
        end

        if current + n <= limit then
            redis.call('INCRBY', key, n)
            redis.call('EXPIRE', key, window)
            return 1
        else
            return 0
        end
    `

    result, err := l.rdb.Eval(ctx, luaScript, []string{windowKey},
                             l.limit, n, int(l.window.Seconds())).Result()
    if err != nil {
        return false, err
    }

    return result.(int64) == 1, nil
}

// 2. 滑动窗口限流器
type SlidingWindowLimiter struct {
    rdb    *redis.Client
    limit  int           // 限制次数
    window time.Duration // 时间窗口
}

func NewSlidingWindowLimiter(rdb *redis.Client, limit int, window time.Duration) *SlidingWindowLimiter {
    return &SlidingWindowLimiter{
        rdb:    rdb,
        limit:  limit,
        window: window,
    }
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
    return l.AllowN(ctx, key, 1)
}

func (l *SlidingWindowLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
    now := time.Now()
    windowStart := now.Add(-l.window)

    luaScript := `
        local key = KEYS[1]
        local limit = tonumber(ARGV[1])
        local n = tonumber(ARGV[2])
        local now = tonumber(ARGV[3])
        local window_start = tonumber(ARGV[4])

        -- 清理过期记录
        redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

        -- 检查当前窗口内的请求数
        local current = redis.call('ZCARD', key)

        if current + n <= limit then
            -- 添加当前请求
            for i = 1, n do
                redis.call('ZADD', key, now, now .. ':' .. i)
            end

            -- 设置过期时间
            redis.call('EXPIRE', key, math.ceil(ARGV[5]))
            return 1
        else
            return 0
        end
    `

    result, err := l.rdb.Eval(ctx, luaScript, []string{key},
                             l.limit, n, now.UnixNano(), windowStart.UnixNano(),
                             l.window.Seconds()).Result()
    if err != nil {
        return false, err
    }

    return result.(int64) == 1, nil
}

// 3. 令牌桶限流器
type TokenBucketLimiter struct {
    rdb      *redis.Client
    capacity int           // 桶容量
    rate     int           // 令牌生成速率（每秒）
}

func NewTokenBucketLimiter(rdb *redis.Client, capacity, rate int) *TokenBucketLimiter {
    return &TokenBucketLimiter{
        rdb:      rdb,
        capacity: capacity,
        rate:     rate,
    }
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
    return l.AllowN(ctx, key, 1)
}

func (l *TokenBucketLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
    now := time.Now().Unix()

    luaScript := `
        local key = KEYS[1]
        local capacity = tonumber(ARGV[1])
        local rate = tonumber(ARGV[2])
        local n = tonumber(ARGV[3])
        local now = tonumber(ARGV[4])

        local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
        local tokens = tonumber(bucket[1]) or capacity
        local last_refill = tonumber(bucket[2]) or now

        -- 计算需要添加的令牌数
        local time_passed = now - last_refill
        local new_tokens = math.min(capacity, tokens + time_passed * rate)

        if new_tokens >= n then
            new_tokens = new_tokens - n
            redis.call('HMSET', key, 'tokens', new_tokens, 'last_refill', now)
            redis.call('EXPIRE', key, 3600)
            return 1
        else
            redis.call('HMSET', key, 'tokens', new_tokens, 'last_refill', now)
            redis.call('EXPIRE', key, 3600)
            return 0
        end
    `

    result, err := l.rdb.Eval(ctx, luaScript, []string{key},
                             l.capacity, l.rate, n, now).Result()
    if err != nil {
        return false, err
    }

    return result.(int64) == 1, nil
}

// 4. 测试函数
func TestRateLimiters() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer rdb.Close()

    ctx := context.Background()

    // 测试固定窗口限流器
    fmt.Println("=== 固定窗口限流器测试 ===")
    fixedLimiter := NewFixedWindowLimiter(rdb, 5, time.Minute)

    for i := 0; i < 7; i++ {
        allowed, _ := fixedLimiter.Allow(ctx, "user:123")
        fmt.Printf("请求 %d: %v\n", i+1, allowed)
    }

    // 测试滑动窗口限流器
    fmt.Println("\n=== 滑动窗口限流器测试 ===")
    slidingLimiter := NewSlidingWindowLimiter(rdb, 5, time.Minute)

    for i := 0; i < 7; i++ {
        allowed, _ := slidingLimiter.Allow(ctx, "user:456")
        fmt.Printf("请求 %d: %v\n", i+1, allowed)
        time.Sleep(10 * time.Second) // 模拟时间间隔
    }

    // 测试令牌桶限流器
    fmt.Println("\n=== 令牌桶限流器测试 ===")
    tokenLimiter := NewTokenBucketLimiter(rdb, 10, 2) // 容量10，每秒生成2个令牌

    for i := 0; i < 15; i++ {
        allowed, _ := tokenLimiter.Allow(ctx, "user:789")
        fmt.Printf("请求 %d: %v\n", i+1, allowed)
        time.Sleep(time.Second)
    }
}

/*
解析说明：
1. 固定窗口：简单高效，但可能出现突刺流量
2. 滑动窗口：更平滑的限流，但内存开销较大
3. 令牌桶：支持突发流量，适合大多数场景
4. 使用Lua脚本保证操作的原子性
5. 合理设置过期时间避免内存泄漏

扩展思考：
- 如何实现分布式环境下的限流？
- 如何处理Redis故障时的降级策略？
- 如何监控和调整限流参数？
- 如何实现更复杂的限流策略（如用户级别限流）？
*/
```

### 练习题2：Redis缓存一致性解决方案（⭐⭐⭐）

**题目描述：**
设计一个Redis缓存一致性解决方案，支持多种更新策略，处理并发更新问题，确保缓存与数据库的数据一致性。

```go
// 练习题2：Redis缓存一致性解决方案
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// 解答：
// 缓存一致性策略枚举
type ConsistencyStrategy int

const (
    CacheAside ConsistencyStrategy = iota
    WriteThrough
    WriteBehind
    RefreshAhead
)

// 缓存一致性管理器
type CacheConsistencyManager struct {
    rdb      *redis.Client
    db       *gorm.DB
    strategy ConsistencyStrategy

    // 写后延迟队列
    writeQueue chan *WriteTask

    // 刷新任务管理
    refreshTasks sync.Map

    // 版本控制
    versionManager *VersionManager
}

// 写任务结构
type WriteTask struct {
    Key   string
    Value interface{}
    TTL   time.Duration
}

// 版本管理器
type VersionManager struct {
    rdb *redis.Client
}

func NewVersionManager(rdb *redis.Client) *VersionManager {
    return &VersionManager{rdb: rdb}
}

func (vm *VersionManager) GetVersion(ctx context.Context, key string) (int64, error) {
    versionKey := fmt.Sprintf("version:%s", key)
    version, err := vm.rdb.Get(ctx, versionKey).Int64()
    if err == redis.Nil {
        return 0, nil
    }
    return version, err
}

func (vm *VersionManager) IncrVersion(ctx context.Context, key string) (int64, error) {
    versionKey := fmt.Sprintf("version:%s", key)
    return vm.rdb.Incr(ctx, versionKey).Result()
}

func (vm *VersionManager) CompareAndSet(ctx context.Context, key string, expectedVersion int64, value interface{}, ttl time.Duration) (bool, error) {
    versionKey := fmt.Sprintf("version:%s", key)

    luaScript := `
        local version_key = KEYS[1]
        local data_key = KEYS[2]
        local expected_version = tonumber(ARGV[1])
        local new_value = ARGV[2]
        local ttl = tonumber(ARGV[3])

        local current_version = redis.call('GET', version_key)
        if current_version == false then
            current_version = 0
        else
            current_version = tonumber(current_version)
        end

        if current_version == expected_version then
            local new_version = current_version + 1
            redis.call('SET', version_key, new_version)
            redis.call('SET', data_key, new_value, 'EX', ttl)
            return new_version
        else
            return 0
        end
    `

    valueJSON, _ := json.Marshal(value)
    result, err := vm.rdb.Eval(ctx, luaScript, []string{versionKey, key},
                              expectedVersion, valueJSON, int(ttl.Seconds())).Result()
    if err != nil {
        return false, err
    }

    return result.(int64) > 0, nil
}

// 创建缓存一致性管理器
func NewCacheConsistencyManager(rdb *redis.Client, db *gorm.DB, strategy ConsistencyStrategy) *CacheConsistencyManager {
    manager := &CacheConsistencyManager{
        rdb:            rdb,
        db:             db,
        strategy:       strategy,
        writeQueue:     make(chan *WriteTask, 1000),
        versionManager: NewVersionManager(rdb),
    }

    // 启动写后延迟处理器
    if strategy == WriteBehind {
        go manager.processBehindWrites()
    }

    return manager
}

// 1. Cache-Aside策略实现
func (m *CacheConsistencyManager) CacheAsideGet(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
    // 先从缓存获取
    cached, err := m.rdb.Get(ctx, key).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }

    // 缓存未命中，使用分布式锁防止缓存击穿
    lockKey := fmt.Sprintf("lock:%s", key)
    lock := NewDistributedLock(m.rdb, lockKey, 10*time.Second)

    if err := lock.TryAcquire(ctx, 50*time.Millisecond, 5); err != nil {
        // 获取锁失败，等待后重试
        time.Sleep(100 * time.Millisecond)
        cached, err := m.rdb.Get(ctx, key).Result()
        if err == nil {
            var result interface{}
            if err := json.Unmarshal([]byte(cached), &result); err == nil {
                return result, nil
            }
        }
        return nil, fmt.Errorf("获取数据失败: %w", err)
    }
    defer lock.Release(ctx)

    // 双重检查
    cached, err = m.rdb.Get(ctx, key).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(cached), &result); err == nil {
            return result, nil
        }
    }

    // 从数据源加载
    data, err := loader()
    if err != nil {
        return nil, err
    }

    // 写入缓存
    dataJSON, _ := json.Marshal(data)
    m.rdb.Set(ctx, key, dataJSON, time.Hour)

    return data, nil
}

func (m *CacheConsistencyManager) CacheAsideUpdate(ctx context.Context, key string, updater func() error) error {
    // 先更新数据库
    if err := updater(); err != nil {
        return err
    }

    // 删除缓存
    m.rdb.Del(ctx, key)

    // 增加版本号
    m.versionManager.IncrVersion(ctx, key)

    return nil
}

// 2. Write-Through策略实现
func (m *CacheConsistencyManager) WriteThroughSet(ctx context.Context, key string, value interface{}, ttl time.Duration, dbUpdater func() error) error {
    // 同时更新缓存和数据库
    errChan := make(chan error, 2)

    // 并发更新缓存
    go func() {
        valueJSON, _ := json.Marshal(value)
        errChan <- m.rdb.Set(ctx, key, valueJSON, ttl).Err()
    }()

    // 并发更新数据库
    go func() {
        errChan <- dbUpdater()
    }()

    // 等待两个操作完成
    var errors []error
    for i := 0; i < 2; i++ {
        if err := <-errChan; err != nil {
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        // 如果有错误，回滚操作
        m.rdb.Del(ctx, key)
        return fmt.Errorf("write-through failed: %v", errors)
    }

    return nil
}

// 3. Write-Behind策略实现
func (m *CacheConsistencyManager) WriteBehindSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // 立即更新缓存
    valueJSON, _ := json.Marshal(value)
    if err := m.rdb.Set(ctx, key, valueJSON, ttl).Err(); err != nil {
        return err
    }

    // 异步更新数据库
    task := &WriteTask{
        Key:   key,
        Value: value,
        TTL:   ttl,
    }

    select {
    case m.writeQueue <- task:
        return nil
    default:
        return fmt.Errorf("write queue is full")
    }
}

func (m *CacheConsistencyManager) processBehindWrites() {
    for task := range m.writeQueue {
        // 批量处理写任务
        m.processBatchWrites([]*WriteTask{task})
    }
}

func (m *CacheConsistencyManager) processBatchWrites(tasks []*WriteTask) {
    // 批量更新数据库
    for _, task := range tasks {
        // 这里应该调用实际的数据库更新逻辑
        fmt.Printf("异步更新数据库: %s\n", task.Key)

        // 模拟数据库更新
        time.Sleep(10 * time.Millisecond)
    }
}

// 4. Refresh-Ahead策略实现
func (m *CacheConsistencyManager) RefreshAheadGet(ctx context.Context, key string, loader func() (interface{}, error), refreshThreshold time.Duration) (interface{}, error) {
    // 获取缓存数据和TTL
    pipe := m.rdb.Pipeline()
    getCmd := pipe.Get(ctx, key)
    ttlCmd := pipe.TTL(ctx, key)
    _, err := pipe.Exec(ctx)

    if err != nil && err != redis.Nil {
        return nil, err
    }

    cached, err := getCmd.Result()
    if err == redis.Nil {
        // 缓存不存在，直接加载
        return m.loadAndCache(ctx, key, loader)
    }

    // 检查是否需要刷新
    ttl, _ := ttlCmd.Result()
    if ttl > 0 && ttl < refreshThreshold {
        // 异步刷新缓存
        go m.refreshCache(context.Background(), key, loader)
    }

    // 返回当前缓存数据
    var result interface{}
    json.Unmarshal([]byte(cached), &result)
    return result, nil
}

func (m *CacheConsistencyManager) refreshCache(ctx context.Context, key string, loader func() (interface{}, error)) {
    // 防止重复刷新
    if _, exists := m.refreshTasks.LoadOrStore(key, true); exists {
        return
    }
    defer m.refreshTasks.Delete(key)

    // 加载新数据
    data, err := loader()
    if err != nil {
        fmt.Printf("刷新缓存失败: %s, error: %v\n", key, err)
        return
    }

    // 更新缓存
    dataJSON, _ := json.Marshal(data)
    m.rdb.Set(ctx, key, dataJSON, time.Hour)

    fmt.Printf("缓存刷新完成: %s\n", key)
}

func (m *CacheConsistencyManager) loadAndCache(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
    data, err := loader()
    if err != nil {
        return nil, err
    }

    dataJSON, _ := json.Marshal(data)
    m.rdb.Set(ctx, key, dataJSON, time.Hour)

    return data, nil
}

// 5. 测试函数
func TestCacheConsistency() {
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    defer rdb.Close()

    // 模拟数据库
    var db *gorm.DB // 实际项目中应该是真实的数据库连接

    ctx := context.Background()

    // 测试Cache-Aside策略
    fmt.Println("=== Cache-Aside策略测试 ===")
    manager := NewCacheConsistencyManager(rdb, db, CacheAside)

    // 模拟数据加载器
    loader := func() (interface{}, error) {
        fmt.Println("从数据库加载数据...")
        return map[string]interface{}{
            "id":   1,
            "name": "测试用户",
            "age":  25,
        }, nil
    }

    // 第一次获取（缓存未命中）
    data1, _ := manager.CacheAsideGet(ctx, "user:1", loader)
    fmt.Printf("第一次获取: %+v\n", data1)

    // 第二次获取（缓存命中）
    data2, _ := manager.CacheAsideGet(ctx, "user:1", loader)
    fmt.Printf("第二次获取: %+v\n", data2)

    // 更新数据
    updater := func() error {
        fmt.Println("更新数据库...")
        return nil
    }
    manager.CacheAsideUpdate(ctx, "user:1", updater)

    // 再次获取（缓存已失效）
    data3, _ := manager.CacheAsideGet(ctx, "user:1", loader)
    fmt.Printf("更新后获取: %+v\n", data3)
}

/*
解析说明：
1. Cache-Aside：应用程序管理缓存，适合读多写少场景
2. Write-Through：同步更新缓存和数据库，保证强一致性
3. Write-Behind：异步更新数据库，提高写性能
4. Refresh-Ahead：主动刷新即将过期的缓存，减少缓存未命中
5. 使用版本控制和分布式锁解决并发问题

扩展思考：
- 如何处理缓存和数据库的事务一致性？
- 如何实现多级缓存的一致性？
- 如何监控缓存一致性的健康状态？
- 如何处理网络分区时的一致性问题？
*/
```

---

## 📚 章节总结

### 🎯 本章学习成果

通过本章的学习，你已经掌握了：

#### 📖 理论知识
- **Redis核心概念**：数据类型、持久化机制、集群架构
- **缓存策略设计**：Cache-Aside、Write-Through、Write-Behind等模式
- **分布式系统理论**：CAP定理、一致性模型、分布式锁原理
- **性能优化理论**：内存管理、网络优化、并发控制

#### 🛠️ 实践技能
- **Redis数据操作**：String、Hash、List、Set、ZSet的高级用法
- **缓存问题解决**：穿透、击穿、雪崩的完整解决方案
- **分布式锁实现**：基于Redis的各种锁机制和应用场景
- **消息队列应用**：List队列、Pub/Sub、延迟队列、优先级队列
- **性能优化实践**：连接池配置、管道操作、内存优化

#### 🏗️ 架构能力
- **缓存架构设计**：多级缓存、缓存预热、过期策略
- **高可用方案**：主从复制、哨兵模式、集群部署
- **监控体系建设**：性能监控、内存分析、慢查询优化
- **企业级实践**：电商缓存架构、秒杀系统、会话管理

### 🆚 与其他技术方案对比总结

| 特性 | Redis | Memcached | Ehcache | Hazelcast |
|------|-------|-----------|---------|-----------|
| **数据类型** | 丰富(5+种) | 仅String | 对象缓存 | 丰富数据结构 |
| **持久化** | RDB+AOF | 无 | 可选 | 可选 |
| **分布式** | 原生支持 | 客户端分片 | 分布式缓存 | 原生分布式 |
| **性能** | 极高 | 极高 | 高(JVM内) | 高 |
| **功能丰富度** | 非常丰富 | 基础 | 丰富 | 丰富 |
| **学习成本** | 中等 | 低 | 中等 | 高 |
| **生态成熟度** | 非常成熟 | 成熟 | 成熟 | 较成熟 |

### 🎯 面试准备要点

#### 核心概念掌握
- Redis数据类型的底层实现和应用场景
- 缓存策略的选择和实现原理
- 分布式锁的实现方式和注意事项
- Redis持久化机制的优缺点对比

#### 实践经验展示
- 大型项目中的Redis架构设计经验
- 缓存问题的诊断和解决实践
- 高并发场景下的性能优化案例
- 分布式系统中的一致性保证方案

#### 问题解决能力
- 常见Redis问题的排查思路
- 内存优化和性能调优经验
- 集群运维和故障处理能力
- 监控体系的建设和维护

### 🚀 下一步学习建议

#### 深入学习方向
1. **Redis源码分析**
   - 数据结构实现原理
   - 事件驱动模型
   - 内存管理机制
   - 集群通信协议

2. **高级特性探索**
   - Redis Modules开发
   - Lua脚本高级应用
   - Stream消息流处理
   - 地理位置应用

3. **企业级运维**
   - 监控告警体系
   - 自动化运维
   - 容量规划
   - 灾备恢复

#### 实践项目建议
1. **个人项目**：构建一个完整的缓存中间件
2. **开源贡献**：参与Redis相关开源项目
3. **企业实践**：在生产环境中应用所学知识

### 💡 学习心得

Redis作为现代互联网架构中不可或缺的组件，不仅仅是一个简单的缓存工具，更是一个功能强大的数据结构服务器。通过本章的学习，我们不仅掌握了Redis的使用技巧，更重要的是培养了分布式系统的设计思维。

在实际应用中，要始终记住：
- **合适优于完美**：选择最适合业务场景的方案
- **简单优于复杂**：避免过度设计和不必要的复杂性
- **监控优于猜测**：建立完善的监控体系
- **实践优于理论**：通过实际项目验证所学知识

### 🎉 恭喜完成

恭喜你完成了Redis缓存应用与实践的学习！你现在已经具备了：

✅ **扎实的理论基础** - 深入理解Redis原理和缓存设计
✅ **丰富的实践经验** - 掌握各种复杂场景的解决方案
✅ **优秀的架构能力** - 能够设计高性能、高可用的缓存系统
✅ **完善的面试准备** - 具备回答各种Redis相关问题的能力

继续保持学习的热情，在Go语言的道路上不断前进！下一章我们将学习消息队列集成，进一步提升系统的可扩展性和可靠性。

---

*"缓存是性能优化的银弹，Redis是缓存界的王者。掌握了Redis，你就掌握了现代互联网架构的核心技能！"* 🚀✨
