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
	// newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)

	// record.Attrs(func(attr slog.Attr) bool {
	// 	if !h.option.IgnorePath {
	// 		newRecord.AddAttrs(
	// 			PrefixAttrKeys(
	// 				lo.Ternary(h.option.Prefix == "" || h.option.IgnorePath, "", h.option.Prefix+h.option.Separator),
	// 				FlattenAttrsWithPrefix(h.option.Separator, h.option.Prefix, []slog.Attr{attr}),
	// 			)...,
	// 		)
	// 	} else {
	// 		newRecord.AddAttrs(FlattenAttrs([]slog.Attr{attr})...)
	// 	}
	// 	return true
	// })

	// return h.next.Handle(ctx, newRecord)
	return h.next.Handle(ctx, record)
}

// Implements slog.Handler
func (h *FlattenFormatterMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	if !h.option.IgnorePath {
		attrs = PrefixAttrKeys(
			lo.Ternary(h.option.Prefix == "" || h.option.IgnorePath, "", h.option.Prefix+h.option.Separator),
			FlattenAttrsWithPrefix(h.option.Separator, h.option.Prefix, attrs),
		)
	} else {
		attrs = FlattenAttrs(attrs)
	}

	return &FlattenFormatterMiddleware{
		next:   h.next.WithAttrs(attrs),
		option: h.option,
	}
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
	return lo.Map(attrs, func(item slog.Attr, _ int) slog.Attr {
		return slog.Attr{
			Key:   prefix + item.Key,
			Value: item.Value,
		}
	})
}

// FlattenAttrs flatten attributes recursively.
func FlattenAttrs(attrs []slog.Attr) []slog.Attr {
	output := []slog.Attr{}

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
	output := []slog.Attr{}

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
