# Go语言并发编程基础

> *"Don't communicate by sharing memory; share memory by communicating."* - Rob Pike

欢迎来到Go语言最激动人心的部分——并发编程！🚀 如果说错误处理是Go的理性之美，那么并发编程就是Go的感性之美。Go的并发模型不仅优雅，而且强大，让我们能够轻松构建高性能的并发程序。

## 🎯 本章学习目标

通过本章学习，你将掌握：

- **Goroutine的工作原理** - 轻量级线程的奥秘 🧵
- **Channel通信机制** - Go并发的核心哲学 📡
- **Select多路复用** - 优雅的并发控制 🎛️
- **同步原语使用** - 传统并发工具的Go实现 🔒
- **并发安全编程** - 避免竞态条件的艺术 ⚡
- **Context上下文管理** - 优雅的取消和超时 ⏰
- **经典并发模式** - 工业级并发设计 🏭
- **性能调优技巧** - 让并发程序飞起来 🚁

## 📚 章节大纲

1. **Goroutine基础** - 轻量级并发的魅力
2. **Channel通信** - CSP模型的Go实现
3. **Select语句** - 多路复用的艺术
4. **同步原语** - 传统并发工具箱
5. **并发安全** - 竞态条件的预防和治理
6. **Context包** - 优雅的上下文管理
7. **并发模式** - 经典设计模式实战
8. **调试和优化** - 并发程序的性能调优
9. **实战案例** - Mall-Go项目并发实现
10. **练习题** - 从基础到高级的完整训练

---

## 🧵 Goroutine基础

### 什么是Goroutine？

Goroutine是Go语言的轻量级线程，是Go并发编程的基石。与传统的操作系统线程相比，Goroutine有着显著的优势。

#### 1. Goroutine vs 传统线程对比

```go
// Go - Goroutine创建
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    for i := 0; i < 5; i++ {
        fmt.Printf("Hello %s! (%d)\n", name, i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    // 创建Goroutine就这么简单！
    go sayHello("Alice")   // 启动第一个Goroutine
    go sayHello("Bob")     // 启动第二个Goroutine
    go sayHello("Charlie") // 启动第三个Goroutine
    
    // 主Goroutine等待一段时间
    time.Sleep(600 * time.Millisecond)
    fmt.Println("主程序结束")
}

/*
输出示例：
Hello Alice! (0)
Hello Bob! (0)
Hello Charlie! (0)
Hello Alice! (1)
Hello Bob! (1)
Hello Charlie! (1)
...
*/
```

```java
// Java - 传统线程创建
public class ThreadExample {
    public static void sayHello(String name) {
        for (int i = 0; i < 5; i++) {
            System.out.printf("Hello %s! (%d)%n", name, i);
            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
    }
    
    public static void main(String[] args) throws InterruptedException {
        // 创建和启动线程
        Thread t1 = new Thread(() -> sayHello("Alice"));
        Thread t2 = new Thread(() -> sayHello("Bob"));
        Thread t3 = new Thread(() -> sayHello("Charlie"));
        
        t1.start();
        t2.start();
        t3.start();
        
        // 等待所有线程完成
        t1.join();
        t2.join();
        t3.join();
        
        System.out.println("主程序结束");
    }
}
```

```python
# Python - 线程创建
import threading
import time

def say_hello(name):
    for i in range(5):
        print(f"Hello {name}! ({i})")
        time.sleep(0.1)

def main():
    # 创建线程
    threads = []
    for name in ["Alice", "Bob", "Charlie"]:
        t = threading.Thread(target=say_hello, args=(name,))
        threads.append(t)
        t.start()
    
    # 等待所有线程完成
    for t in threads:
        t.join()
    
    print("主程序结束")

if __name__ == "__main__":
    main()
```

#### 2. Goroutine的特点和优势

```go
// 来自 mall-go/pkg/concurrent/goroutine.go
package concurrent

import (
    "fmt"
    "runtime"
    "sync"
    "time"
)

// Goroutine特性演示
func DemonstrateGoroutineFeatures() {
    fmt.Println("=== Goroutine特性演示 ===")
    
    // 1. 轻量级 - 可以创建大量Goroutine
    demonstrateLightweight()
    
    // 2. 快速启动 - 启动开销极小
    demonstrateFastStartup()
    
    // 3. 动态栈 - 栈大小自动调整
    demonstrateDynamicStack()
    
    // 4. M:N调度 - 多个Goroutine映射到少数OS线程
    demonstrateScheduling()
}

// 轻量级特性：创建大量Goroutine
func demonstrateLightweight() {
    fmt.Println("\n1. 轻量级特性：")
    
    const numGoroutines = 100000
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // 模拟一些工作
            _ = id * 2
        }(i)
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    fmt.Printf("创建并执行 %d 个Goroutine耗时: %v\n", numGoroutines, duration)
    fmt.Printf("平均每个Goroutine: %v\n", duration/numGoroutines)
    
    // 显示内存使用情况
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("内存使用: %.2f MB\n", float64(m.Alloc)/1024/1024)
}

// 快速启动特性
func demonstrateFastStartup() {
    fmt.Println("\n2. 快速启动特性：")
    
    const iterations = 10000
    
    // 测试Goroutine启动时间
    start := time.Now()
    var wg sync.WaitGroup
    
    for i := 0; i < iterations; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
        }()
    }
    
    wg.Wait()
    goroutineDuration := time.Since(start)
    
    fmt.Printf("启动 %d 个Goroutine耗时: %v\n", iterations, goroutineDuration)
    fmt.Printf("平均启动时间: %v\n", goroutineDuration/iterations)
}

// 动态栈特性
func demonstrateDynamicStack() {
    fmt.Println("\n3. 动态栈特性：")
    
    // 递归函数测试栈增长
    var recursiveFunc func(int)
    recursiveFunc = func(depth int) {
        if depth <= 0 {
            return
        }
        
        // 创建一些栈变量
        var buffer [1024]byte
        _ = buffer
        
        recursiveFunc(depth - 1)
    }
    
    go func() {
        fmt.Println("开始深度递归...")
        recursiveFunc(1000) // 深度递归
        fmt.Println("递归完成，栈自动调整大小")
    }()
    
    time.Sleep(100 * time.Millisecond)
}

// M:N调度特性
func demonstrateScheduling() {
    fmt.Println("\n4. M:N调度特性：")
    
    fmt.Printf("CPU核心数: %d\n", runtime.NumCPU())
    fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
    
    // 创建比CPU核心数多得多的Goroutine
    const numWorkers = 1000
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // CPU密集型任务
            sum := 0
            for j := 0; j < 1000000; j++ {
                sum += j
            }
            
            if id < 5 { // 只打印前5个
                fmt.Printf("Worker %d 完成，结果: %d\n", id, sum)
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Printf("所有 %d 个Worker完成\n", numWorkers)
}
```

### Goroutine的生命周期管理

```go
// 来自 mall-go/internal/service/concurrent_service.go
package service

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "go.uber.org/zap"
)

type ConcurrentService struct {
    logger *zap.Logger
    wg     sync.WaitGroup
    ctx    context.Context
    cancel context.CancelFunc
}

func NewConcurrentService(logger *zap.Logger) *ConcurrentService {
    ctx, cancel := context.WithCancel(context.Background())
    return &ConcurrentService{
        logger: logger,
        ctx:    ctx,
        cancel: cancel,
    }
}

// 启动后台任务
func (s *ConcurrentService) StartBackgroundTasks() {
    s.logger.Info("启动后台任务")
    
    // 启动多个后台Goroutine
    s.startTask("数据清理", s.dataCleanupTask, 5*time.Second)
    s.startTask("健康检查", s.healthCheckTask, 10*time.Second)
    s.startTask("指标收集", s.metricsCollectionTask, 3*time.Second)
}

func (s *ConcurrentService) startTask(name string, taskFunc func(), interval time.Duration) {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        defer func() {
            if r := recover(); r != nil {
                s.logger.Error("任务panic", 
                    zap.String("task", name),
                    zap.Any("panic", r))
            }
        }()
        
        s.logger.Info("任务启动", zap.String("task", name))
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-s.ctx.Done():
                s.logger.Info("任务停止", zap.String("task", name))
                return
            case <-ticker.C:
                taskFunc()
            }
        }
    }()
}

func (s *ConcurrentService) dataCleanupTask() {
    s.logger.Debug("执行数据清理任务")
    // 模拟数据清理工作
    time.Sleep(100 * time.Millisecond)
}

func (s *ConcurrentService) healthCheckTask() {
    s.logger.Debug("执行健康检查任务")
    // 模拟健康检查工作
    time.Sleep(50 * time.Millisecond)
}

func (s *ConcurrentService) metricsCollectionTask() {
    s.logger.Debug("执行指标收集任务")
    // 模拟指标收集工作
    time.Sleep(200 * time.Millisecond)
}

// 优雅停止
func (s *ConcurrentService) Stop() {
    s.logger.Info("停止后台任务")
    s.cancel() // 发送取消信号
    s.wg.Wait() // 等待所有任务完成
    s.logger.Info("所有后台任务已停止")
}
```

---

## 📡 Channel通信机制

### Channel的基本概念

Channel是Go语言并发编程的核心，实现了CSP（Communicating Sequential Processes）模型。

#### 1. Channel类型和基本操作

```go
// 来自 mall-go/pkg/concurrent/channel.go
package concurrent

import (
    "fmt"
    "time"
)

// Channel基础操作演示
func DemonstrateChannelBasics() {
    fmt.Println("=== Channel基础操作 ===")
    
    // 1. 无缓冲Channel
    demonstrateUnbufferedChannel()
    
    // 2. 有缓冲Channel
    demonstrateBufferedChannel()
    
    // 3. 单向Channel
    demonstrateDirectionalChannel()
    
    // 4. Channel关闭
    demonstrateChannelClose()
}

// 无缓冲Channel - 同步通信
func demonstrateUnbufferedChannel() {
    fmt.Println("\n1. 无缓冲Channel：")
    
    // 创建无缓冲Channel
    ch := make(chan string)
    
    // 启动接收者Goroutine
    go func() {
        message := <-ch // 阻塞等待消息
        fmt.Printf("接收到消息: %s\n", message)
    }()
    
    // 发送消息（会阻塞直到有接收者）
    fmt.Println("发送消息...")
    ch <- "Hello, Channel!"
    fmt.Println("消息已发送")
    
    time.Sleep(100 * time.Millisecond)
}

// 有缓冲Channel - 异步通信
func demonstrateBufferedChannel() {
    fmt.Println("\n2. 有缓冲Channel：")
    
    // 创建容量为3的缓冲Channel
    ch := make(chan int, 3)
    
    // 发送数据（不会阻塞，直到缓冲区满）
    ch <- 1
    ch <- 2
    ch <- 3
    fmt.Printf("缓冲区已满，长度: %d, 容量: %d\n", len(ch), cap(ch))
    
    // 接收数据
    for i := 0; i < 3; i++ {
        value := <-ch
        fmt.Printf("接收到: %d\n", value)
    }
}

// 单向Channel - 类型安全
func demonstrateDirectionalChannel() {
    fmt.Println("\n3. 单向Channel：")
    
    // 双向Channel
    ch := make(chan string, 1)
    
    // 只发送Channel
    sendOnly := func(ch chan<- string) {
        ch <- "只能发送"
    }
    
    // 只接收Channel
    receiveOnly := func(ch <-chan string) {
        message := <-ch
        fmt.Printf("只能接收: %s\n", message)
    }
    
    go sendOnly(ch)
    go receiveOnly(ch)
    
    time.Sleep(100 * time.Millisecond)
}

// Channel关闭
func demonstrateChannelClose() {
    fmt.Println("\n4. Channel关闭：")
    
    ch := make(chan int, 3)
    
    // 发送数据
    go func() {
        for i := 1; i <= 5; i++ {
            ch <- i
            fmt.Printf("发送: %d\n", i)
        }
        close(ch) // 关闭Channel
        fmt.Println("Channel已关闭")
    }()
    
    // 接收数据直到Channel关闭
    for {
        value, ok := <-ch
        if !ok {
            fmt.Println("Channel已关闭，退出接收循环")
            break
        }
        fmt.Printf("接收: %d\n", value)
    }
    
    // 使用range接收（更简洁）
    fmt.Println("\n使用range接收：")
    ch2 := make(chan string, 2)
    
    go func() {
        ch2 <- "消息1"
        ch2 <- "消息2"
        close(ch2)
    }()
    
    for message := range ch2 {
        fmt.Printf("Range接收: %s\n", message)
    }
}
```

#### 2. Channel设计模式

```go
// 来自 mall-go/pkg/patterns/channel_patterns.go
package patterns

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

// 生产者-消费者模式
func ProducerConsumerPattern() {
    fmt.Println("=== 生产者-消费者模式 ===")

    // 创建任务Channel
    taskChan := make(chan Task, 10)
    resultChan := make(chan Result, 10)

    // 启动生产者
    go producer(taskChan)

    // 启动多个消费者
    const numConsumers = 3
    var wg sync.WaitGroup

    for i := 0; i < numConsumers; i++ {
        wg.Add(1)
        go consumer(i, taskChan, resultChan, &wg)
    }

    // 启动结果收集器
    go resultCollector(resultChan)

    // 等待消费者完成
    wg.Wait()
    close(resultChan)

    time.Sleep(100 * time.Millisecond)
}

type Task struct {
    ID   int
    Data string
}

type Result struct {
    TaskID int
    Output string
    Worker int
}

func producer(taskChan chan<- Task) {
    defer close(taskChan)

    for i := 1; i <= 10; i++ {
        task := Task{
            ID:   i,
            Data: fmt.Sprintf("任务数据-%d", i),
        }

        taskChan <- task
        fmt.Printf("生产者: 生产任务 %d\n", i)
        time.Sleep(100 * time.Millisecond)
    }

    fmt.Println("生产者: 所有任务已生产完成")
}

func consumer(workerID int, taskChan <-chan Task, resultChan chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()

    for task := range taskChan {
        fmt.Printf("消费者%d: 处理任务 %d\n", workerID, task.ID)

        // 模拟处理时间
        processingTime := time.Duration(rand.Intn(500)) * time.Millisecond
        time.Sleep(processingTime)

        result := Result{
            TaskID: task.ID,
            Output: fmt.Sprintf("处理结果-%d", task.ID),
            Worker: workerID,
        }

        resultChan <- result
        fmt.Printf("消费者%d: 完成任务 %d\n", workerID, task.ID)
    }

    fmt.Printf("消费者%d: 退出\n", workerID)
}

func resultCollector(resultChan <-chan Result) {
    fmt.Println("结果收集器: 开始收集结果")

    for result := range resultChan {
        fmt.Printf("结果收集器: 任务%d的结果 - %s (由消费者%d处理)\n",
            result.TaskID, result.Output, result.Worker)
    }

    fmt.Println("结果收集器: 所有结果已收集完成")
}

// 管道模式 (Pipeline)
func PipelinePattern() {
    fmt.Println("\n=== 管道模式 ===")

    // 创建管道阶段
    numbers := generateNumbers(10)
    squares := squareNumbers(numbers)
    results := filterEven(squares)

    // 消费最终结果
    fmt.Println("管道处理结果:")
    for result := range results {
        fmt.Printf("最终结果: %d\n", result)
    }
}

// 第一阶段：生成数字
func generateNumbers(count int) <-chan int {
    out := make(chan int)

    go func() {
        defer close(out)
        for i := 1; i <= count; i++ {
            out <- i
            fmt.Printf("生成: %d\n", i)
        }
    }()

    return out
}

// 第二阶段：计算平方
func squareNumbers(in <-chan int) <-chan int {
    out := make(chan int)

    go func() {
        defer close(out)
        for num := range in {
            square := num * num
            out <- square
            fmt.Printf("平方: %d -> %d\n", num, square)
        }
    }()

    return out
}

// 第三阶段：过滤偶数
func filterEven(in <-chan int) <-chan int {
    out := make(chan int)

    go func() {
        defer close(out)
        for num := range in {
            if num%2 == 0 {
                out <- num
                fmt.Printf("过滤: %d (偶数)\n", num)
            } else {
                fmt.Printf("过滤: %d (奇数，丢弃)\n", num)
            }
        }
    }()

    return out
}

// 扇入模式 (Fan-in)
func FanInPattern() {
    fmt.Println("\n=== 扇入模式 ===")

    // 创建多个输入Channel
    input1 := make(chan string)
    input2 := make(chan string)
    input3 := make(chan string)

    // 启动多个生产者
    go func() {
        defer close(input1)
        for i := 1; i <= 3; i++ {
            input1 <- fmt.Sprintf("来源1-消息%d", i)
            time.Sleep(100 * time.Millisecond)
        }
    }()

    go func() {
        defer close(input2)
        for i := 1; i <= 3; i++ {
            input2 <- fmt.Sprintf("来源2-消息%d", i)
            time.Sleep(150 * time.Millisecond)
        }
    }()

    go func() {
        defer close(input3)
        for i := 1; i <= 3; i++ {
            input3 <- fmt.Sprintf("来源3-消息%d", i)
            time.Sleep(200 * time.Millisecond)
        }
    }()

    // 扇入合并
    merged := fanIn(input1, input2, input3)

    // 消费合并后的消息
    for message := range merged {
        fmt.Printf("扇入结果: %s\n", message)
    }
}

func fanIn(inputs ...<-chan string) <-chan string {
    out := make(chan string)
    var wg sync.WaitGroup

    // 为每个输入Channel启动一个Goroutine
    for _, input := range inputs {
        wg.Add(1)
        go func(ch <-chan string) {
            defer wg.Done()
            for message := range ch {
                out <- message
            }
        }(input)
    }

    // 等待所有输入完成后关闭输出Channel
    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

// 扇出模式 (Fan-out)
func FanOutPattern() {
    fmt.Println("\n=== 扇出模式 ===")

    // 创建输入Channel
    input := make(chan int, 10)

    // 生产数据
    go func() {
        defer close(input)
        for i := 1; i <= 10; i++ {
            input <- i
            fmt.Printf("生产: %d\n", i)
        }
    }()

    // 扇出到多个处理器
    const numProcessors = 3
    var wg sync.WaitGroup

    for i := 0; i < numProcessors; i++ {
        wg.Add(1)
        go processor(i, input, &wg)
    }

    wg.Wait()
    fmt.Println("所有处理器完成")
}

func processor(id int, input <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for data := range input {
        // 模拟处理时间
        time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
        fmt.Printf("处理器%d: 处理数据 %d\n", id, data)
    }

    fmt.Printf("处理器%d: 完成\n", id)
}
```

---

## 🎛️ Select语句和多路复用

### Select的基本用法

Select语句是Go并发编程中的多路复用器，类似于网络编程中的select/poll。

