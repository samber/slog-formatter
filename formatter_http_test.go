package slogformatter

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"

	slogmock "github.com/samber/slog-mock"
	"github.com/stretchr/testify/assert"
)

func TestHTTPRequestFormatter_HideHeaders(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPRequestFormatter(true))

	req := &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{
			Scheme:   "https",
			Host:     "example.com",
			Path:     "/api/test",
			RawQuery: "foo=bar&baz=qux",
		},
		Header: http.Header{
			"Authorization": []string{"Bearer secret"},
			"Content-Type":  []string{"application/json"},
		},
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "request" && attr.Value.Kind() == slog.KindGroup {
							found := map[string]bool{}
							for _, a := range attr.Value.Group() {
								found[a.Key] = true
								switch a.Key {
								case "host":
									is.Equal("example.com", a.Value.String())
								case "method":
									is.Equal("GET", a.Value.String())
								case "headers":
									// ignoreHeaders=true, so headers is a string "[hidden]"
									is.Equal("[hidden]", a.Value.String())
								}
							}
							is.True(found["host"])
							is.True(found["method"])
							is.True(found["headers"])
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("request", req))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPRequestFormatter_ShowHeaders(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPRequestFormatter(false))

	req := &http.Request{
		Method: "POST",
		Host:   "api.example.com",
		URL: &url.URL{
			Scheme: "https",
			Host:   "api.example.com",
			Path:   "/v1/users",
		},
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "request" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "headers" && a.Value.Kind() == slog.KindGroup {
									for _, h := range a.Value.Group() {
										if h.Key == "Content-Type" {
											is.Equal("application/json", h.Value.String())
										}
									}
									atomic.AddInt32(&checked, 1)
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

	logger.Info("test", slog.Any("request", req))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPRequestFormatter_WithQueryParams(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPRequestFormatter(true))

	req := &http.Request{
		Method: "GET",
		Host:   "example.com",
		URL: &url.URL{
			Scheme:   "https",
			Host:     "example.com",
			Path:     "/search",
			RawQuery: "q=test&page=1",
		},
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "request" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "url" && a.Value.Kind() == slog.KindGroup {
									for _, u := range a.Value.Group() {
										if u.Key == "raw_query" {
											is.Equal("q=test&page=1", u.Value.String())
											atomic.AddInt32(&checked, 1)
										}
									}
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

	logger.Info("test", slog.Any("request", req))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPRequestFormatter_NonRequestUnchanged(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPRequestFormatter(true))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "request" {
							is.Equal("not-a-request", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("request", "not-a-request"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPResponseFormatter_HideHeaders(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPResponseFormatter(true))

	resp := &http.Response{
		StatusCode:    200,
		Status:        "200 OK",
		ContentLength: 1234,
		Uncompressed:  false,
		Header: http.Header{
			"X-Secret": []string{"should-be-hidden"},
		},
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "response" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								switch a.Key {
								case "status":
									is.Equal(int64(200), a.Value.Int64())
								case "status_text":
									is.Equal("200 OK", a.Value.String())
								case "content_length":
									is.Equal(int64(1234), a.Value.Int64())
								case "uncompressed":
									is.False(a.Value.Bool())
								case "headers":
									is.Equal("[hidden]", a.Value.String())
								}
							}
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("response", resp))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPResponseFormatter_ShowHeaders(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPResponseFormatter(false))

	resp := &http.Response{
		StatusCode:    404,
		Status:        "404 Not Found",
		ContentLength: 0,
		Uncompressed:  true,
		Header: http.Header{
			"Content-Type": []string{"text/html"},
		},
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "response" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "headers" && a.Value.Kind() == slog.KindGroup {
									for _, h := range a.Value.Group() {
										if h.Key == "Content-Type" {
											is.Equal("text/html", h.Value.String())
										}
									}
									atomic.AddInt32(&checked, 1)
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

	logger.Info("test", slog.Any("response", resp))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPResponseFormatter_NonResponseUnchanged(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPResponseFormatter(true))

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "response" {
							is.Equal("not-a-response", attr.Value.String())
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.String("response", "not-a-response"))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPRequestFormatter_EmptyURL(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPRequestFormatter(true))

	req := &http.Request{
		Method: "GET",
		Host:   "",
		URL:    &url.URL{},
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "request" && attr.Value.Kind() == slog.KindGroup {
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("request", req))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}

func TestHTTPResponseFormatter_ZeroStatus(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var checked int32
	handler := NewFormatterMiddleware(HTTPResponseFormatter(true))

	resp := &http.Response{
		StatusCode:    0,
		Status:        "",
		ContentLength: -1,
	}

	logger := slog.New(
		handler(
			slogmock.Option{
				Handle: func(ctx context.Context, record slog.Record) error {
					record.Attrs(func(attr slog.Attr) bool {
						if attr.Key == "response" && attr.Value.Kind() == slog.KindGroup {
							for _, a := range attr.Value.Group() {
								if a.Key == "status" {
									is.Equal(int64(0), a.Value.Int64())
								}
								if a.Key == "content_length" {
									is.Equal(int64(-1), a.Value.Int64())
								}
							}
							atomic.AddInt32(&checked, 1)
						}
						return true
					})
					return nil
				},
			}.NewMockHandler(),
		),
	)

	logger.Info("test", slog.Any("response", resp))
	is.Equal(int32(1), atomic.LoadInt32(&checked))
}
