# 实战篇第四章：消息队列集成与实践 🚀

> *"在分布式系统中，消息队列是解耦服务、提高系统可靠性的核心组件。掌握消息队列，就掌握了构建高可用系统的关键技能！"* 💪

## 📚 本章学习目标

通过本章学习，你将掌握：

- 🎯 **消息队列核心概念**：理解消息队列的作用、优势和应用场景
- 🛠️ **主流MQ框架对比**：RabbitMQ vs Kafka vs NSQ vs Redis Stream
- 🏗️ **Go语言集成实践**：在Go项目中集成和使用各种消息队列
- 📨 **消息模式设计**：发布订阅、点对点、请求响应等消息模式
- ⚡ **高级特性应用**：消息持久化、事务、死信队列、延迟消息
- 🎪 **事件驱动架构**：基于消息队列构建事件驱动的微服务架构
- 🔧 **性能优化技巧**：消息队列的性能调优和监控
- 🏢 **企业级实践**：结合mall-go项目的真实业务场景

---

## 🌟 消息队列概述

### 什么是消息队列？

消息队列（Message Queue，简称MQ）是一种应用程序间的通信方法，通过在消息的传输过程中保存消息来实现应用程序间的异步通信。

```go
// 消息队列的基本概念示例
type Message struct {
    ID        string                 `json:"id"`
    Topic     string                 `json:"topic"`
    Payload   map[string]interface{} `json:"payload"`
    Timestamp int64                  `json:"timestamp"`
    Headers   map[string]string      `json:"headers"`
}

// 生产者接口
type Producer interface {
    Send(ctx context.Context, topic string, message *Message) error
    SendBatch(ctx context.Context, topic string, messages []*Message) error
    Close() error
}

// 消费者接口
type Consumer interface {
    Subscribe(ctx context.Context, topic string, handler MessageHandler) error
    Unsubscribe(topic string) error
    Close() error
}

// 消息处理器
type MessageHandler func(ctx context.Context, message *Message) error
```

### 消息队列的核心优势

#### 1. 系统解耦 🔗
```go
// ❌ 紧耦合的同步调用
func ProcessOrder(order *Order) error {
    // 直接调用各个服务
    if err := inventoryService.UpdateStock(order.Items); err != nil {
        return err
    }
    if err := paymentService.ProcessPayment(order.Payment); err != nil {
        return err
    }
    if err := notificationService.SendNotification(order.UserID); err != nil {
        return err
    }
    return nil
}

// ✅ 通过消息队列解耦
func ProcessOrderWithMQ(order *Order, producer Producer) error {
    // 发布订单创建事件
    event := &Message{
        ID:      generateID(),
        Topic:   "order.created",
        Payload: map[string]interface{}{
            "order_id": order.ID,
            "user_id":  order.UserID,
            "items":    order.Items,
            "amount":   order.Amount,
        },
        Timestamp: time.Now().Unix(),
    }
    
    return producer.Send(context.Background(), "order.created", event)
}
```

#### 2. 异步处理 ⚡
```go
// 异步处理订单相关业务
func SetupOrderEventHandlers(consumer Consumer) {
    // 库存服务监听订单事件
    consumer.Subscribe(context.Background(), "order.created", func(ctx context.Context, msg *Message) error {
        orderID := msg.Payload["order_id"].(string)
        items := msg.Payload["items"].([]interface{})
        
        // 异步更新库存
        return inventoryService.UpdateStockAsync(orderID, items)
    })
    
    // 支付服务监听订单事件
    consumer.Subscribe(context.Background(), "order.created", func(ctx context.Context, msg *Message) error {
        orderID := msg.Payload["order_id"].(string)
        amount := msg.Payload["amount"].(float64)
        
        // 异步处理支付
        return paymentService.ProcessPaymentAsync(orderID, amount)
    })
    
    // 通知服务监听订单事件
    consumer.Subscribe(context.Background(), "order.created", func(ctx context.Context, msg *Message) error {
        userID := msg.Payload["user_id"].(string)
        orderID := msg.Payload["order_id"].(string)
        
        // 异步发送通知
        return notificationService.SendNotificationAsync(userID, orderID)
    })
}
```

#### 3. 流量削峰 📈
```go
// 秒杀场景的流量削峰
type SeckillService struct {
    producer Producer
    redis    *redis.Client
}

func (s *SeckillService) HandleSeckillRequest(ctx context.Context, userID, productID string) error {
    // 快速响应用户请求
    requestID := generateRequestID()
    
    // 将秒杀请求放入队列
    message := &Message{
        ID:    requestID,
        Topic: "seckill.request",
        Payload: map[string]interface{}{
            "request_id": requestID,
            "user_id":    userID,
            "product_id": productID,
            "timestamp":  time.Now().Unix(),
        },
    }
    
    if err := s.producer.Send(ctx, "seckill.request", message); err != nil {
        return err
    }
    
    // 立即返回，告知用户请求已提交
    return nil
}

// 异步处理秒杀请求
func (s *SeckillService) ProcessSeckillRequests(consumer Consumer) {
    consumer.Subscribe(context.Background(), "seckill.request", func(ctx context.Context, msg *Message) error {
        userID := msg.Payload["user_id"].(string)
        productID := msg.Payload["product_id"].(string)
        
        // 检查库存并处理秒杀
        success, err := s.processSeckillLogic(ctx, userID, productID)
        if err != nil {
            return err
        }
        
        // 发布处理结果
        resultMsg := &Message{
            ID:    generateID(),
            Topic: "seckill.result",
            Payload: map[string]interface{}{
                "user_id":    userID,
                "product_id": productID,
                "success":    success,
                "timestamp":  time.Now().Unix(),
            },
        }
        
        return s.producer.Send(ctx, "seckill.result", resultMsg)
    })
}
```

### 消息队列应用场景

#### 1. 订单处理流程 📦
```go
// 订单处理的完整消息流
type OrderEventType string

const (
    OrderCreated   OrderEventType = "order.created"
    OrderPaid      OrderEventType = "order.paid"
    OrderShipped   OrderEventType = "order.shipped"
    OrderDelivered OrderEventType = "order.delivered"
    OrderCancelled OrderEventType = "order.cancelled"
)

// 订单状态机通过消息驱动
func (s *OrderService) CreateOrder(ctx context.Context, order *Order) error {
    // 1. 保存订单到数据库
    if err := s.db.Create(order).Error; err != nil {
        return err
    }
    
    // 2. 发布订单创建事件
    return s.publishOrderEvent(ctx, OrderCreated, order)
}

func (s *OrderService) publishOrderEvent(ctx context.Context, eventType OrderEventType, order *Order) error {
    message := &Message{
        ID:    generateID(),
        Topic: string(eventType),
        Payload: map[string]interface{}{
            "order_id":   order.ID,
            "user_id":    order.UserID,
            "status":     order.Status,
            "amount":     order.Amount,
            "items":      order.Items,
            "created_at": order.CreatedAt,
        },
        Headers: map[string]string{
            "event_type": string(eventType),
            "source":     "order-service",
        },
    }
    
    return s.producer.Send(ctx, string(eventType), message)
}
```

#### 2. 数据同步 🔄
```go
// 用户信息变更同步到各个服务
type UserSyncService struct {
    producer Producer
}

func (s *UserSyncService) UpdateUserProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
    // 1. 更新主数据库
    if err := s.updateUserInDB(ctx, userID, updates); err != nil {
        return err
    }
    
    // 2. 发布用户更新事件
    message := &Message{
        ID:    generateID(),
        Topic: "user.profile.updated",
        Payload: map[string]interface{}{
            "user_id": userID,
            "updates": updates,
            "version": time.Now().Unix(),
        },
    }
    
    return s.producer.Send(ctx, "user.profile.updated", message)
}

// 各个服务监听用户更新事件
func SetupUserSyncHandlers(consumer Consumer) {
    // 订单服务更新用户缓存
    consumer.Subscribe(context.Background(), "user.profile.updated", func(ctx context.Context, msg *Message) error {
        userID := msg.Payload["user_id"].(string)
        updates := msg.Payload["updates"].(map[string]interface{})
        
        return orderService.UpdateUserCache(ctx, userID, updates)
    })
    
    // 推荐服务更新用户画像
    consumer.Subscribe(context.Background(), "user.profile.updated", func(ctx context.Context, msg *Message) error {
        userID := msg.Payload["user_id"].(string)
        updates := msg.Payload["updates"].(map[string]interface{})
        
        return recommendService.UpdateUserProfile(ctx, userID, updates)
    })
}
```

#### 3. 日志收集与分析 📊
```go
// 业务日志收集
type LogCollector struct {
    producer Producer
}

func (lc *LogCollector) LogUserAction(ctx context.Context, action *UserAction) error {
    message := &Message{
        ID:    generateID(),
        Topic: "user.action.log",
        Payload: map[string]interface{}{
            "user_id":    action.UserID,
            "action":     action.Action,
            "resource":   action.Resource,
            "ip":         action.IP,
            "user_agent": action.UserAgent,
            "timestamp":  action.Timestamp,
        },
    }
    
    return lc.producer.Send(ctx, "user.action.log", message)
}

// 日志分析服务
func SetupLogAnalysis(consumer Consumer) {
    consumer.Subscribe(context.Background(), "user.action.log", func(ctx context.Context, msg *Message) error {
        // 实时分析用户行为
        return analyticsService.ProcessUserAction(ctx, msg.Payload)
    })
}
```

---

## 🆚 主流消息队列对比

### 技术选型对比表

| 特性 | RabbitMQ | Apache Kafka | NSQ | Redis Stream |
|------|----------|--------------|-----|--------------|
| **语言** | Erlang | Scala/Java | Go | C |
| **协议** | AMQP | 自定义 | HTTP/TCP | Redis协议 |
| **性能** | 中等 | 极高 | 高 | 高 |
| **可靠性** | 极高 | 高 | 高 | 中等 |
| **复杂度** | 高 | 高 | 低 | 低 |
| **运维成本** | 高 | 高 | 低 | 低 |
| **生态成熟度** | 非常成熟 | 非常成熟 | 较成熟 | 较新 |
| **适用场景** | 企业级应用 | 大数据流处理 | 简单消息队列 | 轻量级队列 |

### 详细对比分析

#### 1. RabbitMQ 🐰
```go
// RabbitMQ的特点和适用场景
/*
优势：
✅ 功能丰富：支持多种消息模式（发布订阅、路由、RPC等）
✅ 可靠性高：支持消息持久化、事务、确认机制
✅ 管理界面：提供Web管理界面，运维友好
✅ 插件生态：丰富的插件系统
✅ 标准协议：支持AMQP标准协议

劣势：
❌ 性能一般：相比Kafka性能较低
❌ 学习成本：概念较多，学习曲线陡峭
❌ 资源消耗：内存和CPU消耗较高
❌ 扩展性：集群扩展相对复杂

适用场景：
🎯 企业级应用，对可靠性要求极高
🎯 复杂的消息路由需求
🎯 需要事务支持的场景
🎯 传统企业架构升级
*/

// RabbitMQ Go客户端示例
import "github.com/streadway/amqp"

type RabbitMQClient struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    
    return &RabbitMQClient{
        conn:    conn,
        channel: ch,
    }, nil
}
```

#### 2. Apache Kafka ⚡
```go
// Kafka的特点和适用场景
/*
优势：
✅ 极高性能：百万级TPS，低延迟
✅ 水平扩展：天然支持分布式，易于扩展
✅ 持久化：数据持久化到磁盘，支持数据回放
✅ 流处理：与Kafka Streams集成，支持流处理
✅ 生态丰富：与大数据生态深度集成

劣势：
❌ 复杂度高：配置复杂，运维成本高
❌ 资源消耗：需要较多内存和磁盘空间
❌ 学习成本：概念较多，需要深入理解
❌ 实时性：不适合低延迟场景

适用场景：
🎯 大数据流处理
🎯 高吞吐量场景
🎯 日志收集和分析
🎯 事件溯源架构
*/

// Kafka Go客户端示例
import "github.com/Shopify/sarama"

type KafkaClient struct {
    producer sarama.SyncProducer
    consumer sarama.Consumer
}

func NewKafkaClient(brokers []string) (*KafkaClient, error) {
    config := sarama.NewConfig()
    config.Producer.Return.Successes = true
    
    producer, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, err
    }
    
    consumer, err := sarama.NewConsumer(brokers, config)
    if err != nil {
        return nil, err
    }
    
    return &KafkaClient{
        producer: producer,
        consumer: consumer,
    }, nil
}
```

#### 3. NSQ 🚀
```go
// NSQ的特点和适用场景
/*
优势：
✅ 简单易用：Go原生，配置简单
✅ 去中心化：无单点故障
✅ 高性能：纯Go实现，性能优秀
✅ 运维友好：内置Web管理界面
✅ 自动发现：支持服务自动发现

劣势：
❌ 功能相对简单：不支持复杂路由
❌ 生态较小：相比RabbitMQ和Kafka生态较小
❌ 持久化有限：持久化功能相对简单
❌ 顺序保证：不保证消息顺序

适用场景：
🎯 Go微服务架构
🎯 简单的发布订阅需求
🎯 对运维成本敏感的场景
🎯 中小型项目
*/

// NSQ Go客户端示例
import "github.com/nsqio/go-nsq"

type NSQClient struct {
    producer *nsq.Producer
    consumer *nsq.Consumer
}

func NewNSQClient(nsqdAddr string) (*NSQClient, error) {
    config := nsq.NewConfig()
    
    producer, err := nsq.NewProducer(nsqdAddr, config)
    if err != nil {
        return nil, err
    }
    
    return &NSQClient{
        producer: producer,
    }, nil
}
```

#### 4. Redis Stream 📡
```go
// Redis Stream的特点和适用场景
/*
优势：
✅ 轻量级：基于Redis，部署简单
✅ 高性能：内存存储，性能优秀
✅ 持久化：支持AOF和RDB持久化
✅ 消费组：支持消费者组模式
✅ 学习成本低：基于熟悉的Redis

劣势：
❌ 功能有限：相比专业MQ功能较少
❌ 内存限制：受Redis内存限制
❌ 生态较新：相对较新，生态不够成熟
❌ 集群复杂：Redis集群配置相对复杂

适用场景：
🎯 已有Redis基础设施
🎯 轻量级消息队列需求
🎯 实时数据流处理
🎯 简单的事件驱动架构
*/

// Redis Stream Go客户端示例
import "github.com/redis/go-redis/v9"

type RedisStreamClient struct {
    rdb *redis.Client
}

func NewRedisStreamClient(addr string) *RedisStreamClient {
    rdb := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    
    return &RedisStreamClient{rdb: rdb}
}
```

---

## 🛠️ RabbitMQ集成实践

RabbitMQ是功能最丰富的消息队列之一，让我们深入学习如何在Go项目中集成和使用RabbitMQ。

### 基础配置与连接

```go
// 来自 mall-go/internal/mq/rabbitmq/client.go
package rabbitmq

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/streadway/amqp"
)

// RabbitMQ配置
type Config struct {
    URL          string        `yaml:"url"`
    Exchange     string        `yaml:"exchange"`
    ExchangeType string        `yaml:"exchange_type"`
    Durable      bool          `yaml:"durable"`
    AutoDelete   bool          `yaml:"auto_delete"`
    Internal     bool          `yaml:"internal"`
    NoWait       bool          `yaml:"no_wait"`
    
    // 连接池配置
    MaxConnections int           `yaml:"max_connections"`
    MaxChannels    int           `yaml:"max_channels"`
    HeartBeat      time.Duration `yaml:"heartbeat"`
    
    // 重连配置
    ReconnectDelay time.Duration `yaml:"reconnect_delay"`
    MaxRetries     int           `yaml:"max_retries"`
}

// RabbitMQ客户端
type Client struct {
    config   *Config
    conn     *amqp.Connection
    channels chan *amqp.Channel
    mutex    sync.RWMutex
    closed   bool
}

// 创建RabbitMQ客户端
func NewClient(config *Config) (*Client, error) {
    client := &Client{
        config:   config,
        channels: make(chan *amqp.Channel, config.MaxChannels),
    }
    
    if err := client.connect(); err != nil {
        return nil, err
    }
    
    // 预创建通道池
    for i := 0; i < config.MaxChannels; i++ {
        ch, err := client.conn.Channel()
        if err != nil {
            return nil, err
        }
        client.channels <- ch
    }
    
    // 监听连接状态
    go client.handleConnectionEvents()
    
    return client, nil
}

// 建立连接
func (c *Client) connect() error {
    var err error
    
    // 配置连接参数
    config := amqp.Config{
        Heartbeat: c.config.HeartBeat,
        Locale:    "en_US",
    }
    
    c.conn, err = amqp.DialConfig(c.config.URL, config)
    if err != nil {
        return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }
    
    log.Printf("Connected to RabbitMQ: %s", c.config.URL)
    return nil
}

// 获取通道
func (c *Client) getChannel() (*amqp.Channel, error) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if c.closed {
        return nil, fmt.Errorf("client is closed")
    }
    
    select {
    case ch := <-c.channels:
        return ch, nil
    case <-time.After(5 * time.Second):
        return nil, fmt.Errorf("timeout getting channel")
    }
}

// 归还通道
func (c *Client) returnChannel(ch *amqp.Channel) {
    if ch == nil || ch.IsClosed() {
        // 创建新通道替换
        newCh, err := c.conn.Channel()
        if err != nil {
            log.Printf("Failed to create new channel: %v", err)
            return
        }
        ch = newCh
    }
    
    select {
    case c.channels <- ch:
    default:
        ch.Close() // 通道池满了，关闭通道
    }
}

// 处理连接事件
func (c *Client) handleConnectionEvents() {
    for {
        reason, ok := <-c.conn.NotifyClose(make(chan *amqp.Error))
        if !ok {
            break
        }
        
        log.Printf("Connection closed: %v", reason)
        
        // 尝试重连
        for i := 0; i < c.config.MaxRetries; i++ {
            time.Sleep(c.config.ReconnectDelay)
            
            if err := c.connect(); err != nil {
                log.Printf("Reconnect attempt %d failed: %v", i+1, err)
                continue
            }
            
            log.Printf("Reconnected to RabbitMQ")
            break
        }
    }
}

// 关闭客户端
func (c *Client) Close() error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if c.closed {
        return nil
    }
    
    c.closed = true
    
    // 关闭所有通道
    close(c.channels)
    for ch := range c.channels {
        ch.Close()
    }
    
    // 关闭连接
    return c.conn.Close()
}
```

