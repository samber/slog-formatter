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

func TestFormat_GroupsSlicePropagation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var capturedGroups [][]string
	handler := NewFormatterMiddleware(
		Format(func(groups []string, key string, value slog.Value) slog.Value {
			cp := make([]string, len(groups))
			copy(cp, groups)
			capturedGroups = append(capturedGroups, cp)
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

	logger.Info("test",
		slog.String("top", "val"),
		slog.Group("g1",
			slog.String("mid", "val"),
			slog.Group("g2",
				slog.String("deep", "val"),
			),
		),
	)

	is.Len(capturedGroups, 3)
	is.Empty(capturedGroups[0])          // top-level
	is.Equal([]string{"g1"}, capturedGroups[1]) // inside g1
	is.Equal([]string{"g1", "g2"}, capturedGroups[2]) // inside g1.g2
}

func TestFormat_EmptyGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(
		Format(func(groups []string, key string, value slog.Value) slog.Value {
			return value
		}),
	)

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "empty" {
							is.Equal(slog.KindGroup, attr.Value.Kind())
							is.Empty(attr.Value.Group())
						}
						return true
					})
					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Group("empty"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestFormatByType_NoMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByType[time.Duration](func(v time.Duration) slog.Value {
		return slog.StringValue(v.String())
	})

	// String attr should not match Duration type
	val, ok := formatter(nil, slog.String("key", "not-a-duration"))
	is.False(ok)
	is.Equal("not-a-duration", val.String())
}

func TestFormatByType_MatchFlat(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByType[time.Duration](func(v time.Duration) slog.Value {
		return slog.StringValue(v.String())
	})

	val, ok := formatter(nil, slog.Duration("d", 5*time.Second))
	is.True(ok)
	is.Equal("5s", val.String())
}

func TestFormatByFieldType_GroupSkipped(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByFieldType[string]("key", func(v string) slog.Value {
		return slog.StringValue("formatted")
	})

	groupAttr := slog.Group("key", slog.String("inner", "val"))
	val, ok := formatter(nil, groupAttr)
	is.False(ok)
	is.Equal(slog.KindGroup, val.Kind())
}

func TestFormatByFieldType_WrongKey(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByFieldType[string]("target", func(v string) slog.Value {
		return slog.StringValue("formatted")
	})

	val, ok := formatter(nil, slog.String("other", "value"))
	is.False(ok)
	is.Equal("value", val.String())
}

func TestFormatByFieldType_WrongType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByFieldType[int64]("key", func(v int64) slog.Value {
		return slog.Int64Value(v * 2)
	})

	// Key matches but type is string, not int64
	val, ok := formatter(nil, slog.String("key", "not-an-int"))
	is.False(ok)
	is.Equal("not-an-int", val.String())
}

func TestFormatByFieldType_Match(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByFieldType[string]("key", func(v string) slog.Value {
		return slog.StringValue("formatted_" + v)
	})

	val, ok := formatter(nil, slog.String("key", "hello"))
	is.True(ok)
	is.Equal("formatted_hello", val.String())
}

func TestFormatByGroup_PathMismatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroup([]string{"a", "b"}, func(attrs []slog.Attr) slog.Value {
		return slog.GroupValue(attrs...)
	})

	groupAttr := slog.Group("b", slog.String("inner", "val"))
	val, ok := formatter([]string{"x"}, groupAttr) // path is ["x","b"], want ["a","b"]
	is.False(ok)
	is.Equal(slog.KindGroup, val.Kind())
}

func TestFormatByGroup_Match(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroup([]string{"parent", "child"}, func(attrs []slog.Attr) slog.Value {
		return slog.StringValue("flattened")
	})

	groupAttr := slog.Group("child", slog.String("inner", "val"))
	val, ok := formatter([]string{"parent"}, groupAttr) // path is ["parent","child"]
	is.True(ok)
	is.Equal("flattened", val.String())
}

func TestFormatByGroup_NonGroupAttr(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroup([]string{"a"}, func(attrs []slog.Attr) slog.Value {
		return slog.StringValue("formatted")
	})

	// Non-group attr should not match
	val, ok := formatter(nil, slog.String("a", "val"))
	is.False(ok)
	is.Equal("val", val.String())
}

func TestFormatByGroupKey_Match(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroupKey([]string{"parent"}, "target", func(v slog.Value) slog.Value {
		return slog.StringValue("formatted_" + v.String())
	})

	val, ok := formatter([]string{"parent"}, slog.String("target", "hello"))
	is.True(ok)
	is.Equal("formatted_hello", val.String())
}

func TestFormatByGroupKey_GroupMismatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroupKey([]string{"parent"}, "target", func(v slog.Value) slog.Value {
		return slog.StringValue("formatted")
	})

	val, ok := formatter([]string{"other"}, slog.String("target", "hello"))
	is.False(ok)
	is.Equal("hello", val.String())
}

