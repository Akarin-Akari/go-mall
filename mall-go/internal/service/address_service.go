package service

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 预编译正则表达式，避免重复编译提高性能
var (
	// 中国大陆手机号正则表达式
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	// 中国邮政编码正则表达式（6位数字）
	postalCodeRegex = regexp.MustCompile(`^\d{6}$`)
)

// AddressService 地址服务
// 实现 IAddressService 接口
type AddressService struct {
	db                *gorm.DB
	configManager     *config.AddressConfigManager
	auditLogger       *AuditLogger
	timeoutManager    *TimeoutManager
	timeoutWrapper    *TimeoutWrapper
	performanceMonitor *PerformanceMonitor
	cacheService      *CacheService
}

// 确保 AddressService 实现了 IAddressService 接口
var _ IAddressService = (*AddressService)(nil)

// NewAddressService 创建新的地址服务实例
func NewAddressService(db *gorm.DB) *AddressService {
	// 使用默认配置创建配置管理器
	configManager, err := config.NewAddressConfigManager(nil)
	if err != nil {
		logger.Error("创建地址配置管理器失败，使用默认配置", zap.Error(err))
		// 如果配置管理器创建失败，使用默认配置重试
		defaultConfig := config.DefaultAddressConfig()
		configManager, _ = config.NewAddressConfigManager(defaultConfig)
	}

	// 创建审计日志记录器
	auditLogger := NewAuditLogger()

	// 创建超时管理器
	timeoutManager := NewTimeoutManager(configManager.GetConfig())
	timeoutWrapper := NewTimeoutWrapper(timeoutManager)

	// 创建性能监控器
	performanceMonitor := NewPerformanceMonitor()

	// 创建缓存服务（Redis客户端为nil时自动禁用缓存）
	cacheService := NewCacheService(nil, configManager.GetConfig())

	return &AddressService{
		db:                 db,
		configManager:      configManager,
		auditLogger:        auditLogger,
		timeoutManager:     timeoutManager,
		timeoutWrapper:     timeoutWrapper,
		performanceMonitor: performanceMonitor,
		cacheService:       cacheService,
	}
}

// NewAddressServiceWithConfig 使用指定配置创建地址服务实例
func NewAddressServiceWithConfig(db *gorm.DB, cfg *config.AddressConfig) *AddressService {
	configManager, err := config.NewAddressConfigManager(cfg)
	if err != nil {
		logger.Error("创建地址配置管理器失败，使用默认配置", zap.Error(err))
		// 如果配置管理器创建失败，使用默认配置
		configManager, _ = config.NewAddressConfigManager(nil)
	}

	// 创建审计日志记录器
	auditLogger := NewAuditLogger()

	// 创建超时管理器
	timeoutManager := NewTimeoutManager(cfg)
	timeoutWrapper := NewTimeoutWrapper(timeoutManager)

	return &AddressService{
		db:             db,
		configManager:  configManager,
		auditLogger:    auditLogger,
		timeoutManager: timeoutManager,
		timeoutWrapper: timeoutWrapper,
	}
}

