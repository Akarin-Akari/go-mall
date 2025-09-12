# 高级篇第四章：性能优化技巧 🚀

> *"性能优化是一门艺术，需要在功能、性能、可维护性之间找到最佳平衡点。优化不是银弹，而是基于数据驱动的科学决策过程！"* ⚡

## 📚 本章学习目标

通过本章学习，你将掌握：

- 🔍 **Go程序性能分析工具**：深入掌握pprof、trace、go tool等性能分析利器
- 🧠 **内存优化策略**：内存泄漏检测、GC调优、对象池等内存管理技巧
- ⚡ **并发性能调优**：goroutine池、channel优化、锁优化等并发编程优化
- 🗄️ **数据库性能优化**：连接池、查询优化、索引设计等数据库优化策略
- 💾 **缓存策略优化**：多级缓存、缓存穿透/雪崩防护等缓存优化技术
- 🌐 **网络性能优化**：HTTP/2、连接复用、压缩等网络层优化
- 🏢 **Mall-Go性能优化实战**：结合电商项目的完整性能优化案例

---

## 🔍 Go程序性能分析工具详解

### pprof性能分析

pprof是Go语言内置的性能分析工具，能够帮助我们识别程序的性能瓶颈。

```go
// 性能分析工具集成
package profiling

import (
    "context"
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "os"
    "runtime"
    "runtime/pprof"
    "runtime/trace"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/pkg/profile"
)

// 性能分析管理器
type ProfilingManager struct {
    enabled     bool
    profileType string
    outputDir   string
    server      *http.Server
}

// 性能分析配置
type ProfilingConfig struct {
    Enabled     bool   `json:"enabled"`
    Port        int    `json:"port"`
    ProfileType string `json:"profile_type"` // cpu, mem, block, mutex, goroutine, trace
    OutputDir   string `json:"output_dir"`
    Duration    int    `json:"duration"`     // 分析持续时间（秒）
}

// 创建性能分析管理器
func NewProfilingManager(config *ProfilingConfig) *ProfilingManager {
    return &ProfilingManager{
        enabled:     config.Enabled,
        profileType: config.ProfileType,
        outputDir:   config.OutputDir,
    }
}

// 启动性能分析服务
func (pm *ProfilingManager) StartProfilingServer(port int) error {
    if !pm.enabled {
        return nil
    }
    
    mux := http.NewServeMux()
    
    // 注册pprof路由
    mux.HandleFunc("/debug/pprof/", http.HandlerFunc(pprof.Index))
    mux.HandleFunc("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
    mux.HandleFunc("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
    mux.HandleFunc("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
    mux.HandleFunc("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
    
    // 自定义分析端点
    mux.HandleFunc("/debug/pprof/heap", pm.heapProfile)
    mux.HandleFunc("/debug/pprof/goroutine", pm.goroutineProfile)
    mux.HandleFunc("/debug/pprof/block", pm.blockProfile)
    mux.HandleFunc("/debug/pprof/mutex", pm.mutexProfile)
    
    // 性能分析控制端点
    mux.HandleFunc("/profiling/start", pm.startProfiling)
    mux.HandleFunc("/profiling/stop", pm.stopProfiling)
    mux.HandleFunc("/profiling/status", pm.profilingStatus)
    
    pm.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    
    go func() {
        if err := pm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("Profiling server error: %v\n", err)
        }
    }()
    
    fmt.Printf("Profiling server started on port %d\n", port)
    return nil
}

// 堆内存分析
func (pm *ProfilingManager) heapProfile(w http.ResponseWriter, r *http.Request) {
    runtime.GC() // 强制GC，获取准确的堆信息
    pprof.WriteHeapProfile(w)
}

// Goroutine分析
func (pm *ProfilingManager) goroutineProfile(w http.ResponseWriter, r *http.Request) {
    pprof.Lookup("goroutine").WriteTo(w, 1)
}

// 阻塞分析
func (pm *ProfilingManager) blockProfile(w http.ResponseWriter, r *http.Request) {
    pprof.Lookup("block").WriteTo(w, 1)
}

// 互斥锁分析
func (pm *ProfilingManager) mutexProfile(w http.ResponseWriter, r *http.Request) {
    pprof.Lookup("mutex").WriteTo(w, 1)
}

// 开始性能分析
func (pm *ProfilingManager) startProfiling(w http.ResponseWriter, r *http.Request) {
    profileType := r.URL.Query().Get("type")
    if profileType == "" {
        profileType = "cpu"
    }
    
    duration := 30 * time.Second
    if d := r.URL.Query().Get("duration"); d != "" {
        if parsed, err := time.ParseDuration(d); err == nil {
            duration = parsed
        }
    }
    
    switch profileType {
    case "cpu":
        pm.startCPUProfiling(duration)
    case "mem":
        pm.startMemoryProfiling(duration)
    case "trace":
        pm.startTraceProfiling(duration)
    default:
        http.Error(w, "Unsupported profile type", http.StatusBadRequest)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Started %s profiling for %v\n", profileType, duration)
}

// CPU性能分析
func (pm *ProfilingManager) startCPUProfiling(duration time.Duration) {
    filename := fmt.Sprintf("%s/cpu_profile_%d.prof", pm.outputDir, time.Now().Unix())
    f, err := os.Create(filename)
    if err != nil {
        fmt.Printf("Failed to create CPU profile file: %v\n", err)
        return
    }
    
    if err := pprof.StartCPUProfile(f); err != nil {
        fmt.Printf("Failed to start CPU profile: %v\n", err)
        f.Close()
        return
    }
    
    go func() {
        time.Sleep(duration)
        pprof.StopCPUProfile()
        f.Close()
        fmt.Printf("CPU profile saved to %s\n", filename)
    }()
}

// 内存性能分析
func (pm *ProfilingManager) startMemoryProfiling(duration time.Duration) {
    go func() {
        time.Sleep(duration)
        filename := fmt.Sprintf("%s/mem_profile_%d.prof", pm.outputDir, time.Now().Unix())
        f, err := os.Create(filename)
        if err != nil {
            fmt.Printf("Failed to create memory profile file: %v\n", err)
            return
        }
        defer f.Close()
        
        runtime.GC() // 强制GC
        if err := pprof.WriteHeapProfile(f); err != nil {
            fmt.Printf("Failed to write memory profile: %v\n", err)
            return
        }
        
        fmt.Printf("Memory profile saved to %s\n", filename)
    }()
}

// 执行追踪分析
func (pm *ProfilingManager) startTraceProfiling(duration time.Duration) {
    filename := fmt.Sprintf("%s/trace_%d.out", pm.outputDir, time.Now().Unix())
    f, err := os.Create(filename)
    if err != nil {
        fmt.Printf("Failed to create trace file: %v\n", err)
        return
    }
    
    if err := trace.Start(f); err != nil {
        fmt.Printf("Failed to start trace: %v\n", err)
        f.Close()
        return
    }
    
    go func() {
        time.Sleep(duration)
        trace.Stop()
        f.Close()
        fmt.Printf("Trace saved to %s\n", filename)
    }()
}

// 停止性能分析
func (pm *ProfilingManager) stopProfiling(w http.ResponseWriter, r *http.Request) {
    pprof.StopCPUProfile()
    trace.Stop()
    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "Profiling stopped\n")
}

// 性能分析状态
func (pm *ProfilingManager) profilingStatus(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "enabled":     pm.enabled,
        "goroutines":  runtime.NumGoroutine(),
        "memory":      getMemoryStats(),
        "gc_stats":    getGCStats(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, "%+v\n", status)
}

// 获取内存统计信息
func getMemoryStats() map[string]interface{} {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    return map[string]interface{}{
        "alloc":         m.Alloc,
        "total_alloc":   m.TotalAlloc,
        "sys":           m.Sys,
        "heap_alloc":    m.HeapAlloc,
        "heap_sys":      m.HeapSys,
        "heap_idle":     m.HeapIdle,
        "heap_inuse":    m.HeapInuse,
        "heap_objects":  m.HeapObjects,
        "stack_inuse":   m.StackInuse,
        "stack_sys":     m.StackSys,
        "next_gc":       m.NextGC,
        "last_gc":       m.LastGC,
        "num_gc":        m.NumGC,
        "gc_cpu_fraction": m.GCCPUFraction,
    }
}

// 获取GC统计信息
func getGCStats() map[string]interface{} {
    var stats runtime.GCStats
    runtime.ReadGCStats(&stats)
    
    return map[string]interface{}{
        "last_gc":       stats.LastGC,
        "num_gc":        stats.NumGC,
        "pause_total":   stats.PauseTotal,
        "pause_ns":      stats.PauseNs,
        "pause_end":     stats.PauseEnd,
        "pause_quantiles": stats.PauseQuantiles,
    }
}

// Gin中间件 - 性能监控
func (pm *ProfilingManager) PerformanceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !pm.enabled {
            c.Next()
            return
        }
        
        start := time.Now()
        
        // 记录请求开始时的资源状态
        var startMem runtime.MemStats
        runtime.ReadMemStats(&startMem)
        startGoroutines := runtime.NumGoroutine()
        
        // 处理请求
        c.Next()
        
        // 记录请求结束时的资源状态
        var endMem runtime.MemStats
        runtime.ReadMemStats(&endMem)
        endGoroutines := runtime.NumGoroutine()
        
        duration := time.Since(start)
        
        // 记录性能指标
        c.Header("X-Response-Time", duration.String())
        c.Header("X-Memory-Alloc", fmt.Sprintf("%d", endMem.Alloc-startMem.Alloc))
        c.Header("X-Goroutines", fmt.Sprintf("%d", endGoroutines))
        c.Header("X-Goroutine-Delta", fmt.Sprintf("%d", endGoroutines-startGoroutines))
        
        // 如果请求时间过长，记录详细信息
        if duration > 1*time.Second {
            fmt.Printf("Slow request detected: %s %s took %v\n", 
                c.Request.Method, c.Request.URL.Path, duration)
        }
    }
}

// 自动性能分析器
type AutoProfiler struct {
    manager     *ProfilingManager
    thresholds  *PerformanceThresholds
    enabled     bool
    lastProfile time.Time
}

// 性能阈值配置
type PerformanceThresholds struct {
    CPUUsage        float64       `json:"cpu_usage"`         // CPU使用率阈值
    MemoryUsage     uint64        `json:"memory_usage"`      // 内存使用量阈值
    GoroutineCount  int           `json:"goroutine_count"`   // Goroutine数量阈值
    GCPause         time.Duration `json:"gc_pause"`          // GC暂停时间阈值
    ResponseTime    time.Duration `json:"response_time"`     // 响应时间阈值
    ProfileInterval time.Duration `json:"profile_interval"`  // 自动分析间隔
}

// 创建自动性能分析器
func NewAutoProfiler(manager *ProfilingManager, thresholds *PerformanceThresholds) *AutoProfiler {
    return &AutoProfiler{
        manager:    manager,
        thresholds: thresholds,
        enabled:    true,
    }
}

// 启动自动性能分析
func (ap *AutoProfiler) Start(ctx context.Context) {
    if !ap.enabled {
        return
    }
    
    ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ap.checkAndProfile()
        }
    }
}

// 检查并执行性能分析
func (ap *AutoProfiler) checkAndProfile() {
    // 检查是否需要进行性能分析
    if time.Since(ap.lastProfile) < ap.thresholds.ProfileInterval {
        return
    }
    
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    goroutineCount := runtime.NumGoroutine()
    
    // 检查各项指标是否超过阈值
    needProfile := false
    profileType := "cpu"
    
    if m.Alloc > ap.thresholds.MemoryUsage {
        needProfile = true
        profileType = "mem"
        fmt.Printf("Memory usage threshold exceeded: %d > %d\n", m.Alloc, ap.thresholds.MemoryUsage)
    }
    
    if goroutineCount > ap.thresholds.GoroutineCount {
        needProfile = true
        profileType = "goroutine"
        fmt.Printf("Goroutine count threshold exceeded: %d > %d\n", goroutineCount, ap.thresholds.GoroutineCount)
    }
    
    if m.PauseNs[(m.NumGC+255)%256] > uint64(ap.thresholds.GCPause) {
        needProfile = true
        profileType = "cpu"
        fmt.Printf("GC pause threshold exceeded: %d > %d\n", 
            m.PauseNs[(m.NumGC+255)%256], ap.thresholds.GCPause)
    }
    
    if needProfile {
        ap.triggerProfiling(profileType)
        ap.lastProfile = time.Now()
    }
}

// 触发性能分析
func (ap *AutoProfiler) triggerProfiling(profileType string) {
    fmt.Printf("Auto-triggering %s profiling due to threshold breach\n", profileType)
    
    switch profileType {
    case "cpu":
        ap.manager.startCPUProfiling(30 * time.Second)
    case "mem":
        ap.manager.startMemoryProfiling(10 * time.Second)
    case "trace":
        ap.manager.startTraceProfiling(15 * time.Second)
    }
}
```

---

## 🧠 内存优化策略

### 内存泄漏检测

内存泄漏是Go程序性能问题的常见原因，需要系统性的检测和预防机制。