### RabbitMQ生产者实现

```go
// 来自 mall-go/internal/mq/rabbitmq/producer.go
package rabbitmq

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/streadway/amqp"
)

// 生产者
type Producer struct {
    client   *Client
    exchange string
}

// 创建生产者
func NewProducer(client *Client, exchange string) (*Producer, error) {
    producer := &Producer{
        client:   client,
        exchange: exchange,
    }

    // 声明交换机
    if err := producer.declareExchange(); err != nil {
        return nil, err
    }

    return producer, nil
}

// 声明交换机
func (p *Producer) declareExchange() error {
    ch, err := p.client.getChannel()
    if err != nil {
        return err
    }
    defer p.client.returnChannel(ch)

    return ch.ExchangeDeclare(
        p.exchange,                    // 交换机名称
        p.client.config.ExchangeType,  // 交换机类型
        p.client.config.Durable,       // 是否持久化
        p.client.config.AutoDelete,    // 是否自动删除
        p.client.config.Internal,      // 是否内部使用
        p.client.config.NoWait,        // 是否等待服务器响应
        nil,                           // 额外参数
    )
}

// 发送消息
func (p *Producer) Send(ctx context.Context, routingKey string, message interface{}) error {
    return p.SendWithOptions(ctx, routingKey, message, PublishOptions{})
}

// 发送选项
type PublishOptions struct {
    ContentType  string
    DeliveryMode uint8  // 1=非持久化, 2=持久化
    Priority     uint8  // 0-255
    Expiration   string // 消息过期时间(毫秒)
    Headers      amqp.Table
    Mandatory    bool   // 如果无法路由是否返回
    Immediate    bool   // 如果无消费者是否返回
}

// 带选项发送消息
func (p *Producer) SendWithOptions(ctx context.Context, routingKey string, message interface{}, options PublishOptions) error {
    ch, err := p.client.getChannel()
    if err != nil {
        return err
    }
    defer p.client.returnChannel(ch)

    // 序列化消息
    body, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    // 设置默认值
    if options.ContentType == "" {
        options.ContentType = "application/json"
    }
    if options.DeliveryMode == 0 {
        options.DeliveryMode = 2 // 默认持久化
    }

    // 构建发布消息
    publishing := amqp.Publishing{
        ContentType:  options.ContentType,
        DeliveryMode: options.DeliveryMode,
        Priority:     options.Priority,
        Expiration:   options.Expiration,
        Headers:      options.Headers,
        Timestamp:    time.Now(),
        Body:         body,
    }

    // 发布消息
    return ch.Publish(
        p.exchange,         // 交换机
        routingKey,         // 路由键
        options.Mandatory,  // mandatory
        options.Immediate,  // immediate
        publishing,         // 消息
    )
}

// 批量发送消息
func (p *Producer) SendBatch(ctx context.Context, routingKey string, messages []interface{}) error {
    ch, err := p.client.getChannel()
    if err != nil {
        return err
    }
    defer p.client.returnChannel(ch)

    for _, message := range messages {
        body, err := json.Marshal(message)
        if err != nil {
            return fmt.Errorf("failed to marshal message: %w", err)
        }

        publishing := amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: 2, // 持久化
            Timestamp:    time.Now(),
            Body:         body,
        }

        if err := ch.Publish(p.exchange, routingKey, false, false, publishing); err != nil {
            return err
        }
    }

    return nil
}

// 发送延迟消息
func (p *Producer) SendDelayed(ctx context.Context, routingKey string, message interface{}, delay time.Duration) error {
    options := PublishOptions{
        Expiration: fmt.Sprintf("%d", delay.Milliseconds()),
        Headers: amqp.Table{
            "x-delay": delay.Milliseconds(),
        },
    }

    return p.SendWithOptions(ctx, routingKey, message, options)
}

// 发送事务消息
func (p *Producer) SendTransactional(ctx context.Context, routingKey string, messages []interface{}) error {
    ch, err := p.client.getChannel()
    if err != nil {
        return err
    }
    defer p.client.returnChannel(ch)

    // 开启事务
    if err := ch.Tx(); err != nil {
        return err
    }

    // 发送消息
    for _, message := range messages {
        body, err := json.Marshal(message)
        if err != nil {
            ch.TxRollback() // 回滚事务
            return fmt.Errorf("failed to marshal message: %w", err)
        }

        publishing := amqp.Publishing{
            ContentType:  "application/json",
            DeliveryMode: 2,
            Timestamp:    time.Now(),
            Body:         body,
        }

        if err := ch.Publish(p.exchange, routingKey, false, false, publishing); err != nil {
            ch.TxRollback() // 回滚事务
            return err
        }
    }

    // 提交事务
    return ch.TxCommit()
}
```

### RabbitMQ消费者实现

```go
// 来自 mall-go/internal/mq/rabbitmq/consumer.go
package rabbitmq

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/streadway/amqp"
)

// 消息处理器
type MessageHandler func(ctx context.Context, delivery amqp.Delivery) error

// 消费者配置
type ConsumerConfig struct {
    QueueName    string `yaml:"queue_name"`
    RoutingKey   string `yaml:"routing_key"`
    ConsumerTag  string `yaml:"consumer_tag"`
    AutoAck      bool   `yaml:"auto_ack"`
    Exclusive    bool   `yaml:"exclusive"`
    NoLocal      bool   `yaml:"no_local"`
    NoWait       bool   `yaml:"no_wait"`

    // 队列配置
    QueueDurable    bool `yaml:"queue_durable"`
    QueueAutoDelete bool `yaml:"queue_auto_delete"`
    QueueExclusive  bool `yaml:"queue_exclusive"`

    // 消费者配置
    PrefetchCount int `yaml:"prefetch_count"`
    PrefetchSize  int `yaml:"prefetch_size"`

    // 重试配置
    MaxRetries    int           `yaml:"max_retries"`
    RetryDelay    time.Duration `yaml:"retry_delay"`
    DeadLetterExchange string   `yaml:"dead_letter_exchange"`
}

// 消费者
type Consumer struct {
    client   *Client
    config   *ConsumerConfig
    exchange string
    handlers map[string]MessageHandler
    mutex    sync.RWMutex
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
}

// 创建消费者
func NewConsumer(client *Client, exchange string, config *ConsumerConfig) (*Consumer, error) {
    ctx, cancel := context.WithCancel(context.Background())

    consumer := &Consumer{
        client:   client,
        config:   config,
        exchange: exchange,
        handlers: make(map[string]MessageHandler),
        ctx:      ctx,
        cancel:   cancel,
    }

    // 声明队列
    if err := consumer.declareQueue(); err != nil {
        return nil, err
    }

    return consumer, nil
}

// 声明队列
func (c *Consumer) declareQueue() error {
    ch, err := c.client.getChannel()
    if err != nil {
        return err
    }
    defer c.client.returnChannel(ch)

    // 声明队列
    args := amqp.Table{}
    if c.config.DeadLetterExchange != "" {
        args["x-dead-letter-exchange"] = c.config.DeadLetterExchange
    }

    _, err = ch.QueueDeclare(
        c.config.QueueName,        // 队列名称
        c.config.QueueDurable,     // 是否持久化
        c.config.QueueAutoDelete,  // 是否自动删除
        c.config.QueueExclusive,   // 是否排他
        c.config.NoWait,           // 是否等待
        args,                      // 额外参数
    )
    if err != nil {
        return err
    }

    // 绑定队列到交换机
    return ch.QueueBind(
        c.config.QueueName, // 队列名称
        c.config.RoutingKey, // 路由键
        c.exchange,         // 交换机
        c.config.NoWait,    // 是否等待
        nil,                // 额外参数
    )
}

// 注册消息处理器
func (c *Consumer) RegisterHandler(routingKey string, handler MessageHandler) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.handlers[routingKey] = handler
}

// 开始消费
func (c *Consumer) Start() error {
    ch, err := c.client.getChannel()
    if err != nil {
        return err
    }

    // 设置QoS
    if err := ch.Qos(c.config.PrefetchCount, c.config.PrefetchSize, false); err != nil {
        return err
    }

    // 开始消费
    deliveries, err := ch.Consume(
        c.config.QueueName,   // 队列名称
        c.config.ConsumerTag, // 消费者标签
        c.config.AutoAck,     // 自动确认
        c.config.Exclusive,   // 排他
        c.config.NoLocal,     // 不接收自己发布的消息
        c.config.NoWait,      // 不等待
        nil,                  // 额外参数
    )
    if err != nil {
        return err
    }

    // 启动消息处理协程
    c.wg.Add(1)
    go c.handleMessages(deliveries)

    log.Printf("Consumer started for queue: %s", c.config.QueueName)
    return nil
}

// 处理消息
func (c *Consumer) handleMessages(deliveries <-chan amqp.Delivery) {
    defer c.wg.Done()

    for {
        select {
        case <-c.ctx.Done():
            return
        case delivery, ok := <-deliveries:
            if !ok {
                return
            }

            c.processMessage(delivery)
        }
    }
}

// 处理单个消息
func (c *Consumer) processMessage(delivery amqp.Delivery) {
    c.mutex.RLock()
    handler, exists := c.handlers[delivery.RoutingKey]
    c.mutex.RUnlock()

    if !exists {
        log.Printf("No handler for routing key: %s", delivery.RoutingKey)
        delivery.Nack(false, false) // 拒绝消息，不重新入队
        return
    }

    // 处理消息
    err := c.processWithRetry(handler, delivery)

    if err != nil {
        log.Printf("Failed to process message after retries: %v", err)
        delivery.Nack(false, false) // 拒绝消息，发送到死信队列
    } else {
        delivery.Ack(false) // 确认消息
    }
}

// 带重试的消息处理
func (c *Consumer) processWithRetry(handler MessageHandler, delivery amqp.Delivery) error {
    var err error

    for i := 0; i <= c.config.MaxRetries; i++ {
        err = handler(c.ctx, delivery)
        if err == nil {
            return nil
        }

        if i < c.config.MaxRetries {
            log.Printf("Message processing failed (attempt %d/%d): %v",
                      i+1, c.config.MaxRetries+1, err)
            time.Sleep(c.config.RetryDelay)
        }
    }

    return err
}

// 停止消费
func (c *Consumer) Stop() error {
    c.cancel()
    c.wg.Wait()

    log.Printf("Consumer stopped for queue: %s", c.config.QueueName)
    return nil
}
```

### RabbitMQ实际应用示例

```go
// 来自 mall-go/internal/service/order_message_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "mall-go/internal/mq/rabbitmq"
    "mall-go/internal/model"
)

// 订单消息服务
type OrderMessageService struct {
    producer *rabbitmq.Producer
    consumer *rabbitmq.Consumer
}

// 订单事件类型
type OrderEvent struct {
    EventType string                 `json:"event_type"`
    OrderID   uint                   `json:"order_id"`
    UserID    uint                   `json:"user_id"`
    Data      map[string]interface{} `json:"data"`
    Timestamp int64                  `json:"timestamp"`
}

// 创建订单消息服务
func NewOrderMessageService(client *rabbitmq.Client) (*OrderMessageService, error) {
    // 创建生产者
    producer, err := rabbitmq.NewProducer(client, "mall.orders")
    if err != nil {
        return nil, err
    }

    // 创建消费者配置
    consumerConfig := &rabbitmq.ConsumerConfig{
        QueueName:       "order.processing",
        RoutingKey:      "order.*",
        ConsumerTag:     "order-processor",
        AutoAck:         false,
        PrefetchCount:   10,
        MaxRetries:      3,
        RetryDelay:      time.Second * 5,
        DeadLetterExchange: "mall.orders.dlx",
    }

    // 创建消费者
    consumer, err := rabbitmq.NewConsumer(client, "mall.orders", consumerConfig)
    if err != nil {
        return nil, err
    }

    service := &OrderMessageService{
        producer: producer,
        consumer: consumer,
    }

    // 注册消息处理器
    service.registerHandlers()

    return service, nil
}

// 注册消息处理器
func (s *OrderMessageService) registerHandlers() {
    // 订单创建处理器
    s.consumer.RegisterHandler("order.created", s.handleOrderCreated)

    // 订单支付处理器
    s.consumer.RegisterHandler("order.paid", s.handleOrderPaid)

    // 订单发货处理器
    s.consumer.RegisterHandler("order.shipped", s.handleOrderShipped)

    // 订单取消处理器
    s.consumer.RegisterHandler("order.cancelled", s.handleOrderCancelled)
}

// 发布订单创建事件
func (s *OrderMessageService) PublishOrderCreated(ctx context.Context, order *model.Order) error {
    event := &OrderEvent{
        EventType: "order.created",
        OrderID:   order.ID,
        UserID:    order.UserID,
        Data: map[string]interface{}{
            "amount":     order.Amount,
            "items":      order.Items,
            "address":    order.Address,
            "created_at": order.CreatedAt,
        },
        Timestamp: time.Now().Unix(),
    }

    return s.producer.Send(ctx, "order.created", event)
}

// 发布订单支付事件
func (s *OrderMessageService) PublishOrderPaid(ctx context.Context, orderID uint, paymentInfo map[string]interface{}) error {
    event := &OrderEvent{
        EventType: "order.paid",
        OrderID:   orderID,
        Data:      paymentInfo,
        Timestamp: time.Now().Unix(),
    }

    return s.producer.Send(ctx, "order.paid", event)
}

// 处理订单创建事件
func (s *OrderMessageService) handleOrderCreated(ctx context.Context, delivery amqp.Delivery) error {
    var event OrderEvent
    if err := json.Unmarshal(delivery.Body, &event); err != nil {
        return fmt.Errorf("failed to unmarshal order created event: %w", err)
    }

    log.Printf("Processing order created event: OrderID=%d, UserID=%d",
               event.OrderID, event.UserID)

    // 1. 更新库存
    if err := s.updateInventory(ctx, event); err != nil {
        return fmt.Errorf("failed to update inventory: %w", err)
    }

    // 2. 发送通知
    if err := s.sendOrderNotification(ctx, event); err != nil {
        log.Printf("Failed to send notification: %v", err)
        // 通知失败不影响主流程
    }

    // 3. 更新用户积分
    if err := s.updateUserPoints(ctx, event); err != nil {
        log.Printf("Failed to update user points: %v", err)
        // 积分更新失败不影响主流程
    }

    log.Printf("Order created event processed successfully: OrderID=%d", event.OrderID)
    return nil
}

// 处理订单支付事件
func (s *OrderMessageService) handleOrderPaid(ctx context.Context, delivery amqp.Delivery) error {
    var event OrderEvent
    if err := json.Unmarshal(delivery.Body, &event); err != nil {
        return fmt.Errorf("failed to unmarshal order paid event: %w", err)
    }

    log.Printf("Processing order paid event: OrderID=%d", event.OrderID)

    // 1. 更新订单状态
    if err := s.updateOrderStatus(ctx, event.OrderID, "paid"); err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    // 2. 触发发货流程
    if err := s.triggerShipping(ctx, event); err != nil {
        return fmt.Errorf("failed to trigger shipping: %w", err)
    }

    // 3. 发送支付成功通知
    if err := s.sendPaymentNotification(ctx, event); err != nil {
        log.Printf("Failed to send payment notification: %v", err)
    }

    log.Printf("Order paid event processed successfully: OrderID=%d", event.OrderID)
    return nil
}

// 处理订单发货事件
func (s *OrderMessageService) handleOrderShipped(ctx context.Context, delivery amqp.Delivery) error {
    var event OrderEvent
    if err := json.Unmarshal(delivery.Body, &event); err != nil {
        return fmt.Errorf("failed to unmarshal order shipped event: %w", err)
    }

    log.Printf("Processing order shipped event: OrderID=%d", event.OrderID)

    // 1. 更新订单状态
    if err := s.updateOrderStatus(ctx, event.OrderID, "shipped"); err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    // 2. 发送发货通知
    if err := s.sendShippingNotification(ctx, event); err != nil {
        log.Printf("Failed to send shipping notification: %v", err)
    }

    // 3. 启动物流跟踪
    if err := s.startLogisticsTracking(ctx, event); err != nil {
        log.Printf("Failed to start logistics tracking: %v", err)
    }

    log.Printf("Order shipped event processed successfully: OrderID=%d", event.OrderID)
    return nil
}

// 处理订单取消事件
func (s *OrderMessageService) handleOrderCancelled(ctx context.Context, delivery amqp.Delivery) error {
    var event OrderEvent
    if err := json.Unmarshal(delivery.Body, &event); err != nil {
        return fmt.Errorf("failed to unmarshal order cancelled event: %w", err)
    }

    log.Printf("Processing order cancelled event: OrderID=%d", event.OrderID)

    // 1. 恢复库存
    if err := s.restoreInventory(ctx, event); err != nil {
        return fmt.Errorf("failed to restore inventory: %w", err)
    }

    // 2. 处理退款
    if err := s.processRefund(ctx, event); err != nil {
        return fmt.Errorf("failed to process refund: %w", err)
    }

    // 3. 发送取消通知
    if err := s.sendCancellationNotification(ctx, event); err != nil {
        log.Printf("Failed to send cancellation notification: %v", err)
    }

    log.Printf("Order cancelled event processed successfully: OrderID=%d", event.OrderID)
    return nil
}

// 辅助方法实现
func (s *OrderMessageService) updateInventory(ctx context.Context, event OrderEvent) error {
    // 模拟库存更新
    log.Printf("Updating inventory for order: %d", event.OrderID)
    time.Sleep(100 * time.Millisecond) // 模拟处理时间
    return nil
}

func (s *OrderMessageService) sendOrderNotification(ctx context.Context, event OrderEvent) error {
    // 模拟发送通知
    log.Printf("Sending order notification for user: %d", event.UserID)
    return nil
}

func (s *OrderMessageService) updateUserPoints(ctx context.Context, event OrderEvent) error {
    // 模拟积分更新
    log.Printf("Updating user points for user: %d", event.UserID)
    return nil
}

func (s *OrderMessageService) updateOrderStatus(ctx context.Context, orderID uint, status string) error {
    // 模拟订单状态更新
    log.Printf("Updating order %d status to: %s", orderID, status)
    return nil
}

func (s *OrderMessageService) triggerShipping(ctx context.Context, event OrderEvent) error {
    // 模拟触发发货
    log.Printf("Triggering shipping for order: %d", event.OrderID)
    return nil
}

func (s *OrderMessageService) sendPaymentNotification(ctx context.Context, event OrderEvent) error {
    // 模拟发送支付通知
    log.Printf("Sending payment notification for order: %d", event.OrderID)
    return nil
}

func (s *OrderMessageService) sendShippingNotification(ctx context.Context, event OrderEvent) error {
    // 模拟发送发货通知
    log.Printf("Sending shipping notification for order: %d", event.OrderID)
    return nil
}

func (s *OrderMessageService) startLogisticsTracking(ctx context.Context, event OrderEvent) error {
    // 模拟启动物流跟踪
    log.Printf("Starting logistics tracking for order: %d", event.OrderID)
    return nil
}

func (s *OrderMessageService) restoreInventory(ctx context.Context, event OrderEvent) error {
    // 模拟恢复库存
    log.Printf("Restoring inventory for order: %d", event.OrderID)
    return nil
}

func (s *OrderMessageService) processRefund(ctx context.Context, event OrderEvent) error {
    // 模拟处理退款
    log.Printf("Processing refund for order: %d", event.OrderID)
    return nil
}

func (s *OrderMessageService) sendCancellationNotification(ctx context.Context, event OrderEvent) error {
    // 模拟发送取消通知
    log.Printf("Sending cancellation notification for order: %d", event.OrderID)
    return nil
}

// 启动消费者
func (s *OrderMessageService) Start() error {
    return s.consumer.Start()
}

// 停止消费者
func (s *OrderMessageService) Stop() error {
    return s.consumer.Stop()
}
```

