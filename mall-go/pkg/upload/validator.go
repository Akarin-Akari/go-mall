package upload

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// FileValidator 文件验证器
type FileValidator struct {
	config *UploadConfig
}

// NewFileValidator 创建文件验证器
func NewFileValidator(config *UploadConfig) *FileValidator {
	return &FileValidator{
		config: config,
	}
}

// ValidateFile 验证文件
func (fv *FileValidator) ValidateFile(filename string, reader io.Reader, size int64) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Filename: filename,
		Size:     size,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// 读取文件头部用于检测
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 创建新的reader，包含已读取的数据
	fullReader := io.MultiReader(bytes.NewReader(buffer[:n]), reader)

	// 检测文件类型
	contentType := http.DetectContentType(buffer[:n])
	result.ContentType = contentType

	// 1. 文件大小验证
	if err := fv.validateFileSize(size, result); err != nil {
		return result, err
	}

	// 2. 文件扩展名验证
	if err := fv.validateFileExtension(filename, result); err != nil {
		return result, err
	}

	// 3. 文件类型验证
	if err := fv.validateContentType(contentType, result); err != nil {
		return result, err
	}

	// 4. 安全检查
	if fv.config.EnableSecurity {
		if err := fv.performSecurityCheck(filename, contentType, buffer[:n], result); err != nil {
			return result, err
		}
	}

	// 5. 图片特殊验证
	if fv.isImageFile(contentType) {
		if err := fv.validateImage(fullReader, result); err != nil {
			return result, err
		}
	}

	// 设置最终验证结果
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid       bool       `json:"valid"`                // 是否通过验证
	Filename    string     `json:"filename"`             // 文件名
	Size        int64      `json:"size"`                 // 文件大小
	ContentType string     `json:"content_type"`         // 文件类型
	Errors      []string   `json:"errors"`               // 错误信息
	Warnings    []string   `json:"warnings"`             // 警告信息
	ImageInfo   *ImageInfo `json:"image_info,omitempty"` // 图片信息
}

// ImageInfo 图片信息
type ImageInfo struct {
	Width  int    `json:"width"`  // 宽度
	Height int    `json:"height"` // 高度
	Format string `json:"format"` // 格式
}

// validateFileSize 验证文件大小
func (fv *FileValidator) validateFileSize(size int64, result *ValidationResult) error {
	if size > fv.config.MaxFileSize {
		result.Errors = append(result.Errors, fmt.Sprintf("文件大小超过限制，最大允许 %d 字节", fv.config.MaxFileSize))
		return nil
	}

	if size == 0 {
		result.Errors = append(result.Errors, "文件为空")
		return nil
	}

	return nil
}

// validateFileExtension 验证文件扩展名
func (fv *FileValidator) validateFileExtension(filename string, result *ValidationResult) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// 检查是否为禁止的扩展名
	for _, forbiddenExt := range fv.config.Security.ForbiddenExts {
		if ext == strings.ToLower(forbiddenExt) {
			result.Errors = append(result.Errors, fmt.Sprintf("文件类型被禁止: %s", ext))
			return nil
		}
	}

	// 检查是否为允许的扩展名
	if len(fv.config.AllowedExts) > 0 {
		allowed := false
		for _, allowedExt := range fv.config.AllowedExts {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Errors = append(result.Errors, fmt.Sprintf("不支持的文件扩展名: %s", ext))
			return nil
		}
	}

	return nil
}

// validateContentType 验证文件类型
func (fv *FileValidator) validateContentType(contentType string, result *ValidationResult) error {
	if len(fv.config.AllowedTypes) > 0 {
		allowed := false
		for _, allowedType := range fv.config.AllowedTypes {
			if contentType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Errors = append(result.Errors, fmt.Sprintf("不支持的文件类型: %s", contentType))
			return nil
		}
	}

	return nil
}

// performSecurityCheck 执行安全检查
func (fv *FileValidator) performSecurityCheck(filename, contentType string, header []byte, result *ValidationResult) error {
	// 1. 检查文件头部是否与扩展名匹配
	if err := fv.checkFileHeaderMismatch(filename, contentType, result); err != nil {
		return err
	}

	// 2. 检查可执行文件特征
	if err := fv.checkExecutableFile(header, result); err != nil {
		return err
	}

	// 3. 检查脚本文件特征
	if err := fv.checkScriptFile(header, result); err != nil {
		return err
	}

	// 4. 检查恶意文件特征
	if err := fv.checkMaliciousPatterns(header, result); err != nil {
		return err
	}

	return nil
}

// checkFileHeaderMismatch 检查文件头部与扩展名是否匹配
func (fv *FileValidator) checkFileHeaderMismatch(filename, contentType string, result *ValidationResult) error {
	ext := strings.ToLower(filepath.Ext(filename))
	expectedType := mime.TypeByExtension(ext)

	if expectedType != "" && contentType != expectedType {
		// 某些情况下可能不完全匹配，给出警告而不是错误
		if !fv.isAcceptableMismatch(ext, contentType, expectedType) {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("文件类型与扩展名不匹配: 扩展名 %s 期望 %s，实际 %s", ext, expectedType, contentType))
		}
	}

	return nil
}

