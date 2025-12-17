package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// LogLevel 日志级别类型
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// LogConfig 日志配置
type LogConfig struct {
	Level      LogLevel `json:"level" yaml:"level"`
	Format     string   `json:"format" yaml:"format"`           // json or console
	Output     string   `json:"output" yaml:"output"`           // stdout, stderr, file
	Filename   string   `json:"filename" yaml:"filename"`       // 日志文件名
	MaxSize    int      `json:"max_size" yaml:"max_size"`       // 日志文件最大大小(MB)
	MaxAge     int      `json:"max_age" yaml:"max_age"`         // 保留天数
	MaxBackups int      `json:"max_backups" yaml:"max_backups"` // 保留文件数
	Compress   bool     `json:"compress" yaml:"compress"`       // 是否压缩
}

// StructuredLog 结构化日志字段
type StructuredLog struct {
	TraceID    string                 `json:"trace_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Action     string                 `json:"action,omitempty"`
	Resource   string                 `json:"resource,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RequestContext 请求上下文
type RequestContext struct {
	TraceID   string
	UserID    string
	RequestID string
	StartTime time.Time
}

// 上下文键
type contextKey string

const (
	requestContextKey contextKey = "request_context"
	traceIDKey        contextKey = "trace_id"
)

// Init 初始化日志系统
func Init() {
	config := getDefaultConfig()
	InitWithConfig(config)
}

// InitWithConfig 使用配置初始化日志系统
func InitWithConfig(config LogConfig) {
	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据格式选择编码器
	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	switch config.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	case PanicLevel:
		level = zapcore.PanicLevel
	}

	// 设置输出
	var writer zapcore.WriteSyncer
	switch config.Output {
	case "stderr":
		writer = zapcore.AddSync(os.Stderr)
	case "file":
		writer = getFileWriter(config)
	default:
		writer = zapcore.AddSync(os.Stdout)
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writer, level)

	// 创建日志器
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = Logger.Sugar()

	Info("Logger initialized successfully",
		zap.String("level", string(config.Level)),
		zap.String("format", config.Format),
		zap.String("output", config.Output))
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() LogConfig {
	return LogConfig{
		Level:      InfoLevel,
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 3,
		Compress:   true,
	}
}

// getFileWriter 获取文件写入器
func getFileWriter(config LogConfig) zapcore.WriteSyncer {
	// 确保日志目录存在
	if config.Filename != "" {
		dir := filepath.Dir(config.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("failed to create log directory: %v", err))
		}
	}

	// 这里应该使用 lumberjack 来实现日志轮转，但为了简化先使用文件
	file, err := os.OpenFile(config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file: %v", err))
	}
	return zapcore.AddSync(file)
}

// WithContext 创建带上下文的日志器
func WithContext(ctx context.Context) *zap.Logger {
	if Logger == nil {
		return nil
	}

	fields := []zap.Field{}

	// 添加追踪ID
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// 添加请求上下文
	if reqCtx, ok := ctx.Value(requestContextKey).(*RequestContext); ok {
		if reqCtx.TraceID != "" {
			fields = append(fields, zap.String("trace_id", reqCtx.TraceID))
		}
		if reqCtx.UserID != "" {
			fields = append(fields, zap.String("user_id", reqCtx.UserID))
		}
		if reqCtx.RequestID != "" {
			fields = append(fields, zap.String("request_id", reqCtx.RequestID))
		}
	}

	if len(fields) > 0 {
		return Logger.With(fields...)
	}

	return Logger
}

// NewRequestContext 创建新的请求上下文
func NewRequestContext(userID string) *RequestContext {
	return &RequestContext{
		TraceID:   uuid.New().String(),
		UserID:    userID,
		RequestID: uuid.New().String(),
		StartTime: time.Now(),
	}
}

// ContextWithTrace 在上下文中添加追踪ID
func ContextWithTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// ContextWithRequest 在上下文中添加请求上下文
func ContextWithRequest(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, reqCtx)
}

// GinMiddleware Gin日志中间件
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 创建请求上下文
		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			if uidStr, ok := uid.(string); ok {
				userID = uidStr
			} else if uidUint, ok := uid.(uint); ok {
				userID = fmt.Sprintf("%d", uidUint)
			}
		}

		reqCtx := NewRequestContext(userID)
		ctx := ContextWithRequest(c.Request.Context(), reqCtx)
		c.Request = c.Request.WithContext(ctx)

		// 设置响应头
		c.Header("X-Trace-ID", reqCtx.TraceID)
		c.Header("X-Request-ID", reqCtx.RequestID)

		// 记录请求开始
		WithContext(ctx).Info("Request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("remote_addr", c.ClientIP()),
		)

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 记录请求完成
		logger := WithContext(ctx)
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.Int("response_size", c.Writer.Size()),
		}

		// 添加错误信息（如果有）
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		if c.Writer.Status() >= 400 {
			logger.Error("Request completed with error", fields...)
		} else {
			logger.Info("Request completed", fields...)
		}
	}
}