---

## ⚡ Apache Kafka集成实践

Kafka是高性能的分布式流处理平台，特别适合大数据场景和高吞吐量的消息处理。

### Kafka基础配置与连接

```go
// 来自 mall-go/internal/mq/kafka/client.go
package kafka

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/Shopify/sarama"
)

// Kafka配置
type Config struct {
    Brokers       []string      `yaml:"brokers"`
    ClientID      string        `yaml:"client_id"`
    Version       string        `yaml:"version"`

    // 生产者配置
    Producer ProducerConfig `yaml:"producer"`

    // 消费者配置
    Consumer ConsumerConfig `yaml:"consumer"`

    // 安全配置
    Security SecurityConfig `yaml:"security"`
}

// 生产者配置
type ProducerConfig struct {
    RequiredAcks      int           `yaml:"required_acks"`      // 0=不等待, 1=等待leader, -1=等待所有副本
    Timeout           time.Duration `yaml:"timeout"`
    Compression       string        `yaml:"compression"`        // none, gzip, snappy, lz4, zstd
    MaxMessageBytes   int           `yaml:"max_message_bytes"`
    RetryMax          int           `yaml:"retry_max"`
    RetryBackoff      time.Duration `yaml:"retry_backoff"`
    FlushFrequency    time.Duration `yaml:"flush_frequency"`
    FlushMessages     int           `yaml:"flush_messages"`
    FlushBytes        int           `yaml:"flush_bytes"`

    // 幂等性配置
    Idempotent        bool          `yaml:"idempotent"`
    MaxInFlight       int           `yaml:"max_in_flight"`
}

// 消费者配置
type ConsumerConfig struct {
    GroupID          string        `yaml:"group_id"`
    AutoOffsetReset  string        `yaml:"auto_offset_reset"`  // earliest, latest
    EnableAutoCommit bool          `yaml:"enable_auto_commit"`
    AutoCommitInterval time.Duration `yaml:"auto_commit_interval"`
    SessionTimeout   time.Duration `yaml:"session_timeout"`
    HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
    MaxProcessingTime time.Duration `yaml:"max_processing_time"`
    FetchMin         int32         `yaml:"fetch_min"`
    FetchMax         int32         `yaml:"fetch_max"`
    FetchDefault     int32         `yaml:"fetch_default"`
}

// 安全配置
type SecurityConfig struct {
    EnableSASL bool   `yaml:"enable_sasl"`
    SASLMechanism string `yaml:"sasl_mechanism"`
    Username   string `yaml:"username"`
    Password   string `yaml:"password"`
    EnableTLS  bool   `yaml:"enable_tls"`
    TLSConfig  TLSConfig `yaml:"tls_config"`
}

type TLSConfig struct {
    CertFile string `yaml:"cert_file"`
    KeyFile  string `yaml:"key_file"`
    CAFile   string `yaml:"ca_file"`
}

// Kafka客户端
type Client struct {
    config   *Config
    producer sarama.SyncProducer
    consumer sarama.Consumer
}

// 创建Kafka客户端
func NewClient(config *Config) (*Client, error) {
    // 解析Kafka版本
    version, err := sarama.ParseKafkaVersion(config.Version)
    if err != nil {
        return nil, fmt.Errorf("failed to parse Kafka version: %w", err)
    }

    // 创建Sarama配置
    saramaConfig := sarama.NewConfig()
    saramaConfig.Version = version
    saramaConfig.ClientID = config.ClientID

    // 配置生产者
    saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(config.Producer.RequiredAcks)
    saramaConfig.Producer.Timeout = config.Producer.Timeout
    saramaConfig.Producer.Compression = parseCompression(config.Producer.Compression)
    saramaConfig.Producer.MaxMessageBytes = config.Producer.MaxMessageBytes
    saramaConfig.Producer.Retry.Max = config.Producer.RetryMax
    saramaConfig.Producer.Retry.Backoff = config.Producer.RetryBackoff
    saramaConfig.Producer.Flush.Frequency = config.Producer.FlushFrequency
    saramaConfig.Producer.Flush.Messages = config.Producer.FlushMessages
    saramaConfig.Producer.Flush.Bytes = config.Producer.FlushBytes
    saramaConfig.Producer.Idempotent = config.Producer.Idempotent
    saramaConfig.Net.MaxOpenRequests = config.Producer.MaxInFlight
    saramaConfig.Producer.Return.Successes = true
    saramaConfig.Producer.Return.Errors = true

    // 配置消费者
    saramaConfig.Consumer.Group.Session.Timeout = config.Consumer.SessionTimeout
    saramaConfig.Consumer.Group.Heartbeat.Interval = config.Consumer.HeartbeatInterval
    saramaConfig.Consumer.MaxProcessingTime = config.Consumer.MaxProcessingTime
    saramaConfig.Consumer.Fetch.Min = config.Consumer.FetchMin
    saramaConfig.Consumer.Fetch.Max = config.Consumer.FetchMax
    saramaConfig.Consumer.Fetch.Default = config.Consumer.FetchDefault
    saramaConfig.Consumer.Return.Errors = true

    // 设置偏移量重置策略
    switch config.Consumer.AutoOffsetReset {
    case "earliest":
        saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
    case "latest":
        saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
    }

    // 配置安全设置
    if err := configureSecurity(saramaConfig, &config.Security); err != nil {
        return nil, err
    }

    // 创建生产者
    producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create producer: %w", err)
    }

    // 创建消费者
    consumer, err := sarama.NewConsumer(config.Brokers, saramaConfig)
    if err != nil {
        producer.Close()
        return nil, fmt.Errorf("failed to create consumer: %w", err)
    }

    client := &Client{
        config:   config,
        producer: producer,
        consumer: consumer,
    }

    log.Printf("Kafka client created successfully, brokers: %s",
               strings.Join(config.Brokers, ","))

    return client, nil
}

// 解析压缩算法
func parseCompression(compression string) sarama.CompressionCodec {
    switch strings.ToLower(compression) {
    case "gzip":
        return sarama.CompressionGZIP
    case "snappy":
        return sarama.CompressionSnappy
    case "lz4":
        return sarama.CompressionLZ4
    case "zstd":
        return sarama.CompressionZSTD
    default:
        return sarama.CompressionNone
    }
}

// 配置安全设置
func configureSecurity(config *sarama.Config, security *SecurityConfig) error {
    if security.EnableSASL {
        config.Net.SASL.Enable = true
        config.Net.SASL.Mechanism = sarama.SASLMechanism(security.SASLMechanism)
        config.Net.SASL.User = security.Username
        config.Net.SASL.Password = security.Password
    }

    if security.EnableTLS {
        config.Net.TLS.Enable = true
        // 这里可以配置TLS证书等
    }

    return nil
}

// 关闭客户端
func (c *Client) Close() error {
    var errs []error

    if err := c.producer.Close(); err != nil {
        errs = append(errs, err)
    }

    if err := c.consumer.Close(); err != nil {
        errs = append(errs, err)
    }

    if len(errs) > 0 {
        return fmt.Errorf("errors closing client: %v", errs)
    }

    return nil
}
```

### Kafka生产者实现

```go
// 来自 mall-go/internal/mq/kafka/producer.go
package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/Shopify/sarama"
)

// Kafka生产者
type Producer struct {
    client   *Client
    producer sarama.SyncProducer
}

// 创建生产者
func NewProducer(client *Client) *Producer {
    return &Producer{
        client:   client,
        producer: client.producer,
    }
}

// 消息结构
type KafkaMessage struct {
    Key       string                 `json:"key"`
    Value     interface{}            `json:"value"`
    Headers   map[string]string      `json:"headers"`
    Timestamp time.Time              `json:"timestamp"`
}

// 发送消息
func (p *Producer) Send(ctx context.Context, topic string, message *KafkaMessage) error {
    // 序列化消息值
    valueBytes, err := json.Marshal(message.Value)
    if err != nil {
        return fmt.Errorf("failed to marshal message value: %w", err)
    }

    // 构建Kafka消息
    kafkaMsg := &sarama.ProducerMessage{
        Topic:     topic,
        Key:       sarama.StringEncoder(message.Key),
        Value:     sarama.ByteEncoder(valueBytes),
        Timestamp: message.Timestamp,
    }

    // 添加消息头
    if message.Headers != nil {
        kafkaMsg.Headers = make([]sarama.RecordHeader, 0, len(message.Headers))
        for k, v := range message.Headers {
            kafkaMsg.Headers = append(kafkaMsg.Headers, sarama.RecordHeader{
                Key:   []byte(k),
                Value: []byte(v),
            })
        }
    }

    // 发送消息
    partition, offset, err := p.producer.SendMessage(kafkaMsg)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }

    fmt.Printf("Message sent successfully: topic=%s, partition=%d, offset=%d\n",
               topic, partition, offset)
    return nil
}

// 批量发送消息
func (p *Producer) SendBatch(ctx context.Context, topic string, messages []*KafkaMessage) error {
    kafkaMessages := make([]*sarama.ProducerMessage, len(messages))

    for i, msg := range messages {
        valueBytes, err := json.Marshal(msg.Value)
        if err != nil {
            return fmt.Errorf("failed to marshal message %d: %w", i, err)
        }

        kafkaMsg := &sarama.ProducerMessage{
            Topic:     topic,
            Key:       sarama.StringEncoder(msg.Key),
            Value:     sarama.ByteEncoder(valueBytes),
            Timestamp: msg.Timestamp,
        }

        if msg.Headers != nil {
            kafkaMsg.Headers = make([]sarama.RecordHeader, 0, len(msg.Headers))
            for k, v := range msg.Headers {
                kafkaMsg.Headers = append(kafkaMsg.Headers, sarama.RecordHeader{
                    Key:   []byte(k),
                    Value: []byte(v),
                })
            }
        }

        kafkaMessages[i] = kafkaMsg
    }

    // 批量发送
    return p.producer.SendMessages(kafkaMessages)
}

// 发送带分区键的消息
func (p *Producer) SendWithPartition(ctx context.Context, topic string, partition int32, message *KafkaMessage) error {
    valueBytes, err := json.Marshal(message.Value)
    if err != nil {
        return fmt.Errorf("failed to marshal message value: %w", err)
    }

    kafkaMsg := &sarama.ProducerMessage{
        Topic:     topic,
        Partition: partition,
        Key:       sarama.StringEncoder(message.Key),
        Value:     sarama.ByteEncoder(valueBytes),
        Timestamp: message.Timestamp,
    }

    if message.Headers != nil {
        kafkaMsg.Headers = make([]sarama.RecordHeader, 0, len(message.Headers))
        for k, v := range message.Headers {
            kafkaMsg.Headers = append(kafkaMsg.Headers, sarama.RecordHeader{
                Key:   []byte(k),
                Value: []byte(v),
            })
        }
    }

    partition, offset, err := p.producer.SendMessage(kafkaMsg)
    if err != nil {
        return fmt.Errorf("failed to send message: %w", err)
    }

    fmt.Printf("Message sent to specific partition: topic=%s, partition=%d, offset=%d\n",
               topic, partition, offset)
    return nil
}
```

### Kafka消费者实现

```go
// 来自 mall-go/internal/mq/kafka/consumer.go
package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"

    "github.com/Shopify/sarama"
)

// 消息处理器
type KafkaMessageHandler func(ctx context.Context, message *sarama.ConsumerMessage) error

// 消费者组处理器
type ConsumerGroupHandler struct {
    handlers map[string]KafkaMessageHandler
    mutex    sync.RWMutex
}

// 创建消费者组处理器
func NewConsumerGroupHandler() *ConsumerGroupHandler {
    return &ConsumerGroupHandler{
        handlers: make(map[string]KafkaMessageHandler),
    }
}

// 注册消息处理器
func (h *ConsumerGroupHandler) RegisterHandler(topic string, handler KafkaMessageHandler) {
    h.mutex.Lock()
    defer h.mutex.Unlock()

    h.handlers[topic] = handler
}

// 实现sarama.ConsumerGroupHandler接口
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
    return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
    return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    for {
        select {
        case message := <-claim.Messages():
            if message == nil {
                return nil
            }

            if err := h.processMessage(session.Context(), message); err != nil {
                log.Printf("Failed to process message: %v", err)
                // 根据业务需求决定是否继续处理
                continue
            }

            // 标记消息已处理
            session.MarkMessage(message, "")

        case <-session.Context().Done():
            return nil
        }
    }
}

// 处理消息
func (h *ConsumerGroupHandler) processMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
    h.mutex.RLock()
    handler, exists := h.handlers[message.Topic]
    h.mutex.RUnlock()

    if !exists {
        return fmt.Errorf("no handler registered for topic: %s", message.Topic)
    }

    return handler(ctx, message)
}

// Kafka消费者
type Consumer struct {
    client        *Client
    consumerGroup sarama.ConsumerGroup
    handler       *ConsumerGroupHandler
    topics        []string
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
}

// 创建消费者
func NewConsumer(client *Client, groupID string, topics []string) (*Consumer, error) {
    consumerGroup, err := sarama.NewConsumerGroupFromClient(groupID, client.consumer.(*sarama.consumer).client)
    if err != nil {
        return nil, fmt.Errorf("failed to create consumer group: %w", err)
    }

    ctx, cancel := context.WithCancel(context.Background())

    consumer := &Consumer{
        client:        client,
        consumerGroup: consumerGroup,
        handler:       NewConsumerGroupHandler(),
        topics:        topics,
        ctx:           ctx,
        cancel:        cancel,
    }

    return consumer, nil
}

// 注册消息处理器
func (c *Consumer) RegisterHandler(topic string, handler KafkaMessageHandler) {
    c.handler.RegisterHandler(topic, handler)
}

// 开始消费
func (c *Consumer) Start() error {
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()

        for {
            select {
            case <-c.ctx.Done():
                return
            default:
                if err := c.consumerGroup.Consume(c.ctx, c.topics, c.handler); err != nil {
                    log.Printf("Consumer group error: %v", err)
                    return
                }
            }
        }
    }()

    // 监听错误
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()

        for {
            select {
            case <-c.ctx.Done():
                return
            case err := <-c.consumerGroup.Errors():
                if err != nil {
                    log.Printf("Consumer group error: %v", err)
                }
            }
        }
    }()

    log.Printf("Kafka consumer started for topics: %v", c.topics)
    return nil
}

// 停止消费
func (c *Consumer) Stop() error {
    c.cancel()
    c.wg.Wait()

    if err := c.consumerGroup.Close(); err != nil {
        return fmt.Errorf("failed to close consumer group: %w", err)
    }

    log.Printf("Kafka consumer stopped for topics: %v", c.topics)
    return nil
}
```

### Kafka实际应用示例

