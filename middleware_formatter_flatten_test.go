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

func TestFlattenAttrsWithPrefix_DeepNested_KnownBehavior(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	attrs := []slog.Attr{
		slog.Group("a",
			slog.Group("b",
				slog.String("leaf", "val"),
			),
		),
	}

	// Note: the double separator ("root.a..b.leaf") is a known quirk of the
	// current FlattenAttrsWithPrefix implementation. The inner recursive call
	// passes an empty prefix which produces an extra separator when concatenated.
	// This test documents the existing behavior to prevent accidental changes.
	result := FlattenAttrsWithPrefix(".", "root", attrs)
	is.Len(result, 1)
	is.Equal("root.a..b.leaf", result[0].Key)
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

func flattenOutput(t *testing.T, opts FlattenFormatterMiddlewareOptions, log func(logger *slog.Logger)) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	handler := opts.NewFlattenFormatterMiddlewareOptions()(slog.NewJSONHandler(&buf, nil))
	log(slog.New(handler))

	output := map[string]any{}
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON %q: %v", buf.String(), err)
	}
	delete(output, slog.TimeKey)
	delete(output, slog.LevelKey)
	delete(output, slog.MessageKey)
	return output
}

func TestFlattenFormatterMiddleware_Handle_RecordGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	output := flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: "_"}, func(logger *slog.Logger) {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "test",
			slog.Group("request", slog.String("foo", "bar"), slog.Int("baz", 42)))
	})
	is.Equal(map[string]any{"request_foo": "bar", "request_baz": float64(42)}, output)
}

func TestFlattenFormatterMiddleware_Handle_UnderWithGroup(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	output := flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: ".", Prefix: "attrs"}, func(logger *slog.Logger) {
		logger.
			With("email", "samuel@acme.org").
			WithGroup("group1").
			With("hello", "world").
			WithGroup("group2").
			Error("A message", "foo", "bar", slog.Group("a", slog.Group("b", slog.String("leaf", "val"))))
	})
	is.Equal(map[string]any{
		"attrs.email":                  "samuel@acme.org",
		"attrs.group1.hello":           "world",
		"attrs.group1.group2.foo":      "bar",
		"attrs.group1.group2.a.b.leaf": "val",
	}, output)
}

func TestFlattenFormatterMiddleware_WithAttrs_GroupKeys(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	group := slog.Group("request", slog.String("foo", "bar"))

	output := flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: "_"}, func(logger *slog.Logger) {
		logger.With(group).Info("test")
	})
	is.Equal(map[string]any{"request_foo": "bar"}, output)

	output = flattenOutput(t, FlattenFormatterMiddlewareOptions{Separator: "_", Prefix: "p"}, func(logger *slog.Logger) {
		logger.With(group).Info("test")
	})
	is.Equal(map[string]any{"p_request_foo": "bar"}, output)
}

func TestFlattenFormatterMiddleware_Handle_IgnorePath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	output := flattenOutput(t, FlattenFormatterMiddlewareOptions{IgnorePath: true, Prefix: "attrs"}, func(logger *slog.Logger) {
		logger.WithGroup("g").Info("test", slog.Group("a", slog.Group("b", slog.String("leaf", "val"))), "k", "v")
	})
	is.Equal(map[string]any{"leaf": "val", "k": "v"}, output)
}

func TestFlattenFormatterMiddleware_Handle_EmptyAttrs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	output := flattenOutput(t, FlattenFormatterMiddlewareOptions{Prefix: "attrs"}, func(logger *slog.Logger) {
		logger.Info("test", slog.Attr{}, slog.Group("empty"), slog.Group("", slog.String("inline", "v")))
	})
	is.Equal(map[string]any{"attrs.inline": "v"}, output)
}