```go
// 内存泄漏检测器
package memory

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "time"
)

// 内存泄漏检测器
type MemoryLeakDetector struct {
    enabled         bool
    checkInterval   time.Duration
    thresholds      *MemoryThresholds
    history         []MemorySnapshot
    historyMutex    sync.RWMutex
    maxHistorySize  int
    alertCallback   func(alert *MemoryAlert)
}

// 内存阈值配置
type MemoryThresholds struct {
    HeapGrowthRate    float64 `json:"heap_growth_rate"`    // 堆增长率阈值 (%)
    GoroutineGrowth   int     `json:"goroutine_growth"`    // Goroutine增长阈值
    GCFrequency       int     `json:"gc_frequency"`        // GC频率阈值 (次/分钟)
    MemoryUsageLimit  uint64  `json:"memory_usage_limit"`  // 内存使用限制 (bytes)
    LeakDetectionTime time.Duration `json:"leak_detection_time"` // 泄漏检测时间窗口
}

// 内存快照
type MemorySnapshot struct {
    Timestamp       time.Time `json:"timestamp"`
    HeapAlloc       uint64    `json:"heap_alloc"`
    HeapSys         uint64    `json:"heap_sys"`
    HeapIdle        uint64    `json:"heap_idle"`
    HeapInuse       uint64    `json:"heap_inuse"`
    HeapObjects     uint64    `json:"heap_objects"`
    GoroutineCount  int       `json:"goroutine_count"`
    NumGC           uint32    `json:"num_gc"`
    GCCPUFraction   float64   `json:"gc_cpu_fraction"`
    NextGC          uint64    `json:"next_gc"`
    LastGC          uint64    `json:"last_gc"`
}

// 内存告警
type MemoryAlert struct {
    Type        string    `json:"type"`
    Severity    string    `json:"severity"`
    Message     string    `json:"message"`
    Timestamp   time.Time `json:"timestamp"`
    CurrentMem  uint64    `json:"current_mem"`
    PreviousMem uint64    `json:"previous_mem"`
    GrowthRate  float64   `json:"growth_rate"`
    Suggestions []string  `json:"suggestions"`
}

// 创建内存泄漏检测器
func NewMemoryLeakDetector(thresholds *MemoryThresholds, alertCallback func(*MemoryAlert)) *MemoryLeakDetector {
    return &MemoryLeakDetector{
        enabled:        true,
        checkInterval:  30 * time.Second,
        thresholds:     thresholds,
        history:        make([]MemorySnapshot, 0),
        maxHistorySize: 100,
        alertCallback:  alertCallback,
    }
}

// 启动内存监控
func (mld *MemoryLeakDetector) Start(ctx context.Context) {
    if !mld.enabled {
        return
    }

    ticker := time.NewTicker(mld.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            mld.checkMemoryUsage()
        }
    }
}

// 检查内存使用情况
func (mld *MemoryLeakDetector) checkMemoryUsage() {
    snapshot := mld.takeSnapshot()
    mld.addSnapshot(snapshot)

    // 分析内存趋势
    if len(mld.history) >= 2 {
        mld.analyzeMemoryTrend()
    }

    // 检查内存泄漏
    if len(mld.history) >= 5 {
        mld.detectMemoryLeak()
    }
}

// 获取内存快照
func (mld *MemoryLeakDetector) takeSnapshot() MemorySnapshot {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return MemorySnapshot{
        Timestamp:      time.Now(),
        HeapAlloc:      m.Alloc,
        HeapSys:        m.Sys,
        HeapIdle:       m.HeapIdle,
        HeapInuse:      m.HeapInuse,
        HeapObjects:    m.HeapObjects,
        GoroutineCount: runtime.NumGoroutine(),
        NumGC:          m.NumGC,
        GCCPUFraction:  m.GCCPUFraction,
        NextGC:         m.NextGC,
        LastGC:         m.LastGC,
    }
}

// 添加快照到历史记录
func (mld *MemoryLeakDetector) addSnapshot(snapshot MemorySnapshot) {
    mld.historyMutex.Lock()
    defer mld.historyMutex.Unlock()

    mld.history = append(mld.history, snapshot)

    // 保持历史记录大小限制
    if len(mld.history) > mld.maxHistorySize {
        mld.history = mld.history[1:]
    }
}

// 分析内存趋势
func (mld *MemoryLeakDetector) analyzeMemoryTrend() {
    mld.historyMutex.RLock()
    defer mld.historyMutex.RUnlock()

    if len(mld.history) < 2 {
        return
    }

    current := mld.history[len(mld.history)-1]
    previous := mld.history[len(mld.history)-2]

    // 计算内存增长率
    heapGrowthRate := float64(current.HeapAlloc-previous.HeapAlloc) / float64(previous.HeapAlloc) * 100
    goroutineGrowth := current.GoroutineCount - previous.GoroutineCount

    // 检查堆内存增长
    if heapGrowthRate > mld.thresholds.HeapGrowthRate {
        alert := &MemoryAlert{
            Type:        "heap_growth",
            Severity:    "warning",
            Message:     fmt.Sprintf("Heap memory growth rate %.2f%% exceeds threshold %.2f%%", heapGrowthRate, mld.thresholds.HeapGrowthRate),
            Timestamp:   time.Now(),
            CurrentMem:  current.HeapAlloc,
            PreviousMem: previous.HeapAlloc,
            GrowthRate:  heapGrowthRate,
            Suggestions: []string{
                "Check for memory leaks in recent code changes",
                "Review object lifecycle management",
                "Consider using object pools for frequently allocated objects",
            },
        }
        mld.sendAlert(alert)
    }

    // 检查Goroutine增长
    if goroutineGrowth > mld.thresholds.GoroutineGrowth {
        alert := &MemoryAlert{
            Type:      "goroutine_growth",
            Severity:  "warning",
            Message:   fmt.Sprintf("Goroutine count increased by %d, exceeds threshold %d", goroutineGrowth, mld.thresholds.GoroutineGrowth),
            Timestamp: time.Now(),
            Suggestions: []string{
                "Check for goroutine leaks",
                "Ensure all goroutines have proper exit conditions",
                "Review context cancellation usage",
            },
        }
        mld.sendAlert(alert)
    }

    // 检查内存使用限制
    if current.HeapAlloc > mld.thresholds.MemoryUsageLimit {
        alert := &MemoryAlert{
            Type:       "memory_limit",
            Severity:   "critical",
            Message:    fmt.Sprintf("Memory usage %d bytes exceeds limit %d bytes", current.HeapAlloc, mld.thresholds.MemoryUsageLimit),
            Timestamp:  time.Now(),
            CurrentMem: current.HeapAlloc,
            Suggestions: []string{
                "Immediate memory cleanup required",
                "Consider triggering manual GC",
                "Review large object allocations",
                "Implement memory pressure handling",
            },
        }
        mld.sendAlert(alert)
    }
}

// 检测内存泄漏
func (mld *MemoryLeakDetector) detectMemoryLeak() {
    mld.historyMutex.RLock()
    defer mld.historyMutex.RUnlock()

    if len(mld.history) < 5 {
        return
    }

    // 分析最近5个快照的趋势
    recent := mld.history[len(mld.history)-5:]

    // 检查内存是否持续增长
    isLeaking := true
    for i := 1; i < len(recent); i++ {
        if recent[i].HeapAlloc <= recent[i-1].HeapAlloc {
            isLeaking = false
            break
        }
    }

    if isLeaking {
        // 计算平均增长率
        totalGrowth := float64(recent[len(recent)-1].HeapAlloc - recent[0].HeapAlloc)
        avgGrowthRate := totalGrowth / float64(recent[0].HeapAlloc) * 100

        alert := &MemoryAlert{
            Type:      "memory_leak",
            Severity:  "critical",
            Message:   fmt.Sprintf("Potential memory leak detected: continuous growth over %v with average rate %.2f%%", mld.thresholds.LeakDetectionTime, avgGrowthRate),
            Timestamp: time.Now(),
            CurrentMem: recent[len(recent)-1].HeapAlloc,
            PreviousMem: recent[0].HeapAlloc,
            GrowthRate: avgGrowthRate,
            Suggestions: []string{
                "Perform heap dump analysis",
                "Review recent code changes for resource leaks",
                "Check for unclosed resources (files, connections, etc.)",
                "Analyze goroutine stack traces",
                "Consider implementing memory profiling",
            },
        }
        mld.sendAlert(alert)
    }
}

// 发送告警
func (mld *MemoryLeakDetector) sendAlert(alert *MemoryAlert) {
    if mld.alertCallback != nil {
        mld.alertCallback(alert)
    }
}

// 获取内存统计报告
func (mld *MemoryLeakDetector) GetMemoryReport() *MemoryReport {
    mld.historyMutex.RLock()
    defer mld.historyMutex.RUnlock()

    if len(mld.history) == 0 {
        return nil
    }

    current := mld.history[len(mld.history)-1]

    report := &MemoryReport{
        Timestamp:      time.Now(),
        CurrentSnapshot: current,
        HistoryCount:   len(mld.history),
    }

    if len(mld.history) >= 2 {
        previous := mld.history[len(mld.history)-2]
        report.GrowthRate = float64(current.HeapAlloc-previous.HeapAlloc) / float64(previous.HeapAlloc) * 100
        report.GoroutineGrowth = current.GoroutineCount - previous.GoroutineCount
    }

    return report
}

// 内存报告
type MemoryReport struct {
    Timestamp       time.Time      `json:"timestamp"`
    CurrentSnapshot MemorySnapshot `json:"current_snapshot"`
    HistoryCount    int            `json:"history_count"`
    GrowthRate      float64        `json:"growth_rate"`
    GoroutineGrowth int            `json:"goroutine_growth"`
    Recommendations []string       `json:"recommendations"`
}
```

### GC调优策略

```go
// GC调优管理器
package gc

import (
    "fmt"
    "runtime"
    "runtime/debug"
    "time"
)

// GC调优管理器
type GCTuningManager struct {
    config          *GCConfig
    originalGCGoal  int
    originalMaxProcs int
    stats           *GCStats
}

// GC配置
type GCConfig struct {
    // GC目标百分比 (默认100)
    GCGoal int `json:"gc_goal"`

    // 最大并行GC线程数
    MaxGCProcs int `json:"max_gc_procs"`

    // 内存限制 (bytes)
    MemoryLimit int64 `json:"memory_limit"`

    // 是否启用GC调优
    Enabled bool `json:"enabled"`

    // 自动调优配置
    AutoTuning struct {
        Enabled           bool          `json:"enabled"`
        TargetLatency     time.Duration `json:"target_latency"`     // 目标GC延迟
        TargetThroughput  float64       `json:"target_throughput"`  // 目标吞吐量
        AdjustmentFactor  float64       `json:"adjustment_factor"`  // 调整因子
        MonitorInterval   time.Duration `json:"monitor_interval"`   // 监控间隔
    } `json:"auto_tuning"`
}

// GC统计信息
type GCStats struct {
    NumGC           uint32        `json:"num_gc"`
    PauseTotal      time.Duration `json:"pause_total"`
    PauseAvg        time.Duration `json:"pause_avg"`
    PauseMax        time.Duration `json:"pause_max"`
    PauseMin        time.Duration `json:"pause_min"`
    LastGC          time.Time     `json:"last_gc"`
    GCCPUFraction   float64       `json:"gc_cpu_fraction"`
    HeapSize        uint64        `json:"heap_size"`
    NextGC          uint64        `json:"next_gc"`
    Frequency       float64       `json:"frequency"` // GC频率 (次/秒)
}

// 创建GC调优管理器
func NewGCTuningManager(config *GCConfig) *GCTuningManager {
    return &GCTuningManager{
        config:          config,
        originalGCGoal:  debug.SetGCPercent(-1), // 获取原始值
        originalMaxProcs: runtime.GOMAXPROCS(0),
        stats:           &GCStats{},
    }
}

// 应用GC配置
func (gtm *GCTuningManager) ApplyGCConfig() {
    if !gtm.config.Enabled {
        return
    }

    // 设置GC目标百分比
    if gtm.config.GCGoal > 0 {
        debug.SetGCPercent(gtm.config.GCGoal)
        fmt.Printf("Set GC goal to %d%%\n", gtm.config.GCGoal)
    }

    // 设置内存限制
    if gtm.config.MemoryLimit > 0 {
        debug.SetMemoryLimit(gtm.config.MemoryLimit)
        fmt.Printf("Set memory limit to %d bytes\n", gtm.config.MemoryLimit)
    }

    // 设置最大GC线程数
    if gtm.config.MaxGCProcs > 0 {
        runtime.GOMAXPROCS(gtm.config.MaxGCProcs)
        fmt.Printf("Set GOMAXPROCS to %d\n", gtm.config.MaxGCProcs)
    }
}

// 恢复原始GC配置
func (gtm *GCTuningManager) RestoreOriginalConfig() {
    debug.SetGCPercent(gtm.originalGCGoal)
    runtime.GOMAXPROCS(gtm.originalMaxProcs)
    debug.SetMemoryLimit(-1) // 移除内存限制
}

// 收集GC统计信息
func (gtm *GCTuningManager) CollectGCStats() *GCStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    var gcStats debug.GCStats
    debug.ReadGCStats(&gcStats)

    // 计算平均暂停时间
    var pauseAvg time.Duration
    if gcStats.NumGC > 0 {
        pauseAvg = gcStats.PauseTotal / time.Duration(gcStats.NumGC)
    }

    // 计算最大和最小暂停时间
    var pauseMax, pauseMin time.Duration
    if len(gcStats.Pause) > 0 {
        pauseMax = gcStats.Pause[0]
        pauseMin = gcStats.Pause[0]
        for _, pause := range gcStats.Pause {
            if pause > pauseMax {
                pauseMax = pause
            }
            if pause < pauseMin && pause > 0 {
                pauseMin = pause
            }
        }
    }

    // 计算GC频率
    var frequency float64
    if !gcStats.LastGC.IsZero() {
        timeSinceLastGC := time.Since(gcStats.LastGC)
        if timeSinceLastGC > 0 {
            frequency = 1.0 / timeSinceLastGC.Seconds()
        }
    }

    gtm.stats = &GCStats{
        NumGC:         uint32(gcStats.NumGC),
        PauseTotal:    gcStats.PauseTotal,
        PauseAvg:      pauseAvg,
        PauseMax:      pauseMax,
        PauseMin:      pauseMin,
        LastGC:        gcStats.LastGC,
        GCCPUFraction: m.GCCPUFraction,
        HeapSize:      m.HeapAlloc,
        NextGC:        m.NextGC,
        Frequency:     frequency,
    }

    return gtm.stats
}

// 自动GC调优
func (gtm *GCTuningManager) AutoTuneGC() {
    if !gtm.config.AutoTuning.Enabled {
        return
    }

    stats := gtm.CollectGCStats()

    // 根据目标延迟调整GC目标
    if stats.PauseAvg > gtm.config.AutoTuning.TargetLatency {
        // 暂停时间过长，降低GC目标百分比
        newGoal := int(float64(gtm.config.GCGoal) * (1.0 - gtm.config.AutoTuning.AdjustmentFactor))
        if newGoal < 50 { // 最小值限制
            newGoal = 50
        }
        gtm.config.GCGoal = newGoal
        debug.SetGCPercent(newGoal)
        fmt.Printf("Auto-tuned GC goal to %d%% (reducing pause time)\n", newGoal)
    } else if stats.GCCPUFraction > gtm.config.AutoTuning.TargetThroughput {
        // GC CPU占用过高，增加GC目标百分比
        newGoal := int(float64(gtm.config.GCGoal) * (1.0 + gtm.config.AutoTuning.AdjustmentFactor))
        if newGoal > 500 { // 最大值限制
            newGoal = 500
        }
        gtm.config.GCGoal = newGoal
        debug.SetGCPercent(newGoal)
        fmt.Printf("Auto-tuned GC goal to %d%% (improving throughput)\n", newGoal)
    }
}

// 强制GC并测量性能
func (gtm *GCTuningManager) ForceGCAndMeasure() *GCMeasurement {
    start := time.Now()

    var beforeMem runtime.MemStats
    runtime.ReadMemStats(&beforeMem)

    // 强制执行GC
    runtime.GC()

    var afterMem runtime.MemStats
    runtime.ReadMemStats(&afterMem)

    duration := time.Since(start)

    return &GCMeasurement{
        Duration:     duration,
        BeforeAlloc:  beforeMem.Alloc,
        AfterAlloc:   afterMem.Alloc,
        Freed:        beforeMem.Alloc - afterMem.Alloc,
        BeforeNextGC: beforeMem.NextGC,
        AfterNextGC:  afterMem.NextGC,
        Timestamp:    time.Now(),
    }
}

// GC测量结果
type GCMeasurement struct {
    Duration     time.Duration `json:"duration"`
    BeforeAlloc  uint64        `json:"before_alloc"`
    AfterAlloc   uint64        `json:"after_alloc"`
    Freed        uint64        `json:"freed"`
    BeforeNextGC uint64        `json:"before_next_gc"`
    AfterNextGC  uint64        `json:"after_next_gc"`
    Timestamp    time.Time     `json:"timestamp"`
}

// 获取GC调优建议
func (gtm *GCTuningManager) GetTuningRecommendations() []string {
    stats := gtm.CollectGCStats()
    recommendations := make([]string, 0)

    // 基于暂停时间的建议
    if stats.PauseAvg > 10*time.Millisecond {
        recommendations = append(recommendations,
            "GC pause time is high (>10ms). Consider reducing GOGC value or implementing object pooling.")
    }

    // 基于GC频率的建议
    if stats.Frequency > 10 { // 每秒超过10次GC
        recommendations = append(recommendations,
            "GC frequency is high (>10/sec). Consider increasing GOGC value or reducing allocation rate.")
    }

    // 基于CPU占用的建议
    if stats.GCCPUFraction > 0.25 { // GC占用超过25% CPU
        recommendations = append(recommendations,
            "GC CPU usage is high (>25%). Consider optimizing allocation patterns or increasing heap size.")
    }

    // 基于堆大小的建议
    if stats.HeapSize > 1<<30 { // 堆大小超过1GB
        recommendations = append(recommendations,
            "Heap size is large (>1GB). Consider implementing memory pressure handling and object lifecycle management.")
    }

    if len(recommendations) == 0 {
        recommendations = append(recommendations, "GC performance is within acceptable ranges.")
    }

    return recommendations
}
```

