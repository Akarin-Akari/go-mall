# 架构篇第四章：分布式系统概念 🌐

> *"分布式系统是现代软件架构的基石，它让我们能够构建跨越地理边界、处理海量数据、服务亿万用户的系统。理解分布式系统的核心概念，就是掌握了构建大规模系统的钥匙！"* 🔑

## 📚 本章学习目标

通过本章学习，你将掌握：

- 🎯 **分布式系统基础**：理解分布式系统的定义、特征和挑战
- 📐 **CAP理论**：深入理解一致性、可用性、分区容错性的权衡
- 🔄 **一致性模型**：掌握强一致性、最终一致性等不同一致性级别
- 🗳️ **共识算法**：深入学习Raft、PBFT等共识算法的原理和实现
- 🔒 **分布式锁**：实现基于Redis、Zookeeper等的分布式锁
- 💳 **分布式事务**：掌握2PC、3PC、Saga等分布式事务模式
- 🕰️ **分布式时钟**：理解逻辑时钟、向量时钟等时序概念
- 🛠️ **Go语言实现**：使用Go实现分布式系统的核心组件
- 🏢 **企业级实践**：结合mall-go项目的分布式架构设计

---

## 🌟 分布式系统概述

### 什么是分布式系统？

分布式系统是由多个独立的计算机节点组成的系统，这些节点通过网络连接，协同工作来完成共同的任务，对用户来说就像一个统一的系统。

```go
// 分布式系统的核心概念
package distributed

import (
    "context"
    "time"
    "sync"
)

// 分布式系统节点
type Node struct {
    ID       string            `json:"id"`
    Address  string            `json:"address"`
    Role     NodeRole          `json:"role"`
    Status   NodeStatus        `json:"status"`
    Metadata map[string]string `json:"metadata"`
    
    // 节点通信
    Network  NetworkInterface  `json:"-"`
    
    // 状态管理
    State    NodeState         `json:"-"`
    
    // 时钟同步
    Clock    LogicalClock      `json:"-"`
    
    // 故障检测
    FailureDetector FailureDetector `json:"-"`
}

// 节点角色
type NodeRole string

const (
    RoleLeader    NodeRole = "leader"
    RoleFollower  NodeRole = "follower"
    RoleCandidate NodeRole = "candidate"
    RoleLearner   NodeRole = "learner"
)

// 节点状态
type NodeStatus string

const (
    StatusActive   NodeStatus = "active"
    StatusInactive NodeStatus = "inactive"
    StatusSuspect  NodeStatus = "suspect"
    StatusFailed   NodeStatus = "failed"
)

// 分布式系统特征
type DistributedSystemCharacteristics struct {
    // 1. 并发性
    Concurrency struct {
        Description string `json:"description"` // "多个节点同时执行"
        Challenges  []string `json:"challenges"` // ["竞态条件", "死锁", "活锁"]
        Solutions   []string `json:"solutions"`  // ["锁机制", "无锁算法", "事务"]
    } `json:"concurrency"`
    
    // 2. 缺乏全局时钟
    NoGlobalClock struct {
        Description string `json:"description"` // "节点间时钟不同步"
        Challenges  []string `json:"challenges"` // ["事件排序", "因果关系", "一致性"]
        Solutions   []string `json:"solutions"`  // ["逻辑时钟", "向量时钟", "NTP"]
    } `json:"no_global_clock"`
    
    // 3. 独立故障
    IndependentFailures struct {
        Description string `json:"description"` // "节点可能独立失败"
        Challenges  []string `json:"challenges"` // ["部分失败", "网络分区", "脑裂"]
        Solutions   []string `json:"solutions"`  // ["冗余", "故障检测", "恢复机制"]
    } `json:"independent_failures"`
    
    // 4. 网络不可靠
    UnreliableNetwork struct {
        Description string `json:"description"` // "网络可能丢包、延迟、重复"
        Challenges  []string `json:"challenges"` // ["消息丢失", "网络分区", "延迟"]
        Solutions   []string `json:"solutions"`  // ["重试机制", "幂等性", "超时处理"]
    } `json:"unreliable_network"`
}

// 分布式系统挑战
type DistributedSystemChallenges struct {
    // 异构性
    Heterogeneity struct {
        Networks    []string `json:"networks"`    // ["以太网", "WiFi", "移动网络"]
        Hardware    []string `json:"hardware"`    // ["x86", "ARM", "GPU"]
        OS          []string `json:"os"`          // ["Linux", "Windows", "macOS"]
        Languages   []string `json:"languages"`   // ["Go", "Java", "Python"]
        Middleware  []string `json:"middleware"`  // ["gRPC", "HTTP", "消息队列"]
    } `json:"heterogeneity"`
    
    // 开放性
    Openness struct {
        Extensibility   bool `json:"extensibility"`   // 可扩展性
        Interoperability bool `json:"interoperability"` // 互操作性
        Portability     bool `json:"portability"`     // 可移植性
        Standards       []string `json:"standards"`   // ["HTTP", "TCP/IP", "JSON"]
    } `json:"openness"`
    
    // 安全性
    Security struct {
        Confidentiality bool `json:"confidentiality"` // 机密性
        Integrity       bool `json:"integrity"`       // 完整性
        Availability    bool `json:"availability"`    // 可用性
        Authentication  bool `json:"authentication"`  // 认证
        Authorization   bool `json:"authorization"`   // 授权
    } `json:"security"`
    
    // 可扩展性
    Scalability struct {
        Size        string `json:"size"`        // "节点数量扩展"
        Geography   string `json:"geography"`   // "地理分布扩展"
        Management  string `json:"management"`  // "管理复杂度扩展"
    } `json:"scalability"`
    
    // 故障处理
    FailureHandling struct {
        Detection   string `json:"detection"`   // "故障检测"
        Masking     string `json:"masking"`     // "故障屏蔽"
        Tolerance   string `json:"tolerance"`   // "故障容忍"
        Recovery    string `json:"recovery"`    // "故障恢复"
    } `json:"failure_handling"`
    
    // 并发控制
    ConcurrencyControl struct {
        Coordination string `json:"coordination"` // "节点协调"
        Synchronization string `json:"synchronization"` // "同步机制"
        Consistency string `json:"consistency"` // "一致性保证"
    } `json:"concurrency_control"`
    
    // 透明性
    Transparency struct {
        Access      bool `json:"access"`      // 访问透明性
        Location    bool `json:"location"`    // 位置透明性
        Migration   bool `json:"migration"`   // 迁移透明性
        Replication bool `json:"replication"` // 复制透明性
        Concurrency bool `json:"concurrency"` // 并发透明性
        Failure     bool `json:"failure"`     // 故障透明性
    } `json:"transparency"`
}
```

### 分布式系统的目标

```go
// 分布式系统设计目标
type DistributedSystemGoals struct {
    // 1. 资源共享
    ResourceSharing struct {
        Hardware []string `json:"hardware"` // ["CPU", "内存", "存储", "网络"]
        Software []string `json:"software"` // ["数据库", "文件系统", "应用服务"]
        Data     []string `json:"data"`     // ["文件", "数据库", "缓存"]
    } `json:"resource_sharing"`
    
    // 2. 开放性
    Openness struct {
        Standards    []string `json:"standards"`    // ["HTTP", "TCP/IP", "REST"]
        Interfaces   []string `json:"interfaces"`   // ["API", "协议", "格式"]
        Portability  bool     `json:"portability"`  // 可移植性
        Extensibility bool    `json:"extensibility"` // 可扩展性
    } `json:"openness"`
    
    // 3. 并发性
    Concurrency struct {
        MultiUser    bool `json:"multi_user"`    // 多用户并发
        MultiProcess bool `json:"multi_process"` // 多进程并发
        MultiThread  bool `json:"multi_thread"`  // 多线程并发
    } `json:"concurrency"`
    
    // 4. 可扩展性
    Scalability struct {
        Horizontal bool `json:"horizontal"` // 水平扩展
        Vertical   bool `json:"vertical"`   // 垂直扩展
        Geographic bool `json:"geographic"` // 地理扩展
    } `json:"scalability"`
    
    // 5. 容错性
    FaultTolerance struct {
        Redundancy   bool `json:"redundancy"`   // 冗余
        Replication  bool `json:"replication"`  // 复制
        Recovery     bool `json:"recovery"`     // 恢复
        Graceful     bool `json:"graceful"`     // 优雅降级
    } `json:"fault_tolerance"`
    
    // 6. 透明性
    Transparency struct {
        Access      bool `json:"access"`      // 访问透明性
        Location    bool `json:"location"`    // 位置透明性
        Concurrency bool `json:"concurrency"` // 并发透明性
        Replication bool `json:"replication"` // 复制透明性
        Failure     bool `json:"failure"`     // 故障透明性
        Migration   bool `json:"migration"`   // 迁移透明性
        Performance bool `json:"performance"` // 性能透明性
        Scaling     bool `json:"scaling"`     // 扩展透明性
    } `json:"transparency"`
}
```

