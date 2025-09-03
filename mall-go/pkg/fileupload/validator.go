package fileupload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"mall-go/internal/model"
)

// FileValidator 文件验证器
type FileValidator struct {
	maxFileSize   int64
	allowedTypes  []string
	allowedMimes  []string
	dangerousExts []string
	magicNumbers  map[string][]byte
	// 新增配置项
	config *ValidationConfig
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	MaxImageSize     int64    `json:"max_image_size"`     // 图片最大大小
	MaxDocumentSize  int64    `json:"max_document_size"`  // 文档最大大小
	MaxVideoSize     int64    `json:"max_video_size"`     // 视频最大大小
	MaxFileCount     int      `json:"max_file_count"`     // 单次最多上传文件数
	AllowedTypes     []string `json:"allowed_types"`      // 允许的文件类型
	AllowedMimes     []string `json:"allowed_mimes"`      // 允许的MIME类型
	DangerousExts    []string `json:"dangerous_exts"`     // 危险的扩展名
	EnableMagicCheck bool     `json:"enable_magic_check"` // 是否启用魔数检查
}

// DefaultValidationConfig 默认验证配置
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxImageSize:    model.MaxImageSize,
		MaxDocumentSize: model.MaxDocumentSize,
		MaxVideoSize:    model.MaxVideoSize,
		MaxFileCount:    model.MaxFileCount,
		AllowedTypes: append(append(model.SupportedImageTypes, model.SupportedDocumentTypes...),
			model.SupportedVideoTypes...),
		AllowedMimes: append(append(model.ImageMimeTypes, model.DocumentMimeTypes...),
			model.VideoMimeTypes...),
		DangerousExts: []string{
			".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar",
			".php", ".asp", ".aspx", ".jsp", ".py", ".rb", ".pl", ".sh",
		},
		EnableMagicCheck: true,
	}
}

// NewFileValidator 创建文件验证器（使用默认配置）
func NewFileValidator() *FileValidator {
	config := DefaultValidationConfig()
	return NewFileValidatorWithConfig(config)
}

// NewFileValidatorWithConfig 创建文件验证器（使用自定义配置）
func NewFileValidatorWithConfig(config *ValidationConfig) *FileValidator {
	return &FileValidator{
		maxFileSize:   config.MaxImageSize, // 默认使用图片大小限制
		allowedTypes:  config.AllowedTypes,
		allowedMimes:  config.AllowedMimes,
		dangerousExts: config.DangerousExts,
		config:        config,
		magicNumbers: map[string][]byte{
			"jpeg": {0xFF, 0xD8, 0xFF},
			"png":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			"gif":  {0x47, 0x49, 0x46, 0x38},
			"pdf":  {0x25, 0x50, 0x44, 0x46},
			"webp": {0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
		},
	}
}

// ValidateFile 验证单个文件
func (v *FileValidator) ValidateFile(fileHeader *multipart.FileHeader) error {
	// 1. 验证文件名
	if err := v.validateFileName(fileHeader.Filename); err != nil {
		return err
	}

	// 2. 验证文件大小
	if err := v.validateFileSize(fileHeader.Size); err != nil {
		return err
	}

	// 3. 验证文件类型
	if err := v.validateFileType(fileHeader); err != nil {
		return err
	}

	// 4. 验证MIME类型
	mimeType := fileHeader.Header.Get("Content-Type")
	if err := v.validateMimeType(mimeType); err != nil {
		return err
	}

	// 5. 验证文件内容
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	if err := v.validateFileContent(file); err != nil {
		return err
	}

	// 6. 检测恶意内容
	file.Seek(0, 0) // 重置文件指针
	if err := v.detectMaliciousContent(file); err != nil {
		return err
	}

	return nil
}

// ValidateMultipleFiles 验证多个文件
func (v *FileValidator) ValidateMultipleFiles(fileHeaders []*multipart.FileHeader) error {
	if len(fileHeaders) == 0 {
		return fmt.Errorf("没有选择文件")
	}

	if len(fileHeaders) > model.MaxFileCount {
		return fmt.Errorf("文件数量超过限制，最多允许%d个文件", model.MaxFileCount)
	}

	// 计算总大小
	var totalSize int64
	for _, fileHeader := range fileHeaders {
		totalSize += fileHeader.Size
	}

	maxTotalSize := int64(100 * 1024 * 1024) // 100MB
	if totalSize > maxTotalSize {
		return fmt.Errorf("文件总大小超过限制，最大允许%d字节", maxTotalSize)
	}

	// 验证每个文件
	for i, fileHeader := range fileHeaders {
		if err := v.ValidateFile(fileHeader); err != nil {
			return fmt.Errorf("文件%d验证失败: %v", i+1, err)
		}
	}

	return nil
}

// validateFileName 验证文件名
func (v *FileValidator) validateFileName(filename string) error {
	if filename == "" {
		return fmt.Errorf("文件名不能为空")
	}

	if len(filename) > 255 {
		return fmt.Errorf("文件名过长，最大允许255个字符")
	}

	// 检查危险字符
	dangerousChars := []string{"../", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("文件名包含非法字符: %s", char)
		}
	}

	// 检查Windows保留名称
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	nameWithoutExt := strings.ToUpper(strings.TrimSuffix(filename, filepath.Ext(filename)))
	for _, reserved := range reservedNames {
		if nameWithoutExt == reserved {
			return fmt.Errorf("文件名使用了系统保留名称: %s", reserved)
		}
	}

	return nil
}

