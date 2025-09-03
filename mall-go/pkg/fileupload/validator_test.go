package fileupload

import (
	"bytes"
	"mime/multipart"
	"testing"

	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
)

// createTestFileHeader 创建测试文件头
func createTestFileHeader(filename, contentType string, size int64, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", filename)
	if content != nil {
		part.Write(content)
	} else {
		// 写入指定大小的内容
		data := make([]byte, size)
		for i := range data {
			data[i] = 'A'
		}
		part.Write(data)
	}
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)

	files := form.File["file"]
	if len(files) == 0 {
		// 创建一个空的文件头用于测试
		fileHeader := &multipart.FileHeader{
			Filename: filename,
			Size:     size,
			Header:   make(map[string][]string),
		}
		fileHeader.Header.Set("Content-Type", contentType)
		return fileHeader
	}

	fileHeader := files[0]
	fileHeader.Header.Set("Content-Type", contentType)
	fileHeader.Size = size

	return fileHeader
}

// TestNewFileValidator 测试创建文件验证器
func TestNewFileValidator(t *testing.T) {
	validator := NewFileValidator()

	assert.NotNil(t, validator)
	assert.Equal(t, int64(model.MaxImageSize), validator.maxFileSize)
	assert.NotEmpty(t, validator.allowedTypes)
	assert.NotEmpty(t, validator.allowedMimes)
	assert.NotEmpty(t, validator.dangerousExts)
}

// TestValidateFile_ValidImage 测试验证有效图片文件
func TestValidateFile_ValidImage(t *testing.T) {
	validator := NewFileValidator()

	// 创建有效的JPEG文件头
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	fileHeader := createTestFileHeader("test.jpg", "image/jpeg", 1024, jpegHeader)

	err := validator.ValidateFile(fileHeader)
	assert.NoError(t, err)
}

// TestValidateFile_ValidPNG 测试验证有效PNG文件
func TestValidateFile_ValidPNG(t *testing.T) {
	validator := NewFileValidator()

	// 创建有效的PNG文件头
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fileHeader := createTestFileHeader("test.png", "image/png", 2048, pngHeader)

	err := validator.ValidateFile(fileHeader)
	assert.NoError(t, err)
}