```go
// 来自 mall-go/pkg/concurrent/select_demo.go
package concurrent

import (
    "fmt"
    "math/rand"
    "time"
)

// Select基础用法演示
func DemonstrateSelectBasics() {
    fmt.Println("=== Select基础用法 ===")

    // 1. 基本Select
    demonstrateBasicSelect()

    // 2. 非阻塞Select
    demonstrateNonBlockingSelect()

    // 3. 超时处理
    demonstrateTimeout()

    // 4. Select with default
    demonstrateSelectDefault()
}

// 基本Select用法
func demonstrateBasicSelect() {
    fmt.Println("\n1. 基本Select用法：")

    ch1 := make(chan string)
    ch2 := make(chan string)

    // 启动两个Goroutine发送数据
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "来自Channel 1的消息"
    }()

    go func() {
        time.Sleep(200 * time.Millisecond)
        ch2 <- "来自Channel 2的消息"
    }()

    // 使用Select等待任一Channel
    for i := 0; i < 2; i++ {
        select {
        case msg1 := <-ch1:
            fmt.Printf("接收到: %s\n", msg1)
        case msg2 := <-ch2:
            fmt.Printf("接收到: %s\n", msg2)
        }
    }
}

// 非阻塞Select
func demonstrateNonBlockingSelect() {
    fmt.Println("\n2. 非阻塞Select：")

    ch := make(chan string, 1)

    // 尝试非阻塞发送
    select {
    case ch <- "非阻塞发送":
        fmt.Println("发送成功")
    default:
        fmt.Println("Channel已满，发送失败")
    }

    // 尝试非阻塞接收
    select {
    case msg := <-ch:
        fmt.Printf("接收到: %s\n", msg)
    default:
        fmt.Println("Channel为空，接收失败")
    }

    // 再次尝试接收（此时Channel为空）
    select {
    case msg := <-ch:
        fmt.Printf("接收到: %s\n", msg)
    default:
        fmt.Println("Channel为空，使用默认处理")
    }
}

// 超时处理
func demonstrateTimeout() {
    fmt.Println("\n3. 超时处理：")

    ch := make(chan string)

    // 启动一个慢速的Goroutine
    go func() {
        time.Sleep(2 * time.Second)
        ch <- "延迟消息"
    }()

    // 使用Select实现超时
    select {
    case msg := <-ch:
        fmt.Printf("接收到消息: %s\n", msg)
    case <-time.After(1 * time.Second):
        fmt.Println("操作超时！")
    }
}

// Select with default
func demonstrateSelectDefault() {
    fmt.Println("\n4. Select with default：")

    ch := make(chan int, 2)

    // 填充Channel
    ch <- 1
    ch <- 2

    // 持续尝试发送，直到Channel满
    for i := 3; i <= 5; i++ {
        select {
        case ch <- i:
            fmt.Printf("发送成功: %d\n", i)
        default:
            fmt.Printf("Channel已满，无法发送: %d\n", i)

            // 接收一个值腾出空间
            received := <-ch
            fmt.Printf("接收了: %d，腾出空间\n", received)

            // 重新尝试发送
            ch <- i
            fmt.Printf("重新发送成功: %d\n", i)
        }
    }
}
```

### Select高级应用

```go
// 来自 mall-go/internal/service/notification_service.go
package service

import (
    "context"
    "fmt"
    "sync"
    "time"

    "go.uber.org/zap"
)

// 通知服务 - Select高级应用示例
type NotificationService struct {
    logger        *zap.Logger
    emailChan     chan EmailNotification
    smsChan       chan SMSNotification
    pushChan      chan PushNotification
    priorityChan  chan PriorityNotification
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
}

type EmailNotification struct {
    To      string
    Subject string
    Body    string
}

type SMSNotification struct {
    Phone   string
    Message string
}

type PushNotification struct {
    DeviceID string
    Title    string
    Message  string
}

type PriorityNotification struct {
    Type    string
    Content interface{}
    Level   int // 1-高优先级, 2-中优先级, 3-低优先级
}

func NewNotificationService(logger *zap.Logger) *NotificationService {
    ctx, cancel := context.WithCancel(context.Background())

    return &NotificationService{
        logger:       logger,
        emailChan:    make(chan EmailNotification, 100),
        smsChan:      make(chan SMSNotification, 50),
        pushChan:     make(chan PushNotification, 200),
        priorityChan: make(chan PriorityNotification, 10),
        ctx:          ctx,
        cancel:       cancel,
    }
}

// 启动通知处理器
func (ns *NotificationService) Start() {
    ns.logger.Info("启动通知服务")

    ns.wg.Add(1)
    go ns.notificationProcessor()
}

// 通知处理器 - 使用Select处理多种通知类型
func (ns *NotificationService) notificationProcessor() {
    defer ns.wg.Done()

    // 创建定时器用于批量处理
    batchTimer := time.NewTicker(5 * time.Second)
    defer batchTimer.Stop()

    // 批量处理缓存
    var emailBatch []EmailNotification
    var smsBatch []SMSNotification
    var pushBatch []PushNotification

    for {
        select {
        case <-ns.ctx.Done():
            ns.logger.Info("通知服务停止")
            return

        case priority := <-ns.priorityChan:
            // 优先级通知立即处理
            ns.processPriorityNotification(priority)

        case email := <-ns.emailChan:
            // 邮件通知加入批量处理
            emailBatch = append(emailBatch, email)
            ns.logger.Debug("邮件通知加入批量队列",
                zap.String("to", email.To),
                zap.Int("batch_size", len(emailBatch)))

        case sms := <-ns.smsChan:
            // 短信通知加入批量处理
            smsBatch = append(smsBatch, sms)
            ns.logger.Debug("短信通知加入批量队列",
                zap.String("phone", sms.Phone),
                zap.Int("batch_size", len(smsBatch)))

        case push := <-ns.pushChan:
            // 推送通知加入批量处理
            pushBatch = append(pushBatch, push)
            ns.logger.Debug("推送通知加入批量队列",
                zap.String("device_id", push.DeviceID),
                zap.Int("batch_size", len(pushBatch)))

        case <-batchTimer.C:
            // 定时批量处理
            if len(emailBatch) > 0 {
                ns.processBatchEmails(emailBatch)
                emailBatch = emailBatch[:0] // 清空切片
            }

            if len(smsBatch) > 0 {
                ns.processBatchSMS(smsBatch)
                smsBatch = smsBatch[:0]
            }

            if len(pushBatch) > 0 {
                ns.processBatchPush(pushBatch)
                pushBatch = pushBatch[:0]
            }

        default:
            // 非阻塞检查，避免CPU空转
            time.Sleep(10 * time.Millisecond)
        }

        // 检查批量大小，达到阈值立即处理
        if len(emailBatch) >= 10 {
            ns.processBatchEmails(emailBatch)
            emailBatch = emailBatch[:0]
        }

        if len(smsBatch) >= 5 {
            ns.processBatchSMS(smsBatch)
            smsBatch = smsBatch[:0]
        }

        if len(pushBatch) >= 20 {
            ns.processBatchPush(pushBatch)
            pushBatch = pushBatch[:0]
        }
    }
}

// 处理优先级通知
func (ns *NotificationService) processPriorityNotification(notification PriorityNotification) {
    ns.logger.Info("处理优先级通知",
        zap.String("type", notification.Type),
        zap.Int("level", notification.Level))

    // 根据优先级决定处理方式
    switch notification.Level {
    case 1: // 高优先级 - 立即处理
        ns.processHighPriorityNotification(notification)
    case 2: // 中优先级 - 快速处理
        ns.processMediumPriorityNotification(notification)
    case 3: // 低优先级 - 延迟处理
        ns.processLowPriorityNotification(notification)
    }
}

func (ns *NotificationService) processHighPriorityNotification(notification PriorityNotification) {
    // 高优先级通知的处理逻辑
    ns.logger.Warn("处理高优先级通知", zap.Any("content", notification.Content))
    // 实际实现：立即发送，可能使用多种渠道
}

func (ns *NotificationService) processMediumPriorityNotification(notification PriorityNotification) {
    // 中优先级通知的处理逻辑
    ns.logger.Info("处理中优先级通知", zap.Any("content", notification.Content))
    // 实际实现：快速处理，单一渠道
}

func (ns *NotificationService) processLowPriorityNotification(notification PriorityNotification) {
    // 低优先级通知的处理逻辑
    ns.logger.Debug("处理低优先级通知", zap.Any("content", notification.Content))
    // 实际实现：加入队列，延迟处理
}

// 批量处理邮件
func (ns *NotificationService) processBatchEmails(emails []EmailNotification) {
    ns.logger.Info("批量处理邮件", zap.Int("count", len(emails)))

    // 模拟批量发送邮件
    for _, email := range emails {
        ns.logger.Debug("发送邮件",
            zap.String("to", email.To),
            zap.String("subject", email.Subject))
        // 实际实现：调用邮件服务API
    }
}

// 批量处理短信
func (ns *NotificationService) processBatchSMS(messages []SMSNotification) {
    ns.logger.Info("批量处理短信", zap.Int("count", len(messages)))

    // 模拟批量发送短信
    for _, sms := range messages {
        ns.logger.Debug("发送短信",
            zap.String("phone", sms.Phone),
            zap.String("message", sms.Message))
        // 实际实现：调用短信服务API
    }
}

// 批量处理推送
func (ns *NotificationService) processBatchPush(notifications []PushNotification) {
    ns.logger.Info("批量处理推送", zap.Int("count", len(notifications)))

    // 模拟批量推送
    for _, push := range notifications {
        ns.logger.Debug("发送推送",
            zap.String("device_id", push.DeviceID),
            zap.String("title", push.Title))
        // 实际实现：调用推送服务API
    }
}

// 发送邮件通知
func (ns *NotificationService) SendEmail(to, subject, body string) error {
    email := EmailNotification{
        To:      to,
        Subject: subject,
        Body:    body,
    }

    select {
    case ns.emailChan <- email:
        return nil
    case <-time.After(1 * time.Second):
        return fmt.Errorf("邮件通知队列已满")
    }
}

// 发送短信通知
func (ns *NotificationService) SendSMS(phone, message string) error {
    sms := SMSNotification{
        Phone:   phone,
        Message: message,
    }

    select {
    case ns.smsChan <- sms:
        return nil
    case <-time.After(1 * time.Second):
        return fmt.Errorf("短信通知队列已满")
    }
}

// 发送推送通知
func (ns *NotificationService) SendPush(deviceID, title, message string) error {
    push := PushNotification{
        DeviceID: deviceID,
        Title:    title,
        Message:  message,
    }

    select {
    case ns.pushChan <- push:
        return nil
    case <-time.After(1 * time.Second):
        return fmt.Errorf("推送通知队列已满")
    }
}

// 发送优先级通知
func (ns *NotificationService) SendPriorityNotification(notificationType string, content interface{}, level int) error {
    priority := PriorityNotification{
        Type:    notificationType,
        Content: content,
        Level:   level,
    }

    select {
    case ns.priorityChan <- priority:
        return nil
    case <-time.After(500 * time.Millisecond):
        return fmt.Errorf("优先级通知队列已满")
    }
}

// 停止服务
func (ns *NotificationService) Stop() {
    ns.logger.Info("停止通知服务")
    ns.cancel()
    ns.wg.Wait()
    ns.logger.Info("通知服务已停止")
}
```

---

## 🔒 同步原语

### Mutex和RWMutex

虽然Go推荐使用Channel进行通信，但在某些场景下，传统的同步原语仍然有其用武之地。

```go
// 来自 mall-go/pkg/concurrent/sync_primitives.go
package concurrent

import (
    "fmt"
    "sync"
    "time"
)

// 同步原语演示
func DemonstrateSyncPrimitives() {
    fmt.Println("=== 同步原语演示 ===")

    // 1. Mutex使用
    demonstrateMutex()

    // 2. RWMutex使用
    demonstrateRWMutex()

    // 3. WaitGroup使用
    demonstrateWaitGroup()

    // 4. Once使用
    demonstrateOnce()
}

// Mutex演示 - 保护共享资源
func demonstrateMutex() {
    fmt.Println("\n1. Mutex演示：")

    // 共享计数器
    type Counter struct {
        mu    sync.Mutex
        value int
    }

    counter := &Counter{}

    // 启动多个Goroutine并发修改计数器
    var wg sync.WaitGroup
    const numGoroutines = 10
    const incrementsPerGoroutine = 1000

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < incrementsPerGoroutine; j++ {
                // 使用Mutex保护临界区
                counter.mu.Lock()
                counter.value++
                counter.mu.Unlock()
            }

            fmt.Printf("Goroutine %d 完成\n", id)
        }(i)
    }

    wg.Wait()

    expected := numGoroutines * incrementsPerGoroutine
    fmt.Printf("期望值: %d, 实际值: %d\n", expected, counter.value)

    if counter.value == expected {
        fmt.Println("✅ Mutex正确保护了共享资源")
    } else {
        fmt.Println("❌ 存在竞态条件")
    }
}

// RWMutex演示 - 读写锁
func demonstrateRWMutex() {
    fmt.Println("\n2. RWMutex演示：")

    // 共享数据结构
    type SafeMap struct {
        mu   sync.RWMutex
        data map[string]int
    }

    safeMap := &SafeMap{
        data: make(map[string]int),
    }

    // 写入方法
    safeMap.Set = func(key string, value int) {
        safeMap.mu.Lock()
        defer safeMap.mu.Unlock()
        safeMap.data[key] = value
        fmt.Printf("写入: %s = %d\n", key, value)
    }

    // 读取方法
    safeMap.Get = func(key string) (int, bool) {
        safeMap.mu.RLock()
        defer safeMap.mu.RUnlock()
        value, exists := safeMap.data[key]
        fmt.Printf("读取: %s = %d (存在: %t)\n", key, value, exists)
        return value, exists
    }

    var wg sync.WaitGroup

    // 启动写入Goroutine
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < 5; i++ {
            key := fmt.Sprintf("key%d", i)
            safeMap.Set(key, i*10)
            time.Sleep(100 * time.Millisecond)
        }
    }()

    // 启动多个读取Goroutine
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(readerID int) {
            defer wg.Done()
            for j := 0; j < 5; j++ {
                key := fmt.Sprintf("key%d", j)
                safeMap.Get(key)
                time.Sleep(50 * time.Millisecond)
            }
        }(i)
    }

    wg.Wait()
    fmt.Println("RWMutex演示完成")
}

// WaitGroup演示 - 等待多个Goroutine完成
func demonstrateWaitGroup() {
    fmt.Println("\n3. WaitGroup演示：")

    var wg sync.WaitGroup

    // 模拟多个任务
    tasks := []string{"任务A", "任务B", "任务C", "任务D", "任务E"}

    for _, task := range tasks {
        wg.Add(1) // 增加等待计数

        go func(taskName string) {
            defer wg.Done() // 完成时减少计数

            fmt.Printf("开始执行: %s\n", taskName)

            // 模拟任务执行时间
            duration := time.Duration(100+len(taskName)*50) * time.Millisecond
            time.Sleep(duration)

            fmt.Printf("完成执行: %s\n", taskName)
        }(task)
    }

    fmt.Println("等待所有任务完成...")
    wg.Wait() // 等待所有Goroutine完成
    fmt.Println("所有任务已完成")
}

// Once演示 - 确保函数只执行一次
func demonstrateOnce() {
    fmt.Println("\n4. Once演示：")

    var once sync.Once
    var initValue string

    // 初始化函数
    initFunc := func() {
        fmt.Println("执行初始化...")
        time.Sleep(100 * time.Millisecond) // 模拟初始化耗时
        initValue = "已初始化"
        fmt.Println("初始化完成")
    }

    var wg sync.WaitGroup

    // 启动多个Goroutine尝试初始化
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            fmt.Printf("Goroutine %d 尝试初始化\n", id)
            once.Do(initFunc) // 只有第一次调用会执行
            fmt.Printf("Goroutine %d 看到的值: %s\n", id, initValue)
        }(i)
    }

    wg.Wait()
    fmt.Println("Once演示完成")
}
```

### 原子操作

```go
// 来自 mall-go/pkg/concurrent/atomic_demo.go
package concurrent

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// 原子操作演示
func DemonstrateAtomicOperations() {
    fmt.Println("=== 原子操作演示 ===")

    // 1. 基本原子操作
    demonstrateBasicAtomic()

    // 2. 原子操作vs Mutex性能对比
    compareAtomicVsMutex()

    // 3. 原子操作的实际应用
    demonstrateAtomicUseCases()
}

// 基本原子操作
func demonstrateBasicAtomic() {
    fmt.Println("\n1. 基本原子操作：")

    var counter int64
    var wg sync.WaitGroup

    const numGoroutines = 10
    const incrementsPerGoroutine = 1000

    // 使用原子操作并发增加计数器
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < incrementsPerGoroutine; j++ {
                atomic.AddInt64(&counter, 1)
            }

            fmt.Printf("Goroutine %d 完成\n", id)
        }(i)
    }

    wg.Wait()

    expected := int64(numGoroutines * incrementsPerGoroutine)
    actual := atomic.LoadInt64(&counter)

    fmt.Printf("期望值: %d, 实际值: %d\n", expected, actual)

    if actual == expected {
        fmt.Println("✅ 原子操作正确保护了共享变量")
    } else {
        fmt.Println("❌ 存在问题")
    }
}

// 原子操作vs Mutex性能对比
func compareAtomicVsMutex() {
    fmt.Println("\n2. 原子操作vs Mutex性能对比：")

    const operations = 1000000
    const numGoroutines = 10

    // 测试原子操作性能
    var atomicCounter int64
    var atomicWg sync.WaitGroup

    start := time.Now()
    for i := 0; i < numGoroutines; i++ {
        atomicWg.Add(1)
        go func() {
            defer atomicWg.Done()
            for j := 0; j < operations/numGoroutines; j++ {
                atomic.AddInt64(&atomicCounter, 1)
            }
        }()
    }
    atomicWg.Wait()
    atomicDuration := time.Since(start)

    // 测试Mutex性能
    var mutexCounter int64
    var mu sync.Mutex
    var mutexWg sync.WaitGroup

    start = time.Now()
    for i := 0; i < numGoroutines; i++ {
        mutexWg.Add(1)
        go func() {
            defer mutexWg.Done()
            for j := 0; j < operations/numGoroutines; j++ {
                mu.Lock()
                mutexCounter++
                mu.Unlock()
            }
        }()
    }
    mutexWg.Wait()
    mutexDuration := time.Since(start)

    fmt.Printf("原子操作耗时: %v (结果: %d)\n", atomicDuration, atomicCounter)
    fmt.Printf("Mutex耗时: %v (结果: %d)\n", mutexDuration, mutexCounter)
    fmt.Printf("原子操作比Mutex快: %.2fx\n", float64(mutexDuration)/float64(atomicDuration))
}

// 原子操作的实际应用
func demonstrateAtomicUseCases() {
    fmt.Println("\n3. 原子操作实际应用：")

    // 应用场景1：统计计数器
    demonstrateStatsCounter()

    // 应用场景2：配置热更新
    demonstrateConfigUpdate()

    // 应用场景3：状态标志
    demonstrateStatusFlag()
}

// 统计计数器
func demonstrateStatsCounter() {
    fmt.Println("\n统计计数器应用：")

    type StatsCounter struct {
        requests    int64
        errors      int64
        successRate float64
    }

    stats := &StatsCounter{}

    var wg sync.WaitGroup

    // 模拟请求处理
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(requestID int) {
            defer wg.Done()

            // 增加请求计数
            atomic.AddInt64(&stats.requests, 1)

            // 模拟处理，10%概率出错
            if requestID%10 == 0 {
                atomic.AddInt64(&stats.errors, 1)
            }

            // 模拟处理时间
            time.Sleep(time.Millisecond)
        }(i)
    }

    wg.Wait()

    requests := atomic.LoadInt64(&stats.requests)
    errors := atomic.LoadInt64(&stats.errors)
    successRate := float64(requests-errors) / float64(requests) * 100

    fmt.Printf("总请求数: %d\n", requests)
    fmt.Printf("错误数: %d\n", errors)
    fmt.Printf("成功率: %.2f%%\n", successRate)
}

// 配置热更新
func demonstrateConfigUpdate() {
    fmt.Println("\n配置热更新应用：")

    type Config struct {
        maxConnections int64
        timeout        int64 // 毫秒
    }

    // 使用原子操作存储配置指针
    var configPtr atomic.Value

    // 初始配置
    initialConfig := &Config{
        maxConnections: 100,
        timeout:        5000,
    }
    configPtr.Store(initialConfig)

    // 模拟配置使用者
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()

            for j := 0; j < 3; j++ {
                config := configPtr.Load().(*Config)
                fmt.Printf("Worker %d 使用配置: 最大连接=%d, 超时=%dms\n",
                    workerID, config.maxConnections, config.timeout)
                time.Sleep(100 * time.Millisecond)
            }
        }(i)
    }

    // 模拟配置更新
    go func() {
        time.Sleep(150 * time.Millisecond)

        newConfig := &Config{
            maxConnections: 200,
            timeout:        3000,
        }
        configPtr.Store(newConfig)
        fmt.Println("🔄 配置已更新")
    }()

    wg.Wait()
}

// 状态标志
func demonstrateStatusFlag() {
    fmt.Println("\n状态标志应用：")

    var isShuttingDown int32
    var wg sync.WaitGroup

    // 启动工作Goroutine
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()

            for {
                // 检查关闭标志
                if atomic.LoadInt32(&isShuttingDown) == 1 {
                    fmt.Printf("Worker %d 检测到关闭信号，退出\n", workerID)
                    return
                }

                // 模拟工作
                fmt.Printf("Worker %d 正在工作...\n", workerID)
                time.Sleep(200 * time.Millisecond)
            }
        }(i)
    }

    // 模拟运行一段时间后关闭
    time.Sleep(600 * time.Millisecond)

    fmt.Println("🛑 发送关闭信号")
    atomic.StoreInt32(&isShuttingDown, 1)

    wg.Wait()
    fmt.Println("所有Worker已停止")
}
```

