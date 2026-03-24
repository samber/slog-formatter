package slogformatter

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// helpers

func newDiscardLogger(formatters ...Formatter) *slog.Logger {
	handler := NewFormatterHandler(formatters...)
	return slog.New(handler(slog.NewJSONHandler(io.Discard, nil)))
}

func buildNestedGroup(depth int, leafKey, leafVal string) slog.Attr {
	leaf := slog.String(leafKey, leafVal)
	attr := leaf
	for i := depth; i >= 1; i-- {
		attr = slog.Group(fmt.Sprintf("g%d", i), attr)
	}
	return attr
}

// Benchmarks

func BenchmarkFormatByKey(b *testing.B) {
	b.Run("flat_match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKey("target", func(v slog.Value) slog.Value {
				return slog.StringValue("formatted")
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("target", "value"))
		}
	})

	b.Run("flat_no_match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKey("target", func(v slog.Value) slog.Value {
				return slog.StringValue("formatted")
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("other", "value"))
		}
	})

	b.Run("nested_3_levels", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKey("leaf", func(v slog.Value) slog.Value {
				return slog.StringValue("formatted")
			}),
		)
		attr := buildNestedGroup(3, "leaf", "value")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", attr)
		}
	})

	b.Run("nested_10_levels", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKey("leaf", func(v slog.Value) slog.Value {
				return slog.StringValue("formatted")
			}),
		)
		attr := buildNestedGroup(10, "leaf", "value")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", attr)
		}
	})
}

func BenchmarkFormatByKind(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKind(slog.KindDuration, func(v slog.Value) slog.Value {
				return slog.StringValue(v.Duration().String())
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Duration("d", time.Second))
		}
	})

	b.Run("no_match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKind(slog.KindDuration, func(v slog.Value) slog.Value {
				return slog.StringValue(v.Duration().String())
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("s", "value"))
		}
	})

	b.Run("nested_match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKind(slog.KindInt64, func(v slog.Value) slog.Value {
				return slog.Int64Value(v.Int64() * 2)
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench",
				slog.Group("g1",
					slog.Group("g2",
						slog.Int64("num", 42),
					),
				),
			)
		}
	})
}

func BenchmarkFormatByType(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByType[time.Duration](func(v time.Duration) slog.Value {
				return slog.StringValue(v.String())
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Duration("d", time.Second))
		}
	})

	b.Run("no_match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByType[time.Duration](func(v time.Duration) slog.Value {
				return slog.StringValue(v.String())
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("s", "value"))
		}
	})
}

func BenchmarkFormat(b *testing.B) {
	b.Run("passthrough", func(b *testing.B) {
		logger := newDiscardLogger(
			Format(func(groups []string, key string, value slog.Value) slog.Value {
				return value
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("a", "1"), slog.String("b", "2"))
		}
	})

	b.Run("transform", func(b *testing.B) {
		logger := newDiscardLogger(
			Format(func(groups []string, key string, value slog.Value) slog.Value {
				if value.Kind() == slog.KindString {
					return slog.StringValue("X")
				}
				return value
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("a", "1"), slog.String("b", "2"))
		}
	})
}

func BenchmarkFormatByFieldType(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByFieldType[string]("target", func(v string) slog.Value {
				return slog.StringValue("formatted_" + v)
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("target", "value"))
		}
	})

	b.Run("no_match_key", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByFieldType[string]("target", func(v string) slog.Value {
				return slog.StringValue("formatted_" + v)
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("other", "value"))
		}
	})
}

func BenchmarkFormatByGroup(b *testing.B) {
	b.Run("match", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByGroup([]string{"parent", "child"}, func(attrs []slog.Attr) slog.Value {
				return slog.GroupValue(attrs...)
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench",
				slog.Group("parent",
					slog.Group("child",
						slog.String("a", "1"),
					),
				),
			)
		}
	})
}

