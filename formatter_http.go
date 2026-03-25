package slogformatter

import (
	"net/http"
	"strings"

	"log/slog"
)

func headerToAttrs(header http.Header) []any {
	attrs := make([]any, 0, len(header))
	for key, values := range header {
		attrs = append(attrs, slog.String(key, strings.Join(values, ",")))
	}
	return attrs
}

// HTTPRequestFormatter transforms a *http.Request into a readable object.
func HTTPRequestFormatter(ignoreHeaders bool) Formatter {
	headers := slog.String("headers", "[hidden]")

	return FormatByType(func(req *http.Request) slog.Value {
		if !ignoreHeaders {
			headers = slog.Group("headers", headerToAttrs(req.Header)...)
		}

		queryParams := req.URL.Query()
		queryAttrs := make([]any, 0, len(queryParams))
		for key, values := range queryParams {
			queryAttrs = append(queryAttrs, slog.String(key, strings.Join(values, ",")))
		}

		return slog.GroupValue(
			slog.String("host", req.Host),
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Group(
				"url",
				slog.String("url", req.URL.String()),
				slog.String("scheme", req.URL.Scheme),
				slog.String("host", req.URL.Host),
				slog.String("path", req.URL.Path),
				slog.String("raw_query", req.URL.RawQuery),
				slog.String("fragment", req.URL.Fragment),
				slog.Group("query", queryAttrs...),
			),
			headers,
		)
	})
}

// HTTPResponseFormatter transforms a *http.Response into a readable object.
func HTTPResponseFormatter(ignoreHeaders bool) Formatter {
	headers := slog.String("headers", "[hidden]")

	return FormatByType(func(res *http.Response) slog.Value {
		if !ignoreHeaders {
			headers = slog.Group("headers", headerToAttrs(res.Header)...)
		}

		return slog.GroupValue(
			slog.Int("status", res.StatusCode),
			slog.String("status_text", res.Status),
			slog.Int64("content_length", res.ContentLength),
			slog.Bool("uncompressed", res.Uncompressed),
			headers,
		)
	})
}
