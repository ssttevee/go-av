package avutil

// #include <libavutil/log.h>
//
// void goavLogSetup();
import "C"
import (
	"log"
	"sync"
)

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

var initLoggingOnce sync.Once

func initLogging() {
	initLoggingOnce.Do(func() {
		C.goavLogSetup()
	})
}

var Logger *log.Logger
var Verbosity = Info

//export goavLog
func goavLog(level C.int, cstr *C.char) {
	if level > C.int(Verbosity) {
		return
	}

	if Logger == nil {
		log.Print(C.GoString(cstr))
	} else {
		Logger.Print(C.GoString(cstr))
	}
}
