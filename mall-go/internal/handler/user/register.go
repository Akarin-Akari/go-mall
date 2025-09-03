package user

import (
	"net/http"

	"mall-go/pkg/response"
	"mall-go/pkg/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterHandler 用户注册处理器
type RegisterHandler struct {
	db              *gorm.DB
	registerService *user.RegisterService
}

// NewRegisterHandler 创建用户注册处理器
func NewRegisterHandler(db *gorm.DB) *RegisterHandler {
	return &RegisterHandler{
		db:              db,
		registerService: user.NewRegisterService(db),
	}
}

// Register 用户注册
func (h *RegisterHandler) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 执行注册
	resp, err := h.registerService.Register(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "注册成功", resp)
}

// SendEmailCode 发送邮箱验证码
func (h *RegisterHandler) SendEmailCode(c *gin.Context) {
	var req user.SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	err := h.registerService.SendEmailCode(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "验证码发送成功", gin.H{
		"email": req.Email,
		"type":  req.Type,
	})
}

// SendPhoneCode 发送手机验证码
func (h *RegisterHandler) SendPhoneCode(c *gin.Context) {
	var req user.SendPhoneCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	err := h.registerService.SendPhoneCode(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "验证码发送成功", gin.H{
		"phone": req.Phone,
		"type":  req.Type,
	})
}

// CheckUsername 检查用户名可用性
func (h *RegisterHandler) CheckUsername(c *gin.Context) {
	var req user.CheckUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	available, err := h.registerService.CheckUsername(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var message string
	if available {
		message = "用户名可用"
	} else {
		message = "用户名已被使用"
	}

	response.Success(c, message, gin.H{
		"username":  req.Username,
		"available": available,
	})
}

// CheckEmail 检查邮箱可用性
func (h *RegisterHandler) CheckEmail(c *gin.Context) {
	var req user.CheckEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	available, err := h.registerService.CheckEmail(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var message string
	if available {
		message = "邮箱可用"
	} else {
		message = "邮箱已被注册"
	}

	response.Success(c, message, gin.H{
		"email":     req.Email,
		"available": available,
	})
}

// CheckUsernameByQuery 通过查询参数检查用户名可用性
func (h *RegisterHandler) CheckUsernameByQuery(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		response.Error(c, http.StatusBadRequest, "用户名不能为空")
		return
	}

	req := user.CheckUsernameRequest{Username: username}
	available, err := h.registerService.CheckUsername(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var message string
	if available {
		message = "用户名可用"
	} else {
		message = "用户名已被使用"
	}

	response.Success(c, message, gin.H{
		"username":  username,
		"available": available,
	})
}

// CheckEmailByQuery 通过查询参数检查邮箱可用性
func (h *RegisterHandler) CheckEmailByQuery(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "邮箱不能为空")
		return
	}

	req := user.CheckEmailRequest{Email: email}
	available, err := h.registerService.CheckEmail(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var message string
	if available {
		message = "邮箱可用"
	} else {
		message = "邮箱已被注册"
	}

	response.Success(c, message, gin.H{
		"email":     email,
		"available": available,
	})
}

// GetRegisterConfig 获取注册配置
func (h *RegisterHandler) GetRegisterConfig(c *gin.Context) {
	config := gin.H{
		"username_rules": gin.H{
			"min_length":    3,
			"max_length":    50,
			"allowed_chars": "字母、数字、下划线",
			"pattern":       "^[a-zA-Z0-9_]{3,50}$",
		},
		"password_rules": gin.H{
			"min_length":     6,
			"max_length":     50,
			"require_number": false,
			"require_upper":  false,
			"require_lower":  false,
			"require_symbol": false,
		},
		"email_rules": gin.H{
			"required":      true,
			"need_verify":   true,
			"verify_expire": 600, // 10分钟
		},
		"phone_rules": gin.H{
			"required":      false,
			"need_verify":   true,
			"verify_expire": 300, // 5分钟
			"pattern":       "^1[3-9]\\d{9}$",
		},
		"verification_code": gin.H{
			"email_expire": 600, // 10分钟
			"phone_expire": 300, // 5分钟
			"code_length":  6,
		},
		"terms": gin.H{
			"required": true,
			"url":      "/terms-of-service",
		},
	}

	response.Success(c, "获取注册配置成功", config)
}

// ResendEmailCode 重新发送邮箱验证码
func (h *RegisterHandler) ResendEmailCode(c *gin.Context) {
	email := c.PostForm("email")
	codeType := c.DefaultPostForm("type", "register")

	if email == "" {
		response.Error(c, http.StatusBadRequest, "邮箱不能为空")
		return
	}

	req := user.SendEmailCodeRequest{
		Email: email,
		Type:  codeType,
	}

	err := h.registerService.SendEmailCode(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "验证码重新发送成功", gin.H{
		"email": email,
		"type":  codeType,
	})
}

// ResendPhoneCode 重新发送手机验证码
func (h *RegisterHandler) ResendPhoneCode(c *gin.Context) {
	phone := c.PostForm("phone")
	codeType := c.DefaultPostForm("type", "register")

	if phone == "" {
		response.Error(c, http.StatusBadRequest, "手机号不能为空")
		return
	}

	req := user.SendPhoneCodeRequest{
		Phone: phone,
		Type:  codeType,
	}

	err := h.registerService.SendPhoneCode(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, "验证码重新发送成功", gin.H{
		"phone": phone,
		"type":  codeType,
	})
}

// ValidateRegisterData 验证注册数据（用于前端实时验证）
func (h *RegisterHandler) ValidateRegisterData(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 执行各项验证但不实际注册
	errors := make(map[string]string)

	// 检查用户名
	if exists, err := h.registerService.IsUsernameExists(req.Username); err != nil {
		errors["username"] = "检查用户名失败"
	} else if exists {
		errors["username"] = "用户名已存在"
	}

	// 检查邮箱
	if exists, err := h.registerService.IsEmailExists(req.Email); err != nil {
		errors["email"] = "检查邮箱失败"
	} else if exists {
		errors["email"] = "邮箱已被注册"
	}

	// 检查手机号
	if req.Phone != "" {
		if exists, err := h.registerService.IsPhoneExists(req.Phone); err != nil {
			errors["phone"] = "检查手机号失败"
		} else if exists {
			errors["phone"] = "手机号已被注册"
		}
	}

	// 检查密码确认
	if req.Password != req.ConfirmPassword {
		errors["confirm_password"] = "两次输入的密码不一致"
	}

	valid := len(errors) == 0
	message := "验证通过"
	if !valid {
		message = "验证失败"
	}

	response.Success(c, message, gin.H{
		"valid":  valid,
		"errors": errors,
	})
}
