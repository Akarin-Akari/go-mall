package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *UploadConfig
	configFile string
	mutex      sync.RWMutex
	watchers   []ConfigWatcher
}

// ConfigWatcher 配置变更监听器
type ConfigWatcher interface {
	OnConfigChanged(config *UploadConfig)
}

// ConfigWatcherFunc 配置变更监听器函数类型
type ConfigWatcherFunc func(config *UploadConfig)

// OnConfigChanged 实现ConfigWatcher接口
func (f ConfigWatcherFunc) OnConfigChanged(config *UploadConfig) {
	f(config)
}

var (
	globalConfigManager *ConfigManager
	once                sync.Once
)

// NewConfigManager 创建配置管理器
func NewConfigManager(configFile string) (*ConfigManager, error) {
	config, err := LoadConfigFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %v", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 确保上传目录存在
	if config.StorageType == StorageTypeLocal {
		if err := os.MkdirAll(config.Local.UploadPath, 0755); err != nil {
			return nil, fmt.Errorf("创建上传目录失败: %v", err)
		}
	}

	manager := &ConfigManager{
		config:     config,
		configFile: configFile,
		watchers:   make([]ConfigWatcher, 0),
	}

	return manager, nil
}

// GetGlobalConfigManager 获取全局配置管理器
func GetGlobalConfigManager() *ConfigManager {
	return globalConfigManager
}

// InitGlobalConfigManager 初始化全局配置管理器
func InitGlobalConfigManager(configFile string) error {
	var err error
	once.Do(func() {
		globalConfigManager, err = NewConfigManager(configFile)
	})
	return err
}

// InitGlobalConfigManagerWithEnv 使用环境变量初始化全局配置管理器
func InitGlobalConfigManagerWithEnv() error {
	var err error
	once.Do(func() {
		config := LoadConfigFromEnv()
		if validateErr := config.Validate(); validateErr != nil {
			err = fmt.Errorf("配置验证失败: %v", validateErr)
			return
		}

		// 确保上传目录存在
		if config.StorageType == StorageTypeLocal {
			if mkdirErr := os.MkdirAll(config.Local.UploadPath, 0755); mkdirErr != nil {
				err = fmt.Errorf("创建上传目录失败: %v", mkdirErr)
				return
			}
		}

		globalConfigManager = &ConfigManager{
			config:   config,
			watchers: make([]ConfigWatcher, 0),
		}
	})
	return err
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *UploadConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 返回配置的副本，避免外部修改
	configCopy := *cm.config
	return &configCopy
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(newConfig *UploadConfig) error {
	// 验证新配置
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 备份旧配置
	oldConfig := cm.config

	// 更新配置
	cm.config = newConfig

	// 如果是本地存储，确保目录存在
	if newConfig.StorageType == StorageTypeLocal {
		if err := os.MkdirAll(newConfig.Local.UploadPath, 0755); err != nil {
			// 恢复旧配置
			cm.config = oldConfig
			return fmt.Errorf("创建上传目录失败: %v", err)
		}
	}

	// 保存配置到文件
	if cm.configFile != "" {
		if err := newConfig.SaveToFile(cm.configFile); err != nil {
			// 恢复旧配置
			cm.config = oldConfig
			return fmt.Errorf("保存配置文件失败: %v", err)
		}
	}

	// 通知监听器
	cm.notifyWatchers(newConfig)

	return nil
}

// ReloadConfig 重新加载配置
func (cm *ConfigManager) ReloadConfig() error {
	if cm.configFile == "" {
		return fmt.Errorf("未指定配置文件")
	}

	newConfig, err := LoadConfigFromFile(cm.configFile)
	if err != nil {
		return fmt.Errorf("重新加载配置失败: %v", err)
	}

	return cm.UpdateConfig(newConfig)
}

// AddWatcher 添加配置变更监听器
func (cm *ConfigManager) AddWatcher(watcher ConfigWatcher) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.watchers = append(cm.watchers, watcher)
}

// RemoveWatcher 移除配置变更监听器
func (cm *ConfigManager) RemoveWatcher(watcher ConfigWatcher) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for i, w := range cm.watchers {
		if w == watcher {
			cm.watchers = append(cm.watchers[:i], cm.watchers[i+1:]...)
			break
		}
	}
}

// notifyWatchers 通知所有监听器
func (cm *ConfigManager) notifyWatchers(config *UploadConfig) {
	for _, watcher := range cm.watchers {
		go watcher.OnConfigChanged(config)
	}
}

// GetStorageType 获取存储类型
func (cm *ConfigManager) GetStorageType() StorageType {
	config := cm.GetConfig()
	return config.StorageType
}

// GetMaxFileSize 获取最大文件大小
func (cm *ConfigManager) GetMaxFileSize() int64 {
	config := cm.GetConfig()
	return config.MaxFileSize
}

// GetMaxFiles 获取最大文件数
func (cm *ConfigManager) GetMaxFiles() int {
	config := cm.GetConfig()
	return config.MaxFiles
}