---

## 📐 CAP理论

### CAP理论基础

CAP理论是分布式系统设计的基础理论，由Eric Brewer在2000年提出。它指出在分布式系统中，一致性(Consistency)、可用性(Availability)、分区容错性(Partition tolerance)三者不能同时满足，最多只能同时满足其中两个。

```go
// CAP理论实现
package cap

import (
    "context"
    "errors"
    "sync"
    "time"
)

// CAP特性定义
type CAPProperty string

const (
    Consistency         CAPProperty = "consistency"
    Availability        CAPProperty = "availability"
    PartitionTolerance  CAPProperty = "partition_tolerance"
)

// CAP系统类型
type CAPSystemType string

const (
    CP_System  CAPSystemType = "cp"  // 一致性 + 分区容错性
    AP_System  CAPSystemType = "ap"  // 可用性 + 分区容错性
    CA_System  CAPSystemType = "ca"  // 一致性 + 可用性（理论上，实际不存在）
)

// CAP系统配置
type CAPSystemConfig struct {
    Type                CAPSystemType `json:"type"`
    ConsistencyLevel    ConsistencyLevel `json:"consistency_level"`
    AvailabilityTarget  float64 `json:"availability_target"` // 99.9%
    PartitionStrategy   PartitionStrategy `json:"partition_strategy"`
    
    // 权衡策略
    TradeoffStrategy    TradeoffStrategy `json:"tradeoff_strategy"`
}

// 一致性级别
type ConsistencyLevel string

const (
    StrongConsistency     ConsistencyLevel = "strong"
    EventualConsistency   ConsistencyLevel = "eventual"
    WeakConsistency       ConsistencyLevel = "weak"
    CausalConsistency     ConsistencyLevel = "causal"
    MonotonicConsistency  ConsistencyLevel = "monotonic"
)

// 分区策略
type PartitionStrategy string

const (
    PartitionIgnore       PartitionStrategy = "ignore"
    PartitionDetectAndWait PartitionStrategy = "detect_and_wait"
    PartitionDetectAndContinue PartitionStrategy = "detect_and_continue"
)

// CAP权衡策略
type TradeoffStrategy struct {
    // 网络正常时的策略
    NormalOperation struct {
        PrioritizeConsistency bool `json:"prioritize_consistency"`
        PrioritizeAvailability bool `json:"prioritize_availability"`
        PrioritizePerformance bool `json:"prioritize_performance"`
    } `json:"normal_operation"`
    
    // 网络分区时的策略
    PartitionOperation struct {
        Strategy PartitionHandlingStrategy `json:"strategy"`
        Timeout  time.Duration `json:"timeout"`
        Retry    RetryConfig `json:"retry"`
    } `json:"partition_operation"`
}

// 分区处理策略
type PartitionHandlingStrategy string

const (
    FailFast        PartitionHandlingStrategy = "fail_fast"
    WaitForHealing  PartitionHandlingStrategy = "wait_for_healing"
    ContinueServing PartitionHandlingStrategy = "continue_serving"
    DegradedService PartitionHandlingStrategy = "degraded_service"
)

// CAP系统实现
type CAPSystem struct {
    config      CAPSystemConfig
    nodes       map[string]*Node
    partitions  map[string]bool // 记录分区状态
    mutex       sync.RWMutex

    // 一致性管理
    consistency ConsistencyManager

    // 可用性管理
    availability AvailabilityManager

    // 分区检测
    partitionDetector PartitionDetector
}

// 一致性管理器
type ConsistencyManager interface {
    // 读操作
    Read(ctx context.Context, key string) (interface{}, error)

    // 写操作
    Write(ctx context.Context, key string, value interface{}) error

    // 检查一致性
    CheckConsistency(ctx context.Context) (bool, error)

    // 修复不一致
    RepairInconsistency(ctx context.Context) error
}

// 强一致性实现
type StrongConsistencyManager struct {
    quorum      int
    nodes       []*Node
    coordinator *Node
}

func (s *StrongConsistencyManager) Write(ctx context.Context, key string, value interface{}) error {
    // 1. 选择协调者
    if s.coordinator == nil {
        return errors.New("no coordinator available")
    }

    // 2. 两阶段提交
    // Phase 1: Prepare
    prepareCount := 0
    for _, node := range s.nodes {
        if err := node.Prepare(ctx, key, value); err == nil {
            prepareCount++
        }
    }

    // 检查是否达到法定人数
    if prepareCount < s.quorum {
        // 回滚
        for _, node := range s.nodes {
            node.Abort(ctx, key)
        }
        return errors.New("insufficient nodes for strong consistency")
    }

    // Phase 2: Commit
    for _, node := range s.nodes {
        node.Commit(ctx, key, value)
    }

    return nil
}

func (s *StrongConsistencyManager) Read(ctx context.Context, key string) (interface{}, error) {
    // 从法定人数的节点读取
    values := make(map[interface{}]int)
    readCount := 0

    for _, node := range s.nodes {
        if value, err := node.Read(ctx, key); err == nil {
            values[value]++
            readCount++
            if readCount >= s.quorum {
                break
            }
        }
    }

    if readCount < s.quorum {
        return nil, errors.New("insufficient nodes for consistent read")
    }

    // 返回多数派的值
    var majorityValue interface{}
    maxCount := 0
    for value, count := range values {
        if count > maxCount {
            maxCount = count
            majorityValue = value
        }
    }

    return majorityValue, nil
}

// 最终一致性实现
type EventualConsistencyManager struct {
    nodes           []*Node
    replicationFactor int
    conflictResolver ConflictResolver
}

func (e *EventualConsistencyManager) Write(ctx context.Context, key string, value interface{}) error {
    // 异步复制到多个节点
    successCount := 0

    for i, node := range e.nodes {
        if i < e.replicationFactor {
            go func(n *Node) {
                n.AsyncWrite(ctx, key, value)
            }(node)
            successCount++
        }
    }

    // 只要有一个节点写入成功就返回
    if successCount > 0 {
        return nil
    }

    return errors.New("no nodes available for write")
}

func (e *EventualConsistencyManager) Read(ctx context.Context, key string) (interface{}, error) {
    // 从任意可用节点读取
    for _, node := range e.nodes {
        if value, err := node.Read(ctx, key); err == nil {
            return value, nil
        }
    }

    return nil, errors.New("no nodes available for read")
}

// 可用性管理器
type AvailabilityManager interface {
    // 检查系统可用性
    CheckAvailability(ctx context.Context) (float64, error)

    // 处理节点故障
    HandleNodeFailure(nodeID string) error

    // 节点恢复
    HandleNodeRecovery(nodeID string) error

    // 负载均衡
    LoadBalance(request interface{}) (*Node, error)
}

// 高可用性实现
type HighAvailabilityManager struct {
    nodes           map[string]*Node
    healthChecker   HealthChecker
    loadBalancer    LoadBalancer
    failoverManager FailoverManager

    // 可用性目标
    availabilityTarget float64

    // 统计信息
    stats AvailabilityStats
}

type AvailabilityStats struct {
    TotalRequests   int64     `json:"total_requests"`
    SuccessRequests int64     `json:"success_requests"`
    FailedRequests  int64     `json:"failed_requests"`
    Uptime          time.Duration `json:"uptime"`
    Downtime        time.Duration `json:"downtime"`
    LastFailure     time.Time `json:"last_failure"`
}

func (h *HighAvailabilityManager) CheckAvailability(ctx context.Context) (float64, error) {
    if h.stats.TotalRequests == 0 {
        return 1.0, nil
    }

    availability := float64(h.stats.SuccessRequests) / float64(h.stats.TotalRequests)
    return availability, nil
}

func (h *HighAvailabilityManager) HandleNodeFailure(nodeID string) error {
    h.mutex.Lock()
    defer h.mutex.Unlock()

    // 1. 标记节点为失败状态
    if node, exists := h.nodes[nodeID]; exists {
        node.Status = StatusFailed
    }

    // 2. 触发故障转移
    return h.failoverManager.Failover(nodeID)
}

// 分区检测器
type PartitionDetector interface {
    // 检测网络分区
    DetectPartition(ctx context.Context) ([]Partition, error)

    // 监控分区状态
    MonitorPartitions(ctx context.Context) <-chan PartitionEvent

    // 分区恢复检测
    DetectPartitionHealing(ctx context.Context) error
}

type Partition struct {
    ID       string    `json:"id"`
    Nodes    []string  `json:"nodes"`
    StartTime time.Time `json:"start_time"`
    EndTime   *time.Time `json:"end_time,omitempty"`
    Status   PartitionStatus `json:"status"`
}

type PartitionStatus string

const (
    PartitionActive  PartitionStatus = "active"
    PartitionHealed  PartitionStatus = "healed"
)

type PartitionEvent struct {
    Type      PartitionEventType `json:"type"`
    Partition Partition          `json:"partition"`
    Timestamp time.Time          `json:"timestamp"`
}

type PartitionEventType string

const (
    PartitionDetected PartitionEventType = "detected"
    PartitionHealed   PartitionEventType = "healed"
)
```