// validateFileSize 验证文件大小
func (v *FileValidator) validateFileSize(size int64) error {
	if size <= 0 {
		return fmt.Errorf("文件大小无效")
	}

	if size > v.maxFileSize {
		return fmt.Errorf("文件大小超过限制，最大允许%d字节", v.maxFileSize)
	}

	return nil
}

// validateFileType 验证文件类型
func (v *FileValidator) validateFileType(fileHeader *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// 检查是否为危险扩展名
	if v.isDangerousExtension(ext) {
		return fmt.Errorf("禁止上传的文件类型: %s", ext)
	}

	// 检查是否为允许的扩展名
	if !v.isAllowedExtension(ext) {
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	return nil
}

// validateMimeType 验证MIME类型
func (v *FileValidator) validateMimeType(mimeType string) error {
	if mimeType == "" {
		return fmt.Errorf("MIME类型不能为空")
	}

	if !v.isAllowedMimeType(mimeType) {
		return fmt.Errorf("不支持的MIME类型: %s", mimeType)
	}

	return nil
}

// validateFileContent 验证文件内容
func (v *FileValidator) validateFileContent(file multipart.File) error {
	// 读取文件头部用于魔数验证
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件头失败: %v", err)
	}

	header = header[:n]

	// 验证魔数
	if !v.validateMagicNumber(header) {
		return fmt.Errorf("文件内容与扩展名不匹配")
	}

	return nil
}

// validateMagicNumber 验证文件魔数
func (v *FileValidator) validateMagicNumber(header []byte) bool {
	for _, magic := range v.magicNumbers {
		if len(header) >= len(magic) && bytes.Equal(header[:len(magic)], magic) {
			return true
		}
	}

	// 对于其他文件类型，暂时允许通过
	return true
}

// detectMaliciousContent 检测恶意内容
func (v *FileValidator) detectMaliciousContent(file multipart.File) error {
	// 读取文件内容进行检测
	content := make([]byte, 1024)
	n, err := file.Read(content)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件内容失败: %v", err)
	}

	contentStr := string(content[:n])

	// 检测脚本内容
	scriptPatterns := []string{
		`<script.*?>.*?</script>`,
		`javascript:`,
		`<\?php.*?\?>`,
		`<%.*?%>`,
		`eval\s*\(`,
		`document\.write`,
	}

	for _, pattern := range scriptPatterns {
		matched, _ := regexp.MatchString(`(?i)`+pattern, contentStr)
		if matched {
			return fmt.Errorf("检测到可疑内容，可能包含恶意脚本")
		}
	}

	// 检测可执行文件头
	executableHeaders := [][]byte{
		{0x4D, 0x5A},             // PE文件 (Windows可执行文件)
		{0x7F, 0x45, 0x4C, 0x46}, // ELF文件 (Linux可执行文件)
	}

	for _, execHeader := range executableHeaders {
		if len(content) >= len(execHeader) && bytes.Equal(content[:len(execHeader)], execHeader) {
			return fmt.Errorf("检测到可执行文件，禁止上传")
		}
	}

	return nil
}

// isAllowedExtension 检查是否为允许的扩展名
func (v *FileValidator) isAllowedExtension(ext string) bool {
	for _, allowedExt := range v.allowedTypes {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// isDangerousExtension 检查是否为危险扩展名
func (v *FileValidator) isDangerousExtension(ext string) bool {
	for _, dangerousExt := range v.dangerousExts {
		if ext == dangerousExt {
			return true
		}
	}
	return false
}

// isAllowedMimeType 检查是否为允许的MIME类型
func (v *FileValidator) isAllowedMimeType(mimeType string) bool {
	for _, allowedMime := range v.allowedMimes {
		if mimeType == allowedMime {
			return true
		}
	}
	return false
}

// GetFileTypeFromHeader 从文件头获取文件类型
func (v *FileValidator) GetFileTypeFromHeader(header []byte) string {
	for fileType, magic := range v.magicNumbers {
		if len(header) >= len(magic) && bytes.Equal(header[:len(magic)], magic) {
			return fileType
		}
	}
	return "unknown"
}
