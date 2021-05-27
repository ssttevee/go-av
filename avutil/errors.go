package avutil

// #include <libavutil/error.h>
// #include <libavutil/common.h>
import "C"
import (
	"github.com/ssttevee/go-av/internal/common"
)

type Error int

const (
	ErrAgain Error = -C.EAGAIN
	ErrNoMem Error = -C.ENOMEM
	ErrInval Error = -C.EINVAL

	ErrOptionNotFound Error = C.AVERROR_OPTION_NOT_FOUND
	ErrInvalidData    Error = C.AVERROR_INVALIDDATA
	ErrStreamNotFound Error = C.AVERROR_STREAM_NOT_FOUND
	ErrEOF            Error = C.AVERROR_EOF
)

func (e Error) Error() string {
	switch int(e) {
	case common.IOError:
		return "internal io error; placeholder for error in pinned file"
	case common.FormatError:
		return "internal format error; placeholder for error in pinned format context data"
	case common.CodecError:
		return "internal codec error; placeholder for error in pinned format context data"
	}

	return getErrorString(e)
}