func BenchmarkFormatByGroupKey(b *testing.B) {
	logger := newDiscardLogger(
		FormatByGroupKey([]string{"parent"}, "target", func(v slog.Value) slog.Value {
			return slog.StringValue("formatted")
		}),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench",
			slog.Group("parent",
				slog.String("target", "value"),
			),
		)
	}
}

func BenchmarkPIIFormatter(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		logger := newDiscardLogger(PIIFormatter("user"))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench",
				slog.Group("user",
					slog.String("id", "uuid-123"),
					slog.String("email", "user@example.com"),
					slog.String("name", "John Doe"),
				),
			)
		}
	})

	b.Run("nested_address", func(b *testing.B) {
		logger := newDiscardLogger(PIIFormatter("user"))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench",
				slog.Group("user",
					slog.String("id", "uuid-123"),
					slog.String("email", "user@example.com"),
					slog.Group("address",
						slog.String("street", "123 Main St"),
						slog.String("city", "New York"),
						slog.Int("zip", 10001),
					),
				),
			)
		}
	})
}

func BenchmarkIPAddressFormatter(b *testing.B) {
	logger := newDiscardLogger(IPAddressFormatter("ip"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench", slog.String("ip", "192.168.1.1"))
	}
}

func BenchmarkHTTPRequestFormatter(b *testing.B) {
	b.Run("hide_headers", func(b *testing.B) {
		logger := newDiscardLogger(HTTPRequestFormatter(true))
		req := &http.Request{
			Method: "GET",
			Host:   "example.com",
			URL: &url.URL{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/api/test",
				RawQuery: "foo=bar",
			},
			Header: http.Header{
				"Authorization": []string{"Bearer token"},
				"Content-Type":  []string{"application/json"},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Any("request", req))
		}
	})

	b.Run("show_headers", func(b *testing.B) {
		logger := newDiscardLogger(HTTPRequestFormatter(false))
		req := &http.Request{
			Method: "POST",
			Host:   "example.com",
			URL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/api/test",
			},
			Header: http.Header{
				"Authorization": []string{"Bearer token"},
				"Content-Type":  []string{"application/json"},
				"Accept":        []string{"*/*"},
			},
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Any("request", req))
		}
	})
}

func BenchmarkHTTPResponseFormatter(b *testing.B) {
	logger := newDiscardLogger(HTTPResponseFormatter(true))
	resp := &http.Response{
		StatusCode:    200,
		Status:        "200 OK",
		ContentLength: 1234,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench", slog.Any("response", resp))
	}
}

func BenchmarkTimeFormatter(b *testing.B) {
	b.Run("rfc3339", func(b *testing.B) {
		logger := newDiscardLogger(TimeFormatter(time.RFC3339, nil))
		ts := time.Now()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Time("ts", ts))
		}
	})

	b.Run("custom_format", func(b *testing.B) {
		logger := newDiscardLogger(TimeFormatter("2006-01-02 15:04:05", nil))
		ts := time.Now()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Time("ts", ts))
		}
	})

	b.Run("with_location", func(b *testing.B) {
		loc, _ := time.LoadLocation("America/New_York")
		logger := newDiscardLogger(TimeFormatter(time.RFC3339, loc))
		ts := time.Now()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.Time("ts", ts))
		}
	})
}

func BenchmarkUnixTimestampFormatter(b *testing.B) {
	precisions := []struct {
		name string
		p    time.Duration
	}{
		{"second", time.Second},
		{"millisecond", time.Millisecond},
		{"microsecond", time.Microsecond},
		{"nanosecond", time.Nanosecond},
	}
	ts := time.Now()

	for _, p := range precisions {
		b.Run(p.name, func(b *testing.B) {
			logger := newDiscardLogger(UnixTimestampFormatter(p.p))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("bench", slog.Time("ts", ts))
			}
		})
	}
}

func BenchmarkTimezoneConverter(b *testing.B) {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	logger := newDiscardLogger(TimezoneConverter(loc))
	ts := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench", slog.Time("ts", ts))
	}
}

