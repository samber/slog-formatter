package slogformatter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	slogmock "github.com/samber/slog-mock"
	"github.com/stretchr/testify/assert"
)

func TestFormatterHandler_Enabled(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	inner := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})
	h := NewFormatterHandler()(inner)

	is.False(h.Enabled(context.Background(), slog.LevelDebug))
	is.False(h.Enabled(context.Background(), slog.LevelInfo))
	is.True(h.Enabled(context.Background(), slog.LevelWarn))
	is.True(h.Enabled(context.Background(), slog.LevelError))
}

func TestFormatterHandler_Handle_NoFormatters(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterHandler()
	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					attrs := map[string]slog.Value{}
					record.Attrs(func(attr slog.Attr) bool {
						attrs[attr.Key] = attr.Value
						return true
					})
					is.Equal("value", attrs["key"].String())
					is.Equal(int64(42), attrs["num"].Int64())
					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("key", "value"), slog.Int64("num", 42))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatterHandler_Handle_MultipleFormatters(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// First formatter uppercases strings, second prefixes them
	var order []string
	f1 := Format(func(groups []string, key string, value slog.Value) slog.Value {
		order = append(order, "f1")
		if value.Kind() == slog.KindString {
			return slog.StringValue("UPPER_" + value.String())
		}
		return value
	})
	f2 := Format(func(groups []string, key string, value slog.Value) slog.Value {
		order = append(order, "f2")
		if value.Kind() == slog.KindString {
			return slog.StringValue("PREFIX_" + value.String())
		}
		return value
	})

	var checked int32
	handler := NewFormatterHandler(f1, f2)
	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "msg" {
							// f1 applied first, then f2
							is.Equal("PREFIX_UPPER_hello", attr.Value.String())
						}
						return true
					})
					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("msg", "hello"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
	// Both formatters called in order
	is.Equal([]string{"f1", "f2"}, order)
}

func TestFormatterHandler_WithAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterHandler(
		FormatByKey("preset", func(v slog.Value) slog.Value {
			return slog.StringValue("formatted_" + v.String())
		}),
	)

	h := handler(
		slogmock.Option{
			Handle: func(ctx context.Context, record slog.Record) error {
				record.Attrs(func(attr slog.Attr) bool {
					if attr.Key == "preset" {
						is.Equal("formatted_original", attr.Value.String())
						atomic.AddInt32(&checked, 1)
					}
					return true
				})
				return nil
			},
		}.NewMockHandler(),
	)

	// WithAttrs should pre-transform the attrs
	h = h.WithAttrs([]slog.Attr{slog.String("preset", "original")})
	// Trigger a log to see the pre-transformed attr
	h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatterHandler_WithGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := NewFormatterHandler()
	inner := slog.NewJSONHandler(io.Discard, nil)
	h := handler(inner)

	// WithGroup returns a new handler with the group
	h2 := h.WithGroup("mygroup")
	is.NotEqual(h, h2)

	// Verify it's a FormatterHandler with groups set
	fh, ok := h2.(*FormatterHandler)
	is.True(ok)
	is.Equal([]string{"mygroup"}, fh.groups)
}

func TestFormatterHandler_WithGroup_EmptyName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := NewFormatterHandler()
	inner := slog.NewJSONHandler(io.Discard, nil)
	h := handler(inner)

	// Empty group name returns same handler
	h2 := h.WithGroup("")
	is.Equal(h, h2)
}

func TestFormatterHandler_WithGroup_Chained(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := NewFormatterHandler()
	inner := slog.NewJSONHandler(io.Discard, nil)
	h := handler(inner)

	h = h.WithGroup("a").WithGroup("b").WithGroup("c")
	fh := h.(*FormatterHandler)
	is.Equal([]string{"a", "b", "c"}, fh.groups)
}

type testLogValuer struct {
	val string
}

func (v testLogValuer) LogValue() slog.Value {
	return slog.StringValue(v.val)
}

func TestFormatterHandler_LogValuerResolution(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterHandler(
		FormatByKey("resolved", func(v slog.Value) slog.Value {
			// By the time we get here, LogValuer should be resolved
			is.Equal(slog.KindString, v.Kind())
			is.Equal("resolved_value", v.String())
			return slog.StringValue("formatted_" + v.String())
		}),
	)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "resolved" {
							is.Equal("formatted_resolved_value", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("resolved", testLogValuer{val: "resolved_value"}))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatterHandler_Handle_EmptyRecord(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterHandler(
		Format(func(groups []string, key string, value slog.Value) slog.Value {
			t.Fatal("formatter should not be called on empty record")
			return value
		}),
	)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					is.Equal(0, record.NumAttrs())
					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("empty")
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatterHandler_ConcurrentUse(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var count int64
	handler := NewFormatterHandler(
		FormatByKind(slog.KindInt64, func(v slog.Value) slog.Value {
			return slog.Int64Value(v.Int64() * 2)
		}),
	)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					atomic.AddInt64(&count, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Info("concurrent", slog.Int64("val", int64(n)))
		}(i)
	}
	wg.Wait()

	is.Equal(int64(100), atomic.LoadInt64(&count))
}

func TestFormatterHandler_Handle_PreservesRecordFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterHandler()
	now := time.Now()

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					is.Equal("test message", record.Message)
					is.Equal(slog.LevelWarn, record.Level)
					// Time should be close (slog sets it)
					is.WithinDuration(now, record.Time, 1*time.Second)
					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Warn("test message")
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatterHandler_Handle_ErrorPropagation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := NewFormatterHandler()
	expectedErr := fmt.Errorf("handler error")

	h := handler(
		slogmock.Option{
			Handle: func(ctx context.Context, record slog.Record) error {
				return expectedErr
			},
		}.NewMockHandler(),
	)

	err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
	is.ErrorIs(err, expectedErr)
}

// nestedLogValuer wraps another LogValuer
type nestedLogValuer struct {
	inner slog.LogValuer
}

func (v nestedLogValuer) LogValue() slog.Value {
	return slog.AnyValue(v.inner)
}

func TestFormatterHandler_NestedLogValuer(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterHandler(
		FormatByKey("nested", func(v slog.Value) slog.Value {
			is.Equal("deep_value", v.String())
			return slog.StringValue("found_" + v.String())
		}),
	)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "nested" {
							is.Equal("found_deep_value", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	// Double-wrapped LogValuer
	logger.Info("test", slog.Any("nested", nestedLogValuer{inner: testLogValuer{val: "deep_value"}}))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}