### CAP理论的实际应用

```go
// 不同CAP组合的实际系统示例
type CAPExamples struct {
    // CP系统：一致性 + 分区容错性
    CPSystems []CPSystemExample `json:"cp_systems"`

    // AP系统：可用性 + 分区容错性
    APSystems []APSystemExample `json:"ap_systems"`

    // CA系统：一致性 + 可用性（理论上）
    CASystems []CASystemExample `json:"ca_systems"`
}

type CPSystemExample struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Examples    []string `json:"examples"`
    Scenarios   []string `json:"scenarios"`
}

type APSystemExample struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Examples    []string `json:"examples"`
    Scenarios   []string `json:"scenarios"`
}

type CASystemExample struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Examples    []string `json:"examples"`
    Limitations []string `json:"limitations"`
}

func GetCAPExamples() CAPExamples {
    return CAPExamples{
        CPSystems: []CPSystemExample{
            {
                Name:        "分布式数据库",
                Description: "优先保证数据一致性，在网络分区时可能不可用",
                Examples:    []string{"MongoDB", "HBase", "Redis Cluster"},
                Scenarios:   []string{"金融交易", "库存管理", "用户账户"},
            },
            {
                Name:        "分布式锁服务",
                Description: "保证锁的一致性，分区时停止服务",
                Examples:    []string{"Zookeeper", "etcd", "Consul"},
                Scenarios:   []string{"配置管理", "服务发现", "分布式协调"},
            },
        },
        APSystems: []APSystemExample{
            {
                Name:        "内容分发网络",
                Description: "优先保证服务可用性，允许数据暂时不一致",
                Examples:    []string{"DNS", "CDN", "Cassandra"},
                Scenarios:   []string{"静态内容分发", "用户会话", "日志收集"},
            },
            {
                Name:        "社交网络",
                Description: "保证用户能够访问服务，允许数据最终一致",
                Examples:    []string{"Facebook", "Twitter", "DynamoDB"},
                Scenarios:   []string{"用户动态", "消息推送", "内容推荐"},
            },
        },
        CASystems: []CASystemExample{
            {
                Name:        "单机数据库",
                Description: "在没有网络分区的环境中保证一致性和可用性",
                Examples:    []string{"MySQL", "PostgreSQL", "SQLite"},
                Limitations: []string{"无法处理网络分区", "单点故障", "扩展性限制"},
            },
        },
    }
}
```

---

## 🗳️ 共识算法

### Raft共识算法

Raft是一种易于理解的共识算法，它将共识问题分解为领导者选举、日志复制和安全性三个子问题。

```go
// Raft算法实现
package raft

import (
    "context"
    "math/rand"
    "sync"
    "time"
)

// Raft节点状态
type RaftState string

const (
    Follower  RaftState = "follower"
    Candidate RaftState = "candidate"
    Leader    RaftState = "leader"
)

// Raft节点
type RaftNode struct {
    // 基本信息
    ID      string    `json:"id"`
    State   RaftState `json:"state"`

    // 持久化状态
    CurrentTerm int        `json:"current_term"`
    VotedFor    *string    `json:"voted_for"`
    Log         []LogEntry `json:"log"`

    // 易失状态
    CommitIndex int `json:"commit_index"`
    LastApplied int `json:"last_applied"`

    // 领导者状态
    NextIndex  map[string]int `json:"next_index,omitempty"`
    MatchIndex map[string]int `json:"match_index,omitempty"`

    // 网络和定时器
    peers       map[string]*RaftPeer
    electionTimer *time.Timer
    heartbeatTimer *time.Timer

    // 同步控制
    mutex sync.RWMutex

    // 状态机
    stateMachine StateMachine

    // 配置
    config RaftConfig
}

// 日志条目
type LogEntry struct {
    Index   int         `json:"index"`
    Term    int         `json:"term"`
    Command interface{} `json:"command"`
    Type    LogType     `json:"type"`
}

type LogType string

const (
    LogCommand      LogType = "command"
    LogConfiguration LogType = "configuration"
    LogNoOp         LogType = "noop"
)

// Raft配置
type RaftConfig struct {
    ElectionTimeoutMin time.Duration `json:"election_timeout_min"`
    ElectionTimeoutMax time.Duration `json:"election_timeout_max"`
    HeartbeatInterval  time.Duration `json:"heartbeat_interval"`
    MaxLogEntries      int           `json:"max_log_entries"`
    SnapshotThreshold  int           `json:"snapshot_threshold"`
}

// 投票请求
type VoteRequest struct {
    Term         int    `json:"term"`
    CandidateID  string `json:"candidate_id"`
    LastLogIndex int    `json:"last_log_index"`
    LastLogTerm  int    `json:"last_log_term"`
}

// 投票响应
type VoteResponse struct {
    Term        int  `json:"term"`
    VoteGranted bool `json:"vote_granted"`
}

// 追加条目请求
type AppendEntriesRequest struct {
    Term         int        `json:"term"`
    LeaderID     string     `json:"leader_id"`
    PrevLogIndex int        `json:"prev_log_index"`
    PrevLogTerm  int        `json:"prev_log_term"`
    Entries      []LogEntry `json:"entries"`
    LeaderCommit int        `json:"leader_commit"`
}

// 追加条目响应
type AppendEntriesResponse struct {
    Term    int  `json:"term"`
    Success bool `json:"success"`
}

// 领导者选举
func (r *RaftNode) StartElection() {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    // 1. 增加当前任期
    r.CurrentTerm++

    // 2. 转换为候选者状态
    r.State = Candidate

    // 3. 为自己投票
    r.VotedFor = &r.ID

    // 4. 重置选举定时器
    r.resetElectionTimer()

    // 5. 向所有其他节点发送投票请求
    voteCount := 1 // 自己的票
    totalNodes := len(r.peers) + 1

    for peerID, peer := range r.peers {
        go func(id string, p *RaftPeer) {
            request := VoteRequest{
                Term:         r.CurrentTerm,
                CandidateID:  r.ID,
                LastLogIndex: len(r.Log) - 1,
                LastLogTerm:  r.getLastLogTerm(),
            }

            response, err := p.RequestVote(context.Background(), request)
            if err != nil {
                return
            }

            r.mutex.Lock()
            defer r.mutex.Unlock()

            // 检查任期
            if response.Term > r.CurrentTerm {
                r.CurrentTerm = response.Term
                r.State = Follower
                r.VotedFor = nil
                return
            }

            // 统计选票
            if response.VoteGranted && r.State == Candidate {
                voteCount++
                if voteCount > totalNodes/2 {
                    r.becomeLeader()
                }
            }
        }(peerID, peer)
    }
}

// 成为领导者
func (r *RaftNode) becomeLeader() {
    r.State = Leader

    // 初始化领导者状态
    r.NextIndex = make(map[string]int)
    r.MatchIndex = make(map[string]int)

    for peerID := range r.peers {
        r.NextIndex[peerID] = len(r.Log)
        r.MatchIndex[peerID] = 0
    }

    // 发送心跳
    r.sendHeartbeats()

    // 启动心跳定时器
    r.startHeartbeatTimer()
}

// 发送心跳
func (r *RaftNode) sendHeartbeats() {
    for peerID, peer := range r.peers {
        go func(id string, p *RaftPeer) {
            r.sendAppendEntries(id, p)
        }(peerID, peer)
    }
}

// 发送追加条目
func (r *RaftNode) sendAppendEntries(peerID string, peer *RaftPeer) {
    r.mutex.RLock()

    if r.State != Leader {
        r.mutex.RUnlock()
        return
    }

    nextIndex := r.NextIndex[peerID]
    prevLogIndex := nextIndex - 1
    prevLogTerm := 0

    if prevLogIndex >= 0 && prevLogIndex < len(r.Log) {
        prevLogTerm = r.Log[prevLogIndex].Term
    }

    entries := []LogEntry{}
    if nextIndex < len(r.Log) {
        entries = r.Log[nextIndex:]
    }

    request := AppendEntriesRequest{
        Term:         r.CurrentTerm,
        LeaderID:     r.ID,
        PrevLogIndex: prevLogIndex,
        PrevLogTerm:  prevLogTerm,
        Entries:      entries,
        LeaderCommit: r.CommitIndex,
    }

    r.mutex.RUnlock()

    response, err := peer.AppendEntries(context.Background(), request)
    if err != nil {
        return
    }

    r.mutex.Lock()
    defer r.mutex.Unlock()

    // 检查任期
    if response.Term > r.CurrentTerm {
        r.CurrentTerm = response.Term
        r.State = Follower
        r.VotedFor = nil
        return
    }

    if response.Success {
        // 更新匹配索引
        r.MatchIndex[peerID] = prevLogIndex + len(entries)
        r.NextIndex[peerID] = r.MatchIndex[peerID] + 1

        // 更新提交索引
        r.updateCommitIndex()
    } else {
        // 减少下一个索引并重试
        if r.NextIndex[peerID] > 0 {
            r.NextIndex[peerID]--
        }
    }
}

// 处理投票请求
func (r *RaftNode) HandleVoteRequest(request VoteRequest) VoteResponse {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    response := VoteResponse{
        Term:        r.CurrentTerm,
        VoteGranted: false,
    }

    // 检查任期
    if request.Term > r.CurrentTerm {
        r.CurrentTerm = request.Term
        r.State = Follower
        r.VotedFor = nil
    }

    if request.Term < r.CurrentTerm {
        return response
    }

    // 检查是否已经投票
    if r.VotedFor != nil && *r.VotedFor != request.CandidateID {
        return response
    }

    // 检查日志是否至少和自己一样新
    lastLogIndex := len(r.Log) - 1
    lastLogTerm := r.getLastLogTerm()

    logUpToDate := request.LastLogTerm > lastLogTerm ||
        (request.LastLogTerm == lastLogTerm && request.LastLogIndex >= lastLogIndex)

    if logUpToDate {
        r.VotedFor = &request.CandidateID
        response.VoteGranted = true
        r.resetElectionTimer()
    }

    return response
}

// 处理追加条目请求
func (r *RaftNode) HandleAppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    response := AppendEntriesResponse{
        Term:    r.CurrentTerm,
        Success: false,
    }

    // 检查任期
    if request.Term > r.CurrentTerm {
        r.CurrentTerm = request.Term
        r.State = Follower
        r.VotedFor = nil
    }

    if request.Term < r.CurrentTerm {
        return response
    }

    // 重置选举定时器
    r.resetElectionTimer()

    // 检查日志一致性
    if request.PrevLogIndex >= 0 {
        if request.PrevLogIndex >= len(r.Log) ||
           r.Log[request.PrevLogIndex].Term != request.PrevLogTerm {
            return response
        }
    }

    // 追加新条目
    if len(request.Entries) > 0 {
        // 删除冲突的条目
        r.Log = r.Log[:request.PrevLogIndex+1]

        // 追加新条目
        r.Log = append(r.Log, request.Entries...)
    }

    // 更新提交索引
    if request.LeaderCommit > r.CommitIndex {
        r.CommitIndex = min(request.LeaderCommit, len(r.Log)-1)
        r.applyLogEntries()
    }

    response.Success = true
    return response
}
```