---

## ⚡ 并发安全和竞态条件

### 竞态条件的识别和避免

```go
// 来自 mall-go/pkg/concurrent/race_conditions.go
package concurrent

import (
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)

// 竞态条件演示和解决方案
func DemonstrateRaceConditions() {
    fmt.Println("=== 竞态条件演示 ===")

    // 1. 典型的竞态条件
    demonstrateRaceCondition()

    // 2. 使用Mutex解决
    demonstrateMutexSolution()

    // 3. 使用Channel解决
    demonstrateChannelSolution()

    // 4. 使用原子操作解决
    demonstrateAtomicSolution()
}

// 典型的竞态条件 - 银行账户示例
func demonstrateRaceCondition() {
    fmt.Println("\n1. 竞态条件示例（银行账户）：")

    // 不安全的银行账户
    type UnsafeBankAccount struct {
        balance int64
    }

    account := &UnsafeBankAccount{balance: 1000}

    var wg sync.WaitGroup

    // 模拟多个并发交易
    transactions := []struct {
        amount int64
        desc   string
    }{
        {100, "存款"},
        {-50, "取款"},
        {200, "存款"},
        {-30, "取款"},
        {-80, "取款"},
    }

    for _, tx := range transactions {
        wg.Add(1)
        go func(amount int64, desc string) {
            defer wg.Done()

            // 读取当前余额
            currentBalance := account.balance

            // 模拟处理时间
            time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

            // 更新余额
            newBalance := currentBalance + amount
            account.balance = newBalance

            fmt.Printf("%s %d, 余额: %d -> %d\n", desc, amount, currentBalance, newBalance)
        }(tx.amount, tx.desc)
    }

    wg.Wait()

    fmt.Printf("最终余额: %d (可能不正确)\n", account.balance)
    fmt.Println("⚠️  由于竞态条件，结果可能不一致")
}

// 使用Mutex解决竞态条件
func demonstrateMutexSolution() {
    fmt.Println("\n2. 使用Mutex解决竞态条件：")

    // 安全的银行账户
    type SafeBankAccount struct {
        mu      sync.Mutex
        balance int64
    }

    // 存款方法
    deposit := func(account *SafeBankAccount, amount int64) {
        account.mu.Lock()
        defer account.mu.Unlock()

        oldBalance := account.balance
        account.balance += amount
        fmt.Printf("存款 %d, 余额: %d -> %d\n", amount, oldBalance, account.balance)
    }

    // 取款方法
    withdraw := func(account *SafeBankAccount, amount int64) bool {
        account.mu.Lock()
        defer account.mu.Unlock()

        if account.balance >= amount {
            oldBalance := account.balance
            account.balance -= amount
            fmt.Printf("取款 %d, 余额: %d -> %d\n", amount, oldBalance, account.balance)
            return true
        }

        fmt.Printf("取款 %d 失败, 余额不足: %d\n", amount, account.balance)
        return false
    }

    // 查询余额方法
    getBalance := func(account *SafeBankAccount) int64 {
        account.mu.Lock()
        defer account.mu.Unlock()
        return account.balance
    }

    account := &SafeBankAccount{balance: 1000}
    var wg sync.WaitGroup

    // 并发交易
    wg.Add(1)
    go func() {
        defer wg.Done()
        deposit(account, 100)
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        withdraw(account, 50)
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        deposit(account, 200)
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        withdraw(account, 300)
    }()

    wg.Wait()

    finalBalance := getBalance(account)
    fmt.Printf("最终余额: %d\n", finalBalance)
    fmt.Println("✅ Mutex确保了操作的原子性")
}

// 使用Channel解决竞态条件
func demonstrateChannelSolution() {
    fmt.Println("\n3. 使用Channel解决竞态条件：")

    type Transaction struct {
        amount   int64
        response chan bool // 用于返回操作结果
    }

    type BankAccountActor struct {
        balance     int64
        transactions chan Transaction
        done        chan struct{}
    }

    // 创建银行账户Actor
    newBankAccountActor := func(initialBalance int64) *BankAccountActor {
        actor := &BankAccountActor{
            balance:      initialBalance,
            transactions: make(chan Transaction),
            done:         make(chan struct{}),
        }

        // 启动处理Goroutine
        go func() {
            for {
                select {
                case tx := <-actor.transactions:
                    oldBalance := actor.balance

                    if tx.amount > 0 {
                        // 存款
                        actor.balance += tx.amount
                        fmt.Printf("存款 %d, 余额: %d -> %d\n", tx.amount, oldBalance, actor.balance)
                        tx.response <- true
                    } else {
                        // 取款
                        amount := -tx.amount
                        if actor.balance >= amount {
                            actor.balance -= amount
                            fmt.Printf("取款 %d, 余额: %d -> %d\n", amount, oldBalance, actor.balance)
                            tx.response <- true
                        } else {
                            fmt.Printf("取款 %d 失败, 余额不足: %d\n", amount, actor.balance)
                            tx.response <- false
                        }
                    }

                case <-actor.done:
                    return
                }
            }
        }()

        return actor
    }

    // 交易方法
    transact := func(actor *BankAccountActor, amount int64) bool {
        response := make(chan bool)
        tx := Transaction{
            amount:   amount,
            response: response,
        }

        actor.transactions <- tx
        return <-response
    }

    // 停止Actor
    stop := func(actor *BankAccountActor) {
        close(actor.done)
    }

    account := newBankAccountActor(1000)
    var wg sync.WaitGroup

    // 并发交易
    transactions := []int64{100, -50, 200, -300, -80}

    for _, amount := range transactions {
        wg.Add(1)
        go func(amt int64) {
            defer wg.Done()
            transact(account, amt)
        }(amount)
    }

    wg.Wait()

    // 查询最终余额
    balanceResponse := make(chan bool)
    balanceTx := Transaction{amount: 0, response: balanceResponse}
    account.transactions <- balanceTx
    <-balanceResponse

    stop(account)
    fmt.Println("✅ Channel确保了串行化处理")
}

// 使用原子操作解决竞态条件
func demonstrateAtomicSolution() {
    fmt.Println("\n4. 使用原子操作解决竞态条件：")

    // 简单计数器的原子操作
    var counter int64
    var wg sync.WaitGroup

    const numGoroutines = 10
    const incrementsPerGoroutine = 1000

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < incrementsPerGoroutine; j++ {
                // 原子递增
                atomic.AddInt64(&counter, 1)
            }
        }(i)
    }

    wg.Wait()

    expected := int64(numGoroutines * incrementsPerGoroutine)
    actual := atomic.LoadInt64(&counter)

    fmt.Printf("期望值: %d, 实际值: %d\n", expected, actual)

    if actual == expected {
        fmt.Println("✅ 原子操作确保了并发安全")
    }

    // 复杂数据结构的原子操作
    demonstrateAtomicValue()
}

func demonstrateAtomicValue() {
    fmt.Println("\n原子Value操作：")

    type Config struct {
        MaxConnections int
        Timeout        time.Duration
        EnableLogging  bool
    }

    var configValue atomic.Value

    // 初始配置
    initialConfig := &Config{
        MaxConnections: 100,
        Timeout:        5 * time.Second,
        EnableLogging:  true,
    }
    configValue.Store(initialConfig)

    var wg sync.WaitGroup

    // 配置读取者
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(readerID int) {
            defer wg.Done()

            for j := 0; j < 3; j++ {
                config := configValue.Load().(*Config)
                fmt.Printf("读取者%d: 最大连接=%d, 超时=%v, 日志=%t\n",
                    readerID, config.MaxConnections, config.Timeout, config.EnableLogging)
                time.Sleep(100 * time.Millisecond)
            }
        }(i)
    }

    // 配置更新者
    wg.Add(1)
    go func() {
        defer wg.Done()

        time.Sleep(150 * time.Millisecond)

        newConfig := &Config{
            MaxConnections: 200,
            Timeout:        3 * time.Second,
            EnableLogging:  false,
        }
        configValue.Store(newConfig)
        fmt.Println("🔄 配置已原子性更新")
    }()

    wg.Wait()
    fmt.Println("✅ 原子Value确保了配置的一致性读取")
}
```

---

## ⏰ Context包的使用和最佳实践

### Context基础概念

Context是Go语言中用于处理请求范围数据、取消信号和超时的标准包，是并发编程中不可或缺的工具。

#### 1. Context的基本用法

```go
// 来自 mall-go/pkg/context/context_demo.go
package context

import (
    "context"
    "fmt"
    "time"
)

// Context基础演示
func DemonstrateContextBasics() {
    fmt.Println("=== Context基础演示 ===")

    // 1. 基本Context创建
    demonstrateBasicContext()

    // 2. 带取消的Context
    demonstrateCancelContext()

    // 3. 带超时的Context
    demonstrateTimeoutContext()

    // 4. 带截止时间的Context
    demonstrateDeadlineContext()

    // 5. 带值的Context
    demonstrateValueContext()
}

// 基本Context创建
func demonstrateBasicContext() {
    fmt.Println("\n1. 基本Context创建：")

    // 创建根Context
    ctx := context.Background()
    fmt.Printf("根Context: %v\n", ctx)

    // 创建TODO Context（用于不确定使用哪种Context的场景）
    todoCtx := context.TODO()
    fmt.Printf("TODO Context: %v\n", todoCtx)

    // 检查Context状态
    select {
    case <-ctx.Done():
        fmt.Println("Context已取消")
    default:
        fmt.Println("Context正常运行")
    }
}

// 带取消的Context
func demonstrateCancelContext() {
    fmt.Println("\n2. 带取消的Context：")

    // 创建可取消的Context
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel() // 确保资源清理

    // 启动工作Goroutine
    go func() {
        for {
            select {
            case <-ctx.Done():
                fmt.Println("工作Goroutine收到取消信号，退出")
                fmt.Printf("取消原因: %v\n", ctx.Err())
                return
            default:
                fmt.Println("工作Goroutine正在运行...")
                time.Sleep(500 * time.Millisecond)
            }
        }
    }()

    // 运行2秒后取消
    time.Sleep(2 * time.Second)
    fmt.Println("发送取消信号")
    cancel()

    // 等待Goroutine退出
    time.Sleep(100 * time.Millisecond)
}

// 带超时的Context
func demonstrateTimeoutContext() {
    fmt.Println("\n3. 带超时的Context：")

    // 创建1秒超时的Context
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    // 模拟长时间运行的操作
    go func() {
        select {
        case <-time.After(2 * time.Second):
            fmt.Println("操作完成")
        case <-ctx.Done():
            fmt.Println("操作被取消")
            fmt.Printf("取消原因: %v\n", ctx.Err())
        }
    }()

    // 等待操作完成或超时
    <-ctx.Done()
    fmt.Printf("Context状态: %v\n", ctx.Err())
}

// 带截止时间的Context
func demonstrateDeadlineContext() {
    fmt.Println("\n4. 带截止时间的Context：")

    // 创建截止时间为1.5秒后的Context
    deadline := time.Now().Add(1500 * time.Millisecond)
    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()

    fmt.Printf("截止时间: %v\n", deadline.Format("15:04:05.000"))

    // 检查剩余时间
    if deadline, ok := ctx.Deadline(); ok {
        remaining := time.Until(deadline)
        fmt.Printf("剩余时间: %v\n", remaining)
    }

    // 等待截止时间到达
    <-ctx.Done()
    fmt.Printf("Context已过期: %v\n", ctx.Err())
}

// 带值的Context
func demonstrateValueContext() {
    fmt.Println("\n5. 带值的Context：")

    // 定义Context键类型（避免键冲突）
    type contextKey string

    const (
        userIDKey    contextKey = "user_id"
        requestIDKey contextKey = "request_id"
        traceIDKey   contextKey = "trace_id"
    )

    // 创建带值的Context
    ctx := context.Background()
    ctx = context.WithValue(ctx, userIDKey, "user123")
    ctx = context.WithValue(ctx, requestIDKey, "req456")
    ctx = context.WithValue(ctx, traceIDKey, "trace789")

    // 从Context中获取值
    if userID := ctx.Value(userIDKey); userID != nil {
        fmt.Printf("用户ID: %v\n", userID)
    }

    if requestID := ctx.Value(requestIDKey); requestID != nil {
        fmt.Printf("请求ID: %v\n", requestID)
    }

    if traceID := ctx.Value(traceIDKey); traceID != nil {
        fmt.Printf("追踪ID: %v\n", traceID)
    }

    // 获取不存在的值
    if sessionID := ctx.Value("session_id"); sessionID != nil {
        fmt.Printf("会话ID: %v\n", sessionID)
    } else {
        fmt.Println("会话ID不存在")
    }
}
```

#### 2. Context在HTTP服务中的应用

```go
// 来自 mall-go/internal/handler/context_handler.go
package handler

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Context在HTTP服务中的应用
type ContextHandler struct {
    logger *zap.Logger
}

func NewContextHandler(logger *zap.Logger) *ContextHandler {
    return &ContextHandler{logger: logger}
}

// 中间件：添加请求上下文信息
func (h *ContextHandler) ContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 生成请求ID
        requestID := generateRequestID()

        // 从Header获取用户ID
        userID := c.GetHeader("X-User-ID")
        if userID == "" {
            userID = "anonymous"
        }

        // 从Header获取追踪ID
        traceID := c.GetHeader("X-Trace-ID")
        if traceID == "" {
            traceID = generateTraceID()
        }

        // 创建带值的Context
        ctx := c.Request.Context()
        ctx = context.WithValue(ctx, "request_id", requestID)
        ctx = context.WithValue(ctx, "user_id", userID)
        ctx = context.WithValue(ctx, "trace_id", traceID)
        ctx = context.WithValue(ctx, "start_time", time.Now())

        // 设置超时
        ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()

        // 更新请求Context
        c.Request = c.Request.WithContext(ctx)

        // 记录请求开始
        h.logger.Info("请求开始",
            zap.String("request_id", requestID),
            zap.String("user_id", userID),
            zap.String("trace_id", traceID),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
        )

        c.Next()

        // 记录请求结束
        if startTime := ctx.Value("start_time"); startTime != nil {
            duration := time.Since(startTime.(time.Time))
            h.logger.Info("请求结束",
                zap.String("request_id", requestID),
                zap.Duration("duration", duration),
                zap.Int("status", c.Writer.Status()),
            )
        }
    }
}

// 处理长时间运行的任务
func (h *ContextHandler) LongRunningTask(c *gin.Context) {
    ctx := c.Request.Context()

    // 从Context获取信息
    requestID := getStringFromContext(ctx, "request_id")
    userID := getStringFromContext(ctx, "user_id")

    h.logger.Info("开始长时间任务",
        zap.String("request_id", requestID),
        zap.String("user_id", userID),
    )

    // 模拟长时间运行的任务
    result, err := h.performLongTask(ctx, userID)
    if err != nil {
        if err == context.DeadlineExceeded {
            c.JSON(http.StatusRequestTimeout, gin.H{
                "error": "请求超时",
                "request_id": requestID,
            })
            return
        }

        if err == context.Canceled {
            c.JSON(http.StatusRequestTimeout, gin.H{
                "error": "请求被取消",
                "request_id": requestID,
            })
            return
        }

        h.logger.Error("任务执行失败",
            zap.String("request_id", requestID),
            zap.Error(err),
        )

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "内部服务器错误",
            "request_id": requestID,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "result": result,
        "request_id": requestID,
    })
}

// 执行长时间任务
func (h *ContextHandler) performLongTask(ctx context.Context, userID string) (string, error) {
    // 创建子任务
    tasks := []func(context.Context) error{
        h.taskStep1,
        h.taskStep2,
        h.taskStep3,
    }

    for i, task := range tasks {
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
            h.logger.Debug("执行任务步骤",
                zap.Int("step", i+1),
                zap.String("user_id", userID),
            )

            if err := task(ctx); err != nil {
                return "", fmt.Errorf("步骤 %d 失败: %w", i+1, err)
            }
        }
    }

    return "任务完成", nil
}

func (h *ContextHandler) taskStep1(ctx context.Context) error {
    // 模拟步骤1：数据库查询
    select {
    case <-time.After(2 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (h *ContextHandler) taskStep2(ctx context.Context) error {
    // 模拟步骤2：外部API调用
    select {
    case <-time.After(3 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (h *ContextHandler) taskStep3(ctx context.Context) error {
    // 模拟步骤3：数据处理
    select {
    case <-time.After(1 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// 辅助函数
func getStringFromContext(ctx context.Context, key string) string {
    if value := ctx.Value(key); value != nil {
        if str, ok := value.(string); ok {
            return str
        }
    }
    return ""
}

func generateRequestID() string {
    return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func generateTraceID() string {
    return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}
```

