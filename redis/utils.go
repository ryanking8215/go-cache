package redis

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/ryanking8215/go-cache"
)

func notRedisError(err error) bool {
	switch err {
	case context.Canceled, context.DeadlineExceeded:
		return true
	case io.EOF:
		return true
	}
	if _, ok := err.(net.Error); ok {
		return true
	}

	return false
}

func timeUnixNanoToString(t time.Time) string {
	return strconv.FormatInt(t.UnixNano(), 10)
}

func usePrecise(dur time.Duration) bool {
	return dur < time.Second || dur%time.Second != 0
}

// toString converts i to a string, encoder is as a fallback method if we can't handle it by default
// Copy from [goframe](https://github.com/gogf/gf/), thanks for it.
func toString(i interface{}, encoder cache.Encoder) string {
	if i == nil {
		return ""
	}
	switch value := i.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case []byte:
		return string(value)
	case *time.Time:
		if value == nil {
			return ""
		}
		return value.String()
	default:
		// Empty checks.
		if value == nil {
			return ""
		}
		if f, ok := value.(interface{ String() string }); ok {
			// If the variable implements the String() interface,
			// then use that interface to perform the conversion
			return f.String()
		} else if f, ok := value.(error); ok {
			// If the variable implements the Error() interface,
			// then use that interface to perform the conversion
			return f.Error()
		} else {
			// Reflect checks.
			rv := reflect.ValueOf(value)
			kind := rv.Kind()
			switch kind {
			case reflect.Chan,
				reflect.Map,
				reflect.Slice,
				reflect.Func,
				reflect.Ptr,
				reflect.Interface,
				reflect.UnsafePointer:
				if rv.IsNil() {
					return ""
				}
			}
			if kind == reflect.Ptr {
				return toString(rv.Elem().Interface(), encoder)
			}
			// Finally we use encoder to convert.
			if b, err := encoder.Encode(value); err != nil {
				return fmt.Sprint(value)
			} else {
				return string(b)
			}
		}
	}
}
