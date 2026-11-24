package slogformatter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefixAttrKeys_Basic(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.String("a", "1"),
		slog.String("b", "2"),
	}

	result := PrefixAttrKeys("prefix.", attrs)
	is.Len(result, 2)
	is.Equal("prefix.a", result[0].Key)
	is.Equal("1", result[0].Value.String())
	is.Equal("prefix.b", result[1].Key)
	is.Equal("2", result[1].Value.String())
}

func TestPrefixAttrKeys_EmptyPrefix(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{slog.String("a", "1")}
	result := PrefixAttrKeys("", attrs)
	is.Len(result, 1)
	is.Equal("a", result[0].Key)
}

func TestPrefixAttrKeys_EmptyAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result := PrefixAttrKeys("prefix.", nil)
	is.Empty(result)
}

func TestFlattenAttrs_FlatAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.String("a", "1"),
		slog.Int("b", 2),
		slog.Bool("c", true),
	}

	result := FlattenAttrs(attrs)
	is.Len(result, 3)
	is.Equal("a", result[0].Key)
	is.Equal("b", result[1].Key)
	is.Equal("c", result[2].Key)
}

func TestFlattenAttrs_NestedGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.Group("g",
			slog.String("inner1", "val1"),
			slog.String("inner2", "val2"),
		),
	}

	result := FlattenAttrs(attrs)
	is.Len(result, 2)
	is.Equal("inner1", result[0].Key)
	is.Equal("val1", result[0].Value.String())
	is.Equal("inner2", result[1].Key)
	is.Equal("val2", result[1].Value.String())
}

func TestFlattenAttrs_MixedTypes(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.String("flat", "val"),
		slog.Group("g", slog.String("nested", "val2")),
	}

	result := FlattenAttrs(attrs)
	is.Len(result, 2)
	is.Equal("flat", result[0].Key)
	is.Equal("nested", result[1].Key)
}

func TestFlattenAttrs_EmptyGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{slog.Group("empty")}
	result := FlattenAttrs(attrs)
	is.Empty(result)
}

func TestFlattenAttrs_AllKinds(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.Any("any", struct{ X int }{42}),
		slog.Bool("bool", true),
		slog.Duration("dur", 5),
		slog.Float64("float", 3.14),
		slog.Int64("int", 99),
		slog.Uint64("uint", 100),
		slog.String("str", "hello"),
		slog.Time("time", time.Now()),
		slog.Group("group", slog.String("nested", "val")),
	}

	result := FlattenAttrs(attrs)
	// 7 scalar kinds + 1 time + 1 group (flattened to its child "nested") = 9
	is.Len(result, 9)
	// Verify group was flattened (last element should be the nested attr)
	is.Equal("nested", result[8].Key)
}

func TestFlattenAttrsWithPrefix_Simple(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.String("a", "1"),
		slog.Int("b", 2),
	}

	result := FlattenAttrsWithPrefix(".", "root", attrs)
	is.Len(result, 2)
	is.Equal("a", result[0].Key)
	is.Equal("b", result[1].Key)
}

func TestFlattenAttrsWithPrefix_Nested(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.Group("g",
			slog.String("inner", "val"),
		),
	}

	result := FlattenAttrsWithPrefix(".", "root", attrs)
	is.Len(result, 1)
	is.Equal("root.g.inner", result[0].Key)
	is.Equal("val", result[0].Value.String())
}

func TestFlattenAttrsWithPrefix_DeepNested(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.Group("a",
			slog.Group("b",
				slog.String("leaf", "val"),
			),
		),
	}

	result := FlattenAttrsWithPrefix(".", "root", attrs)
	is.Len(result, 1)
	is.Equal("root.a.b.leaf", result[0].Key)
	is.Equal("val", result[0].Value.String())
}

func TestFlattenAttrsWithPrefix_CustomSeparator(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.Group("g",
			slog.String("inner", "val"),
		),
	}

	result := FlattenAttrsWithPrefix("_", "pfx", attrs)
	is.Len(result, 1)
	is.Equal("pfx_g_inner", result[0].Key)
}