### 对象池优化

对象池是减少内存分配和GC压力的重要技术，特别适用于频繁创建和销毁的对象。

```go
// 对象池管理器
package pool

import (
    "bytes"
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// 通用对象池接口
type ObjectPool interface {
    Get() interface{}
    Put(interface{})
    Size() int
    Stats() *PoolStats
    Clear()
}

// 池统计信息
type PoolStats struct {
    Gets        int64     `json:"gets"`         // 获取次数
    Puts        int64     `json:"puts"`         // 归还次数
    Hits        int64     `json:"hits"`         // 命中次数
    Misses      int64     `json:"misses"`       // 未命中次数
    Size        int       `json:"size"`         // 当前大小
    MaxSize     int       `json:"max_size"`     // 最大大小
    HitRate     float64   `json:"hit_rate"`     // 命中率
    LastReset   time.Time `json:"last_reset"`   // 上次重置时间
}

// 高性能对象池
type HighPerformancePool struct {
    pool        sync.Pool
    newFunc     func() interface{}
    resetFunc   func(interface{})
    maxSize     int
    currentSize int64
    stats       *PoolStats
    mutex       sync.RWMutex
}

// 创建高性能对象池
func NewHighPerformancePool(newFunc func() interface{}, resetFunc func(interface{}), maxSize int) *HighPerformancePool {
    return &HighPerformancePool{
        pool: sync.Pool{
            New: newFunc,
        },
        newFunc:   newFunc,
        resetFunc: resetFunc,
        maxSize:   maxSize,
        stats: &PoolStats{
            MaxSize:   maxSize,
            LastReset: time.Now(),
        },
    }
}

// 获取对象
func (hpp *HighPerformancePool) Get() interface{} {
    atomic.AddInt64(&hpp.stats.Gets, 1)

    obj := hpp.pool.Get()
    if obj != nil {
        atomic.AddInt64(&hpp.stats.Hits, 1)
        atomic.AddInt64(&hpp.currentSize, -1)
    } else {
        atomic.AddInt64(&hpp.stats.Misses, 1)
        obj = hpp.newFunc()
    }

    return obj
}

// 归还对象
func (hpp *HighPerformancePool) Put(obj interface{}) {
    if obj == nil {
        return
    }

    atomic.AddInt64(&hpp.stats.Puts, 1)

    // 检查池大小限制
    if int(atomic.LoadInt64(&hpp.currentSize)) >= hpp.maxSize {
        return // 丢弃对象，避免池过大
    }

    // 重置对象状态
    if hpp.resetFunc != nil {
        hpp.resetFunc(obj)
    }

    hpp.pool.Put(obj)
    atomic.AddInt64(&hpp.currentSize, 1)
}

// 获取池大小
func (hpp *HighPerformancePool) Size() int {
    return int(atomic.LoadInt64(&hpp.currentSize))
}

// 获取统计信息
func (hpp *HighPerformancePool) Stats() *PoolStats {
    hpp.mutex.RLock()
    defer hpp.mutex.RUnlock()

    gets := atomic.LoadInt64(&hpp.stats.Gets)
    hits := atomic.LoadInt64(&hpp.stats.Hits)

    stats := *hpp.stats
    stats.Gets = gets
    stats.Puts = atomic.LoadInt64(&hpp.stats.Puts)
    stats.Hits = hits
    stats.Misses = atomic.LoadInt64(&hpp.stats.Misses)
    stats.Size = hpp.Size()

    if gets > 0 {
        stats.HitRate = float64(hits) / float64(gets) * 100
    }

    return &stats
}

// 清空池
func (hpp *HighPerformancePool) Clear() {
    // 创建新的sync.Pool来清空
    hpp.pool = sync.Pool{
        New: hpp.newFunc,
    }
    atomic.StoreInt64(&hpp.currentSize, 0)
}

// 字节缓冲池
type ByteBufferPool struct {
    pool        *HighPerformancePool
    initialSize int
    maxSize     int
}

// 创建字节缓冲池
func NewByteBufferPool(initialSize, maxSize int, poolMaxSize int) *ByteBufferPool {
    bbp := &ByteBufferPool{
        initialSize: initialSize,
        maxSize:     maxSize,
    }

    bbp.pool = NewHighPerformancePool(
        func() interface{} {
            return bytes.NewBuffer(make([]byte, 0, bbp.initialSize))
        },
        func(obj interface{}) {
            if buf, ok := obj.(*bytes.Buffer); ok {
                buf.Reset()
                // 如果缓冲区太大，重新创建
                if buf.Cap() > bbp.maxSize {
                    buf = bytes.NewBuffer(make([]byte, 0, bbp.initialSize))
                }
            }
        },
        poolMaxSize,
    )

    return bbp
}

// 获取字节缓冲
func (bbp *ByteBufferPool) Get() *bytes.Buffer {
    return bbp.pool.Get().(*bytes.Buffer)
}

// 归还字节缓冲
func (bbp *ByteBufferPool) Put(buf *bytes.Buffer) {
    bbp.pool.Put(buf)
}

// HTTP响应池
type HTTPResponsePool struct {
    pool *HighPerformancePool
}

// HTTP响应对象
type HTTPResponse struct {
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Body       []byte            `json:"body"`
    Timestamp  time.Time         `json:"timestamp"`
}

// 创建HTTP响应池
func NewHTTPResponsePool(maxSize int) *HTTPResponsePool {
    hrp := &HTTPResponsePool{}

    hrp.pool = NewHighPerformancePool(
        func() interface{} {
            return &HTTPResponse{
                Headers: make(map[string]string),
                Body:    make([]byte, 0, 1024),
            }
        },
        func(obj interface{}) {
            if resp, ok := obj.(*HTTPResponse); ok {
                resp.StatusCode = 0
                // 清空headers但保留容量
                for k := range resp.Headers {
                    delete(resp.Headers, k)
                }
                resp.Body = resp.Body[:0] // 保留容量
                resp.Timestamp = time.Time{}
            }
        },
        maxSize,
    )

    return hrp
}

// 获取HTTP响应对象
func (hrp *HTTPResponsePool) Get() *HTTPResponse {
    return hrp.pool.Get().(*HTTPResponse)
}

// 归还HTTP响应对象
func (hrp *HTTPResponsePool) Put(resp *HTTPResponse) {
    hrp.pool.Put(resp)
}

// 连接池管理器
type ConnectionPool struct {
    pool        chan interface{}
    factory     func() (interface{}, error)
    close       func(interface{}) error
    ping        func(interface{}) error
    maxIdle     int
    maxActive   int
    idleTimeout time.Duration
    active      int64
    stats       *PoolStats
    mutex       sync.RWMutex
}

// 连接池配置
type ConnectionPoolConfig struct {
    MaxIdle     int           `json:"max_idle"`
    MaxActive   int           `json:"max_active"`
    IdleTimeout time.Duration `json:"idle_timeout"`
    Factory     func() (interface{}, error)
    Close       func(interface{}) error
    Ping        func(interface{}) error
}

// 创建连接池
func NewConnectionPool(config *ConnectionPoolConfig) *ConnectionPool {
    return &ConnectionPool{
        pool:        make(chan interface{}, config.MaxIdle),
        factory:     config.Factory,
        close:       config.Close,
        ping:        config.Ping,
        maxIdle:     config.MaxIdle,
        maxActive:   config.MaxActive,
        idleTimeout: config.IdleTimeout,
        stats: &PoolStats{
            MaxSize:   config.MaxIdle,
            LastReset: time.Now(),
        },
    }
}

// 获取连接
func (cp *ConnectionPool) Get() (interface{}, error) {
    atomic.AddInt64(&cp.stats.Gets, 1)

    // 尝试从池中获取连接
    select {
    case conn := <-cp.pool:
        // 检查连接是否有效
        if cp.ping != nil && cp.ping(conn) != nil {
            cp.close(conn)
            atomic.AddInt64(&cp.stats.Misses, 1)
            return cp.createConnection()
        }
        atomic.AddInt64(&cp.stats.Hits, 1)
        return conn, nil
    default:
        // 池中没有可用连接，创建新连接
        atomic.AddInt64(&cp.stats.Misses, 1)
        return cp.createConnection()
    }
}

// 创建新连接
func (cp *ConnectionPool) createConnection() (interface{}, error) {
    // 检查活跃连接数限制
    if int(atomic.LoadInt64(&cp.active)) >= cp.maxActive {
        return nil, fmt.Errorf("connection pool exhausted")
    }

    conn, err := cp.factory()
    if err != nil {
        return nil, err
    }

    atomic.AddInt64(&cp.active, 1)
    return conn, nil
}

// 归还连接
func (cp *ConnectionPool) Put(conn interface{}) {
    if conn == nil {
        return
    }

    atomic.AddInt64(&cp.stats.Puts, 1)

    // 尝试将连接放回池中
    select {
    case cp.pool <- conn:
        // 成功放回池中
    default:
        // 池已满，关闭连接
        cp.close(conn)
        atomic.AddInt64(&cp.active, -1)
    }
}

// 关闭连接池
func (cp *ConnectionPool) Close() error {
    close(cp.pool)

    // 关闭所有池中的连接
    for conn := range cp.pool {
        cp.close(conn)
        atomic.AddInt64(&cp.active, -1)
    }

    return nil
}

// 池管理器
type PoolManager struct {
    pools map[string]ObjectPool
    mutex sync.RWMutex
}

// 创建池管理器
func NewPoolManager() *PoolManager {
    return &PoolManager{
        pools: make(map[string]ObjectPool),
    }
}

// 注册池
func (pm *PoolManager) RegisterPool(name string, pool ObjectPool) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    pm.pools[name] = pool
}

// 获取池
func (pm *PoolManager) GetPool(name string) ObjectPool {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    return pm.pools[name]
}

// 获取所有池的统计信息
func (pm *PoolManager) GetAllStats() map[string]*PoolStats {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    stats := make(map[string]*PoolStats)
    for name, pool := range pm.pools {
        stats[name] = pool.Stats()
    }
    return stats
}

// 清空所有池
func (pm *PoolManager) ClearAll() {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()

    for _, pool := range pm.pools {
        pool.Clear()
    }
}
```

