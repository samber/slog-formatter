package slogformatter

import (
	"strings"

	"log/slog"

	"github.com/samber/lo"
)

// IPAddressFormatter transforms an IP address into "********".
//
// Example:
//
//	"context": {
//	  "ip_address": "bd57ffbd-8858-4cc4-a93b-426cef16de61"
//	}
//
// passed to IPAddressFormatter("ip_address"), will be transformed into:
//
//	"context": {
//	  "ip_address": "********",
//	}
func IPAddressFormatter(key string) Formatter {
	return FormatByKey(key, func(v slog.Value) slog.Value {
		return slog.StringValue("*******")
	})
}

// PIIFormatter transforms any value under provided key into "********".
// IDs are kept as is.
//
// Example:
//
//	  "user": {
//	    "id": "bd57ffbd-8858-4cc4-a93b-426cef16de61",
//	    "email": "foobar@example.com",
//	    "address": {
//	      "street": "1st street",
//		     "city": "New York",
//	      "country": USA",
//		     "zip": 123456
//	    }
//	  }
//
// passed to PIIFormatter("user"), will be transformed into:
//
//	  "user": {
//	    "id": "bd57ffbd-8858-4cc4-a93b-426cef16de61",
//	    "email": "foob********",
//	    "address": {
//	      "street": "1st *******",
//		     "city": "New *******",
//	  	   "country": "*******",
//		     "zip": "*******"
//	    }
//	  }
func PIIFormatter(key string) Formatter {
	return FormatByKey(key, func(v slog.Value) slog.Value {
		return recursivelyHidePII(key, v)
	})
}

func recursivelyHidePII(key string, v slog.Value) slog.Value {
	if v.Kind() == slog.KindLogValuer {
		v = v.LogValuer().LogValue()
	}

	if v.Kind() == slog.KindGroup {
		return slog.GroupValue(
			lo.Map(v.Group(), func(item slog.Attr, _ int) slog.Attr {
				return slog.Any(item.Key, recursivelyHidePII(item.Key, item.Value))
			})...,
		)
	}

	key = strings.ToLower(key)
	if key == "id" || strings.HasSuffix(key, "_id") || strings.HasSuffix(key, "-id") {
		return v
	}

	if v.Kind() != slog.KindString {
		return slog.StringValue("*******")
	}

	if len(v.String()) <= 5 {
		return slog.StringValue("*******")
	}

	return slog.StringValue(v.String()[0:4] + "*******")
}
