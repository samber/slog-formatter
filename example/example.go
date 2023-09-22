package main

import (
	"fmt"
	"os"
	"time"

	"log/slog"

	"github.com/samber/lo"
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
)

type myError struct {
	code int
	msg  string
}

func (e myError) Error() string {
	return e.msg
}

func fail() error {
	return &myError{42, "oops"}
}

type Company struct {
	CEO Person
	CTO Person
}

func (c Company) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("ceo", c.CEO),
		slog.Any("cto", c.CTO),
	)
}

type Person struct {
	firstname string
	lastname  string
}

func (p Person) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("%s %s", p.firstname, p.lastname))
}

func example() {
	logger := slog.New(
		slogformatter.NewFormatterHandler(
			slogformatter.FormatByType(func(e *myError) slog.Value {
				return slog.GroupValue(
					slog.Int("code", e.code),
					slog.String("message", e.msg),
				)
			}),
			slogformatter.ErrorFormatter("error_with_generic_formatter"),
			slogformatter.FormatByKey("email", func(v slog.Value) slog.Value {
				return slog.StringValue("***********")
			}),
			slogformatter.FormatByGroupKey([]string{"a-group"}, "hello", func(v slog.Value) slog.Value {
				return slog.StringValue("eve")
			}),
			slogformatter.FormatByGroup([]string{"hq"}, func(attrs []slog.Attr) slog.Value {
				return slog.GroupValue(
					slog.Group(
						"address",
						lo.ToAnySlice(attrs)...,
					),
				)
			}),
			slogformatter.PIIFormatter("hq"),
		)(
			slog.NewJSONHandler(os.Stdout, nil),
		),
	)

	logger.
		With(
			slog.Group(
				"companies",
				slog.Any(
					"acme",
					Company{
						CEO: Person{firstname: "Al", lastname: "ice"},
						CTO: Person{firstname: "Bo", lastname: "b"},
					},
				),
			),
		).
		With(
			slog.Group(
				"hq",
				slog.String("street", "1st street"),
				slog.String("city", "New York"),
				slog.String("country", "USA"),
				slog.Int("zip", 12345),
			),
		).
		With("email", "samuel@acme.org").
		With("environment", "dev").
		With("error_without_formatter", fmt.Errorf("an error")).
		With("error_with_generic_formatter", fmt.Errorf("an error")).
		With("error_with_formatter", fail()).
		WithGroup("a-group").
		With("hello", "world").
		Error("A message")
}

func exampleFlatten() {
	logger := slog.New(
		slogmulti.
			Pipe(slogformatter.FlattenFormatterMiddlewareOptions{Separator: ".", Prefix: "attrs", IgnorePath: false}.NewFlattenFormatterMiddlewareOptions()).
			Handler(slog.NewJSONHandler(os.Stdout, nil)),
	)

	logger.
		With("email", "samuel@acme.org").
		With("environment", "dev").
		WithGroup("group1").
		With("hello", "world").
		WithGroup("group2").
		With("hello", "world").
		Error("A message", "foo", "bar")
}

func exampleTime() {
	logger := slog.New(
		slogmulti.
			Pipe(
				slogformatter.NewFormatterHandler(
					slogformatter.TimeFormatter(time.DateTime, time.UTC),
				),
			).
			Handler(slog.NewJSONHandler(os.Stdout, nil)),
	)

	logger.
		With("time", time.Now()).
		Error("A message", "foo", "bar")
}

func main() {
	example()
	exampleFlatten()
	exampleTime()
}
