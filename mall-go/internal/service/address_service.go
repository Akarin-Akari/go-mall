package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddressService 地址服务
type AddressService struct {
	db *gorm.DB
}

// NewAddressService 创建新的地址服务实例
func NewAddressService(db *gorm.DB) *AddressService {
	return &AddressService{
		db: db,
	}
}

// CreateAddress 创建新地址
func (s *AddressService) CreateAddress(ctx context.Context, userID uint, req *model.AddressCreateRequest) (*model.Address, error) {
	// 验证地址字段
	if err := s.validateAddressRequest(req); err != nil {
		return nil, err
	}

	// 检查用户地址数量限制
	var count int64
	if err := s.db.Model(&model.Address{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("查询用户地址数量失败", zap.Error(err))
		return nil, err
	}
	if count >= 20 { // 限制每个用户最多20个地址
		return nil, errors.New("地址数量已达上限")
	}

	// 设置默认地址类型
	addressType := req.AddressType
	if addressType == "" {
		addressType = model.AddressTypeHome
	}

	// 创建地址对象
	address := &model.Address{
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

	// 保存到数据库
	if err := s.db.Create(address).Error; err != nil {
		logger.Error("创建地址失败", zap.Error(err))
		return nil, err
	}

	logger.Info("创建地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID),
		zap.String("receiver_name", address.ReceiverName))

	return address, nil
}

// GetUserAddresses 获取用户所有地址
func (s *AddressService) GetUserAddresses(ctx context.Context, userID uint) ([]*model.Address, error) {
	var addresses []*model.Address
	if err := s.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error; err != nil {
		logger.Error("获取用户地址列表失败", zap.Error(err))
		return nil, err
	}
	return addresses, nil
}

// GetAddressByID 根据ID获取地址
func (s *AddressService) GetAddressByID(ctx context.Context, userID, addressID uint) (*model.Address, error) {
	var address model.Address
	if err := s.db.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("地址不存在")
		}
		logger.Error("查询地址失败", zap.Error(err))
		return nil, err
	}
	return &address, nil
}

// UpdateAddress 更新地址
func (s *AddressService) UpdateAddress(ctx context.Context, userID, addressID uint, req *model.AddressUpdateRequest) (*model.Address, error) {
	// 查找地址
	address, err := s.GetAddressByID(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.ReceiverName != "" {
		address.ReceiverName = req.ReceiverName
	}
	if req.ReceiverPhone != "" {
		// 验证手机号格式
		if !s.isValidPhone(req.ReceiverPhone) {
			return nil, errors.New("手机号格式不正确")
		}
		address.ReceiverPhone = req.ReceiverPhone
	}
	if req.Province != "" {
		address.Province = req.Province
	}
	if req.City != "" {
		address.City = req.City
	}
	if req.District != "" {
		address.District = req.District
	}
	if req.DetailAddress != "" {
		address.DetailAddress = req.DetailAddress
	}
	if req.PostalCode != "" {
		// 验证邮编格式
		if !s.isValidPostalCode(req.PostalCode) {
			return nil, errors.New("邮政编码格式不正确")
		}
		address.PostalCode = req.PostalCode
	}
	if req.IsDefault != nil {
		address.IsDefault = *req.IsDefault
	}
	if req.AddressType != "" {
		address.AddressType = req.AddressType
	}

	// 保存更新
	if err := s.db.Save(address).Error; err != nil {
		logger.Error("更新地址失败", zap.Error(err))
		return nil, err
	}

	logger.Info("更新地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	return address, nil
}

// DeleteAddress 删除地址
func (s *AddressService) DeleteAddress(ctx context.Context, userID, addressID uint) error {
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
	// 查找地址
	address, err := s.GetAddressByID(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}

	// 设置为默认地址
	address.IsDefault = true
	if err := s.db.Save(address).Error; err != nil {
		logger.Error("设置默认地址失败", zap.Error(err))
		return nil, err
	}

	logger.Info("设置默认地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	return address, nil
}

// GetDefaultAddress 获取默认地址
func (s *AddressService) GetDefaultAddress(ctx context.Context, userID uint) (*model.Address, error) {
	var address model.Address
	if err := s.db.Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("未找到默认地址")
		}
		logger.Error("查询默认地址失败", zap.Error(err))
		return nil, err
	}
	return &address, nil
}

// validateAddressRequest 验证地址请求
func (s *AddressService) validateAddressRequest(req *model.AddressCreateRequest) error {
	if req.ReceiverName == "" {
		return errors.New("收货人姓名不能为空")
	}
	if req.ReceiverPhone == "" {
		return errors.New("收货人电话不能为空")
	}
	if !s.isValidPhone(req.ReceiverPhone) {
		return errors.New("手机号格式不正确")
	}
	if req.Province == "" {
		return errors.New("省份不能为空")
	}
	if req.City == "" {
		return errors.New("城市不能为空")
	}
	if req.District == "" {
		return errors.New("区县不能为空")
	}
	if req.DetailAddress == "" {
		return errors.New("详细地址不能为空")
	}
	if req.PostalCode != "" && !s.isValidPostalCode(req.PostalCode) {
		return errors.New("邮政编码格式不正确")
	}
	return nil
}

// isValidPhone 验证手机号格式
func (s *AddressService) isValidPhone(phone string) bool {
	// 中国大陆手机号正则表达式
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return phoneRegex.MatchString(phone)
}

// isValidPostalCode 验证邮政编码格式
func (s *AddressService) isValidPostalCode(code string) bool {
	// 中国邮政编码正则表达式（6位数字）
	codeRegex := regexp.MustCompile(`^\d{6}$`)
	return codeRegex.MatchString(code)
}

// GetAddressesWithFilter 根据条件获取地址列表
func (s *AddressService) GetAddressesWithFilter(ctx context.Context, req *model.AddressListRequest) ([]*model.Address, error) {
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
		return nil, err
	}

	return addresses, nil
}

// ValidateAddressOwnership 验证地址归属
func (s *AddressService) ValidateAddressOwnership(ctx context.Context, userID, addressID uint) error {
	var count int64
	if err := s.db.Model(&model.Address{}).
		Where("id = ? AND user_id = ?", addressID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("地址不存在或无权限访问")
	}
	return nil
}
