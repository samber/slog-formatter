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
					is.Equal(attrs["subgroup"].Kind(), slog.KindGroup)
					is.Equal(attrs["subgroup"].Group()[0].Key, "duration")
					is.Equal(attrs["subgroup"].Group()[0].Value, slog.StringValue("1s"))

					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("hello world",
		slog.String("key", "value"),
		slog.Group("subgroup",
			slog.Duration("duration", 1*time.Second),
		),
	)

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
					is.Equal(attrs["subgroup"].Kind(), slog.KindGroup)
					is.Equal(attrs["subgroup"].Group()[0].Key, "duration")
					is.Equal(attrs["subgroup"].Group()[0].Value, slog.StringValue("1s"))

					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("hello world",
		slog.String("key", "value"),
		slog.Group("subgroup",
			slog.Duration("duration", 1*time.Second),
		),
	)

	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestRecursiveNestedGroups(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Test FormatByKey with deeply nested groups
	t.Run("FormatByKey recursive", func(t *testing.T) {
		handler := NewFormatterMiddleware(
			FormatByKey("deep_value", func(v slog.Value) slog.Value {
				return slog.StringValue("formatted_" + v.String())
			}),
		)

		var checked int32
		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						found := false
						record.Attrs(func(attr slog.Attr) bool {
							// Traverse the nested structure to find the formatted value
							if attr.Key == "level1" && attr.Value.Kind() == slog.KindGroup {
								level1Group := attr.Value.Group()
								for _, level2Attr := range level1Group {
									if level2Attr.Key == "level2" && level2Attr.Value.Kind() == slog.KindGroup {
										level2Group := level2Attr.Value.Group()
										for _, level3Attr := range level2Group {
											if level3Attr.Key == "level3" && level3Attr.Value.Kind() == slog.KindGroup {
												level3Group := level3Attr.Value.Group()
												for _, deepAttr := range level3Group {
													if deepAttr.Key == "deep_value" {
														is.Equal("formatted_value", deepAttr.Value.String())
														found = true
													}
												}
											}
										}
									}
								}
							}
							return true
						})
						is.True(found, "formatted deep_value should be found")
						atomic.AddInt32(&checked, 1)
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("test",
			slog.Group("level1",
				slog.Group("level2",
					slog.Group("level3",
						slog.String("deep_value", "value"),
					),
				),
			),
		)

		is.Equal(int32(1), atomic.LoadInt32(&checked))
	})

	// Test FormatByType with deeply nested groups
	t.Run("FormatByType recursive", func(t *testing.T) {
		handler := NewFormatterMiddleware(
			FormatByType[time.Duration](func(v time.Duration) slog.Value {
				return slog.StringValue(v.String())
			}),
		)

		var checked int32
		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						found := false
						record.Attrs(func(attr slog.Attr) bool {
							// Traverse the nested structure to find the formatted duration
							if attr.Key == "level1" && attr.Value.Kind() == slog.KindGroup {
								level1Group := attr.Value.Group()
								for _, level2Attr := range level1Group {
									if level2Attr.Key == "level2" && level2Attr.Value.Kind() == slog.KindGroup {
										level2Group := level2Attr.Value.Group()
										for _, level3Attr := range level2Group {
											if level3Attr.Key == "level3" && level3Attr.Value.Kind() == slog.KindGroup {
												level3Group := level3Attr.Value.Group()
												for _, durationAttr := range level3Group {
													if durationAttr.Key == "deep_duration" {
														is.Equal("1s", durationAttr.Value.String())
														found = true
													}
												}
											}
										}
									}
								}
							}
							return true
						})
						is.True(found, "formatted deep_duration should be found")
						atomic.AddInt32(&checked, 1)
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("test",
			slog.Group("level1",
				slog.Group("level2",
					slog.Group("level3",
						slog.Duration("deep_duration", 1*time.Second),
					),
				),
			),
		)

		is.Equal(int32(1), atomic.LoadInt32(&checked))
	})

	// Test FormatByKind with deeply nested groups
	t.Run("FormatByKind recursive", func(t *testing.T) {
		handler := NewFormatterMiddleware(
			FormatByKind(slog.KindInt64, func(v slog.Value) slog.Value {
				return slog.Int64Value(v.Int64() * 2)
			}),
		)

		var checked int32
		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						found := false
						record.Attrs(func(attr slog.Attr) bool {
							// Traverse the nested structure to find the formatted int
							if attr.Key == "level1" && attr.Value.Kind() == slog.KindGroup {
								level1Group := attr.Value.Group()
								for _, level2Attr := range level1Group {
									if level2Attr.Key == "level2" && level2Attr.Value.Kind() == slog.KindGroup {
										level2Group := level2Attr.Value.Group()
										for _, level3Attr := range level2Group {
											if level3Attr.Key == "level3" && level3Attr.Value.Kind() == slog.KindGroup {
												level3Group := level3Attr.Value.Group()
												for _, intAttr := range level3Group {
													if intAttr.Key == "deep_int" {
														is.Equal(int64(200), intAttr.Value.Int64())
														found = true
													}
												}
											}
										}
									}
								}
							}
							return true
						})
						is.True(found, "formatted deep_int should be found")
						atomic.AddInt32(&checked, 1)
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("test",
			slog.Group("level1",
				slog.Group("level2",
					slog.Group("level3",
						slog.Int64("deep_int", 100),
					),
				),
			),
		)

		is.Equal(int32(1), atomic.LoadInt32(&checked))
	})
}

func TestRecursiveFormatterEdgeCases(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Test empty groups
	t.Run("empty groups in Format", func(t *testing.T) {
		called := false
		handler := NewFormatterMiddleware(
			Format(func(groups []string, key string, value slog.Value) slog.Value {
				is.Empty(groups, "groups should be empty for top-level attributes")
				called = true
				return value
			}),
		)

		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						return nil
					},
				}.NewMockHandler(),
			),
		)

		logger.Info("test", slog.String("key", "value"))
		is.True(called, "formatter should be called")
	})

	// Test multiple matching keys in different levels
	t.Run("multiple matches in different levels", func(t *testing.T) {
		handler := NewFormatterMiddleware(
			FormatByKey("target", func(v slog.Value) slog.Value {
				return slog.StringValue("formatted_" + v.String())
			}),
		)

		var formattedCount int
		logger := slog.New(
			handler(
				slogmock.Option{
					Handle: func(ctx context.Context, record slog.Record) error {
						record.Attrs(func(attr slog.Attr) bool {
							// Count formatted values
							if attr.Key == "target" {
								is.Equal("formatted_level1", attr.Value.String())
								formattedCount++
							} else if attr.Key == "level1" && attr.Value.Kind() == slog.KindGroup {
								group := attr.Value.Group()
								for _, subAttr := range group {
									if subAttr.Key == "target" {
										is.Equal("formatted_level2", subAttr.Value.String())
										formattedCount++
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
			slog.String("target", "level1"),
			slog.Group("level1",
				slog.String("other", "value"),
				slog.String("target", "level2"),
			),
		)

		is.Equal(2, formattedCount, "should format all 2 matching attributes")
	})
}