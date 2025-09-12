package performance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PerformanceReport 性能测试报告
type PerformanceReport struct {
	ReportID          string                    `json:"report_id"`
	GeneratedAt       time.Time                 `json:"generated_at"`
	ProjectName       string                    `json:"project_name"`
	TestPhase         string                    `json:"test_phase"`
	Summary           *ReportSummary            `json:"summary"`
	TestResults       []*TestResult             `json:"test_results"`
	PerformanceGoals  *PerformanceGoals         `json:"performance_goals"`
	Achievements      *Achievements             `json:"achievements"`
	Recommendations   []*Recommendation         `json:"recommendations"`
	TechnicalDetails  *TechnicalDetails         `json:"technical_details"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	TotalTests        int     `json:"total_tests"`
	PassedTests       int     `json:"passed_tests"`
	FailedTests       int     `json:"failed_tests"`
	OverallScore      float64 `json:"overall_score"`
	ExecutionTime     string  `json:"execution_time"`
	TestEnvironment   string  `json:"test_environment"`
}

// TestResult 测试结果
type TestResult struct {
	TestName          string        `json:"test_name"`
	Status            string        `json:"status"`
	ExecutionTime     time.Duration `json:"execution_time"`
	QPS               float64       `json:"qps"`
	AvgResponseTime   time.Duration `json:"avg_response_time"`
	P95ResponseTime   time.Duration `json:"p95_response_time"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
	ErrorRate         float64       `json:"error_rate"`
	ConcurrentUsers   int           `json:"concurrent_users"`
	TotalRequests     int64         `json:"total_requests"`
	Passed            bool          `json:"passed"`
	FailureReasons    []string      `json:"failure_reasons,omitempty"`
}

// PerformanceGoals 性能目标
type PerformanceGoals struct {
	TargetQPS              float64       `json:"target_qps"`
	MaxAvgResponseTime     time.Duration `json:"max_avg_response_time"`
	MaxP95ResponseTime     time.Duration `json:"max_p95_response_time"`
	MinCacheHitRate        float64       `json:"min_cache_hit_rate"`
	MaxErrorRate           float64       `json:"max_error_rate"`
	MinDBQueryReduction    float64       `json:"min_db_query_reduction"`
}

// Achievements 成就指标
type Achievements struct {
	ActualQPS           float64       `json:"actual_qps"`
	ActualAvgResponse   time.Duration `json:"actual_avg_response"`
	ActualP95Response   time.Duration `json:"actual_p95_response"`
	ActualCacheHitRate  float64       `json:"actual_cache_hit_rate"`
	ActualErrorRate     float64       `json:"actual_error_rate"`
	ActualDBReduction   float64       `json:"actual_db_reduction"`
	QPSImprovement      float64       `json:"qps_improvement"`
	ResponseImprovement float64       `json:"response_improvement"`
	CacheEffectiveness  float64       `json:"cache_effectiveness"`
}

// Recommendation 优化建议
type Recommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
}

// TechnicalDetails 技术细节
type TechnicalDetails struct {
	CacheArchitecture   string            `json:"cache_architecture"`
	RedisConfiguration  map[string]string `json:"redis_configuration"`
	TestConfiguration   map[string]string `json:"test_configuration"`
	SystemSpecifications map[string]string `json:"system_specifications"`
	CacheStrategies     []string          `json:"cache_strategies"`
	MonitoringMetrics   []string          `json:"monitoring_metrics"`
}

// PerformanceReportGenerator 性能报告生成器
type PerformanceReportGenerator struct {
	reportDir string
}

// NewPerformanceReportGenerator 创建性能报告生成器
func NewPerformanceReportGenerator(reportDir string) *PerformanceReportGenerator {
	return &PerformanceReportGenerator{
		reportDir: reportDir,
	}
}