---

## 🔒 分布式锁

### 分布式锁基础

分布式锁是分布式系统中用于协调多个节点对共享资源访问的机制，确保在任意时刻只有一个节点能够访问特定资源。

```go
// 分布式锁接口
package distributedlock

import (
    "context"
    "time"
)

// 分布式锁接口
type DistributedLock interface {
    // 获取锁
    Lock(ctx context.Context) error

    // 尝试获取锁
    TryLock(ctx context.Context) (bool, error)

    // 带超时的获取锁
    LockWithTimeout(ctx context.Context, timeout time.Duration) error

    // 释放锁
    Unlock(ctx context.Context) error

    // 续期锁
    Renew(ctx context.Context, duration time.Duration) error

    // 检查锁状态
    IsLocked(ctx context.Context) (bool, error)

    // 获取锁信息
    GetLockInfo(ctx context.Context) (*LockInfo, error)
}

// 锁信息
type LockInfo struct {
    Key        string        `json:"key"`
    Owner      string        `json:"owner"`
    AcquiredAt time.Time     `json:"acquired_at"`
    ExpiresAt  time.Time     `json:"expires_at"`
    TTL        time.Duration `json:"ttl"`
    Metadata   map[string]interface{} `json:"metadata"`
}

// 锁配置
type LockConfig struct {
    Key           string        `json:"key"`
    Owner         string        `json:"owner"`
    TTL           time.Duration `json:"ttl"`
    RetryInterval time.Duration `json:"retry_interval"`
    MaxRetries    int           `json:"max_retries"`
    AutoRenew     bool          `json:"auto_renew"`
    RenewInterval time.Duration `json:"renew_interval"`
}

// Redis分布式锁实现
type RedisDistributedLock struct {
    client RedisClient
    config LockConfig

    // 锁状态
    locked    bool
    lockValue string

    // 自动续期
    renewCancel context.CancelFunc
}

func NewRedisDistributedLock(client RedisClient, config LockConfig) *RedisDistributedLock {
    return &RedisDistributedLock{
        client: client,
        config: config,
        lockValue: generateLockValue(),
    }
}

// Lua脚本：原子性获取锁
const acquireLockScript = `
    if redis.call("GET", KEYS[1]) == false then
        redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
        return 1
    else
        return 0
    end
`

// Lua脚本：原子性释放锁
const releaseLockScript = `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("DEL", KEYS[1])
    else
        return 0
    end
`

// Lua脚本：原子性续期锁
const renewLockScript = `
    if redis.call("GET", KEYS[1]) == ARGV[1] then
        return redis.call("PEXPIRE", KEYS[1], ARGV[2])
    else
        return 0
    end