func TestFlattenFormatterMiddleware_DefaultSeparator(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{
		Separator: "",
	}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	fh := handler.(*FlattenFormatterMiddleware)
	is.Equal(".", fh.option.Separator)
}

func TestFlattenFormatterMiddleware_CustomSeparator(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{
		Separator: "_",
	}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	fh := handler.(*FlattenFormatterMiddleware)
	is.Equal("_", fh.option.Separator)
}

func TestFlattenFormatterMiddleware_WithGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{
		Prefix: "root",
	}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	h2 := handler.WithGroup("sub")
	fh := h2.(*FlattenFormatterMiddleware)
	is.Equal("root.sub", fh.option.Prefix)
}

func TestFlattenFormatterMiddleware_WithGroup_IgnorePath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{
		Prefix:     "root",
		IgnorePath: true,
	}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	h2 := handler.WithGroup("sub")
	fh := h2.(*FlattenFormatterMiddleware)
	is.Equal("sub", fh.option.Prefix)
}

func TestFlattenFormatterMiddleware_WithGroup_NoPrefix(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	h2 := handler.WithGroup("first")
	fh := h2.(*FlattenFormatterMiddleware)
	is.Equal("first", fh.option.Prefix)
}

func TestFlattenFormatterMiddleware_WithAttrs_IgnorePath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{
		IgnorePath: true,
	}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	// Should flatten without path prefixes
	h2 := handler.WithAttrs([]slog.Attr{
		slog.Group("g", slog.String("inner", "val")),
	})
	is.NotNil(h2)
}

func TestFlattenFormatterMiddleware_WithAttrs_WithPath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{
		Prefix: "root",
	}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	handler := middleware(slog.NewJSONHandler(io.Discard, nil))

	h2 := handler.WithAttrs([]slog.Attr{
		slog.String("key", "val"),
	})
	is.NotNil(h2)
}

func TestFlattenFormatterMiddleware_Enabled(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	opts := FlattenFormatterMiddlewareOptions{}
	middleware := opts.NewFlattenFormatterMiddlewareOptions()
	inner := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn})
	handler := middleware(inner)

	is.False(handler.Enabled(context.Background(), slog.LevelInfo))
	is.True(handler.Enabled(context.Background(), slog.LevelWarn))
	is.True(handler.Enabled(context.Background(), slog.LevelError))
}

// flattenOutput runs fn against a logger built on the flatten middleware and
// returns the JSON record that reached the sink, without time and level.
func flattenOutput(t *testing.T, opts FlattenFormatterMiddlewareOptions, fn func(l *slog.Logger)) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	sink := slog.NewJSONHandler(&buf, nil)
	fn(slog.New(opts.NewFlattenFormatterMiddlewareOptions()(sink)))

	out := map[string]any{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	delete(out, "time")
	delete(out, "level")
	delete(out, "msg")
	return out
}

func TestFlattenFormatterMiddleware_Handle_Group(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	out := flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: "_"}, func(l *slog.Logger) {
		l.LogAttrs(context.Background(), slog.LevelInfo, "test",
			slog.Group("request", slog.String("foo", "bar"), slog.Int("baz", 42)),
		)
	})

	is.Equal(map[string]any{"request_foo": "bar", "request_baz": float64(42)}, out)
}

func TestFlattenFormatterMiddleware_Handle_NestedGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	out := flattenOutput(t, FlattenFormatterMiddlewareOptions{}, func(l *slog.Logger) {
		l.Info("test", slog.Group("a", slog.Group("b", slog.String("leaf", "val")), slog.String("x", "y")))
	})

	is.Equal(map[string]any{"a.b.leaf": "val", "a.x": "y"}, out)
}