#### 3. Context与Java/Python的对比

```java
// Java - CompletableFuture with timeout
public class JavaContextExample {

    public CompletableFuture<String> longRunningTask(String userId) {
        return CompletableFuture
            .supplyAsync(() -> {
                // 模拟长时间任务
                try {
                    Thread.sleep(5000);
                    return "任务完成: " + userId;
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("任务被中断", e);
                }
            })
            .orTimeout(3, TimeUnit.SECONDS) // 3秒超时
            .exceptionally(throwable -> {
                if (throwable instanceof TimeoutException) {
                    return "任务超时";
                }
                return "任务失败: " + throwable.getMessage();
            });
    }

    // 使用示例
    public void handleRequest(HttpServletRequest request) {
        String userId = request.getHeader("X-User-ID");

        longRunningTask(userId)
            .thenAccept(result -> {
                // 处理结果
                System.out.println("结果: " + result);
            })
            .join(); // 等待完成
    }
}

/*
Java vs Go Context对比：

1. 取消机制：
   - Java: CompletableFuture.cancel()，但不能传播到子任务
   - Go: Context取消会自动传播到所有子Context

2. 超时处理：
   - Java: orTimeout()方法，但需要手动处理
   - Go: WithTimeout()自动处理，统一的Done()通道

3. 值传递：
   - Java: 没有内置机制，通常使用ThreadLocal
   - Go: WithValue()提供类型安全的值传递

4. 组合性：
   - Java: 需要手动组合多个CompletableFuture
   - Go: Context自然支持嵌套和组合
*/
```

```python
# Python - asyncio with timeout and cancellation
import asyncio
import contextvars
from typing import Optional

class PythonContextExample:

    # 使用contextvars传递请求上下文
    request_id: contextvars.ContextVar[str] = contextvars.ContextVar('request_id')
    user_id: contextvars.ContextVar[str] = contextvars.ContextVar('user_id')

    async def long_running_task(self, user_id: str) -> str:
        """长时间运行的任务"""
        try:
            # 模拟任务步骤
            await asyncio.sleep(2)  # 步骤1
            await asyncio.sleep(3)  # 步骤2
            await asyncio.sleep(1)  # 步骤3

            return f"任务完成: {user_id}"

        except asyncio.CancelledError:
            print(f"任务被取消: {user_id}")
            raise

    async def handle_request_with_timeout(self, user_id: str) -> Optional[str]:
        """带超时的请求处理"""
        try:
            # 设置3秒超时
            result = await asyncio.wait_for(
                self.long_running_task(user_id),
                timeout=3.0
            )
            return result

        except asyncio.TimeoutError:
            print("请求超时")
            return None
        except asyncio.CancelledError:
            print("请求被取消")
            return None

    async def handle_request_with_cancellation(self, user_id: str):
        """带取消机制的请求处理"""
        # 创建任务
        task = asyncio.create_task(self.long_running_task(user_id))

        try:
            # 等待2秒后取消
            await asyncio.sleep(2)
            task.cancel()

            result = await task
            print(f"结果: {result}")

        except asyncio.CancelledError:
            print("任务已取消")

"""
Python vs Go Context对比：

1. 异步模型：
   - Python: 基于事件循环的协程
   - Go: 基于CSP模型的Goroutine

2. 取消机制：
   - Python: Task.cancel()，需要手动传播
   - Go: Context.Done()自动传播

3. 超时处理：
   - Python: asyncio.wait_for()
   - Go: context.WithTimeout()

4. 值传递：
   - Python: contextvars（Python 3.7+）
   - Go: context.WithValue()

5. 错误处理：
   - Python: 异常机制
   - Go: 错误返回值
"""
```

---

## 🏭 常见并发模式

### Worker Pool模式

Worker Pool是最常用的并发模式之一，通过固定数量的工作者处理任务队列。

```go
// 来自 mall-go/pkg/patterns/worker_pool.go
package patterns

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"
)

// Worker Pool模式实现
type WorkerPool struct {
    workerCount int
    taskQueue   chan Task
    resultQueue chan Result
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

type Task struct {
    ID       int
    Data     interface{}
    Priority int // 优先级：1-高，2-中，3-低
}

type Result struct {
    TaskID   int
    Output   interface{}
    Error    error
    Duration time.Duration
    WorkerID int
}

// 创建Worker Pool
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())

    return &WorkerPool{
        workerCount: workerCount,
        taskQueue:   make(chan Task, queueSize),
        resultQueue: make(chan Result, queueSize),
        ctx:         ctx,
        cancel:      cancel,
    }
}

// 启动Worker Pool
func (wp *WorkerPool) Start() {
    fmt.Printf("启动 %d 个Worker\n", wp.workerCount)

    for i := 0; i < wp.workerCount; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

// Worker实现
func (wp *WorkerPool) worker(workerID int) {
    defer wp.wg.Done()

    fmt.Printf("Worker %d 启动\n", workerID)

    for {
        select {
        case <-wp.ctx.Done():
            fmt.Printf("Worker %d 收到停止信号\n", workerID)
            return

        case task := <-wp.taskQueue:
            fmt.Printf("Worker %d 开始处理任务 %d\n", workerID, task.ID)

            start := time.Now()
            output, err := wp.processTask(task)
            duration := time.Since(start)

            result := Result{
                TaskID:   task.ID,
                Output:   output,
                Error:    err,
                Duration: duration,
                WorkerID: workerID,
            }

            // 发送结果
            select {
            case wp.resultQueue <- result:
                fmt.Printf("Worker %d 完成任务 %d，耗时 %v\n",
                    workerID, task.ID, duration)
            case <-wp.ctx.Done():
                return
            }
        }
    }
}

// 处理任务
func (wp *WorkerPool) processTask(task Task) (interface{}, error) {
    // 模拟不同优先级的处理时间
    var processingTime time.Duration
    switch task.Priority {
    case 1: // 高优先级
        processingTime = time.Duration(100+rand.Intn(200)) * time.Millisecond
    case 2: // 中优先级
        processingTime = time.Duration(200+rand.Intn(300)) * time.Millisecond
    case 3: // 低优先级
        processingTime = time.Duration(300+rand.Intn(500)) * time.Millisecond
    default:
        processingTime = time.Duration(200+rand.Intn(300)) * time.Millisecond
    }

    time.Sleep(processingTime)

    // 模拟10%的失败率
    if rand.Float32() < 0.1 {
        return nil, fmt.Errorf("任务 %d 处理失败", task.ID)
    }

    return fmt.Sprintf("任务 %d 的处理结果", task.ID), nil
}

// 提交任务
func (wp *WorkerPool) SubmitTask(task Task) error {
    select {
    case wp.taskQueue <- task:
        return nil
    case <-wp.ctx.Done():
        return fmt.Errorf("Worker Pool已停止")
    default:
        return fmt.Errorf("任务队列已满")
    }
}

// 获取结果
func (wp *WorkerPool) GetResult() <-chan Result {
    return wp.resultQueue
}

// 停止Worker Pool
func (wp *WorkerPool) Stop() {
    fmt.Println("停止Worker Pool...")
    wp.cancel()
    close(wp.taskQueue)
    wp.wg.Wait()
    close(wp.resultQueue)
    fmt.Println("Worker Pool已停止")
}

// Worker Pool使用示例
func DemonstrateWorkerPool() {
    fmt.Println("=== Worker Pool模式演示 ===")

    // 创建Worker Pool：3个Worker，队列大小10
    pool := NewWorkerPool(3, 10)
    pool.Start()

    // 启动结果收集器
    go func() {
        for result := range pool.GetResult() {
            if result.Error != nil {
                fmt.Printf("❌ 任务 %d 失败: %v (Worker %d, 耗时 %v)\n",
                    result.TaskID, result.Error, result.WorkerID, result.Duration)
            } else {
                fmt.Printf("✅ 任务 %d 成功: %v (Worker %d, 耗时 %v)\n",
                    result.TaskID, result.Output, result.WorkerID, result.Duration)
            }
        }
    }()

    // 提交任务
    for i := 1; i <= 15; i++ {
        task := Task{
            ID:       i,
            Data:     fmt.Sprintf("数据-%d", i),
            Priority: (i%3)+1, // 循环设置优先级
        }

        if err := pool.SubmitTask(task); err != nil {
            fmt.Printf("提交任务 %d 失败: %v\n", i, err)
        } else {
            fmt.Printf("提交任务 %d (优先级 %d)\n", i, task.Priority)
        }

        time.Sleep(100 * time.Millisecond)
    }

    // 等待所有任务完成
    time.Sleep(5 * time.Second)

    // 停止Worker Pool
    pool.Stop()
}
```

### Pipeline模式

Pipeline模式将复杂的处理过程分解为多个阶段，每个阶段专注于特定的处理逻辑。

```go
// 来自 mall-go/pkg/patterns/pipeline.go
package patterns

import (
    "context"
    "fmt"
    "strings"
    "time"
)

// Pipeline模式实现
type Pipeline struct {
    stages []Stage
    ctx    context.Context
    cancel context.CancelFunc
}

type Stage interface {
    Process(ctx context.Context, input <-chan interface{}) <-chan interface{}
    Name() string
}

// 数据清洗阶段
type DataCleaningStage struct{}

func (s *DataCleaningStage) Name() string {
    return "数据清洗"
}

func (s *DataCleaningStage) Process(ctx context.Context, input <-chan interface{}) <-chan interface{} {
    output := make(chan interface{})

    go func() {
        defer close(output)

        for {
            select {
            case <-ctx.Done():
                fmt.Printf("%s阶段收到取消信号\n", s.Name())
                return
            case data, ok := <-input:
                if !ok {
                    fmt.Printf("%s阶段输入完成\n", s.Name())
                    return
                }

                // 模拟数据清洗
                if str, ok := data.(string); ok {
                    cleaned := strings.TrimSpace(strings.ToLower(str))
                    fmt.Printf("%s: %s -> %s\n", s.Name(), str, cleaned)

                    select {
                    case output <- cleaned:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
    }()

    return output
}

// 数据验证阶段
type DataValidationStage struct{}

func (s *DataValidationStage) Name() string {
    return "数据验证"
}

func (s *DataValidationStage) Process(ctx context.Context, input <-chan interface{}) <-chan interface{} {
    output := make(chan interface{})

    go func() {
        defer close(output)

        for {
            select {
            case <-ctx.Done():
                fmt.Printf("%s阶段收到取消信号\n", s.Name())
                return
            case data, ok := <-input:
                if !ok {
                    fmt.Printf("%s阶段输入完成\n", s.Name())
                    return
                }

                // 模拟数据验证
                if str, ok := data.(string); ok {
                    if len(str) > 0 && !strings.Contains(str, "invalid") {
                        fmt.Printf("%s: %s ✅\n", s.Name(), str)

                        select {
                        case output <- str:
                        case <-ctx.Done():
                            return
                        }
                    } else {
                        fmt.Printf("%s: %s ❌ (无效数据)\n", s.Name(), str)
                    }
                }
            }
        }
    }()

    return output
}

// 数据转换阶段
type DataTransformationStage struct{}

func (s *DataTransformationStage) Name() string {
    return "数据转换"
}

func (s *DataTransformationStage) Process(ctx context.Context, input <-chan interface{}) <-chan interface{} {
    output := make(chan interface{})

    go func() {
        defer close(output)

        for {
            select {
            case <-ctx.Done():
                fmt.Printf("%s阶段收到取消信号\n", s.Name())
                return
            case data, ok := <-input:
                if !ok {
                    fmt.Printf("%s阶段输入完成\n", s.Name())
                    return
                }

                // 模拟数据转换
                if str, ok := data.(string); ok {
                    transformed := fmt.Sprintf("PROCESSED_%s_%d", strings.ToUpper(str), time.Now().Unix())
                    fmt.Printf("%s: %s -> %s\n", s.Name(), str, transformed)

                    select {
                    case output <- transformed:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
    }()

    return output
}

// 创建Pipeline
func NewPipeline(stages ...Stage) *Pipeline {
    ctx, cancel := context.WithCancel(context.Background())

    return &Pipeline{
        stages: stages,
        ctx:    ctx,
        cancel: cancel,
    }
}

// 执行Pipeline
func (p *Pipeline) Execute(input <-chan interface{}) <-chan interface{} {
    current := input

    // 依次通过每个阶段
    for _, stage := range p.stages {
        current = stage.Process(p.ctx, current)
    }

    return current
}

// 停止Pipeline
func (p *Pipeline) Stop() {
    p.cancel()
}

// Pipeline使用示例
func DemonstratePipeline() {
    fmt.Println("\n=== Pipeline模式演示 ===")

    // 创建Pipeline
    pipeline := NewPipeline(
        &DataCleaningStage{},
        &DataValidationStage{},
        &DataTransformationStage{},
    )

    // 创建输入数据
    input := make(chan interface{}, 10)

    // 发送测试数据
    go func() {
        defer close(input)

        testData := []string{
            "  Hello World  ",
            "Go Programming",
            "invalid data",
            "  CONCURRENT  ",
            "Pipeline Pattern",
            "",
            "Final Test",
        }

        for _, data := range testData {
            input <- data
            time.Sleep(200 * time.Millisecond)
        }
    }()

    // 执行Pipeline并收集结果
    output := pipeline.Execute(input)

    fmt.Println("\n=== Pipeline处理结果 ===")
    for result := range output {
        fmt.Printf("最终结果: %v\n", result)
    }

    pipeline.Stop()
}
```

### Fan-in/Fan-out模式

```go
// 来自 mall-go/pkg/patterns/fan_patterns.go
package patterns

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "time"
)

// Fan-out模式：将任务分发给多个处理器
func FanOutPattern() {
    fmt.Println("\n=== Fan-out模式演示 ===")

    // 创建输入Channel
    input := make(chan int, 20)

    // 生产数据
    go func() {
        defer close(input)
        for i := 1; i <= 20; i++ {
            input <- i
            fmt.Printf("生产数据: %d\n", i)
            time.Sleep(50 * time.Millisecond)
        }
    }()

    // Fan-out到多个处理器
    const numProcessors = 4
    var wg sync.WaitGroup

    for i := 0; i < numProcessors; i++ {
        wg.Add(1)
        go func(processorID int) {
            defer wg.Done()

            for data := range input {
                // 模拟处理时间
                processingTime := time.Duration(rand.Intn(300)) * time.Millisecond
                time.Sleep(processingTime)

                result := data * data
                fmt.Printf("处理器%d: %d -> %d (耗时 %v)\n",
                    processorID, data, result, processingTime)
            }

            fmt.Printf("处理器%d 完成\n", processorID)
        }(i)
    }

    wg.Wait()
    fmt.Println("Fan-out处理完成")
}

// Fan-in模式：将多个输入合并到一个输出
func FanInPattern() {
    fmt.Println("\n=== Fan-in模式演示 ===")

    // 创建多个输入源
    sources := make([]<-chan string, 3)

    for i := 0; i < 3; i++ {
        source := make(chan string)
        sources[i] = source

        // 启动数据生产者
        go func(sourceID int, ch chan<- string) {
            defer close(ch)

            for j := 1; j <= 5; j++ {
                message := fmt.Sprintf("源%d-消息%d", sourceID, j)
                ch <- message
                fmt.Printf("源%d 生产: %s\n", sourceID, message)

                // 不同源有不同的生产速度
                delay := time.Duration((sourceID+1)*100) * time.Millisecond
                time.Sleep(delay)
            }

            fmt.Printf("源%d 完成\n", sourceID)
        }(i, source)
    }

    // Fan-in合并所有输入
    merged := fanInMultiple(sources...)

    // 消费合并后的数据
    fmt.Println("\n=== Fan-in合并结果 ===")
    for message := range merged {
        fmt.Printf("合并输出: %s\n", message)
    }

    fmt.Println("Fan-in处理完成")
}

// 合并多个Channel
func fanInMultiple(inputs ...<-chan string) <-chan string {
    output := make(chan string)
    var wg sync.WaitGroup

    // 为每个输入Channel启动一个Goroutine
    for i, input := range inputs {
        wg.Add(1)
        go func(id int, ch <-chan string) {
            defer wg.Done()

            for message := range ch {
                output <- message
            }
        }(i, input)
    }

    // 等待所有输入完成后关闭输出
    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}

// 复合模式：Fan-out + Fan-in
func FanOutFanInPattern() {
    fmt.Println("\n=== Fan-out + Fan-in复合模式演示 ===")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // 第一阶段：生成数据
    numbers := generateNumbers(ctx, 1, 20)

    // 第二阶段：Fan-out处理
    const numWorkers = 3
    processedChannels := make([]<-chan int, numWorkers)

    for i := 0; i < numWorkers; i++ {
        processedChannels[i] = processNumbers(ctx, i, numbers)
    }

    // 第三阶段：Fan-in合并结果
    results := fanInNumbers(ctx, processedChannels...)

    // 第四阶段：收集最终结果
    fmt.Println("\n=== 最终处理结果 ===")
    var totalResults int
    for result := range results {
        fmt.Printf("最终结果: %d\n", result)
        totalResults++
    }

    fmt.Printf("总共处理了 %d 个结果\n", totalResults)
}

// 生成数字
func generateNumbers(ctx context.Context, start, end int) <-chan int {
    output := make(chan int)

    go func() {
        defer close(output)

        for i := start; i <= end; i++ {
            select {
            case output <- i:
                fmt.Printf("生成数字: %d\n", i)
                time.Sleep(100 * time.Millisecond)
            case <-ctx.Done():
                fmt.Println("数字生成被取消")
                return
            }
        }

        fmt.Println("数字生成完成")
    }()

    return output
}

// 处理数字
func processNumbers(ctx context.Context, workerID int, input <-chan int) <-chan int {
    output := make(chan int)

    go func() {
        defer close(output)
        defer fmt.Printf("Worker%d 完成\n", workerID)

        for {
            select {
            case num, ok := <-input:
                if !ok {
                    return
                }

                // 模拟处理时间
                processingTime := time.Duration(rand.Intn(200)) * time.Millisecond
                time.Sleep(processingTime)

                result := num * num
                fmt.Printf("Worker%d: %d -> %d\n", workerID, num, result)

                select {
                case output <- result:
                case <-ctx.Done():
                    return
                }

            case <-ctx.Done():
                fmt.Printf("Worker%d 被取消\n", workerID)
                return
            }
        }
    }()

    return output
}

// 合并数字Channel
func fanInNumbers(ctx context.Context, inputs ...<-chan int) <-chan int {
    output := make(chan int)
    var wg sync.WaitGroup

    for i, input := range inputs {
        wg.Add(1)
        go func(id int, ch <-chan int) {
            defer wg.Done()

            for {
                select {
                case num, ok := <-ch:
                    if !ok {
                        return
                    }

                    select {
                    case output <- num:
                    case <-ctx.Done():
                        return
                    }

                case <-ctx.Done():
                    return
                }
            }
        }(i, input)
    }

    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}
```

---

## 🔧 并发程序的调试和性能优化

### Race Detector使用

