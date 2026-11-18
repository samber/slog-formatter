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

// flatten turns nested groups into flat attributes. Unless IgnorePath is set,
// keys are joined with the separator and start with the prefix built by
// WithGroup, so record attributes and WithAttrs attributes get the same keys.
func (h *FlattenFormatterMiddleware) flatten(attrs []slog.Attr) []slog.Attr {
	prefix := h.option.Prefix
	if h.option.IgnorePath {
		prefix = ""
	}
	return flattenAttrsWithPath(h.option.IgnorePath, h.option.Separator, prefix, attrs, make([]slog.Attr, 0, len(attrs)))
}

func flattenAttrsWithPath(ignorePath bool, separator string, path string, attrs []slog.Attr, output []slog.Attr) []slog.Attr {
	for _, attr := range attrs {
		value := attr.Value.Resolve()
		key := attr.Key
		if !ignorePath && path != "" && key != "" {
			key = path + separator + key
		} else if !ignorePath && key == "" {
			key = path
		}

		if value.Kind() == slog.KindGroup {
			output = flattenAttrsWithPath(ignorePath, separator, key, value.Group(), output)
			continue
		}
		// Handlers ignore empty attributes; do not turn one into a prefixed key.
		if attr.Key == "" && value.Any() == nil {
			continue
		}

		output = append(output, slog.Attr{Key: key, Value: value})
	}

	return output
}

// Implements slog.Handler
func (h *FlattenFormatterMiddleware) WithGroup(name string) slog.Handler {
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
		switch attr.Value.Kind() {
		case slog.KindAny, slog.KindBool, slog.KindDuration, slog.KindFloat64, slog.KindInt64, slog.KindUint64, slog.KindString, slog.KindTime:
			output = append(output, attr)
		case slog.KindGroup:
			output = append(output, attr.Value.Group()...)
		case slog.KindLogValuer:
			output = append(output, slog.Any(attr.Key, attr.Value.Resolve().Any()))
		}
	}

	return output
}

// FlattenAttrsWithPrefix flatten attributes recursively, with prefix.
func FlattenAttrsWithPrefix(separator string, prefix string, attrs []slog.Attr) []slog.Attr {
	output := make([]slog.Attr, 0, len(attrs))

	for _, attr := range attrs {
		switch attr.Value.Kind() {
		case slog.KindAny, slog.KindBool, slog.KindDuration, slog.KindFloat64, slog.KindInt64, slog.KindUint64, slog.KindString, slog.KindTime:
			output = append(output, attr)
		case slog.KindGroup:
			output = append(output, PrefixAttrKeys(prefix+separator+attr.Key+separator, FlattenAttrsWithPrefix(separator, "", attr.Value.Group()))...)
		case slog.KindLogValuer:
			attr := slog.Any(attr.Key, attr.Value.Resolve().Any())
			output = append(output, FlattenAttrsWithPrefix(separator, prefix, []slog.Attr{attr})...)
		}
	}

	return output
}