`

func (r *RedisDistributedLock) Lock(ctx context.Context) error {
    for {
        acquired, err := r.TryLock(ctx)
        if err != nil {
            return err
        }

        if acquired {
            return nil
        }

        // 等待重试
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(r.config.RetryInterval):
            continue
        }
    }
}

func (r *RedisDistributedLock) TryLock(ctx context.Context) (bool, error) {
    result, err := r.client.Eval(ctx, acquireLockScript,
        []string{r.config.Key},
        r.lockValue,
        int64(r.config.TTL/time.Millisecond))

    if err != nil {
        return false, err
    }

    acquired := result.(int64) == 1
    if acquired {
        r.locked = true

        // 启动自动续期
        if r.config.AutoRenew {
            r.startAutoRenew(ctx)
        }
    }

    return acquired, nil
}

func (r *RedisDistributedLock) Unlock(ctx context.Context) error {
    if !r.locked {
        return nil
    }

    // 停止自动续期
    if r.renewCancel != nil {
        r.renewCancel()
    }

    result, err := r.client.Eval(ctx, releaseLockScript,
        []string{r.config.Key},
        r.lockValue)

    if err != nil {
        return err
    }

    released := result.(int64) == 1
    if released {
        r.locked = false
    }

    return nil
}

func (r *RedisDistributedLock) Renew(ctx context.Context, duration time.Duration) error {
    result, err := r.client.Eval(ctx, renewLockScript,
        []string{r.config.Key},
        r.lockValue,
        int64(duration/time.Millisecond))

    if err != nil {
        return err
    }

    renewed := result.(int64) == 1
    if !renewed {
        r.locked = false
        return errors.New("lock not owned by this instance")
    }

    return nil
}

func (r *RedisDistributedLock) startAutoRenew(ctx context.Context) {
    renewCtx, cancel := context.WithCancel(ctx)
    r.renewCancel = cancel

    go func() {
        ticker := time.NewTicker(r.config.RenewInterval)
        defer ticker.Stop()

        for {
            select {
            case <-renewCtx.Done():
                return
            case <-ticker.C:
                if err := r.Renew(renewCtx, r.config.TTL); err != nil {
                    // 续期失败，停止自动续期
                    return
                }
            }
        }
    }()
}

// Zookeeper分布式锁实现
type ZookeeperDistributedLock struct {
    client ZookeeperClient
    config LockConfig

    // 锁路径
    lockPath     string
    sequencePath string

    // 监听器
    watcher Watcher
}

func (z *ZookeeperDistributedLock) Lock(ctx context.Context) error {
    // 1. 创建临时顺序节点
    path, err := z.client.CreateSequential(z.config.Key+"/lock-",
        []byte(z.config.Owner), true)
    if err != nil {
        return err
    }
    z.sequencePath = path

    for {
        // 2. 获取所有子节点
        children, err := z.client.GetChildren(z.config.Key)
        if err != nil {
            return err
        }

        // 3. 排序并检查是否是最小节点
        sort.Strings(children)

        mySequence := filepath.Base(z.sequencePath)
        if children[0] == mySequence {
            // 获得锁
            return nil
        }

        // 4. 监听前一个节点
        myIndex := -1
        for i, child := range children {
            if child == mySequence {
                myIndex = i
                break
            }
        }

        if myIndex > 0 {
            prevPath := z.config.Key + "/" + children[myIndex-1]

            // 监听前一个节点的删除事件
            exists, _, eventCh, err := z.client.ExistsW(prevPath)
            if err != nil {
                return err
            }

            if !exists {
                // 前一个节点已经不存在，重新检查
                continue
            }

            // 等待前一个节点删除
            select {
            case <-ctx.Done():
                z.Unlock(ctx)
                return ctx.Err()
            case event := <-eventCh:
                if event.Type == EventNodeDeleted {
                    continue
                }
            }
        }
    }
}

func (z *ZookeeperDistributedLock) Unlock(ctx context.Context) error {
    if z.sequencePath != "" {
        return z.client.Delete(z.sequencePath, -1)
    }
    return nil
}
```

---

## 💳 分布式事务

### 两阶段提交(2PC)

```go
// 两阶段提交协议实现
package twophasecommit

import (
    "context"
    "errors"
    "sync"
    "time"
)

// 事务协调者
type TransactionCoordinator struct {
    transactionID string
    participants  []Participant
    state         TransactionState
    timeout       time.Duration

    // 投票结果
    votes map[string]VoteResult
    mutex sync.RWMutex
}

// 事务状态
type TransactionState string

const (
    TxInit      TransactionState = "init"
    TxPreparing TransactionState = "preparing"
    TxCommitting TransactionState = "committing"
    TxAborting  TransactionState = "aborting"
    TxCommitted TransactionState = "committed"
    TxAborted   TransactionState = "aborted"
)

// 投票结果
type VoteResult string

const (
    VoteYes     VoteResult = "yes"
    VoteNo      VoteResult = "no"
    VoteTimeout VoteResult = "timeout"
)

// 参与者接口
type Participant interface {
    // 准备阶段
    Prepare(ctx context.Context, txID string, data interface{}) (VoteResult, error)

    // 提交阶段
    Commit(ctx context.Context, txID string) error

    // 回滚阶段
    Abort(ctx context.Context, txID string) error

    // 获取参与者ID
    GetID() string
}

// 两阶段提交执行
func (tc *TransactionCoordinator) Execute(ctx context.Context, data interface{}) error {
    tc.state = TxPreparing
    tc.votes = make(map[string]VoteResult)

    // Phase 1: Prepare
    if !tc.preparePhase(ctx, data) {
        return tc.abortTransaction(ctx)
    }

    // Phase 2: Commit
    return tc.commitPhase(ctx)
}

// 准备阶段
func (tc *TransactionCoordinator) preparePhase(ctx context.Context, data interface{}) bool {
    var wg sync.WaitGroup
    voteChan := make(chan struct {
        participantID string
        vote         VoteResult
        err          error
    }, len(tc.participants))

    // 并发发送准备请求
    for _, participant := range tc.participants {
        wg.Add(1)
        go func(p Participant) {
            defer wg.Done()

            prepareCtx, cancel := context.WithTimeout(ctx, tc.timeout)
            defer cancel()

            vote, err := p.Prepare(prepareCtx, tc.transactionID, data)

            voteChan <- struct {
                participantID string
                vote         VoteResult
                err          error
            }{
                participantID: p.GetID(),
                vote:         vote,
                err:          err,
            }
        }(participant)
    }

    // 等待所有投票
    go func() {
        wg.Wait()
        close(voteChan)
    }()

    // 收集投票结果
    for result := range voteChan {
        tc.mutex.Lock()
        if result.err != nil {
            tc.votes[result.participantID] = VoteTimeout
        } else {
            tc.votes[result.participantID] = result.vote
        }
        tc.mutex.Unlock()
    }

    // 检查是否所有参与者都投了赞成票
    tc.mutex.RLock()
    defer tc.mutex.RUnlock()

    for _, vote := range tc.votes {
        if vote != VoteYes {
            return false
        }
    }

    return true
}

// 提交阶段
func (tc *TransactionCoordinator) commitPhase(ctx context.Context) error {
    tc.state = TxCommitting

    var wg sync.WaitGroup
    errorChan := make(chan error, len(tc.participants))

    // 并发发送提交请求
    for _, participant := range tc.participants {
        wg.Add(1)
        go func(p Participant) {
            defer wg.Done()

            commitCtx, cancel := context.WithTimeout(ctx, tc.timeout)
            defer cancel()

            if err := p.Commit(commitCtx, tc.transactionID); err != nil {
                errorChan <- err
            }
        }(participant)
    }

    wg.Wait()
    close(errorChan)

    // 检查是否有错误
    var errors []error
    for err := range errorChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        tc.state = TxAborted
        return fmt.Errorf("commit failed: %v", errors)
    }

    tc.state = TxCommitted
    return nil
}

// 回滚事务
func (tc *TransactionCoordinator) abortTransaction(ctx context.Context) error {
    tc.state = TxAborting

    var wg sync.WaitGroup

    // 并发发送回滚请求
    for _, participant := range tc.participants {
        wg.Add(1)
        go func(p Participant) {
            defer wg.Done()

            abortCtx, cancel := context.WithTimeout(ctx, tc.timeout)
            defer cancel()

            p.Abort(abortCtx, tc.transactionID)
        }(participant)
    }

    wg.Wait()
    tc.state = TxAborted

    return errors.New("transaction aborted")
}
```

### Saga模式

Saga是一种长事务处理模式，将长事务分解为一系列短事务，每个短事务都有对应的补偿操作。

