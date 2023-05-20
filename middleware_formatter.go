package slogformatter

import (
	slogmulti "github.com/samber/slog-multi"
)

// NewFormatterMiddleware returns slog-multi middleware.
func NewFormatterMiddleware(formatters ...Formatter) slogmulti.Middleware {
	return NewFormatterHandler(formatters...)
}