---

## ⚡ 并发性能调优

### Goroutine池优化

```go
// Goroutine池管理器
package goroutine

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    "sync/atomic"
    "time"
)

// 工作任务接口
type Task interface {
    Execute(ctx context.Context) error
}

// 函数任务
type FuncTask struct {
    Fn func(ctx context.Context) error
}

func (ft *FuncTask) Execute(ctx context.Context) error {
    return ft.Fn(ctx)
}

// Goroutine池
type GoroutinePool struct {
    // 配置参数
    minWorkers    int
    maxWorkers    int
    maxIdleTime   time.Duration
    taskQueueSize int

    // 运行时状态
    workers       int64
    activeWorkers int64
    taskQueue     chan Task
    workerQueue   chan chan Task
    quit          chan bool
    wg            sync.WaitGroup

    // 统计信息
    stats         *PoolStats

    // 控制
    mutex         sync.RWMutex
    started       bool
}

// 池统计信息
type PoolStats struct {
    Workers       int64     `json:"workers"`
    ActiveWorkers int64     `json:"active_workers"`
    QueuedTasks   int       `json:"queued_tasks"`
    CompletedTasks int64    `json:"completed_tasks"`
    FailedTasks   int64     `json:"failed_tasks"`
    TotalTasks    int64     `json:"total_tasks"`
    AvgTaskTime   time.Duration `json:"avg_task_time"`
    LastTaskTime  time.Time `json:"last_task_time"`
}

// 池配置
type PoolConfig struct {
    MinWorkers    int           `json:"min_workers"`
    MaxWorkers    int           `json:"max_workers"`
    MaxIdleTime   time.Duration `json:"max_idle_time"`
    TaskQueueSize int           `json:"task_queue_size"`
}

// 创建Goroutine池
func NewGoroutinePool(config *PoolConfig) *GoroutinePool {
    if config.MinWorkers <= 0 {
        config.MinWorkers = 1
    }
    if config.MaxWorkers <= 0 {
        config.MaxWorkers = runtime.NumCPU() * 2
    }
    if config.MaxIdleTime <= 0 {
        config.MaxIdleTime = 30 * time.Second
    }
    if config.TaskQueueSize <= 0 {
        config.TaskQueueSize = 1000
    }

    return &GoroutinePool{
        minWorkers:    config.MinWorkers,
        maxWorkers:    config.MaxWorkers,
        maxIdleTime:   config.MaxIdleTime,
        taskQueueSize: config.TaskQueueSize,
        taskQueue:     make(chan Task, config.TaskQueueSize),
        workerQueue:   make(chan chan Task, config.MaxWorkers),
        quit:          make(chan bool),
        stats:         &PoolStats{},
    }
}

// 启动池
func (gp *GoroutinePool) Start() error {
    gp.mutex.Lock()
    defer gp.mutex.Unlock()

    if gp.started {
        return fmt.Errorf("pool already started")
    }

    // 启动最小数量的worker
    for i := 0; i < gp.minWorkers; i++ {
        gp.startWorker()
    }

    // 启动调度器
    go gp.dispatcher()

    // 启动监控器
    go gp.monitor()

    gp.started = true
    return nil
}

// 停止池
func (gp *GoroutinePool) Stop() error {
    gp.mutex.Lock()
    defer gp.mutex.Unlock()

    if !gp.started {
        return fmt.Errorf("pool not started")
    }

    close(gp.quit)
    gp.wg.Wait()

    gp.started = false
    return nil
}

// 提交任务
func (gp *GoroutinePool) Submit(task Task) error {
    if !gp.started {
        return fmt.Errorf("pool not started")
    }

    select {
    case gp.taskQueue <- task:
        atomic.AddInt64(&gp.stats.TotalTasks, 1)
        return nil
    default:
        return fmt.Errorf("task queue full")
    }
}

// 提交函数任务
func (gp *GoroutinePool) SubmitFunc(fn func(ctx context.Context) error) error {
    return gp.Submit(&FuncTask{Fn: fn})
}

// 调度器
func (gp *GoroutinePool) dispatcher() {
    for {
        select {
        case task := <-gp.taskQueue:
            // 尝试获取可用worker
            select {
            case workerTaskQueue := <-gp.workerQueue:
                // 有可用worker，分配任务
                workerTaskQueue <- task
            default:
                // 没有可用worker，尝试创建新worker
                if atomic.LoadInt64(&gp.workers) < int64(gp.maxWorkers) {
                    gp.startWorker()
                    // 重新尝试分配任务
                    select {
                    case workerTaskQueue := <-gp.workerQueue:
                        workerTaskQueue <- task
                    case <-time.After(100 * time.Millisecond):
                        // 超时，将任务放回队列
                        select {
                        case gp.taskQueue <- task:
                        default:
                            // 队列满，丢弃任务
                            atomic.AddInt64(&gp.stats.FailedTasks, 1)
                        }
                    }
                } else {
                    // 已达到最大worker数，等待可用worker
                    select {
                    case workerTaskQueue := <-gp.workerQueue:
                        workerTaskQueue <- task
                    case <-time.After(1 * time.Second):
                        // 超时，将任务放回队列
                        select {
                        case gp.taskQueue <- task:
                        default:
                            atomic.AddInt64(&gp.stats.FailedTasks, 1)
                        }
                    }
                }
            }
        case <-gp.quit:
            return
        }
    }
}

// 启动worker
func (gp *GoroutinePool) startWorker() {
    atomic.AddInt64(&gp.workers, 1)
    gp.wg.Add(1)

    go func() {
        defer func() {
            atomic.AddInt64(&gp.workers, -1)
            gp.wg.Done()
        }()

        taskQueue := make(chan Task)
        idleTimer := time.NewTimer(gp.maxIdleTime)

        for {
            // 将worker注册到worker队列
            select {
            case gp.workerQueue <- taskQueue:
            case <-gp.quit:
                return
            }

            // 等待任务或超时
            select {
            case task := <-taskQueue:
                // 重置空闲计时器
                if !idleTimer.Stop() {
                    <-idleTimer.C
                }
                idleTimer.Reset(gp.maxIdleTime)

                // 执行任务
                gp.executeTask(task)

            case <-idleTimer.C:
                // 空闲超时，检查是否可以退出
                if atomic.LoadInt64(&gp.workers) > int64(gp.minWorkers) {
                    return
                }
                idleTimer.Reset(gp.maxIdleTime)

            case <-gp.quit:
                return
            }
        }
    }()
}

// 执行任务
func (gp *GoroutinePool) executeTask(task Task) {
    start := time.Now()
    atomic.AddInt64(&gp.activeWorkers, 1)

    defer func() {
        atomic.AddInt64(&gp.activeWorkers, -1)
        duration := time.Since(start)

        // 更新统计信息
        atomic.AddInt64(&gp.stats.CompletedTasks, 1)
        gp.stats.LastTaskTime = time.Now()

        // 计算平均任务时间
        completed := atomic.LoadInt64(&gp.stats.CompletedTasks)
        if completed > 0 {
            gp.stats.AvgTaskTime = time.Duration(int64(gp.stats.AvgTaskTime)*completed + int64(duration)) / time.Duration(completed+1)
        }

        // 恢复panic
        if r := recover(); r != nil {
            atomic.AddInt64(&gp.stats.FailedTasks, 1)
            fmt.Printf("Task panic recovered: %v\n", r)
        }
    }()

    // 执行任务
    ctx := context.Background()
    if err := task.Execute(ctx); err != nil {
        atomic.AddInt64(&gp.stats.FailedTasks, 1)
        fmt.Printf("Task execution failed: %v\n", err)
    }
}

// 监控器
func (gp *GoroutinePool) monitor() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            gp.updateStats()
            gp.adjustWorkers()
        case <-gp.quit:
            return
        }
    }
}

// 更新统计信息
func (gp *GoroutinePool) updateStats() {
    gp.stats.Workers = atomic.LoadInt64(&gp.workers)
    gp.stats.ActiveWorkers = atomic.LoadInt64(&gp.activeWorkers)
    gp.stats.QueuedTasks = len(gp.taskQueue)
}

// 动态调整worker数量
func (gp *GoroutinePool) adjustWorkers() {
    queuedTasks := len(gp.taskQueue)
    workers := atomic.LoadInt64(&gp.workers)
    activeWorkers := atomic.LoadInt64(&gp.activeWorkers)

    // 如果队列中有很多任务且活跃worker比例高，增加worker
    if queuedTasks > gp.taskQueueSize/2 && float64(activeWorkers)/float64(workers) > 0.8 {
        if workers < int64(gp.maxWorkers) {
            gp.startWorker()
            fmt.Printf("Increased workers to %d due to high load\n", workers+1)
        }
    }

    // 如果队列空且活跃worker比例低，可能需要减少worker（由空闲超时处理）
}

// 获取统计信息
func (gp *GoroutinePool) Stats() *PoolStats {
    gp.updateStats()
    return gp.stats
}

// 获取池状态
func (gp *GoroutinePool) Status() map[string]interface{} {
    stats := gp.Stats()
    return map[string]interface{}{
        "workers":         stats.Workers,
        "active_workers":  stats.ActiveWorkers,
        "queued_tasks":    stats.QueuedTasks,
        "completed_tasks": stats.CompletedTasks,
        "failed_tasks":    stats.FailedTasks,
        "total_tasks":     stats.TotalTasks,
        "avg_task_time":   stats.AvgTaskTime,
        "last_task_time":  stats.LastTaskTime,
        "utilization":     float64(stats.ActiveWorkers) / float64(stats.Workers) * 100,
        "queue_usage":     float64(stats.QueuedTasks) / float64(gp.taskQueueSize) * 100,
    }
}
```

### Channel优化策略

```go
// Channel优化管理器
package channel

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// 高性能Channel包装器
type OptimizedChannel struct {
    ch          chan interface{}
    capacity    int
    sendCount   int64
    recvCount   int64
    blockCount  int64
    closeOnce   sync.Once
    closed      int32
    stats       *ChannelStats
}

// Channel统计信息
type ChannelStats struct {
    Capacity      int           `json:"capacity"`
    Length        int           `json:"length"`
    SendCount     int64         `json:"send_count"`
    RecvCount     int64         `json:"recv_count"`
    BlockCount    int64         `json:"block_count"`
    Throughput    float64       `json:"throughput"`    // 消息/秒
    AvgLatency    time.Duration `json:"avg_latency"`
    LastActivity  time.Time     `json:"last_activity"`
}

// 创建优化的Channel
func NewOptimizedChannel(capacity int) *OptimizedChannel {
    return &OptimizedChannel{
        ch:       make(chan interface{}, capacity),
        capacity: capacity,
        stats:    &ChannelStats{Capacity: capacity},
    }
}

// 非阻塞发送
func (oc *OptimizedChannel) TrySend(data interface{}) bool {
    if atomic.LoadInt32(&oc.closed) == 1 {
        return false
    }

    select {
    case oc.ch <- data:
        atomic.AddInt64(&oc.sendCount, 1)
        oc.stats.LastActivity = time.Now()
        return true
    default:
        atomic.AddInt64(&oc.blockCount, 1)
        return false
    }
}

// 带超时的发送
func (oc *OptimizedChannel) SendWithTimeout(data interface{}, timeout time.Duration) error {
    if atomic.LoadInt32(&oc.closed) == 1 {
        return fmt.Errorf("channel closed")
    }

    select {
    case oc.ch <- data:
        atomic.AddInt64(&oc.sendCount, 1)
        oc.stats.LastActivity = time.Now()
        return nil
    case <-time.After(timeout):
        atomic.AddInt64(&oc.blockCount, 1)
        return fmt.Errorf("send timeout")
    }
}

// 非阻塞接收
func (oc *OptimizedChannel) TryRecv() (interface{}, bool) {
    select {
    case data := <-oc.ch:
        atomic.AddInt64(&oc.recvCount, 1)
        oc.stats.LastActivity = time.Now()
        return data, true
    default:
        return nil, false
    }
}

// 带超时的接收
func (oc *OptimizedChannel) RecvWithTimeout(timeout time.Duration) (interface{}, error) {
    select {
    case data := <-oc.ch:
        atomic.AddInt64(&oc.recvCount, 1)
        oc.stats.LastActivity = time.Now()
        return data, nil
    case <-time.After(timeout):
        return nil, fmt.Errorf("recv timeout")
    }
}

// 关闭Channel
func (oc *OptimizedChannel) Close() {
    oc.closeOnce.Do(func() {
        atomic.StoreInt32(&oc.closed, 1)
        close(oc.ch)
    })
}

// 获取统计信息
func (oc *OptimizedChannel) Stats() *ChannelStats {
    stats := *oc.stats
    stats.Length = len(oc.ch)
    stats.SendCount = atomic.LoadInt64(&oc.sendCount)
    stats.RecvCount = atomic.LoadInt64(&oc.recvCount)
    stats.BlockCount = atomic.LoadInt64(&oc.blockCount)

    // 计算吞吐量
    if !stats.LastActivity.IsZero() {
        duration := time.Since(stats.LastActivity)
        if duration > 0 {
            stats.Throughput = float64(stats.SendCount+stats.RecvCount) / duration.Seconds()
        }
    }

    return &stats
}

// 批量Channel处理器
type BatchChannelProcessor struct {
    input       chan interface{}
    output      chan []interface{}
    batchSize   int
    flushTime   time.Duration
    processor   func([]interface{}) []interface{}
    ctx         context.Context
    cancel      context.CancelFunc
    wg          sync.WaitGroup
}

// 创建批量处理器
func NewBatchChannelProcessor(batchSize int, flushTime time.Duration, processor func([]interface{}) []interface{}) *BatchChannelProcessor {
    ctx, cancel := context.WithCancel(context.Background())

    return &BatchChannelProcessor{
        input:     make(chan interface{}, batchSize*2),
        output:    make(chan []interface{}, 10),
        batchSize: batchSize,
        flushTime: flushTime,
        processor: processor,
        ctx:       ctx,
        cancel:    cancel,
    }
}

// 启动批量处理
func (bcp *BatchChannelProcessor) Start() {
    bcp.wg.Add(1)
    go bcp.processBatches()
}

// 停止批量处理
func (bcp *BatchChannelProcessor) Stop() {
    bcp.cancel()
    bcp.wg.Wait()
    close(bcp.input)
    close(bcp.output)
}

// 发送数据
func (bcp *BatchChannelProcessor) Send(data interface{}) error {
    select {
    case bcp.input <- data:
        return nil
    case <-bcp.ctx.Done():
        return bcp.ctx.Err()
    }
}

// 接收处理结果
func (bcp *BatchChannelProcessor) Recv() ([]interface{}, error) {
    select {
    case result := <-bcp.output:
        return result, nil
    case <-bcp.ctx.Done():
        return nil, bcp.ctx.Err()
    }
}

// 批量处理逻辑
func (bcp *BatchChannelProcessor) processBatches() {
    defer bcp.wg.Done()

    batch := make([]interface{}, 0, bcp.batchSize)
    flushTimer := time.NewTimer(bcp.flushTime)

    for {
        select {
        case data := <-bcp.input:
            batch = append(batch, data)

            // 检查是否达到批量大小
            if len(batch) >= bcp.batchSize {
                bcp.processBatch(batch)
                batch = batch[:0] // 重置切片

                // 重置定时器
                if !flushTimer.Stop() {
                    <-flushTimer.C
                }
                flushTimer.Reset(bcp.flushTime)
            }

        case <-flushTimer.C:
            // 定时刷新
            if len(batch) > 0 {
                bcp.processBatch(batch)
                batch = batch[:0]
            }
            flushTimer.Reset(bcp.flushTime)

        case <-bcp.ctx.Done():
            // 处理剩余数据
            if len(batch) > 0 {
                bcp.processBatch(batch)
            }
            return
        }
    }
}

// 处理单个批次
func (bcp *BatchChannelProcessor) processBatch(batch []interface{}) {
    if bcp.processor != nil {
        result := bcp.processor(batch)
        if result != nil {
            select {
            case bcp.output <- result:
            case <-bcp.ctx.Done():
                return
            }
        }
    }
}
```

