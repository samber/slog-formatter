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

func TestTimeFormatter_DefaultFormat(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(TimeFormatter("", nil))

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "created_at" {
							is.Equal(ts.Format(time.RFC3339), attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("created_at", ts))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestTimeFormatter_CustomFormat(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(TimeFormatter("2006/01/02", nil))

	ts := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "date" {
							is.Equal("2024/06/15", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("date", ts))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestTimeFormatter_WithLocation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	loc, err := time.LoadLocation("America/New_York")
	is.NoError(err)

	var checked int32
	handler := NewFormatterMiddleware(TimeFormatter(time.RFC3339, loc))

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ts" {
							expected := ts.In(loc).Format(time.RFC3339)
							is.Equal(expected, attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("ts", ts))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestTimeFormatter_NilLocation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(TimeFormatter(time.RFC3339, nil))

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ts" {
							is.Equal(ts.Format(time.RFC3339), attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("ts", ts))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestTimeFormatter_NonTimeUnchanged(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(TimeFormatter("", nil))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "str" {
							is.Equal("not-a-time", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("str", "not-a-time"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestUnixTimestampFormatter_AllPrecisions(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name      string
		precision time.Duration
		expected  int64
	}{
		{"second", time.Second, ts.Unix()},
		{"millisecond", time.Millisecond, ts.UnixMilli()},
		{"microsecond", time.Microsecond, ts.UnixMicro()},
		{"nanosecond", time.Nanosecond, ts.UnixNano()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			var checked int32
			handler := NewFormatterMiddleware(UnixTimestampFormatter(tt.precision))

			logger := slog.New(
				handler(
					slogmock.Option{
						Handle: func(ctx context.Context, record slog.Record) error {
							record.Attrs(func(attr slog.Attr) bool {
								if attr.Key == "ts" {
									is.Equal(tt.expected, attr.Value.Int64())
									atomic.AddInt32(&checked, 1)
								}
								return true
							})
							return nil
						},
					}.NewMockHandler(),
				),
			)

			logger.Info("test", slog.Time("ts", ts))
			is.Equal(int32(1), atomic.LoadInt32(&checked))
		})
	}
}

func TestUnixTimestampFormatter_InvalidPrecision(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.Panics(func() {
		UnixTimestampFormatter(time.Minute)
	})

	is.Panics(func() {
		UnixTimestampFormatter(time.Hour)
	})

	is.Panics(func() {
		UnixTimestampFormatter(0)
	})
}

func TestTimezoneConverter_Basic(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	loc, err := time.LoadLocation("Asia/Tokyo")
	is.NoError(err)

	var checked int32
	handler := NewFormatterMiddleware(TimezoneConverter(loc))

	ts := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ts" {
							converted := attr.Value.Time()
							is.Equal("Asia/Tokyo", converted.Location().String())
							is.Equal(19, converted.Hour()) // UTC+9
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("ts", ts))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestTimezoneConverter_NilLocationDefaultsUTC(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	loc, err := time.LoadLocation("America/New_York")
	is.NoError(err)

	var checked int32
	handler := NewFormatterMiddleware(TimezoneConverter(nil))

	ts := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ts" {
							converted := attr.Value.Time()
							is.Equal("UTC", converted.Location().String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("ts", ts))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestTimeFormatter_ZeroTime(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(TimeFormatter(time.RFC3339, nil))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ts" {
							// Zero time should still format
							is.Contains(attr.Value.String(), "0001")
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Time("ts", time.Time{}))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}
