package slogformatter

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	slogmock "github.com/samber/slog-mock"
)

func FuzzFormatByKey(f *testing.F) {
	f.Add("key", "value")
	f.Add("", "")
	f.Add("a.b.c", "hello world")
	f.Add("key\x00with\nnewlines", "val\ttab")
	f.Add("très-spécial", "日本語")
	f.Add("key with spaces", "value with\x00null")
	f.Add(string(make([]byte, 1024)), "large key")

	f.Fuzz(func(t *testing.T, key, value string) {
		formatter := FormatByKey(key, func(v slog.Value) slog.Value {
			return slog.StringValue("formatted_" + v.String())
		})

		// Should not panic on any input
		attr := slog.String(key, value)
		val, ok := formatter(nil, attr)
		if ok {
			_ = val.String()
		}

		// Also test with a non-matching key
		attr2 := slog.String("other_"+key, value)
		val2, ok2 := formatter(nil, attr2)
		if ok2 {
			_ = val2.String()
		}

		// Test nested in group
		groupAttr := slog.Group("g", slog.String(key, value))
		val3, ok3 := formatter(nil, groupAttr)
		if ok3 {
			_ = val3.String()
		}
	})
}

func FuzzFormatByKind(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(-1))
	f.Add(int64(9223372036854775807))
	f.Add(int64(-9223372036854775808))
	f.Add(int64(42))

	f.Fuzz(func(t *testing.T, val int64) {
		// Fuzz Int64 kind
		formatter := FormatByKind(slog.KindInt64, func(v slog.Value) slog.Value {
			return slog.Int64Value(v.Int64() + 1)
		})

		attr := slog.Int64("key", val)
		result, ok := formatter(nil, attr)
		if ok {
			_ = result.Int64()
		}

		// Test with non-matching kind
		strAttr := slog.String("key", "str")
		result2, ok2 := formatter(nil, strAttr)
		if ok2 {
			_ = result2.String()
		}
	})
}

func FuzzPIIFormatter(f *testing.F) {
	f.Add("email", "user@example.com")
	f.Add("id", "uuid-123")
	f.Add("user_id", "abc")
	f.Add("name", "Jo")
	f.Add("", "")
	f.Add("ID", "preserved")
	f.Add("field-id", "also-preserved")
	f.Add("secret", "a")
	f.Add("data", "12345")   // exactly 5 chars
	f.Add("data", "123456")  // exactly 6 chars
	f.Add("key", string(make([]byte, 0)))
	f.Add("key\x00null", "val\x00null")

	f.Fuzz(func(t *testing.T, key, value string) {
		// Should never panic
		handler := NewFormatterMiddleware(PIIFormatter("data"))
		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						record.Attrs(func(attr slog.Attr) bool {
							_ = attr.Value.String()
							return true
						})
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("fuzz", slog.Group("data", slog.String(key, value)))
	})
}

func FuzzFlattenAttrs(f *testing.F) {
	f.Add("key1", "val1", "key2", "val2", "group")
	f.Add("", "", "", "", "")
	f.Add("a.b", "x", "c.d", "y", "g.h")
	f.Add("key\x00", "val\x00", "k2", "v2", "grp")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2, groupName string) {
		// Test FlattenAttrs with various attr combinations
		attrs := []slog.Attr{
			slog.String(k1, v1),
			slog.Group(groupName,
				slog.String(k2, v2),
			),
		}

		result := FlattenAttrs(attrs)
		for _, a := range result {
			_ = a.Key
			_ = a.Value.String()
		}

		// Test FlattenAttrsWithPrefix
		result2 := FlattenAttrsWithPrefix(".", "prefix", attrs)
		for _, a := range result2 {
			_ = a.Key
			_ = a.Value.String()
		}

		// Test PrefixAttrKeys
		result3 := PrefixAttrKeys("pfx.", attrs)
		for _, a := range result3 {
			_ = a.Key
			_ = a.Value.String()
		}
	})
}

func FuzzTimeFormatter(f *testing.F) {
	f.Add("2006-01-02", int64(0))
	f.Add(time.RFC3339, int64(1705312200))
	f.Add("", int64(-1))
	f.Add("Mon Jan _2 15:04:05 2006", int64(9999999999))
	f.Add("invalid-format-%%", int64(42))
	f.Add(string(make([]byte, 100)), int64(0))

	f.Fuzz(func(t *testing.T, format string, unixSec int64) {
		// Should never panic
		handler := NewFormatterMiddleware(TimeFormatter(format, nil))
		ts := time.Unix(unixSec, 0)

		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						record.Attrs(func(attr slog.Attr) bool {
							_ = attr.Value.String()
							return true
						})
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("fuzz", slog.Time("ts", ts))
	})
}

func FuzzErrorFormatter(f *testing.F) {
	f.Add("error", "something went wrong")
	f.Add("err", "")
	f.Add("", "msg")
	f.Add("error", string(make([]byte, 4096)))
	f.Add("key\x00null", "err\nnewline")

	f.Fuzz(func(t *testing.T, key, msg string) {
		handler := NewFormatterMiddleware(ErrorFormatter(key))

		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						record.Attrs(func(attr slog.Attr) bool {
							_ = attr.Value.String()
							return true
						})
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("fuzz", slog.Any(key, errors.New(msg)))
		// Also test with non-error value
		logger.Info("fuzz", slog.String(key, msg))
	})
}

func FuzzFormat(f *testing.F) {
	f.Add("key", "value", "group")
	f.Add("", "", "")
	f.Add("k\x00ey", "v\nal", "g\tr")

	f.Fuzz(func(t *testing.T, key, value, groupName string) {
		formatter := Format(func(groups []string, k string, v slog.Value) slog.Value {
			return v
		})

		// Flat attr
		attr := slog.String(key, value)
		val, _ := formatter(nil, attr)
		_ = val.String()

		// Nested attr
		groupAttr := slog.Group(groupName, slog.String(key, value))
		val2, _ := formatter(nil, groupAttr)
		_ = val2.String()

		// Double nested
		deepAttr := slog.Group("outer", slog.Group(groupName, slog.String(key, value)))
		val3, _ := formatter(nil, deepAttr)
		_ = val3.String()
	})
}

func FuzzIPAddressFormatter(f *testing.F) {
	f.Add("ip", "192.168.1.1")
	f.Add("ip", "::1")
	f.Add("ip", "not-an-ip")
	f.Add("ip", "")
	f.Add("other", "10.0.0.1")

	f.Fuzz(func(t *testing.T, key, value string) {
		handler := NewFormatterMiddleware(IPAddressFormatter(key))

		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						record.Attrs(func(attr slog.Attr) bool {
							_ = attr.Value.String()
							return true
						})
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("fuzz", slog.String(key, value))
		logger.Info("fuzz", slog.String("other_"+key, value))
	})
}
