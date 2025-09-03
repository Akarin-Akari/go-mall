package file

import (
	"net/http"
	"os"
	"strconv"

	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/fileupload"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FileHandler 文件处理器
type FileHandler struct {
	uploadService *fileupload.FileUploadService
	validator     *fileupload.FileValidator
}

// NewFileHandler 创建文件处理器
func NewFileHandler(db *gorm.DB, uploadPath, baseURL string) *FileHandler {
	return &FileHandler{
		uploadService: fileupload.NewFileUploadService(db, uploadPath, baseURL),
		validator:     fileupload.NewFileValidator(),
	}
}

// UploadSingle 上传单个文件
// @Summary 上传单个文件
// @Description 上传单个文件到服务器
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要上传的文件"
// @Param business_type formData string true "业务类型" Enums(avatar,product,store,other)
// @Param business_id formData int false "关联业务ID"
// @Param is_public formData bool false "是否公开访问"
// @Success 200 {object} response.Response{data=model.FileUploadResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/files/upload [post]
// @Security BearerAuth
func (h *FileHandler) UploadSingle(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 权限检查：验证用户是否有文件上传权限
	hasPermission, err := auth.CheckUserPermission(userID.(uint), model.ResourceFile, model.ActionCreate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "权限检查失败")
		return
	}
	if !hasPermission {
		response.Error(c, http.StatusForbidden, "无权限上传文件")
		return
	}

	// 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	// 获取业务参数
	businessType := model.BusinessType(c.PostForm("business_type"))
	if businessType == "" {
		businessType = model.BusinessTypeOther
	}

	var businessID *uint
	if businessIDStr := c.PostForm("business_id"); businessIDStr != "" {
		if id, err := strconv.ParseUint(businessIDStr, 10, 32); err == nil {
			businessIDUint := uint(id)
			businessID = &businessIDUint
		}
	}

	isPublic := c.PostForm("is_public") == "true"

	// 验证文件
	if err := h.validator.ValidateFile(fileHeader); err != nil {
		response.Error(c, http.StatusBadRequest, "文件验证失败: "+err.Error())
		return
	}

	// 上传文件
	result, err := h.uploadService.UploadFile(
		fileHeader,
		userID.(uint),
		businessType,
		businessID,
		isPublic,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "文件上传失败: "+err.Error())
		return
	}

	response.Success(c, "文件上传成功", result)
}

// UploadMultiple 上传多个文件
// @Summary 上传多个文件
// @Description 批量上传多个文件到服务器
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "要上传的文件（可多选）"
// @Param business_type formData string true "业务类型" Enums(avatar,product,store,other)
// @Param business_id formData int false "关联业务ID"
// @Param is_public formData bool false "是否公开访问"
// @Success 200 {object} response.Response{data=[]model.FileUploadResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/files/upload/multiple [post]
// @Security BearerAuth
func (h *FileHandler) UploadMultiple(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 权限检查：验证用户是否有文件上传权限
	hasPermission, err := auth.CheckUserPermission(userID.(uint), model.ResourceFile, model.ActionCreate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "权限检查失败")
		return
	}
	if !hasPermission {
		response.Error(c, http.StatusForbidden, "无权限上传文件")
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.Error(c, http.StatusBadRequest, "没有选择文件")
		return
	}

	// 获取业务参数
	businessType := model.BusinessType(c.PostForm("business_type"))
	if businessType == "" {
		businessType = model.BusinessTypeOther
	}

	var businessID *uint
	if businessIDStr := c.PostForm("business_id"); businessIDStr != "" {
		if id, err := strconv.ParseUint(businessIDStr, 10, 32); err == nil {
			businessIDUint := uint(id)
			businessID = &businessIDUint
		}
	}

	isPublic := c.PostForm("is_public") == "true"

	// 验证所有文件
	if err := h.validator.ValidateMultipleFiles(files); err != nil {
		response.Error(c, http.StatusBadRequest, "文件验证失败: "+err.Error())
		return
	}

	// 上传文件
	results, err := h.uploadService.UploadMultipleFiles(
		files,
		userID.(uint),
		businessType,
		businessID,
		isPublic,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "文件上传失败: "+err.Error())
		return
	}

	response.Success(c, "文件上传成功", results)
}