```go
// Saga模式实现
package saga

import (
    "context"
    "fmt"
)

// Saga事务
type Saga struct {
    ID          string      `json:"id"`
    Steps       []SagaStep  `json:"steps"`
    CurrentStep int         `json:"current_step"`
    State       SagaState   `json:"state"`
    Context     SagaContext `json:"context"`
}

// Saga状态
type SagaState string

const (
    SagaRunning    SagaState = "running"
    SagaCompleted  SagaState = "completed"
    SagaFailed     SagaState = "failed"
    SagaCompensating SagaState = "compensating"
    SagaCompensated  SagaState = "compensated"
)

// Saga步骤
type SagaStep struct {
    Name        string                 `json:"name"`
    Action      SagaAction            `json:"-"`
    Compensation SagaCompensation     `json:"-"`
    Timeout     time.Duration         `json:"timeout"`
    Retries     int                   `json:"retries"`
    Data        map[string]interface{} `json:"data"`
}

// Saga动作接口
type SagaAction interface {
    Execute(ctx context.Context, sagaCtx SagaContext) error
}

// Saga补偿接口
type SagaCompensation interface {
    Compensate(ctx context.Context, sagaCtx SagaContext) error
}

// Saga上下文
type SagaContext struct {
    Data     map[string]interface{} `json:"data"`
    Results  map[string]interface{} `json:"results"`
    Metadata map[string]interface{} `json:"metadata"`
}

// Saga执行器
type SagaExecutor struct {
    saga    *Saga
    logger  Logger
    metrics MetricsCollector
}

// 执行Saga
func (se *SagaExecutor) Execute(ctx context.Context) error {
    se.saga.State = SagaRunning

    // 顺序执行每个步骤
    for i, step := range se.saga.Steps {
        se.saga.CurrentStep = i

        if err := se.executeStep(ctx, step); err != nil {
            se.saga.State = SagaFailed

            // 执行补偿操作
            return se.compensate(ctx)
        }
    }

    se.saga.State = SagaCompleted
    return nil
}

// 执行单个步骤
func (se *SagaExecutor) executeStep(ctx context.Context, step SagaStep) error {
    stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
    defer cancel()

    var err error
    for attempt := 0; attempt <= step.Retries; attempt++ {
        err = step.Action.Execute(stepCtx, se.saga.Context)
        if err == nil {
            se.logger.Info("Saga step completed", "step", step.Name, "saga", se.saga.ID)
            return nil
        }

        se.logger.Warn("Saga step failed", "step", step.Name, "attempt", attempt, "error", err)

        if attempt < step.Retries {
            time.Sleep(time.Second * time.Duration(attempt+1))
        }
    }

    return fmt.Errorf("saga step %s failed after %d attempts: %v", step.Name, step.Retries+1, err)
}

// 执行补偿
func (se *SagaExecutor) compensate(ctx context.Context) error {
    se.saga.State = SagaCompensating

    // 逆序执行补偿操作
    for i := se.saga.CurrentStep; i >= 0; i-- {
        step := se.saga.Steps[i]

        if step.Compensation != nil {
            if err := step.Compensation.Compensate(ctx, se.saga.Context); err != nil {
                se.logger.Error("Compensation failed", "step", step.Name, "error", err)
                // 补偿失败，需要人工干预
                return fmt.Errorf("compensation failed for step %s: %v", step.Name, err)
            }

            se.logger.Info("Compensation completed", "step", step.Name)
        }
    }

    se.saga.State = SagaCompensated
    return errors.New("saga compensated due to failure")
}

// Mall-Go电商订单Saga示例
type OrderSaga struct {
    OrderID   string  `json:"order_id"`
    UserID    string  `json:"user_id"`
    ProductID string  `json:"product_id"`
    Quantity  int     `json:"quantity"`
    Amount    float64 `json:"amount"`
}

// 创建订单Saga
func CreateOrderSaga(orderID, userID, productID string, quantity int, amount float64) *Saga {
    sagaCtx := SagaContext{
        Data: map[string]interface{}{
            "order_id":   orderID,
            "user_id":    userID,
            "product_id": productID,
            "quantity":   quantity,
            "amount":     amount,
        },
        Results:  make(map[string]interface{}),
        Metadata: make(map[string]interface{}),
    }

    return &Saga{
        ID:      orderID,
        Context: sagaCtx,
        Steps: []SagaStep{
            {
                Name:         "validate_order",
                Action:       &ValidateOrderAction{},
                Compensation: &ValidateOrderCompensation{},
                Timeout:      5 * time.Second,
                Retries:      2,
            },
            {
                Name:         "reserve_inventory",
                Action:       &ReserveInventoryAction{},
                Compensation: &ReserveInventoryCompensation{},
                Timeout:      10 * time.Second,
                Retries:      3,
            },
            {
                Name:         "process_payment",
                Action:       &ProcessPaymentAction{},
                Compensation: &ProcessPaymentCompensation{},
                Timeout:      30 * time.Second,
                Retries:      2,
            },
            {
                Name:         "create_order",
                Action:       &CreateOrderAction{},
                Compensation: &CreateOrderCompensation{},
                Timeout:      5 * time.Second,
                Retries:      2,
            },
            {
                Name:         "send_notification",
                Action:       &SendNotificationAction{},
                Compensation: nil, // 通知失败不需要补偿
                Timeout:      10 * time.Second,
                Retries:      1,
            },
        },
    }
}

// 验证订单动作
type ValidateOrderAction struct{}

func (v *ValidateOrderAction) Execute(ctx context.Context, sagaCtx SagaContext) error {
    userID := sagaCtx.Data["user_id"].(string)
    productID := sagaCtx.Data["product_id"].(string)
    quantity := sagaCtx.Data["quantity"].(int)

    // 验证用户
    if !isValidUser(userID) {
        return errors.New("invalid user")
    }

    // 验证商品
    if !isValidProduct(productID) {
        return errors.New("invalid product")
    }

    // 验证库存
    if !hasEnoughInventory(productID, quantity) {
        return errors.New("insufficient inventory")
    }

    sagaCtx.Results["validation"] = "success"
    return nil
}

// 库存预留动作
type ReserveInventoryAction struct{}

func (r *ReserveInventoryAction) Execute(ctx context.Context, sagaCtx SagaContext) error {
    productID := sagaCtx.Data["product_id"].(string)
    quantity := sagaCtx.Data["quantity"].(int)

    reservationID, err := reserveInventory(ctx, productID, quantity)
    if err != nil {
        return fmt.Errorf("failed to reserve inventory: %v", err)
    }

    sagaCtx.Results["reservation_id"] = reservationID
    return nil
}

// 库存预留补偿
type ReserveInventoryCompensation struct{}

func (r *ReserveInventoryCompensation) Compensate(ctx context.Context, sagaCtx SagaContext) error {
    reservationID, exists := sagaCtx.Results["reservation_id"]
    if !exists {
        return nil // 没有预留，无需补偿
    }

    return releaseInventoryReservation(ctx, reservationID.(string))
}

// 支付处理动作
type ProcessPaymentAction struct{}

func (p *ProcessPaymentAction) Execute(ctx context.Context, sagaCtx SagaContext) error {
    userID := sagaCtx.Data["user_id"].(string)
    amount := sagaCtx.Data["amount"].(float64)

    paymentID, err := processPayment(ctx, userID, amount)
    if err != nil {
        return fmt.Errorf("payment failed: %v", err)
    }

    sagaCtx.Results["payment_id"] = paymentID
    return nil
}

// 支付补偿
type ProcessPaymentCompensation struct{}

func (p *ProcessPaymentCompensation) Compensate(ctx context.Context, sagaCtx SagaContext) error {
    paymentID, exists := sagaCtx.Results["payment_id"]
    if !exists {
        return nil // 没有支付，无需补偿
    }

    return refundPayment(ctx, paymentID.(string))
}
```

---

## 🏢 Mall-Go项目分布式实践

### 电商分布式架构设计

