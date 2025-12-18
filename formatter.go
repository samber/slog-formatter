package slogformatter

import (
	"log/slog"

	"slices"
)

type LogValuerFunc func(any) (slog.Value, bool)
type Formatter func(groups []string, attr slog.Attr) (slog.Value, bool)

// Format returns a Formatter that applies formatter to each attribute,
// recursively traversing nested groups. The groups slice contains the keys
// of enclosing groups, from outermost to innermost.
func Format(formatter func([]string, string, slog.Value) slog.Value) Formatter {
	var formatRecursive func([]string, slog.Attr) slog.Value
	formatRecursive = func(groups []string, attr slog.Attr) slog.Value {
		value := attr.Value

		if value.Kind() == slog.KindGroup {
			attrs := make([]slog.Attr, 0, len(value.Group()))
			nestedGroups := make([]string, len(groups)+1)
			copy(nestedGroups, groups)
			nestedGroups[len(groups)] = attr.Key

			for _, nestedAttr := range value.Group() {
				formattedValue := formatRecursive(nestedGroups, nestedAttr)
				attrs = append(attrs, slog.Attr{Key: nestedAttr.Key, Value: formattedValue})
			}

			return slog.GroupValue(attrs...)
		}

		return formatter(groups, attr.Key, value)
	}

	return func(groups []string, attr slog.Attr) (slog.Value, bool) {
		return formatRecursive(groups, attr), true
	}
}

// FormatByType pass attributes matching generic type into a formatter.
// This function performs recursive lookup through nested groups to find matching types.
func FormatByType[T any](formatter func(T) slog.Value) Formatter {
	var formatRecursive func([]string, slog.Attr) (slog.Value, bool)
	formatRecursive = func(groups []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() == slog.KindGroup {
			updated := false
			attrs := make([]slog.Attr, 0, len(value.Group()))
			nestedGroups := make([]string, len(groups)+1)
			copy(nestedGroups, groups)
			nestedGroups[len(groups)] = attr.Key

			for _, nestedAttr := range value.Group() {
				if nestedFormatted, ok := formatRecursive(nestedGroups, nestedAttr); ok {
					attrs = append(attrs, slog.Attr{Key: nestedAttr.Key, Value: nestedFormatted})
					updated = true
				} else {
					attrs = append(attrs, nestedAttr)
				}
			}

			if updated {
				return slog.GroupValue(attrs...), true
			}
			return value, false
		}

		if v, ok := value.Any().(T); ok {
			return formatter(v), true
		}

		return value, false
	}

	return formatRecursive
}

// FormatByKind pass attributes matching `slog.Kind` into a formatter.
// This function performs recursive lookup through nested groups to find matching kinds.
func FormatByKind(kind slog.Kind, formatter func(slog.Value) slog.Value) Formatter {
	var formatRecursive func([]string, slog.Attr) (slog.Value, bool)
	formatRecursive = func(groups []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if value.Kind() == slog.KindGroup {
			updated := false
			attrs := make([]slog.Attr, 0, len(value.Group()))
			nestedGroups := make([]string, len(groups)+1)
			copy(nestedGroups, groups)
			nestedGroups[len(groups)] = attr.Key

			for _, nestedAttr := range value.Group() {
				if nestedFormatted, ok := formatRecursive(nestedGroups, nestedAttr); ok {
					attrs = append(attrs, slog.Attr{Key: nestedAttr.Key, Value: nestedFormatted})
					updated = true
				} else {
					attrs = append(attrs, nestedAttr)
				}
			}

			if updated {
				return slog.GroupValue(attrs...), true
			}
			return value, false
		}

		if value.Kind() == kind {
			return formatter(value), true
		}

		return value, false
	}

	return formatRecursive
}

// FormatByKey pass attributes matching key into a formatter.
// This function performs recursive lookup through nested groups to find matching keys.
func FormatByKey(key string, formatter func(slog.Value) slog.Value) Formatter {
	var formatRecursive func([]string, slog.Attr) (slog.Value, bool)
	formatRecursive = func(groups []string, attr slog.Attr) (slog.Value, bool) {
		value := attr.Value

		if attr.Key == key {
			return formatter(value), true
		}

		if value.Kind() == slog.KindGroup {
			updated := false
			attrs := make([]slog.Attr, 0, len(value.Group()))
			nestedGroups := make([]string, len(groups)+1)
			copy(nestedGroups, groups)
			nestedGroups[len(groups)] = attr.Key

			for _, nestedAttr := range value.Group() {
				if nestedFormatted, ok := formatRecursive(nestedGroups, nestedAttr); ok {
					attrs = append(attrs, slog.Attr{Key: nestedAttr.Key, Value: nestedFormatted})
					updated = true
				} else {
					attrs = append(attrs, nestedAttr)
				}
			}

			if updated {
				return slog.GroupValue(attrs...), true
			}
			return value, false
		}

		return value, false
	}

	return formatRecursive
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