```go
// 来自 mall-go/internal/service/analytics_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/Shopify/sarama"
    "mall-go/internal/mq/kafka"
)

// 用户行为分析服务
type AnalyticsService struct {
    producer *kafka.Producer
    consumer *kafka.Consumer
}

// 用户行为事件
type UserBehaviorEvent struct {
    UserID    uint                   `json:"user_id"`
    EventType string                 `json:"event_type"`
    Page      string                 `json:"page"`
    Action    string                 `json:"action"`
    Properties map[string]interface{} `json:"properties"`
    Timestamp  int64                  `json:"timestamp"`
    SessionID  string                 `json:"session_id"`
    IP         string                 `json:"ip"`
    UserAgent  string                 `json:"user_agent"`
}

// 创建分析服务
func NewAnalyticsService(client *kafka.Client) (*AnalyticsService, error) {
    producer := kafka.NewProducer(client)

    consumer, err := kafka.NewConsumer(client, "analytics-group", []string{
        "user.behavior",
        "order.events",
        "product.events",
    })
    if err != nil {
        return nil, err
    }

    service := &AnalyticsService{
        producer: producer,
        consumer: consumer,
    }

    // 注册消息处理器
    service.registerHandlers()

    return service, nil
}

// 注册消息处理器
func (s *AnalyticsService) registerHandlers() {
    // 用户行为分析
    s.consumer.RegisterHandler("user.behavior", s.handleUserBehavior)

    // 订单事件分析
    s.consumer.RegisterHandler("order.events", s.handleOrderEvents)

    // 商品事件分析
    s.consumer.RegisterHandler("product.events", s.handleProductEvents)
}

// 记录用户行为
func (s *AnalyticsService) TrackUserBehavior(ctx context.Context, event *UserBehaviorEvent) error {
    message := &kafka.KafkaMessage{
        Key:   fmt.Sprintf("user_%d", event.UserID),
        Value: event,
        Headers: map[string]string{
            "event_type": event.EventType,
            "source":     "web",
        },
        Timestamp: time.Now(),
    }

    return s.producer.Send(ctx, "user.behavior", message)
}

// 处理用户行为事件
func (s *AnalyticsService) handleUserBehavior(ctx context.Context, message *sarama.ConsumerMessage) error {
    var event UserBehaviorEvent
    if err := json.Unmarshal(message.Value, &event); err != nil {
        return fmt.Errorf("failed to unmarshal user behavior event: %w", err)
    }

    log.Printf("Processing user behavior: UserID=%d, EventType=%s, Action=%s",
               event.UserID, event.EventType, event.Action)

    // 1. 实时用户画像更新
    if err := s.updateUserProfile(ctx, &event); err != nil {
        log.Printf("Failed to update user profile: %v", err)
    }

    // 2. 实时推荐计算
    if err := s.updateRecommendations(ctx, &event); err != nil {
        log.Printf("Failed to update recommendations: %v", err)
    }

    // 3. 异常行为检测
    if err := s.detectAnomalies(ctx, &event); err != nil {
        log.Printf("Failed to detect anomalies: %v", err)
    }

    // 4. 存储到数据仓库
    if err := s.storeToWarehouse(ctx, &event); err != nil {
        log.Printf("Failed to store to warehouse: %v", err)
    }

    return nil
}

// 处理订单事件
func (s *AnalyticsService) handleOrderEvents(ctx context.Context, message *sarama.ConsumerMessage) error {
    log.Printf("Processing order event: Partition=%d, Offset=%d",
               message.Partition, message.Offset)

    // 解析订单事件
    var orderEvent map[string]interface{}
    if err := json.Unmarshal(message.Value, &orderEvent); err != nil {
        return fmt.Errorf("failed to unmarshal order event: %w", err)
    }

    // 1. 更新销售统计
    if err := s.updateSalesStats(ctx, orderEvent); err != nil {
        log.Printf("Failed to update sales stats: %v", err)
    }

    // 2. 更新商品热度
    if err := s.updateProductPopularity(ctx, orderEvent); err != nil {
        log.Printf("Failed to update product popularity: %v", err)
    }

    // 3. 用户价值分析
    if err := s.analyzeUserValue(ctx, orderEvent); err != nil {
        log.Printf("Failed to analyze user value: %v", err)
    }

    return nil
}

// 处理商品事件
func (s *AnalyticsService) handleProductEvents(ctx context.Context, message *sarama.ConsumerMessage) error {
    log.Printf("Processing product event: Partition=%d, Offset=%d",
               message.Partition, message.Offset)

    // 解析商品事件
    var productEvent map[string]interface{}
    if err := json.Unmarshal(message.Value, &productEvent); err != nil {
        return fmt.Errorf("failed to unmarshal product event: %w", err)
    }

    // 1. 更新商品统计
    if err := s.updateProductStats(ctx, productEvent); err != nil {
        log.Printf("Failed to update product stats: %v", err)
    }

    // 2. 库存预警分析
    if err := s.analyzeInventoryAlerts(ctx, productEvent); err != nil {
        log.Printf("Failed to analyze inventory alerts: %v", err)
    }

    return nil
}

// 辅助方法实现
func (s *AnalyticsService) updateUserProfile(ctx context.Context, event *UserBehaviorEvent) error {
    // 模拟用户画像更新
    log.Printf("Updating user profile for user: %d", event.UserID)
    return nil
}

func (s *AnalyticsService) updateRecommendations(ctx context.Context, event *UserBehaviorEvent) error {
    // 模拟推荐更新
    log.Printf("Updating recommendations for user: %d", event.UserID)
    return nil
}

func (s *AnalyticsService) detectAnomalies(ctx context.Context, event *UserBehaviorEvent) error {
    // 模拟异常检测
    log.Printf("Detecting anomalies for user: %d", event.UserID)
    return nil
}

func (s *AnalyticsService) storeToWarehouse(ctx context.Context, event *UserBehaviorEvent) error {
    // 模拟数据仓库存储
    log.Printf("Storing event to warehouse: %s", event.EventType)
    return nil
}

func (s *AnalyticsService) updateSalesStats(ctx context.Context, orderEvent map[string]interface{}) error {
    // 模拟销售统计更新
    log.Printf("Updating sales statistics")
    return nil
}

func (s *AnalyticsService) updateProductPopularity(ctx context.Context, orderEvent map[string]interface{}) error {
    // 模拟商品热度更新
    log.Printf("Updating product popularity")
    return nil
}

func (s *AnalyticsService) analyzeUserValue(ctx context.Context, orderEvent map[string]interface{}) error {
    // 模拟用户价值分析
    log.Printf("Analyzing user value")
    return nil
}

func (s *AnalyticsService) updateProductStats(ctx context.Context, productEvent map[string]interface{}) error {
    // 模拟商品统计更新
    log.Printf("Updating product statistics")
    return nil
}

func (s *AnalyticsService) analyzeInventoryAlerts(ctx context.Context, productEvent map[string]interface{}) error {
    // 模拟库存预警分析
    log.Printf("Analyzing inventory alerts")
    return nil
}

// 启动服务
func (s *AnalyticsService) Start() error {
    return s.consumer.Start()
}

// 停止服务
func (s *AnalyticsService) Stop() error {
    return s.consumer.Stop()
}
```

---

## 🚀 NSQ集成实践

NSQ是Go语言原生的分布式消息队列，具有简单易用、高性能、去中心化等特点。

### NSQ基础配置与连接

```go
// 来自 mall-go/internal/mq/nsq/client.go
package nsq

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/nsqio/go-nsq"
)

// NSQ配置
type Config struct {
    NSQDAddrs    []string      `yaml:"nsqd_addrs"`
    LookupdAddrs []string      `yaml:"lookupd_addrs"`

    // 生产者配置
    Producer ProducerConfig `yaml:"producer"`

    // 消费者配置
    Consumer ConsumerConfig `yaml:"consumer"`
}

// 生产者配置
type ProducerConfig struct {
    MaxInFlight int           `yaml:"max_in_flight"`
    DialTimeout time.Duration `yaml:"dial_timeout"`
    ReadTimeout time.Duration `yaml:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout"`
    LocalAddr   string        `yaml:"local_addr"`
    UserAgent   string        `yaml:"user_agent"`
}

// 消费者配置
type ConsumerConfig struct {
    MaxInFlight        int           `yaml:"max_in_flight"`
    MaxAttempts        uint16        `yaml:"max_attempts"`
    RequeueDelay       time.Duration `yaml:"requeue_delay"`
    DefaultRequeueDelay time.Duration `yaml:"default_requeue_delay"`
    BackoffStrategy    string        `yaml:"backoff_strategy"`
    MaxBackoffDuration time.Duration `yaml:"max_backoff_duration"`
    BackoffMultiplier  time.Duration `yaml:"backoff_multiplier"`
    LookupdPollInterval time.Duration `yaml:"lookupd_poll_interval"`
    LookupdPollJitter  float64       `yaml:"lookupd_poll_jitter"`
}

// NSQ客户端
type Client struct {
    config    *Config
    producers map[string]*nsq.Producer
    consumers map[string]*nsq.Consumer
    mutex     sync.RWMutex
}

// 创建NSQ客户端
func NewClient(config *Config) (*Client, error) {
    client := &Client{
        config:    config,
        producers: make(map[string]*nsq.Producer),
        consumers: make(map[string]*nsq.Consumer),
    }

    return client, nil
}

// 获取生产者
func (c *Client) GetProducer(nsqdAddr string) (*nsq.Producer, error) {
    c.mutex.RLock()
    if producer, exists := c.producers[nsqdAddr]; exists {
        c.mutex.RUnlock()
        return producer, nil
    }
    c.mutex.RUnlock()

    c.mutex.Lock()
    defer c.mutex.Unlock()

    // 双重检查
    if producer, exists := c.producers[nsqdAddr]; exists {
        return producer, nil
    }

    // 创建生产者配置
    config := nsq.NewConfig()
    config.MaxInFlight = c.config.Producer.MaxInFlight
    config.DialTimeout = c.config.Producer.DialTimeout
    config.ReadTimeout = c.config.Producer.ReadTimeout
    config.WriteTimeout = c.config.Producer.WriteTimeout
    config.LocalAddr = c.config.Producer.LocalAddr
    config.UserAgent = c.config.Producer.UserAgent

    // 创建生产者
    producer, err := nsq.NewProducer(nsqdAddr, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create NSQ producer: %w", err)
    }

    c.producers[nsqdAddr] = producer
    log.Printf("NSQ producer created for: %s", nsqdAddr)

    return producer, nil
}

// 创建消费者
func (c *Client) CreateConsumer(topic, channel string) (*nsq.Consumer, error) {
    // 创建消费者配置
    config := nsq.NewConfig()
    config.MaxInFlight = c.config.Consumer.MaxInFlight
    config.MaxAttempts = c.config.Consumer.MaxAttempts
    config.RequeueDelay = c.config.Consumer.RequeueDelay
    config.DefaultRequeueDelay = c.config.Consumer.DefaultRequeueDelay
    config.MaxBackoffDuration = c.config.Consumer.MaxBackoffDuration
    config.BackoffMultiplier = c.config.Consumer.BackoffMultiplier
    config.LookupdPollInterval = c.config.Consumer.LookupdPollInterval
    config.LookupdPollJitter = c.config.Consumer.LookupdPollJitter

    // 设置退避策略
    switch c.config.Consumer.BackoffStrategy {
    case "exponential":
        config.BackoffStrategy = nsq.ExponentialStrategy{}
    case "full_jitter":
        config.BackoffStrategy = nsq.FullJitterStrategy{}
    default:
        config.BackoffStrategy = nsq.ExponentialStrategy{}
    }

    // 创建消费者
    consumer, err := nsq.NewConsumer(topic, channel, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create NSQ consumer: %w", err)
    }

    consumerKey := fmt.Sprintf("%s:%s", topic, channel)
    c.mutex.Lock()
    c.consumers[consumerKey] = consumer
    c.mutex.Unlock()

    log.Printf("NSQ consumer created for topic: %s, channel: %s", topic, channel)

    return consumer, nil
}

// 关闭客户端
func (c *Client) Close() error {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    var errs []error

    // 关闭所有生产者
    for addr, producer := range c.producers {
        producer.Stop()
        if err := <-producer.StopChan; err != nil {
            errs = append(errs, fmt.Errorf("error stopping producer %s: %w", addr, err))
        }
    }

    // 关闭所有消费者
    for key, consumer := range c.consumers {
        consumer.Stop()
        if err := <-consumer.StopChan; err != nil {
            errs = append(errs, fmt.Errorf("error stopping consumer %s: %w", key, err))
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("errors closing NSQ client: %v", errs)
    }

    return nil
}
```

### NSQ生产者和消费者实现

```go
// 来自 mall-go/internal/mq/nsq/producer.go
package nsq

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/nsqio/go-nsq"
)

// NSQ生产者包装器
type Producer struct {
    client   *Client
    nsqdAddr string
    producer *nsq.Producer
}

// 创建生产者
func NewProducer(client *Client, nsqdAddr string) (*Producer, error) {
    producer, err := client.GetProducer(nsqdAddr)
    if err != nil {
        return nil, err
    }

    return &Producer{
        client:   client,
        nsqdAddr: nsqdAddr,
        producer: producer,
    }, nil
}

// NSQ消息结构
type NSQMessage struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Timestamp int64                  `json:"timestamp"`
    Retry     int                    `json:"retry"`
}

// 发布消息
func (p *Producer) Publish(ctx context.Context, topic string, message *NSQMessage) error {
    messageBytes, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    if err := p.producer.Publish(topic, messageBytes); err != nil {
        return fmt.Errorf("failed to publish message: %w", err)
    }

    fmt.Printf("Message published to NSQ: topic=%s, id=%s\n", topic, message.ID)
    return nil
}

// 延迟发布消息
func (p *Producer) DeferredPublish(ctx context.Context, topic string, delay time.Duration, message *NSQMessage) error {
    messageBytes, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    if err := p.producer.DeferredPublish(topic, delay, messageBytes); err != nil {
        return fmt.Errorf("failed to publish deferred message: %w", err)
    }

    fmt.Printf("Deferred message published to NSQ: topic=%s, delay=%v, id=%s\n",
               topic, delay, message.ID)
    return nil
}

// 批量发布消息
func (p *Producer) MultiPublish(ctx context.Context, topic string, messages []*NSQMessage) error {
    messageBytes := make([][]byte, len(messages))

    for i, msg := range messages {
        bytes, err := json.Marshal(msg)
        if err != nil {
            return fmt.Errorf("failed to marshal message %d: %w", i, err)
        }
        messageBytes[i] = bytes
    }

    if err := p.producer.MultiPublish(topic, messageBytes); err != nil {
        return fmt.Errorf("failed to multi-publish messages: %w", err)
    }

    fmt.Printf("Multi-published %d messages to NSQ: topic=%s\n", len(messages), topic)
    return nil
}

// NSQ消费者包装器
type Consumer struct {
    client   *Client
    consumer *nsq.Consumer
    topic    string
    channel  string
    handlers map[string]NSQMessageHandler
    mutex    sync.RWMutex
}

// 消息处理器
type NSQMessageHandler func(ctx context.Context, message *NSQMessage) error

// 创建消费者
func NewConsumer(client *Client, topic, channel string) (*Consumer, error) {
    consumer, err := client.CreateConsumer(topic, channel)
    if err != nil {
        return nil, err
    }

    c := &Consumer{
        client:   client,
        consumer: consumer,
        topic:    topic,
        channel:  channel,
        handlers: make(map[string]NSQMessageHandler),
    }

    // 设置消息处理器
    consumer.AddHandler(c)

    return c, nil
}

// 实现nsq.Handler接口
func (c *Consumer) HandleMessage(message *nsq.Message) error {
    var nsqMsg NSQMessage
    if err := json.Unmarshal(message.Body, &nsqMsg); err != nil {
        return fmt.Errorf("failed to unmarshal NSQ message: %w", err)
    }

    c.mutex.RLock()
    handler, exists := c.handlers[nsqMsg.Type]
    c.mutex.RUnlock()

    if !exists {
        return fmt.Errorf("no handler registered for message type: %s", nsqMsg.Type)
    }

    ctx := context.Background()
    if err := handler(ctx, &nsqMsg); err != nil {
        // 增加重试次数
        nsqMsg.Retry++

        // 如果重试次数超过限制，记录错误但不重新入队
        if nsqMsg.Retry >= 3 {
            log.Printf("Message processing failed after %d retries: %v", nsqMsg.Retry, err)
            return nil // 返回nil避免重新入队
        }

        return err // 返回错误会导致消息重新入队
    }

    return nil
}

// 注册消息处理器
func (c *Consumer) RegisterHandler(messageType string, handler NSQMessageHandler) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.handlers[messageType] = handler
}

// 连接到NSQ
func (c *Consumer) ConnectToNSQD(nsqdAddr string) error {
    return c.consumer.ConnectToNSQD(nsqdAddr)
}

// 连接到NSQLookupd
func (c *Consumer) ConnectToNSQLookupd(lookupdAddr string) error {
    return c.consumer.ConnectToNSQLookupd(lookupdAddr)
}

// 停止消费者
func (c *Consumer) Stop() {
    c.consumer.Stop()
}
```

### NSQ实际应用示例

```go
// 来自 mall-go/internal/service/notification_service.go
package service

import (
    "context"
    "fmt"
    "log"
    "time"

    "mall-go/internal/mq/nsq"
)

// 通知服务
type NotificationService struct {
    producer *nsq.Producer
    consumer *nsq.Consumer
}

// 通知类型
const (
    NotificationTypeEmail = "email"
    NotificationTypeSMS   = "sms"
    NotificationTypePush  = "push"
    NotificationTypeInApp = "in_app"
)