```go
// Mall-Go分布式系统架构
package mall

import (
    "context"
    "time"
)

// 分布式电商系统
type DistributedMallSystem struct {
    // 服务注册与发现
    ServiceRegistry ServiceRegistry

    // 分布式锁
    DistributedLock DistributedLock

    // 分布式事务
    TransactionManager TransactionManager

    // 分布式缓存
    DistributedCache DistributedCache

    // 消息队列
    MessageQueue MessageQueue

    // 配置中心
    ConfigCenter ConfigCenter

    // 监控系统
    MonitoringSystem MonitoringSystem
}

// 分布式订单处理
type DistributedOrderProcessor struct {
    userService     UserService
    productService  ProductService
    inventoryService InventoryService
    paymentService  PaymentService
    orderService    OrderService
    notificationService NotificationService

    // 分布式组件
    distributedLock DistributedLock
    sagaExecutor    SagaExecutor
    eventBus        EventBus
}

// 处理订单创建
func (d *DistributedOrderProcessor) ProcessOrder(ctx context.Context, orderReq *OrderRequest) (*OrderResponse, error) {
    // 1. 分布式锁防止重复下单
    lockKey := fmt.Sprintf("order:user:%s", orderReq.UserID)

    if err := d.distributedLock.LockWithTimeout(ctx, lockKey, 30*time.Second); err != nil {
        return nil, fmt.Errorf("failed to acquire lock: %v", err)
    }
    defer d.distributedLock.Unlock(ctx, lockKey)

    // 2. 创建Saga事务
    saga := CreateOrderSaga(
        orderReq.OrderID,
        orderReq.UserID,
        orderReq.ProductID,
        orderReq.Quantity,
        orderReq.Amount,
    )

    // 3. 执行Saga
    if err := d.sagaExecutor.Execute(ctx, saga); err != nil {
        return nil, fmt.Errorf("order processing failed: %v", err)
    }

    // 4. 发布订单创建事件
    event := &OrderCreatedEvent{
        OrderID:   orderReq.OrderID,
        UserID:    orderReq.UserID,
        ProductID: orderReq.ProductID,
        Amount:    orderReq.Amount,
        Timestamp: time.Now(),
    }

    if err := d.eventBus.Publish(ctx, "order.created", event); err != nil {
        // 事件发布失败不影响主流程
        log.Warn("Failed to publish order created event", "error", err)
    }

    return &OrderResponse{
        OrderID: orderReq.OrderID,
        Status:  "created",
        Message: "Order created successfully",
    }, nil
}

// 分布式库存管理
type DistributedInventoryManager struct {
    inventoryDB     Database
    distributedLock DistributedLock
    cache          DistributedCache
    eventBus       EventBus
}

func (d *DistributedInventoryManager) ReserveInventory(ctx context.Context, productID string, quantity int) (string, error) {
    // 1. 分布式锁保证库存操作原子性
    lockKey := fmt.Sprintf("inventory:product:%s", productID)

    if err := d.distributedLock.Lock(ctx, lockKey); err != nil {
        return "", err
    }
    defer d.distributedLock.Unlock(ctx, lockKey)

    // 2. 检查库存
    currentInventory, err := d.getCurrentInventory(ctx, productID)
    if err != nil {
        return "", err
    }

    if currentInventory < quantity {
        return "", errors.New("insufficient inventory")
    }

    // 3. 预留库存
    reservationID := generateReservationID()

    tx, err := d.inventoryDB.BeginTx(ctx, nil)
    if err != nil {
        return "", err
    }
    defer tx.Rollback()

    // 更新库存
    _, err = tx.ExecContext(ctx,
        "UPDATE inventory SET available = available - ?, reserved = reserved + ? WHERE product_id = ?",
        quantity, quantity, productID)
    if err != nil {
        return "", err
    }

    // 记录预留
    _, err = tx.ExecContext(ctx,
        "INSERT INTO inventory_reservations (id, product_id, quantity, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
        reservationID, productID, quantity, time.Now(), time.Now().Add(30*time.Minute))
    if err != nil {
        return "", err
    }

    if err = tx.Commit(); err != nil {
        return "", err
    }

    // 4. 更新缓存
    d.cache.Delete(ctx, fmt.Sprintf("inventory:%s", productID))

    // 5. 发布库存变更事件
    event := &InventoryReservedEvent{
        ProductID:     productID,
        Quantity:      quantity,
        ReservationID: reservationID,
        Timestamp:     time.Now(),
    }
    d.eventBus.Publish(ctx, "inventory.reserved", event)

    return reservationID, nil
}
```

---

## 🎯 面试常考知识点

### 核心概念面试题

**Q1: 什么是CAP理论？请详细解释三个特性及其权衡。**

**标准答案：**
CAP理论是分布式系统设计的基础理论，指出分布式系统不能同时满足以下三个特性：

1. **一致性(Consistency)**：所有节点在同一时间看到相同的数据
2. **可用性(Availability)**：系统保持可操作状态，即使部分节点失败
3. **分区容错性(Partition tolerance)**：系统在网络分区时仍能继续运行

**权衡策略：**
- **CP系统**：牺牲可用性保证一致性，如MongoDB、HBase
- **AP系统**：牺牲一致性保证可用性，如Cassandra、DynamoDB
- **CA系统**：理论上存在，实际中不可能（网络分区不可避免）

**Q2: 解释强一致性、弱一致性、最终一致性的区别。**

**标准答案：**
```go
// 一致性级别对比
type ConsistencyLevels struct {
    Strong struct {
        Definition   string `json:"definition"`   // "所有节点同时看到相同数据"
        Guarantees   string `json:"guarantees"`   // "读取总是返回最新写入的值"
        Performance  string `json:"performance"`  // "性能较低，延迟较高"
        Examples     []string `json:"examples"`   // ["银行转账", "库存扣减"]
        Implementation string `json:"implementation"` // "同步复制、法定人数"
    } `json:"strong"`

    Weak struct {
        Definition   string `json:"definition"`   // "不保证何时达到一致性"
        Guarantees   string `json:"guarantees"`   // "可能读到旧数据"
        Performance  string `json:"performance"`  // "性能较高，延迟较低"
        Examples     []string `json:"examples"`   // ["DNS缓存", "CDN内容"]
        Implementation string `json:"implementation"` // "异步复制、缓存"
    } `json:"weak"`

    Eventual struct {
        Definition   string `json:"definition"`   // "最终会达到一致性"
        Guarantees   string `json:"guarantees"`   // "停止更新后最终一致"
        Performance  string `json:"performance"`  // "性能好，可用性高"
        Examples     []string `json:"examples"`   // ["社交媒体", "购物车"]
        Implementation string `json:"implementation"` // "异步复制、冲突解决"
    } `json:"eventual"`
}
```

**Q3: Raft算法的核心思想是什么？包含哪些关键组件？**

**标准答案：**
Raft算法将共识问题分解为三个子问题：

1. **领导者选举(Leader Election)**
   - 使用随机超时避免选举冲突
   - 候选者获得多数票成为领导者
   - 任期(Term)机制保证唯一性

2. **日志复制(Log Replication)**
   - 领导者接收客户端请求并复制到跟随者
   - 使用心跳机制维持权威
   - 多数派确认后提交日志

3. **安全性(Safety)**
   - 选举安全性：每个任期最多一个领导者
   - 日志匹配性：相同索引和任期的日志条目相同
   - 领导者完整性：已提交的日志不会丢失

**Q4: 分布式锁有哪些实现方式？各有什么优缺点？**

**标准答案：**

| 实现方式 | 优点 | 缺点 | 适用场景 |
|----------|------|------|----------|
| **Redis** | 性能高、实现简单 | 单点故障、时钟依赖 | 高性能、短时间锁 |
| **Zookeeper** | 强一致性、可靠性高 | 性能较低、复杂度高 | 配置管理、协调服务 |
| **etcd** | 强一致性、云原生 | 学习成本高 | Kubernetes、微服务 |
| **数据库** | 事务保证、简单 | 性能差、单点故障 | 传统应用、简单场景 |

**Q5: 什么是分布式事务？2PC和Saga模式的区别是什么？**

**标准答案：**

**分布式事务**：跨越多个网络节点的事务，需要保证ACID特性。

**2PC vs Saga对比：**

| 特性 | 2PC | Saga |
|------|-----|------|
| **一致性** | 强一致性 | 最终一致性 |
| **性能** | 较低（阻塞） | 较高（非阻塞） |
| **可用性** | 较低（协调者单点） | 较高（去中心化） |
| **复杂度** | 相对简单 | 较复杂（补偿逻辑） |
| **适用场景** | 短事务、强一致性要求 | 长事务、高可用要求 |
| **故障处理** | 阻塞等待 | 补偿回滚 |

### 技术实现面试题

**Q6: 如何实现一个高可用的分布式锁？**

**标准答案：**
```go
// 高可用分布式锁设计要点
type HighAvailabilityLockDesign struct {
    // 1. 多节点部署
    MultiNode struct {
        RedisCluster   bool `json:"redis_cluster"`   // Redis集群
        ZKEnsemble     bool `json:"zk_ensemble"`     // ZK集群
        EtcdCluster    bool `json:"etcd_cluster"`    // etcd集群
    } `json:"multi_node"`

    // 2. 故障检测
    FailureDetection struct {
        HealthCheck    bool `json:"health_check"`    // 健康检查
        Timeout        bool `json:"timeout"`         // 超时机制
        Heartbeat      bool `json:"heartbeat"`       // 心跳检测
    } `json:"failure_detection"`

    // 3. 自动恢复
    AutoRecovery struct {
        LockExpiration bool `json:"lock_expiration"` // 锁自动过期
        LeaderElection bool `json:"leader_election"` // 领导者选举
        Failover       bool `json:"failover"`        // 故障转移
    } `json:"auto_recovery"`

    // 4. 一致性保证
    ConsistencyGuarantee struct {
        Quorum         bool `json:"quorum"`          // 法定人数
        Consensus      bool `json:"consensus"`       // 共识算法
        Linearizability bool `json:"linearizability"` // 线性一致性
    } `json:"consistency_guarantee"`
}
```

**Q7: 分布式系统中如何处理时钟同步问题？**

**标准答案：**
1. **逻辑时钟(Logical Clock)**
   - Lamport时间戳：单调递增的逻辑时钟
   - 向量时钟：捕获因果关系的时钟

2. **物理时钟同步**
   - NTP协议：网络时间协议
   - PTP协议：精确时间协议
   - GPS时钟：卫星授时

