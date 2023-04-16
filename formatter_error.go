package slogformatter

import (
	"reflect"

	"golang.org/x/exp/slog"
)

// ErrorFormatter transforms a go error into a readable error.
//
// Example:
//
//	err := reader.Close()
//	err = fmt.Errorf("could not close reader: %v", err)
//	logger.With("error", reader.Close()).Log("error")
//
// passed to ErrorFormatter("error"), will be transformed into:
//
//	"error": {
//	  "message": "could not close reader: file already closed",
//	  "type": "*io.ErrClosedPipe"
//	}
func ErrorFormatter(fieldName string) Formatter {
	return FormatByFieldType(fieldName, func(err error) slog.Value {
		values := []slog.Attr{
			slog.String("message", err.Error()),
			slog.String("type", reflect.TypeOf(err).String()),
			// @TODO: inject stracktrace
		}

		return slog.GroupValue(values...)
	})
}