// isAcceptableMismatch 检查是否为可接受的类型不匹配
func (fv *FileValidator) isAcceptableMismatch(ext, actual, expected string) bool {
	// JPEG文件的不同MIME类型
	if ext == ".jpg" || ext == ".jpeg" {
		return actual == "image/jpeg" || expected == "image/jpeg"
	}

	// 其他可接受的不匹配情况
	acceptableMismatches := map[string][]string{
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".webp": {"image/webp"},
		".pdf":  {"application/pdf"},
	}

	if acceptable, exists := acceptableMismatches[ext]; exists {
		for _, acceptableType := range acceptable {
			if actual == acceptableType {
				return true
			}
		}
	}

	return false
}

// checkExecutableFile 检查可执行文件
func (fv *FileValidator) checkExecutableFile(header []byte, result *ValidationResult) error {
	// Windows可执行文件特征
	if len(header) >= 2 && header[0] == 0x4D && header[1] == 0x5A { // MZ header
		result.Errors = append(result.Errors, "检测到Windows可执行文件")
		return nil
	}

	// ELF可执行文件特征
	if len(header) >= 4 && header[0] == 0x7F && header[1] == 0x45 &&
		header[2] == 0x4C && header[3] == 0x46 { // ELF header
		result.Errors = append(result.Errors, "检测到Linux可执行文件")
		return nil
	}

	return nil
}

// checkScriptFile 检查脚本文件
func (fv *FileValidator) checkScriptFile(header []byte, result *ValidationResult) error {
	headerStr := string(header)

	// 检查脚本文件特征
	scriptPatterns := []string{
		"#!/bin/sh",
		"#!/bin/bash",
		"#!/usr/bin/env",
		"<?php",
		"<script",
		"javascript:",
		"vbscript:",
	}

	for _, pattern := range scriptPatterns {
		if strings.Contains(strings.ToLower(headerStr), strings.ToLower(pattern)) {
			result.Errors = append(result.Errors, fmt.Sprintf("检测到脚本文件特征: %s", pattern))
			return nil
		}
	}

	return nil
}

// checkMaliciousPatterns 检查恶意文件特征
func (fv *FileValidator) checkMaliciousPatterns(header []byte, result *ValidationResult) error {
	headerStr := strings.ToLower(string(header))

	// 恶意代码特征
	maliciousPatterns := []string{
		"eval(",
		"exec(",
		"system(",
		"shell_exec(",
		"passthru(",
		"base64_decode(",
		"<iframe",
		"<object",
		"<embed",
	}

	for _, pattern := range maliciousPatterns {
		if strings.Contains(headerStr, pattern) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("检测到可疑代码特征: %s", pattern))
		}
	}

	return nil
}

// validateImage 验证图片文件
func (fv *FileValidator) validateImage(reader io.Reader, result *ValidationResult) error {
	// 解码图片获取尺寸信息
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("无法解析图片: %v", err))
		return nil
	}

	// 设置图片信息
	result.ImageInfo = &ImageInfo{
		Width:  config.Width,
		Height: config.Height,
		Format: format,
	}

	// 检查图片尺寸
	if fv.config.Security.MaxImageWidth > 0 && config.Width > fv.config.Security.MaxImageWidth {
		result.Errors = append(result.Errors,
			fmt.Sprintf("图片宽度超过限制: %d > %d", config.Width, fv.config.Security.MaxImageWidth))
	}

	if fv.config.Security.MaxImageHeight > 0 && config.Height > fv.config.Security.MaxImageHeight {
		result.Errors = append(result.Errors,
			fmt.Sprintf("图片高度超过限制: %d > %d", config.Height, fv.config.Security.MaxImageHeight))
	}

	// 检查图片格式是否与文件类型匹配
	expectedFormat := fv.getExpectedImageFormat(result.ContentType)
	if expectedFormat != "" && format != expectedFormat {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("图片格式与文件类型不匹配: 格式 %s，类型 %s", format, result.ContentType))
	}

	return nil
}

// isImageFile 检查是否为图片文件
func (fv *FileValidator) isImageFile(contentType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/tiff",
	}

	for _, imageType := range imageTypes {
		if contentType == imageType {
			return true
		}
	}

	return false
}

// getExpectedImageFormat 根据MIME类型获取期望的图片格式
func (fv *FileValidator) getExpectedImageFormat(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return "jpeg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/bmp":
		return "bmp"
	case "image/tiff":
		return "tiff"
	default:
		return ""
	}
}

// 全局验证器
var globalValidator *FileValidator

// InitGlobalValidator 初始化全局验证器
func InitGlobalValidator(config *UploadConfig) {
	globalValidator = NewFileValidator(config)
}

// GetGlobalValidator 获取全局验证器
func GetGlobalValidator() *FileValidator {
	return globalValidator
}

// ValidateUploadFile 验证上传文件（全局函数）
func ValidateUploadFile(filename string, reader io.Reader, size int64) (*ValidationResult, error) {
	if globalValidator == nil {
		return nil, fmt.Errorf("文件验证器未初始化")
	}
	return globalValidator.ValidateFile(filename, reader, size)
}