Go提供了内置的竞态条件检测工具，是调试并发程序的利器。

```go
// 来自 mall-go/pkg/debug/race_example.go
package debug

import (
    "fmt"
    "sync"
    "time"
)

// 竞态条件示例 - 用于演示race detector
func RaceConditionExample() {
    fmt.Println("=== 竞态条件示例 ===")

    // 有竞态条件的代码
    demonstrateRaceCondition()

    // 修复后的代码
    demonstrateFixedCode()
}

// 存在竞态条件的代码
func demonstrateRaceCondition() {
    fmt.Println("\n1. 存在竞态条件的代码：")

    var counter int
    var wg sync.WaitGroup

    // 启动多个Goroutine并发修改counter
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // 这里存在竞态条件！
            for j := 0; j < 1000; j++ {
                counter++ // 非原子操作
            }

            fmt.Printf("Goroutine %d 完成\n", id)
        }(i)
    }

    wg.Wait()
    fmt.Printf("最终计数: %d (期望: 10000)\n", counter)

    /*
    使用race detector检测：
    go run -race race_example.go

    输出会显示：
    ==================
    WARNING: DATA RACE
    Write at 0x00c000014088 by goroutine 7:
      main.demonstrateRaceCondition.func1()
          /path/to/race_example.go:XX +0x4e

    Previous write at 0x00c000014088 by goroutine 6:
      main.demonstrateRaceCondition.func1()
          /path/to/race_example.go:XX +0x4e
    ==================
    */
}

// 修复竞态条件的代码
func demonstrateFixedCode() {
    fmt.Println("\n2. 修复后的代码：")

    var counter int
    var mu sync.Mutex
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < 1000; j++ {
                mu.Lock()
                counter++
                mu.Unlock()
            }

            fmt.Printf("Goroutine %d 完成\n", id)
        }(i)
    }

    wg.Wait()
    fmt.Printf("最终计数: %d (期望: 10000)\n", counter)
}

// 更复杂的竞态条件示例
type UnsafeCounter struct {
    value int
    name  string
}

func (c *UnsafeCounter) Increment() {
    // 竞态条件：读取-修改-写入不是原子操作
    temp := c.value
    time.Sleep(time.Nanosecond) // 增加竞态条件发生的概率
    c.value = temp + 1
}

func (c *UnsafeCounter) Get() int {
    return c.value
}

func (c *UnsafeCounter) SetName(name string) {
    c.name = name
}

func (c *UnsafeCounter) GetName() string {
    return c.name
}

// 安全的计数器
type SafeCounter struct {
    mu    sync.RWMutex
    value int
    name  string
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *SafeCounter) Get() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.value
}

func (c *SafeCounter) SetName(name string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.name = name
}

func (c *SafeCounter) GetName() string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.name
}

// 复杂竞态条件演示
func ComplexRaceConditionExample() {
    fmt.Println("\n=== 复杂竞态条件示例 ===")

    // 不安全的计数器
    unsafeCounter := &UnsafeCounter{name: "unsafe"}

    var wg sync.WaitGroup

    // 并发递增
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < 100; j++ {
                unsafeCounter.Increment()

                // 并发修改名称
                if j%10 == 0 {
                    unsafeCounter.SetName(fmt.Sprintf("counter-%d-%d", id, j))
                }
            }
        }(i)
    }

    // 并发读取
    wg.Add(1)
    go func() {
        defer wg.Done()

        for i := 0; i < 50; i++ {
            value := unsafeCounter.Get()
            name := unsafeCounter.GetName()
            fmt.Printf("读取: 值=%d, 名称=%s\n", value, name)
            time.Sleep(10 * time.Millisecond)
        }
    }()

    wg.Wait()

    fmt.Printf("不安全计数器最终值: %d\n", unsafeCounter.Get())

    // 安全的计数器对比
    fmt.Println("\n安全计数器对比：")
    safeCounter := &SafeCounter{name: "safe"}

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < 100; j++ {
                safeCounter.Increment()

                if j%10 == 0 {
                    safeCounter.SetName(fmt.Sprintf("counter-%d-%d", id, j))
                }
            }
        }(i)
    }

    wg.Wait()
    fmt.Printf("安全计数器最终值: %d\n", safeCounter.Get())
}
```

### 性能分析工具

```go
// 来自 mall-go/pkg/debug/profiling.go
package debug

import (
    "context"
    "fmt"
    "math/rand"
    "runtime"
    "sync"
    "time"
)

// 性能分析示例
func PerformanceAnalysisExample() {
    fmt.Println("=== 性能分析示例 ===")

    // 1. CPU密集型任务
    demonstrateCPUIntensiveTask()

    // 2. 内存密集型任务
    demonstrateMemoryIntensiveTask()

    // 3. Goroutine泄漏检测
    demonstrateGoroutineLeakDetection()
}

// CPU密集型任务
func demonstrateCPUIntensiveTask() {
    fmt.Println("\n1. CPU密集型任务：")

    start := time.Now()

    // 并发计算斐波那契数列
    const numWorkers = 4
    const numTasks = 20

    taskChan := make(chan int, numTasks)
    resultChan := make(chan int, numTasks)

    var wg sync.WaitGroup

    // 启动Worker
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()

            for n := range taskChan {
                result := fibonacci(n)
                fmt.Printf("Worker %d: fibonacci(%d) = %d\n", workerID, n, result)
                resultChan <- result
            }
        }(i)
    }

    // 发送任务
    go func() {
        defer close(taskChan)
        for i := 30; i < 30+numTasks; i++ {
            taskChan <- i
        }
    }()

    // 收集结果
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    var results []int
    for result := range resultChan {
        results = append(results, result)
    }

    duration := time.Since(start)
    fmt.Printf("CPU密集型任务完成，耗时: %v，结果数量: %d\n", duration, len(results))

    /*
    使用pprof分析CPU性能：

    1. 在代码中添加：
    import _ "net/http/pprof"
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    2. 运行程序后访问：
    go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

    3. 在pprof交互模式中使用：
    (pprof) top10    # 显示CPU使用最多的10个函数
    (pprof) list fibonacci  # 显示fibonacci函数的详细信息
    (pprof) web      # 生成调用图
    */
}

// 斐波那契数列（CPU密集型）
func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

// 内存密集型任务
func demonstrateMemoryIntensiveTask() {
    fmt.Println("\n2. 内存密集型任务：")

    start := time.Now()

    // 创建大量数据结构
    const numSlices = 100
    const sliceSize = 100000

    var slices [][]int
    var wg sync.WaitGroup

    for i := 0; i < numSlices; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // 创建大切片
            slice := make([]int, sliceSize)
            for j := 0; j < sliceSize; j++ {
                slice[j] = rand.Intn(1000)
            }

            // 模拟处理
            sum := 0
            for _, v := range slice {
                sum += v
            }

            fmt.Printf("切片 %d 处理完成，和: %d\n", id, sum)

            // 注意：这里故意不释放slice，模拟内存泄漏
            slices = append(slices, slice)
        }(i)
    }

    wg.Wait()

    duration := time.Since(start)
    fmt.Printf("内存密集型任务完成，耗时: %v，切片数量: %d\n", duration, len(slices))

    // 显示内存统计
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("内存使用: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d\n",
        bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)

    /*
    使用pprof分析内存：

    1. 堆内存分析：
    go tool pprof http://localhost:6060/debug/pprof/heap

    2. 内存分配分析：
    go tool pprof http://localhost:6060/debug/pprof/allocs

    3. 在pprof中使用：
    (pprof) top10 -cum    # 按累计分配排序
    (pprof) list demonstrateMemoryIntensiveTask
    (pprof) web
    */
}

func bToKb(b uint64) uint64 {
    return b / 1024
}

// Goroutine泄漏检测
func demonstrateGoroutineLeakDetection() {
    fmt.Println("\n3. Goroutine泄漏检测：")

    fmt.Printf("开始时Goroutine数量: %d\n", runtime.NumGoroutine())

    // 创建会泄漏的Goroutine
    createLeakyGoroutines()

    time.Sleep(100 * time.Millisecond)
    fmt.Printf("创建泄漏Goroutine后数量: %d\n", runtime.NumGoroutine())

    // 创建正常的Goroutine
    createNormalGoroutines()

    time.Sleep(2 * time.Second)
    fmt.Printf("正常Goroutine完成后数量: %d\n", runtime.NumGoroutine())

    /*
    Goroutine泄漏检测方法：

    1. 使用pprof查看Goroutine：
    go tool pprof http://localhost:6060/debug/pprof/goroutine

    2. 在pprof中使用：
    (pprof) top10
    (pprof) list createLeakyGoroutines
    (pprof) traces  # 显示Goroutine的调用栈

    3. 使用go-leak检测库：
    import "go.uber.org/goleak"

    func TestMain(m *testing.M) {
        goleak.VerifyTestMain(m)
    }
    */
}

// 创建会泄漏的Goroutine
func createLeakyGoroutines() {
    for i := 0; i < 10; i++ {
        go func(id int) {
            // 这个Goroutine永远不会退出，造成泄漏
            ch := make(chan struct{})
            <-ch // 永远阻塞
        }(i)
    }
}

// 创建正常的Goroutine
func createNormalGoroutines() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // 正常的工作，会正常退出
            time.Sleep(1 * time.Second)
            fmt.Printf("正常Goroutine %d 完成\n", id)
        }(i)
    }

    wg.Wait()
}

// 性能基准测试示例
func BenchmarkConcurrentVsSequential() {
    fmt.Println("\n=== 并发vs顺序性能对比 ===")

    const dataSize = 1000000
    data := make([]int, dataSize)
    for i := range data {
        data[i] = rand.Intn(1000)
    }

    // 顺序处理
    start := time.Now()
    sequentialSum := sequentialSum(data)
    sequentialTime := time.Since(start)

    // 并发处理
    start = time.Now()
    concurrentSum := concurrentSum(data, 4)
    concurrentTime := time.Since(start)

    fmt.Printf("数据大小: %d\n", dataSize)
    fmt.Printf("顺序处理: 结果=%d, 耗时=%v\n", sequentialSum, sequentialTime)
    fmt.Printf("并发处理: 结果=%d, 耗时=%v\n", concurrentSum, concurrentTime)
    fmt.Printf("性能提升: %.2fx\n", float64(sequentialTime)/float64(concurrentTime))

    /*
    运行基准测试：
    go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

    分析结果：
    go tool pprof cpu.prof
    go tool pprof mem.prof
    */
}

// 顺序求和
func sequentialSum(data []int) int {
    sum := 0
    for _, v := range data {
        sum += v
    }
    return sum
}

// 并发求和
func concurrentSum(data []int, numWorkers int) int {
    chunkSize := len(data) / numWorkers
    resultChan := make(chan int, numWorkers)

    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(start, end int) {
            defer wg.Done()

            sum := 0
            for j := start; j < end; j++ {
                sum += data[j]
            }
            resultChan <- sum
        }(i*chunkSize, (i+1)*chunkSize)
    }

    go func() {
        wg.Wait()
        close(resultChan)
    }()

    totalSum := 0
    for partialSum := range resultChan {
        totalSum += partialSum
    }

    return totalSum
}
```

---

## 🏢 实战案例分析

### Mall-Go项目并发实现

让我们通过一个完整的电商系统案例，看看Go并发编程在实际项目中的应用。

```go
// 来自 mall-go/internal/service/order_service.go
package service

import (
    "context"
    "fmt"
    "sync"
    "time"

    "go.uber.org/zap"
    "gorm.io/gorm"
)

// 订单服务 - 展示复杂的并发处理场景
type OrderService struct {
    db           *gorm.DB
    logger       *zap.Logger
    stockService *StockService
    payService   *PaymentService
    notifyService *NotificationService

    // 并发控制
    orderProcessSemaphore chan struct{} // 限制并发订单处理数量
    ctx                   context.Context
    cancel                context.CancelFunc
    wg                    sync.WaitGroup
}

type Order struct {
    ID          int64     `json:"id"`
    UserID      int64     `json:"user_id"`
    ProductID   int64     `json:"product_id"`
    Quantity    int       `json:"quantity"`
    TotalAmount float64   `json:"total_amount"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type OrderProcessResult struct {
    OrderID   int64
    Success   bool
    Error     error
    Duration  time.Duration
    Steps     []ProcessStep
}

type ProcessStep struct {
    Name      string
    Success   bool
    Duration  time.Duration
    Error     error
}

func NewOrderService(db *gorm.DB, logger *zap.Logger) *OrderService {
    ctx, cancel := context.WithCancel(context.Background())

    return &OrderService{
        db:                    db,
        logger:               logger,
        stockService:         NewStockService(db, logger),
        payService:           NewPaymentService(logger),
        notifyService:        NewNotificationService(logger),
        orderProcessSemaphore: make(chan struct{}, 10), // 最多10个并发订单处理
        ctx:                   ctx,
        cancel:               cancel,
    }
}

// 批量处理订单 - 展示Fan-out模式
func (os *OrderService) ProcessOrdersBatch(orders []Order) []OrderProcessResult {
    os.logger.Info("开始批量处理订单", zap.Int("count", len(orders)))

    resultChan := make(chan OrderProcessResult, len(orders))

    // Fan-out: 并发处理每个订单
    for _, order := range orders {
        os.wg.Add(1)
        go os.processOrderConcurrently(order, resultChan)
    }

    // 等待所有订单处理完成
    go func() {
        os.wg.Wait()
        close(resultChan)
    }()

    // 收集结果
    var results []OrderProcessResult
    for result := range resultChan {
        results = append(results, result)
    }

    os.logger.Info("批量订单处理完成",
        zap.Int("total", len(results)),
        zap.Int("success", os.countSuccessfulOrders(results)),
        zap.Int("failed", len(results)-os.countSuccessfulOrders(results)),
    )

    return results
}

// 并发处理单个订单
func (os *OrderService) processOrderConcurrently(order Order, resultChan chan<- OrderProcessResult) {
    defer os.wg.Done()

    // 获取信号量，限制并发数
    os.orderProcessSemaphore <- struct{}{}
    defer func() { <-os.orderProcessSemaphore }()

    start := time.Now()
    result := OrderProcessResult{
        OrderID: order.ID,
        Steps:   make([]ProcessStep, 0),
    }

    os.logger.Info("开始处理订单", zap.Int64("order_id", order.ID))

    // 创建订单处理的Context，设置超时
    ctx, cancel := context.WithTimeout(os.ctx, 30*time.Second)
    defer cancel()

    // 步骤1: 库存检查和扣减
    if success, duration, err := os.processStockReduction(ctx, order); !success {
        result.Steps = append(result.Steps, ProcessStep{
            Name: "库存扣减", Success: false, Duration: duration, Error: err,
        })
        result.Success = false
        result.Error = err
        result.Duration = time.Since(start)
        resultChan <- result
        return
    } else {
        result.Steps = append(result.Steps, ProcessStep{
            Name: "库存扣减", Success: true, Duration: duration,
        })
    }

    // 步骤2: 支付处理
    if success, duration, err := os.processPayment(ctx, order); !success {
        // 支付失败，需要回滚库存
        os.rollbackStock(ctx, order)

        result.Steps = append(result.Steps, ProcessStep{
            Name: "支付处理", Success: false, Duration: duration, Error: err,
        })
        result.Success = false
        result.Error = err
        result.Duration = time.Since(start)
        resultChan <- result
        return
    } else {
        result.Steps = append(result.Steps, ProcessStep{
            Name: "支付处理", Success: true, Duration: duration,
        })
    }

    // 步骤3: 更新订单状态
    if success, duration, err := os.updateOrderStatus(ctx, order.ID, "paid"); !success {
        // 状态更新失败，需要回滚支付和库存
        os.rollbackPayment(ctx, order)
        os.rollbackStock(ctx, order)

        result.Steps = append(result.Steps, ProcessStep{
            Name: "状态更新", Success: false, Duration: duration, Error: err,
        })
        result.Success = false
        result.Error = err
        result.Duration = time.Since(start)
        resultChan <- result
        return
    } else {
        result.Steps = append(result.Steps, ProcessStep{
            Name: "状态更新", Success: true, Duration: duration,
        })
    }

    // 步骤4: 异步发送通知（不影响主流程）
    go os.sendOrderNotifications(order)

    result.Success = true
    result.Duration = time.Since(start)

    os.logger.Info("订单处理成功",
        zap.Int64("order_id", order.ID),
        zap.Duration("duration", result.Duration),
    )

    resultChan <- result
}

// 库存扣减处理
func (os *OrderService) processStockReduction(ctx context.Context, order Order) (bool, time.Duration, error) {
    start := time.Now()

    // 使用Context控制超时
    stockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    err := os.stockService.ReduceStock(stockCtx, order.ProductID, order.Quantity)
    duration := time.Since(start)

    if err != nil {
        os.logger.Error("库存扣减失败",
            zap.Int64("order_id", order.ID),
            zap.Int64("product_id", order.ProductID),
            zap.Int("quantity", order.Quantity),
            zap.Error(err),
        )
        return false, duration, err
    }

    return true, duration, nil
}

// 支付处理
func (os *OrderService) processPayment(ctx context.Context, order Order) (bool, time.Duration, error) {
    start := time.Now()

    payCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    err := os.payService.ProcessPayment(payCtx, order.ID, order.TotalAmount)
    duration := time.Since(start)

    if err != nil {
        os.logger.Error("支付处理失败",
            zap.Int64("order_id", order.ID),
            zap.Float64("amount", order.TotalAmount),
            zap.Error(err),
        )
        return false, duration, err
    }

    return true, duration, nil
}

// 更新订单状态
func (os *OrderService) updateOrderStatus(ctx context.Context, orderID int64, status string) (bool, time.Duration, error) {
    start := time.Now()

    err := os.db.WithContext(ctx).Model(&Order{}).
        Where("id = ?", orderID).
        Update("status", status).Error

    duration := time.Since(start)

    if err != nil {
        os.logger.Error("订单状态更新失败",
            zap.Int64("order_id", orderID),
            zap.String("status", status),
            zap.Error(err),
        )
        return false, duration, err
    }

    return true, duration, nil
}

// 发送订单通知（异步）
func (os *OrderService) sendOrderNotifications(order Order) {
    os.logger.Info("发送订单通知", zap.Int64("order_id", order.ID))

    // 并发发送多种通知
    var wg sync.WaitGroup

    // 发送邮件通知
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := os.notifyService.SendEmail(
            fmt.Sprintf("user%d@example.com", order.UserID),
            "订单确认",
            fmt.Sprintf("您的订单 %d 已确认", order.ID),
        ); err != nil {
            os.logger.Error("邮件通知发送失败", zap.Error(err))
        }
    }()

    // 发送短信通知
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := os.notifyService.SendSMS(
            fmt.Sprintf("1380000%04d", order.UserID),
            fmt.Sprintf("订单 %d 已确认，感谢您的购买", order.ID),
        ); err != nil {
            os.logger.Error("短信通知发送失败", zap.Error(err))
        }
    }()

    // 发送推送通知
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := os.notifyService.SendPush(
            fmt.Sprintf("device_%d", order.UserID),
            "订单确认",
            fmt.Sprintf("订单 %d 已确认", order.ID),
        ); err != nil {
            os.logger.Error("推送通知发送失败", zap.Error(err))
        }
    }()

    wg.Wait()
    os.logger.Info("订单通知发送完成", zap.Int64("order_id", order.ID))
}

