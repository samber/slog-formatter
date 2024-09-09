package slogformatter

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	slogmock "github.com/samber/slog-mock"
	"github.com/stretchr/testify/assert"
)

func TestFormatByKind(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := NewFormatterMiddleware(
		FormatByKind(slog.KindDuration, func(v slog.Value) slog.Value {
			return slog.StringValue(v.Duration().String())
		}),
	)

	var checked int32

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					is.Equal("hello world", record.Message)

					attrs := map[string]slog.Value{}
					record.Attrs(func(attr slog.Attr) bool {
						attrs[attr.Key] = attr.Value
						return true
					})

					is.Len(attrs, 2)
					is.Equal(attrs["key"], slog.StringValue("value"))
					is.Equal(attrs["duration"], slog.StringValue("1s"))

					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("hello world",
		slog.String("key", "value"),
		slog.Duration("duration", 1*time.Second))

	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatByKey(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := NewFormatterMiddleware(
		FormatByKey("duration", func(v slog.Value) slog.Value {
			return slog.StringValue(v.Duration().String())
		}),
	)

	var checked int32

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					is.Equal("hello world", record.Message)

					attrs := map[string]slog.Value{}
					record.Attrs(func(attr slog.Attr) bool {
						attrs[attr.Key] = attr.Value
						return true
					})

					is.Len(attrs, 2)
					is.Equal(attrs["key"], slog.StringValue("value"))
					is.Equal(attrs["duration"], slog.StringValue("1s"))

					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("hello world",
		slog.String("key", "value"),
		slog.Duration("duration", 1*time.Second))

	is.Equal(int32(1), atomic.LoadInt32(&checked))
}
