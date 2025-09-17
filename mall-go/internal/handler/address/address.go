package address

import (
	"net/http"
	"strconv"

	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 地址处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建地址处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// GetAddresses 获取地址列表
// @Summary 获取地址列表
// @Description 获取用户的地址列表
// @Tags 地址管理
// @Accept json
// @Produce json
// @Param user_id query uint false "用户ID(管理员使用)"
// @Param address_type query string false "地址类型"
// @Param is_default query bool false "是否默认地址"
// @Success 200 {object} response.Response{data=[]model.AddressResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/addresses [get]
// @Security ApiKeyAuth
func (h *Handler) GetAddresses(c *gin.Context) {
	var req model.AddressListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 获取当前用户ID
	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 如果不是管理员，只能查看自己的地址
	if req.UserID > 0 && req.UserID != userID && !h.isAdmin(c) {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	// 使用当前用户ID（除非是管理员查看其他用户）
	if req.UserID == 0 || !h.isAdmin(c) {
		req.UserID = userID
	}

	// 构建查询
	query := h.db.Model(&model.Address{}).Where("user_id = ?", req.UserID)

	// 添加筛选条件
	if req.AddressType != "" {
		query = query.Where("address_type = ?", req.AddressType)
	}
	if req.IsDefault != nil {
		query = query.Where("is_default = ?", *req.IsDefault)
	}

	// 查询地址列表
	var addresses []model.Address
	if err := query.Order("is_default DESC, created_at DESC").Find(&addresses).Error; err != nil {
		logger.Error("获取地址列表失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "获取地址列表失败")
		return
	}

	// 转换为响应格式
	var addressResponses []*model.AddressResponse
	for _, addr := range addresses {
		addressResponses = append(addressResponses, addr.ToResponse())
	}

	response.Success(c, "获取地址列表成功", addressResponses)
}

// CreateAddress 创建地址
// @Summary 创建地址
// @Description 创建新的收货地址
// @Tags 地址管理
// @Accept json
// @Produce json
// @Param request body model.AddressCreateRequest true "地址信息"
// @Success 200 {object} response.Response{data=model.AddressResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/addresses [post]
// @Security ApiKeyAuth
func (h *Handler) CreateAddress(c *gin.Context) {
	var req model.AddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 验证地址字段
	if err := model.ValidateAddressFields(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "地址信息不完整")
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 设置默认地址类型
	if req.AddressType == "" {
		req.AddressType = model.AddressTypeHome
	}

	// 检查用户地址数量限制（可选）
	var count int64
	h.db.Model(&model.Address{}).Where("user_id = ?", userID).Count(&count)
	if count >= 20 { // 限制每个用户最多20个地址
		response.Error(c, http.StatusBadRequest, "地址数量已达上限")
		return
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
		AddressType:   req.AddressType,
	}

	// 保存到数据库
	if err := h.db.Create(address).Error; err != nil {
		logger.Error("创建地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "创建地址失败")
		return
	}

	logger.Info("创建地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID),
		zap.String("receiver_name", address.ReceiverName))

	response.Success(c, "创建地址成功", address.ToResponse())
}

// UpdateAddress 更新地址
// @Summary 更新地址
// @Description 更新指定的收货地址
// @Tags 地址管理
// @Accept json
// @Produce json
// @Param id path uint true "地址ID"
// @Param request body model.AddressUpdateRequest true "地址信息"
// @Success 200 {object} response.Response{data=model.AddressResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/addresses/{id} [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的地址ID")
		return
	}

	var req model.AddressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 查找地址
	var address model.Address
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "地址不存在")
			return
		}
		logger.Error("查询地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "查询地址失败")
		return
	}

	// 更新字段
	if req.ReceiverName != "" {
		address.ReceiverName = req.ReceiverName
	}
	if req.ReceiverPhone != "" {
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
		address.PostalCode = req.PostalCode
	}
	if req.IsDefault != nil {
		address.IsDefault = *req.IsDefault
	}
	if req.AddressType != "" {
		address.AddressType = req.AddressType
	}

	// 保存更新
	if err := h.db.Save(&address).Error; err != nil {
		logger.Error("更新地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "更新地址失败")
		return
	}

	logger.Info("更新地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	response.Success(c, "更新地址成功", address.ToResponse())
}

// DeleteAddress 删除地址
// @Summary 删除地址
// @Description 删除指定的收货地址
// @Tags 地址管理
// @Accept json
// @Produce json
// @Param id path uint true "地址ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/addresses/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的地址ID")
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 查找地址
	var address model.Address
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "地址不存在")
			return
		}
		logger.Error("查询地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "查询地址失败")
		return
	}

	// 删除地址
	if err := h.db.Delete(&address).Error; err != nil {
		logger.Error("删除地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "删除地址失败")
		return
	}

	logger.Info("删除地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	response.Success(c, "删除地址成功", nil)
}

// SetDefaultAddress 设置默认地址
// @Summary 设置默认地址
// @Description 将指定地址设置为默认地址
// @Tags 地址管理
// @Accept json
// @Produce json
// @Param id path uint true "地址ID"
// @Success 200 {object} response.Response{data=model.AddressResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/addresses/{id}/default [put]
// @Security ApiKeyAuth
func (h *Handler) SetDefaultAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的地址ID")
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "用户未登录")
		return
	}

	// 查找地址
	var address model.Address
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "地址不存在")
			return
		}
		logger.Error("查询地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "查询地址失败")
		return
	}

	// 设置为默认地址
	address.IsDefault = true
	if err := h.db.Save(&address).Error; err != nil {
		logger.Error("设置默认地址失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "设置默认地址失败")
		return
	}

	logger.Info("设置默认地址成功", 
		zap.Uint("user_id", userID),
		zap.Uint("address_id", address.ID))

	response.Success(c, "设置默认地址成功", address.ToResponse())
}

// GetRegions 获取地区数据
// @Summary 获取地区数据
// @Description 获取省市区三级地区数据
// @Tags 地址管理
// @Accept json
// @Produce json
// @Param parent_code query string false "父级地区代码"
// @Param level query int false "地区级别(1:省份 2:城市 3:区县)"
// @Success 200 {object} response.Response{data=[]model.RegionResponse}
// @Failure 500 {object} response.Response
// @Router /api/v1/addresses/regions [get]
func (h *Handler) GetRegions(c *gin.Context) {
	// 这里可以从数据库或配置文件读取地区数据
	// 为了演示，我们返回一些静态数据
	regions := []model.RegionResponse{
		{
			Code:  "110000",
			Name:  "北京市",
			Level: 1,
			Children: []model.RegionResponse{
				{Code: "110100", Name: "北京市", Level: 2},
			},
		},
		{
			Code:  "310000",
			Name:  "上海市",
			Level: 1,
			Children: []model.RegionResponse{
				{Code: "310100", Name: "上海市", Level: 2},
			},
		},
		{
			Code:  "440000",
			Name:  "广东省",
			Level: 1,
			Children: []model.RegionResponse{
				{Code: "440100", Name: "广州市", Level: 2},
				{Code: "440300", Name: "深圳市", Level: 2},
			},
		},
	}

	response.Success(c, "获取地区数据成功", regions)
}

// getUserID 获取用户ID
func (h *Handler) getUserID(c *gin.Context) uint {
	if uid, exists := c.Get("user_id"); exists {
		return uid.(uint)
	}
	return 0
}

// isAdmin 检查是否为管理员
func (h *Handler) isAdmin(c *gin.Context) bool {
	role, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return role == model.RoleAdmin
}