### 锁优化策略

```go
// 锁优化管理器
package locks

import (
    "runtime"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

// 读写锁统计信息
type RWMutexStats struct {
    ReadLocks    int64         `json:"read_locks"`
    WriteLocks   int64         `json:"write_locks"`
    ReadWaits    int64         `json:"read_waits"`
    WriteWaits   int64         `json:"write_waits"`
    AvgReadTime  time.Duration `json:"avg_read_time"`
    AvgWriteTime time.Duration `json:"avg_write_time"`
    Contention   float64       `json:"contention"`
}

// 优化的读写锁
type OptimizedRWMutex struct {
    mu           sync.RWMutex
    stats        *RWMutexStats
    readCount    int64
    writeCount   int64
    readWaits    int64
    writeWaits   int64
    readTime     int64  // 纳秒
    writeTime    int64  // 纳秒
}

// 创建优化的读写锁
func NewOptimizedRWMutex() *OptimizedRWMutex {
    return &OptimizedRWMutex{
        stats: &RWMutexStats{},
    }
}

// 读锁
func (orm *OptimizedRWMutex) RLock() {
    start := time.Now()

    // 尝试快速获取读锁
    if orm.tryRLock() {
        atomic.AddInt64(&orm.readCount, 1)
        return
    }

    // 快速获取失败，记录等待
    atomic.AddInt64(&orm.readWaits, 1)
    orm.mu.RLock()

    duration := time.Since(start)
    atomic.AddInt64(&orm.readCount, 1)
    atomic.AddInt64(&orm.readTime, int64(duration))
}

// 读解锁
func (orm *OptimizedRWMutex) RUnlock() {
    orm.mu.RUnlock()
}

// 写锁
func (orm *OptimizedRWMutex) Lock() {
    start := time.Now()

    // 尝试快速获取写锁
    if orm.tryLock() {
        atomic.AddInt64(&orm.writeCount, 1)
        return
    }

    // 快速获取失败，记录等待
    atomic.AddInt64(&orm.writeWaits, 1)
    orm.mu.Lock()

    duration := time.Since(start)
    atomic.AddInt64(&orm.writeCount, 1)
    atomic.AddInt64(&orm.writeTime, int64(duration))
}

// 写解锁
func (orm *OptimizedRWMutex) Unlock() {
    orm.mu.Unlock()
}

// 尝试获取读锁（非阻塞）
func (orm *OptimizedRWMutex) tryRLock() bool {
    // 使用unsafe操作尝试快速路径
    // 注意：这是一个简化的实现，实际情况更复杂
    return false // 简化实现，总是返回false
}

// 尝试获取写锁（非阻塞）
func (orm *OptimizedRWMutex) tryLock() bool {
    // 使用unsafe操作尝试快速路径
    return false // 简化实现，总是返回false
}

// 获取统计信息
func (orm *OptimizedRWMutex) Stats() *RWMutexStats {
    readCount := atomic.LoadInt64(&orm.readCount)
    writeCount := atomic.LoadInt64(&orm.writeCount)
    readWaits := atomic.LoadInt64(&orm.readWaits)
    writeWaits := atomic.LoadInt64(&orm.writeWaits)
    readTime := atomic.LoadInt64(&orm.readTime)
    writeTime := atomic.LoadInt64(&orm.writeTime)

    stats := &RWMutexStats{
        ReadLocks:  readCount,
        WriteLocks: writeCount,
        ReadWaits:  readWaits,
        WriteWaits: writeWaits,
    }

    // 计算平均时间
    if readCount > 0 {
        stats.AvgReadTime = time.Duration(readTime / readCount)
    }
    if writeCount > 0 {
        stats.AvgWriteTime = time.Duration(writeTime / writeCount)
    }

    // 计算竞争率
    totalOps := readCount + writeCount
    totalWaits := readWaits + writeWaits
    if totalOps > 0 {
        stats.Contention = float64(totalWaits) / float64(totalOps) * 100
    }

    return stats
}

// 自适应自旋锁
type AdaptiveSpinLock struct {
    state       int32
    spinCount   int32
    maxSpin     int32
    backoff     time.Duration
}

// 创建自适应自旋锁
func NewAdaptiveSpinLock() *AdaptiveSpinLock {
    return &AdaptiveSpinLock{
        maxSpin: int32(runtime.NumCPU() * 4),
        backoff: time.Microsecond,
    }
}

// 获取锁
func (asl *AdaptiveSpinLock) Lock() {
    for {
        // 尝试原子获取锁
        if atomic.CompareAndSwapInt32(&asl.state, 0, 1) {
            return
        }

        // 自适应自旋
        spinCount := atomic.LoadInt32(&asl.spinCount)
        if spinCount < asl.maxSpin {
            // 自旋等待
            for i := int32(0); i < spinCount; i++ {
                if atomic.LoadInt32(&asl.state) == 0 {
                    break
                }
                runtime.Gosched() // 让出CPU
            }

            // 增加自旋次数
            atomic.AddInt32(&asl.spinCount, 1)
        } else {
            // 自旋次数过多，使用退避策略
            time.Sleep(asl.backoff)
            asl.backoff *= 2
            if asl.backoff > time.Millisecond {
                asl.backoff = time.Millisecond
            }

            // 重置自旋计数
            atomic.StoreInt32(&asl.spinCount, 0)
        }
    }
}

// 释放锁
func (asl *AdaptiveSpinLock) Unlock() {
    atomic.StoreInt32(&asl.state, 0)

    // 重置退避时间
    asl.backoff = time.Microsecond
}

// 无锁队列
type LockFreeQueue struct {
    head unsafe.Pointer
    tail unsafe.Pointer
}

type queueNode struct {
    data interface{}
    next unsafe.Pointer
}

// 创建无锁队列
func NewLockFreeQueue() *LockFreeQueue {
    node := &queueNode{}
    return &LockFreeQueue{
        head: unsafe.Pointer(node),
        tail: unsafe.Pointer(node),
    }
}

// 入队
func (lfq *LockFreeQueue) Enqueue(data interface{}) {
    node := &queueNode{data: data}

    for {
        tail := (*queueNode)(atomic.LoadPointer(&lfq.tail))
        next := (*queueNode)(atomic.LoadPointer(&tail.next))

        if tail == (*queueNode)(atomic.LoadPointer(&lfq.tail)) {
            if next == nil {
                if atomic.CompareAndSwapPointer(&tail.next, unsafe.Pointer(next), unsafe.Pointer(node)) {
                    break
                }
            } else {
                atomic.CompareAndSwapPointer(&lfq.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
            }
        }
    }

    atomic.CompareAndSwapPointer(&lfq.tail, unsafe.Pointer((*queueNode)(atomic.LoadPointer(&lfq.tail))), unsafe.Pointer(node))
}

// 出队
func (lfq *LockFreeQueue) Dequeue() (interface{}, bool) {
    for {
        head := (*queueNode)(atomic.LoadPointer(&lfq.head))
        tail := (*queueNode)(atomic.LoadPointer(&lfq.tail))
        next := (*queueNode)(atomic.LoadPointer(&head.next))

        if head == (*queueNode)(atomic.LoadPointer(&lfq.head)) {
            if head == tail {
                if next == nil {
                    return nil, false // 队列为空
                }
                atomic.CompareAndSwapPointer(&lfq.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
            } else {
                if next == nil {
                    continue
                }

                data := next.data
                if atomic.CompareAndSwapPointer(&lfq.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
                    return data, true
                }
            }
        }
    }
}

// 锁管理器
type LockManager struct {
    locks map[string]*OptimizedRWMutex
    mutex sync.RWMutex
}

// 创建锁管理器
func NewLockManager() *LockManager {
    return &LockManager{
        locks: make(map[string]*OptimizedRWMutex),
    }
}

// 获取锁
func (lm *LockManager) GetLock(name string) *OptimizedRWMutex {
    lm.mutex.RLock()
    lock, exists := lm.locks[name]
    lm.mutex.RUnlock()

    if exists {
        return lock
    }

    lm.mutex.Lock()
    defer lm.mutex.Unlock()

    // 双重检查
    if lock, exists := lm.locks[name]; exists {
        return lock
    }

    lock = NewOptimizedRWMutex()
    lm.locks[name] = lock
    return lock
}

// 获取所有锁的统计信息
func (lm *LockManager) GetAllStats() map[string]*RWMutexStats {
    lm.mutex.RLock()
    defer lm.mutex.RUnlock()

    stats := make(map[string]*RWMutexStats)
    for name, lock := range lm.locks {
        stats[name] = lock.Stats()
    }
    return stats
}
```

---

## 🗄️ 数据库性能优化

### 连接池优化

```go
// 数据库连接池优化
package database

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
    "sync/atomic"
    "time"

    "gorm.io/gorm"
)

// 连接池配置
type ConnectionPoolConfig struct {
    MaxOpenConns    int           `json:"max_open_conns"`
    MaxIdleConns    int           `json:"max_idle_conns"`
    ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`

    // 高级配置
    ReadTimeout     time.Duration `json:"read_timeout"`
    WriteTimeout    time.Duration `json:"write_timeout"`
    ConnectTimeout  time.Duration `json:"connect_timeout"`

    // 监控配置
    EnableMetrics   bool          `json:"enable_metrics"`
    MetricsInterval time.Duration `json:"metrics_interval"`
}

// 连接池统计信息
type ConnectionPoolStats struct {
    OpenConnections int           `json:"open_connections"`
    InUseConns      int           `json:"in_use_conns"`
    IdleConns       int           `json:"idle_conns"`
    WaitCount       int64         `json:"wait_count"`
    WaitDuration    time.Duration `json:"wait_duration"`
    MaxIdleClosed   int64         `json:"max_idle_closed"`
    MaxLifetimeClosed int64       `json:"max_lifetime_closed"`

    // 性能指标
    AvgConnTime     time.Duration `json:"avg_conn_time"`
    AvgQueryTime    time.Duration `json:"avg_query_time"`
    QueryCount      int64         `json:"query_count"`
    ErrorCount      int64         `json:"error_count"`

    // 健康状态
    HealthScore     float64       `json:"health_score"`
    LastCheck       time.Time     `json:"last_check"`
}

// 优化的数据库管理器
type OptimizedDBManager struct {
    db              *gorm.DB
    config          *ConnectionPoolConfig
    stats           *ConnectionPoolStats
    queryStats      map[string]*QueryStats
    queryStatsMutex sync.RWMutex

    // 监控
    monitoring      bool
    stopMonitoring  chan bool
    wg              sync.WaitGroup
}

// 查询统计信息
type QueryStats struct {
    Count       int64         `json:"count"`
    TotalTime   time.Duration `json:"total_time"`
    AvgTime     time.Duration `json:"avg_time"`
    MinTime     time.Duration `json:"min_time"`
    MaxTime     time.Duration `json:"max_time"`
    ErrorCount  int64         `json:"error_count"`
    LastExecuted time.Time    `json:"last_executed"`
}

// 创建优化的数据库管理器
func NewOptimizedDBManager(db *gorm.DB, config *ConnectionPoolConfig) *OptimizedDBManager {
    manager := &OptimizedDBManager{
        db:             db,
        config:         config,
        stats:          &ConnectionPoolStats{},
        queryStats:     make(map[string]*QueryStats),
        stopMonitoring: make(chan bool),
    }

    // 配置连接池
    manager.configureConnectionPool()

    // 启动监控
    if config.EnableMetrics {
        manager.startMonitoring()
    }

    return manager
}

