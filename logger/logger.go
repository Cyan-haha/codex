package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	fixedTag                       = "T"
	defaultServiceName             = "DC"
	defaultTraceIDField            = "traceId"
	defaultTimeFormat              = "2006-01-02 15:04:05.000"
	missingTraceID                 = "-"
	traceIDContextKey   contextKey = "traceId"
)

// Config 定义 logger 初始化配置。
type Config struct {
	Level       string
	ServiceName string
}

// Logger 使用 zap 字段体系，按约定格式输出日志：
// T|服务名|级别|traceID|2006-01-02 15:04:05.000 日志内容
// 其他字段会追加在末尾（key=value）。
type Logger struct {
	mu          sync.Mutex
	level       zapcore.Level
	serviceName string
	writer      io.Writer
}

// New 根据配置创建 Logger。
func New(cfg Config) (*Logger, error) {
	return &Logger{
		level:       parseLevel(cfg.Level),
		serviceName: serviceNameOrDefault(cfg.ServiceName),
		writer:      os.Stdout,
	}, nil
}

// MustNew 创建 Logger，失败时 panic。
func MustNew(cfg Config) *Logger {
	l, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return l
}

func serviceNameOrDefault(v string) string {
	if strings.TrimSpace(v) == "" {
		return defaultServiceName
	}
	return v
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

// WithTraceID 将 traceId 注入 context。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceIDContextKey, traceID)
}

// TraceIDFromContext 从 context 中提取 traceId。
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx.Value(traceIDContextKey).(string)
	if !ok {
		return ""
	}
	return v
}

func (l *Logger) enabled(level zapcore.Level) bool {
	return level >= l.level
}

func (l *Logger) log(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	if l == nil || !l.enabled(level) {
		return
	}

	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		traceID = missingTraceID
	}
	line := fmt.Sprintf(
		"%s|%s|%s|%s|%s %s",
		fixedTag,
		l.serviceName,
		strings.ToUpper(level.String()),
		traceID,
		time.Now().Format(defaultTimeFormat),
		msg,
	)

	if tail := formatFields(fields...); tail != "" {
		line += " " + tail
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.writer.WriteString(line + "\n")
}

func formatFields(fields ...zap.Field) string {
	if len(fields) == 0 {
		return ""
	}

	enc := zapcore.NewMapObjectEncoder()
	for i := range fields {
		fields[i].AddTo(enc)
	}
	if len(enc.Fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(enc.Fields))
	for k := range enc.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(fmt.Sprintf("%s=%v", k, enc.Fields[k]))
	}
	return buf.String()
}

// Debug 输出 debug 级别日志。
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zap.DebugLevel, msg, fields...)
}

// Info 输出 info 级别日志。
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zap.InfoLevel, msg, fields...)
}

// Warn 输出 warn 级别日志。
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zap.WarnLevel, msg, fields...)
}

// Error 输出 error 级别日志。
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zap.ErrorLevel, msg, fields...)
}

// Sync 刷新缓冲日志。
func (l *Logger) Sync() error {
	return nil
}

var std = MustNew(Config{})

// ReplaceGlobal 替换包级别默认 Logger。
func ReplaceGlobal(l *Logger) {
	if l != nil {
		std = l
	}
}

// Global 返回包级默认 Logger。
func Global() *Logger {
	return std
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	std.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	std.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	std.Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	std.Error(ctx, msg, fields...)
}