// GenerateComprehensiveReport 生成综合性能报告
func (g *PerformanceReportGenerator) GenerateComprehensiveReport(
	testResults []*TestResult,
	goals *PerformanceGoals,
	achievements *Achievements,
) (*PerformanceReport, error) {
	
	// 创建报告目录
	if err := os.MkdirAll(g.reportDir, 0755); err != nil {
		return nil, fmt.Errorf("创建报告目录失败: %w", err)
	}
	
	// 生成报告ID
	reportID := fmt.Sprintf("perf_report_%s", time.Now().Format("20060102_150405"))
	
	// 计算摘要信息
	summary := g.calculateSummary(testResults)
	
	// 生成优化建议
	recommendations := g.generateRecommendations(testResults, achievements)
	
	// 生成技术细节
	technicalDetails := g.generateTechnicalDetails()
	
	// 创建报告
	report := &PerformanceReport{
		ReportID:         reportID,
		GeneratedAt:      time.Now(),
		ProjectName:      "Mall-Go电商系统",
		TestPhase:        "第三周缓存优化验证",
		Summary:          summary,
		TestResults:      testResults,
		PerformanceGoals: goals,
		Achievements:     achievements,
		Recommendations:  recommendations,
		TechnicalDetails: technicalDetails,
	}
	
	// 保存JSON报告
	if err := g.saveJSONReport(report); err != nil {
		return nil, fmt.Errorf("保存JSON报告失败: %w", err)
	}
	
	// 生成Markdown报告
	if err := g.generateMarkdownReport(report); err != nil {
		return nil, fmt.Errorf("生成Markdown报告失败: %w", err)
	}
	
	// 生成HTML报告
	if err := g.generateHTMLReport(report); err != nil {
		return nil, fmt.Errorf("生成HTML报告失败: %w", err)
	}
	
	return report, nil
}

// calculateSummary 计算摘要信息
func (g *PerformanceReportGenerator) calculateSummary(testResults []*TestResult) *ReportSummary {
	totalTests := len(testResults)
	passedTests := 0
	totalExecutionTime := time.Duration(0)
	
	for _, result := range testResults {
		if result.Passed {
			passedTests++
		}
		totalExecutionTime += result.ExecutionTime
	}
	
	overallScore := float64(passedTests) / float64(totalTests) * 100
	
	return &ReportSummary{
		TotalTests:      totalTests,
		PassedTests:     passedTests,
		FailedTests:     totalTests - passedTests,
		OverallScore:    overallScore,
		ExecutionTime:   totalExecutionTime.String(),
		TestEnvironment: "测试环境 (SQLite + Redis)",
	}
}

// generateRecommendations 生成优化建议
func (g *PerformanceReportGenerator) generateRecommendations(testResults []*TestResult, achievements *Achievements) []*Recommendation {
	var recommendations []*Recommendation
	
	// 基于测试结果生成建议
	for _, result := range testResults {
		if !result.Passed {
			for _, reason := range result.FailureReasons {
				if strings.Contains(reason, "QPS") {
					recommendations = append(recommendations, &Recommendation{
						Category:    "性能优化",
						Priority:    "高",
						Title:       "提升系统QPS",
						Description: "当前QPS未达到目标，建议优化缓存策略和数据库查询",
						Impact:      "高",
						Effort:      "中",
					})
				}
				if strings.Contains(reason, "响应时间") {
					recommendations = append(recommendations, &Recommendation{
						Category:    "性能优化",
						Priority:    "高",
						Title:       "优化响应时间",
						Description: "响应时间超标，建议增加缓存预热和优化热点数据访问",
						Impact:      "高",
						Effort:      "中",
					})
				}
				if strings.Contains(reason, "缓存命中率") {
					recommendations = append(recommendations, &Recommendation{
						Category:    "缓存优化",
						Priority:    "高",
						Title:       "提升缓存命中率",
						Description: "缓存命中率偏低，建议优化缓存策略和预热机制",
						Impact:      "高",
						Effort:      "中",
					})
				}
			}
		}
	}
	
	// 基于成就指标生成建议
	if achievements.ActualCacheHitRate < 90.0 {
		recommendations = append(recommendations, &Recommendation{
			Category:    "缓存优化",
			Priority:    "中",
			Title:       "进一步优化缓存命中率",
			Description: "虽然已达标，但仍有提升空间，建议分析热点数据模式",
			Impact:      "中",
			Effort:      "低",
		})
	}
	
	if achievements.ActualErrorRate > 0.5 {
		recommendations = append(recommendations, &Recommendation{
			Category:    "稳定性",
			Priority:    "中",
			Title:       "降低错误率",
			Description: "系统错误率可进一步降低，建议增强错误处理和重试机制",
			Impact:      "中",
			Effort:      "低",
		})
	}
	
	// 通用优化建议
	recommendations = append(recommendations, &Recommendation{
		Category:    "监控",
		Priority:    "低",
		Title:       "完善监控体系",
		Description: "建议建立生产环境监控告警体系，实时跟踪性能指标",
		Impact:      "中",
		Effort:      "中",
	})
	
	return recommendations
}