// 通知消息
type NotificationMessage struct {
    UserID      uint                   `json:"user_id"`
    Type        string                 `json:"type"`
    Title       string                 `json:"title"`
    Content     string                 `json:"content"`
    Template    string                 `json:"template"`
    Variables   map[string]interface{} `json:"variables"`
    Priority    int                    `json:"priority"`
    ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// 创建通知服务
func NewNotificationService(client *nsq.Client) (*NotificationService, error) {
    // 创建生产者
    producer, err := nsq.NewProducer(client, "127.0.0.1:4150")
    if err != nil {
        return nil, err
    }

    // 创建消费者
    consumer, err := nsq.NewConsumer(client, "notifications", "notification-processor")
    if err != nil {
        return nil, err
    }

    service := &NotificationService{
        producer: producer,
        consumer: consumer,
    }

    // 注册消息处理器
    service.registerHandlers()

    return service, nil
}

// 注册消息处理器
func (s *NotificationService) registerHandlers() {
    s.consumer.RegisterHandler(NotificationTypeEmail, s.handleEmailNotification)
    s.consumer.RegisterHandler(NotificationTypeSMS, s.handleSMSNotification)
    s.consumer.RegisterHandler(NotificationTypePush, s.handlePushNotification)
    s.consumer.RegisterHandler(NotificationTypeInApp, s.handleInAppNotification)
}

// 发送通知
func (s *NotificationService) SendNotification(ctx context.Context, notification *NotificationMessage) error {
    message := &nsq.NSQMessage{
        ID:   generateMessageID(),
        Type: notification.Type,
        Payload: map[string]interface{}{
            "user_id":      notification.UserID,
            "title":        notification.Title,
            "content":      notification.Content,
            "template":     notification.Template,
            "variables":    notification.Variables,
            "priority":     notification.Priority,
            "scheduled_at": notification.ScheduledAt,
        },
        Timestamp: time.Now().Unix(),
    }

    // 如果是定时通知，使用延迟发布
    if notification.ScheduledAt != nil && notification.ScheduledAt.After(time.Now()) {
        delay := notification.ScheduledAt.Sub(time.Now())
        return s.producer.DeferredPublish(ctx, "notifications", delay, message)
    }

    return s.producer.Publish(ctx, "notifications", message)
}

// 批量发送通知
func (s *NotificationService) SendBatchNotifications(ctx context.Context, notifications []*NotificationMessage) error {
    messages := make([]*nsq.NSQMessage, len(notifications))

    for i, notification := range notifications {
        messages[i] = &nsq.NSQMessage{
            ID:   generateMessageID(),
            Type: notification.Type,
            Payload: map[string]interface{}{
                "user_id":   notification.UserID,
                "title":     notification.Title,
                "content":   notification.Content,
                "template":  notification.Template,
                "variables": notification.Variables,
                "priority":  notification.Priority,
            },
            Timestamp: time.Now().Unix(),
        }
    }

    return s.producer.MultiPublish(ctx, "notifications", messages)
}

// 处理邮件通知
func (s *NotificationService) handleEmailNotification(ctx context.Context, message *nsq.NSQMessage) error {
    userID := uint(message.Payload["user_id"].(float64))
    title := message.Payload["title"].(string)
    content := message.Payload["content"].(string)

    log.Printf("Sending email notification: UserID=%d, Title=%s", userID, title)

    // 模拟邮件发送
    if err := s.sendEmail(ctx, userID, title, content); err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }

    log.Printf("Email notification sent successfully: UserID=%d", userID)
    return nil
}

// 处理短信通知
func (s *NotificationService) handleSMSNotification(ctx context.Context, message *nsq.NSQMessage) error {
    userID := uint(message.Payload["user_id"].(float64))
    content := message.Payload["content"].(string)

    log.Printf("Sending SMS notification: UserID=%d", userID)

    // 模拟短信发送
    if err := s.sendSMS(ctx, userID, content); err != nil {
        return fmt.Errorf("failed to send SMS: %w", err)
    }

    log.Printf("SMS notification sent successfully: UserID=%d", userID)
    return nil
}

// 处理推送通知
func (s *NotificationService) handlePushNotification(ctx context.Context, message *nsq.NSQMessage) error {
    userID := uint(message.Payload["user_id"].(float64))
    title := message.Payload["title"].(string)
    content := message.Payload["content"].(string)

    log.Printf("Sending push notification: UserID=%d, Title=%s", userID, title)

    // 模拟推送发送
    if err := s.sendPush(ctx, userID, title, content); err != nil {
        return fmt.Errorf("failed to send push: %w", err)
    }

    log.Printf("Push notification sent successfully: UserID=%d", userID)
    return nil
}

// 处理应用内通知
func (s *NotificationService) handleInAppNotification(ctx context.Context, message *nsq.NSQMessage) error {
    userID := uint(message.Payload["user_id"].(float64))
    title := message.Payload["title"].(string)
    content := message.Payload["content"].(string)

    log.Printf("Sending in-app notification: UserID=%d, Title=%s", userID, title)

    // 模拟应用内通知
    if err := s.sendInAppNotification(ctx, userID, title, content); err != nil {
        return fmt.Errorf("failed to send in-app notification: %w", err)
    }

    log.Printf("In-app notification sent successfully: UserID=%d", userID)
    return nil
}

// 辅助方法
func (s *NotificationService) sendEmail(ctx context.Context, userID uint, title, content string) error {
    // 模拟邮件发送逻辑
    time.Sleep(100 * time.Millisecond)
    return nil
}

func (s *NotificationService) sendSMS(ctx context.Context, userID uint, content string) error {
    // 模拟短信发送逻辑
    time.Sleep(50 * time.Millisecond)
    return nil
}

func (s *NotificationService) sendPush(ctx context.Context, userID uint, title, content string) error {
    // 模拟推送发送逻辑
    time.Sleep(30 * time.Millisecond)
    return nil
}

func (s *NotificationService) sendInAppNotification(ctx context.Context, userID uint, title, content string) error {
    // 模拟应用内通知逻辑
    time.Sleep(10 * time.Millisecond)
    return nil
}

func generateMessageID() string {
    return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// 启动服务
func (s *NotificationService) Start() error {
    // 连接到NSQLookupd
    if err := s.consumer.ConnectToNSQLookupd("127.0.0.1:4161"); err != nil {
        return err
    }

    log.Printf("Notification service started")
    return nil
}

// 停止服务
func (s *NotificationService) Stop() {
    s.consumer.Stop()
    log.Printf("Notification service stopped")
}
```

---

## 🎪 事件驱动架构设计

事件驱动架构（Event-Driven Architecture，EDA）是一种基于事件的软件架构模式，通过事件的产生、传播和消费来实现系统间的松耦合通信。

### 事件驱动架构核心概念

```go
// 来自 mall-go/internal/event/event.go
package event

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
)

// 事件接口
type Event interface {
    GetEventID() string
    GetEventType() string
    GetAggregateID() string
    GetVersion() int
    GetTimestamp() time.Time
    GetPayload() interface{}
    GetMetadata() map[string]interface{}
}