// 回滚操作
func (os *OrderService) rollbackStock(ctx context.Context, order Order) {
    if err := os.stockService.RestoreStock(ctx, order.ProductID, order.Quantity); err != nil {
        os.logger.Error("库存回滚失败", zap.Int64("order_id", order.ID), zap.Error(err))
    }
}

func (os *OrderService) rollbackPayment(ctx context.Context, order Order) {
    if err := os.payService.RefundPayment(ctx, order.ID, order.TotalAmount); err != nil {
        os.logger.Error("支付回滚失败", zap.Int64("order_id", order.ID), zap.Error(err))
    }
}

// 统计成功订单数量
func (os *OrderService) countSuccessfulOrders(results []OrderProcessResult) int {
    count := 0
    for _, result := range results {
        if result.Success {
            count++
        }
    }
    return count
}

// 停止服务
func (os *OrderService) Stop() {
    os.logger.Info("停止订单服务")
    os.cancel()
    os.wg.Wait()
    os.logger.Info("订单服务已停止")
}
```

### 库存服务实现

```go
// 来自 mall-go/internal/service/stock_service.go
package service

import (
    "context"
    "fmt"
    "sync"
    "time"

    "go.uber.org/zap"
    "gorm.io/gorm"
)

// 库存服务 - 展示高并发下的数据一致性处理
type StockService struct {
    db     *gorm.DB
    logger *zap.Logger

    // 使用分段锁减少锁竞争
    stockLocks map[int64]*sync.RWMutex
    locksMu    sync.RWMutex
}

type Stock struct {
    ProductID int64 `json:"product_id" gorm:"primaryKey"`
    Quantity  int   `json:"quantity"`
    Reserved  int   `json:"reserved"` // 预留库存
    UpdatedAt time.Time `json:"updated_at"`
}

func NewStockService(db *gorm.DB, logger *zap.Logger) *StockService {
    return &StockService{
        db:         db,
        logger:     logger,
        stockLocks: make(map[int64]*sync.RWMutex),
    }
}

// 获取产品锁（分段锁实现）
func (ss *StockService) getProductLock(productID int64) *sync.RWMutex {
    ss.locksMu.RLock()
    lock, exists := ss.stockLocks[productID]
    ss.locksMu.RUnlock()

    if exists {
        return lock
    }

    // 需要创建新锁
    ss.locksMu.Lock()
    defer ss.locksMu.Unlock()

    // 双重检查
    if lock, exists := ss.stockLocks[productID]; exists {
        return lock
    }

    lock = &sync.RWMutex{}
    ss.stockLocks[productID] = lock
    return lock
}

// 扣减库存 - 使用乐观锁和悲观锁结合
func (ss *StockService) ReduceStock(ctx context.Context, productID int64, quantity int) error {
    ss.logger.Info("开始扣减库存",
        zap.Int64("product_id", productID),
        zap.Int("quantity", quantity),
    )

    // 获取产品锁
    lock := ss.getProductLock(productID)
    lock.Lock()
    defer lock.Unlock()

    // 使用事务确保一致性
    return ss.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var stock Stock

        // 悲观锁查询当前库存
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            Where("product_id = ?", productID).
            First(&stock).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                return fmt.Errorf("产品 %d 库存不存在", productID)
            }
            return fmt.Errorf("查询库存失败: %w", err)
        }

        // 检查库存是否充足
        availableStock := stock.Quantity - stock.Reserved
        if availableStock < quantity {
            return fmt.Errorf("库存不足: 可用=%d, 需要=%d", availableStock, quantity)
        }

        // 扣减库存
        if err := tx.Model(&stock).
            Where("product_id = ?", productID).
            Update("quantity", gorm.Expr("quantity - ?", quantity)).Error; err != nil {
            return fmt.Errorf("扣减库存失败: %w", err)
        }

        ss.logger.Info("库存扣减成功",
            zap.Int64("product_id", productID),
            zap.Int("quantity", quantity),
            zap.Int("remaining", stock.Quantity-quantity),
        )

        return nil
    })
}

// 恢复库存
func (ss *StockService) RestoreStock(ctx context.Context, productID int64, quantity int) error {
    ss.logger.Info("开始恢复库存",
        zap.Int64("product_id", productID),
        zap.Int("quantity", quantity),
    )

    lock := ss.getProductLock(productID)
    lock.Lock()
    defer lock.Unlock()

    return ss.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&Stock{}).
            Where("product_id = ?", productID).
            Update("quantity", gorm.Expr("quantity + ?", quantity)).Error; err != nil {
            return fmt.Errorf("恢复库存失败: %w", err)
        }

        ss.logger.Info("库存恢复成功",
            zap.Int64("product_id", productID),
            zap.Int("quantity", quantity),
        )

        return nil
    })
}

// 预留库存
func (ss *StockService) ReserveStock(ctx context.Context, productID int64, quantity int) error {
    lock := ss.getProductLock(productID)
    lock.Lock()
    defer lock.Unlock()

    return ss.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var stock Stock

        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            Where("product_id = ?", productID).
            First(&stock).Error; err != nil {
            return fmt.Errorf("查询库存失败: %w", err)
        }

        availableStock := stock.Quantity - stock.Reserved
        if availableStock < quantity {
            return fmt.Errorf("可用库存不足")
        }

        if err := tx.Model(&stock).
            Where("product_id = ?", productID).
            Update("reserved", gorm.Expr("reserved + ?", quantity)).Error; err != nil {
            return fmt.Errorf("预留库存失败: %w", err)
        }

        return nil
    })
}

// 批量库存操作 - 展示并发批处理
func (ss *StockService) BatchReduceStock(ctx context.Context, operations []StockOperation) []StockResult {
    ss.logger.Info("开始批量库存操作", zap.Int("count", len(operations)))

    resultChan := make(chan StockResult, len(operations))
    semaphore := make(chan struct{}, 5) // 限制并发数

    var wg sync.WaitGroup

    for _, op := range operations {
        wg.Add(1)
        go func(operation StockOperation) {
            defer wg.Done()

            // 获取信号量
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            start := time.Now()
            err := ss.ReduceStock(ctx, operation.ProductID, operation.Quantity)
            duration := time.Since(start)

            result := StockResult{
                ProductID: operation.ProductID,
                Quantity:  operation.Quantity,
                Success:   err == nil,
                Error:     err,
                Duration:  duration,
            }

            resultChan <- result
        }(op)
    }

    go func() {
        wg.Wait()
        close(resultChan)
    }()

    var results []StockResult
    for result := range resultChan {
        results = append(results, result)
    }

    successCount := 0
    for _, result := range results {
        if result.Success {
            successCount++
        }
    }

    ss.logger.Info("批量库存操作完成",
        zap.Int("total", len(results)),
        zap.Int("success", successCount),
        zap.Int("failed", len(results)-successCount),
    )

    return results
}

type StockOperation struct {
    ProductID int64
    Quantity  int
}

type StockResult struct {
    ProductID int64
    Quantity  int
    Success   bool
    Error     error
    Duration  time.Duration
}
```

---

## 🎯 面试常考点

### Go并发编程面试题精选

```go
// 来自 mall-go/docs/interview/concurrency_questions.go
package interview

import (
    "context"
    "fmt"
    "sync"
    "time"
)

/*
=== Go并发编程面试常考点 ===

1. Goroutine vs 线程的区别
2. Channel的内部实现原理
3. Select语句的执行机制
4. 内存模型和happens-before关系
5. 常见的并发模式
6. 竞态条件和数据竞争
7. Context的使用场景
8. 性能优化技巧
*/

// 面试题1: 实现一个并发安全的计数器
type SafeCounter struct {
    mu    sync.Mutex
    value int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

/*
面试官可能的追问：
1. 为什么读操作也需要加锁？
   答：因为在没有同步机制的情况下，读写操作之间可能存在数据竞争。
   即使是简单的int读取，在某些架构上也可能不是原子的。

2. 如何优化这个计数器的性能？
   答：可以使用原子操作 sync/atomic 包，或者使用读写锁 sync.RWMutex。
*/

// 面试题2: 实现一个带超时的Channel操作
func ChannelWithTimeout() {
    ch := make(chan string, 1)

    // 发送操作带超时
    select {
    case ch <- "hello":
        fmt.Println("发送成功")
    case <-time.After(1 * time.Second):
        fmt.Println("发送超时")
    }

    // 接收操作带超时
    select {
    case msg := <-ch:
        fmt.Printf("接收到: %s\n", msg)
    case <-time.After(1 * time.Second):
        fmt.Println("接收超时")
    }
}

/*
面试官可能的追问：
1. time.After()会造成内存泄漏吗？
   答：在Go 1.23之前，如果select的其他case先执行，time.After创建的Timer
   不会被立即回收，可能造成临时的内存泄漏。建议使用context.WithTimeout。

2. 如何实现可取消的超时？
   答：使用context.WithTimeout或context.WithCancel。
*/

// 面试题3: 实现一个工作池
type WorkerPool struct {
    workerCount int
    jobQueue    chan Job
    quit        chan bool
    wg          sync.WaitGroup
}

type Job struct {
    ID   int
    Data interface{}
}

func NewWorkerPool(workerCount int) *WorkerPool {
    return &WorkerPool{
        workerCount: workerCount,
        jobQueue:    make(chan Job, 100),
        quit:        make(chan bool),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workerCount; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()

    for {
        select {
        case job := <-wp.jobQueue:
            fmt.Printf("Worker %d processing job %d\n", id, job.ID)
            // 处理任务
            time.Sleep(100 * time.Millisecond)

        case <-wp.quit:
            fmt.Printf("Worker %d stopping\n", id)
            return
        }
    }
}

func (wp *WorkerPool) Submit(job Job) {
    wp.jobQueue <- job
}

func (wp *WorkerPool) Stop() {
    close(wp.quit)
    wp.wg.Wait()
}

/*
面试官可能的追问：
1. 如何优雅地关闭工作池？
   答：先关闭jobQueue，让worker处理完剩余任务后再退出。

2. 如何处理任务执行失败的情况？
   答：可以添加重试机制、错误处理回调、或者死信队列。

3. 如何动态调整worker数量？
   答：可以实现动态扩缩容机制，根据队列长度调整worker数量。
*/

// 面试题4: 实现一个并发安全的LRU缓存
type LRUCache struct {
    capacity int
    cache    map[int]*Node
    head     *Node
    tail     *Node
    mu       sync.RWMutex
}

type Node struct {
    key   int
    value int
    prev  *Node
    next  *Node
}

func NewLRUCache(capacity int) *LRUCache {
    head := &Node{}
    tail := &Node{}
    head.next = tail
    tail.prev = head

    return &LRUCache{
        capacity: capacity,
        cache:    make(map[int]*Node),
        head:     head,
        tail:     tail,
    }
}

func (lru *LRUCache) Get(key int) int {
    lru.mu.Lock()
    defer lru.mu.Unlock()

    if node, exists := lru.cache[key]; exists {
        lru.moveToHead(node)
        return node.value
    }

    return -1
}

func (lru *LRUCache) Put(key, value int) {
    lru.mu.Lock()
    defer lru.mu.Unlock()

    if node, exists := lru.cache[key]; exists {
        node.value = value
        lru.moveToHead(node)
        return
    }

    newNode := &Node{key: key, value: value}

    if len(lru.cache) >= lru.capacity {
        // 删除尾部节点
        tail := lru.removeTail()
        delete(lru.cache, tail.key)
    }

    lru.cache[key] = newNode
    lru.addToHead(newNode)
}

func (lru *LRUCache) addToHead(node *Node) {
    node.prev = lru.head
    node.next = lru.head.next
    lru.head.next.prev = node
    lru.head.next = node
}

func (lru *LRUCache) removeNode(node *Node) {
    node.prev.next = node.next
    node.next.prev = node.prev
}

func (lru *LRUCache) moveToHead(node *Node) {
    lru.removeNode(node)
    lru.addToHead(node)
}

func (lru *LRUCache) removeTail() *Node {
    lastNode := lru.tail.prev
    lru.removeNode(lastNode)
    return lastNode
}

/*
面试官可能的追问：
1. 为什么使用读写锁而不是普通互斥锁？
   答：Get操作只读取数据，使用读锁可以允许多个并发读取，提高性能。

2. 如何进一步优化性能？
   答：可以使用分段锁、无锁数据结构、或者为每个操作使用单独的锁。

3. 如何处理缓存穿透和缓存雪崩？
   答：可以添加布隆过滤器、设置随机过期时间、使用熔断器等。
*/

// 面试题5: 实现一个发布订阅系统
type PubSub struct {
    mu          sync.RWMutex
    subscribers map[string][]chan interface{}
    closed      bool
}

func NewPubSub() *PubSub {
    return &PubSub{
        subscribers: make(map[string][]chan interface{}),
    }
}

func (ps *PubSub) Subscribe(topic string) <-chan interface{} {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.closed {
        return nil
    }

    ch := make(chan interface{}, 10)
    ps.subscribers[topic] = append(ps.subscribers[topic], ch)
    return ch
}

func (ps *PubSub) Publish(topic string, data interface{}) {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    if ps.closed {
        return
    }

    for _, ch := range ps.subscribers[topic] {
        select {
        case ch <- data:
        default:
            // 订阅者处理太慢，丢弃消息
        }
    }
}

func (ps *PubSub) Close() {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.closed {
        return
    }

    ps.closed = true

    for _, subscribers := range ps.subscribers {
        for _, ch := range subscribers {
            close(ch)
        }
    }

    ps.subscribers = nil
}

/*
面试官可能的追问：
1. 如何处理慢消费者问题？
   答：可以使用带缓冲的channel、丢弃策略、或者背压机制。

2. 如何实现取消订阅？
   答：可以返回一个取消函数，或者使用context进行管理。

3. 如何保证消息的可靠传递？
   答：可以添加确认机制、重试机制、或者持久化存储。
*/
```

---

## ⚠️ 踩坑提醒

### 常见并发编程陷阱

```go
// 来自 mall-go/docs/pitfalls/concurrency_pitfalls.go
package pitfalls

import (
    "context"
    "fmt"
    "sync"
    "time"
)

/*
=== Go并发编程常见陷阱 ===

1. Goroutine泄漏
2. Channel死锁
3. 竞态条件
4. 内存逃逸
5. Context误用
6. 锁的误用
7. 性能陷阱
*/

// 陷阱1: Goroutine泄漏
func GoroutineLeakExample() {
    fmt.Println("=== Goroutine泄漏示例 ===")

    // ❌ 错误示例：Goroutine永远不会退出
    badExample := func() {
        ch := make(chan int)

        go func() {
            // 这个Goroutine会永远阻塞，造成泄漏
            <-ch
        }()

        // 函数返回，但Goroutine仍在运行
        fmt.Println("函数返回，但Goroutine泄漏了")
    }

    badExample()

    // ✅ 正确示例：使用Context控制Goroutine生命周期
    goodExample := func() {
        ctx, cancel := context.WithCancel(context.Background())
        defer cancel() // 确保清理

        ch := make(chan int)

        go func() {
            select {
            case <-ch:
                fmt.Println("收到数据")
            case <-ctx.Done():
                fmt.Println("Goroutine正常退出")
                return
            }
        }()

        // 函数返回前取消Context
        fmt.Println("函数返回，Goroutine会正常退出")
    }

    goodExample()
    time.Sleep(100 * time.Millisecond)
}

// 陷阱2: Channel死锁
func ChannelDeadlockExample() {
    fmt.Println("\n=== Channel死锁示例 ===")

    // ❌ 错误示例1：向无缓冲Channel发送数据但没有接收者
    deadlockExample1 := func() {
        defer func() {
            if r := recover(); r != nil {
                fmt.Printf("捕获死锁: %v\n", r)
            }
        }()

        ch := make(chan int)
        ch <- 1 // 死锁！没有接收者
    }

    go deadlockExample1()
    time.Sleep(100 * time.Millisecond)

    // ❌ 错误示例2：循环依赖导致死锁
    deadlockExample2 := func() {
        ch1 := make(chan int)
        ch2 := make(chan int)

        go func() {
            ch1 <- 1
            <-ch2 // 等待ch2
        }()

        go func() {
            ch2 <- 2
            <-ch1 // 等待ch1
        }()

        time.Sleep(100 * time.Millisecond)
        fmt.Println("可能的死锁情况")
    }

    deadlockExample2()

    // ✅ 正确示例：使用select避免死锁
    goodExample := func() {
        ch1 := make(chan int)
        ch2 := make(chan int)

        go func() {
            select {
            case ch1 <- 1:
                fmt.Println("发送到ch1成功")
            case <-time.After(100 * time.Millisecond):
                fmt.Println("发送到ch1超时")
            }
        }()

        go func() {
            select {
            case val := <-ch1:
                fmt.Printf("从ch1接收: %d\n", val)
            case <-time.After(200 * time.Millisecond):
                fmt.Println("从ch1接收超时")
            }
        }()

        time.Sleep(300 * time.Millisecond)
    }

    goodExample()
}

// 陷阱3: 竞态条件
func RaceConditionPitfall() {
    fmt.Println("\n=== 竞态条件陷阱 ===")

    // ❌ 错误示例：map并发读写
    badMapExample := func() {
        m := make(map[int]int)

        var wg sync.WaitGroup

        // 并发写入
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(key int) {
                defer wg.Done()
                m[key] = key * key // 竞态条件！
            }(i)
        }

        // 并发读取
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < 10; i++ {
                _ = m[i] // 竞态条件！
            }
        }()

        wg.Wait()
        fmt.Println("map操作完成（可能崩溃）")
    }

    // 使用defer recover避免程序崩溃
    func() {
        defer func() {
            if r := recover(); r != nil {
                fmt.Printf("捕获panic: %v\n", r)
            }
        }()
        badMapExample()
    }()

    // ✅ 正确示例：使用sync.Map或加锁
    goodMapExample := func() {
        var m sync.Map
        var wg sync.WaitGroup

        // 并发写入
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func(key int) {
                defer wg.Done()
                m.Store(key, key*key)
            }(i)
        }

        // 并发读取
        wg.Add(1)
        go func() {
            defer wg.Done()
            for i := 0; i < 10; i++ {
                if val, ok := m.Load(i); ok {
                    fmt.Printf("读取: %d = %v\n", i, val)
                }
            }
        }()

        wg.Wait()
        fmt.Println("sync.Map操作完成")
    }

    goodMapExample()
}

