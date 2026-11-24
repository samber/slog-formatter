package slogformatter

import (
	"context"

	"log/slog"

	"github.com/samber/lo"
	slogmulti "github.com/samber/slog-multi"
)

type FlattenFormatterMiddlewareOptions struct {
	// Ignore attribute path and therefore ignore attribute key prefix.
	// Some attribute keys may collide.
	IgnorePath bool
	// Separator between prefix and key.
	Separator string
	// Attribute key prefix.
	Prefix string
}

// NewFlattenFormatterMiddlewareOptions returns a formatter middleware that flatten attributes recursively.
func (o FlattenFormatterMiddlewareOptions) NewFlattenFormatterMiddlewareOptions() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &FlattenFormatterMiddleware{
			next: next,
			option: FlattenFormatterMiddlewareOptions{
				IgnorePath: o.IgnorePath,
				Separator:  lo.Ternary(o.Separator == "", ".", o.Separator),
				Prefix:     o.Prefix,
			},
		}
	}
}

var _ slog.Handler = (*FlattenFormatterMiddleware)(nil)

type FlattenFormatterMiddleware struct {
	next   slog.Handler
	option FlattenFormatterMiddlewareOptions
}

// Implements slog.Handler
func (h *FlattenFormatterMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Implements slog.Handler
func (h *FlattenFormatterMiddleware) Handle(ctx context.Context, record slog.Record) error {
	if record.NumAttrs() == 0 {
		return h.next.Handle(ctx, record)
	}

	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(h.flatten(attrs)...)

	return h.next.Handle(ctx, newRecord)
}

// Implements slog.Handler
func (h *FlattenFormatterMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FlattenFormatterMiddleware{
		next:   h.next.WithAttrs(h.flatten(attrs)),
		option: h.option,
	}
}

// Implements slog.Handler
func (h *FlattenFormatterMiddleware) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	prefix := h.option.Prefix + h.option.Separator
	if h.option.IgnorePath || h.option.Prefix == "" {
		prefix = ""
	}

	return &FlattenFormatterMiddleware{
		next: h.next,
		option: FlattenFormatterMiddlewareOptions{
			IgnorePath: h.option.IgnorePath,
			Separator:  h.option.Separator,
			Prefix:     prefix + name,
		},
	}
}

// flatten flattens attrs and prefixes their keys with the current group path.
func (h *FlattenFormatterMiddleware) flatten(attrs []slog.Attr) []slog.Attr {
	if h.option.IgnorePath {
		return FlattenAttrs(attrs)
	}

	prefix := ""
	if h.option.Prefix != "" {
		prefix = h.option.Prefix + h.option.Separator
	}

	return PrefixAttrKeys(prefix, FlattenAttrsWithPrefix(h.option.Separator, "", attrs))
}

// PrefixAttrKeys prefix attribute keys.
func PrefixAttrKeys(prefix string, attrs []slog.Attr) []slog.Attr {
	result := make([]slog.Attr, len(attrs))
	for i, item := range attrs {
		result[i] = slog.Attr{
			Key:   prefix + item.Key,
			Value: item.Value,
		}
	}
	return result
}

// FlattenAttrs flatten attributes recursively.
func FlattenAttrs(attrs []slog.Attr) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs))

	for _, attr := range attrs {
		attr.Value = attr.Value.Resolve()

		if attr.Value.Kind() == slog.KindGroup {
			output = append(output, FlattenAttrs(attr.Value.Group())...)
		} else {
			output = append(output, attr)
		}
	}

	return output
}

// FlattenAttrsWithPrefix flatten attributes recursively, with prefix.
// The prefix and the separator are only added to the keys of attributes
// that were nested in a group. Top level attributes keep their key.
func FlattenAttrsWithPrefix(separator string, prefix string, attrs []slog.Attr) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs))

	for _, attr := range attrs {
		attr.Value = attr.Value.Resolve()

		if attr.Value.Kind() == slog.KindGroup {
			output = append(output, flattenGroup(separator, joinKeys(separator, prefix, attr.Key), attr.Value.Group())...)
		} else {
			output = append(output, attr)
		}
	}

	return output
}

// flattenGroup flattens the members of a group and prefixes every key with the group path.
func flattenGroup(separator string, path string, attrs []slog.Attr) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs))

	for _, attr := range attrs {
		attr.Value = attr.Value.Resolve()

		if attr.Value.Kind() == slog.KindGroup {
			output = append(output, flattenGroup(separator, joinKeys(separator, path, attr.Key), attr.Value.Group())...)
		} else {
			attr.Key = joinKeys(separator, path, attr.Key)
			output = append(output, attr)
		}
	}

	return output
}

// joinKeys joins two key parts, ignoring the empty ones.
func joinKeys(separator string, left string, right string) string {
	if left == "" {
		return right
	}
	if right == "" {
		return left
	}
	return left + separator + right
}
