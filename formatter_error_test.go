package slogformatter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"testing"

	slogmock "github.com/samber/slog-mock"
	"github.com/stretchr/testify/assert"
)

func TestErrorFormatter_Basic(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" && attr.Value.Kind() == slog.KindGroup {
							found := map[string]string{}
							for _, a := range attr.Value.Group() {
								found[a.Key] = a.Value.String()
							}
							is.Equal("something failed", found["message"])
							is.Contains(found["type"], "*errors.errorString")
							is.NotEmpty(found["stacktrace"])
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", errors.New("something failed")))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestErrorFormatter_WrongKey(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "err" {
							// Should not be transformed since key doesn't match
							is.NotEqual(slog.KindGroup, attr.Value.Kind())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("err", errors.New("wrong key")))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestErrorFormatter_NonErrorType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" {
							// Key matches but value is string not error, should not be group
							is.NotEqual(slog.KindGroup, attr.Value.Kind())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("error", "just a string"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestErrorFormatter_WrappedError(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	inner := errors.New("root cause")
	wrapped := fmt.Errorf("context: %w", inner)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" && attr.Value.Kind() == slog.KindGroup {
							found := map[string]string{}
							for _, a := range attr.Value.Group() {
								found[a.Key] = a.Value.String()
							}
							is.Equal("context: root cause", found["message"])
							is.Contains(found["type"], "fmt.wrapError")
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", wrapped))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

type customError struct {
	code int
	msg  string
}

func (e *customError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, e.msg)
}

func TestErrorFormatter_CustomErrorType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" && attr.Value.Kind() == slog.KindGroup {
							found := map[string]string{}
							for _, a := range attr.Value.Group() {
								found[a.Key] = a.Value.String()
							}
							is.Equal("[404] not found", found["message"])
							is.Contains(found["type"], "customError")
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", &customError{code: 404, msg: "not found"}))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestStacktrace_NonEmpty(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var stacktraceValue string
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "stacktrace" {
									stacktraceValue = a.Value.String()
								}
							}
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", errors.New("test error")))
	is.NotEmpty(stacktraceValue, "stacktrace should not be empty")
}

func TestStacktrace_SkipsSlogFrames(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var stacktraceValue string
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "stacktrace" {
									stacktraceValue = a.Value.String()
								}
							}
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", errors.New("test error")))
	is.NotContains(stacktraceValue, "log/slog", "stacktrace should not contain slog frames")
}

func TestStacktrace_ContainsCallerInfo(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var stacktraceValue string
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "stacktrace" {
									stacktraceValue = a.Value.String()
								}
							}
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", errors.New("test error")))
	is.Contains(stacktraceValue, "formatter_error_test.go", "stacktrace should contain the test file name")
}

func TestErrorFormatter_NilError(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(ErrorFormatter("error"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "error" {
							// nil error should not be formatted as group
							is.NotEqual(slog.KindGroup, attr.Value.Kind())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("error", nil))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}
