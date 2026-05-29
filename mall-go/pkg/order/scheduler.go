package order

import (
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Scheduler 订单定时调度器
//
// 职责：周期性扫描并处理超时订单——
//   - 支付超时（pending 且 pay_expire_time < now）→ 自动取消并恢复库存
//   - 收货超时（delivered 且 receive_expire_time < now）→ 自动确认收货
//
// 实际处理逻辑复用 StatusService.AutoUpdateExpiredOrders()，本调度器只负责"按节拍触发"，
// 遵循 DRY 原则，不重复实现关单/恢复库存逻辑。
type Scheduler struct {
	statusService *StatusService
	interval      time.Duration
	stopCh        chan struct{}
}

// NewScheduler 创建订单调度器。interval <= 0 时默认 1 分钟。
func NewScheduler(db *gorm.DB, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &Scheduler{
		statusService: NewStatusService(db),
		interval:      interval,
		stopCh:        make(chan struct{}),
	}
}

// Start 在后台 goroutine 中启动周期性扫描。非阻塞。
func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// 启动后先立即执行一次，避免等待首个 interval 周期
		s.runOnce()

		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				logger.Info("订单调度器已停止")
				return
			}
		}
	}()
	logger.Info("订单调度器已启动", zap.Duration("interval", s.interval))
}

// runOnce 执行一次过期订单处理，错误仅记录日志不中断调度循环。
func (s *Scheduler) runOnce() {
	if err := s.statusService.AutoUpdateExpiredOrders(); err != nil {
		logger.Error("自动处理过期订单失败", zap.Error(err))
	}
}

// Stop 停止调度器。多次调用会 panic，调用方需保证只调用一次。
func (s *Scheduler) Stop() {
	close(s.stopCh)
}
