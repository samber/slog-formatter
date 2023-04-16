package slogformatter

import (
	"time"

	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

// TimeFormatter transforms a `time.Time` into a readable string.
func TimeFormatter(timeFormat string, location *time.Location) Formatter {
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	return FormatByKind(slog.KindTime, func(value slog.Value) slog.Value {
		t := value.Time()

		if location != nil {
			t = t.In(location)
		}

		return slog.StringValue(t.Format(timeFormat))
	})
}

// UnixTimestampFormatter transforms a `time.Time` into a unix timestamp.
func UnixTimestampFormatter(precision time.Duration) Formatter {
	if !lo.Contains([]time.Duration{time.Nanosecond, time.Microsecond, time.Millisecond, time.Second}, precision) {
		panic("slog-formatter: unexpected precision")
	}

	return FormatByKind(slog.KindTime, func(value slog.Value) slog.Value {
		t := value.Time()

		switch precision {
		case time.Nanosecond:
			return slog.Int64Value(t.UnixNano())
		case time.Microsecond:
			return slog.Int64Value(t.UnixMicro())
		case time.Millisecond:
			return slog.Int64Value(t.UnixMilli())
		case time.Second:
			return slog.Int64Value(t.Unix())
		}

		panic("slog-formatter: unexpected precision")
	})
}

// TimezoneConverter set a `time.Time` to a different timezone.
func TimezoneConverter(location *time.Location) Formatter {
	return FormatByKind(slog.KindTime, func(value slog.Value) slog.Value {
		t := value.Time()

		if location == nil {
			location = time.UTC
		}

		return slog.TimeValue(t.In(location))
	})
}