func TestFormatByGroupKey_KeyMismatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroupKey([]string{"parent"}, "target", func(v slog.Value) slog.Value {
		return slog.StringValue("formatted")
	})

	val, ok := formatter([]string{"parent"}, slog.String("other", "hello"))
	is.False(ok)
	is.Equal("hello", val.String())
}

func TestFormatByGroupKeyType_Match(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroupKeyType[string]([]string{"parent"}, "target", func(v string) slog.Value {
		return slog.StringValue("formatted_" + v)
	})

	val, ok := formatter([]string{"parent"}, slog.String("target", "hello"))
	is.True(ok)
	is.Equal("formatted_hello", val.String())
}

func TestFormatByGroupKeyType_TypeMismatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroupKeyType[int64]([]string{"parent"}, "target", func(v int64) slog.Value {
		return slog.Int64Value(v * 2)
	})

	// Key and group match, but type is string not int64
	val, ok := formatter([]string{"parent"}, slog.String("target", "not-int"))
	is.False(ok)
	is.Equal("not-int", val.String())
}

func TestFormatByGroupKeyType_GroupIsSkipped(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByGroupKeyType[string]([]string{"parent"}, "target", func(v string) slog.Value {
		return slog.StringValue("formatted")
	})

	groupAttr := slog.Group("target", slog.String("inner", "val"))
	val, ok := formatter([]string{"parent"}, groupAttr)
	is.False(ok)
	is.Equal(slog.KindGroup, val.Kind())
}

func TestFormatByKind_NoMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByKind(slog.KindDuration, func(v slog.Value) slog.Value {
		return slog.StringValue("formatted")
	})

	val, ok := formatter(nil, slog.String("key", "not-duration"))
	is.False(ok)
	is.Equal("not-duration", val.String())
}

func TestFormatByKey_NoMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByKey("target", func(v slog.Value) slog.Value {
		return slog.StringValue("formatted")
	})

	val, ok := formatter(nil, slog.String("other", "value"))
	is.False(ok)
	is.Equal("value", val.String())
}

func TestFormat_DeepNesting_10Levels(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var maxDepth int
	handler := NewFormatterMiddleware(
		Format(func(groups []string, key string, value slog.Value) slog.Value {
			if len(groups) > maxDepth {
				maxDepth = len(groups)
			}
			if key == "leaf" {
				return slog.StringValue("formatted_" + value.String())
			}
			return value
		}),
	)

	var checked int32
	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					atomic.AddInt32(&checked, 1)
					return nil
				},
			}.NewMockHandler(),
		),
	)

	// Build 10-level deep group
	leaf := slog.String("leaf", "value")
	var attr slog.Attr
	attr = slog.Group("g10", leaf)
	for i := 9; i >= 1; i-- {
		attr = slog.Group("g"+string(rune('0'+i)), attr)
	}

	logger.Info("test", attr)
	is.Equal(int32(1), atomic.LoadInt32(&checked))
	is.Equal(10, maxDepth)
}

func TestFormatByKey_TopLevelMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByKey("target", func(v slog.Value) slog.Value {
		return slog.StringValue("formatted_" + v.String())
	})

	val, ok := formatter(nil, slog.String("target", "hello"))
	is.True(ok)
	is.Equal("formatted_hello", val.String())
}

func TestFormatByKind_NestedNoMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// All children are strings, looking for Int64
	formatter := FormatByKind(slog.KindInt64, func(v slog.Value) slog.Value {
		return slog.Int64Value(v.Int64() * 2)
	})

	groupAttr := slog.Group("g", slog.String("a", "val"), slog.String("b", "val2"))
	val, ok := formatter(nil, groupAttr)
	is.False(ok)
	is.Equal(slog.KindGroup, val.Kind())
}

func TestFormatByType_NestedMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByType[time.Duration](func(v time.Duration) slog.Value {
		return slog.StringValue(v.String())
	})

	groupAttr := slog.Group("g",
		slog.String("a", "val"),
		slog.Duration("d", 3*time.Second),
	)
	val, ok := formatter(nil, groupAttr)
	is.True(ok)
	is.Equal(slog.KindGroup, val.Kind())
	group := val.Group()
	is.Len(group, 2)
	is.Equal("val", group[0].Value.String())
	is.Equal("3s", group[1].Value.String())
}

func TestFormatByType_NestedNoMatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	formatter := FormatByType[time.Duration](func(v time.Duration) slog.Value {
		return slog.StringValue(v.String())
	})

	groupAttr := slog.Group("g", slog.String("a", "val"))
	val, ok := formatter(nil, groupAttr)
	is.False(ok)
	is.Equal(slog.KindGroup, val.Kind())
}