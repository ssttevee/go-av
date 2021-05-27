// +build av_dynamic

package avutil

// #cgo LDFLAGS: -ldl
//
// #include <stdlib.h>
// #include <stdint.h>
// #include <stdarg.h>
// #include <dlfcn.h>
//
// static void *handle = 0;
//
// static int (*_av_strerror)(int, char*, size_t);
//
// int av_strerror(int err, char *buf, size_t len) {
//     return _av_strerror(err, buf, len);
// }
//
// static void (*_av_log_set_callback)(void (*callback)(void *, int, const char *, va_list));
//
// void av_log_set_callback(void (*callback)(void *, int, const char *, va_list)) {
//     _av_log_set_callback(callback);
// }
//
// static void (*_av_log_format_line)(void *, int, const char *, va_list, char *, int, int *);
//
// void av_log_format_line(void *ptr, int level, const char *fmt, va_list vl, char *line, int line_size, int *print_prefix) {
//     _av_log_format_line(ptr, level, fmt, vl, line, line_size, print_prefix);
// }
//
// void load_avutil_2() {
//     handle = dlopen("libavutil.so", RTLD_NOW | RTLD_GLOBAL);
//     _av_strerror = dlsym(handle, "av_strerror");
//     _av_log_set_callback = dlsym(handle, "av_log_set_callback");
//     _av_log_format_line = dlsym(handle, "av_log_format_line");
// }
import "C"
import "sync"

var initLib2Once sync.Once

func dynamicInit2() {
	initLib2Once.Do(func() {
		C.load_avutil_2()
	})
}

func getErrorString(e Error) string {
	dynamicInit2()
	var buf [64]C.char
	if C.av_strerror(C.int(e), &buf[0], 64) != 0 {
		return "unknown error"
	}

	return C.GoString(&buf[0])
}
