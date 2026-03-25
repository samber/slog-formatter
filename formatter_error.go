package slogformatter

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"log/slog"
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
			slog.String("stacktrace", stacktrace()),
		}

		return slog.GroupValue(values...)
	})
}

func stacktrace() string {
	var pcs [32]uintptr
	n := runtime.Callers(1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.Function, "log/slog") {
			fmt.Fprintf(&b, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		}
		if !more {
			break
		}
	}
	return b.String()
}
