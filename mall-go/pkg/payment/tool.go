package payment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// ConfigTool 配置管理工具
type ConfigTool struct {
	configPath   string
	templatePath string
	backupPath   string
}

// NewConfigTool 创建配置工具
func NewConfigTool(configPath string) *ConfigTool {
	dir := filepath.Dir(configPath)

	return &ConfigTool{
		configPath:   configPath,
		templatePath: filepath.Join(dir, "templates"),
		backupPath:   filepath.Join(dir, "backups"),
	}
}

// GenerateConfigForEnvironment 为指定环境生成配置文件
func (ct *ConfigTool) GenerateConfigForEnvironment(env string, force bool) error {
	logger.Info("生成环境配置文件",
		zap.String("environment", env),
		zap.String("config_path", ct.configPath),
		zap.Bool("force", force))

	// 检查文件是否已存在
	if !force {
		if _, err := os.Stat(ct.configPath); err == nil {
			return fmt.Errorf("配置文件已存在，使用 --force 参数覆盖: %s", ct.configPath)
		}
	}

	// 备份现有配置
	if err := ct.backupExistingConfig(); err != nil {
		logger.Warn("备份现有配置失败", zap.Error(err))
	}

	// 生成配置模板
	template := LoadTemplateByEnvironment(env)

	// 创建配置目录
	if err := os.MkdirAll(filepath.Dir(ct.configPath), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(ct.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 生成示例环境变量文件
	if err := ct.generateEnvExample(env); err != nil {
		logger.Warn("生成环境变量示例文件失败", zap.Error(err))
	}

	logger.Info("配置文件生成成功",
		zap.String("environment", env),
		zap.String("config_path", ct.configPath))

	return nil
}

// ValidateConfig 验证配置文件
func (ct *ConfigTool) ValidateConfig() (*ValidationReport, error) {
	logger.Info("验证配置文件", zap.String("config_path", ct.configPath))

	// 加载配置
	config, err := LoadConfigFromFile(ct.configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %v", err)
	}

	// 执行验证
	errors := ValidateEnvironmentConfig(config)

	report := &ValidationReport{
		ConfigPath:  ct.configPath,
		Environment: config.Environment,
		IsValid:     len(errors) == 0,
		ErrorCount:  len(errors),
		Errors:      errors,
		ValidatedAt: time.Now(),
	}

	// 输出验证结果
	if report.IsValid {
		logger.Info("✅ 配置验证通过", zap.String("environment", config.Environment))
	} else {
		logger.Error("❌ 配置验证失败",
			zap.String("environment", config.Environment),
			zap.Int("error_count", len(errors)))

		for _, err := range errors {
			logger.Error("配置错误",
				zap.String("field", err.Field),
				zap.String("message", err.Message),
				zap.String("code", err.Code))
		}
	}

	return report, nil
}

// ValidationReport 验证报告
type ValidationReport struct {
	ConfigPath  string            `json:"config_path"`
	Environment string            `json:"environment"`
	IsValid     bool              `json:"is_valid"`
	ErrorCount  int               `json:"error_count"`
	Errors      []ValidationError `json:"errors"`
	ValidatedAt time.Time         `json:"validated_at"`
}

// MigrateConfig 配置迁移
func (ct *ConfigTool) MigrateConfig(fromVersion, toVersion string) error {
	logger.Info("迁移配置",
		zap.String("from_version", fromVersion),
		zap.String("to_version", toVersion))

	// 备份当前配置
	if err := ct.backupExistingConfig(); err != nil {
		return fmt.Errorf("备份配置失败: %v", err)
	}

	// 加载当前配置
	config, err := LoadConfigFromFile(ct.configPath)
	if err != nil {
		return fmt.Errorf("加载当前配置失败: %v", err)
	}

	// 执行迁移逻辑
	migratedConfig, err := ct.performMigration(config, fromVersion, toVersion)
	if err != nil {
		return fmt.Errorf("执行迁移失败: %v", err)
	}

	// 保存迁移后的配置
	if err := migratedConfig.SaveToFile(ct.configPath); err != nil {
		return fmt.Errorf("保存迁移后配置失败: %v", err)
	}

	logger.Info("配置迁移完成",
		zap.String("from_version", fromVersion),
		zap.String("to_version", toVersion))

	return nil
}

// performMigration 执行配置迁移
func (ct *ConfigTool) performMigration(config *PaymentConfig, fromVersion, toVersion string) (*PaymentConfig, error) {
	// 这里实现具体的迁移逻辑
	// 根据版本差异进行字段添加、删除、修改等操作

	switch {
	case fromVersion == "1.0" && toVersion == "1.1":
		// 示例：添加新的安全配置
		if config.Security.MaxRequestSize == 0 {
			config.Security.MaxRequestSize = 1024 * 1024 // 1MB
		}
		if config.Security.RateLimitRPS == 0 {
			config.Security.RateLimitRPS = 100
		}

	case fromVersion == "1.1" && toVersion == "1.2":
		// 示例：添加银联支付配置
		if !config.UnionPay.Enabled {
			config.UnionPay.Enabled = false
			config.UnionPay.GatewayURL = "https://gateway.test.95516.com"
			config.UnionPay.Timeout = 30 * time.Second
		}
	}

	return config, nil
}

// backupExistingConfig 备份现有配置
func (ct *ConfigTool) backupExistingConfig() error {
	// 检查配置文件是否存在
	if _, err := os.Stat(ct.configPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	// 创建备份目录
	if err := os.MkdirAll(ct.backupPath, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	baseName := strings.TrimSuffix(filepath.Base(ct.configPath), filepath.Ext(ct.configPath))
	backupFile := filepath.Join(ct.backupPath, fmt.Sprintf("%s_%s.json", baseName, timestamp))

	// 复制文件
	data, err := os.ReadFile(ct.configPath)
	if err != nil {
		return fmt.Errorf("读取原配置文件失败: %v", err)
	}

	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return fmt.Errorf("写入备份文件失败: %v", err)
	}

	logger.Info("配置文件备份成功", zap.String("backup_file", backupFile))
	return nil
}

// generateEnvExample 生成环境变量示例文件
func (ct *ConfigTool) generateEnvExample(env string) error {
	envFile := filepath.Join(filepath.Dir(ct.configPath), fmt.Sprintf(".env.%s.example", env))

	var envContent strings.Builder
	envContent.WriteString(fmt.Sprintf("# Mall-Go Payment Configuration - %s Environment\n", strings.ToUpper(env)))
	envContent.WriteString(fmt.Sprintf("# Generated at: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	envContent.WriteString("# 基础配置\n")
	envContent.WriteString(fmt.Sprintf("PAYMENT_ENVIRONMENT=%s\n", env))
	envContent.WriteString("PAYMENT_DEBUG=true\n")
	envContent.WriteString("PAYMENT_LOG_LEVEL=info\n\n")

	envContent.WriteString("# 支付宝配置\n")
	if env == "prod" {
		envContent.WriteString("ALIPAY_APP_ID=your_production_app_id\n")
		envContent.WriteString("ALIPAY_PRIVATE_KEY=your_production_private_key\n")
		envContent.WriteString("ALIPAY_PUBLIC_KEY=your_production_public_key\n")
	} else {
		envContent.WriteString("ALIPAY_APP_ID=2021000000000000\n")
		envContent.WriteString("ALIPAY_PRIVATE_KEY=your_sandbox_private_key\n")
		envContent.WriteString("ALIPAY_PUBLIC_KEY=your_sandbox_public_key\n")
	}
	envContent.WriteString("\n")

	envContent.WriteString("# 微信支付配置\n")
	if env == "prod" {
		envContent.WriteString("WECHAT_APP_ID=your_production_app_id\n")
		envContent.WriteString("WECHAT_MCH_ID=your_production_mch_id\n")
		envContent.WriteString("WECHAT_API_KEY=your_production_api_key\n")
	} else {
		envContent.WriteString("WECHAT_APP_ID=wx1234567890abcdef\n")
		envContent.WriteString("WECHAT_MCH_ID=1234567890\n")
		envContent.WriteString("WECHAT_API_KEY=your_test_api_key\n")
	}
	envContent.WriteString("\n")

	envContent.WriteString("# 银联支付配置\n")
	envContent.WriteString("UNIONPAY_MERCHANT_ID=your_merchant_id\n")
	envContent.WriteString("UNIONPAY_CERT_PATH=/path/to/cert.pfx\n")
	envContent.WriteString("UNIONPAY_KEY_PATH=/path/to/key.key\n")

	if err := os.WriteFile(envFile, []byte(envContent.String()), 0644); err != nil {
		return fmt.Errorf("写入环境变量示例文件失败: %v", err)
	}

	logger.Info("环境变量示例文件生成成功", zap.String("env_file", envFile))
	return nil
}

// CompareConfigs 比较两个配置文件
func (ct *ConfigTool) CompareConfigs(otherConfigPath string) (*ConfigComparison, error) {
	// 加载当前配置
	config1, err := LoadConfigFromFile(ct.configPath)
	if err != nil {
		return nil, fmt.Errorf("加载当前配置失败: %v", err)
	}

	// 加载比较配置
	config2, err := LoadConfigFromFile(otherConfigPath)
	if err != nil {
		return nil, fmt.Errorf("加载比较配置失败: %v", err)
	}

	comparison := &ConfigComparison{
		Config1Path: ct.configPath,
		Config2Path: otherConfigPath,
		ComparedAt:  time.Now(),
		Differences: make([]ConfigDifference, 0),
	}

	// 比较基础配置
	ct.compareField(comparison, "environment", config1.Environment, config2.Environment)
	ct.compareField(comparison, "debug", config1.Debug, config2.Debug)
	ct.compareField(comparison, "log_level", config1.LogLevel, config2.LogLevel)

	// 比较支付宝配置
	ct.compareField(comparison, "alipay.enabled", config1.Alipay.Enabled, config2.Alipay.Enabled)
	ct.compareField(comparison, "alipay.app_id", config1.Alipay.AppID, config2.Alipay.AppID)
	ct.compareField(comparison, "alipay.gateway_url", config1.Alipay.GatewayURL, config2.Alipay.GatewayURL)

	// 比较微信配置
	ct.compareField(comparison, "wechat.enabled", config1.Wechat.Enabled, config2.Wechat.Enabled)
	ct.compareField(comparison, "wechat.app_id", config1.Wechat.AppID, config2.Wechat.AppID)
	ct.compareField(comparison, "wechat.mch_id", config1.Wechat.MchID, config2.Wechat.MchID)

	logger.Info("配置比较完成",
		zap.String("config1", ct.configPath),
		zap.String("config2", otherConfigPath),
		zap.Int("differences", len(comparison.Differences)))

	return comparison, nil
}

// ConfigComparison 配置比较结果
type ConfigComparison struct {
	Config1Path string             `json:"config1_path"`
	Config2Path string             `json:"config2_path"`
	ComparedAt  time.Time          `json:"compared_at"`
	Differences []ConfigDifference `json:"differences"`
}

// ConfigDifference 配置差异
type ConfigDifference struct {
	Field  string      `json:"field"`
	Value1 interface{} `json:"value1"`
	Value2 interface{} `json:"value2"`
	Type   string      `json:"type"` // changed, added, removed
}

// compareField 比较单个字段
func (ct *ConfigTool) compareField(comparison *ConfigComparison, field string, value1, value2 interface{}) {
	if fmt.Sprintf("%v", value1) != fmt.Sprintf("%v", value2) {
		comparison.Differences = append(comparison.Differences, ConfigDifference{
			Field:  field,
			Value1: value1,
			Value2: value2,
			Type:   "changed",
		})
	}
}

// ListBackups 列出备份文件
func (ct *ConfigTool) ListBackups() ([]BackupInfo, error) {
	backups := make([]BackupInfo, 0)

	// 检查备份目录
	if _, err := os.Stat(ct.backupPath); os.IsNotExist(err) {
		return backups, nil
	}

	// 读取备份目录
	files, err := os.ReadDir(ct.backupPath)
	if err != nil {
		return nil, fmt.Errorf("读取备份目录失败: %v", err)
	}

	// 解析备份文件
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			backup := BackupInfo{
				FileName:  file.Name(),
				FilePath:  filepath.Join(ct.backupPath, file.Name()),
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			}

			backups = append(backups, backup)
		}
	}

	return backups, nil
}

// BackupInfo 备份信息
type BackupInfo struct {
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// RestoreFromBackup 从备份恢复配置
func (ct *ConfigTool) RestoreFromBackup(backupFileName string) error {
	backupFilePath := filepath.Join(ct.backupPath, backupFileName)

	// 检查备份文件是否存在
	if _, err := os.Stat(backupFilePath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", backupFileName)
	}

	// 备份当前配置
	if err := ct.backupExistingConfig(); err != nil {
		logger.Warn("备份当前配置失败", zap.Error(err))
	}

	// 复制备份文件到配置路径
	data, err := os.ReadFile(backupFilePath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %v", err)
	}

	if err := os.WriteFile(ct.configPath, data, 0644); err != nil {
		return fmt.Errorf("恢复配置文件失败: %v", err)
	}

	logger.Info("配置恢复成功",
		zap.String("backup_file", backupFileName),
		zap.String("config_path", ct.configPath))

	return nil
}