// 配置连接池
func (odm *OptimizedDBManager) configureConnectionPool() {
    sqlDB, err := odm.db.DB()
    if err != nil {
        fmt.Printf("Failed to get sql.DB: %v\n", err)
        return
    }

    // 设置连接池参数
    sqlDB.SetMaxOpenConns(odm.config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(odm.config.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(odm.config.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(odm.config.ConnMaxIdleTime)

    fmt.Printf("Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v\n",
        odm.config.MaxOpenConns, odm.config.MaxIdleConns, odm.config.ConnMaxLifetime)
}

// 启动监控
func (odm *OptimizedDBManager) startMonitoring() {
    odm.monitoring = true
    odm.wg.Add(1)

    go func() {
        defer odm.wg.Done()
        ticker := time.NewTicker(odm.config.MetricsInterval)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                odm.collectStats()
            case <-odm.stopMonitoring:
                return
            }
        }
    }()
}

// 停止监控
func (odm *OptimizedDBManager) StopMonitoring() {
    if odm.monitoring {
        close(odm.stopMonitoring)
        odm.wg.Wait()
        odm.monitoring = false
    }
}

// 收集统计信息
func (odm *OptimizedDBManager) collectStats() {
    sqlDB, err := odm.db.DB()
    if err != nil {
        return
    }

    dbStats := sqlDB.Stats()

    odm.stats.OpenConnections = dbStats.OpenConnections
    odm.stats.InUseConns = dbStats.InUse
    odm.stats.IdleConns = dbStats.Idle
    odm.stats.WaitCount = dbStats.WaitCount
    odm.stats.WaitDuration = dbStats.WaitDuration
    odm.stats.MaxIdleClosed = dbStats.MaxIdleClosed
    odm.stats.MaxLifetimeClosed = dbStats.MaxLifetimeClosed
    odm.stats.LastCheck = time.Now()

    // 计算健康分数
    odm.calculateHealthScore()
}

// 计算健康分数
func (odm *OptimizedDBManager) calculateHealthScore() {
    score := 100.0

    // 连接使用率
    if odm.config.MaxOpenConns > 0 {
        usage := float64(odm.stats.InUseConns) / float64(odm.config.MaxOpenConns)
        if usage > 0.9 {
            score -= 20 // 使用率过高扣分
        } else if usage > 0.7 {
            score -= 10
        }
    }

    // 等待时间
    if odm.stats.WaitDuration > 100*time.Millisecond {
        score -= 15 // 等待时间过长扣分
    }

    // 错误率
    if odm.stats.QueryCount > 0 {
        errorRate := float64(odm.stats.ErrorCount) / float64(odm.stats.QueryCount)
        if errorRate > 0.05 {
            score -= 25 // 错误率过高扣分
        } else if errorRate > 0.01 {
            score -= 10
        }
    }

    if score < 0 {
        score = 0
    }

    odm.stats.HealthScore = score
}

// 执行查询并记录统计信息
func (odm *OptimizedDBManager) ExecuteQuery(query string, args ...interface{}) *gorm.DB {
    start := time.Now()

    // 执行查询
    result := odm.db.Raw(query, args...)

    duration := time.Since(start)

    // 记录统计信息
    odm.recordQueryStats(query, duration, result.Error)

    return result
}

// 记录查询统计信息
func (odm *OptimizedDBManager) recordQueryStats(query string, duration time.Duration, err error) {
    // 简化查询字符串作为key
    queryKey := odm.normalizeQuery(query)

    odm.queryStatsMutex.Lock()
    defer odm.queryStatsMutex.Unlock()

    stats, exists := odm.queryStats[queryKey]
    if !exists {
        stats = &QueryStats{
            MinTime: duration,
            MaxTime: duration,
        }
        odm.queryStats[queryKey] = stats
    }

    // 更新统计信息
    stats.Count++
    stats.TotalTime += duration
    stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)
    stats.LastExecuted = time.Now()

    if duration < stats.MinTime {
        stats.MinTime = duration
    }
    if duration > stats.MaxTime {
        stats.MaxTime = duration
    }

    if err != nil {
        stats.ErrorCount++
        atomic.AddInt64(&odm.stats.ErrorCount, 1)
    }

    atomic.AddInt64(&odm.stats.QueryCount, 1)
}

// 标准化查询字符串
func (odm *OptimizedDBManager) normalizeQuery(query string) string {
    // 简化实现：移除参数，只保留查询结构
    // 实际实现中可能需要更复杂的解析
    if len(query) > 100 {
        return query[:100] + "..."
    }
    return query
}

// 获取连接池统计信息
func (odm *OptimizedDBManager) GetStats() *ConnectionPoolStats {
    return odm.stats
}

// 获取查询统计信息
func (odm *OptimizedDBManager) GetQueryStats() map[string]*QueryStats {
    odm.queryStatsMutex.RLock()
    defer odm.queryStatsMutex.RUnlock()

    // 复制统计信息
    result := make(map[string]*QueryStats)
    for k, v := range odm.queryStats {
        statsCopy := *v
        result[k] = &statsCopy
    }

    return result
}

// 获取慢查询
func (odm *OptimizedDBManager) GetSlowQueries(threshold time.Duration) map[string]*QueryStats {
    odm.queryStatsMutex.RLock()
    defer odm.queryStatsMutex.RUnlock()

    slowQueries := make(map[string]*QueryStats)
    for query, stats := range odm.queryStats {
        if stats.AvgTime > threshold || stats.MaxTime > threshold*2 {
            statsCopy := *stats
            slowQueries[query] = &statsCopy
        }
    }

    return slowQueries
}

// 健康检查
func (odm *OptimizedDBManager) HealthCheck() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    sqlDB, err := odm.db.DB()
    if err != nil {
        return fmt.Errorf("failed to get sql.DB: %w", err)
    }

    if err := sqlDB.PingContext(ctx); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }

    // 检查连接池状态
    stats := sqlDB.Stats()
    if stats.OpenConnections == 0 {
        return fmt.Errorf("no open connections")
    }

    return nil
}

// 优化建议
func (odm *OptimizedDBManager) GetOptimizationRecommendations() []string {
    recommendations := make([]string, 0)

    // 基于统计信息提供建议
    if odm.stats.HealthScore < 70 {
        recommendations = append(recommendations, "Database health score is low, consider investigating connection pool settings")
    }

    if odm.stats.WaitDuration > 100*time.Millisecond {
        recommendations = append(recommendations, "High connection wait time detected, consider increasing MaxOpenConns")
    }

    if float64(odm.stats.InUseConns)/float64(odm.config.MaxOpenConns) > 0.8 {
        recommendations = append(recommendations, "Connection pool utilization is high, consider scaling up")
    }

    // 检查慢查询
    slowQueries := odm.GetSlowQueries(1 * time.Second)
    if len(slowQueries) > 0 {
        recommendations = append(recommendations, fmt.Sprintf("Found %d slow queries, consider optimization", len(slowQueries)))
    }

    if odm.stats.ErrorCount > 0 && odm.stats.QueryCount > 0 {
        errorRate := float64(odm.stats.ErrorCount) / float64(odm.stats.QueryCount)
        if errorRate > 0.01 {
            recommendations = append(recommendations, fmt.Sprintf("High error rate detected: %.2f%%", errorRate*100))
        }
    }

    if len(recommendations) == 0 {
        recommendations = append(recommendations, "Database performance is within acceptable ranges")
    }

    return recommendations
}
```

---

## 💾 缓存策略优化

### 多级缓存架构

```go
// 多级缓存管理器
package cache

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"
)

// 缓存层级定义
type CacheLevel int

const (
    L1Cache CacheLevel = iota // 内存缓存
    L2Cache                   // Redis缓存
    L3Cache                   // 分布式缓存
)

// 缓存项
type CacheItem struct {
    Key       string      `json:"key"`
    Value     interface{} `json:"value"`
    TTL       time.Duration `json:"ttl"`
    Level     CacheLevel  `json:"level"`
    CreatedAt time.Time   `json:"created_at"`
    AccessCount int64     `json:"access_count"`
    LastAccess time.Time  `json:"last_access"`
}

// 多级缓存管理器
type MultiLevelCacheManager struct {
    l1Cache    *MemoryCache
    l2Cache    *RedisCache
    l3Cache    *DistributedCache
    config     *CacheConfig
    stats      *CacheStats
    mutex      sync.RWMutex
}

// 缓存配置
type CacheConfig struct {
    L1 struct {
        MaxSize     int           `json:"max_size"`
        TTL         time.Duration `json:"ttl"`
        CleanupInterval time.Duration `json:"cleanup_interval"`
    } `json:"l1"`

    L2 struct {
        Addr        string        `json:"addr"`
        Password    string        `json:"password"`
        DB          int           `json:"db"`
        PoolSize    int           `json:"pool_size"`
        TTL         time.Duration `json:"ttl"`
    } `json:"l2"`

    L3 struct {
        Nodes       []string      `json:"nodes"`
        TTL         time.Duration `json:"ttl"`
    } `json:"l3"`

    // 策略配置
    WriteThrough   bool `json:"write_through"`
    WriteBack      bool `json:"write_back"`
    ReadThrough    bool `json:"read_through"`

    // 性能配置
    PrefetchEnabled bool          `json:"prefetch_enabled"`
    PrefetchRatio   float64       `json:"prefetch_ratio"`
    CompressionEnabled bool       `json:"compression_enabled"`
    SerializationFormat string    `json:"serialization_format"` // json, msgpack, protobuf
}

// 缓存统计信息
type CacheStats struct {
    L1Stats CacheLevelStats `json:"l1_stats"`
    L2Stats CacheLevelStats `json:"l2_stats"`
    L3Stats CacheLevelStats `json:"l3_stats"`

    TotalHits   int64   `json:"total_hits"`
    TotalMisses int64   `json:"total_misses"`
    HitRate     float64 `json:"hit_rate"`

    AvgLatency  time.Duration `json:"avg_latency"`
    Throughput  float64       `json:"throughput"`
}

// 缓存层级统计
type CacheLevelStats struct {
    Hits        int64         `json:"hits"`
    Misses      int64         `json:"misses"`
    Sets        int64         `json:"sets"`
    Deletes     int64         `json:"deletes"`
    Size        int           `json:"size"`
    HitRate     float64       `json:"hit_rate"`
    AvgLatency  time.Duration `json:"avg_latency"`
}

// 创建多级缓存管理器
func NewMultiLevelCacheManager(config *CacheConfig) *MultiLevelCacheManager {
    manager := &MultiLevelCacheManager{
        config: config,
        stats:  &CacheStats{},
    }

    // 初始化各级缓存
    manager.l1Cache = NewMemoryCache(config.L1.MaxSize, config.L1.TTL)
    manager.l2Cache = NewRedisCache(&RedisCacheConfig{
        Addr:     config.L2.Addr,
        Password: config.L2.Password,
        DB:       config.L2.DB,
        PoolSize: config.L2.PoolSize,
        TTL:      config.L2.TTL,
    })

    return manager
}

// 获取缓存值
func (mlcm *MultiLevelCacheManager) Get(ctx context.Context, key string) (interface{}, bool) {
    start := time.Now()
    defer func() {
        latency := time.Since(start)
        mlcm.updateLatencyStats(latency)
    }()

    // L1缓存查找
    if value, found := mlcm.l1Cache.Get(key); found {
        mlcm.stats.L1Stats.Hits++
        mlcm.stats.TotalHits++
        return value, true
    }
    mlcm.stats.L1Stats.Misses++

    // L2缓存查找
    if value, found := mlcm.l2Cache.Get(ctx, key); found {
        mlcm.stats.L2Stats.Hits++
        mlcm.stats.TotalHits++

        // 回写到L1缓存
        mlcm.l1Cache.Set(key, value, mlcm.config.L1.TTL)
        return value, true
    }
    mlcm.stats.L2Stats.Misses++

    // L3缓存查找（如果配置了）
    if mlcm.l3Cache != nil {
        if value, found := mlcm.l3Cache.Get(ctx, key); found {
            mlcm.stats.L3Stats.Hits++
            mlcm.stats.TotalHits++

            // 回写到L1和L2缓存
            mlcm.l1Cache.Set(key, value, mlcm.config.L1.TTL)
            mlcm.l2Cache.Set(ctx, key, value, mlcm.config.L2.TTL)
            return value, true
        }
        mlcm.stats.L3Stats.Misses++
    }

    mlcm.stats.TotalMisses++
    return nil, false
}

// 设置缓存值
func (mlcm *MultiLevelCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // 根据策略决定写入哪些层级
    if mlcm.config.WriteThrough {
        // 写穿策略：同时写入所有层级
        mlcm.l1Cache.Set(key, value, ttl)
        mlcm.stats.L1Stats.Sets++

        if err := mlcm.l2Cache.Set(ctx, key, value, ttl); err != nil {
            return fmt.Errorf("L2 cache set failed: %w", err)
        }
        mlcm.stats.L2Stats.Sets++

        if mlcm.l3Cache != nil {
            if err := mlcm.l3Cache.Set(ctx, key, value, ttl); err != nil {
                return fmt.Errorf("L3 cache set failed: %w", err)
            }
            mlcm.stats.L3Stats.Sets++
        }
    } else {
        // 写回策略：只写入L1缓存
        mlcm.l1Cache.Set(key, value, ttl)
        mlcm.stats.L1Stats.Sets++
    }

    return nil
}