// TestValidateFile_EmptyFilename 测试空文件名
func TestValidateFile_EmptyFilename(t *testing.T) {
	validator := NewFileValidator()

	fileHeader := createTestFileHeader("", "image/jpeg", 1024, nil)

	err := validator.ValidateFile(fileHeader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件名不能为空")
}

// TestValidateFile_InvalidSize 测试无效文件大小
func TestValidateFile_InvalidSize(t *testing.T) {
	validator := NewFileValidator()

	fileHeader := createTestFileHeader("test.jpg", "image/jpeg", 0, nil)

	err := validator.ValidateFile(fileHeader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件大小无效")
}

// TestValidateFile_ExceedsMaxSize 测试文件大小超限
func TestValidateFile_ExceedsMaxSize(t *testing.T) {
	validator := NewFileValidator()

	// 创建超过图片大小限制的文件
	fileHeader := createTestFileHeader("large.jpg", "image/jpeg", model.MaxImageSize+1, nil)

	err := validator.ValidateFile(fileHeader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件大小超过限制")
}

// TestValidateFile_UnsupportedExtension 测试不支持的扩展名
func TestValidateFile_UnsupportedExtension(t *testing.T) {
	validator := NewFileValidator()

	fileHeader := createTestFileHeader("test.xyz", "application/octet-stream", 1024, nil)

	err := validator.ValidateFile(fileHeader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的文件类型")
}

// TestValidateFile_DangerousExtension 测试危险扩展名
func TestValidateFile_DangerousExtension(t *testing.T) {
	validator := NewFileValidator()

	fileHeader := createTestFileHeader("malware.exe", "application/octet-stream", 1024, nil)

	err := validator.ValidateFile(fileHeader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "禁止上传的文件类型")
}

// TestValidateFile_UnsupportedMimeType 测试不支持的MIME类型
func TestValidateFile_UnsupportedMimeType(t *testing.T) {
	validator := NewFileValidator()

	fileHeader := createTestFileHeader("test.jpg", "application/unknown", 1024, nil)

	err := validator.ValidateFile(fileHeader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的MIME类型")
}

// TestValidateFileName_LongName 测试过长文件名
func TestValidateFileName_LongName(t *testing.T) {
	validator := NewFileValidator()

	// 创建超过255字符的文件名
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	longName = append(longName, []byte(".jpg")...)

	err := validator.validateFileName(string(longName))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件名过长")
}

// TestValidateFileName_DangerousChars 测试危险字符
func TestValidateFileName_DangerousChars(t *testing.T) {
	validator := NewFileValidator()

	dangerousNames := []string{
		"../test.jpg",
		"test/file.jpg",
		"test\\file.jpg",
		"test:file.jpg",
		"test*file.jpg",
		"test?file.jpg",
		"test\"file.jpg",
		"test<file.jpg",
		"test>file.jpg",
		"test|file.jpg",
	}

	for _, name := range dangerousNames {
		err := validator.validateFileName(name)
		assert.Error(t, err, "应该拒绝危险文件名: %s", name)
		assert.Contains(t, err.Error(), "文件名包含非法字符")
	}
}

// TestValidateFileName_ReservedNames 测试保留文件名
func TestValidateFileName_ReservedNames(t *testing.T) {
	validator := NewFileValidator()

	reservedNames := []string{
		"CON.jpg",
		"PRN.txt",
		"AUX.pdf",
		"NUL.doc",
		"COM1.jpg",
		"LPT1.txt",
	}

	for _, name := range reservedNames {
		err := validator.validateFileName(name)
		assert.Error(t, err, "应该拒绝保留文件名: %s", name)
		assert.Contains(t, err.Error(), "文件名使用了系统保留名称")
	}
}

// TestValidateMagicNumber_JPEG 测试JPEG魔数验证
func TestValidateMagicNumber_JPEG(t *testing.T) {
	validator := NewFileValidator()

	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	result := validator.validateMagicNumber(jpegHeader)
	assert.True(t, result)
}

// TestValidateMagicNumber_PNG 测试PNG魔数验证
func TestValidateMagicNumber_PNG(t *testing.T) {
	validator := NewFileValidator()

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	result := validator.validateMagicNumber(pngHeader)
	assert.True(t, result)
}

// TestValidateMagicNumber_PDF 测试PDF魔数验证
func TestValidateMagicNumber_PDF(t *testing.T) {
	validator := NewFileValidator()

	pdfHeader := []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}
	result := validator.validateMagicNumber(pdfHeader)
	assert.True(t, result)
}

// TestDetectMaliciousContent_Script 测试检测脚本内容
func TestDetectMaliciousContent_Script(t *testing.T) {
	validator := NewFileValidator()

	maliciousContents := []string{
		"<script>alert('xss')</script>",
		"javascript:void(0)",
		"<?php system($_GET['cmd']); ?>",
		"eval(base64_decode('...'))",
		"document.write('<script>...')",
	}

	for _, content := range maliciousContents {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte(content))
		writer.Close()

		reader := multipart.NewReader(body, writer.Boundary())
		form, _ := reader.ReadForm(32 << 20)
		file, _ := form.File["file"][0].Open()

		err := validator.detectMaliciousContent(file)
		assert.Error(t, err, "应该检测到恶意内容: %s", content)
		assert.Contains(t, err.Error(), "检测到可疑内容")

		file.Close()
	}
}

// TestDetectMaliciousContent_Executable 测试检测可执行文件
func TestDetectMaliciousContent_Executable(t *testing.T) {
	validator := NewFileValidator()

	// Windows PE文件头
	peHeader := []byte{0x4D, 0x5A, 0x90, 0x00}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.exe")
	part.Write(peHeader)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)
	file, _ := form.File["file"][0].Open()

	err := validator.detectMaliciousContent(file)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "检测到可执行文件")

	file.Close()
}

// TestValidateMultipleFiles 测试验证多个文件
func TestValidateMultipleFiles(t *testing.T) {
	validator := NewFileValidator()

	// 创建有效的文件列表
	validFiles := []*multipart.FileHeader{
		createTestFileHeader("test1.jpg", "image/jpeg", 1024, []byte{0xFF, 0xD8, 0xFF}),
		createTestFileHeader("test2.png", "image/png", 2048, []byte{0x89, 0x50, 0x4E, 0x47}),
	}

	err := validator.ValidateMultipleFiles(validFiles)
	assert.NoError(t, err)

	// 测试空文件列表
	err = validator.ValidateMultipleFiles([]*multipart.FileHeader{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "没有选择文件")

	// 测试文件数量超限
	tooManyFiles := make([]*multipart.FileHeader, model.MaxFileCount+1)
	for i := range tooManyFiles {
		tooManyFiles[i] = createTestFileHeader("test.jpg", "image/jpeg", 1024, nil)
	}

	err = validator.ValidateMultipleFiles(tooManyFiles)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件数量超过限制")
}

// TestGetFileTypeFromHeader 测试从文件头获取文件类型
func TestGetFileTypeFromHeader(t *testing.T) {
	validator := NewFileValidator()

	testCases := []struct {
		header   []byte
		expected string
	}{
		{[]byte{0xFF, 0xD8, 0xFF}, "jpeg"},
		{[]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "png"},
		{[]byte{0x47, 0x49, 0x46, 0x38}, "gif"},
		{[]byte{0x25, 0x50, 0x44, 0x46}, "pdf"},
		{[]byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}, "webp"},
		{[]byte{0x00, 0x00, 0x00}, "unknown"},
	}

	for _, tc := range testCases {
		result := validator.GetFileTypeFromHeader(tc.header)
		assert.Equal(t, tc.expected, result, "文件头 %v 应该返回类型 %s", tc.header, tc.expected)
	}
}

// TestIsAllowedExtension 测试扩展名验证
func TestIsAllowedExtension(t *testing.T) {
	validator := NewFileValidator()

	// 测试允许的扩展名
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt"}
	for _, ext := range allowedExts {
		result := validator.isAllowedExtension(ext)
		assert.True(t, result, "扩展名 %s 应该被允许", ext)
	}

	// 测试不允许的扩展名
	disallowedExts := []string{".exe", ".bat", ".cmd", ".xyz"}
	for _, ext := range disallowedExts {
		result := validator.isAllowedExtension(ext)
		assert.False(t, result, "扩展名 %s 应该被拒绝", ext)
	}
}

// TestIsDangerousExtension 测试危险扩展名检查
func TestIsDangerousExtension(t *testing.T) {
	validator := NewFileValidator()

	// 测试危险扩展名
	dangerousExts := []string{".exe", ".bat", ".cmd", ".php", ".jsp", ".asp"}
	for _, ext := range dangerousExts {
		result := validator.isDangerousExtension(ext)
		assert.True(t, result, "扩展名 %s 应该被识别为危险", ext)
	}

	// 测试安全扩展名
	safeExts := []string{".jpg", ".png", ".pdf", ".txt"}
	for _, ext := range safeExts {
		result := validator.isDangerousExtension(ext)
		assert.False(t, result, "扩展名 %s 应该被识别为安全", ext)
	}
}

// TestIsAllowedMimeType 测试MIME类型验证
func TestIsAllowedMimeType(t *testing.T) {
	validator := NewFileValidator()

	// 测试允许的MIME类型
	allowedMimes := []string{"image/jpeg", "image/png", "application/pdf", "text/plain"}
	for _, mime := range allowedMimes {
		result := validator.isAllowedMimeType(mime)
		assert.True(t, result, "MIME类型 %s 应该被允许", mime)
	}

	// 测试不允许的MIME类型
	disallowedMimes := []string{"application/x-executable", "application/unknown"}
	for _, mime := range disallowedMimes {
		result := validator.isAllowedMimeType(mime)
		assert.False(t, result, "MIME类型 %s 应该被拒绝", mime)
	}
}
