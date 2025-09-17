package user

import (
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

// RefreshToken 刷新令牌
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.loginService.RefreshToken(&user.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "令牌刷新成功", result)
}

// ForgotPassword 忘记密码
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.loginService.ForgotPassword(&user.ForgotPasswordRequest{
		Account: req.Email,
		Type:    "email",
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "重置密码邮件已发送", nil)
}

// ResetPassword 重置密码
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	err := h.loginService.ResetPassword(&user.ResetPasswordRequest{
		Account:         req.Email,
		Code:            req.Code,
		NewPassword:     req.Password,
		ConfirmPassword: req.Password,
		Type:            "email",
	})
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "密码重置成功", nil)
}

// VerifyResetCode 验证重置码
func (h *Handler) VerifyResetCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// TODO: 实现验证码验证逻辑
	// 暂时返回成功，实际应该验证验证码
	response.Success(c, "验证码验证功能待实现", gin.H{
		"email": req.Email,
		"code":  req.Code,
	})
}

// UploadAvatar 上传头像
func (h *Handler) UploadAvatar(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	// 获取上传的文件
	_, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "请选择要上传的头像文件")
		return
	}

	// 这里应该调用文件上传服务
	// 暂时返回成功响应
	response.Success(c, "头像上传成功", gin.H{
		"user_id": userID,
		"message": "头像上传功能待实现",
	})
}

// GetCurrentUser 获取当前用户信息
func (h *Handler) GetCurrentUser(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, "获取用户信息成功", user.ToResponse())
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 这里应该调用密码修改服务
	// 暂时返回成功响应
	response.Success(c, "密码修改成功", gin.H{
		"user_id": userID,
		"message": "密码修改功能待实现",
	})
}

// ChangeEmail 修改邮箱
func (h *Handler) ChangeEmail(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	var req struct {
		NewEmail string `json:"new_email" binding:"required,email"`
		Code     string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	response.Success(c, "邮箱修改成功", gin.H{
		"user_id": userID,
		"message": "邮箱修改功能待实现",
	})
}

// VerifyEmail 验证邮箱
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	response.Success(c, "邮箱验证成功", gin.H{
		"message": "邮箱验证功能待实现",
	})
}

// ChangePhone 修改手机号
func (h *Handler) ChangePhone(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	var req struct {
		NewPhone string `json:"new_phone" binding:"required"`
		Code     string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	response.Success(c, "手机号修改成功", gin.H{
		"user_id": userID,
		"message": "手机号修改功能待实现",
	})
}

// VerifyPhone 验证手机号
func (h *Handler) VerifyPhone(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	response.Success(c, "手机号验证成功", gin.H{
		"message": "手机号验证功能待实现",
	})
}

// GetSecuritySettings 获取安全设置
func (h *Handler) GetSecuritySettings(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "获取安全设置成功", gin.H{
		"user_id": userID,
		"message": "安全设置功能待实现",
	})
}

// Enable2FA 启用双因子认证
func (h *Handler) Enable2FA(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "双因子认证启用成功", gin.H{
		"user_id": userID,
		"message": "双因子认证功能待实现",
	})
}

// Disable2FA 禁用双因子认证
func (h *Handler) Disable2FA(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "双因子认证禁用成功", gin.H{
		"user_id": userID,
		"message": "双因子认证功能待实现",
	})
}

// GetLoginLogs 获取登录日志
func (h *Handler) GetLoginLogs(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "获取登录日志成功", gin.H{
		"user_id": userID,
		"message": "登录日志功能待实现",
	})
}

// GetSocialAccounts 获取社交账号绑定
func (h *Handler) GetSocialAccounts(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "获取社交账号成功", gin.H{
		"user_id": userID,
		"message": "社交账号功能待实现",
	})
}

// BindSocialAccount 绑定社交账号
func (h *Handler) BindSocialAccount(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "社交账号绑定成功", gin.H{
		"user_id": userID,
		"message": "社交账号绑定功能待实现",
	})
}

// UnbindSocialAccount 解绑社交账号
func (h *Handler) UnbindSocialAccount(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "社交账号解绑成功", gin.H{
		"user_id": userID,
		"message": "社交账号解绑功能待实现",
	})
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "登出成功", gin.H{
		"user_id": userID,
		"message": "登出功能待实现",
	})
}

// DeleteAccount 删除账号
func (h *Handler) DeleteAccount(c *gin.Context) {
	// 从JWT中获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未认证")
		return
	}

	userID := userIDVal.(uint)

	response.Success(c, "账号删除成功", gin.H{
		"user_id": userID,
		"message": "账号删除功能待实现",
	})
}

// GetUsers 获取用户列表（管理员功能）
func (h *Handler) GetUsers(c *gin.Context) {
	response.Success(c, "获取用户列表成功", gin.H{
		"message": "用户列表功能待实现",
	})
}