// 基础事件结构
type BaseEvent struct {
    EventID     string                 `json:"event_id"`
    EventType   string                 `json:"event_type"`
    AggregateID string                 `json:"aggregate_id"`
    Version     int                    `json:"version"`
    Timestamp   time.Time              `json:"timestamp"`
    Payload     interface{}            `json:"payload"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// 实现Event接口
func (e *BaseEvent) GetEventID() string                 { return e.EventID }
func (e *BaseEvent) GetEventType() string               { return e.EventType }
func (e *BaseEvent) GetAggregateID() string             { return e.AggregateID }
func (e *BaseEvent) GetVersion() int                    { return e.Version }
func (e *BaseEvent) GetTimestamp() time.Time            { return e.Timestamp }
func (e *BaseEvent) GetPayload() interface{}            { return e.Payload }
func (e *BaseEvent) GetMetadata() map[string]interface{} { return e.Metadata }

// 事件总线接口
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
    Start() error
    Stop() error
}

// 事件处理器
type EventHandler interface {
    Handle(ctx context.Context, event Event) error
    GetHandlerName() string
}

// 事件存储接口
type EventStore interface {
    SaveEvent(ctx context.Context, event Event) error
    GetEvents(ctx context.Context, aggregateID string) ([]Event, error)
    GetEventsByType(ctx context.Context, eventType string, limit int) ([]Event, error)
}

// 事件发布器
type EventPublisher interface {
    PublishEvent(ctx context.Context, event Event) error
    PublishEvents(ctx context.Context, events []Event) error
}

// 聚合根接口
type AggregateRoot interface {
    GetID() string
    GetVersion() int
    GetUncommittedEvents() []Event
    MarkEventsAsCommitted()
    LoadFromHistory(events []Event) error
}
```

### 事件总线实现

```go
// 来自 mall-go/internal/event/bus.go
package event

import (
    "context"
    "fmt"
    "log"
    "sync"

    "mall-go/internal/mq/rabbitmq"
)

// 基于RabbitMQ的事件总线
type RabbitMQEventBus struct {
    producer  *rabbitmq.Producer
    consumer  *rabbitmq.Consumer
    handlers  map[string][]EventHandler
    mutex     sync.RWMutex
    started   bool
}

// 创建事件总线
func NewRabbitMQEventBus(client *rabbitmq.Client) (*RabbitMQEventBus, error) {
    producer, err := rabbitmq.NewProducer(client, "events")
    if err != nil {
        return nil, err
    }

    consumerConfig := &rabbitmq.ConsumerConfig{
        QueueName:    "event-bus",
        RoutingKey:   "event.*",
        ConsumerTag:  "event-bus-consumer",
        AutoAck:      false,
        PrefetchCount: 10,
        MaxRetries:   3,
    }

    consumer, err := rabbitmq.NewConsumer(client, "events", consumerConfig)
    if err != nil {
        return nil, err
    }

    bus := &RabbitMQEventBus{
        producer: producer,
        consumer: consumer,
        handlers: make(map[string][]EventHandler),
    }

    return bus, nil
}

// 发布事件
func (bus *RabbitMQEventBus) Publish(ctx context.Context, event Event) error {
    routingKey := fmt.Sprintf("event.%s", event.GetEventType())

    return bus.producer.Send(ctx, routingKey, event)
}

// 订阅事件
func (bus *RabbitMQEventBus) Subscribe(eventType string, handler EventHandler) error {
    bus.mutex.Lock()
    defer bus.mutex.Unlock()

    if bus.handlers[eventType] == nil {
        bus.handlers[eventType] = make([]EventHandler, 0)
    }

    bus.handlers[eventType] = append(bus.handlers[eventType], handler)

    log.Printf("Event handler registered: %s -> %s", eventType, handler.GetHandlerName())
    return nil
}

// 取消订阅
func (bus *RabbitMQEventBus) Unsubscribe(eventType string, handler EventHandler) error {
    bus.mutex.Lock()
    defer bus.mutex.Unlock()

    handlers := bus.handlers[eventType]
    for i, h := range handlers {
        if h.GetHandlerName() == handler.GetHandlerName() {
            bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
            break
        }
    }

    return nil
}

// 启动事件总线
func (bus *RabbitMQEventBus) Start() error {
    if bus.started {
        return nil
    }

    // 注册消息处理器
    bus.consumer.RegisterHandler("event.*", bus.handleEvent)

    // 启动消费者
    if err := bus.consumer.Start(); err != nil {
        return err
    }

    bus.started = true
    log.Printf("Event bus started")
    return nil
}

// 停止事件总线
func (bus *RabbitMQEventBus) Stop() error {
    if !bus.started {
        return nil
    }

    if err := bus.consumer.Stop(); err != nil {
        return err
    }

    bus.started = false
    log.Printf("Event bus stopped")
    return nil
}

// 处理事件消息
func (bus *RabbitMQEventBus) handleEvent(ctx context.Context, delivery amqp.Delivery) error {
    var baseEvent BaseEvent
    if err := json.Unmarshal(delivery.Body, &baseEvent); err != nil {
        return fmt.Errorf("failed to unmarshal event: %w", err)
    }

    bus.mutex.RLock()
    handlers := bus.handlers[baseEvent.EventType]
    bus.mutex.RUnlock()

    if len(handlers) == 0 {
        log.Printf("No handlers registered for event type: %s", baseEvent.EventType)
        return nil
    }

    // 并发处理事件
    var wg sync.WaitGroup
    errChan := make(chan error, len(handlers))

    for _, handler := range handlers {
        wg.Add(1)
        go func(h EventHandler) {
            defer wg.Done()

            if err := h.Handle(ctx, &baseEvent); err != nil {
                errChan <- fmt.Errorf("handler %s failed: %w", h.GetHandlerName(), err)
            }
        }(handler)
    }

    wg.Wait()
    close(errChan)

    // 收集错误
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("event handling errors: %v", errors)
    }

    return nil
}
```

### 领域事件实现

```go
// 来自 mall-go/internal/domain/order/events.go
package order

import (
    "time"

    "mall-go/internal/event"
)

// 订单事件类型
const (
    OrderCreatedEventType   = "order.created"
    OrderPaidEventType      = "order.paid"
    OrderShippedEventType   = "order.shipped"
    OrderDeliveredEventType = "order.delivered"
    OrderCancelledEventType = "order.cancelled"
)

// 订单创建事件
type OrderCreatedEvent struct {
    *event.BaseEvent
    OrderData OrderCreatedData `json:"order_data"`
}

type OrderCreatedData struct {
    OrderID   uint                   `json:"order_id"`
    UserID    uint                   `json:"user_id"`
    Items     []OrderItem            `json:"items"`
    Amount    float64                `json:"amount"`
    Address   string                 `json:"address"`
    CreatedAt time.Time              `json:"created_at"`
}

// 创建订单创建事件
func NewOrderCreatedEvent(orderID uint, orderData OrderCreatedData) *OrderCreatedEvent {
    return &OrderCreatedEvent{
        BaseEvent: &event.BaseEvent{
            EventID:     generateEventID(),
            EventType:   OrderCreatedEventType,
            AggregateID: fmt.Sprintf("order_%d", orderID),
            Version:     1,
            Timestamp:   time.Now(),
            Metadata: map[string]interface{}{
                "source": "order-service",
                "user_id": orderData.UserID,
            },
        },
        OrderData: orderData,
    }
}

// 订单支付事件
type OrderPaidEvent struct {
    *event.BaseEvent
    PaymentData OrderPaidData `json:"payment_data"`
}

type OrderPaidData struct {
    OrderID       uint      `json:"order_id"`
    PaymentMethod string    `json:"payment_method"`
    Amount        float64   `json:"amount"`
    TransactionID string    `json:"transaction_id"`
    PaidAt        time.Time `json:"paid_at"`
}

// 创建订单支付事件
func NewOrderPaidEvent(orderID uint, paymentData OrderPaidData) *OrderPaidEvent {
    return &OrderPaidEvent{
        BaseEvent: &event.BaseEvent{
            EventID:     generateEventID(),
            EventType:   OrderPaidEventType,
            AggregateID: fmt.Sprintf("order_%d", orderID),
            Version:     2,
            Timestamp:   time.Now(),
            Metadata: map[string]interface{}{
                "source": "payment-service",
                "transaction_id": paymentData.TransactionID,
            },
        },
        PaymentData: paymentData,
    }
}

// 订单发货事件
type OrderShippedEvent struct {
    *event.BaseEvent
    ShippingData OrderShippedData `json:"shipping_data"`
}

type OrderShippedData struct {
    OrderID        uint      `json:"order_id"`
    TrackingNumber string    `json:"tracking_number"`
    Carrier        string    `json:"carrier"`
    ShippedAt      time.Time `json:"shipped_at"`
    EstimatedDelivery time.Time `json:"estimated_delivery"`
}

// 创建订单发货事件
func NewOrderShippedEvent(orderID uint, shippingData OrderShippedData) *OrderShippedEvent {
    return &OrderShippedEvent{
        BaseEvent: &event.BaseEvent{
            EventID:     generateEventID(),
            EventType:   OrderShippedEventType,
            AggregateID: fmt.Sprintf("order_%d", orderID),
            Version:     3,
            Timestamp:   time.Now(),
            Metadata: map[string]interface{}{
                "source": "shipping-service",
                "tracking_number": shippingData.TrackingNumber,
            },
        },
        ShippingData: shippingData,
    }
}

// 订单取消事件
type OrderCancelledEvent struct {
    *event.BaseEvent
    CancellationData OrderCancelledData `json:"cancellation_data"`
}

type OrderCancelledData struct {
    OrderID     uint      `json:"order_id"`
    Reason      string    `json:"reason"`
    CancelledBy string    `json:"cancelled_by"`
    CancelledAt time.Time `json:"cancelled_at"`
}

// 创建订单取消事件
func NewOrderCancelledEvent(orderID uint, cancellationData OrderCancelledData) *OrderCancelledEvent {
    return &OrderCancelledEvent{
        BaseEvent: &event.BaseEvent{
            EventID:     generateEventID(),
            EventType:   OrderCancelledEventType,
            AggregateID: fmt.Sprintf("order_%d", orderID),
            Version:     4,
            Timestamp:   time.Now(),
            Metadata: map[string]interface{}{
                "source": "order-service",
                "cancelled_by": cancellationData.CancelledBy,
            },
        },
        CancellationData: cancellationData,
    }
}

func generateEventID() string {
    return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
```

### 事件处理器实现

```go
// 来自 mall-go/internal/handler/order_event_handler.go
package handler

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "mall-go/internal/event"
    "mall-go/internal/domain/order"
    "mall-go/internal/service"
)

// 订单事件处理器
type OrderEventHandler struct {
    inventoryService    *service.InventoryService
    notificationService *service.NotificationService
    analyticsService    *service.AnalyticsService
    handlerName         string
}

// 创建订单事件处理器
func NewOrderEventHandler(
    inventoryService *service.InventoryService,
    notificationService *service.NotificationService,
    analyticsService *service.AnalyticsService,
) *OrderEventHandler {
    return &OrderEventHandler{
        inventoryService:    inventoryService,
        notificationService: notificationService,
        analyticsService:    analyticsService,
        handlerName:         "OrderEventHandler",
    }
}

// 获取处理器名称
func (h *OrderEventHandler) GetHandlerName() string {
    return h.handlerName
}

// 处理事件
func (h *OrderEventHandler) Handle(ctx context.Context, event event.Event) error {
    switch event.GetEventType() {
    case order.OrderCreatedEventType:
        return h.handleOrderCreated(ctx, event)
    case order.OrderPaidEventType:
        return h.handleOrderPaid(ctx, event)
    case order.OrderShippedEventType:
        return h.handleOrderShipped(ctx, event)
    case order.OrderCancelledEventType:
        return h.handleOrderCancelled(ctx, event)
    default:
        return fmt.Errorf("unsupported event type: %s", event.GetEventType())
    }
}

// 处理订单创建事件
func (h *OrderEventHandler) handleOrderCreated(ctx context.Context, event event.Event) error {
    var orderCreatedEvent order.OrderCreatedEvent
    if err := h.unmarshalEvent(event, &orderCreatedEvent); err != nil {
        return err
    }

    log.Printf("Handling order created event: OrderID=%d", orderCreatedEvent.OrderData.OrderID)

    // 1. 更新库存
    if err := h.updateInventory(ctx, orderCreatedEvent.OrderData); err != nil {
        return fmt.Errorf("failed to update inventory: %w", err)
    }

    // 2. 发送通知
    if err := h.sendOrderCreatedNotification(ctx, orderCreatedEvent.OrderData); err != nil {
        log.Printf("Failed to send notification: %v", err)
        // 通知失败不影响主流程
    }

    // 3. 记录分析数据
    if err := h.recordOrderAnalytics(ctx, orderCreatedEvent.OrderData); err != nil {
        log.Printf("Failed to record analytics: %v", err)
        // 分析数据记录失败不影响主流程
    }

    return nil
}

// 处理订单支付事件
func (h *OrderEventHandler) handleOrderPaid(ctx context.Context, event event.Event) error {
    var orderPaidEvent order.OrderPaidEvent
    if err := h.unmarshalEvent(event, &orderPaidEvent); err != nil {
        return err
    }

    log.Printf("Handling order paid event: OrderID=%d", orderPaidEvent.PaymentData.OrderID)

    // 1. 触发发货流程
    if err := h.triggerShipping(ctx, orderPaidEvent.PaymentData); err != nil {
        return fmt.Errorf("failed to trigger shipping: %w", err)
    }

    // 2. 发送支付成功通知
    if err := h.sendPaymentSuccessNotification(ctx, orderPaidEvent.PaymentData); err != nil {
        log.Printf("Failed to send payment notification: %v", err)
    }

    // 3. 更新用户积分
    if err := h.updateUserPoints(ctx, orderPaidEvent.PaymentData); err != nil {
        log.Printf("Failed to update user points: %v", err)
    }

    return nil
}

// 处理订单发货事件
func (h *OrderEventHandler) handleOrderShipped(ctx context.Context, event event.Event) error {
    var orderShippedEvent order.OrderShippedEvent
    if err := h.unmarshalEvent(event, &orderShippedEvent); err != nil {
        return err
    }

    log.Printf("Handling order shipped event: OrderID=%d", orderShippedEvent.ShippingData.OrderID)

    // 1. 发送发货通知
    if err := h.sendShippingNotification(ctx, orderShippedEvent.ShippingData); err != nil {
        log.Printf("Failed to send shipping notification: %v", err)
    }

    // 2. 启动物流跟踪
    if err := h.startLogisticsTracking(ctx, orderShippedEvent.ShippingData); err != nil {
        log.Printf("Failed to start logistics tracking: %v", err)
    }

    return nil
}

// 处理订单取消事件
func (h *OrderEventHandler) handleOrderCancelled(ctx context.Context, event event.Event) error {
    var orderCancelledEvent order.OrderCancelledEvent
    if err := h.unmarshalEvent(event, &orderCancelledEvent); err != nil {
        return err
    }

    log.Printf("Handling order cancelled event: OrderID=%d", orderCancelledEvent.CancellationData.OrderID)

    // 1. 恢复库存
    if err := h.restoreInventory(ctx, orderCancelledEvent.CancellationData); err != nil {
        return fmt.Errorf("failed to restore inventory: %w", err)
    }

    // 2. 处理退款
    if err := h.processRefund(ctx, orderCancelledEvent.CancellationData); err != nil {
        return fmt.Errorf("failed to process refund: %w", err)
    }

    // 3. 发送取消通知
    if err := h.sendCancellationNotification(ctx, orderCancelledEvent.CancellationData); err != nil {
        log.Printf("Failed to send cancellation notification: %v", err)
    }

    return nil
}

// 辅助方法
func (h *OrderEventHandler) unmarshalEvent(event event.Event, target interface{}) error {
    eventBytes, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    if err := json.Unmarshal(eventBytes, target); err != nil {
        return fmt.Errorf("failed to unmarshal event: %w", err)
    }

    return nil
}

func (h *OrderEventHandler) updateInventory(ctx context.Context, orderData order.OrderCreatedData) error {
    return h.inventoryService.UpdateStock(ctx, orderData.Items)
}

func (h *OrderEventHandler) sendOrderCreatedNotification(ctx context.Context, orderData order.OrderCreatedData) error {
    return h.notificationService.SendOrderCreatedNotification(ctx, orderData.UserID, orderData.OrderID)
}

func (h *OrderEventHandler) recordOrderAnalytics(ctx context.Context, orderData order.OrderCreatedData) error {
    return h.analyticsService.RecordOrderCreated(ctx, orderData)
}

func (h *OrderEventHandler) triggerShipping(ctx context.Context, paymentData order.OrderPaidData) error {
    // 模拟触发发货
    log.Printf("Triggering shipping for order: %d", paymentData.OrderID)
    return nil
}

func (h *OrderEventHandler) sendPaymentSuccessNotification(ctx context.Context, paymentData order.OrderPaidData) error {
    // 模拟发送支付成功通知
    log.Printf("Sending payment success notification for order: %d", paymentData.OrderID)
    return nil
}

func (h *OrderEventHandler) updateUserPoints(ctx context.Context, paymentData order.OrderPaidData) error {
    // 模拟更新用户积分
    log.Printf("Updating user points for order: %d", paymentData.OrderID)
    return nil
}

func (h *OrderEventHandler) sendShippingNotification(ctx context.Context, shippingData order.OrderShippedData) error {
    // 模拟发送发货通知
    log.Printf("Sending shipping notification for order: %d", shippingData.OrderID)
    return nil
}

func (h *OrderEventHandler) startLogisticsTracking(ctx context.Context, shippingData order.OrderShippedData) error {
    // 模拟启动物流跟踪
    log.Printf("Starting logistics tracking for order: %d", shippingData.OrderID)
    return nil
}

func (h *OrderEventHandler) restoreInventory(ctx context.Context, cancellationData order.OrderCancelledData) error {
    // 模拟恢复库存
    log.Printf("Restoring inventory for order: %d", cancellationData.OrderID)
    return nil
}

func (h *OrderEventHandler) processRefund(ctx context.Context, cancellationData order.OrderCancelledData) error {
    // 模拟处理退款
    log.Printf("Processing refund for order: %d", cancellationData.OrderID)
    return nil
}

func (h *OrderEventHandler) sendCancellationNotification(ctx context.Context, cancellationData order.OrderCancelledData) error {
    // 模拟发送取消通知
    log.Printf("Sending cancellation notification for order: %d", cancellationData.OrderID)
    return nil
}
```

---

## 🎯 面试常考点

### 1. 消息队列基础概念

**问题：** 什么是消息队列？有什么优势和劣势？

**答案：**
```go
/*
消息队列（Message Queue）是一种应用程序间的通信方法，通过在消息的传输过程中保存消息来实现应用程序间的异步通信。

优势：
✅ 系统解耦：生产者和消费者不需要直接交互
✅ 异步处理：提高系统响应速度和吞吐量
✅ 流量削峰：缓解系统压力，防止系统过载
✅ 可靠性：消息持久化，保证消息不丢失
✅ 扩展性：易于水平扩展和负载均衡

劣势：
❌ 系统复杂性：增加了系统的复杂度
❌ 一致性问题：异步处理可能导致数据一致性问题
❌ 消息重复：网络问题可能导致消息重复消费
❌ 消息顺序：分布式环境下难以保证消息顺序
❌ 调试困难：异步处理增加了调试难度
*/

// 消息队列的核心组件
type MessageQueueComponents struct {
    Producer  Producer  // 生产者：发送消息
    Consumer  Consumer  // 消费者：接收和处理消息
    Broker    Broker    // 代理：存储和转发消息
    Topic     Topic     // 主题：消息分类
    Queue     Queue     // 队列：消息存储
    Exchange  Exchange  // 交换机：消息路由（RabbitMQ）
}

// 消息传递模式
type MessagePatterns struct {
    PointToPoint    string // 点对点：一对一消息传递
    PublishSubscribe string // 发布订阅：一对多消息传递
    RequestReply    string // 请求响应：同步消息传递
    MessageRouting  string // 消息路由：基于规则的消息分发
}
```

### 2. 消息可靠性保证

**问题：** 如何保证消息的可靠性？消息丢失的场景有哪些？

**答案：**
```go
/*
消息丢失的三个阶段：
1. 生产者发送消息时丢失
2. 消息队列存储时丢失
3. 消费者处理消息时丢失

解决方案：
*/

// 1. 生产者确认机制
type ProducerConfirmation struct {
    // RabbitMQ确认机制
    ConfirmMode bool `json:"confirm_mode"` // 开启确认模式

    // Kafka确认机制
    RequiredAcks int `json:"required_acks"` // 0=不等待, 1=等待leader, -1=等待所有副本

    // 重试机制
    RetryCount int           `json:"retry_count"`
    RetryDelay time.Duration `json:"retry_delay"`
}

// 生产者确认示例
func (p *Producer) SendWithConfirmation(ctx context.Context, message *Message) error {
    // 开启确认模式
    if err := p.channel.Confirm(false); err != nil {
        return err
    }

    // 监听确认
    confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

    // 发送消息
    if err := p.channel.Publish("exchange", "routing.key", false, false, amqp.Publishing{
        Body: message.Body,
        DeliveryMode: 2, // 持久化
    }); err != nil {
        return err
    }

    // 等待确认
    select {
    case confirm := <-confirms:
        if !confirm.Ack {
            return fmt.Errorf("message not acknowledged")
        }
        return nil
    case <-time.After(5 * time.Second):
        return fmt.Errorf("confirmation timeout")
    }
}

// 2. 消息持久化
type MessagePersistence struct {
    // RabbitMQ持久化
    DurableQueue    bool `json:"durable_queue"`    // 队列持久化
    DurableExchange bool `json:"durable_exchange"` // 交换机持久化
    PersistentMessage bool `json:"persistent_message"` // 消息持久化

    // Kafka持久化
    ReplicationFactor int `json:"replication_factor"` // 副本因子
    MinInSyncReplicas int `json:"min_in_sync_replicas"` // 最小同步副本数
}

// 3. 消费者确认机制
type ConsumerAcknowledgment struct {
    AutoAck    bool `json:"auto_ack"`    // 自动确认
    ManualAck  bool `json:"manual_ack"`  // 手动确认
    RejectRequeue bool `json:"reject_requeue"` // 拒绝并重新入队
}

// 消费者手动确认示例
func (c *Consumer) ProcessMessage(delivery amqp.Delivery) {
    defer func() {
        if r := recover(); r != nil {
            // 处理异常，拒绝消息并重新入队
            delivery.Nack(false, true)
        }
    }()

    // 处理消息
    if err := c.handleMessage(delivery.Body); err != nil {
        // 处理失败，拒绝消息
        delivery.Nack(false, false) // 不重新入队，发送到死信队列
        return
    }

    // 处理成功，确认消息
    delivery.Ack(false)
}

// 4. 死信队列处理
type DeadLetterQueue struct {
    Exchange   string `json:"exchange"`
    RoutingKey string `json:"routing_key"`
    TTL        int    `json:"ttl"` // 消息存活时间
}

// 死信队列配置
func ConfigureDeadLetterQueue(ch *amqp.Channel) error {
    // 声明死信交换机
    if err := ch.ExchangeDeclare("dlx", "direct", true, false, false, false, nil); err != nil {
        return err
    }

    // 声明主队列，配置死信交换机
    _, err := ch.QueueDeclare("main.queue", true, false, false, false, amqp.Table{
        "x-dead-letter-exchange":    "dlx",
        "x-dead-letter-routing-key": "dead.letter",
        "x-message-ttl":             300000, // 5分钟TTL
    })

    return err
}
```

### 3. 消息重复和幂等性

**问题：** 如何处理消息重复消费？如何保证消费的幂等性？

**答案：**
```go
/*
消息重复的原因：
1. 网络问题导致确认丢失
2. 消费者处理超时
3. 系统重启或故障恢复
4. 负载均衡导致的重复投递

解决方案：
*/

// 1. 消息去重
type MessageDeduplication struct {
    MessageID string    `json:"message_id"` // 全局唯一消息ID
    Timestamp int64     `json:"timestamp"`  // 消息时间戳
    Hash      string    `json:"hash"`       // 消息内容哈希
}

// 基于Redis的消息去重
type RedisDeduplicator struct {
    rdb *redis.Client
    ttl time.Duration
}

func (d *RedisDeduplicator) IsDuplicate(ctx context.Context, messageID string) (bool, error) {
    key := fmt.Sprintf("msg:processed:%s", messageID)

    // 使用SET NX命令实现原子性检查和设置
    result, err := d.rdb.SetNX(ctx, key, "1", d.ttl).Result()
    if err != nil {
        return false, err
    }

    // 如果设置成功，说明是第一次处理
    return !result, nil
}

// 消费者去重处理
func (c *Consumer) ProcessMessageWithDeduplication(ctx context.Context, delivery amqp.Delivery) error {
    var message struct {
        ID      string      `json:"id"`
        Content interface{} `json:"content"`
    }

    if err := json.Unmarshal(delivery.Body, &message); err != nil {
        return err
    }

    // 检查是否重复
    isDuplicate, err := c.deduplicator.IsDuplicate(ctx, message.ID)
    if err != nil {
        return err
    }

    if isDuplicate {
        log.Printf("Duplicate message detected: %s", message.ID)
        delivery.Ack(false) // 确认消息，避免重复处理
        return nil
    }

    // 处理消息
    if err := c.handleMessage(ctx, message.Content); err != nil {
        return err
    }

    delivery.Ack(false)
    return nil
}

// 2. 幂等性设计
type IdempotentOperation struct {
    OperationID string `json:"operation_id"` // 操作唯一标识
    Version     int    `json:"version"`      // 版本号
    Status      string `json:"status"`       // 操作状态
}

// 幂等性处理示例
func (s *OrderService) ProcessPayment(ctx context.Context, orderID uint, amount float64, operationID string) error {
    // 检查操作是否已执行
    operation, err := s.getOperation(ctx, operationID)
    if err != nil && err != ErrOperationNotFound {
        return err
    }

    if operation != nil {
        switch operation.Status {
        case "completed":
            return nil // 已完成，直接返回
        case "processing":
            return ErrOperationInProgress
        case "failed":
            // 可以重试
            break
        }
    }

    // 记录操作开始
    if err := s.recordOperation(ctx, operationID, "processing"); err != nil {
        return err
    }

    // 执行支付
    if err := s.doPayment(ctx, orderID, amount); err != nil {
        s.recordOperation(ctx, operationID, "failed")
        return err
    }

    // 记录操作完成
    return s.recordOperation(ctx, operationID, "completed")
}

// 3. 基于数据库约束的幂等性
type PaymentRecord struct {
    ID          uint      `gorm:"primaryKey"`
    OrderID     uint      `gorm:"not null"`
    Amount      float64   `gorm:"not null"`
    OperationID string    `gorm:"uniqueIndex;not null"` // 唯一约束
    Status      string    `gorm:"not null"`
    CreatedAt   time.Time
}

// 利用数据库唯一约束保证幂等性
func (s *PaymentService) ProcessPaymentIdempotent(ctx context.Context, orderID uint, amount float64, operationID string) error {
    payment := &PaymentRecord{
        OrderID:     orderID,
        Amount:      amount,
        OperationID: operationID,
        Status:      "processing",
        CreatedAt:   time.Now(),
    }

    // 插入记录，如果operationID重复会失败
    if err := s.db.Create(payment).Error; err != nil {
        if isUniqueConstraintError(err) {
            // 检查已存在记录的状态
            var existing PaymentRecord
            if err := s.db.Where("operation_id = ?", operationID).First(&existing).Error; err != nil {
                return err
            }

            if existing.Status == "completed" {
                return nil // 已完成
            }

            return ErrOperationInProgress
        }
        return err
    }

    // 执行支付逻辑
    if err := s.doPayment(ctx, orderID, amount); err != nil {
        // 更新状态为失败
        s.db.Model(payment).Update("status", "failed")
        return err
    }

    // 更新状态为完成
    return s.db.Model(payment).Update("status", "completed").Error
}
```

### 4. 消息顺序性保证

**问题：** 如何保证消息的顺序性？

**答案：**
```go
/*
消息顺序性的挑战：
1. 分布式环境下的并发处理
2. 多个生产者同时发送消息
3. 消息重试可能打乱顺序
4. 负载均衡导致的乱序

解决方案：
*/

// 1. 单分区/单队列保证顺序
type OrderedMessageProducer struct {
    producer *kafka.Producer
}

// 基于分区键保证顺序
func (p *OrderedMessageProducer) SendOrderedMessage(ctx context.Context, topic string, key string, message interface{}) error {
    // 使用相同的key确保消息发送到同一分区
    kafkaMsg := &kafka.KafkaMessage{
        Key:   key, // 相同业务的消息使用相同key
        Value: message,
        Timestamp: time.Now(),
    }

    return p.producer.Send(ctx, topic, kafkaMsg)
}

// 2. 消息序列号
type SequencedMessage struct {
    SequenceID int64       `json:"sequence_id"`
    BusinessID string      `json:"business_id"` // 业务标识
    Content    interface{} `json:"content"`
    Timestamp  int64       `json:"timestamp"`
}

// 顺序消费者
type OrderedConsumer struct {
    consumer        *kafka.Consumer
    sequenceTracker map[string]int64 // 跟踪每个业务的序列号
    mutex           sync.RWMutex
}

func (c *OrderedConsumer) ProcessOrderedMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
    var seqMsg SequencedMessage
    if err := json.Unmarshal(message.Value, &seqMsg); err != nil {
        return err
    }

    c.mutex.Lock()
    defer c.mutex.Unlock()

    expectedSeq := c.sequenceTracker[seqMsg.BusinessID] + 1

    if seqMsg.SequenceID < expectedSeq {
        // 重复消息，忽略
        log.Printf("Duplicate message: expected=%d, got=%d", expectedSeq, seqMsg.SequenceID)
        return nil
    }

    if seqMsg.SequenceID > expectedSeq {
        // 消息乱序，需要等待或重新排序
        log.Printf("Out of order message: expected=%d, got=%d", expectedSeq, seqMsg.SequenceID)
        return c.handleOutOfOrderMessage(ctx, &seqMsg)
    }

    // 正确顺序，处理消息
    if err := c.handleMessage(ctx, seqMsg.Content); err != nil {
        return err
    }

    // 更新序列号
    c.sequenceTracker[seqMsg.BusinessID] = seqMsg.SequenceID
    return nil
}

// 3. 基于时间戳的排序
type TimestampOrderedConsumer struct {
    consumer    *kafka.Consumer
    buffer      map[string][]*TimestampedMessage // 缓冲区
    bufferSize  int
    waitTimeout time.Duration
}

type TimestampedMessage struct {
    BusinessID string      `json:"business_id"`
    Timestamp  int64       `json:"timestamp"`
    Content    interface{} `json:"content"`
}

func (c *TimestampOrderedConsumer) ProcessTimestampedMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
    var tsMsg TimestampedMessage
    if err := json.Unmarshal(message.Value, &tsMsg); err != nil {
        return err
    }

    // 添加到缓冲区
    c.addToBuffer(&tsMsg)

    // 检查是否可以处理缓冲区中的消息
    return c.processBufferedMessages(ctx, tsMsg.BusinessID)
}

func (c *TimestampOrderedConsumer) addToBuffer(msg *TimestampedMessage) {
    if c.buffer[msg.BusinessID] == nil {
        c.buffer[msg.BusinessID] = make([]*TimestampedMessage, 0)
    }

    // 按时间戳排序插入
    buffer := c.buffer[msg.BusinessID]
    insertIndex := len(buffer)

    for i, bufferedMsg := range buffer {
        if msg.Timestamp < bufferedMsg.Timestamp {
            insertIndex = i
            break
        }
    }

    // 插入消息
    buffer = append(buffer, nil)
    copy(buffer[insertIndex+1:], buffer[insertIndex:])
    buffer[insertIndex] = msg

    c.buffer[msg.BusinessID] = buffer
}

func (c *TimestampOrderedConsumer) processBufferedMessages(ctx context.Context, businessID string) error {
    buffer := c.buffer[businessID]
    if len(buffer) == 0 {
        return nil
    }

    now := time.Now().UnixMilli()
    processedCount := 0

    // 处理超过等待时间的消息
    for i, msg := range buffer {
        if now-msg.Timestamp > c.waitTimeout.Milliseconds() || len(buffer) > c.bufferSize {
            if err := c.handleMessage(ctx, msg.Content); err != nil {
                return err
            }
            processedCount = i + 1
        } else {
            break
        }
    }

    // 移除已处理的消息
    if processedCount > 0 {
        c.buffer[businessID] = buffer[processedCount:]
    }

    return nil
}
```

---

## ⚠️ 踩坑提醒

### 1. 连接管理陷阱

```go
// ❌ 错误：每次发送消息都创建新连接
func BadConnectionManagement() {
    for i := 0; i < 1000; i++ {
        // 每次都创建新连接，资源浪费严重
        conn, _ := amqp.Dial("amqp://localhost")
        ch, _ := conn.Channel()

        ch.Publish("exchange", "key", false, false, amqp.Publishing{
            Body: []byte("message"),
        })

        ch.Close()
        conn.Close()
    }
}

// ✅ 正确：使用连接池管理连接
type ConnectionPool struct {
    connections chan *amqp.Connection
    maxSize     int
    url         string
}

func NewConnectionPool(url string, maxSize int) *ConnectionPool {
    pool := &ConnectionPool{
        connections: make(chan *amqp.Connection, maxSize),
        maxSize:     maxSize,
        url:         url,
    }

    // 预创建连接
    for i := 0; i < maxSize; i++ {
        conn, _ := amqp.Dial(url)
        pool.connections <- conn
    }

    return pool
}

func (p *ConnectionPool) GetConnection() *amqp.Connection {
    return <-p.connections
}

func (p *ConnectionPool) ReturnConnection(conn *amqp.Connection) {
    if !conn.IsClosed() {
        p.connections <- conn
    } else {
        // 连接已关闭，创建新连接
        newConn, _ := amqp.Dial(p.url)
        p.connections <- newConn
    }
}
```

### 2. 消息序列化陷阱

```go
// ❌ 错误：直接序列化复杂对象
type ComplexMessage struct {
    Data     map[string]interface{} `json:"data"`
    Callback func()                 `json:"-"` // 函数无法序列化
    Channel  chan string            `json:"-"` // 通道无法序列化
    Mutex    sync.Mutex             `json:"-"` // 互斥锁无法序列化
}

func BadSerialization(msg *ComplexMessage) error {
    // 这会导致序列化失败或数据丢失
    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    // 发送消息...
    return nil
}

// ✅ 正确：设计专门的消息结构
type MessagePayload struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Data      map[string]interface{} `json:"data"`
    Timestamp int64                  `json:"timestamp"`
    Version   string                 `json:"version"`
}

func GoodSerialization(businessData interface{}) error {
    // 转换为消息格式
    payload := &MessagePayload{
        ID:        generateID(),
        Type:      "business.event",
        Data:      convertToMap(businessData),
        Timestamp: time.Now().Unix(),
        Version:   "1.0",
    }

    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    // 发送消息...
    return nil
}

func convertToMap(data interface{}) map[string]interface{} {
    result := make(map[string]interface{})

    // 使用反射或手动转换
    bytes, _ := json.Marshal(data)
    json.Unmarshal(bytes, &result)

    return result
}
```

### 3. 错误处理陷阱

```go
// ❌ 错误：忽略错误或简单重试
func BadErrorHandling(consumer *Consumer) {
    consumer.RegisterHandler("topic", func(ctx context.Context, msg *Message) error {
        if err := processMessage(msg); err != nil {
            // 简单忽略错误
            log.Printf("Error: %v", err)
            return nil // 这会导致消息丢失
        }
        return nil
    })
}

// ✅ 正确：分类处理不同类型的错误
type ErrorHandler struct {
    retryableErrors map[string]bool
    maxRetries      int
    deadLetterQueue string
}

func (h *ErrorHandler) HandleError(ctx context.Context, msg *Message, err error) error {
    // 分析错误类型
    errorType := classifyError(err)

    switch errorType {
    case "temporary":
        // 临时错误，可以重试
        return h.handleRetryableError(ctx, msg, err)
    case "permanent":
        // 永久错误，发送到死信队列
        return h.sendToDeadLetterQueue(ctx, msg, err)
    case "poison":
        // 毒消息，记录并丢弃
        return h.handlePoisonMessage(ctx, msg, err)
    default:
        // 未知错误，保守处理
        return h.handleUnknownError(ctx, msg, err)
    }
}

func (h *ErrorHandler) handleRetryableError(ctx context.Context, msg *Message, err error) error {
    if msg.RetryCount >= h.maxRetries {
        // 超过最大重试次数，发送到死信队列
        return h.sendToDeadLetterQueue(ctx, msg, err)
    }

    // 增加重试次数
    msg.RetryCount++

    // 计算退避时间
    backoffTime := time.Duration(msg.RetryCount*msg.RetryCount) * time.Second

    // 延迟重试
    time.Sleep(backoffTime)

    return err // 返回错误，触发重试
}

func classifyError(err error) string {
    switch {
    case isNetworkError(err):
        return "temporary"
    case isValidationError(err):
        return "permanent"
    case isSerializationError(err):
        return "poison"
    default:
        return "unknown"
    }
}
```

### 4. 内存泄漏陷阱

```go
// ❌ 错误：不正确的资源管理
func BadResourceManagement() {
    consumers := make([]*Consumer, 0)

    for i := 0; i < 100; i++ {
        consumer := createConsumer()
        consumers = append(consumers, consumer)

        // 启动消费者但没有正确关闭
        consumer.Start()
    }

    // 程序结束时没有清理资源
}

// ✅ 正确：使用资源管理器
type ResourceManager struct {
    consumers []io.Closer
    producers []io.Closer
    mutex     sync.Mutex
}

func NewResourceManager() *ResourceManager {
    return &ResourceManager{
        consumers: make([]io.Closer, 0),
        producers: make([]io.Closer, 0),
    }
}

func (rm *ResourceManager) AddConsumer(consumer io.Closer) {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    rm.consumers = append(rm.consumers, consumer)
}

func (rm *ResourceManager) AddProducer(producer io.Closer) {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    rm.producers = append(rm.producers, producer)
}

func (rm *ResourceManager) Close() error {
    rm.mutex.Lock()
    defer rm.mutex.Unlock()

    var errors []error

    // 关闭所有消费者
    for _, consumer := range rm.consumers {
        if err := consumer.Close(); err != nil {
            errors = append(errors, err)
        }
    }

    // 关闭所有生产者
    for _, producer := range rm.producers {
        if err := producer.Close(); err != nil {
            errors = append(errors, err)
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("errors closing resources: %v", errors)
    }

    return nil
}

// 使用defer确保资源清理
func GoodResourceManagement() {
    rm := NewResourceManager()
    defer rm.Close() // 确保资源被清理

    for i := 0; i < 100; i++ {
        consumer := createConsumer()
        rm.AddConsumer(consumer)
        consumer.Start()
    }
}
```

### 5. 性能优化陷阱

```go
// ❌ 错误：同步处理所有消息
func BadPerformance(messages []*Message) {
    for _, msg := range messages {
        // 同步处理，性能差
        processMessage(msg)
    }
}

// ✅ 正确：批量和并发处理
type BatchProcessor struct {
    batchSize   int
    workerCount int
    timeout     time.Duration
}

func (bp *BatchProcessor) ProcessMessages(messages []*Message) error {
    // 分批处理
    batches := bp.createBatches(messages)

    // 并发处理批次
    var wg sync.WaitGroup
    errChan := make(chan error, len(batches))

    // 限制并发数
    semaphore := make(chan struct{}, bp.workerCount)

    for _, batch := range batches {
        wg.Add(1)
        go func(b []*Message) {
            defer wg.Done()

            // 获取信号量
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            if err := bp.processBatch(b); err != nil {
                errChan <- err
            }
        }(batch)
    }

    wg.Wait()
    close(errChan)

    // 收集错误
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        return fmt.Errorf("batch processing errors: %v", errors)
    }

    return nil
}

func (bp *BatchProcessor) createBatches(messages []*Message) [][]*Message {
    var batches [][]*Message

    for i := 0; i < len(messages); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(messages) {
            end = len(messages)
        }
        batches = append(batches, messages[i:end])
    }

    return batches
}

func (bp *BatchProcessor) processBatch(batch []*Message) error {
    ctx, cancel := context.WithTimeout(context.Background(), bp.timeout)
    defer cancel()

    // 批量处理消息
    for _, msg := range batch {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := processMessage(msg); err != nil {
                return err
            }
        }
    }

    return nil
}
```

---

## 📝 练习题

### 练习题1：分布式事务消息实现（⭐⭐⭐）

**题目描述：**
实现一个基于消息队列的分布式事务解决方案，支持两阶段提交协议，确保消息发送和本地事务的一致性。

```go
// 练习题1：分布式事务消息实现
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    "github.com/streadway/amqp"
    "gorm.io/gorm"
)

// 解答：
// 事务消息状态
type TransactionMessageStatus string

const (
    StatusPrepared  TransactionMessageStatus = "prepared"
    StatusCommitted TransactionMessageStatus = "committed"
    StatusRollback  TransactionMessageStatus = "rollback"
)

// 事务消息记录
type TransactionMessage struct {
    ID          string                    `gorm:"primaryKey"`
    Topic       string                    `gorm:"not null"`
    RoutingKey  string                    `gorm:"not null"`
    Payload     string                    `gorm:"type:text;not null"`
    Status      TransactionMessageStatus `gorm:"not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ExpiredAt   time.Time `gorm:"index"`
}

