package av

// #include <libavutil/opt.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

type Option func(**C.AVDictionary) error

func StringOption(name, value string) Option {
	return func(pm **C.AVDictionary) error {
		cname := C.CString(name)
		defer C.free(unsafe.Pointer(cname))

		cvalue := C.CString(value)
		defer C.free(unsafe.Pointer(cvalue))

		return averror(C.av_dict_set(pm, cname, cvalue, 0))
	}
}

func resolveOptionsDict(opts ...Option) (*C.AVDictionary, error) {
	var dict *C.AVDictionary
	for _, opt := range opts {
		if err := opt(&dict); err != nil {
			C.av_dict_free(&dict)
			return nil, err
		}
	}

	return dict, nil
}

func setOption(ptr unsafe.Pointer, name string, value interface{}, searchFlags int) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	switch v := value.(type) {
	case string:
		cvalue := C.CString(v)
		defer C.free(unsafe.Pointer(cvalue))

		return averror(C.av_opt_set(ptr, cname, cvalue, C.int(searchFlags)))

	case int:
		return averror(C.av_opt_set_int(ptr, cname, C.int64_t(v), C.int(searchFlags)))

	case int64:
		return averror(C.av_opt_set_int(ptr, cname, C.int64_t(v), C.int(searchFlags)))

	case float64:
		return averror(C.av_opt_set_double(ptr, cname, C.double(v), C.int(searchFlags)))

	case Rational:
		return averror(C.av_opt_set_q(ptr, cname, v.ctype(), C.int(searchFlags)))

	case PixelFormat:
		return averror(C.av_opt_set_pixel_fmt(ptr, cname, C.enum_AVPixelFormat(v), C.int(searchFlags)))

	case []byte:
		defer runtime.KeepAlive(v)

		return averror(C.av_opt_set_bin(ptr, cname, (*C.uint8_t)(unsafe.Pointer(&v[0])), C.int(len(v)), C.int(searchFlags)))
	}

	panic(fmt.Sprintf("unexpected option value type: %T", value))
}

func getOption(ptr unsafe.Pointer, name string, searchFlags int) (interface{}, error) {
	type avoption struct {
		name   *C.char
		help   *C.char
		offset C.int
		_type  C.enum_AVOptionType
		defaultValue [8]byte
	}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var target unsafe.Pointer
	opt := (*avoption)(unsafe.Pointer(C.av_opt_find2(ptr, cname, nil, 0, C.int(searchFlags), &target)))
	if opt == nil || target == nil || (opt.offset <= 0 && opt._type != C.AV_OPT_TYPE_CONST) {
		return nil, ErrOptionNotFound
	}

	dst := unsafe.Pointer(uintptr(target) + uintptr(opt.offset))
	switch opt._type {
	case C.AV_OPT_TYPE_BOOL:
		return *(*C.int)(dst) != 0, nil

	case C.AV_OPT_TYPE_FLAGS, C.AV_OPT_TYPE_INT:
		return int(*(*C.int)(dst)), nil

	case C.AV_OPT_TYPE_INT64:
		return int64(*(*C.int64_t)(dst)), nil

	case C.AV_OPT_TYPE_UINT64:
		return uint64(*(*C.uint64_t)(dst)), nil

	case C.AV_OPT_TYPE_FLOAT, C.AV_OPT_TYPE_DOUBLE:
		return float64(*(*C.float)(dst)), nil

	case C.AV_OPT_TYPE_VIDEO_RATE, C.AV_OPT_TYPE_RATIONAL:
		return rational(*(*C.AVRational)(dst)), nil

	case C.AV_OPT_TYPE_CONST:
		return float64(*(*C.double)(unsafe.Pointer(&opt.defaultValue[0]))), nil
	}

	return nil, ErrInval
}