// CreateAddress 创建新地址
func (s *AddressService) CreateAddress(ctx context.Context, userID uint, req *model.AddressCreateRequest) (*model.Address, error) {
	// 记录操作开始时间用于性能监控
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		s.performanceMonitor.RecordMetric("create_address_duration_ms", float64(duration.Milliseconds()), map[string]string{
			"operation": "create_address",
			"user_id":   strconv.FormatUint(uint64(userID), 10),
		})
	}()

	// 创建操作上下文用于审计日志
	opCtx := CreateOperationContext(userID, "CreateAddress", "address", "create").
		WithDetails(map[string]interface{}{
			"receiver_name": req.ReceiverName,
			"phone":         req.ReceiverPhone,
			"province":      req.Province,
			"city":          req.City,
			"district":      req.District,
		})

	// 输入参数验证
	if ctx == nil {
		s.auditLogger.LogError(ctx, userID, "CreateAddress", ErrInvalidContext, "Context为nil")
		return nil, ErrInvalidContext
	}
	if userID == 0 {
		s.auditLogger.LogError(ctx, userID, "CreateAddress", ErrInvalidUserID, "用户ID为0")
		return nil, ErrInvalidUserID
	}
	if req == nil {
		s.auditLogger.LogError(ctx, userID, "CreateAddress", ErrInvalidRequest, "请求对象为nil")
		return nil, ErrInvalidRequest
	}

	logger.Info("开始创建地址",
		zap.Uint("user_id", userID),
		zap.String("receiver_name", req.ReceiverName))

	// 记录用户行为
	s.auditLogger.LogUserBehavior(ctx, userID, "create_address_attempt", map[string]interface{}{
		"receiver_name": req.ReceiverName,
		"phone":         req.ReceiverPhone,
		"province":      req.Province,
		"city":          req.City,
	})

	// 验证地址字段
	if err := s.validateAddressRequest(req); err != nil {
		return nil, err
	}

	// 设置默认地址类型
	addressType := req.AddressType
	if addressType == "" {
		addressType = model.AddressTypeHome
	}

	var address *model.Address

	// 使用事务确保数据一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查用户地址数量限制
		var count int64
		if err := tx.Model(&model.Address{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
			logger.Error("查询用户地址数量失败", zap.Error(err))
			return err
		}
		maxAddresses := s.configManager.GetMaxAddressPerUser()
		if count >= int64(maxAddresses) {
			return ErrAddressLimitReached
		}

		// 2. 如果设置为默认地址，先清除其他默认地址
		if req.IsDefault {
			if err := tx.Model(&model.Address{}).
				Where("user_id = ? AND is_default = ?", userID, true).
				Update("is_default", false).Error; err != nil {
				logger.Error("清除默认地址失败", zap.Error(err))
				return err
			}
		}

		// 3. 创建地址对象
		newAddress := &model.Address{
			UserID:        userID,
			ReceiverName:  req.ReceiverName,
			ReceiverPhone: req.ReceiverPhone,
			Province:      req.Province,
			City:          req.City,
			District:      req.District,
			DetailAddress: req.DetailAddress,
			PostalCode:    req.PostalCode,
			IsDefault:     req.IsDefault,
			AddressType:   addressType,
		}

		// 4. 保存到数据库
		if err := tx.Create(newAddress).Error; err != nil {
			logger.Error("创建地址失败", zap.Error(err))
			return err
		}

		address = newAddress
		return nil
	})

	if err != nil {
		logger.Error("创建地址事务失败",
			zap.Uint("user_id", userID),
			zap.Error(err))

		// 记录操作失败的审计日志
		s.auditLogger.LogOperation(ctx, opCtx, "error", err)

		// 记录失败的性能指标
		s.performanceMonitor.RecordAddressOperation("create_address", err)
		return nil, err
	}

	// 记录数据变更日志
	s.auditLogger.LogDataChange(ctx, userID, "addresses", address.ID, "create", nil, address)

	// 记录操作成功的审计日志
	opCtx.WithResourceID(address.ID).WithChanges(map[string]interface{}{
		"created_address_id": address.ID,
		"is_default":         address.IsDefault,
	})
	s.auditLogger.LogOperation(ctx, opCtx, "success", nil)

	// 检查是否为慢操作
	if s.configManager != nil && s.configManager.GetConfig().EnableDetailedLog {
		s.auditLogger.LogSlowOperation(ctx, opCtx, 1*time.Second)
	}

	// 记录成功的性能指标
	s.performanceMonitor.RecordAddressOperation("create_address", nil)
	s.performanceMonitor.IncrementCounter("address_created_total", map[string]string{
		"user_id": strconv.FormatUint(uint64(userID), 10),
	})

	// 清除用户相关缓存
	if s.cacheService != nil && s.cacheService.IsEnabled() {
		if err := s.cacheService.InvalidateUserCache(ctx, userID); err != nil {
			logger.Warn("清除用户缓存失败",
				zap.Uint("user_id", userID),
				zap.Error(err))
		}
	}

	logger.Info("创建地址成功",
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID),
		zap.String("receiver_name", address.ReceiverName))

	return address, nil
}