// 陷阱4: 闭包变量捕获
func ClosureVariablePitfall() {
    fmt.Println("\n=== 闭包变量捕获陷阱 ===")

    // ❌ 错误示例：循环变量被闭包捕获
    badExample := func() {
        var wg sync.WaitGroup

        for i := 0; i < 5; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                fmt.Printf("错误示例 - i的值: %d\n", i) // 总是打印5
            }()
        }

        wg.Wait()
    }

    badExample()

    // ✅ 正确示例1：传递参数
    goodExample1 := func() {
        var wg sync.WaitGroup

        for i := 0; i < 5; i++ {
            wg.Add(1)
            go func(val int) {
                defer wg.Done()
                fmt.Printf("正确示例1 - 值: %d\n", val)
            }(i)
        }

        wg.Wait()
    }

    goodExample1()

    // ✅ 正确示例2：创建局部变量
    goodExample2 := func() {
        var wg sync.WaitGroup

        for i := 0; i < 5; i++ {
            wg.Add(1)
            i := i // 创建局部变量
            go func() {
                defer wg.Done()
                fmt.Printf("正确示例2 - 值: %d\n", i)
            }()
        }

        wg.Wait()
    }

    goodExample2()
}

// 陷阱5: Context误用
func ContextMisusePitfall() {
    fmt.Println("\n=== Context误用陷阱 ===")

    // ❌ 错误示例：Context存储业务数据
    badExample := func() {
        ctx := context.Background()
        ctx = context.WithValue(ctx, "user_data", map[string]interface{}{
            "id":       123,
            "name":     "张三",
            "password": "secret123", // 不应该存储敏感数据
            "config":   make([]byte, 1024*1024), // 不应该存储大量数据
        })

        processRequest(ctx)
    }

    processRequest := func(ctx context.Context) {
        if userData := ctx.Value("user_data"); userData != nil {
            fmt.Println("❌ 从Context获取业务数据（不推荐）")
        }
    }

    badExample()

    // ✅ 正确示例：Context只存储请求范围的元数据
    goodExample := func() {
        ctx := context.Background()
        ctx = context.WithValue(ctx, "request_id", "req_123")
        ctx = context.WithValue(ctx, "trace_id", "trace_456")
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        defer cancel()

        processRequestCorrectly(ctx, UserData{
            ID:   123,
            Name: "张三",
        })
    }

    type UserData struct {
        ID   int
        Name string
    }

    processRequestCorrectly := func(ctx context.Context, userData UserData) {
        requestID := ctx.Value("request_id")
        fmt.Printf("✅ 请求ID: %v, 用户: %s\n", requestID, userData.Name)
    }

    goodExample()
}

// 陷阱6: 锁的误用
func LockMisusePitfall() {
    fmt.Println("\n=== 锁误用陷阱 ===")

    // ❌ 错误示例：锁粒度过大
    type BadCounter struct {
        mu     sync.Mutex
        values map[string]int
    }

    func (c *BadCounter) IncrementAll() {
        c.mu.Lock()
        defer c.mu.Unlock()

        // 锁住整个map进行批量操作，粒度过大
        for key := range c.values {
            c.values[key]++
            time.Sleep(time.Millisecond) // 模拟耗时操作
        }
    }

    // ✅ 正确示例：细粒度锁
    type GoodCounter struct {
        locks  map[string]*sync.Mutex
        values map[string]int
        mu     sync.RWMutex // 保护locks和values的结构
    }

    func (c *GoodCounter) Increment(key string) {
        // 获取或创建锁
        c.mu.RLock()
        lock, exists := c.locks[key]
        c.mu.RUnlock()

        if !exists {
            c.mu.Lock()
            if lock, exists = c.locks[key]; !exists {
                lock = &sync.Mutex{}
                c.locks[key] = lock
                c.values[key] = 0
            }
            c.mu.Unlock()
        }

        // 只锁定特定的key
        lock.Lock()
        c.values[key]++
        lock.Unlock()
    }

    fmt.Println("锁粒度优化示例完成")
}

// 陷阱7: Channel使用陷阱
func ChannelUsagePitfall() {
    fmt.Println("\n=== Channel使用陷阱 ===")

    // ❌ 错误示例：忘记关闭Channel
    badExample := func() {
        ch := make(chan int, 10)

        go func() {
            for i := 0; i < 5; i++ {
                ch <- i
            }
            // 忘记关闭Channel
        }()

        // range会一直等待，直到Channel关闭
        go func() {
            for val := range ch {
                fmt.Printf("接收: %d\n", val)
                if val == 4 {
                    break // 手动跳出，但Channel仍未关闭
                }
            }
        }()

        time.Sleep(100 * time.Millisecond)
    }

    badExample()

    // ✅ 正确示例：正确关闭Channel
    goodExample := func() {
        ch := make(chan int, 10)

        go func() {
            defer close(ch) // 确保关闭Channel
            for i := 0; i < 5; i++ {
                ch <- i
            }
        }()

        for val := range ch {
            fmt.Printf("正确接收: %d\n", val)
        }
    }

    goodExample()
}
```

---

## 📝 练习题

### 基础练习题

#### 练习题1：并发计算器（基础）

**题目描述：**
实现一个并发安全的计算器，支持加法、减法、乘法、除法操作，要求：
1. 支持多个Goroutine同时进行计算
2. 保证计算结果的准确性
3. 提供获取当前结果的方法

```go
// 练习题1参考答案
package exercises

import (
    "fmt"
    "sync"
)

type ConcurrentCalculator struct {
    mu     sync.RWMutex
    result float64
}

func NewConcurrentCalculator() *ConcurrentCalculator {
    return &ConcurrentCalculator{}
}

func (c *ConcurrentCalculator) Add(value float64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.result += value
}

func (c *ConcurrentCalculator) Subtract(value float64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.result -= value
}

func (c *ConcurrentCalculator) Multiply(value float64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.result *= value
}

func (c *ConcurrentCalculator) Divide(value float64) error {
    if value == 0 {
        return fmt.Errorf("除数不能为零")
    }

    c.mu.Lock()
    defer c.mu.Unlock()
    c.result /= value
    return nil
}

func (c *ConcurrentCalculator) GetResult() float64 {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.result
}

func (c *ConcurrentCalculator) Reset() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.result = 0
}

// 测试函数
func TestConcurrentCalculator() {
    calc := NewConcurrentCalculator()
    var wg sync.WaitGroup

    // 并发执行100次加法操作
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(val float64) {
            defer wg.Done()
            calc.Add(val)
        }(1.0)
    }

    wg.Wait()

    result := calc.GetResult()
    fmt.Printf("并发加法结果: %.2f (期望: 100.00)\n", result)

    if result == 100.0 {
        fmt.Println("✅ 练习题1通过")
    } else {
        fmt.Println("❌ 练习题1失败")
    }
}
```

#### 练习题2：生产者消费者模式（中级）

**题目描述：**
实现一个生产者消费者系统，要求：
1. 支持多个生产者和消费者
2. 使用Channel进行通信
3. 支持优雅关闭
4. 统计生产和消费的数量

```go
// 练习题2参考答案
package exercises

import (
    "context"
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"
)

type ProducerConsumerSystem struct {
    taskQueue     chan Task
    resultQueue   chan Result
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup

    // 统计信息
    producedCount int64
    consumedCount int64
}

type Task struct {
    ID   int
    Data string
}

type Result struct {
    TaskID int
    Output string
    Error  error
}

func NewProducerConsumerSystem(queueSize int) *ProducerConsumerSystem {
    ctx, cancel := context.WithCancel(context.Background())

    return &ProducerConsumerSystem{
        taskQueue:   make(chan Task, queueSize),
        resultQueue: make(chan Result, queueSize),
        ctx:         ctx,
        cancel:      cancel,
    }
}

// 启动生产者
func (pcs *ProducerConsumerSystem) StartProducers(count int) {
    for i := 0; i < count; i++ {
        pcs.wg.Add(1)
        go pcs.producer(i)
    }
}

// 启动消费者
func (pcs *ProducerConsumerSystem) StartConsumers(count int) {
    for i := 0; i < count; i++ {
        pcs.wg.Add(1)
        go pcs.consumer(i)
    }
}

// 生产者
func (pcs *ProducerConsumerSystem) producer(id int) {
    defer pcs.wg.Done()

    for {
        select {
        case <-pcs.ctx.Done():
            fmt.Printf("生产者%d 停止\n", id)
            return
        default:
            task := Task{
                ID:   int(atomic.AddInt64(&pcs.producedCount, 1)),
                Data: fmt.Sprintf("来自生产者%d的数据", id),
            }

            select {
            case pcs.taskQueue <- task:
                fmt.Printf("生产者%d 生产任务%d\n", id, task.ID)
            case <-pcs.ctx.Done():
                return
            }

            // 模拟生产时间
            time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
        }
    }
}

// 消费者
func (pcs *ProducerConsumerSystem) consumer(id int) {
    defer pcs.wg.Done()

    for {
        select {
        case <-pcs.ctx.Done():
            fmt.Printf("消费者%d 停止\n", id)
            return
        case task := <-pcs.taskQueue:
            // 处理任务
            time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

            result := Result{
                TaskID: task.ID,
                Output: fmt.Sprintf("消费者%d处理了任务%d", id, task.ID),
            }

            atomic.AddInt64(&pcs.consumedCount, 1)

            select {
            case pcs.resultQueue <- result:
                fmt.Printf("消费者%d 完成任务%d\n", id, task.ID)
            case <-pcs.ctx.Done():
                return
            }
        }
    }
}

// 获取结果
func (pcs *ProducerConsumerSystem) GetResults() <-chan Result {
    return pcs.resultQueue
}

// 获取统计信息
func (pcs *ProducerConsumerSystem) GetStats() (int64, int64) {
    return atomic.LoadInt64(&pcs.producedCount), atomic.LoadInt64(&pcs.consumedCount)
}

// 停止系统
func (pcs *ProducerConsumerSystem) Stop() {
    pcs.cancel()
    pcs.wg.Wait()
    close(pcs.taskQueue)
    close(pcs.resultQueue)
}

// 测试函数
func TestProducerConsumerSystem() {
    system := NewProducerConsumerSystem(10)

    // 启动2个生产者和3个消费者
    system.StartProducers(2)
    system.StartConsumers(3)

    // 运行5秒
    time.Sleep(5 * time.Second)

    // 停止系统
    system.Stop()

    produced, consumed := system.GetStats()
    fmt.Printf("生产任务数: %d, 消费任务数: %d\n", produced, consumed)

    if produced > 0 && consumed > 0 {
        fmt.Println("✅ 练习题2通过")
    } else {
        fmt.Println("❌ 练习题2失败")
    }
}
```

#### 练习题3：并发Web爬虫（中级）

**题目描述：**
实现一个并发Web爬虫，要求：
1. 支持并发爬取多个URL
2. 限制并发数量
3. 支持超时控制
4. 去重处理

```go
// 练习题3参考答案
package exercises

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type WebCrawler struct {
    client      *http.Client
    semaphore   chan struct{}
    visited     map[string]bool
    visitedMu   sync.RWMutex
    results     []CrawlResult
    resultsMu   sync.Mutex
}

type CrawlResult struct {
    URL        string
    StatusCode int
    Error      error
    Duration   time.Duration
}

func NewWebCrawler(maxConcurrency int, timeout time.Duration) *WebCrawler {
    return &WebCrawler{
        client: &http.Client{
            Timeout: timeout,
        },
        semaphore: make(chan struct{}, maxConcurrency),
        visited:   make(map[string]bool),
    }
}

// 爬取单个URL
func (wc *WebCrawler) crawlURL(ctx context.Context, url string) CrawlResult {
    start := time.Now()

    // 检查是否已访问
    wc.visitedMu.RLock()
    if wc.visited[url] {
        wc.visitedMu.RUnlock()
        return CrawlResult{
            URL:      url,
            Error:    fmt.Errorf("URL已访问"),
            Duration: time.Since(start),
        }
    }
    wc.visitedMu.RUnlock()

    // 标记为已访问
    wc.visitedMu.Lock()
    wc.visited[url] = true
    wc.visitedMu.Unlock()

    // 获取信号量
    select {
    case wc.semaphore <- struct{}{}:
        defer func() { <-wc.semaphore }()
    case <-ctx.Done():
        return CrawlResult{
            URL:      url,
            Error:    ctx.Err(),
            Duration: time.Since(start),
        }
    }

    // 创建请求
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return CrawlResult{
            URL:      url,
            Error:    err,
            Duration: time.Since(start),
        }
    }

    // 发送请求
    resp, err := wc.client.Do(req)
    if err != nil {
        return CrawlResult{
            URL:      url,
            Error:    err,
            Duration: time.Since(start),
        }
    }
    defer resp.Body.Close()

    return CrawlResult{
        URL:        url,
        StatusCode: resp.StatusCode,
        Duration:   time.Since(start),
    }
}

// 批量爬取URL
func (wc *WebCrawler) CrawlURLs(ctx context.Context, urls []string) []CrawlResult {
    var wg sync.WaitGroup
    resultChan := make(chan CrawlResult, len(urls))

    // 启动爬虫Goroutine
    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            result := wc.crawlURL(ctx, u)
            resultChan <- result
        }(url)
    }

    // 等待所有爬虫完成
    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // 收集结果
    var results []CrawlResult
    for result := range resultChan {
        results = append(results, result)
    }

    wc.resultsMu.Lock()
    wc.results = append(wc.results, results...)
    wc.resultsMu.Unlock()

    return results
}

// 获取统计信息
func (wc *WebCrawler) GetStats() (int, int, int) {
    wc.resultsMu.Lock()
    defer wc.resultsMu.Unlock()

    total := len(wc.results)
    success := 0
    failed := 0

    for _, result := range wc.results {
        if result.Error == nil && result.StatusCode == 200 {
            success++
        } else {
            failed++
        }
    }

    return total, success, failed
}

// 测试函数
func TestWebCrawler() {
    crawler := NewWebCrawler(3, 5*time.Second)

    urls := []string{
        "https://httpbin.org/delay/1",
        "https://httpbin.org/status/200",
        "https://httpbin.org/status/404",
        "https://httpbin.org/delay/2",
        "https://httpbin.org/status/500",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    fmt.Println("开始爬取URL...")
    results := crawler.CrawlURLs(ctx, urls)

    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("❌ %s: %v (耗时: %v)\n", result.URL, result.Error, result.Duration)
        } else {
            fmt.Printf("✅ %s: %d (耗时: %v)\n", result.URL, result.StatusCode, result.Duration)
        }
    }

    total, success, failed := crawler.GetStats()
    fmt.Printf("总计: %d, 成功: %d, 失败: %d\n", total, success, failed)

    if total > 0 {
        fmt.Println("✅ 练习题3通过")
    } else {
        fmt.Println("❌ 练习题3失败")
    }
}
```

#### 练习题4：分布式任务调度器（高级）

**题目描述：**
实现一个分布式任务调度器，要求：
1. 支持任务优先级
2. 支持任务重试机制
3. 支持任务超时处理
4. 支持动态添加/移除Worker

```go
// 练习题4参考答案
package exercises

