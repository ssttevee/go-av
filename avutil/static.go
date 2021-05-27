// +build !av_dynamic

package avutil

// #cgo pkg-config: libavutil
//
// #include <libavutil/error.h>
import "C"

func getErrorString(e Error) string {
	var buf [64]C.char
	if C.av_strerror(C.int(e), &buf[0], 64) != 0 {
		return "unknown error"
	}

	return C.GoString(&buf[0])
}
