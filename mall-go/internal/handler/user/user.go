package user

import (
	"net/http"
	"strconv"

	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/logger"
	"mall-go/pkg/response"
	"mall-go/pkg/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	db                *gorm.DB
	loginService      *user.LoginService
	profileService    *user.ProfileService
	permissionService *user.PermissionService
	securityService   *user.SecurityService
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:                db,
		loginService:      user.NewLoginService(db),
		profileService:    user.NewProfileService(db),
		permissionService: user.NewPermissionService(db),
		securityService:   user.NewSecurityService(db),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.UserRegisterRequest true "用户注册信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /users/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req model.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		response.BadRequest(c, "用户名已存在")
		return
	}

	// 检查邮箱是否已存在
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		response.BadRequest(c, "邮箱已存在")
		return
	}

	// 创建用户
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		Role:     req.Role,
		Status:   model.StatusActive,
	}

	// 设置默认角色
	if user.Role == "" {
		user.Role = model.RoleUser
	}

	// 加密密码
	if err := user.SetPassword(req.Password); err != nil {
		logger.Error("密码加密失败", zap.Error(err))
		response.BadRequest(c, "密码格式不正确: "+err.Error())
		return
	}

	if err := h.db.Create(&user).Error; err != nil {
		logger.Error("创建用户失败", zap.Error(err))
		response.ServerError(c, "创建用户失败")
		return
	}

	response.Success(c, "用户注册成功", user.ToResponse())
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.UserLoginRequest true "用户登录信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /users/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req model.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 查找用户
	var user model.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if !user.CanLogin() {
		if user.IsLocked() {
			response.Unauthorized(c, "账户已被锁定，请稍后再试")
			return
		}
		response.Unauthorized(c, "账户状态异常，无法登录")
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		// 增加登录失败次数
		user.IncrementLoginAttempts()
		h.db.Save(&user)

		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 登录成功，重置登录尝试次数并更新最后登录时间
	user.ResetLoginAttempts()
	user.UpdateLastLogin()
	h.db.Save(&user)

	// 生成JWT令牌
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		logger.Error("生成JWT令牌失败", zap.Error(err))
		response.ServerError(c, "登录失败，请稍后重试")
		return
	}

	loginData := map[string]interface{}{
		"token":      token,
		"expires_in": 86400, // 24小时，单位：秒
		"user":       user.ToResponse(),
	}
	response.Success(c, "登录成功", loginData)
}

// GetProfile 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Router /users/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		response.ServerError(c, "用户信息格式错误")
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	userResponse := model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
	response.SuccessWithData(c, userResponse)
}

// UpdateProfile 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body model.UserResponse true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /users/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		response.ServerError(c, "用户信息格式错误")
		return
	}

	var req model.UserResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 更新用户信息
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	user.Phone = req.Phone

	if err := h.db.Save(&user).Error; err != nil {
		logger.Error("更新用户信息失败", zap.Error(err))
		response.ServerError(c, "更新用户信息失败")
		return
	}

	userResponse := model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
	response.Success(c, "用户信息更新成功", userResponse)
}
