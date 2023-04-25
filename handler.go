package slogformatter

import (
	"context"

	"golang.org/x/exp/slog"
)

type FormatterHandler struct {
	groups     []string
	formatters []Formatter
	handler    slog.Handler
}

// NewFormatterHandler returns a slog.Handler that applies formatters to.
func NewFormatterHandler(formatters ...Formatter) func(slog.Handler) slog.Handler {
	return func(handler slog.Handler) slog.Handler {
		return &FormatterHandler{
			groups:     []string{},
			formatters: formatters,
			handler:    handler,
		}
	}
}

// Enabled implements slog.Handler.
func (h *FormatterHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

// Handle implements slog.Handler.
func (h *FormatterHandler) Handle(ctx context.Context, r slog.Record) error {
	r2 := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	r.Attrs(func(attr slog.Attr) bool {
		r2.AddAttrs(h.transformAttr(h.groups, attr))
		return true
	})

	return h.handler.Handle(ctx, r2)
}

// WithAttrs implements slog.Handler.
func (h *FormatterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	attrs = h.transformAttrs(h.groups, attrs)

	return &FormatterHandler{
		groups:     h.groups,
		formatters: h.formatters,
		handler:    h.handler.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler.
func (h *FormatterHandler) WithGroup(name string) slog.Handler {
	return &FormatterHandler{
		groups:     append(h.groups, name),
		formatters: h.formatters,
		handler:    h.handler.WithGroup(name),
	}
}

func (h *FormatterHandler) transformAttrs(groups []string, attrs []slog.Attr) []slog.Attr {
	for i := range attrs {
		attrs[i] = h.transformAttr(groups, attrs[i])
	}
	return attrs
}

func (h *FormatterHandler) transformAttr(groups []string, attr slog.Attr) slog.Attr {
	for attr.Value.Kind() == slog.KindLogValuer {
		attr.Value = attr.Value.LogValuer().LogValue()
	}

	for _, formatter := range h.formatters {
		if v, ok := formatter(groups, attr); ok {
			attr.Value = v
		}
	}

	return attr
}