// generateTechnicalDetails 生成技术细节
func (g *PerformanceReportGenerator) generateTechnicalDetails() *TechnicalDetails {
	return &TechnicalDetails{
		CacheArchitecture: "Redis + 多级缓存架构",
		RedisConfiguration: map[string]string{
			"连接池大小":   "100",
			"最小空闲连接": "10",
			"最大重试次数": "3",
			"连接超时":    "5s",
			"读取超时":    "3s",
			"写入超时":    "3s",
		},
		TestConfiguration: map[string]string{
			"测试数据库":  "SQLite (内存)",
			"测试数据量":  "商品2000, 用户200, 分类5",
			"最大并发数":  "500",
			"测试时长":   "10-30秒",
			"测试环境":   "本地开发环境",
		},
		SystemSpecifications: map[string]string{
			"操作系统": "Windows/Linux",
			"Go版本":  "1.21+",
			"Redis版本": "6.0+",
			"数据库":   "SQLite/MySQL",
		},
		CacheStrategies: []string{
			"Cache-Aside模式",
			"Write-Through模式",
			"Write-Behind模式",
			"Refresh-Ahead模式",
		},
		MonitoringMetrics: []string{
			"缓存命中率",
			"响应时间分布",
			"QPS统计",
			"错误率监控",
			"热点数据分析",
			"内存使用监控",
		},
	}
}

// saveJSONReport 保存JSON报告
func (g *PerformanceReportGenerator) saveJSONReport(report *PerformanceReport) error {
	filename := fmt.Sprintf("%s.json", report.ReportID)
	filepath := filepath.Join(g.reportDir, filename)
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath, data, 0644)
}

// generateMarkdownReport 生成Markdown报告
func (g *PerformanceReportGenerator) generateMarkdownReport(report *PerformanceReport) error {
	filename := fmt.Sprintf("%s.md", report.ReportID)
	filepath := filepath.Join(g.reportDir, filename)
	
	var md strings.Builder
	
	// 报告标题
	md.WriteString(fmt.Sprintf("# %s - %s\n\n", report.ProjectName, report.TestPhase))
	md.WriteString(fmt.Sprintf("**报告ID**: %s  \n", report.ReportID))
	md.WriteString(fmt.Sprintf("**生成时间**: %s  \n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))
	
	// 执行摘要
	md.WriteString("## 📊 执行摘要\n\n")
	md.WriteString(fmt.Sprintf("- **总测试数**: %d\n", report.Summary.TotalTests))
	md.WriteString(fmt.Sprintf("- **通过测试**: %d\n", report.Summary.PassedTests))
	md.WriteString(fmt.Sprintf("- **失败测试**: %d\n", report.Summary.FailedTests))
	md.WriteString(fmt.Sprintf("- **总体得分**: %.2f%%\n", report.Summary.OverallScore))
	md.WriteString(fmt.Sprintf("- **执行时间**: %s\n", report.Summary.ExecutionTime))
	md.WriteString(fmt.Sprintf("- **测试环境**: %s\n\n", report.Summary.TestEnvironment))
	
	// 性能目标 vs 实际成就
	md.WriteString("## 🎯 性能目标 vs 实际成就\n\n")
	md.WriteString("| 指标 | 目标值 | 实际值 | 状态 |\n")
	md.WriteString("|------|--------|--------|------|\n")
	
	goals := report.PerformanceGoals
	achievements := report.Achievements
	
	md.WriteString(fmt.Sprintf("| QPS | %.0f | %.2f | %s |\n", 
		goals.TargetQPS, achievements.ActualQPS, 
		getMarkdownStatus(achievements.ActualQPS >= goals.TargetQPS)))
	
	md.WriteString(fmt.Sprintf("| 平均响应时间 | ≤%v | %v | %s |\n", 
		goals.MaxAvgResponseTime, achievements.ActualAvgResponse, 
		getMarkdownStatus(achievements.ActualAvgResponse <= goals.MaxAvgResponseTime)))
	
	md.WriteString(fmt.Sprintf("| P95响应时间 | ≤%v | %v | %s |\n", 
		goals.MaxP95ResponseTime, achievements.ActualP95Response, 
		getMarkdownStatus(achievements.ActualP95Response <= goals.MaxP95ResponseTime)))
	
	md.WriteString(fmt.Sprintf("| 缓存命中率 | ≥%.1f%% | %.2f%% | %s |\n", 
		goals.MinCacheHitRate, achievements.ActualCacheHitRate, 
		getMarkdownStatus(achievements.ActualCacheHitRate >= goals.MinCacheHitRate)))
	
	md.WriteString(fmt.Sprintf("| 错误率 | ≤%.1f%% | %.2f%% | %s |\n\n", 
		goals.MaxErrorRate, achievements.ActualErrorRate, 
		getMarkdownStatus(achievements.ActualErrorRate <= goals.MaxErrorRate)))
	
	// 详细测试结果
	md.WriteString("## 📋 详细测试结果\n\n")
	for _, result := range report.TestResults {
		status := "✅ 通过"
		if !result.Passed {
			status = "❌ 失败"
		}
		
		md.WriteString(fmt.Sprintf("### %s %s\n\n", result.TestName, status))
		md.WriteString(fmt.Sprintf("- **QPS**: %.2f\n", result.QPS))
		md.WriteString(fmt.Sprintf("- **平均响应时间**: %v\n", result.AvgResponseTime))
		md.WriteString(fmt.Sprintf("- **P95响应时间**: %v\n", result.P95ResponseTime))
		md.WriteString(fmt.Sprintf("- **缓存命中率**: %.2f%%\n", result.CacheHitRate))
		md.WriteString(fmt.Sprintf("- **错误率**: %.2f%%\n", result.ErrorRate))
		md.WriteString(fmt.Sprintf("- **并发用户**: %d\n", result.ConcurrentUsers))
		md.WriteString(fmt.Sprintf("- **总请求数**: %d\n", result.TotalRequests))
		
		if !result.Passed && len(result.FailureReasons) > 0 {
			md.WriteString("\n**失败原因**:\n")
			for _, reason := range result.FailureReasons {
				md.WriteString(fmt.Sprintf("- %s\n", reason))
			}
		}
		md.WriteString("\n")
	}
	
	// 优化建议
	if len(report.Recommendations) > 0 {
		md.WriteString("## 💡 优化建议\n\n")
		for i, rec := range report.Recommendations {
			md.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, rec.Title))
			md.WriteString(fmt.Sprintf("- **类别**: %s\n", rec.Category))
			md.WriteString(fmt.Sprintf("- **优先级**: %s\n", rec.Priority))
			md.WriteString(fmt.Sprintf("- **描述**: %s\n", rec.Description))
			md.WriteString(fmt.Sprintf("- **影响**: %s\n", rec.Impact))
			md.WriteString(fmt.Sprintf("- **工作量**: %s\n\n", rec.Effort))
		}
	}
	
	return os.WriteFile(filepath, []byte(md.String()), 0644)
}

