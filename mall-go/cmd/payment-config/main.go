package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"mall-go/pkg/payment"
)

// ConfigCLI 配置命令行工具
type ConfigCLI struct {
	configPath string
	tool       *payment.ConfigTool
}

func main() {
	// 定义命令行参数
	var (
		configPath = flag.String("config", "./config/payment.json", "配置文件路径")
		command    = flag.String("cmd", "", "执行命令: generate|validate|migrate|compare|backup|restore")
		env        = flag.String("env", "dev", "环境类型: dev|test|prod")
		force      = flag.Bool("force", false, "强制覆盖现有文件")
		fromVer    = flag.String("from", "", "迁移源版本")
		toVer      = flag.String("to", "", "迁移目标版本")
		compare    = flag.String("compare-with", "", "比较的配置文件路径")
		backup     = flag.String("backup", "", "备份文件名")
		output     = flag.String("output", "", "输出格式: json|table")
	)
	flag.Parse()

	// 创建CLI实例
	cli := &ConfigCLI{
		configPath: *configPath,
		tool:       payment.NewConfigTool(*configPath),
	}

	// 执行命令
	switch *command {
	case "generate":
		cli.generateConfig(*env, *force)
	case "validate":
		cli.validateConfig(*output)
	case "migrate":
		cli.migrateConfig(*fromVer, *toVer)
	case "compare":
		cli.compareConfigs(*compare, *output)
	case "backup":
		cli.listBackups(*output)
	case "restore":
		cli.restoreConfig(*backup)
	case "help":
		cli.showHelp()
	default:
		cli.showUsage()
	}
}