// GetUserAddresses 获取用户所有地址
func (s *AddressService) GetUserAddresses(ctx context.Context, userID uint) ([]*model.Address, error) {
	// 输入参数验证
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}

	// 先尝试从缓存获取
	if s.cacheService != nil && s.cacheService.IsEnabled() {
		cachedAddresses, err := s.cacheService.GetUserAddresses(ctx, userID)
		if err == nil && cachedAddresses != nil {
			logger.Debug("从缓存获取用户地址列表成功",
				zap.Uint("user_id", userID),
				zap.Int("count", len(cachedAddresses)))
			return cachedAddresses, nil
		}
		if err != nil {
			logger.Warn("从缓存获取用户地址列表失败",
				zap.Uint("user_id", userID),
				zap.Error(err))
		}
	}

	var addresses []*model.Address

	// 使用超时包装器执行查询操作
	err := s.timeoutWrapper.WrapQuery(ctx, "GetUserAddresses", func(timeoutCtx context.Context) error {
		return s.db.WithContext(timeoutCtx).Where("user_id = ?", userID).
			Order("is_default DESC, created_at DESC").
			Find(&addresses).Error
	})

	if err != nil {
		if IsTimeoutError(err) {
			logger.Error("获取用户地址列表超时",
				zap.Uint("user_id", userID),
				zap.Error(err))
			return nil, ErrOperationTimeout
		}
		logger.Error("获取用户地址列表失败", zap.Error(err))
		return nil, NewServiceError(CodeDatabaseOperation, "获取用户地址列表失败", err)
	}

	// 将查询结果存入缓存
	if s.cacheService != nil && s.cacheService.IsEnabled() {
		if err := s.cacheService.SetUserAddresses(ctx, userID, addresses); err != nil {
			logger.Warn("设置用户地址列表缓存失败",
				zap.Uint("user_id", userID),
				zap.Error(err))
		}
	}

	return addresses, nil
}

// GetAddressByID 根据ID获取地址
func (s *AddressService) GetAddressByID(ctx context.Context, userID, addressID uint) (*model.Address, error) {
	// 输入参数验证
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if addressID == 0 {
		return nil, ErrInvalidAddressID
	}

	var address model.Address
	if err := s.db.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAddressNotFound
		}
		logger.Error("查询地址失败", zap.Error(err))
		return nil, NewServiceError(CodeDatabaseOperation, "查询地址失败", err)
	}
	return &address, nil
}

// UpdateAddress 更新地址
func (s *AddressService) UpdateAddress(ctx context.Context, userID, addressID uint, req *model.AddressUpdateRequest) (*model.Address, error) {
	// 输入参数验证
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if addressID == 0 {
		return nil, ErrInvalidAddressID
	}
	if req == nil {
		return nil, ErrInvalidRequest
	}

	logger.Info("开始更新地址",
		zap.Uint("user_id", userID),
		zap.Uint("address_id", addressID))

	var address *model.Address

	// 使用事务确保数据一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查找地址
		var existingAddress model.Address
		if err := tx.Where("id = ? AND user_id = ?", addressID, userID).First(&existingAddress).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("地址不存在")
			}
			logger.Error("查询地址失败", zap.Error(err))
			return err
		}

		// 2. 验证更新字段
		if req.ReceiverPhone != "" && !s.isValidPhone(req.ReceiverPhone) {
			return ErrInvalidPhone
		}
		if req.PostalCode != "" && !s.isValidPostalCode(req.PostalCode) {
			return ErrInvalidPostalCode
		}

		// 3. 如果要设置为默认地址，先清除其他默认地址
		if req.IsDefault != nil && *req.IsDefault {
			if err := tx.Model(&model.Address{}).
				Where("user_id = ? AND id != ? AND is_default = ?", userID, addressID, true).
				Update("is_default", false).Error; err != nil {
				logger.Error("清除默认地址失败", zap.Error(err))
				return err
			}
		}

		// 4. 更新字段
		updates := make(map[string]interface{})
		if req.ReceiverName != "" {
			updates["receiver_name"] = req.ReceiverName
		}
		if req.ReceiverPhone != "" {
			updates["receiver_phone"] = req.ReceiverPhone
		}
		if req.Province != "" {
			updates["province"] = req.Province
		}
		if req.City != "" {
			updates["city"] = req.City
		}
		if req.District != "" {
			updates["district"] = req.District
		}
		if req.DetailAddress != "" {
			updates["detail_address"] = req.DetailAddress
		}
		if req.PostalCode != "" {
			updates["postal_code"] = req.PostalCode
		}
		if req.IsDefault != nil {
			updates["is_default"] = *req.IsDefault
		}
		if req.AddressType != "" {
			updates["address_type"] = req.AddressType
		}

		// 5. 执行更新
		if len(updates) > 0 {
			if err := tx.Model(&existingAddress).Updates(updates).Error; err != nil {
				logger.Error("更新地址失败", zap.Error(err))
				return err
			}
		}

		// 6. 获取更新后的地址
		if err := tx.Where("id = ? AND user_id = ?", addressID, userID).First(&existingAddress).Error; err != nil {
			logger.Error("获取更新后地址失败", zap.Error(err))
			return err
		}

		address = &existingAddress
		return nil
	})

	if err != nil {
		logger.Error("更新地址事务失败",
			zap.Uint("user_id", userID),
			zap.Uint("address_id", addressID),
			zap.Error(err))
		return nil, err
	}

	logger.Info("更新地址成功",
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	return address, nil
}

