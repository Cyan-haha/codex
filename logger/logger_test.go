package logger

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"go.uber.org/zap"
)

func TestInfoFormatWithTraceID(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{level: zap.InfoLevel, serviceName: "DC", writer: buf}
	ctx := WithTraceID(context.Background(), "trace-1")

	l.Info(ctx, "hello", zap.String("k", "v"))

	line := buf.String()
	pattern := `^T\|DC\|INFO\|trace-1\|\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} hello k=v\n$`
	if !regexp.MustCompile(pattern).MatchString(line) {
		t.Fatalf("line not match pattern, got: %q", line)
	}
}

func TestInfoFormatWithoutTraceID(t *testing.T) {
	buf := &bytes.Buffer{}
	l := &Logger{level: zap.InfoLevel, serviceName: "DC", writer: buf}

	l.Info(context.Background(), "hello")

	line := buf.String()
	pattern := `^T\|DC\|INFO\|-\|\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} hello\n$`
	if !regexp.MustCompile(pattern).MatchString(line) {
		t.Fatalf("line not match pattern, got: %q", line)
	}
}

func TestFormatFieldsSort(t *testing.T) {
	got := formatFields(
		zap.String("b", "2"),
		zap.String("a", "1"),
	)
	if got != "a=1 b=2" {
		t.Fatalf("expected sorted fields, got %s", got)
	}
}