// generateConfig 生成配置文件
func (cli *ConfigCLI) generateConfig(env string, force bool) {
	fmt.Printf("🚀 正在为 %s 环境生成配置文件...\n", env)

	if err := cli.tool.GenerateConfigForEnvironment(env, force); err != nil {
		fmt.Printf("❌ 生成配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 配置文件生成成功: %s\n", cli.configPath)
	fmt.Printf("📝 请根据实际情况修改配置文件中的参数\n")
	fmt.Printf("🔧 环境变量示例文件已生成: .env.%s.example\n", env)
}

// validateConfig 验证配置文件
func (cli *ConfigCLI) validateConfig(output string) {
	fmt.Printf("🔍 正在验证配置文件: %s\n", cli.configPath)

	report, err := cli.tool.ValidateConfig()
	if err != nil {
		fmt.Printf("❌ 验证失败: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		cli.outputJSON(report)
		return
	}

	// 表格输出
	if report.IsValid {
		fmt.Printf("✅ 配置验证通过\n")
		fmt.Printf("   环境: %s\n", report.Environment)
		fmt.Printf("   验证时间: %s\n", report.ValidatedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("❌ 配置验证失败 (发现 %d 个错误)\n", report.ErrorCount)
		fmt.Printf("   环境: %s\n", report.Environment)
		fmt.Printf("   验证时间: %s\n", report.ValidatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("\n错误详情:\n")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "字段\t错误码\t错误信息")
		fmt.Fprintln(w, "----\t------\t--------")

		for _, err := range report.Errors {
			fmt.Fprintf(w, "%s\t%s\t%s\n", err.Field, err.Code, err.Message)
		}
		w.Flush()

		os.Exit(1)
	}
}

// migrateConfig 迁移配置
func (cli *ConfigCLI) migrateConfig(fromVer, toVer string) {
	if fromVer == "" || toVer == "" {
		fmt.Printf("❌ 迁移需要指定源版本和目标版本\n")
		fmt.Printf("   使用方法: -cmd=migrate -from=1.0 -to=1.1\n")
		os.Exit(1)
	}

	fmt.Printf("🔄 正在迁移配置从 %s 到 %s...\n", fromVer, toVer)

	if err := cli.tool.MigrateConfig(fromVer, toVer); err != nil {
		fmt.Printf("❌ 迁移失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 配置迁移成功\n")
	fmt.Printf("📝 建议重新验证配置文件\n")
}

// compareConfigs 比较配置
func (cli *ConfigCLI) compareConfigs(comparePath, output string) {
	if comparePath == "" {
		fmt.Printf("❌ 比较需要指定另一个配置文件路径\n")
		fmt.Printf("   使用方法: -cmd=compare -compare-with=/path/to/other/config.json\n")
		os.Exit(1)
	}

	fmt.Printf("🔍 正在比较配置文件...\n")
	fmt.Printf("   文件1: %s\n", cli.configPath)
	fmt.Printf("   文件2: %s\n", comparePath)

	comparison, err := cli.tool.CompareConfigs(comparePath)
	if err != nil {
		fmt.Printf("❌ 比较失败: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		cli.outputJSON(comparison)
		return
	}

	// 表格输出
	if len(comparison.Differences) == 0 {
		fmt.Printf("✅ 配置文件相同，无差异\n")
	} else {
		fmt.Printf("📊 发现 %d 个差异:\n\n", len(comparison.Differences))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "字段\t文件1值\t文件2值\t类型")
		fmt.Fprintln(w, "----\t------\t------\t----")

		for _, diff := range comparison.Differences {
			fmt.Fprintf(w, "%s\t%v\t%v\t%s\n", diff.Field, diff.Value1, diff.Value2, diff.Type)
		}
		w.Flush()
	}
}

// listBackups 列出备份文件
func (cli *ConfigCLI) listBackups(output string) {
	fmt.Printf("📂 正在查询备份文件...\n")

	backups, err := cli.tool.ListBackups()
	if err != nil {
		fmt.Printf("❌ 查询备份失败: %v\n", err)
		os.Exit(1)
	}

	if len(backups) == 0 {
		fmt.Printf("📭 没有找到备份文件\n")
		return
	}

	if output == "json" {
		cli.outputJSON(backups)
		return
	}

	// 表格输出
	fmt.Printf("📋 找到 %d 个备份文件:\n\n", len(backups))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "文件名\t大小\t创建时间\t文件路径")
	fmt.Fprintln(w, "------\t----\t--------\t--------")

	for _, backup := range backups {
		size := cli.formatSize(backup.Size)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			backup.FileName,
			size,
			backup.CreatedAt.Format("2006-01-02 15:04:05"),
			backup.FilePath)
	}
	w.Flush()

	fmt.Printf("\n💡 使用 -cmd=restore -backup=<文件名> 来恢复配置\n")
}

// restoreConfig 恢复配置
func (cli *ConfigCLI) restoreConfig(backupFile string) {
	if backupFile == "" {
		fmt.Printf("❌ 恢复需要指定备份文件名\n")
		fmt.Printf("   使用方法: -cmd=restore -backup=<文件名>\n")
		fmt.Printf("   使用 -cmd=backup 查看可用的备份文件\n")
		os.Exit(1)
	}

	fmt.Printf("🔄 正在从备份恢复配置: %s\n", backupFile)

	if err := cli.tool.RestoreFromBackup(backupFile); err != nil {
		fmt.Printf("❌ 恢复失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 配置恢复成功\n")
	fmt.Printf("📝 建议重新验证配置文件\n")
}

// outputJSON 输出JSON格式
func (cli *ConfigCLI) outputJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

// formatSize 格式化文件大小
func (cli *ConfigCLI) formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// showHelp 显示帮助信息
func (cli *ConfigCLI) showHelp() {
	fmt.Println("Mall-Go Payment Config Tool")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("这是一个用于管理Mall-Go支付系统配置的命令行工具。")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  generate  生成指定环境的配置文件")
	fmt.Println("  validate  验证配置文件的正确性")
	fmt.Println("  migrate   迁移配置文件版本")
	fmt.Println("  compare   比较两个配置文件")
	fmt.Println("  backup    列出备份文件")
	fmt.Println("  restore   从备份恢复配置")
	fmt.Println("  help      显示帮助信息")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("  -config   配置文件路径 (默认: ./config/payment.json)")
	fmt.Println("  -cmd      执行的命令")
	fmt.Println("  -env      环境类型: dev|test|prod (默认: dev)")
	fmt.Println("  -force    强制覆盖现有文件")
	fmt.Println("  -from     迁移源版本")
	fmt.Println("  -to       迁移目标版本")
	fmt.Println("  -compare-with  比较的配置文件路径")
	fmt.Println("  -backup   备份文件名")
	fmt.Println("  -output   输出格式: json|table (默认: table)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 生成开发环境配置")
	fmt.Println("  ./payment-config -cmd=generate -env=dev")
	fmt.Println()
	fmt.Println("  # 验证配置文件")
	fmt.Println("  ./payment-config -cmd=validate")
	fmt.Println()
	fmt.Println("  # 比较两个配置文件")
	fmt.Println("  ./payment-config -cmd=compare -compare-with=./config/prod.json")
	fmt.Println()
	fmt.Println("  # 列出备份文件")
	fmt.Println("  ./payment-config -cmd=backup")
	fmt.Println()
	fmt.Println("  # 恢复配置")
	fmt.Println("  ./payment-config -cmd=restore -backup=payment_20240101_120000.json")
}

// showUsage 显示使用方法
func (cli *ConfigCLI) showUsage() {
	fmt.Println("Mall-Go Payment Config Tool")
	fmt.Println("使用方法: ./payment-config -cmd=<command> [options]")
	fmt.Println()
	fmt.Println("可用命令:")
	fmt.Println("  generate, validate, migrate, compare, backup, restore, help")
	fmt.Println()
	fmt.Println("使用 -cmd=help 查看详细帮助信息")
	fmt.Println()

	// 自动生成默认配置提示
	configDir := filepath.Dir(cli.configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Printf("💡 配置目录不存在，是否要创建默认开发环境配置？\n")
		fmt.Printf("   运行: ./payment-config -cmd=generate -env=dev\n")
		fmt.Println()
	}
}

// 编译提示
func init() {
	// 检查是否在正确的目录中运行
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("警告: 建议在项目根目录中运行此工具")
	}
}