// DeleteAddress 删除地址
func (s *AddressService) DeleteAddress(ctx context.Context, userID, addressID uint) error {
	// 输入参数验证
	if ctx == nil {
		return ErrInvalidContext
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	if addressID == 0 {
		return ErrInvalidAddressID
	}

	// 查找地址
	address, err := s.GetAddressByID(ctx, userID, addressID)
	if err != nil {
		return err
	}

	// 删除地址
	if err := s.db.Delete(address).Error; err != nil {
		logger.Error("删除地址失败", zap.Error(err))
		return err
	}

	logger.Info("删除地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	return nil
}

// SetDefaultAddress 设置默认地址
func (s *AddressService) SetDefaultAddress(ctx context.Context, userID, addressID uint) (*model.Address, error) {
	// 创建操作上下文用于审计日志
	opCtx := CreateOperationContext(userID, "SetDefaultAddress", "address", "update").
		WithResourceID(addressID).
		WithDetails(map[string]interface{}{
			"target_address_id": addressID,
			"action":            "set_default",
		})

	// 输入参数验证
	if ctx == nil {
		s.auditLogger.LogError(ctx, userID, "SetDefaultAddress", ErrInvalidContext, "Context为nil")
		return nil, ErrInvalidContext
	}
	if userID == 0 {
		s.auditLogger.LogError(ctx, userID, "SetDefaultAddress", ErrInvalidUserID, "用户ID为0")
		return nil, ErrInvalidUserID
	}
	if addressID == 0 {
		s.auditLogger.LogError(ctx, userID, "SetDefaultAddress", ErrInvalidAddressID, "地址ID为0")
		return nil, ErrInvalidAddressID
	}

	logger.Info("开始设置默认地址",
		zap.Uint("user_id", userID),
		zap.Uint("address_id", addressID))

	// 记录用户行为
	s.auditLogger.LogUserBehavior(ctx, userID, "set_default_address_attempt", map[string]interface{}{
		"address_id": addressID,
	})

	var address *model.Address

	// 使用事务确保原子性操作，解决并发安全问题
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 验证地址是否存在且属于该用户
		var count int64
		if err := tx.Model(&model.Address{}).
			Where("id = ? AND user_id = ?", addressID, userID).
			Count(&count).Error; err != nil {
			logger.Error("验证地址归属失败", zap.Error(err))
			return err
		}
		if count == 0 {
			return ErrPermissionDenied
		}

		// 2. 先将该用户的所有地址设为非默认（原子操作）
		if err := tx.Model(&model.Address{}).
			Where("user_id = ? AND is_default = ?", userID, true).
			Update("is_default", false).Error; err != nil {
			logger.Error("清除默认地址失败", zap.Error(err))
			return err
		}

		// 3. 设置指定地址为默认地址（原子操作）
		result := tx.Model(&model.Address{}).
			Where("id = ? AND user_id = ?", addressID, userID).
			Update("is_default", true)

		if result.Error != nil {
			logger.Error("设置默认地址失败", zap.Error(result.Error))
			return result.Error
		}

		// 4. 获取更新后的地址信息
		var updatedAddress model.Address
		if err := tx.Where("id = ? AND user_id = ?", addressID, userID).
			First(&updatedAddress).Error; err != nil {
			logger.Error("获取更新后地址信息失败", zap.Error(err))
			return err
		}

		address = &updatedAddress
		return nil
	})

	if err != nil {
		logger.Error("设置默认地址事务失败",
			zap.Uint("user_id", userID),
			zap.Uint("address_id", addressID),
			zap.Error(err))

		// 记录操作失败的审计日志
		s.auditLogger.LogOperation(ctx, opCtx, "error", err)
		return nil, err
	}

	// 记录操作成功的审计日志
	opCtx.WithChanges(map[string]interface{}{
		"set_default_address_id": address.ID,
		"previous_default":       "cleared",
	})
	s.auditLogger.LogOperation(ctx, opCtx, "success", nil)

	logger.Info("设置默认地址成功",
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	return address, nil
}

// GetDefaultAddress 获取默认地址
func (s *AddressService) GetDefaultAddress(ctx context.Context, userID uint) (*model.Address, error) {
	// 输入参数验证
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}

	var address model.Address
	if err := s.db.Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNoDefaultAddress
		}
		logger.Error("查询默认地址失败", zap.Error(err))
		return nil, NewServiceError(CodeDatabaseOperation, "查询默认地址失败", err)
	}
	return &address, nil
}

