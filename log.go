package av

// #include <stdlib.h>
// #include <libavutil/log.h>
//
// extern void goavLogSetup();
// extern void goavLogDirect(void *class_ptr, int level, char *fmt, va_list vl);
import "C"
import (
	"log"
)

func init() {
	C.goavLogSetup()
}

type VerbosityLevel int

const (
	Quiet   = VerbosityLevel(C.AV_LOG_QUIET)
	Panic   = VerbosityLevel(C.AV_LOG_PANIC)
	Fatal   = VerbosityLevel(C.AV_LOG_FATAL)
	Errors  = VerbosityLevel(C.AV_LOG_ERROR)
	Warning = VerbosityLevel(C.AV_LOG_WARNING)
	Info    = VerbosityLevel(C.AV_LOG_INFO)
	Verbose = VerbosityLevel(C.AV_LOG_VERBOSE)
	Debug   = VerbosityLevel(C.AV_LOG_DEBUG)
	Trace   = VerbosityLevel(C.AV_LOG_TRACE)
)

var logger *log.Logger
var Verbosity = Info

//export goavLog
func goavLog(level C.int, cstr *C.char) {
	if level > C.int(Verbosity) {
		return
	}

	if logger == nil {
		log.Print(C.GoString(cstr))
	} else {
		logger.Print(C.GoString(cstr))
	}
}

func SetLogger(l *log.Logger) {
	logger = l
}