3. **混合方案**
   - TrueTime(Google)：结合GPS和原子钟
   - HLC(Hybrid Logical Clock)：结合逻辑和物理时钟

**Q8: 如何设计一个分布式ID生成器？**

**标准答案：**
```go
// 分布式ID生成策略
type DistributedIDStrategies struct {
    // 1. Snowflake算法
    Snowflake struct {
        Timestamp   int `json:"timestamp"`   // 41位时间戳
        MachineID   int `json:"machine_id"`  // 10位机器ID
        Sequence    int `json:"sequence"`    // 12位序列号
        Advantages  []string `json:"advantages"` // ["高性能", "趋势递增", "无依赖"]
        Disadvantages []string `json:"disadvantages"` // ["时钟依赖", "机器ID管理"]
    } `json:"snowflake"`

    // 2. UUID
    UUID struct {
        Version     string `json:"version"`     // "UUID4"
        Advantages  []string `json:"advantages"` // ["全局唯一", "无依赖", "简单"]
        Disadvantages []string `json:"disadvantages"` // ["无序", "存储空间大"]
    } `json:"uuid"`

    // 3. 数据库自增
    DatabaseAutoIncrement struct {
        Advantages  []string `json:"advantages"` // ["有序", "简单", "唯一"]
        Disadvantages []string `json:"disadvantages"` // ["性能瓶颈", "单点故障"]
    } `json:"database_auto_increment"`

    // 4. Redis计数器
    RedisCounter struct {
        Advantages  []string `json:"advantages"` // ["高性能", "有序", "简单"]
        Disadvantages []string `json:"disadvantages"` // ["依赖Redis", "持久化问题"]
    } `json:"redis_counter"`
}
```

---

## 🏋️ 练习题

### 练习1：实现简化版Raft算法

**题目描述：**
实现一个简化版的Raft共识算法，支持领导者选举和日志复制功能。

**要求：**
1. 实现领导者选举机制
2. 实现日志复制功能
3. 支持节点故障检测和恢复
4. 实现基本的安全性保证
5. 提供状态查询接口

**参考实现框架：**
```go
type SimpleRaft struct {
    nodeID      string
    state       RaftState
    currentTerm int
    votedFor    *string
    log         []LogEntry
    peers       map[string]*RaftPeer

    // TODO: 实现以下方法
}

func (r *SimpleRaft) StartElection() error {
    // 实现选举逻辑
}

func (r *SimpleRaft) AppendEntries(request AppendEntriesRequest) AppendEntriesResponse {
    // 实现日志追加逻辑
}

func (r *SimpleRaft) RequestVote(request VoteRequest) VoteResponse {
    // 实现投票逻辑
}
```

### 练习2：设计电商分布式事务方案

**题目描述：**
为电商系统设计一个完整的分布式事务处理方案，处理订单创建流程。

**要求：**
1. 设计订单创建的完整流程
2. 选择合适的分布式事务模式
3. 实现事务的回滚和补偿机制
4. 考虑异常情况的处理
5. 设计监控和告警机制

**业务流程：**
- 用户下单 → 库存扣减 → 支付处理 → 订单确认 → 发送通知

### 练习3：实现分布式缓存一致性

**题目描述：**
实现一个分布式缓存系统，保证多个节点间的数据一致性。

**要求：**
1. 实现缓存数据的分布式存储
2. 设计缓存一致性协议
3. 实现缓存失效和更新机制
4. 支持缓存的故障恢复
5. 提供性能监控功能

**参考实现框架：**
```go
type DistributedCache struct {
    nodes       map[string]*CacheNode
    hashRing    *ConsistentHash
    replication int

    // TODO: 实现以下方法
}

func (d *DistributedCache) Get(key string) (interface{}, error) {
    // 实现分布式获取逻辑
}

func (d *DistributedCache) Set(key string, value interface{}, ttl time.Duration) error {
    // 实现分布式设置逻辑
}

func (d *DistributedCache) Delete(key string) error {
    // 实现分布式删除逻辑
}

func (d *DistributedCache) InvalidateAll(pattern string) error {
    // 实现批量失效逻辑
}
```

---

## 📚 章节总结

### 🎯 核心知识点回顾

通过本章学习，我们深入掌握了分布式系统的核心概念和关键技术：

1. **分布式系统基础**
   - 理解了分布式系统的定义、特征和挑战
   - 掌握了分布式系统的设计目标和透明性要求
   - 学会了分析分布式系统的复杂性和解决方案

2. **CAP理论深度理解**
   - 深入理解了一致性、可用性、分区容错性的含义
   - 掌握了不同CAP组合的适用场景和权衡策略
   - 学会了根据业务需求选择合适的CAP策略

3. **一致性模型**
   - 掌握了强一致性、弱一致性、最终一致性的区别
   - 理解了因果一致性、单调一致性等高级概念
   - 学会了在不同场景下选择合适的一致性级别

4. **共识算法**
   - 深入学习了Raft算法的原理和实现
   - 理解了领导者选举、日志复制、安全性保证
   - 掌握了共识算法在分布式系统中的应用

5. **分布式锁**
   - 掌握了基于Redis、Zookeeper等的分布式锁实现
   - 理解了分布式锁的设计要点和最佳实践
   - 学会了处理锁的超时、续期、死锁等问题

6. **分布式事务**
   - 深入理解了2PC、3PC、Saga等分布式事务模式
   - 掌握了事务的ACID特性在分布式环境下的挑战
   - 学会了根据业务特点选择合适的事务处理方案

7. **Go语言实现**
   - 掌握了使用Go实现分布式系统组件的核心技术
   - 学会了并发编程、网络通信、状态管理等技能
   - 理解了Go语言在分布式系统开发中的优势

8. **企业级实践**
   - 通过Mall-Go项目实践，掌握了电商场景的分布式设计
   - 学会了处理订单、库存、支付等复杂业务场景
   - 理解了分布式系统在实际项目中的应用模式

### 🚀 实践应用价值

1. **系统架构能力**：能够设计和实现大规模分布式系统架构
2. **技术选型能力**：能够根据业务需求选择合适的分布式技术方案
3. **问题解决能力**：掌握了分布式系统常见问题的解决方案
4. **性能优化能力**：理解了分布式系统的性能瓶颈和优化策略
5. **故障处理能力**：具备了分布式系统故障诊断和恢复的能力

### 🎓 下一步学习建议

1. **深入学习容器化技术**：学习下一章的Docker容器化部署
2. **实践项目应用**：在实际项目中应用分布式系统技术
3. **性能测试验证**：通过压测验证分布式系统的性能表现
4. **监控体系建设**：建立完善的分布式系统监控和告警体系
5. **持续学习新技术**：关注分布式系统领域的新技术和发展趋势

### 💡 关键技术要点

- **分布式系统设计要考虑CAP权衡**，没有完美的解决方案，只有合适的选择
- **一致性和性能往往是矛盾的**，需要根据业务需求找到平衡点
- **共识算法是分布式系统的核心**，理解其原理对系统设计至关重要
- **分布式锁要考虑各种异常情况**，包括网络分区、节点故障、时钟偏移等
- **分布式事务要权衡一致性和可用性**，选择合适的事务模式
- **故障是分布式系统的常态**，设计时要考虑各种故障场景
- **监控和可观测性非常重要**，有助于快速定位和解决问题

### 🌟 技术发展趋势

1. **云原生分布式系统**：与Kubernetes深度集成，支持自动扩缩容
2. **边缘计算分布式**：支持边缘节点的分布式计算和存储
3. **AI增强的分布式系统**：使用机器学习优化分布式系统性能
4. **量子计算影响**：量子计算对分布式系统安全性的影响
5. **绿色分布式计算**：关注能耗和环保的分布式系统设计

### 🔗 与其他章节的联系

- **服务发现**：为分布式系统提供服务注册和发现能力
- **API网关**：作为分布式系统的统一入口点
- **容器化部署**：为分布式系统提供标准化的部署方式
- **监控系统**：为分布式系统提供可观测性支持
- **性能优化**：分布式系统性能优化的具体实践

通过本章的学习，你已经具备了设计和实现企业级分布式系统的能力。分布式系统是现代软件架构的基础，掌握其核心概念和关键技术对于构建高可用、高性能、可扩展的系统至关重要！ 🚀

---

*"分布式系统让我们能够构建跨越时空的计算网络，它不仅是技术的进步，更是人类协作智慧的体现。掌握分布式系统，就是掌握了构建未来数字世界的核心能力！"* 🌐✨
```
```
```
```
```
```