func TestFlattenFormatterMiddleware_Handle_UsesGroupAndPrefix(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	out := flattenOutput(t, FlattenFormatterMiddlewareOptions{Prefix: "attrs"}, func(l *slog.Logger) {
		l.WithGroup("group1").WithGroup("group2").Error("test", "foo", "bar", slog.Group("g", slog.Int("n", 1)))
	})

	is.Equal(map[string]any{
		"attrs.group1.group2.foo": "bar",
		"attrs.group1.group2.g.n": float64(1),
	}, out)
}

func TestFlattenFormatterMiddleware_Handle_IgnorePath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	out := flattenOutput(t, FlattenFormatterMiddlewareOptions{IgnorePath: true}, func(l *slog.Logger) {
		l.WithGroup("ignored").Info("test", slog.Group("a", slog.Group("b", slog.String("leaf", "val"))), "top", 1)
	})

	is.Equal(map[string]any{"leaf": "val", "top": float64(1)}, out)
}

func TestFlattenFormatterMiddleware_Handle_DoesNotMutateRecord(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	mw := FlattenFormatterMiddlewareOptions{}.NewFlattenFormatterMiddlewareOptions()
	h := mw(slog.NewJSONHandler(io.Discard, nil))

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	record.AddAttrs(slog.Group("g", slog.String("k", "v")))
	is.NoError(h.Handle(context.Background(), record))

	var keys []string
	record.Attrs(func(a slog.Attr) bool {
		keys = append(keys, a.Key)
		return true
	})
	is.Equal([]string{"g"}, keys)
}

func TestFlattenFormatterMiddleware_WithAttrs_NoDoublePrefix(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	out := flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: "_"}, func(l *slog.Logger) {
		l.WithGroup("request").With("foo", "bar").With(slog.Group("in", slog.Int("baz", 42))).Info("test")
	})

	is.Equal(map[string]any{"request_foo": "bar", "request_in_baz": float64(42)}, out)
}

func TestFlattenFormatterMiddleware_WithAttrs_GroupWithoutPrefix(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	out := flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: "_"}, func(l *slog.Logger) {
		l.With(slog.Group("g", slog.Int("a", 1))).Info("test")
	})

	is.Equal(map[string]any{"g_a": float64(1)}, out)
}

func TestFlattenFormatterMiddleware_WithGroup_Empty(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	handler := FlattenFormatterMiddlewareOptions{Prefix: "root"}.NewFlattenFormatterMiddlewareOptions()(slog.NewJSONHandler(io.Discard, nil))
	is.Same(handler, handler.WithGroup(""))
}

type flattenGroupValuer struct{}

func (flattenGroupValuer) LogValue() slog.Value {
	return slog.GroupValue(slog.String("id", "42"), slog.Group("inner", slog.Bool("ok", true)))
}

func TestFlattenAttrs_NestedGroupIsFlattenedRecursively(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result := FlattenAttrs([]slog.Attr{
		slog.Group("a", slog.Group("b", slog.String("leaf", "val")), slog.String("x", "y")),
	})

	require.Len(t, result, 2)
	is.Equal("leaf", result[0].Key)
	is.Equal("x", result[1].Key)
	for _, attr := range result {
		is.NotEqual(slog.KindGroup, attr.Value.Kind())
	}
}

func TestFlattenAttrs_LogValuerGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result := FlattenAttrs([]slog.Attr{slog.Any("user", flattenGroupValuer{})})

	require.Len(t, result, 2)
	is.Equal("id", result[0].Key)
	is.Equal("ok", result[1].Key)
}

func TestFlattenAttrsWithPrefix_LogValuerGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result := FlattenAttrsWithPrefix(".", "root", []slog.Attr{slog.Any("user", flattenGroupValuer{})})

	require.Len(t, result, 2)
	is.Equal("root.user.id", result[0].Key)
	is.Equal("root.user.inner.ok", result[1].Key)
}

func TestFlattenAttrsWithPrefix_EmptyPrefix(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result := FlattenAttrsWithPrefix(".", "", []slog.Attr{slog.Group("g", slog.String("k", "v"))})

	require.Len(t, result, 1)
	is.Equal("g.k", result[0].Key)
}
