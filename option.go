package av

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ssttevee/go-av/avutil"
)

type Option func(**avutil.Dictionary) error

func StringOption(name, value string) Option {
	return func(pm **avutil.Dictionary) error {
		return averror(avutil.SetDict(pm, name, value, 0))
	}
}

func resolveOptionsDict(opts ...Option) (*avutil.Dictionary, error) {
	var dict *avutil.Dictionary
	for _, opt := range opts {
		if err := opt(&dict); err != nil {
			avutil.FreeDict(&dict)
			return nil, err
		}
	}

	return dict, nil
}

func setOption(ptr unsafe.Pointer, name string, value interface{}, searchFlags int32) error {
	switch v := value.(type) {
	case string:
		return averror(avutil.SetOpt(ptr, name, v, searchFlags))

	case int:
		return averror(avutil.SetOptInt(ptr, name, int64(v), searchFlags))

	case int64:
		return averror(avutil.SetOptInt(ptr, name, v, searchFlags))

	case float64:
		return averror(avutil.SetOptDouble(ptr, name, v, searchFlags))

	case avutil.Rational:
		return averror(avutil.SetOptRational(ptr, name, v, searchFlags))

	case avutil.PixelFormat:
		return averror(avutil.SetOptPixelFormat(ptr, name, v, searchFlags))

	case []byte:
		defer runtime.KeepAlive(v)

		return averror(avutil.SetOptBin(ptr, name, &v[0], int32(len(v)), searchFlags))
	}

	panic(fmt.Sprintf("unexpected option value type: %T", value))
}

func getOption(ptr unsafe.Pointer, name string, searchFlags int32) (interface{}, error) {
	var target unsafe.Pointer
	opt := avutil.FindOpt(ptr, name, "", 0, searchFlags, &target)
	if opt == nil || target == nil || (opt.Offset <= 0 && opt.Type != avutil.OptionTypeConst) {
		return nil, avutil.ErrOptionNotFound
	}

	dst := unsafe.Pointer(uintptr(target) + uintptr(opt.Offset))
	switch opt.Type {
	case avutil.OptionTypeBool:
		return *(*int)(dst) != 0, nil

	case avutil.OptionTypeFlags, avutil.OptionTypeInt:
		return *(*int)(dst), nil

	case avutil.OptionTypeInt64:
		return *(*int64)(dst), nil

	case avutil.OptionTypeUint64:
		return *(*uint64)(dst), nil

	case avutil.OptionTypeFloat, avutil.OptionTypeDouble:
		return *(*float64)(dst), nil

	case avutil.OptionTypeVideoRate, avutil.OptionTypeRational:
		return *(*avutil.Rational)(dst), nil

	case avutil.OptionTypeConst:
		return *(*float64)(unsafe.Pointer(&opt.DefaultVal[0])), nil
	}

	return nil, avutil.ErrInval
}
