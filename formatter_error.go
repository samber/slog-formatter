package slogformatter

import (
	"reflect"
	"runtime"
	"strconv"
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
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.Function, "log/slog") {
			b.WriteString(frame.Function)
			b.WriteString("\n\t")
			b.WriteString(frame.File)
			b.WriteByte(':')
			b.WriteString(strconv.Itoa(frame.Line))
			b.WriteByte('\n')
		}
		if !more {
			break
		}
	}
	return b.String()
}
