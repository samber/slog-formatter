package main

import (
	"fmt"
	"os"

	slogformatter "github.com/samber/slog-formatter"
	"golang.org/x/exp/slog"
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
						attrs...,
					),
				)
			}),
			slogformatter.PIIFormatter("hq"),
		)(
			slog.NewJSONHandler(os.Stdout),
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

func main() {
	example()
}