// 分布式事务消息管理器
type TransactionMessageManager struct {
    db       *gorm.DB
    producer *amqp.Channel
    consumer *amqp.Channel
}

func NewTransactionMessageManager(db *gorm.DB, conn *amqp.Connection) (*TransactionMessageManager, error) {
    producer, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    consumer, err := conn.Channel()
    if err != nil {
        return nil, err
    }

    manager := &TransactionMessageManager{
        db:       db,
        producer: producer,
        consumer: consumer,
    }

    // 启动定时任务
    go manager.startCleanupTask()
    go manager.startRetryTask()

    return manager, nil
}

// 第一阶段：准备事务消息
func (tm *TransactionMessageManager) PrepareMessage(ctx context.Context, topic, routingKey string, payload interface{}) (string, error) {
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("failed to marshal payload: %w", err)
    }

    messageID := generateMessageID()
    transactionMsg := &TransactionMessage{
        ID:         messageID,
        Topic:      topic,
        RoutingKey: routingKey,
        Payload:    string(payloadBytes),
        Status:     StatusPrepared,
        CreatedAt:  time.Now(),
        ExpiredAt:  time.Now().Add(5 * time.Minute), // 5分钟超时
    }

    if err := tm.db.Create(transactionMsg).Error; err != nil {
        return "", fmt.Errorf("failed to prepare transaction message: %w", err)
    }

    return messageID, nil
}

// 第二阶段：提交事务消息
func (tm *TransactionMessageManager) CommitMessage(ctx context.Context, messageID string) error {
    return tm.db.Transaction(func(tx *gorm.DB) error {
        var transactionMsg TransactionMessage
        if err := tx.Where("id = ? AND status = ?", messageID, StatusPrepared).First(&transactionMsg).Error; err != nil {
            return fmt.Errorf("transaction message not found or already processed: %w", err)
        }

        // 更新状态为已提交
        if err := tx.Model(&transactionMsg).Update("status", StatusCommitted).Error; err != nil {
            return fmt.Errorf("failed to commit transaction message: %w", err)
        }

        // 发送消息到MQ
        if err := tm.publishMessage(&transactionMsg); err != nil {
            return fmt.Errorf("failed to publish message: %w", err)
        }

        return nil
    })
}

// 第二阶段：回滚事务消息
func (tm *TransactionMessageManager) RollbackMessage(ctx context.Context, messageID string) error {
    return tm.db.Model(&TransactionMessage{}).
        Where("id = ? AND status = ?", messageID, StatusPrepared).
        Update("status", StatusRollback).Error
}

// 发布消息到MQ
func (tm *TransactionMessageManager) publishMessage(transactionMsg *TransactionMessage) error {
    return tm.producer.Publish(
        transactionMsg.Topic,
        transactionMsg.RoutingKey,
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         []byte(transactionMsg.Payload),
            DeliveryMode: 2, // 持久化
            Headers: amqp.Table{
                "transaction_id": transactionMsg.ID,
                "timestamp":      time.Now().Unix(),
            },
        },
    )
}

// 业务服务示例
type OrderService struct {
    db  *gorm.DB
    txm *TransactionMessageManager
}

// 创建订单（本地事务 + 事务消息）
func (s *OrderService) CreateOrder(ctx context.Context, order *Order) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 准备事务消息
        messageID, err := s.txm.PrepareMessage(ctx, "order.events", "order.created", map[string]interface{}{
            "order_id": order.ID,
            "user_id":  order.UserID,
            "amount":   order.Amount,
        })
        if err != nil {
            return err
        }

        // 2. 执行本地事务
        if err := tx.Create(order).Error; err != nil {
            // 本地事务失败，回滚事务消息
            s.txm.RollbackMessage(ctx, messageID)
            return err
        }

        // 3. 本地事务成功，提交事务消息
        if err := s.txm.CommitMessage(ctx, messageID); err != nil {
            return err
        }

        return nil
    })
}

// 定时清理过期的事务消息
func (tm *TransactionMessageManager) startCleanupTask() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        // 清理过期的准备状态消息
        tm.db.Where("status = ? AND expired_at < ?", StatusPrepared, time.Now()).
            Update("status", StatusRollback)

        // 删除过期的已处理消息
        tm.db.Where("status IN ? AND created_at < ?",
                   []TransactionMessageStatus{StatusCommitted, StatusRollback},
                   time.Now().Add(-24*time.Hour)).
            Delete(&TransactionMessage{})
    }
}

// 定时重试失败的消息
func (tm *TransactionMessageManager) startRetryTask() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        var messages []TransactionMessage
        tm.db.Where("status = ? AND updated_at < ?", StatusCommitted, time.Now().Add(-time.Minute)).
            Find(&messages)

        for _, msg := range messages {
            if err := tm.publishMessage(&msg); err != nil {
                log.Printf("Failed to retry message %s: %v", msg.ID, err)
            } else {
                log.Printf("Successfully retried message %s", msg.ID)
            }
        }
    }
}

