package slogformatter

import (
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

type LogValuerFunc func(any) (slog.Value, bool)
type Formatter func(groups []string, attr slog.Attr) (slog.Value, bool)

// Format pass every attributes into a formatter.
func Format[T any](formatter func([]string, string, slog.Value) slog.Value) Formatter {
	return func(groups []string, attr slog.Attr) (slog.Value, bool) {
		return formatter(groups, attr.Key, attr.Value), true
	}
}

// FormatByType pass attributes matching generic type into a formatter.
func FormatByType[T any](formatter func(T) slog.Value) Formatter {
	return func(_ []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() == slog.KindGroup {
			return value, false
		}

		if v, ok := value.Any().(T); ok {
			return formatter(v), true
		}

		return value, false
	}
}

// FormatByKind pass attributes matching `slog.Kind` into a formatter.
func FormatByKind(kind slog.Kind, formatter func(slog.Value) slog.Value) Formatter {
	return func(_ []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() == kind {
			return value, false
		}

		return value, false
	}
}

// FormatByKey pass attributes matching key into a formatter.
func FormatByKey(key string, formatter func(slog.Value) slog.Value) Formatter {
	return func(_ []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if attr.Key != key {
			return value, false
		}

		return formatter(value), true

	}
}

// FormatByFieldType pass attributes matching both key and generic type into a formatter.
func FormatByFieldType[T any](key string, formatter func(T) slog.Value) Formatter {
	return func(_ []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() == slog.KindGroup || attr.Key != key {
			return value, false
		}

		if v, ok := value.Any().(T); ok {
			return formatter(v), true
		}

		return value, false
	}
}

// FormatByGroup pass attributes under a group into a formatter.
func FormatByGroup(groups []string, formatter func([]slog.Attr) slog.Value) Formatter {
	return func(currentGroup []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() != slog.KindGroup || !slices.Equal(groups, append(currentGroup, attr.Key)) {
			return value, false
		}

		return formatter(value.Group()), true
	}
}

// FormatByGroupKey pass attributes under a group and matching key, into a formatter.
func FormatByGroupKey(groups []string, key string, formatter func(slog.Value) slog.Value) Formatter {
	return func(currentGroup []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if !slices.Equal(groups, currentGroup) || attr.Key != key {
			return value, false
		}

		return formatter(value), true
	}
}

// FormatByGroupKeyType pass attributes under a group, matching key and matching a generic type, into a formatter.
func FormatByGroupKeyType[T any](groups []string, key string, formatter func(T) slog.Value) Formatter {
	return func(currentGroup []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() == slog.KindGroup || !slices.Equal(groups, currentGroup) || attr.Key != key {
			return value, false
		}

		if v, ok := value.Any().(T); ok {
			return formatter(v), true
		}

		return value, false
	}
}