// GetFile 获取文件信息
// @Summary 获取文件信息
// @Description 根据UUID获取文件详细信息
// @Tags 文件管理
// @Produce json
// @Param uuid path string true "文件UUID"
// @Success 200 {object} response.Response{data=model.File}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/files/{uuid} [get]
// @Security BearerAuth
func (h *FileHandler) GetFile(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		response.Error(c, http.StatusBadRequest, "文件UUID不能为空")
		return
	}

	file, err := h.uploadService.GetFileByUUID(uuid)
	if err != nil {
		response.Error(c, http.StatusNotFound, "文件不存在")
		return
	}

	response.Success(c, "获取文件信息成功", file)
}

// DownloadPublicFile 下载公开文件
// @Summary 下载公开文件
// @Description 下载公开访问的文件
// @Tags 文件管理
// @Produce application/octet-stream
// @Param uuid path string true "文件UUID"
// @Success 200 {file} binary
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/files/public/{uuid} [get]
func (h *FileHandler) DownloadPublicFile(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		response.Error(c, http.StatusBadRequest, "文件UUID不能为空")
		return
	}

	file, err := h.uploadService.GetFileByUUID(uuid)
	if err != nil {
		response.Error(c, http.StatusNotFound, "文件不存在")
		return
	}

	// 检查是否为公开文件
	if !file.IsPublic {
		response.Error(c, http.StatusForbidden, "文件不允许公开访问")
		return
	}

	h.serveFile(c, file)
}

// DownloadPrivateFile 下载私有文件
// @Summary 下载私有文件
// @Description 下载需要认证的私有文件
// @Tags 文件管理
// @Produce application/octet-stream
// @Param uuid path string true "文件UUID"
// @Success 200 {file} binary
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/files/private/{uuid} [get]
// @Security BearerAuth
func (h *FileHandler) DownloadPrivateFile(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		response.Error(c, http.StatusBadRequest, "文件UUID不能为空")
		return
	}

	file, err := h.uploadService.GetFileByUUID(uuid)
	if err != nil {
		response.Error(c, http.StatusNotFound, "文件不存在")
		return
	}

	// 检查访问权限（文件所有者或管理员）
	userRole, _ := c.Get("user_role")
	if file.UploadUserID != userID.(uint) && userRole != model.RoleAdmin {
		response.Error(c, http.StatusForbidden, "无权限访问此文件")
		return
	}

	h.serveFile(c, file)
}

// ListFiles 获取文件列表
// @Summary 获取文件列表
// @Description 获取用户的文件列表
// @Tags 文件管理
// @Produce json
// @Param business_type query string false "业务类型" Enums(avatar,product,store,other)
// @Param business_id query int false "关联业务ID"
// @Param file_type query string false "文件类型" Enums(image,document,video,other)
// @Param is_public query bool false "是否公开"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResult{list=[]model.FileInfo}}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/files [get]
// @Security BearerAuth
func (h *FileHandler) ListFiles(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 解析查询参数
	var req model.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 获取文件列表
	files, total, err := h.uploadService.GetFilesByUser(
		userID.(uint),
		req.BusinessType,
		req.BusinessID,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取文件列表失败: "+err.Error())
		return
	}

	// 构造分页响应
	pageResult := response.PageResult{
		List:     files,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	response.Success(c, "获取文件列表成功", pageResult)
}

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定的文件
// @Tags 文件管理
// @Produce json
// @Param uuid path string true "文件UUID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/files/{uuid} [delete]
// @Security BearerAuth
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 权限检查：验证用户是否有文件删除权限
	hasPermission, err := auth.CheckUserPermission(userID.(uint), model.ResourceFile, model.ActionDelete)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "权限检查失败")
		return
	}
	if !hasPermission {
		response.Error(c, http.StatusForbidden, "无权限删除文件")
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		response.Error(c, http.StatusBadRequest, "文件UUID不能为空")
		return
	}

	// 删除文件
	if err := h.uploadService.DeleteFile(uuid, userID.(uint)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除文件失败: "+err.Error())
		return
	}

	response.Success(c, "文件删除成功", nil)
}

// serveFile 提供文件下载服务
func (h *FileHandler) serveFile(c *gin.Context, file *model.File) {
	// 检查文件是否存在
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		response.Error(c, http.StatusNotFound, "文件不存在")
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+file.OriginalName)
	c.Header("Content-Type", file.MimeType)

	// 发送文件
	c.File(file.FilePath)
}