/*
解析说明：
1. 两阶段提交：先准备事务消息，再根据本地事务结果决定提交或回滚
2. 消息表：使用数据库表记录事务消息状态，确保一致性
3. 定时任务：清理过期消息和重试失败消息
4. 事务边界：本地事务和消息发送在同一个事务中

扩展思考：
- 如何处理消息发送失败的情况？
- 如何优化大量事务消息的性能？
- 如何实现消息的幂等性消费？
- 如何监控事务消息的健康状态？
*/
```

### 练习题2：消息队列监控系统（⭐⭐）

**题目描述：**
设计并实现一个消息队列监控系统，能够实时监控消息队列的健康状态、性能指标和异常情况。

```go
// 练习题2：消息队列监控系统
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 解答：
// 监控指标定义
var (
    // 消息计数器
    messagesProduced = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mq_messages_produced_total",
            Help: "Total number of messages produced",
        },
        []string{"topic", "status"},
    )

    messagesConsumed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mq_messages_consumed_total",
            Help: "Total number of messages consumed",
        },
        []string{"topic", "consumer_group", "status"},
    )

    // 消息处理延迟
    messageProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "mq_message_processing_duration_seconds",
            Help:    "Message processing duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"topic", "consumer_group"},
    )

    // 队列深度
    queueDepth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mq_queue_depth",
            Help: "Current queue depth",
        },
        []string{"queue"},
    )

    // 连接状态
    connectionStatus = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mq_connection_status",
            Help: "Connection status (1=connected, 0=disconnected)",
        },
        []string{"broker", "type"},
    )
)

// 监控数据结构
type MonitoringData struct {
    Timestamp         time.Time              `json:"timestamp"`
    Topic             string                 `json:"topic"`
    ConsumerGroup     string                 `json:"consumer_group,omitempty"`
    MessageCount      int64                  `json:"message_count"`
    ProcessingTime    time.Duration          `json:"processing_time"`
    ErrorCount        int64                  `json:"error_count"`
    QueueDepth        int64                  `json:"queue_depth"`
    ConnectionStatus  bool                   `json:"connection_status"`
    CustomMetrics     map[string]interface{} `json:"custom_metrics,omitempty"`
}

// 监控系统
type MonitoringSystem struct {
    metrics     map[string]*MonitoringData
    mutex       sync.RWMutex
    alertRules  []AlertRule
    collectors  []MetricCollector
    reporters   []MetricReporter
}

// 告警规则
type AlertRule struct {
    Name        string                 `json:"name"`
    Condition   string                 `json:"condition"`
    Threshold   float64                `json:"threshold"`
    Duration    time.Duration          `json:"duration"`
    Severity    string                 `json:"severity"`
    Actions     []AlertAction          `json:"actions"`
    LastTriggered time.Time            `json:"last_triggered"`
}

type AlertAction struct {
    Type   string                 `json:"type"`   // email, webhook, slack
    Config map[string]interface{} `json:"config"`
}

// 指标收集器接口
type MetricCollector interface {
    Collect(ctx context.Context) (*MonitoringData, error)
    GetName() string
}

// 指标报告器接口
type MetricReporter interface {
    Report(ctx context.Context, data *MonitoringData) error
    GetName() string
}

// 创建监控系统
func NewMonitoringSystem() *MonitoringSystem {
    return &MonitoringSystem{
        metrics:    make(map[string]*MonitoringData),
        alertRules: make([]AlertRule, 0),
        collectors: make([]MetricCollector, 0),
        reporters:  make([]MetricReporter, 0),
    }
}

// 注册指标收集器
func (ms *MonitoringSystem) RegisterCollector(collector MetricCollector) {
    ms.collectors = append(ms.collectors, collector)
}

// 注册指标报告器
func (ms *MonitoringSystem) RegisterReporter(reporter MetricReporter) {
    ms.reporters = append(ms.reporters, reporter)
}

// 添加告警规则
func (ms *MonitoringSystem) AddAlertRule(rule AlertRule) {
    ms.alertRules = append(ms.alertRules, rule)
}

// 启动监控
func (ms *MonitoringSystem) Start(ctx context.Context) error {
    // 启动指标收集
    go ms.startMetricCollection(ctx)

    // 启动告警检查
    go ms.startAlertChecking(ctx)

    // 启动指标报告
    go ms.startMetricReporting(ctx)

    return nil
}

// 指标收集循环
func (ms *MonitoringSystem) startMetricCollection(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ms.collectMetrics(ctx)
        }
    }
}

// 收集指标
func (ms *MonitoringSystem) collectMetrics(ctx context.Context) {
    var wg sync.WaitGroup

    for _, collector := range ms.collectors {
        wg.Add(1)
        go func(c MetricCollector) {
            defer wg.Done()

            data, err := c.Collect(ctx)
            if err != nil {
                log.Printf("Failed to collect metrics from %s: %v", c.GetName(), err)
                return
            }

            ms.updateMetrics(c.GetName(), data)
        }(collector)
    }

    wg.Wait()
}

// 更新指标
func (ms *MonitoringSystem) updateMetrics(collectorName string, data *MonitoringData) {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    key := fmt.Sprintf("%s:%s", collectorName, data.Topic)
    ms.metrics[key] = data

    // 更新Prometheus指标
    ms.updatePrometheusMetrics(data)
}

// 更新Prometheus指标
func (ms *MonitoringSystem) updatePrometheusMetrics(data *MonitoringData) {
    // 更新消息计数
    messagesConsumed.WithLabelValues(data.Topic, data.ConsumerGroup, "success").
        Add(float64(data.MessageCount))

    // 更新处理时间
    messageProcessingDuration.WithLabelValues(data.Topic, data.ConsumerGroup).
        Observe(data.ProcessingTime.Seconds())

    // 更新队列深度
    queueDepth.WithLabelValues(data.Topic).Set(float64(data.QueueDepth))

    // 更新连接状态
    connectionValue := 0.0
    if data.ConnectionStatus {
        connectionValue = 1.0
    }
    connectionStatus.WithLabelValues("broker", "consumer").Set(connectionValue)
}

// 告警检查循环
func (ms *MonitoringSystem) startAlertChecking(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ms.checkAlerts(ctx)
        }
    }
}

// 检查告警
func (ms *MonitoringSystem) checkAlerts(ctx context.Context) {
    ms.mutex.RLock()
    metrics := make(map[string]*MonitoringData)
    for k, v := range ms.metrics {
        metrics[k] = v
    }
    ms.mutex.RUnlock()

    for _, rule := range ms.alertRules {
        if ms.evaluateAlertRule(rule, metrics) {
            ms.triggerAlert(ctx, rule, metrics)
        }
    }
}

// 评估告警规则
func (ms *MonitoringSystem) evaluateAlertRule(rule AlertRule, metrics map[string]*MonitoringData) bool {
    switch rule.Condition {
    case "queue_depth_high":
        for _, data := range metrics {
            if float64(data.QueueDepth) > rule.Threshold {
                return true
            }
        }
    case "error_rate_high":
        for _, data := range metrics {
            if data.MessageCount > 0 {
                errorRate := float64(data.ErrorCount) / float64(data.MessageCount)
                if errorRate > rule.Threshold {
                    return true
                }
            }
        }
    case "processing_time_high":
        for _, data := range metrics {
            if data.ProcessingTime.Seconds() > rule.Threshold {
                return true
            }
        }
    case "connection_down":
        for _, data := range metrics {
            if !data.ConnectionStatus {
                return true
            }
        }
    }

    return false
}

// 触发告警
func (ms *MonitoringSystem) triggerAlert(ctx context.Context, rule AlertRule, metrics map[string]*MonitoringData) {
    // 检查告警冷却时间
    if time.Since(rule.LastTriggered) < rule.Duration {
        return
    }

    // 更新最后触发时间
    rule.LastTriggered = time.Now()

    // 执行告警动作
    for _, action := range rule.Actions {
        go ms.executeAlertAction(ctx, action, rule, metrics)
    }

    log.Printf("Alert triggered: %s", rule.Name)
}

// 执行告警动作
func (ms *MonitoringSystem) executeAlertAction(ctx context.Context, action AlertAction, rule AlertRule, metrics map[string]*MonitoringData) {
    switch action.Type {
    case "email":
        ms.sendEmailAlert(ctx, action.Config, rule, metrics)
    case "webhook":
        ms.sendWebhookAlert(ctx, action.Config, rule, metrics)
    case "slack":
        ms.sendSlackAlert(ctx, action.Config, rule, metrics)
    }
}

// 指标报告循环
func (ms *MonitoringSystem) startMetricReporting(ctx context.Context) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ms.reportMetrics(ctx)
        }
    }
}

// 报告指标
func (ms *MonitoringSystem) reportMetrics(ctx context.Context) {
    ms.mutex.RLock()
    metrics := make(map[string]*MonitoringData)
    for k, v := range ms.metrics {
        metrics[k] = v
    }
    ms.mutex.RUnlock()

    for _, reporter := range ms.reporters {
        for _, data := range metrics {
            go func(r MetricReporter, d *MonitoringData) {
                if err := r.Report(ctx, d); err != nil {
                    log.Printf("Failed to report metrics to %s: %v", r.GetName(), err)
                }
            }(reporter, data)
        }
    }
}

// RabbitMQ指标收集器
type RabbitMQCollector struct {
    client *rabbitmq.Client
    queues []string
}

func NewRabbitMQCollector(client *rabbitmq.Client, queues []string) *RabbitMQCollector {
    return &RabbitMQCollector{
        client: client,
        queues: queues,
    }
}

func (c *RabbitMQCollector) GetName() string {
    return "rabbitmq"
}

func (c *RabbitMQCollector) Collect(ctx context.Context) (*MonitoringData, error) {
    // 模拟收集RabbitMQ指标
    data := &MonitoringData{
        Timestamp:        time.Now(),
        Topic:            "orders",
        MessageCount:     1000,
        ProcessingTime:   50 * time.Millisecond,
        ErrorCount:       5,
        QueueDepth:       100,
        ConnectionStatus: true,
        CustomMetrics: map[string]interface{}{
            "memory_usage": "256MB",
            "disk_usage":   "1GB",
        },
    }

    return data, nil
}

// InfluxDB报告器
type InfluxDBReporter struct {
    client   *influxdb.Client
    database string
}

func NewInfluxDBReporter(client *influxdb.Client, database string) *InfluxDBReporter {
    return &InfluxDBReporter{
        client:   client,
        database: database,
    }
}

func (r *InfluxDBReporter) GetName() string {
    return "influxdb"
}

func (r *InfluxDBReporter) Report(ctx context.Context, data *MonitoringData) error {
    // 模拟写入InfluxDB
    log.Printf("Reporting metrics to InfluxDB: %s", data.Topic)
    return nil
}

// 辅助方法
func (ms *MonitoringSystem) sendEmailAlert(ctx context.Context, config map[string]interface{}, rule AlertRule, metrics map[string]*MonitoringData) {
    // 模拟发送邮件告警
    log.Printf("Sending email alert: %s", rule.Name)
}

func (ms *MonitoringSystem) sendWebhookAlert(ctx context.Context, config map[string]interface{}, rule AlertRule, metrics map[string]*MonitoringData) {
    // 模拟发送Webhook告警
    log.Printf("Sending webhook alert: %s", rule.Name)
}

func (ms *MonitoringSystem) sendSlackAlert(ctx context.Context, config map[string]interface{}, rule AlertRule, metrics map[string]*MonitoringData) {
    // 模拟发送Slack告警
    log.Printf("Sending slack alert: %s", rule.Name)
}

// 使用示例
func ExampleMonitoringSystem() {
    // 创建监控系统
    ms := NewMonitoringSystem()

    // 注册收集器
    rabbitmqClient := &rabbitmq.Client{} // 假设已初始化
    collector := NewRabbitMQCollector(rabbitmqClient, []string{"orders", "payments"})
    ms.RegisterCollector(collector)

    // 注册报告器
    influxClient := &influxdb.Client{} // 假设已初始化
    reporter := NewInfluxDBReporter(influxClient, "monitoring")
    ms.RegisterReporter(reporter)

    // 添加告警规则
    ms.AddAlertRule(AlertRule{
        Name:      "High Queue Depth",
        Condition: "queue_depth_high",
        Threshold: 1000,
        Duration:  5 * time.Minute,
        Severity:  "warning",
        Actions: []AlertAction{
            {
                Type: "email",
                Config: map[string]interface{}{
                    "to": "admin@example.com",
                },
            },
        },
    })

    // 启动监控
    ctx := context.Background()
    ms.Start(ctx)
}

/*
解析说明：
1. 多维度监控：消息计数、处理延迟、队列深度、连接状态等
2. 可扩展架构：支持多种收集器和报告器
3. 告警系统：支持多种告警条件和动作
4. Prometheus集成：标准的监控指标格式
5. 实时监控：定时收集和报告指标

扩展思考：
- 如何实现更复杂的告警规则？
- 如何优化大量指标的存储和查询？
- 如何实现监控数据的可视化？
- 如何处理监控系统本身的高可用？
*/
```

---

## 📚 章节总结

### 🎯 本章学习成果

通过本章的学习，你已经掌握了：

#### 📖 理论知识
- **消息队列核心概念**：生产者、消费者、代理、主题、队列等基础概念
- **消息传递模式**：点对点、发布订阅、请求响应等通信模式
- **可靠性保证机制**：消息确认、持久化、死信队列、重试机制
- **事件驱动架构**：事件溯源、CQRS、领域事件等高级架构模式

#### 🛠️ 实践技能
- **多种MQ集成**：RabbitMQ、Kafka、NSQ、Redis Stream的完整集成方案
- **消息可靠性**：生产者确认、消费者确认、事务消息等可靠性保证
- **性能优化**：连接池管理、批量处理、并发消费等性能优化技巧
- **错误处理**：分类错误处理、重试机制、死信队列等容错设计
- **监控告警**：指标收集、实时监控、告警规则等运维能力

#### 🏗️ 架构能力
- **事件驱动设计**：基于事件的微服务架构设计和实现
- **分布式事务**：基于消息的分布式事务解决方案
- **系统解耦**：通过消息队列实现系统间的松耦合
- **流量控制**：消息队列在流量削峰和负载均衡中的应用

### 🆚 消息队列技术选型总结

| 场景 | 推荐方案 | 理由 |
|------|----------|------|
| **企业级应用** | RabbitMQ | 功能丰富、可靠性高、生态成熟 |
| **大数据流处理** | Apache Kafka | 高吞吐量、水平扩展、流处理支持 |
| **Go微服务** | NSQ | 原生Go、简单易用、去中心化 |
| **轻量级队列** | Redis Stream | 部署简单、性能优秀、学习成本低 |
| **实时通信** | WebSocket + Redis | 低延迟、实时性好 |
| **批处理任务** | RabbitMQ + 延迟队列 | 支持延迟消息、任务调度 |

### 🎯 面试准备要点

#### 核心概念掌握
- 消息队列的优势和劣势，适用场景分析
- 不同消息队列的特点对比和选型依据
- 消息可靠性保证的完整方案
- 事件驱动架构的设计原则和实现方式

#### 实践经验展示
- 大型项目中的消息队列架构设计经验
- 消息丢失、重复、乱序等问题的解决实践
- 高并发场景下的性能优化案例
- 分布式事务的实现方案和踩坑经验

#### 问题解决能力
- 常见消息队列问题的排查思路
- 消息积压和性能瓶颈的诊断方法
- 系统故障时的应急处理能力
- 监控体系的建设和维护经验

### 🚀 下一步学习建议

#### 深入学习方向
1. **消息队列源码分析**
   - RabbitMQ的Erlang实现原理
   - Kafka的分布式日志设计
   - NSQ的Go语言实现细节
   - 消息路由和存储机制

2. **高级特性探索**
   - 消息队列集群部署和运维
   - 跨数据中心的消息复制
   - 消息队列的安全机制
   - 流处理和复杂事件处理

3. **企业级实践**
   - 大规模消息队列的容量规划
   - 消息队列的灾备和恢复
   - 多租户消息队列设计
   - 消息队列的成本优化

#### 实践项目建议
1. **个人项目**：构建一个完整的事件驱动电商系统
2. **开源贡献**：参与消息队列相关开源项目
3. **企业实践**：在生产环境中应用消息队列解决实际问题

### 💡 学习心得

消息队列作为现代分布式系统的重要组件，不仅仅是简单的消息传递工具，更是实现系统解耦、提高可扩展性的关键技术。通过本章的学习，我们不仅掌握了各种消息队列的使用方法，更重要的是培养了事件驱动的架构思维。

在实际应用中，要始终记住：
- **选型优于优化**：选择最适合业务场景的消息队列
- **可靠性优于性能**：在保证可靠性的前提下追求性能
- **监控优于调试**：建立完善的监控体系预防问题
- **简单优于复杂**：避免过度设计和不必要的复杂性

### 🔗 与其他章节的联系

本章内容与其他章节紧密相关：
- **Redis缓存章节**：消息队列可以触发缓存更新和失效
- **数据库章节**：消息队列实现数据库的异步同步和备份
- **微服务章节**：消息队列是微服务间通信的重要方式
- **监控章节**：消息队列需要完善的监控和告警机制

### 🎉 恭喜完成

恭喜你完成了消息队列集成与实践的学习！你现在已经具备了：

✅ **扎实的理论基础** - 深入理解消息队列原理和事件驱动架构
✅ **丰富的实践经验** - 掌握多种消息队列的集成和使用方法
✅ **优秀的架构能力** - 能够设计高可用、高性能的消息系统
✅ **完善的面试准备** - 具备回答各种消息队列相关问题的能力

继续保持学习的热情，在Go语言的道路上不断前进！下一章我们将学习微服务架构设计，进一步提升系统的可扩展性和可维护性。

---

*"消息队列是分布式系统的神经网络，事件驱动是现代架构的核心思想。掌握了消息队列，你就掌握了构建大规模分布式系统的关键技能！"* 🚀✨