// 删除缓存值
func (mlcm *MultiLevelCacheManager) Delete(ctx context.Context, key string) error {
    // 从所有层级删除
    mlcm.l1Cache.Delete(key)
    mlcm.stats.L1Stats.Deletes++

    if err := mlcm.l2Cache.Delete(ctx, key); err != nil {
        return fmt.Errorf("L2 cache delete failed: %w", err)
    }
    mlcm.stats.L2Stats.Deletes++

    if mlcm.l3Cache != nil {
        if err := mlcm.l3Cache.Delete(ctx, key); err != nil {
            return fmt.Errorf("L3 cache delete failed: %w", err)
        }
        mlcm.stats.L3Stats.Deletes++
    }

    return nil
}

// 更新延迟统计
func (mlcm *MultiLevelCacheManager) updateLatencyStats(latency time.Duration) {
    // 简化的延迟统计更新
    mlcm.mutex.Lock()
    defer mlcm.mutex.Unlock()

    // 使用指数移动平均
    alpha := 0.1
    if mlcm.stats.AvgLatency == 0 {
        mlcm.stats.AvgLatency = latency
    } else {
        mlcm.stats.AvgLatency = time.Duration(float64(mlcm.stats.AvgLatency)*(1-alpha) + float64(latency)*alpha)
    }
}

// 获取统计信息
func (mlcm *MultiLevelCacheManager) GetStats() *CacheStats {
    mlcm.mutex.RLock()
    defer mlcm.mutex.RUnlock()

    stats := *mlcm.stats

    // 计算命中率
    totalOps := stats.TotalHits + stats.TotalMisses
    if totalOps > 0 {
        stats.HitRate = float64(stats.TotalHits) / float64(totalOps) * 100
    }

    // 计算各层级命中率
    l1Ops := stats.L1Stats.Hits + stats.L1Stats.Misses
    if l1Ops > 0 {
        stats.L1Stats.HitRate = float64(stats.L1Stats.Hits) / float64(l1Ops) * 100
    }

    l2Ops := stats.L2Stats.Hits + stats.L2Stats.Misses
    if l2Ops > 0 {
        stats.L2Stats.HitRate = float64(stats.L2Stats.Hits) / float64(l2Ops) * 100
    }

    return &stats
}

// 内存缓存实现
type MemoryCache struct {
    data      map[string]*CacheItem
    maxSize   int
    defaultTTL time.Duration
    mutex     sync.RWMutex
}

func NewMemoryCache(maxSize int, defaultTTL time.Duration) *MemoryCache {
    return &MemoryCache{
        data:       make(map[string]*CacheItem),
        maxSize:    maxSize,
        defaultTTL: defaultTTL,
    }
}

func (mc *MemoryCache) Get(key string) (interface{}, bool) {
    mc.mutex.RLock()
    defer mc.mutex.RUnlock()

    item, exists := mc.data[key]
    if !exists {
        return nil, false
    }

    // 检查是否过期
    if time.Since(item.CreatedAt) > item.TTL {
        delete(mc.data, key)
        return nil, false
    }

    item.AccessCount++
    item.LastAccess = time.Now()
    return item.Value, true
}

func (mc *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()

    // 检查容量限制
    if len(mc.data) >= mc.maxSize {
        mc.evictLRU()
    }

    if ttl == 0 {
        ttl = mc.defaultTTL
    }

    mc.data[key] = &CacheItem{
        Key:       key,
        Value:     value,
        TTL:       ttl,
        CreatedAt: time.Now(),
        LastAccess: time.Now(),
    }
}

func (mc *MemoryCache) Delete(key string) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()
    delete(mc.data, key)
}

// LRU淘汰
func (mc *MemoryCache) evictLRU() {
    var oldestKey string
    var oldestTime time.Time

    for key, item := range mc.data {
        if oldestKey == "" || item.LastAccess.Before(oldestTime) {
            oldestKey = key
            oldestTime = item.LastAccess
        }
    }

    if oldestKey != "" {
        delete(mc.data, oldestKey)
    }
}

// Redis缓存实现
type RedisCache struct {
    client *redis.Client
    config *RedisCacheConfig
}

type RedisCacheConfig struct {
    Addr     string
    Password string
    DB       int
    PoolSize int
    TTL      time.Duration
}

func NewRedisCache(config *RedisCacheConfig) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:     config.Addr,
        Password: config.Password,
        DB:       config.DB,
        PoolSize: config.PoolSize,
    })

    return &RedisCache{
        client: client,
        config: config,
    }
}

func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
    result, err := rc.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, false
        }
        return nil, false
    }

    return result, true
}

func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if ttl == 0 {
        ttl = rc.config.TTL
    }

    return rc.client.Set(ctx, key, value, ttl).Err()
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
    return rc.client.Del(ctx, key).Err()
}

// 分布式缓存实现（简化）
type DistributedCache struct {
    // 实现分布式缓存逻辑
}

func (dc *DistributedCache) Get(ctx context.Context, key string) (interface{}, bool) {
    // 实现分布式缓存获取逻辑
    return nil, false
}

func (dc *DistributedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // 实现分布式缓存设置逻辑
    return nil
}

func (dc *DistributedCache) Delete(ctx context.Context, key string) error {
    // 实现分布式缓存删除逻辑
    return nil
}
```

---

## 🌐 网络性能优化

### HTTP/2和连接复用

```go
// 网络性能优化管理器
package network

import (
    "compress/gzip"
    "context"
    "crypto/tls"
    "fmt"
    "io"
    "net"
    "net/http"
    "sync"
    "time"

    "golang.org/x/net/http2"
)

// 网络优化配置
type NetworkOptimizationConfig struct {
    // HTTP/2配置
    HTTP2 struct {
        Enabled           bool `json:"enabled"`
        MaxConcurrentStreams uint32 `json:"max_concurrent_streams"`
        MaxFrameSize      uint32 `json:"max_frame_size"`
        MaxHeaderListSize uint32 `json:"max_header_list_size"`
        IdleTimeout       time.Duration `json:"idle_timeout"`
    } `json:"http2"`

    // 连接池配置
    ConnectionPool struct {
        MaxIdleConns        int           `json:"max_idle_conns"`
        MaxIdleConnsPerHost int           `json:"max_idle_conns_per_host"`
        MaxConnsPerHost     int           `json:"max_conns_per_host"`
        IdleConnTimeout     time.Duration `json:"idle_conn_timeout"`
        KeepAlive           time.Duration `json:"keep_alive"`
        TLSHandshakeTimeout time.Duration `json:"tls_handshake_timeout"`
    } `json:"connection_pool"`

    // 压缩配置
    Compression struct {
        Enabled bool     `json:"enabled"`
        Level   int      `json:"level"`
        Types   []string `json:"types"`
    } `json:"compression"`

    // 缓存配置
    Cache struct {
        Enabled    bool          `json:"enabled"`
        MaxAge     time.Duration `json:"max_age"`
        ETagEnabled bool         `json:"etag_enabled"`
    } `json:"cache"`
}

// 优化的HTTP客户端
type OptimizedHTTPClient struct {
    client *http.Client
    config *NetworkOptimizationConfig
    stats  *NetworkStats
    mutex  sync.RWMutex
}

// 网络统计信息
type NetworkStats struct {
    RequestCount    int64         `json:"request_count"`
    ResponseTime    time.Duration `json:"response_time"`
    BytesSent       int64         `json:"bytes_sent"`
    BytesReceived   int64         `json:"bytes_received"`
    ConnectionReuse int64         `json:"connection_reuse"`
    CompressionRatio float64      `json:"compression_ratio"`
    ErrorCount      int64         `json:"error_count"`
}

// 创建优化的HTTP客户端
func NewOptimizedHTTPClient(config *NetworkOptimizationConfig) *OptimizedHTTPClient {
    // 创建自定义Transport
    transport := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: config.ConnectionPool.KeepAlive,
        }).DialContext,

        MaxIdleConns:        config.ConnectionPool.MaxIdleConns,
        MaxIdleConnsPerHost: config.ConnectionPool.MaxIdleConnsPerHost,
        MaxConnsPerHost:     config.ConnectionPool.MaxConnsPerHost,
        IdleConnTimeout:     config.ConnectionPool.IdleConnTimeout,
        TLSHandshakeTimeout: config.ConnectionPool.TLSHandshakeTimeout,

        // 启用HTTP/2
        ForceAttemptHTTP2: config.HTTP2.Enabled,

        // TLS配置
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false,
            MinVersion:         tls.VersionTLS12,
        },
    }

    // 配置HTTP/2
    if config.HTTP2.Enabled {
        http2.ConfigureTransport(transport)
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }

    return &OptimizedHTTPClient{
        client: client,
        config: config,
        stats:  &NetworkStats{},
    }
}

// 执行HTTP请求
func (ohc *OptimizedHTTPClient) Do(req *http.Request) (*http.Response, error) {
    start := time.Now()

    // 添加压缩支持
    if ohc.config.Compression.Enabled {
        req.Header.Set("Accept-Encoding", "gzip, deflate")
    }

    // 添加缓存控制
    if ohc.config.Cache.Enabled {
        req.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d", int(ohc.config.Cache.MaxAge.Seconds())))
    }

    // 执行请求
    resp, err := ohc.client.Do(req)
    if err != nil {
        ohc.stats.ErrorCount++
        return nil, err
    }

    // 更新统计信息
    duration := time.Since(start)
    ohc.updateStats(req, resp, duration)

    return resp, nil
}

// 更新统计信息
func (ohc *OptimizedHTTPClient) updateStats(req *http.Request, resp *http.Response, duration time.Duration) {
    ohc.mutex.Lock()
    defer ohc.mutex.Unlock()

    ohc.stats.RequestCount++
    ohc.stats.ResponseTime = duration

    // 计算传输字节数
    if req.ContentLength > 0 {
        ohc.stats.BytesSent += req.ContentLength
    }
    if resp.ContentLength > 0 {
        ohc.stats.BytesReceived += resp.ContentLength
    }

    // 检查连接复用
    if resp.Header.Get("Connection") == "keep-alive" {
        ohc.stats.ConnectionReuse++
    }

    // 检查压缩
    if resp.Header.Get("Content-Encoding") == "gzip" {
        // 简化的压缩比计算
        ohc.stats.CompressionRatio = 0.7 // 假设70%的压缩率
    }
}

// 获取统计信息
func (ohc *OptimizedHTTPClient) GetStats() *NetworkStats {
    ohc.mutex.RLock()
    defer ohc.mutex.RUnlock()

    stats := *ohc.stats
    return &stats
}

// 压缩中间件
func CompressionMiddleware(level int) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 检查客户端是否支持gzip
            if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
                next.ServeHTTP(w, r)
                return
            }

            // 创建gzip writer
            gz, err := gzip.NewWriterLevel(w, level)
            if err != nil {
                next.ServeHTTP(w, r)
                return
            }
            defer gz.Close()

            // 设置响应头
            w.Header().Set("Content-Encoding", "gzip")
            w.Header().Set("Vary", "Accept-Encoding")

            // 包装ResponseWriter
            gzw := &gzipResponseWriter{
                ResponseWriter: w,
                Writer:         gz,
            }

            next.ServeHTTP(gzw, r)
        })
    }
}

// gzip响应写入器
type gzipResponseWriter struct {
    http.ResponseWriter
    io.Writer
}

func (grw *gzipResponseWriter) Write(data []byte) (int, error) {
    return grw.Writer.Write(data)
}

// HTTP/2服务器配置
func ConfigureHTTP2Server(server *http.Server, config *NetworkOptimizationConfig) error {
    if !config.HTTP2.Enabled {
        return nil
    }

    // 配置HTTP/2
    http2Server := &http2.Server{
        MaxConcurrentStreams: config.HTTP2.MaxConcurrentStreams,
        MaxReadFrameSize:     config.HTTP2.MaxFrameSize,
        MaxUploadBufferPerConnection: 1 << 20, // 1MB
        MaxUploadBufferPerStream:     1 << 16, // 64KB
        IdleTimeout:                  config.HTTP2.IdleTimeout,
    }

    return http2.ConfigureServer(server, http2Server)
}

// 连接池监控
type ConnectionPoolMonitor struct {
    transport *http.Transport
    stats     *ConnectionPoolStats
    mutex     sync.RWMutex
}

type ConnectionPoolStats struct {
    IdleConnections     int `json:"idle_connections"`
    ActiveConnections   int `json:"active_connections"`
    TotalConnections    int `json:"total_connections"`
    ConnectionsCreated  int64 `json:"connections_created"`
    ConnectionsReused   int64 `json:"connections_reused"`
    ConnectionsClosed   int64 `json:"connections_closed"`
}

func NewConnectionPoolMonitor(transport *http.Transport) *ConnectionPoolMonitor {
    return &ConnectionPoolMonitor{
        transport: transport,
        stats:     &ConnectionPoolStats{},
    }
}

func (cpm *ConnectionPoolMonitor) GetStats() *ConnectionPoolStats {
    cpm.mutex.RLock()
    defer cpm.mutex.RUnlock()

    // 注意：这里需要使用反射或其他方法来获取transport的内部状态
    // 这是一个简化的实现
    stats := *cpm.stats
    return &stats
}

// 网络性能分析器
type NetworkPerformanceAnalyzer struct {
    metrics map[string]*NetworkMetric
    mutex   sync.RWMutex
}

type NetworkMetric struct {
    Name         string        `json:"name"`
    Count        int64         `json:"count"`
    TotalTime    time.Duration `json:"total_time"`
    MinTime      time.Duration `json:"min_time"`
    MaxTime      time.Duration `json:"max_time"`
    AvgTime      time.Duration `json:"avg_time"`
    ErrorCount   int64         `json:"error_count"`
    LastMeasured time.Time     `json:"last_measured"`
}

func NewNetworkPerformanceAnalyzer() *NetworkPerformanceAnalyzer {
    return &NetworkPerformanceAnalyzer{
        metrics: make(map[string]*NetworkMetric),
    }
}

