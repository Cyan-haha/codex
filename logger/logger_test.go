package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObserverLogger(traceField string) (*Logger, *observer.ObservedLogs) {
	core, observed := observer.New(zapcore.DebugLevel)
	base := zap.New(core)
	if traceField == "" {
		traceField = defaultTraceIDField
	}
	return &Logger{base: base, traceIDField: traceField}, observed
}

func TestInfoWithTraceID(t *testing.T) {
	l, observed := newObserverLogger("")
	ctx := WithTraceID(context.Background(), "abc-123")

	l.Info(ctx, "hello", zap.String("k", "v"))

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	ctxMap := entries[0].ContextMap()
	if got := ctxMap[defaultTraceIDField]; got != "abc-123" {
		t.Fatalf("expected traceId abc-123, got %#v", got)
	}
	if got := ctxMap["k"]; got != "v" {
		t.Fatalf("expected field k=v, got %#v", got)
	}
}

func TestInfoWithoutTraceID(t *testing.T) {
	l, observed := newObserverLogger("")

	l.Info(context.Background(), "hello")

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	ctxMap := entries[0].ContextMap()
	if _, ok := ctxMap[defaultTraceIDField]; ok {
		t.Fatalf("did not expect traceId field")
	}
}

func TestCustomTraceIDField(t *testing.T) {
	l, observed := newObserverLogger("trace_id")
	ctx := WithTraceID(context.Background(), "abc")

	l.Info(ctx, "hello")

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	ctxMap := entries[0].ContextMap()
	if got := ctxMap["trace_id"]; got != "abc" {
		t.Fatalf("expected trace_id abc, got %#v", got)
	}
}