// generateHTMLReport 生成HTML报告
func (g *PerformanceReportGenerator) generateHTMLReport(report *PerformanceReport) error {
	filename := fmt.Sprintf("%s.html", report.ReportID)
	filepath := filepath.Join(g.reportDir, filename)
	
	// 简化的HTML模板
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s - %s</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { background: #e8f5e8; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .test-result { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .passed { border-left: 5px solid #4CAF50; }
        .failed { border-left: 5px solid #f44336; }
        table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .status-pass { color: #4CAF50; font-weight: bold; }
        .status-fail { color: #f44336; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>%s - %s</h1>
        <p><strong>报告ID:</strong> %s</p>
        <p><strong>生成时间:</strong> %s</p>
    </div>
    
    <div class="summary">
        <h2>📊 执行摘要</h2>
        <p><strong>总体得分:</strong> %.2f%%</p>
        <p><strong>通过测试:</strong> %d/%d</p>
        <p><strong>执行时间:</strong> %s</p>
    </div>
    
    <h2>🎯 性能目标达成情况</h2>
    <p>详细的性能指标对比和测试结果请参考JSON或Markdown报告。</p>
    
    <p><em>完整的HTML报告功能正在开发中...</em></p>
</body>
</html>`, 
		report.ProjectName, report.TestPhase,
		report.ProjectName, report.TestPhase,
		report.ReportID,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		report.Summary.OverallScore,
		report.Summary.PassedTests, report.Summary.TotalTests,
		report.Summary.ExecutionTime)
	
	return os.WriteFile(filepath, []byte(html), 0644)
}

// getMarkdownStatus 获取Markdown状态标记
func getMarkdownStatus(passed bool) string {
	if passed {
		return "✅ 达标"
	}
	return "❌ 未达标"
}