func (npa *NetworkPerformanceAnalyzer) RecordRequest(name string, duration time.Duration, err error) {
    npa.mutex.Lock()
    defer npa.mutex.Unlock()

    metric, exists := npa.metrics[name]
    if !exists {
        metric = &NetworkMetric{
            Name:    name,
            MinTime: duration,
            MaxTime: duration,
        }
        npa.metrics[name] = metric
    }

    metric.Count++
    metric.TotalTime += duration
    metric.AvgTime = metric.TotalTime / time.Duration(metric.Count)
    metric.LastMeasured = time.Now()

    if duration < metric.MinTime {
        metric.MinTime = duration
    }
    if duration > metric.MaxTime {
        metric.MaxTime = duration
    }

    if err != nil {
        metric.ErrorCount++
    }
}

func (npa *NetworkPerformanceAnalyzer) GetMetrics() map[string]*NetworkMetric {
    npa.mutex.RLock()
    defer npa.mutex.RUnlock()

    result := make(map[string]*NetworkMetric)
    for k, v := range npa.metrics {
        metricCopy := *v
        result[k] = &metricCopy
    }

    return result
}

func (npa *NetworkPerformanceAnalyzer) GetSlowRequests(threshold time.Duration) map[string]*NetworkMetric {
    npa.mutex.RLock()
    defer npa.mutex.RUnlock()

    slowRequests := make(map[string]*NetworkMetric)
    for name, metric := range npa.metrics {
        if metric.AvgTime > threshold || metric.MaxTime > threshold*2 {
            metricCopy := *metric
            slowRequests[name] = &metricCopy
        }
    }

    return slowRequests
}
```

---

## 🎯 面试常考知识点

### 1. Go性能分析工具

**Q: 请详细介绍Go语言的性能分析工具pprof的使用方法？**

**A: pprof是Go语言内置的性能分析工具，主要包括以下几个方面：**

1. **CPU性能分析**
```go
// 启动CPU性能分析
import _ "net/http/pprof"

// 在代码中启动
f, err := os.Create("cpu.prof")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

if err := pprof.StartCPUProfile(f); err != nil {
    log.Fatal(err)
}
defer pprof.StopCPUProfile()

// 分析命令
// go tool pprof cpu.prof
// (pprof) top10
// (pprof) list functionName
// (pprof) web
```

2. **内存性能分析**
```go
// 内存分析
runtime.GC() // 强制GC
f, err := os.Create("mem.prof")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

if err := pprof.WriteHeapProfile(f); err != nil {
    log.Fatal(err)
}

// 分析命令
// go tool pprof mem.prof
// (pprof) top10
// (pprof) list functionName
```

3. **Goroutine分析**
```go
// Goroutine分析
pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

// 通过HTTP接口
// http://localhost:6060/debug/pprof/goroutine?debug=1
```

**Q: 如何使用go trace工具分析程序执行？**

**A: go trace工具用于分析程序的执行轨迹：**

```go
// 启动trace
f, err := os.Create("trace.out")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

if err := trace.Start(f); err != nil {
    log.Fatal(err)
}
defer trace.Stop()

// 分析命令
// go tool trace trace.out
// 在浏览器中查看详细的执行时间线
```

### 2. 内存优化策略

**Q: Go语言中如何检测和避免内存泄漏？**

**A: 内存泄漏检测和预防策略：**

1. **常见内存泄漏场景**
```go
// 1. Goroutine泄漏
func badGoroutine() {
    go func() {
        for {
            // 没有退出条件的无限循环
            time.Sleep(1 * time.Second)
        }
    }()
}

// 正确做法
func goodGoroutine(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                time.Sleep(1 * time.Second)
            }
        }
    }()
}

// 2. 切片容量泄漏
func badSlice() []int {
    s := make([]int, 1000000)
    return s[:3] // 仍然引用整个底层数组
}

// 正确做法
func goodSlice() []int {
    s := make([]int, 1000000)
    result := make([]int, 3)
    copy(result, s[:3])
    return result
}

// 3. 定时器泄漏
func badTimer() {
    timer := time.NewTimer(1 * time.Hour)
    // 忘记停止定时器
}

// 正确做法
func goodTimer() {
    timer := time.NewTimer(1 * time.Hour)
    defer timer.Stop()
}
```

2. **内存泄漏检测方法**
```go
// 使用runtime.ReadMemStats监控
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc = %d KB", bToKb(m.Alloc))
fmt.Printf("TotalAlloc = %d KB", bToKb(m.TotalAlloc))
fmt.Printf("Sys = %d KB", bToKb(m.Sys))
fmt.Printf("NumGC = %v\n", m.NumGC)

// 使用pprof分析堆内存
go tool pprof http://localhost:6060/debug/pprof/heap
```

**Q: 如何优化Go程序的GC性能？**

**A: GC优化策略：**

1. **调整GOGC参数**
```go
// 设置GC目标百分比
debug.SetGCPercent(200) // 默认100

// 环境变量设置
// GOGC=200 go run main.go
```

2. **减少内存分配**
```go
// 使用对象池
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func useBuffer() {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf)
    // 使用buf
}

// 预分配切片容量
func preAllocate() {
    // 不好的做法
    var s []int
    for i := 0; i < 1000; i++ {
        s = append(s, i) // 多次重新分配
    }

    // 好的做法
    s := make([]int, 0, 1000) // 预分配容量
    for i := 0; i < 1000; i++ {
        s = append(s, i)
    }
}
```

### 3. 并发性能优化

**Q: 如何设计高性能的Goroutine池？**

**A: Goroutine池设计要点：**

1. **动态调整池大小**
```go
type WorkerPool struct {
    workers    int
    maxWorkers int
    taskQueue  chan Task
    wg         sync.WaitGroup
}

func (wp *WorkerPool) adjustWorkers() {
    queueLen := len(wp.taskQueue)

    // 队列积压过多，增加worker
    if queueLen > cap(wp.taskQueue)/2 && wp.workers < wp.maxWorkers {
        wp.addWorker()
    }

    // 队列空闲，减少worker
    if queueLen == 0 && wp.workers > wp.minWorkers {
        wp.removeWorker()
    }
}
```

2. **避免Goroutine泄漏**
```go
func (wp *WorkerPool) worker(ctx context.Context) {
    defer wp.wg.Done()

    for {
        select {
        case task := <-wp.taskQueue:
            task.Execute()
        case <-ctx.Done():
            return // 优雅退出
        }
    }
}
```

**Q: Channel的性能优化技巧有哪些？**

**A: Channel优化策略：**

1. **选择合适的缓冲区大小**
```go
// 无缓冲channel - 同步通信
ch := make(chan int)

// 有缓冲channel - 异步通信
ch := make(chan int, 100)

// 根据生产者消费者速度调整缓冲区大小
```

2. **使用select避免阻塞**
```go
// 非阻塞发送
select {
case ch <- data:
    // 发送成功
default:
    // 发送失败，处理逻辑
}

// 带超时的操作
select {
case result := <-ch:
    // 接收成功
case <-time.After(1 * time.Second):
    // 超时处理
}
```

### 4. 数据库性能优化

**Q: 如何优化数据库连接池的性能？**

**A: 数据库连接池优化：**

1. **合理设置连接池参数**
```go
// 设置最大连接数
db.SetMaxOpenConns(25)

// 设置最大空闲连接数
db.SetMaxIdleConns(25)

// 设置连接最大生存时间
db.SetConnMaxLifetime(5 * time.Minute)

// 设置连接最大空闲时间
db.SetConnMaxIdleTime(5 * time.Minute)
```

2. **监控连接池状态**
```go
stats := db.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
fmt.Printf("Wait count: %d\n", stats.WaitCount)
fmt.Printf("Wait duration: %v\n", stats.WaitDuration)
```

**Q: 如何进行SQL查询优化？**

**A: SQL查询优化策略：**

1. **使用预编译语句**
```go
// 预编译语句
stmt, err := db.Prepare("SELECT * FROM users WHERE id = ?")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

// 重复使用
for _, id := range userIDs {
    rows, err := stmt.Query(id)
    // 处理结果
}
```

2. **批量操作**
```go
// 批量插入
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

stmt, err := tx.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    _, err := stmt.Exec(user.Name, user.Email)
    if err != nil {
        tx.Rollback()
        log.Fatal(err)
    }
}

stmt.Close()
tx.Commit()
```

### 5. 缓存优化策略

**Q: 如何设计多级缓存系统？**

**A: 多级缓存设计原则：**

1. **缓存层级设计**
```go
// L1: 内存缓存 (最快，容量小)
// L2: Redis缓存 (快，容量中等)
// L3: 分布式缓存 (相对慢，容量大)

type MultiLevelCache struct {
    l1 *MemoryCache
    l2 *RedisCache
    l3 *DistributedCache
}

func (mlc *MultiLevelCache) Get(key string) (interface{}, bool) {
    // 依次查找各级缓存
    if value, found := mlc.l1.Get(key); found {
        return value, true
    }

    if value, found := mlc.l2.Get(key); found {
        mlc.l1.Set(key, value) // 回写到L1
        return value, true
    }

    if value, found := mlc.l3.Get(key); found {
        mlc.l1.Set(key, value) // 回写到L1
        mlc.l2.Set(key, value) // 回写到L2
        return value, true
    }

    return nil, false
}
```

2. **缓存策略选择**
```go
// 写穿策略 (Write-Through)
func (cache *Cache) SetWriteThrough(key string, value interface{}) {
    cache.Set(key, value)
    database.Save(key, value) // 同时写入数据库
}

// 写回策略 (Write-Back)
func (cache *Cache) SetWriteBack(key string, value interface{}) {
    cache.Set(key, value)
    cache.markDirty(key) // 标记为脏数据，稍后写入数据库
}

// 写绕策略 (Write-Around)
func (cache *Cache) SetWriteAround(key string, value interface{}) {
    database.Save(key, value) // 直接写入数据库，不更新缓存
}
```

---

## 🏋️ 实战练习题

### 练习1: 性能分析与优化

**题目**: 给定一个Go程序，使用pprof工具分析性能瓶颈并进行优化

**问题代码**:
```go
func inefficientFunction() {
    var result []string
    for i := 0; i < 100000; i++ {
        result = append(result, fmt.Sprintf("item_%d", i))
    }

    // 字符串拼接
    var combined string
    for _, s := range result {
        combined += s + ","
    }

    // 频繁的map操作
    m := make(map[string]int)
    for i, s := range result {
        m[s] = i
    }
}
```

**要求**:
1. 使用pprof分析CPU和内存使用情况
2. 识别性能瓶颈
3. 提供优化方案
4. 对比优化前后的性能数据

**优化方向**:
- 预分配切片容量
- 使用strings.Builder进行字符串拼接
- 优化map操作
- 减少内存分配

### 练习2: 并发性能优化

**题目**: 设计一个高性能的任务处理系统

**要求**:
1. 实现一个可动态调整大小的Goroutine池
2. 支持任务优先级
3. 实现任务超时机制
4. 提供详细的性能监控
5. 处理优雅关闭

**技术要点**:
- Goroutine池管理
- Channel优化
- 上下文传递
- 性能监控
- 资源清理

### 练习3: 缓存系统设计

**题目**: 实现一个高性能的多级缓存系统

**要求**:
1. 支持L1(内存)、L2(Redis)、L3(分布式)三级缓存
2. 实现LRU、LFU等淘汰策略
3. 支持缓存预热和失效
4. 实现缓存穿透、雪崩防护
5. 提供缓存命中率统计

**技术要点**:
- 多级缓存架构
- 缓存策略实现
- 并发安全
- 性能监控
- 故障处理

---

## 📚 本章总结

通过本章学习，我们深入掌握了Go程序性能优化的完整体系：

### 🎯 核心收获

1. **性能分析工具** 🔍
   - 掌握了pprof、trace等性能分析工具的使用
   - 学会了CPU、内存、Goroutine等性能指标的分析
   - 理解了性能瓶颈的识别和定位方法

2. **内存优化策略** 🧠
   - 深入了解了内存泄漏的检测和预防
   - 掌握了GC调优的方法和技巧
   - 学会了对象池的设计和使用

3. **并发性能调优** ⚡
   - 掌握了Goroutine池的设计和优化
   - 学会了Channel的性能优化技巧
   - 理解了锁优化和无锁编程

4. **数据库性能优化** 🗄️
   - 掌握了连接池的配置和监控
   - 学会了SQL查询的优化策略
   - 理解了数据库性能监控的重要性

5. **缓存策略优化** 💾
   - 深入了解了多级缓存的设计原理
   - 掌握了缓存策略的选择和实现
   - 学会了缓存性能的监控和优化

6. **网络性能优化** 🌐
   - 掌握了HTTP/2和连接复用的优化
   - 学会了压缩和缓存的网络优化技术
   - 理解了网络性能监控的方法

### 🚀 技术进阶

- **微基准测试**: 使用go test -bench进行精确的性能测试
- **火焰图分析**: 使用火焰图可视化性能瓶颈
- **分布式性能优化**: 跨服务的性能优化策略
- **云原生性能优化**: 容器化环境下的性能调优

### 💡 最佳实践

1. **测量驱动优化**: 先测量，再优化，避免过早优化
2. **渐进式优化**: 从最大的瓶颈开始，逐步优化
3. **持续监控**: 建立性能监控体系，及时发现问题
4. **平衡权衡**: 在性能、可维护性、开发效率间找平衡

性能优化是一个持续的过程，需要结合具体的业务场景和系统特点。通过本章的学习，你已经具备了系统性的性能优化能力！ 🎉

---

*至此，Go语言学习文档系列的高级篇全部完成！从生产实践到容器化部署，从监控日志到性能优化，我们构建了完整的企业级Go开发技能体系！* 🚀

**🎊 恭喜你完成了整个Go语言学习文档系列！现在你已经具备了从入门到精通的完整Go开发能力！** 🎊
