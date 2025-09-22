package payment

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// SecurityManager 安全管理器
type SecurityManager struct {
	config      *SecurityConfig
	nonceStore  *NonceStore
	ipWhitelist map[string]bool
	rateLimiter *RateLimiter
	mutex       sync.RWMutex
}

// NonceStore 随机数存储
type NonceStore struct {
	store map[string]time.Time
	mutex sync.RWMutex
	ttl   time.Duration
}

// RateLimiter 限流器
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager(config *SecurityConfig) *SecurityManager {
	sm := &SecurityManager{
		config:      config,
		nonceStore:  NewNonceStore(5 * time.Minute), // 5分钟过期
		ipWhitelist: make(map[string]bool),
		rateLimiter: NewRateLimiter(config.RateLimitRPS, time.Minute),
	}

	// 初始化IP白名单
	for _, ip := range config.AllowedIPs {
		sm.ipWhitelist[ip] = true
	}

	// 启动清理任务
	go sm.startCleanupTask()

	return sm
}

// NewNonceStore 创建随机数存储
func NewNonceStore(ttl time.Duration) *NonceStore {
	return &NonceStore{
		store: make(map[string]time.Time),
		ttl:   ttl,
	}
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// VerifySignature 验证签名
func (sm *SecurityManager) VerifySignature(data map[string]string, signature string, signType string) error {
	if !sm.config.EnableSignature {
		return nil
	}

	// 计算签名
	calculatedSign, err := sm.calculateSignature(data, signType)
	if err != nil {
		return fmt.Errorf("计算签名失败: %v", err)
	}

	// 验证签名
	if signature != calculatedSign {
		logger.Error("签名验证失败",
			zap.String("expected", calculatedSign),
			zap.String("actual", signature))
		return fmt.Errorf("签名验证失败")
	}

	return nil
}

// calculateSignature 计算签名
func (sm *SecurityManager) calculateSignature(data map[string]string, signType string) (string, error) {
	// 排序参数
	var keys []string
	for k := range data {
		if k != "sign" && k != "sign_type" && data[k] != "" {
			keys = append(keys, k)
		}
	}

	// 字典序排序
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// 构建签名字符串
	var signStr strings.Builder
	for i, k := range keys {
		if i > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(data[k])
	}

	// 添加密钥
	signStr.WriteString("&key=")
	signStr.WriteString(sm.config.SecretKey)

	// 根据签名类型计算签名
	switch strings.ToUpper(signType) {
	case "MD5":
		hash := md5.Sum([]byte(signStr.String()))
		return strings.ToUpper(hex.EncodeToString(hash[:])), nil
	case "HMAC-SHA256":
		h := hmac.New(sha256.New, []byte(sm.config.SecretKey))
		h.Write([]byte(signStr.String()))
		return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
	default:
		return "", fmt.Errorf("不支持的签名类型: %s", signType)
	}
}

// CheckNonce 检查随机数（防重放攻击）
func (sm *SecurityManager) CheckNonce(nonce string) error {
	return sm.nonceStore.Check(nonce)
}

// Check 检查随机数
func (ns *NonceStore) Check(nonce string) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// 检查是否已存在
	if _, exists := ns.store[nonce]; exists {
		return fmt.Errorf("随机数已使用")
	}

	// 存储随机数
	ns.store[nonce] = time.Now()

	return nil
}

// CheckTimestamp 检查时间戳（防重放攻击）
func (sm *SecurityManager) CheckTimestamp(timestamp string, tolerance time.Duration) error {
	// 解析时间戳
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的时间戳格式: %v", err)
	}

	// 检查时间差
	now := time.Now().Unix()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}

	if time.Duration(diff)*time.Second > tolerance {
		return fmt.Errorf("时间戳超出允许范围")
	}

	return nil
}

// CheckIPWhitelist 检查IP白名单
func (sm *SecurityManager) CheckIPWhitelist(clientIP string) error {
	// 如果白名单为空，允许所有IP
	if len(sm.ipWhitelist) == 0 {
		return nil
	}

	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 检查IP是否在白名单中
	if !sm.ipWhitelist[clientIP] {
		// 检查是否在CIDR范围内
		for allowedIP := range sm.ipWhitelist {
			if sm.isIPInCIDR(clientIP, allowedIP) {
				return nil
			}
		}
		return fmt.Errorf("IP地址不在白名单中: %s", clientIP)
	}

	return nil
}

// isIPInCIDR 检查IP是否在CIDR范围内
func (sm *SecurityManager) isIPInCIDR(ip, cidr string) bool {
	// 如果不包含/，说明是单个IP
	if !strings.Contains(cidr, "/") {
		return ip == cidr
	}

	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	return network.Contains(clientIP)
}

// CheckRateLimit 检查限流
func (sm *SecurityManager) CheckRateLimit(key string) error {
	return sm.rateLimiter.Check(key)
}

// Check 检查限流
func (rl *RateLimiter) Check(key string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// 获取请求历史
	requests, exists := rl.requests[key]
	if !exists {
		requests = make([]time.Time, 0)
	}

	// 清理过期请求
	var validRequests []time.Time
	for _, reqTime := range requests {
		if now.Sub(reqTime) <= rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= rl.limit {
		return fmt.Errorf("请求频率超过限制")
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return nil
}

// ValidateRequestSize 验证请求大小
func (sm *SecurityManager) ValidateRequestSize(size int64) error {
	if sm.config.MaxRequestSize > 0 && size > sm.config.MaxRequestSize {
		return fmt.Errorf("请求大小超过限制: %d > %d", size, sm.config.MaxRequestSize)
	}
	return nil
}

// GenerateToken 生成安全令牌
func (sm *SecurityManager) GenerateToken(data map[string]string) (string, error) {
	// 添加时间戳
	data["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)

	// 生成签名
	signature, err := sm.calculateSignature(data, "HMAC-SHA256")
	if err != nil {
		return "", err
	}

	return signature, nil
}

// VerifyToken 验证安全令牌
func (sm *SecurityManager) VerifyToken(data map[string]string, token string) error {
	// 检查时间戳
	if timestamp, exists := data["timestamp"]; exists {
		if err := sm.CheckTimestamp(timestamp, sm.config.TokenExpiry); err != nil {
			return err
		}
	}

	// 验证签名
	return sm.VerifySignature(data, token, "HMAC-SHA256")
}

// startCleanupTask 启动清理任务
func (sm *SecurityManager) startCleanupTask() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

// cleanup 清理过期数据
func (sm *SecurityManager) cleanup() {
	// 清理过期的随机数
	sm.nonceStore.cleanup()

	// 清理过期的限流记录
	sm.rateLimiter.cleanup()
}

// cleanup 清理过期的随机数
func (ns *NonceStore) cleanup() {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	now := time.Now()
	for nonce, timestamp := range ns.store {
		if now.Sub(timestamp) > ns.ttl {
			delete(ns.store, nonce)
		}
	}
}

// cleanup 清理过期的限流记录
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) <= rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// GetSecuritySummary 获取安全状态摘要
func (sm *SecurityManager) GetSecuritySummary() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return map[string]interface{}{
		"signature_enabled":  sm.config.EnableSignature,
		"encrypt_enabled":    sm.config.EnableEncrypt,
		"ip_whitelist_count": len(sm.ipWhitelist),
		"rate_limit_rps":     sm.config.RateLimitRPS,
		"max_request_size":   sm.config.MaxRequestSize,
		"token_expiry":       sm.config.TokenExpiry.String(),
		"nonce_store_size":   len(sm.nonceStore.store),
		"rate_limit_keys":    len(sm.rateLimiter.requests),
	}
}