// LogStructured 记录结构化日志
func LogStructured(level LogLevel, message string, log StructuredLog) {
	if Logger == nil {
		return
	}

	fields := []zap.Field{}

	if log.TraceID != "" {
		fields = append(fields, zap.String("trace_id", log.TraceID))
	}
	if log.UserID != "" {
		fields = append(fields, zap.String("user_id", log.UserID))
	}
	if log.RequestID != "" {
		fields = append(fields, zap.String("request_id", log.RequestID))
	}
	if log.Action != "" {
		fields = append(fields, zap.String("action", log.Action))
	}
	if log.Resource != "" {
		fields = append(fields, zap.String("resource", log.Resource))
	}
	if log.Method != "" {
		fields = append(fields, zap.String("method", log.Method))
	}
	if log.Path != "" {
		fields = append(fields, zap.String("path", log.Path))
	}
	if log.StatusCode != 0 {
		fields = append(fields, zap.Int("status_code", log.StatusCode))
	}
	if log.Duration != 0 {
		fields = append(fields, zap.Duration("duration", log.Duration))
	}
	if log.Error != "" {
		fields = append(fields, zap.String("error", log.Error))
	}
	if log.Metadata != nil {
		for k, v := range log.Metadata {
			fields = append(fields, zap.Any(k, v))
		}
	}

	switch level {
	case DebugLevel:
		Logger.Debug(message, fields...)
	case InfoLevel:
		Logger.Info(message, fields...)
	case WarnLevel:
		Logger.Warn(message, fields...)
	case ErrorLevel:
		Logger.Error(message, fields...)
	case FatalLevel:
		Logger.Fatal(message, fields...)
	case PanicLevel:
		Logger.Panic(message, fields...)
	}
}

// LogPerformance 记录性能日志
func LogPerformance(ctx context.Context, operation string, duration time.Duration, metadata map[string]interface{}) {
	logger := WithContext(ctx)

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("performance_type", "timing"),
	}

	if metadata != nil {
		for k, v := range metadata {
			fields = append(fields, zap.Any(k, v))
		}
	}

	// 根据执行时间判断日志级别
	if duration > time.Second {
		logger.Warn("Slow operation detected", fields...)
	} else if duration > 500*time.Millisecond {
		logger.Info("Performance log", fields...)
	} else {
		logger.Debug("Performance log", fields...)
	}
}

// LogBusinessEvent 记录业务事件日志
func LogBusinessEvent(ctx context.Context, event string, entity string, entityID string, action string, metadata map[string]interface{}) {
	logger := WithContext(ctx)

	fields := []zap.Field{
		zap.String("event_type", "business"),
		zap.String("event", event),
		zap.String("entity", entity),
		zap.String("entity_id", entityID),
		zap.String("action", action),
		zap.Time("event_time", time.Now()),
	}

	if metadata != nil {
		for k, v := range metadata {
			fields = append(fields, zap.Any(k, v))
		}
	}

	logger.Info("Business event", fields...)
}

// LogSecurityEvent 记录安全事件日志
func LogSecurityEvent(ctx context.Context, event string, severity string, details map[string]interface{}) {
	logger := WithContext(ctx)

	fields := []zap.Field{
		zap.String("event_type", "security"),
		zap.String("security_event", event),
		zap.String("severity", severity),
		zap.Time("event_time", time.Now()),
	}

	if details != nil {
		for k, v := range details {
			fields = append(fields, zap.Any(k, v))
		}
	}

	// 根据严重程度选择日志级别
	switch severity {
	case "critical", "high":
		logger.Error("Security event", fields...)
	case "medium":
		logger.Warn("Security event", fields...)
	default:
		logger.Info("Security event", fields...)
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}

// Panic panic日志
func Panic(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Panic(msg, fields...)
	}
}

// Sync 同步日志
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// 简化版本的日志方法（使用Sugar）

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Debugf(template, args...)
	}
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Infof(template, args...)
	}
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Warnf(template, args...)
	}
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Errorf(template, args...)
	}
}

// Fatalf 格式化致命错误日志
func Fatalf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Fatalf(template, args...)
	}
}

// Panicf 格式化panic日志
func Panicf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Panicf(template, args...)
	}
}

// With 添加字段
func With(fields ...zap.Field) *zap.Logger {
	if Logger != nil {
		return Logger.With(fields...)
	}
	return nil
}

// WithFields 添加多个字段（使用Sugar）
func WithFields(fields ...interface{}) *zap.SugaredLogger {
	if Sugar != nil {
		return Sugar.With(fields...)
	}
	return nil
}

// GetLogger 获取原始logger
func GetLogger() *zap.Logger {
	return Logger
}

// GetSugar 获取sugar logger
func GetSugar() *zap.SugaredLogger {
	return Sugar
}

// NewDevelopmentLogger 创建开发环境日志器
func NewDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return config.Build()
}

// NewProductionLogger 创建生产环境日志器
func NewProductionLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	return config.Build()
}

// SetLevel 设置日志级别
func SetLevel(level zapcore.Level) {
	if Logger != nil {
		// 重新创建core和logger
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)

		Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		Sugar = Logger.Sugar()
	}
}

// Close 关闭日志器
func Close() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}