// IsAllowedType 检查文件类型是否允许
func (cm *ConfigManager) IsAllowedType(contentType string) bool {
	config := cm.GetConfig()
	return config.IsAllowedType(contentType)
}

// IsAllowedExt 检查文件扩展名是否允许
func (cm *ConfigManager) IsAllowedExt(filename string) bool {
	config := cm.GetConfig()
	return config.IsAllowedExt(filename)
}

// IsForbiddenExt 检查文件扩展名是否被禁止
func (cm *ConfigManager) IsForbiddenExt(filename string) bool {
	config := cm.GetConfig()
	return config.IsForbiddenExt(filename)
}

// GetUploadPath 获取上传路径
func (cm *ConfigManager) GetUploadPath(subPath string) string {
	config := cm.GetConfig()
	return config.GetUploadPath(subPath)
}

// GetFileURL 获取文件访问URL
func (cm *ConfigManager) GetFileURL(filePath string) string {
	config := cm.GetConfig()
	return config.GetFileURL(filePath)
}

// GenerateFileName 生成文件名
func (cm *ConfigManager) GenerateFileName(originalName string) string {
	config := cm.GetConfig()
	return config.GenerateFileName(originalName)
}

// IsSecurityEnabled 检查是否启用安全检查
func (cm *ConfigManager) IsSecurityEnabled() bool {
	config := cm.GetConfig()
	return config.EnableSecurity
}

// IsThumbnailEnabled 检查是否启用缩略图
func (cm *ConfigManager) IsThumbnailEnabled() bool {
	config := cm.GetConfig()
	return config.EnableThumbnail
}

// GetThumbnailSizes 获取缩略图尺寸
func (cm *ConfigManager) GetThumbnailSizes() []ThumbnailSize {
	config := cm.GetConfig()
	return config.Thumbnail.Sizes
}

// ValidateFileSize 验证文件大小
func (cm *ConfigManager) ValidateFileSize(size int64) error {
	maxSize := cm.GetMaxFileSize()
	if size > maxSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %d 字节", maxSize)
	}
	return nil
}

// ValidateFileCount 验证文件数量
func (cm *ConfigManager) ValidateFileCount(count int) error {
	maxFiles := cm.GetMaxFiles()
	if count > maxFiles {
		return fmt.Errorf("文件数量超过限制，最大允许 %d 个文件", maxFiles)
	}
	return nil
}

// ValidateFile 验证文件
func (cm *ConfigManager) ValidateFile(filename, contentType string, size int64) error {
	// 检查文件大小
	if err := cm.ValidateFileSize(size); err != nil {
		return err
	}

	// 检查是否为禁止的扩展名
	if cm.IsForbiddenExt(filename) {
		return fmt.Errorf("文件类型被禁止: %s", filepath.Ext(filename))
	}

	// 检查文件扩展名
	if !cm.IsAllowedExt(filename) {
		return fmt.Errorf("不支持的文件扩展名: %s", filepath.Ext(filename))
	}

	// 检查文件类型
	if contentType != "" && !cm.IsAllowedType(contentType) {
		return fmt.Errorf("不支持的文件类型: %s", contentType)
	}

	return nil
}

// GetConfigSummary 获取配置摘要
func (cm *ConfigManager) GetConfigSummary() map[string]interface{} {
	config := cm.GetConfig()
	return map[string]interface{}{
		"storage_type":     config.StorageType,
		"max_file_size":    config.MaxFileSize,
		"max_files":        config.MaxFiles,
		"allowed_types":    config.AllowedTypes,
		"allowed_exts":     config.AllowedExts,
		"enable_security":  config.EnableSecurity,
		"enable_thumbnail": config.EnableThumbnail,
		"config_file":      cm.configFile,
		"last_updated":     time.Now().Format(time.RFC3339),
	}
}

// 全局便捷函数

// GetConfig 获取全局配置
func GetConfig() *UploadConfig {
	if globalConfigManager == nil {
		return DefaultUploadConfig()
	}
	return globalConfigManager.GetConfig()
}

// ValidateFile 验证文件（全局函数）
func ValidateFile(filename, contentType string, size int64) error {
	if globalConfigManager == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	return globalConfigManager.ValidateFile(filename, contentType, size)
}

// GenerateFileName 生成文件名（全局函数）
func GenerateFileName(originalName string) string {
	if globalConfigManager == nil {
		config := DefaultUploadConfig()
		return config.GenerateFileName(originalName)
	}
	return globalConfigManager.GenerateFileName(originalName)
}

// GetFileURL 获取文件URL（全局函数）
func GetFileURL(filePath string) string {
	if globalConfigManager == nil {
		return filePath
	}
	return globalConfigManager.GetFileURL(filePath)
}

// GetUploadPath 获取上传路径（全局函数）
func GetUploadPath(subPath string) string {
	if globalConfigManager == nil {
		return subPath
	}
	return globalConfigManager.GetUploadPath(subPath)
}