import (
    "context"
    "container/heap"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// 任务调度器
type TaskScheduler struct {
    taskQueue    *PriorityQueue
    workers      map[int]*Worker
    workersMu    sync.RWMutex
    nextWorkerID int64

    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup

    // 统计信息
    totalTasks     int64
    completedTasks int64
    failedTasks    int64
}

// 任务定义
type ScheduledTask struct {
    ID          int64
    Priority    int
    MaxRetries  int
    CurrentTry  int
    Timeout     time.Duration
    Payload     interface{}
    CreatedAt   time.Time
    ExecuteAt   time.Time

    // 任务执行函数
    Execute func(ctx context.Context, payload interface{}) error
}

// 优先级队列实现
type PriorityQueue struct {
    mu    sync.Mutex
    tasks []*ScheduledTask
}

func NewPriorityQueue() *PriorityQueue {
    pq := &PriorityQueue{
        tasks: make([]*ScheduledTask, 0),
    }
    heap.Init(pq)
    return pq
}

func (pq *PriorityQueue) Len() int {
    return len(pq.tasks)
}

func (pq *PriorityQueue) Less(i, j int) bool {
    // 优先级高的排在前面，时间早的排在前面
    if pq.tasks[i].Priority != pq.tasks[j].Priority {
        return pq.tasks[i].Priority > pq.tasks[j].Priority
    }
    return pq.tasks[i].ExecuteAt.Before(pq.tasks[j].ExecuteAt)
}

func (pq *PriorityQueue) Swap(i, j int) {
    pq.tasks[i], pq.tasks[j] = pq.tasks[j], pq.tasks[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
    pq.tasks = append(pq.tasks, x.(*ScheduledTask))
}

func (pq *PriorityQueue) Pop() interface{} {
    old := pq.tasks
    n := len(old)
    task := old[n-1]
    pq.tasks = old[0 : n-1]
    return task
}

func (pq *PriorityQueue) PushTask(task *ScheduledTask) {
    pq.mu.Lock()
    defer pq.mu.Unlock()
    heap.Push(pq, task)
}

func (pq *PriorityQueue) PopTask() *ScheduledTask {
    pq.mu.Lock()
    defer pq.mu.Unlock()

    if pq.Len() == 0 {
        return nil
    }

    return heap.Pop(pq).(*ScheduledTask)
}

// Worker定义
type Worker struct {
    ID        int64
    scheduler *TaskScheduler
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewTaskScheduler() *TaskScheduler {
    ctx, cancel := context.WithCancel(context.Background())

    return &TaskScheduler{
        taskQueue: NewPriorityQueue(),
        workers:   make(map[int]*Worker),
        ctx:       ctx,
        cancel:    cancel,
    }
}

// 添加任务
func (ts *TaskScheduler) ScheduleTask(task *ScheduledTask) {
    task.CreatedAt = time.Now()
    if task.ExecuteAt.IsZero() {
        task.ExecuteAt = time.Now()
    }

    ts.taskQueue.PushTask(task)
    atomic.AddInt64(&ts.totalTasks, 1)

    fmt.Printf("任务 %d 已调度，优先级: %d\n", task.ID, task.Priority)
}

// 添加Worker
func (ts *TaskScheduler) AddWorker() int64 {
    workerID := atomic.AddInt64(&ts.nextWorkerID, 1)

    ctx, cancel := context.WithCancel(ts.ctx)
    worker := &Worker{
        ID:        workerID,
        scheduler: ts,
        ctx:       ctx,
        cancel:    cancel,
    }

    ts.workersMu.Lock()
    ts.workers[int(workerID)] = worker
    ts.workersMu.Unlock()

    ts.wg.Add(1)
    go ts.runWorker(worker)

    fmt.Printf("Worker %d 已添加\n", workerID)
    return workerID
}

// 移除Worker
func (ts *TaskScheduler) RemoveWorker(workerID int64) {
    ts.workersMu.Lock()
    if worker, exists := ts.workers[int(workerID)]; exists {
        worker.cancel()
        delete(ts.workers, int(workerID))
        fmt.Printf("Worker %d 已移除\n", workerID)
    }
    ts.workersMu.Unlock()
}

// Worker运行逻辑
func (ts *TaskScheduler) runWorker(worker *Worker) {
    defer ts.wg.Done()

    fmt.Printf("Worker %d 开始运行\n", worker.ID)

    for {
        select {
        case <-worker.ctx.Done():
            fmt.Printf("Worker %d 停止\n", worker.ID)
            return
        default:
            task := ts.taskQueue.PopTask()
            if task == nil {
                time.Sleep(100 * time.Millisecond)
                continue
            }

            // 检查任务是否到执行时间
            if time.Now().Before(task.ExecuteAt) {
                // 重新放回队列
                ts.taskQueue.PushTask(task)
                time.Sleep(50 * time.Millisecond)
                continue
            }

            ts.executeTask(worker, task)
        }
    }
}

// 执行任务
func (ts *TaskScheduler) executeTask(worker *Worker, task *ScheduledTask) {
    fmt.Printf("Worker %d 开始执行任务 %d (第 %d 次尝试)\n",
        worker.ID, task.ID, task.CurrentTry+1)

    task.CurrentTry++

    // 创建任务执行上下文
    ctx, cancel := context.WithTimeout(worker.ctx, task.Timeout)
    defer cancel()

    // 执行任务
    err := task.Execute(ctx, task.Payload)

    if err != nil {
        fmt.Printf("Worker %d 任务 %d 执行失败: %v\n", worker.ID, task.ID, err)

        // 检查是否需要重试
        if task.CurrentTry < task.MaxRetries {
            // 延迟重试
            task.ExecuteAt = time.Now().Add(time.Duration(task.CurrentTry) * time.Second)
            ts.taskQueue.PushTask(task)
            fmt.Printf("任务 %d 将在 %v 后重试\n", task.ID, task.ExecuteAt)
        } else {
            atomic.AddInt64(&ts.failedTasks, 1)
            fmt.Printf("任务 %d 最终失败，已达到最大重试次数\n", task.ID)
        }
    } else {
        atomic.AddInt64(&ts.completedTasks, 1)
        fmt.Printf("Worker %d 任务 %d 执行成功\n", worker.ID, task.ID)
    }
}

// 获取统计信息
func (ts *TaskScheduler) GetStats() (int64, int64, int64, int) {
    ts.workersMu.RLock()
    workerCount := len(ts.workers)
    ts.workersMu.RUnlock()

    return atomic.LoadInt64(&ts.totalTasks),
           atomic.LoadInt64(&ts.completedTasks),
           atomic.LoadInt64(&ts.failedTasks),
           workerCount
}

// 停止调度器
func (ts *TaskScheduler) Stop() {
    fmt.Println("停止任务调度器...")
    ts.cancel()
    ts.wg.Wait()
    fmt.Println("任务调度器已停止")
}

// 测试函数
func TestTaskScheduler() {
    scheduler := NewTaskScheduler()

    // 添加3个Worker
    scheduler.AddWorker()
    scheduler.AddWorker()
    scheduler.AddWorker()

    // 创建测试任务
    for i := 1; i <= 10; i++ {
        task := &ScheduledTask{
            ID:         int64(i),
            Priority:   i % 3,  // 0, 1, 2 优先级
            MaxRetries: 2,
            Timeout:    2 * time.Second,
            Payload:    fmt.Sprintf("任务数据-%d", i),
            Execute: func(ctx context.Context, payload interface{}) error {
                // 模拟任务执行
                select {
                case <-time.After(time.Duration(100+i*50) * time.Millisecond):
                    // 模拟30%的失败率
                    if i%3 == 0 {
                        return fmt.Errorf("模拟任务失败")
                    }
                    return nil
                case <-ctx.Done():
                    return ctx.Err()
                }
            },
        }

        scheduler.ScheduleTask(task)
    }

    // 运行5秒
    time.Sleep(5 * time.Second)

    // 动态移除一个Worker
    scheduler.RemoveWorker(1)

    // 再运行3秒
    time.Sleep(3 * time.Second)

    // 获取统计信息
    total, completed, failed, workers := scheduler.GetStats()
    fmt.Printf("统计信息 - 总任务: %d, 完成: %d, 失败: %d, Worker数: %d\n",
        total, completed, failed, workers)

    scheduler.Stop()

    if completed > 0 {
        fmt.Println("✅ 练习题4通过")
    } else {
        fmt.Println("❌ 练习题4失败")
    }
}
```

#### 练习题5：实时数据流处理器（高级）

**题目描述：**
实现一个实时数据流处理器，要求：
1. 支持多种数据源输入
2. 支持数据转换和过滤
3. 支持窗口聚合操作
4. 支持背压控制

```go
// 练习题5参考答案
package exercises

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// 数据流处理器
type StreamProcessor struct {
    sources     []DataSource
    transforms  []Transform
    sinks       []DataSink

    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup

    // 背压控制
    backpressure chan struct{}

    // 统计信息
    processedCount int64
    droppedCount   int64
}

// 数据项
type DataItem struct {
    ID        string
    Timestamp time.Time
    Value     interface{}
    Metadata  map[string]interface{}
}

// 数据源接口
type DataSource interface {
    Start(ctx context.Context) <-chan DataItem
    Stop()
}

// 数据转换接口
type Transform interface {
    Process(ctx context.Context, input <-chan DataItem) <-chan DataItem
}

// 数据输出接口
type DataSink interface {
    Write(ctx context.Context, input <-chan DataItem)
}

// 模拟数据源
type MockDataSource struct {
    name     string
    interval time.Duration
    count    int64
}

func NewMockDataSource(name string, interval time.Duration) *MockDataSource {
    return &MockDataSource{
        name:     name,
        interval: interval,
    }
}

func (mds *MockDataSource) Start(ctx context.Context) <-chan DataItem {
    output := make(chan DataItem, 100)

    go func() {
        defer close(output)
        ticker := time.NewTicker(mds.interval)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                id := atomic.AddInt64(&mds.count, 1)
                item := DataItem{
                    ID:        fmt.Sprintf("%s-%d", mds.name, id),
                    Timestamp: time.Now(),
                    Value:     id * 10,
                    Metadata: map[string]interface{}{
                        "source": mds.name,
                    },
                }

                select {
                case output <- item:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    return output
}

func (mds *MockDataSource) Stop() {
    // 清理资源
}

// 过滤转换器
type FilterTransform struct {
    predicate func(DataItem) bool
}

func NewFilterTransform(predicate func(DataItem) bool) *FilterTransform {
    return &FilterTransform{predicate: predicate}
}

func (ft *FilterTransform) Process(ctx context.Context, input <-chan DataItem) <-chan DataItem {
    output := make(chan DataItem, 100)

    go func() {
        defer close(output)

        for {
            select {
            case <-ctx.Done():
                return
            case item, ok := <-input:
                if !ok {
                    return
                }

                if ft.predicate(item) {
                    select {
                    case output <- item:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
    }()

    return output
}

// 映射转换器
type MapTransform struct {
    mapper func(DataItem) DataItem
}

func NewMapTransform(mapper func(DataItem) DataItem) *MapTransform {
    return &MapTransform{mapper: mapper}
}

func (mt *MapTransform) Process(ctx context.Context, input <-chan DataItem) <-chan DataItem {
    output := make(chan DataItem, 100)

    go func() {
        defer close(output)

        for {
            select {
            case <-ctx.Done():
                return
            case item, ok := <-input:
                if !ok {
                    return
                }

                transformed := mt.mapper(item)

                select {
                case output <- transformed:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()

    return output
}

// 窗口聚合转换器
type WindowAggregateTransform struct {
    windowSize time.Duration
    aggregator func([]DataItem) DataItem
}

func NewWindowAggregateTransform(windowSize time.Duration, aggregator func([]DataItem) DataItem) *WindowAggregateTransform {
    return &WindowAggregateTransform{
        windowSize: windowSize,
        aggregator: aggregator,
    }
}

func (wat *WindowAggregateTransform) Process(ctx context.Context, input <-chan DataItem) <-chan DataItem {
    output := make(chan DataItem, 100)

    go func() {
        defer close(output)

        var window []DataItem
        ticker := time.NewTicker(wat.windowSize)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case item, ok := <-input:
                if !ok {
                    // 处理剩余窗口数据
                    if len(window) > 0 {
                        aggregated := wat.aggregator(window)
                        select {
                        case output <- aggregated:
                        case <-ctx.Done():
                        }
                    }
                    return
                }

                window = append(window, item)

            case <-ticker.C:
                if len(window) > 0 {
                    aggregated := wat.aggregator(window)
                    window = window[:0] // 清空窗口

                    select {
                    case output <- aggregated:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
    }()

    return output
}

// 控制台输出
type ConsoleSink struct {
    name string
}

func NewConsoleSink(name string) *ConsoleSink {
    return &ConsoleSink{name: name}
}

func (cs *ConsoleSink) Write(ctx context.Context, input <-chan DataItem) {
    for {
        select {
        case <-ctx.Done():
            return
        case item, ok := <-input:
            if !ok {
                return
            }

            fmt.Printf("[%s] %s: %v (时间: %v)\n",
                cs.name, item.ID, item.Value, item.Timestamp.Format("15:04:05.000"))
        }
    }
}

// 创建流处理器
func NewStreamProcessor(backpressureSize int) *StreamProcessor {
    ctx, cancel := context.WithCancel(context.Background())

    return &StreamProcessor{
        ctx:          ctx,
        cancel:       cancel,
        backpressure: make(chan struct{}, backpressureSize),
    }
}

// 添加数据源
func (sp *StreamProcessor) AddSource(source DataSource) {
    sp.sources = append(sp.sources, source)
}

// 添加转换器
func (sp *StreamProcessor) AddTransform(transform Transform) {
    sp.transforms = append(sp.transforms, transform)
}

// 添加输出
func (sp *StreamProcessor) AddSink(sink DataSink) {
    sp.sinks = append(sp.sinks, sink)
}

// 启动处理器
func (sp *StreamProcessor) Start() {
    fmt.Println("启动数据流处理器...")

    // 合并所有数据源
    var sourceChannels []<-chan DataItem
    for _, source := range sp.sources {
        sourceChannels = append(sourceChannels, source.Start(sp.ctx))
    }

    merged := sp.mergeChannels(sourceChannels...)

    // 应用所有转换
    current := merged
    for _, transform := range sp.transforms {
        current = transform.Process(sp.ctx, current)
    }

    // 启动所有输出
    for _, sink := range sp.sinks {
        sp.wg.Add(1)
        go func(s DataSink, input <-chan DataItem) {
            defer sp.wg.Done()
            s.Write(sp.ctx, input)
        }(sink, current)
    }
}

// 合并多个Channel
func (sp *StreamProcessor) mergeChannels(inputs ...<-chan DataItem) <-chan DataItem {
    output := make(chan DataItem, 100)
    var wg sync.WaitGroup

    for _, input := range inputs {
        wg.Add(1)
        go func(ch <-chan DataItem) {
            defer wg.Done()

            for {
                select {
                case <-sp.ctx.Done():
                    return
                case item, ok := <-ch:
                    if !ok {
                        return
                    }

                    // 背压控制
                    select {
                    case sp.backpressure <- struct{}{}:
                        atomic.AddInt64(&sp.processedCount, 1)

                        select {
                        case output <- item:
                            <-sp.backpressure
                        case <-sp.ctx.Done():
                            <-sp.backpressure
                            return
                        }
                    default:
                        // 背压满了，丢弃数据
                        atomic.AddInt64(&sp.droppedCount, 1)
                    }
                }
            }
        }(input)
    }

    go func() {
        wg.Wait()
        close(output)
    }()

    return output
}

// 获取统计信息
func (sp *StreamProcessor) GetStats() (int64, int64) {
    return atomic.LoadInt64(&sp.processedCount), atomic.LoadInt64(&sp.droppedCount)
}

// 停止处理器
func (sp *StreamProcessor) Stop() {
    fmt.Println("停止数据流处理器...")

    // 停止所有数据源
    for _, source := range sp.sources {
        source.Stop()
    }

    sp.cancel()
    sp.wg.Wait()

    fmt.Println("数据流处理器已停止")
}

// 测试函数
func TestStreamProcessor() {
    processor := NewStreamProcessor(50)

    // 添加数据源
    processor.AddSource(NewMockDataSource("source1", 100*time.Millisecond))
    processor.AddSource(NewMockDataSource("source2", 150*time.Millisecond))

    // 添加转换器
    // 1. 过滤：只保留偶数值
    processor.AddTransform(NewFilterTransform(func(item DataItem) bool {
        if val, ok := item.Value.(int64); ok {
            return val%2 == 0
        }
        return false
    }))

    // 2. 映射：值乘以2
    processor.AddTransform(NewMapTransform(func(item DataItem) DataItem {
        if val, ok := item.Value.(int64); ok {
            item.Value = val * 2
        }
        return item
    }))

    // 3. 窗口聚合：每2秒聚合一次
    processor.AddTransform(NewWindowAggregateTransform(2*time.Second, func(items []DataItem) DataItem {
        var sum int64
        for _, item := range items {
            if val, ok := item.Value.(int64); ok {
                sum += val
            }
        }

        return DataItem{
            ID:        fmt.Sprintf("aggregated-%d", time.Now().Unix()),
            Timestamp: time.Now(),
            Value:     sum,
            Metadata: map[string]interface{}{
                "type":  "aggregated",
                "count": len(items),
            },
        }
    }))

    // 添加输出
    processor.AddSink(NewConsoleSink("output"))

    // 启动处理器
    processor.Start()

    // 运行10秒
    time.Sleep(10 * time.Second)

    // 获取统计信息
    processed, dropped := processor.GetStats()
    fmt.Printf("处理统计 - 已处理: %d, 已丢弃: %d\n", processed, dropped)

    processor.Stop()

    if processed > 0 {
        fmt.Println("✅ 练习题5通过")
    } else {
        fmt.Println("❌ 练习题5失败")
    }
}
```

---

## 📚 章节总结

### 🎯 核心知识点回顾

通过本章的学习，我们深入掌握了Go语言并发编程的核心概念和实践技巧：

#### 1. **Goroutine - 轻量级并发**
- **核心特性**：M:N调度模型、动态栈、低内存占用
- **与传统线程对比**：创建成本低、切换开销小、内存占用少
- **最佳实践**：合理控制Goroutine数量、避免Goroutine泄漏

#### 2. **Channel - 通信机制**
- **设计哲学**："Don't communicate by sharing memory; share memory by communicating"
- **类型分类**：无缓冲Channel、有缓冲Channel、单向Channel
- **使用模式**：生产者-消费者、管道、扇入扇出

#### 3. **Select - 多路复用**
- **核心功能**：非阻塞通信、超时控制、多Channel选择
- **应用场景**：超时处理、取消操作、负载均衡

#### 4. **同步原语**
- **Mutex/RWMutex**：保护共享资源、读写分离
- **WaitGroup**：等待多个Goroutine完成
- **Once**：确保函数只执行一次
- **原子操作**：高性能的基础同步

#### 5. **Context包**
- **核心作用**：请求范围数据传递、取消信号、超时控制
- **最佳实践**：不存储业务数据、及时取消、合理设置超时

#### 6. **并发模式**
- **Worker Pool**：固定数量工作者处理任务队列
- **Pipeline**：流水线处理，分阶段处理数据
- **Fan-in/Fan-out**：数据聚合与分发

### 🆚 与Java/Python并发对比总结

| 特性 | Go | Java | Python |
|------|----|----- |--------|
| **并发模型** | CSP模型，Goroutine+Channel | 共享内存，Thread+锁 | GIL限制，asyncio协程 |
| **创建成本** | 极低（2KB栈） | 高（1MB栈） | 中等（协程轻量） |
| **通信方式** | Channel优先 | 共享变量+锁 | Queue/asyncio |
| **调度器** | Go运行时调度 | OS线程调度 | 事件循环调度 |
| **性能** | 高并发性能优秀 | 线程切换开销大 | GIL限制CPU密集型 |
| **易用性** | 语法简洁，概念清晰 | 复杂的锁机制 | async/await语法 |

### 🛠️ 实战应用场景

#### 1. **Web服务器**
- 每个请求一个Goroutine
- 使用Context传递请求信息
- 连接池管理数据库连接

#### 2. **微服务架构**
- 服务间异步通信
- 熔断器和限流器
- 分布式追踪

#### 3. **数据处理**
- 并行数据处理
- 流式数据处理
- ETL管道

#### 4. **实时系统**
- WebSocket连接管理
- 消息推送系统
- 实时数据同步

### ⚠️ 常见陷阱和解决方案

#### 1. **Goroutine泄漏**
- **问题**：Goroutine无法正常退出
- **解决**：使用Context控制生命周期

#### 2. **Channel死锁**
- **问题**：发送者和接收者相互等待
- **解决**：使用select和超时机制

#### 3. **竞态条件**
- **问题**：并发访问共享资源
- **解决**：使用锁或原子操作

#### 4. **内存泄漏**
- **问题**：Channel或Goroutine未正确清理
- **解决**：及时关闭Channel，使用defer清理

### 🚀 性能优化技巧

#### 1. **合理设置并发数**
- 根据CPU核心数设置Worker数量
- 使用信号量控制并发度
- 监控系统资源使用情况

#### 2. **减少锁竞争**
- 使用读写锁分离读写操作
- 分段锁减少锁粒度
- 优先使用原子操作

#### 3. **Channel优化**
- 合理设置缓冲区大小
- 避免频繁创建Channel
- 及时关闭不再使用的Channel

#### 4. **内存管理**
- 复用对象减少GC压力
- 使用sync.Pool对象池
- 避免不必要的内存分配

### 📈 学习路径建议

#### **已完成的内容** ✅
- **基础篇全部4章**：变量类型、控制结构、函数方法、包管理
- **进阶篇前3章**：结构体接口、错误处理、并发编程基础

#### **下一步学习方向** 🎯

1. **进阶篇第4章：接口设计模式**
   - 设计模式在Go中的实现
   - 依赖注入和控制反转
   - 接口组合和嵌入

2. **进阶篇第5章：反射和泛型**
   - 反射机制的使用和性能考虑
   - Go 1.18+泛型特性
   - 类型约束和类型推断

3. **实战篇：Web应用开发**
   - Gin框架深入使用
   - 数据库操作和ORM
   - 缓存和消息队列集成

4. **架构篇：微服务设计**
   - 服务发现和配置管理
   - API网关和负载均衡
   - 分布式事务和一致性

### 🎊 重要里程碑

**🎉 恭喜！你已经完成了Go语言并发编程基础的学习！**

现在你已经具备了：
- **企业级并发编程能力**：能够设计和实现高性能的并发系统
- **问题诊断和优化能力**：能够识别和解决并发程序中的问题
- **架构设计能力**：能够选择合适的并发模式解决实际问题
- **性能调优能力**：能够使用工具分析和优化并发程序性能

这些技能将为你后续学习微服务架构、分布式系统和云原生开发奠定坚实的基础！

### 🔗 相关资源推荐

#### **官方文档**
- [Go并发模式](https://go.dev/blog/pipelines)
- [Go内存模型](https://go.dev/ref/mem)
- [Context包文档](https://pkg.go.dev/context)

#### **推荐阅读**
- 《Go语言并发之道》- Katherine Cox-Buday
- 《Go语言实战》- William Kennedy
- 《Effective Go》- Go官方团队

#### **实践项目**
- 实现一个简单的Web爬虫
- 构建一个聊天服务器
- 开发一个任务调度系统

继续加油！下一章我们将学习Go语言的接口设计模式，探索更高级的编程技巧！ 🚀
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