// validateAddressRequest 验证地址请求
func (s *AddressService) validateAddressRequest(req *model.AddressCreateRequest) error {
	if req.ReceiverName == "" {
		return ErrEmptyReceiverName
	}
	if req.ReceiverPhone == "" {
		return ErrEmptyReceiverPhone
	}
	if !s.isValidPhone(req.ReceiverPhone) {
		return ErrInvalidPhone
	}
	if req.Province == "" {
		return ErrEmptyProvince
	}
	if req.City == "" {
		return ErrEmptyCity
	}
	if req.District == "" {
		return ErrEmptyDistrict
	}
	if req.DetailAddress == "" {
		return ErrEmptyDetailAddress
	}
	if req.PostalCode != "" && !s.isValidPostalCode(req.PostalCode) {
		return ErrInvalidPostalCode
	}
	return nil
}

// isValidPhone 验证手机号格式
func (s *AddressService) isValidPhone(phone string) bool {
	if s.configManager != nil {
		return s.configManager.ValidatePhone(phone)
	}
	// 降级到预编译的正则表达式
	return phoneRegex.MatchString(phone)
}

// isValidPostalCode 验证邮政编码格式
func (s *AddressService) isValidPostalCode(code string) bool {
	if s.configManager != nil {
		return s.configManager.ValidatePostalCode(code)
	}
	// 降级到预编译的正则表达式
	return postalCodeRegex.MatchString(code)
}

// GetAddressesWithFilter 根据条件获取地址列表
func (s *AddressService) GetAddressesWithFilter(ctx context.Context, req *model.AddressListRequest) ([]*model.Address, error) {
	// 输入参数验证
	if ctx == nil {
		return nil, ErrInvalidContext
	}
	if req == nil {
		return nil, ErrInvalidRequest
	}
	if req.UserID == 0 {
		return nil, ErrInvalidUserID
	}

	// 验证分页参数
	if req.Page < 0 || req.PageSize < 0 {
		return nil, ErrInvalidRequest
	}
	maxPageSize := s.configManager.GetMaxPageSize()
	if req.PageSize > maxPageSize {
		return nil, ErrInvalidRequest
	}

	query := s.db.Model(&model.Address{}).Where("user_id = ?", req.UserID)

	// 添加筛选条件
	if req.AddressType != "" {
		query = query.Where("address_type = ?", req.AddressType)
	}
	if req.IsDefault != nil {
		query = query.Where("is_default = ?", *req.IsDefault)
	}

	// 分页处理
	if req.Page > 0 && req.PageSize > 0 {
		offset := (req.Page - 1) * req.PageSize
		query = query.Offset(offset).Limit(req.PageSize)
	}

	var addresses []*model.Address
	if err := query.Order("is_default DESC, created_at DESC").Find(&addresses).Error; err != nil {
		logger.Error("获取地址列表失败", zap.Error(err))
		return nil, NewServiceError(CodeDatabaseOperation, "获取地址列表失败", err)
	}

	return addresses, nil
}

// ValidateAddressOwnership 验证地址归属
func (s *AddressService) ValidateAddressOwnership(ctx context.Context, userID, addressID uint) error {
	// 输入参数验证
	if ctx == nil {
		return ErrInvalidContext
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	if addressID == 0 {
		return ErrInvalidAddressID
	}

	var count int64
	if err := s.db.Model(&model.Address{}).
		Where("id = ? AND user_id = ?", addressID, userID).
		Count(&count).Error; err != nil {
		return NewServiceError(CodeDatabaseOperation, "验证地址归属失败", err)
	}
	if count == 0 {
		return ErrPermissionDenied
	}
	return nil
}