func BenchmarkErrorFormatter(b *testing.B) {
	logger := newDiscardLogger(ErrorFormatter("error"))
	err := errors.New("something failed")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench", slog.Any("error", err))
	}
}

func BenchmarkFormatterHandler_Handle(b *testing.B) {
	b.Run("no_formatters", func(b *testing.B) {
		logger := newDiscardLogger()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("a", "1"), slog.Int("b", 2))
		}
	})

	b.Run("single_formatter", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKey("a", func(v slog.Value) slog.Value {
				return slog.StringValue("X")
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench", slog.String("a", "1"), slog.Int("b", 2))
		}
	})

	b.Run("three_formatters", func(b *testing.B) {
		logger := newDiscardLogger(
			FormatByKey("a", func(v slog.Value) slog.Value {
				return slog.StringValue("X")
			}),
			FormatByKind(slog.KindInt64, func(v slog.Value) slog.Value {
				return slog.Int64Value(v.Int64() * 2)
			}),
			FormatByType[time.Duration](func(v time.Duration) slog.Value {
				return slog.StringValue(v.String())
			}),
		)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("bench",
				slog.String("a", "1"),
				slog.Int("b", 2),
				slog.Duration("d", time.Second),
			)
		}
	})
}

func BenchmarkFlattenAttrs(b *testing.B) {
	b.Run("flat", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"),
			slog.String("b", "2"),
			slog.String("c", "3"),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FlattenAttrs(attrs)
		}
	})

	b.Run("nested_1", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.Group("g", slog.String("a", "1"), slog.String("b", "2")),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FlattenAttrs(attrs)
		}
	})

	b.Run("nested_3", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.Group("g1",
				slog.Group("g2",
					slog.Group("g3",
						slog.String("leaf", "val"),
					),
				),
			),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FlattenAttrs(attrs)
		}
	})
}

func BenchmarkFlattenAttrsWithPrefix(b *testing.B) {
	b.Run("flat", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.String("a", "1"),
			slog.String("b", "2"),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FlattenAttrsWithPrefix(".", "root", attrs)
		}
	})

	b.Run("nested_3", func(b *testing.B) {
		attrs := []slog.Attr{
			slog.Group("g1",
				slog.Group("g2",
					slog.Group("g3",
						slog.String("leaf", "val"),
					),
				),
			),
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			FlattenAttrsWithPrefix(".", "root", attrs)
		}
	})
}

func BenchmarkDeepNesting(b *testing.B) {
	depths := []int{1, 3, 5, 10}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("depth_%d", depth), func(b *testing.B) {
			logger := newDiscardLogger(
				FormatByKey("leaf", func(v slog.Value) slog.Value {
					return slog.StringValue("formatted")
				}),
			)
			attr := buildNestedGroup(depth, "leaf", "value")
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("bench", attr)
			}
		})
	}
}

func BenchmarkMultipleFormatters(b *testing.B) {
	counts := []int{1, 5, 10}

	for _, count := range counts {
		b.Run(fmt.Sprintf("formatters_%d", count), func(b *testing.B) {
			formatters := make([]Formatter, count)
			for i := 0; i < count; i++ {
				key := fmt.Sprintf("key%d", i)
				formatters[i] = FormatByKey(key, func(v slog.Value) slog.Value {
					return slog.StringValue("X")
				})
			}
			logger := newDiscardLogger(formatters...)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("bench",
					slog.String("key0", "val"),
					slog.String("key1", "val"),
					slog.String("other", "val"),
				)
			}
		})
	}
}

func BenchmarkLogValuerResolution(b *testing.B) {
	logger := newDiscardLogger(
		FormatByKey("resolved", func(v slog.Value) slog.Value {
			return slog.StringValue("formatted_" + v.String())
		}),
	)
	lv := testLogValuer{val: "test_value"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("bench", slog.Any("resolved", lv))
	}
}
