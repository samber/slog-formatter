
# slog formatters

[![tag](https://img.shields.io/github/tag/samber/slog-formatter.svg)](https://github.com/samber/slog-formatter/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20.1-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/slog-formatter?status.svg)](https://pkg.go.dev/github.com/samber/slog-formatter)
![Build Status](https://github.com/samber/slog-formatter/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/slog-formatter)](https://goreportcard.com/report/github.com/samber/slog-formatter)
[![Coverage](https://img.shields.io/codecov/c/github/samber/slog-formatter)](https://codecov.io/gh/samber/slog-formatter)
[![Contributors](https://img.shields.io/github/contributors/samber/slog-formatter)](https://github.com/samber/slog-formatter/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/slog-formatter)](./LICENSE)

A toolset for pipelining formatters to [slog](https://pkg.go.dev/golang.org/x/exp/slog) loggers.

⚠️ In some case, you should consider implementing `slog.LogValuer` instead of using this library.

**See also:**

- [slog-multi](https://github.com/samber/slog-multi): workflows of `slog` handlers (pipeline, fanout, ...)
- [slog-datadog](https://github.com/samber/slog-datadog): A `slog` handler for `Datadog`
- [slog-logstash](https://github.com/samber/slog-logstash): A `slog` handler for `Logstash`
- [slog-slack](https://github.com/samber/slog-slack): A `slog` handler for `Slack`
- [slog-loki](https://github.com/samber/slog-loki): A `slog` handler for `Loki`
- [slog-sentry](https://github.com/samber/slog-sentry): A `slog` handler for `Sentry`
- [slog-fluentd](https://github.com/samber/slog-fluentd): A `slog` handler for `Fluentd`
- [slog-syslog](https://github.com/samber/slog-syslog): A `slog` handler for `Syslog`
- [slog-graylog](https://github.com/samber/slog-graylog): A `slog` handler for `Graylog`

## 🚀 Install

```sh
go get github.com/samber/slog-formatter
```

**Compatibility**: go >= 1.20.1

This library is v0 and follows SemVer strictly. On `slog` final release (go 1.21), this library will go v1.

No breaking changes will be made to exported APIs before v1.0.0.

## 🚀 Getting started

Here is a simple formatter that hides user email. 👇

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"golang.org/x/exp/slog"
)

formatter1 := slogformatter.FormatByKey("very_private_data", func(v slog.Value) slog.Value {
    return slog.StringValue("***********")
})
formatter2 := slogformatter.ErrorFormatter("error")
formatter3 := slogformatter.FormatByType(func(u User) slog.Value {
	return slog.StringValue(fmt.Sprintf("%s %s", u.firstname, u.lastname))
})

logger := slog.New(
    slogformatter.NewFormatterHandler(formatter1, formatter2, formatter3)(
        slog.NewTextHandler(os.Stdout),
    ),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.With("very_private_data", "abcd"),
    slog.With("user", user),
    slog.With("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

Using `slog-multi` pipelines:

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slog"
)

formatter1 := slogformatter.FormatByKey("very_private_data", func(v slog.Value) slog.Value {
    return slog.StringValue("***********")
})
formatter2 := slogformatter.ErrorFormatter("an error")
formatter3 := slogformatter.FormatByType(func(u User) slog.Value {
	return slog.StringValue(fmt.Sprintf("%s %s", u.firstname, u.lastname))
})

formattingMiddleware := slogformatter.NewFormatterHandler(formatter1, formatter2, formatter3)
sink := slog.HandlerOptions{}.NewJSONHandler(os.Stderr)

logger := slog.New(
    slogmulti.
        Pipe(formattingMiddleware).
        Handler(sink),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.With("very_private_data", "abcd"),
    slog.With("user", user),
    slog.With("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

## 💡 Spec

GoDoc: [https://pkg.go.dev/github.com/samber/slog-formatter](https://pkg.go.dev/github.com/samber/slog-formatter)

Handlers:
- [NewFormatterHandler](#NewFormatterHandler): main handler
- [NewFormatterMiddleware](#NewFormatterMiddleware): compatible with `slog-multi` middlewares

Common formattes:
- [PIIFormatter](#PIIFormatter): hide private Personal Identifiable Information (PII)
- [IPAddressFormatter](#IPAddressFormatter): hide ip address from logs
- [ErrorFormatter](#ErrorFormatter): transforms a go error into a readable error

Custom formatter:
- [Format](#Format): pass any attribute into a formatter
- [FormatByType](#FormatByType): pass attributes matching generic type into a formatter
- [FormatByKey](#FormatByKey): pass attributes matching key into a formatter
- [FormatByFieldType](#FormatByFieldType): pass attributes matching both key and generic type into a formatter
- [FormatByGroup](#FormatByGroup): pass attributes under a group into a formatter
- [FormatByGroupKey](#FormatByGroupKey): pass attributes under a group and matching key, into a formatter
- [FormatByGroupKeyType](#FormatByGroupKeyType): pass attributes under a group, matching key and matching a generic type, into a formatter

### NewFormatterHandler

Returns a slog.Handler that applies formatters to.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"golang.org/x/exp/slog"
)

type User struct {
	email     string
	firstname string
	lastname  string
}

formatter1 := slogformatter.FormatByKey("very_private_data", func(v slog.Value) slog.Value {
    return slog.StringValue("***********")
})
formatter2 := slogformatter.ErrorFormatter("error")
formatter3 := slogformatter.FormatByType(func(u User) slog.Value {
	return slog.StringValue(fmt.Sprintf("%s %s", u.firstname, u.lastname))
})

logger := slog.New(
    slogformatter.NewFormatterHandler(formatter1, formatter2, formatter3)(
        slog.NewTextHandler(os.StdErr),
    ),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.With("very_private_data", "abcd"),
    slog.With("user", user),
    slog.With("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

### NewFormatterMiddleware

Returns a `slog-multi` middleware that applies formatters to.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slog"
)

formatter1 := slogformatter.FormatByKey("very_private_data", func(v slog.Value) slog.Value {
    return slog.StringValue("***********")
})
formatter2 := slogformatter.ErrorFormatter("error")
formatter3 := slogformatter.FormatByType(func(u User) slog.Value {
	return slog.StringValue(fmt.Sprintf("%s %s", u.firstname, u.lastname))
})

formattingMiddleware := slogformatter.NewFormatterHandler(formatter1, formatter2, formatter3)
sink := slog.HandlerOptions{}.NewJSONHandler(os.Stderr)

logger := slog.New(
    slogmulti.
        Pipe(formattingMiddleware).
        Handler(sink),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.With("very_private_data", "abcd"),
    slog.With("user", user),
    slog.With("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

### PIIFormatter

Hides private Personal Identifiable Information (PII).

IDs are kept as is. Values longer than 5 characters have a plain text prefix.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"golang.org/x/exp/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.PIIFormatter("user"),
    )(
        slog.NewTextHandler(os.Stdout),
    ),
)

logger.
    With(
        slog.Group(
            "user",
            slog.String("id", "bd57ffbd-8858-4cc4-a93b-426cef16de61"),
            slog.String("email", "foobar@example.com"),
            slog.Group(
                "address",
                slog.String("street", "1st street"),
                slog.String("city", "New York"),
                slog.String("country", "USA"),
                slog.Int("zip", 12345),
            ),
        ),
    ).
    Error("an error")

// outputs:
// {
//   "time":"2023-04-10T14:00:0.000000+00:00",
//   "level": "ERROR",
//   "msg": "an error",
//   "user": {
//     "id": "bd57ffbd-8858-4cc4-a93b-426cef16de61",
//     "email": "foob*******",
//     "address": {
//       "street": "1st *******",
//       "city": "New *******",
//       "country": "*******",
//       "zip": "*******"
//     }
//   }
// }
```

### IPAddressFormatter

Transforms an IP address into "********".

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"golang.org/x/exp/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.IPAddressFormatter("ip_address"),
    )(
        slog.NewTextHandler(os.Stdout),
    ),
)

logger.
    With("ip_address", "1.2.3.4").
    Error("an error")

// outputs:
// {
//   "time":"2023-04-10T14:00:0.000000+00:00",
//   "level": "ERROR",
//   "msg": "an error",
//   "ip_address": "*******",
// }
```

### ErrorFormatter

Transforms a Go error into a readable error.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"golang.org/x/exp/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.ErrorFormatter("error"),
    )(
        slog.NewTextHandler(os.Stdout),
    ),
)

err := fmt.Errorf("an error")
logger.Error("a message", With("error", err))

// outputs:
// {
//   "time":"2023-04-10T14:00:0.000000+00:00",
//   "level": "ERROR",
//   "msg": "a message",
//   "error": {
//     "message": "an error",
//     "type": "*errors.errorString"
//   }
// }
```

### Format

Pass every attributes into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.Format(func(groups []string, key string, value slog.Value) slog.Value {
        // hide everything under "user" group
        if lo.Contains(groups, "user") {
            return slog.StringValue("****")
        }

        return value
    }),
)
```

### FormatByType

Pass attributes matching generic type into a formatter.

```go
slogformatter.NewFormatterHandler(
    // format a custom error type
    slogformatter.FormatByType[*customError](func(err *customError) slog.Value {
        return slog.GroupValue(
            slog.Int("code", err.code),
            slog.String("message", err.msg),
        )
    }),
    // format other errors
    slogformatter.FormatByType[error](func(err error) slog.Value {
        return slog.GroupValue(
            slog.Int("code", err.Error()),
            slog.String("type", reflect.TypeOf(err).String()),
        )
    }),
)
```

⚠️ Consider implementing `slog.LogValuer` when possible:

```go
type customError struct {
    ...
}

func (customError) Error() string {
    ...
}

func (customError) Error() string {
    ...
}

// implements slog.LogValuer
func (customError) LogValue() slog.Value {
	return slog.StringValue(...)
}
```

### FormatByKey

Pass attributes matching key into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.FormatByKey("abcd", func(value slog.Value) slog.Value {
        return ...
    }),
)
```

### FormatByFieldType

Pass attributes matching both key and generic type into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.FormatByFieldType[User]("user", func(u User) slog.Value {
        return ...
    }),
)
```

### FormatByGroup

Pass attributes under a group into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.FormatByGroup([]{"user", "address"}, func(attr []slog.Attr) slog.Value {
        return ...
    }),
)
```

### FormatByGroupKey

Pass attributes under a group and matching key, into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.FormatByGroupKey([]{"user", "address"}, "country", func(value slog.Value) slog.Value {
        return ...
    }),
)
```

### FormatByGroupKeyType

Pass attributes under a group, matching key and matching a generic type, into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.FormatByGroupKeyType[string]([]{"user", "address"}, "country", func(value string) slog.Value {
        return ...
    }),
)
```

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/slog-formatter)
- Fix [open issues](https://github.com/samber/slog-formatter/issues) or request new features

Don't hesitate ;)

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/slog-formatter)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2023 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
