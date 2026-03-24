package slogformatter

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"

	slogmock "github.com/samber/slog-mock"
	"github.com/stretchr/testify/assert"
)

func TestIPAddressFormatter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(IPAddressFormatter("ip"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ip" {
							is.Equal("*******", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("ip", "192.168.1.1"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestIPAddressFormatter_NestedGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(IPAddressFormatter("ip"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "ctx" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "ip" {
									is.Equal("*******", a.Value.String())
									atomic.AddInt32(&checked, 1)
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

	logger.Info("test", slog.Group("ctx", slog.String("ip", "10.0.0.1")))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestIPAddressFormatter_OtherKeyUnchanged(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(IPAddressFormatter("ip"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "host" {
							is.Equal("192.168.1.1", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("host", "192.168.1.1"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestPIIFormatter_BasicFields(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(PIIFormatter("user"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "user" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								switch a.Key {
								case "email":
									is.Equal("foob*******", a.Value.String())
								case "name":
									is.Equal("*******", a.Value.String())
								}
							}
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test",
		slog.Group("user",
			slog.String("email", "foobar@example.com"),
			slog.String("name", "John"),
		),
	)
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestPIIFormatter_IDFieldsPreserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
		val  string
	}{
		{"id field", "id", "uuid-123"},
		{"suffix _id", "user_id", "uuid-456"},
		{"suffix -id", "user-id", "uuid-789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			var checked int32
			handler := NewFormatterMiddleware(PIIFormatter("data"))

			logger := slog.New(
				handler(
					slogmock.Option{
						Handle: func(ctx context.Context, record slog.Record) error {
							record.Attrs(func(attr slog.Attr) bool {
								if attr.Key == "data" && attr.Value.Kind() == slog.KindGroup {
									for _, a := range attr.Value.Group() {
										if a.Key == tt.key {
											is.Equal(tt.val, a.Value.String())
											atomic.AddInt32(&checked, 1)
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

			logger.Info("test", slog.Group("data", slog.String(tt.key, tt.val)))
			is.Equal(int32(1), atomic.LoadInt32(&checked))
		})
	}
}

func TestPIIFormatter_IDCaseInsensitive(t *testing.T) {
	t.Parallel()

	// The code lowercases the key, so "ID", "Id", "iD" should all be preserved
	keys := []string{"ID", "Id", "iD", "id"}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			is := assert.New(t)
			var checked int32
			handler := NewFormatterMiddleware(PIIFormatter("data"))

			logger := slog.New(
				handler(
					slogmock.Option{
						Handle: func(ctx context.Context, record slog.Record) error {
							record.Attrs(func(attr slog.Attr) bool {
								if attr.Key == "data" && attr.Value.Kind() == slog.KindGroup {
									for _, a := range attr.Value.Group() {
										if a.Key == key {
											is.Equal("preserved", a.Value.String())
											atomic.AddInt32(&checked, 1)
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

			logger.Info("test", slog.Group("data", slog.String(key, "preserved")))
			is.Equal(int32(1), atomic.LoadInt32(&checked))
		})
	}
}

func TestPIIFormatter_ShortStrings(t *testing.T) {
	t.Parallel()

	// Strings <= 5 chars are fully masked
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"1 char", "a"},
		{"5 chars", "abcde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			var checked int32
			handler := NewFormatterMiddleware(PIIFormatter("data"))

			logger := slog.New(
				handler(
					slogmock.Option{
						Handle: func(ctx context.Context, record slog.Record) error {
							record.Attrs(func(attr slog.Attr) bool {
								if attr.Key == "data" && attr.Value.Kind() == slog.KindGroup {
									for _, a := range attr.Value.Group() {
										if a.Key == "val" {
											is.Equal("*******", a.Value.String())
											atomic.AddInt32(&checked, 1)
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

			logger.Info("test", slog.Group("data", slog.String("val", tt.input)))
			is.Equal(int32(1), atomic.LoadInt32(&checked))
		})
	}
}

func TestPIIFormatter_LongStrings(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Strings > 5 chars keep first 4 + "*******"
	var checked int32
	handler := NewFormatterMiddleware(PIIFormatter("data"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "data" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "email" {
									is.Equal("test*******", a.Value.String())
									atomic.AddInt32(&checked, 1)
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

	logger.Info("test", slog.Group("data", slog.String("email", "test@example.com")))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestPIIFormatter_NonStringValues(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(PIIFormatter("data"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "data" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								switch a.Key {
								case "age":
									is.Equal("*******", a.Value.String())
									atomic.AddInt32(&checked, 1)
								case "active":
									is.Equal("*******", a.Value.String())
									atomic.AddInt32(&checked, 1)
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

	logger.Info("test",
		slog.Group("data",
			slog.Int("age", 30),
			slog.Bool("active", true),
		),
	)
	is.Equal(int32(2), atomic.LoadInt32(&checked))
}

func TestPIIFormatter_NestedGroups(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(PIIFormatter("user"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "user" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "address" && a.Value.Kind() == slog.KindGroup {
									for _, inner := range a.Value.Group() {
										switch inner.Key {
										case "street":
											is.Equal("123 *******", inner.Value.String())
											atomic.AddInt32(&checked, 1)
										case "zip":
											is.Equal("*******", inner.Value.String())
											atomic.AddInt32(&checked, 1)
										}
									}
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

	logger.Info("test",
		slog.Group("user",
			slog.Group("address",
				slog.String("street", "123 Main St"),
				slog.Int("zip", 12345),
			),
		),
	)
	is.Equal(int32(2), atomic.LoadInt32(&checked))
}

func TestPIIFormatter_ExactBoundary6Chars(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(PIIFormatter("data"))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "data" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "val" {
									// 6 chars: first 4 kept + "*******"
									is.Equal("abcd*******", a.Value.String())
									atomic.AddInt32(&checked, 1)
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

	logger.Info("test", slog.Group("data", slog.String("val", "abcdef")))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}
