
# slog: Attribute formatting

[![tag](https://img.shields.io/github/tag/samber/slog-formatter.svg)](https://github.com/samber/slog-formatter/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/slog-formatter?status.svg)](https://pkg.go.dev/github.com/samber/slog-formatter)
![Build Status](https://github.com/samber/slog-formatter/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/slog-formatter)](https://goreportcard.com/report/github.com/samber/slog-formatter)
[![Coverage](https://img.shields.io/codecov/c/github/samber/slog-formatter)](https://codecov.io/gh/samber/slog-formatter)
[![Contributors](https://img.shields.io/github/contributors/samber/slog-formatter)](https://github.com/samber/slog-formatter/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/slog-formatter)](./LICENSE)

Common formatters for [slog](https://pkg.go.dev/log/slog) library + helpers for building your own.

**Handlers:**
- [NewFormatterHandler](#NewFormatterHandler): main handler
- [NewFormatterMiddleware](#NewFormatterMiddleware): compatible with `slog-multi` middlewares

**Common formatters:**
- [TimeFormatter](#TimeFormatter): transforms a `time.Time` into a readable string
- [UnixTimestampFormatter](#UnixTimestampFormatter): transforms a `time.Time` into a unix timestamp.
- [TimezoneConverter](#TimezoneConverter): set a `time.Time` to a different timezone
- [ErrorFormatter](#ErrorFormatter): transforms a go error into a readable error
- [HTTPRequestFormatter](#HTTPRequestFormatter-and-HTTPResponseFormatter): transforms a *http.Request into a readable object
- [HTTPResponseFormatter](#HTTPRequestFormatter-and-HTTPResponseFormatter): transforms a *http.Response into a readable object
- [PIIFormatter](#PIIFormatter): hide private Personal Identifiable Information (PII)
- [IPAddressFormatter](#IPAddressFormatter): hide ip address from logs
- [FlattenFormatterMiddleware](#FlattenFormatterMiddleware): returns a formatter middleware that flatten attributes recursively

**Custom formatter:**
- [Format](#Format): pass any attribute into a formatter
- [FormatByKind](#FormatByKind): pass attributes matching `slog.Kind` into a formatter
- [FormatByType](#FormatByType): pass attributes matching generic type into a formatter
- [FormatByKey](#FormatByKey): pass attributes matching key into a formatter
- [FormatByFieldType](#FormatByFieldType): pass attributes matching both key and generic type into a formatter
- [FormatByGroup](#FormatByGroup): pass attributes under a group into a formatter
- [FormatByGroupKey](#FormatByGroupKey): pass attributes under a group and matching key, into a formatter
- [FormatByGroupKeyType](#FormatByGroupKeyType): pass attributes under a group, matching key and matching a generic type, into a formatter

**See also:**

- [slog-multi](https://github.com/samber/slog-multi): `slog.Handler` chaining, fanout, routing, failover, load balancing...
- [slog-formatter](https://github.com/samber/slog-formatter): `slog` attribute formatting
- [slog-sampling](https://github.com/samber/slog-sampling): `slog` sampling policy
- [slog-gin](https://github.com/samber/slog-gin): Gin middleware for `slog` logger
- [slog-echo](https://github.com/samber/slog-echo): Echo middleware for `slog` logger
- [slog-fiber](https://github.com/samber/slog-fiber): Fiber middleware for `slog` logger
- [slog-chi](https://github.com/samber/slog-chi): Chi middleware for `slog` logger
- [slog-datadog](https://github.com/samber/slog-datadog): A `slog` handler for `Datadog`
- [slog-rollbar](https://github.com/samber/slog-rollbar): A `slog` handler for `Rollbar`
- [slog-sentry](https://github.com/samber/slog-sentry): A `slog` handler for `Sentry`
- [slog-syslog](https://github.com/samber/slog-syslog): A `slog` handler for `Syslog`
- [slog-logstash](https://github.com/samber/slog-logstash): A `slog` handler for `Logstash`
- [slog-fluentd](https://github.com/samber/slog-fluentd): A `slog` handler for `Fluentd`
- [slog-graylog](https://github.com/samber/slog-graylog): A `slog` handler for `Graylog`
- [slog-loki](https://github.com/samber/slog-loki): A `slog` handler for `Loki`
- [slog-slack](https://github.com/samber/slog-slack): A `slog` handler for `Slack`
- [slog-telegram](https://github.com/samber/slog-telegram): A `slog` handler for `Telegram`
- [slog-mattermost](https://github.com/samber/slog-mattermost): A `slog` handler for `Mattermost`
- [slog-microsoft-teams](https://github.com/samber/slog-microsoft-teams): A `slog` handler for `Microsoft Teams`
- [slog-webhook](https://github.com/samber/slog-webhook): A `slog` handler for `Webhook`
- [slog-kafka](https://github.com/samber/slog-kafka): A `slog` handler for `Kafka`
- [slog-nats](https://github.com/samber/slog-nats): A `slog` handler for `NATS`
- [slog-parquet](https://github.com/samber/slog-parquet): A `slog` handler for `Parquet` + `Object Storage`
- [slog-zap](https://github.com/samber/slog-zap): A `slog` handler for `Zap`
- [slog-zerolog](https://github.com/samber/slog-zerolog): A `slog` handler for `Zerolog`
- [slog-logrus](https://github.com/samber/slog-logrus): A `slog` handler for `Logrus`
- [slog-channel](https://github.com/samber/slog-channel): A `slog` handler for Go channels

## 🚀 Install

```sh
go get github.com/samber/slog-formatter
```

**Compatibility**: go >= 1.21

No breaking changes will be made to exported APIs before v2.0.0.

⚠️ Warnings:
- in some case, you should consider implementing `slog.LogValuer` instead of using this library.
- use this library carefully, log processing can be very costly (!)

## 🚀 Getting started

The following example has 3 formatters that anonymize data, format errors and format user. 👇

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"log/slog"
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
        slog.NewTextHandler(os.Stdout, nil),
    ),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.Any("very_private_data", "abcd"),
    slog.Any("user", user),
    slog.Any("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

## 💡 Spec

GoDoc: [https://pkg.go.dev/github.com/samber/slog-formatter](https://pkg.go.dev/github.com/samber/slog-formatter)

### NewFormatterHandler

Returns a slog.Handler that applies formatters to.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"log/slog"
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
        slog.NewTextHandler(os.StdErr, nil),
    ),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.Any("very_private_data", "abcd"),
    slog.Any("user", user),
    slog.Any("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

### NewFormatterMiddleware

Returns a `slog-multi` middleware that applies formatters to.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"log/slog"
)

formatter1 := slogformatter.FormatByKey("very_private_data", func(v slog.Value) slog.Value {
    return slog.StringValue("***********")
})
formatter2 := slogformatter.ErrorFormatter("error")
formatter3 := slogformatter.FormatByType(func(u User) slog.Value {
	return slog.StringValue(fmt.Sprintf("%s %s", u.firstname, u.lastname))
})

formattingMiddleware := slogformatter.NewFormatterHandler(formatter1, formatter2, formatter3)
sink := slog.NewJSONHandler(os.Stderr, slog.HandlerOptions{})

logger := slog.New(
    slogmulti.
        Pipe(formattingMiddleware).
        Handler(sink),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.Any("very_private_data", "abcd"),
    slog.Any("user", user),
    slog.Any("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
```

### TimeFormatter

Transforms a `time.Time` into a readable string.

```go
slogformatter.NewFormatterHandler(
    slogformatter.TimeFormatter(time.DateTime, time.UTC),
)
```

### UnixTimestampFormatter

Transforms a `time.Time` into a unix timestamp.

```go
slogformatter.NewFormatterHandler(
    slogformatter.UnixTimestampFormatter(time.Millisecond),
)
```

### TimezoneConverter

Set a `time.Time` to a different timezone.

```go
slogformatter.NewFormatterHandler(
    slogformatter.TimezoneConverter(time.UTC),
)
```

### ErrorFormatter

Transforms a Go error into a readable error.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"log/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.ErrorFormatter("error"),
    )(
        slog.NewTextHandler(os.Stdout, nil),
    ),
)

err := fmt.Errorf("an error")
logger.Error("a message", slog.Any("error", err))

// outputs:
// {
//   "time":"2023-04-10T14:00:0.000000+00:00",
//   "level": "ERROR",
//   "msg": "a message",
//   "error": {
//     "message": "an error",
//     "type": "*errors.errorString"
//     "stacktrace": "main.main()\n\t/Users/samber/src/github.com/samber/slog-formatter/example/example.go:108 +0x1c\n"
//   }
// }
```

### HTTPRequestFormatter and HTTPResponseFormatter

Transforms *http.Request and *http.Response into readable objects.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"log/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.HTTPRequestFormatter(false),
        slogformatter.HTTPResponseFormatter(false),
    )(
        slog.NewJSONHandler(os.Stdout, nil),
    ),
)

req, _ := http.NewRequest(http.MethodGet, "https://api.screeb.app", nil)
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-TOKEN", "1234567890")

res, _ := http.DefaultClient.Do(req)

logger.Error("a message",
    slog.Any("request", req),
    slog.Any("response", res))
```

### PIIFormatter

Hides private Personal Identifiable Information (PII).

IDs are kept as is. Values longer than 5 characters have a plain text prefix.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	"log/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.PIIFormatter("user"),
    )(
        slog.NewTextHandler(os.Stdout, nil),
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
	"log/slog"
)

logger := slog.New(
    slogformatter.NewFormatterHandler(
        slogformatter.IPAddressFormatter("ip_address"),
    )(
        slog.NewTextHandler(os.Stdout, nil),
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

### FlattenFormatterMiddleware

A formatter middleware that flatten attributes recursively.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"log/slog"
)

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

// outputs:
// {
//   "time": "2023-05-20T22:14:55.857065+02:00",
//   "level": "ERROR",
//   "msg": "A message",
//   "attrs.email": "samuel@acme.org",
//   "attrs.environment": "dev",
//   "attrs.group1.hello": "world",
//   "attrs.group1.group2.hello": "world",
//   "foo": "bar"
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

### FormatByKind

Pass attributes matching `slog.Kind` into a formatter.

```go
slogformatter.NewFormatterHandler(
    slogformatter.FormatByKind(slog.KindDuration, func(value slog.Value) slog.Value {
        return ...
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
