package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	defaultTraceIDField            = "traceId"
	traceIDContextKey   contextKey = "traceId"
)

// Config 定义 logger 初始化配置。
type Config struct {
	Level        string
	Encoding     string
	TraceIDField string
}

// Logger 封装 zap.Logger，并支持从 context 自动注入 traceId 字段。
type Logger struct {
	base         *zap.Logger
	traceIDField string
}

// New 根据配置创建 Logger。
func New(cfg Config) (*Logger, error) {
	level := parseLevel(cfg.Level)
	enc := cfg.Encoding
	if enc == "" {
		enc = "json"
	}
	traceField := cfg.TraceIDField
	if traceField == "" {
		traceField = defaultTraceIDField
	}

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         enc,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
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
		},
	}

	base, err := zapCfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{
		base:         base,
		traceIDField: traceField,
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

func parseLevel(level string) zapcore.Level {
	switch level {
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

func (l *Logger) attachTraceID(ctx context.Context, fields []zap.Field) []zap.Field {
	if l == nil {
		return fields
	}
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		fields = append(fields, zap.String(l.traceIDField, traceID))
	}
	return fields
}

// Debug 输出 debug 级别日志。
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.base.Debug(msg, l.attachTraceID(ctx, fields)...)
}

// Info 输出 info 级别日志。
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.base.Info(msg, l.attachTraceID(ctx, fields)...)
}

// Warn 输出 warn 级别日志。
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.base.Warn(msg, l.attachTraceID(ctx, fields)...)
}

// Error 输出 error 级别日志。
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.base.Error(msg, l.attachTraceID(ctx, fields)...)
}

// Sync 刷新缓冲日志。
func (l *Logger) Sync() error {
	if l == nil || l.base == nil {
		return nil
	}
	if err := l.base.Sync(); err != nil && err != os.ErrInvalid {
		return err
	}
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
