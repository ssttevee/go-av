package av

// #include <libavutil/error.h>
// #include <libavutil/common.h>
import "C"
import (
	"io"

	"github.com/pkg/errors"
)

type Error int

const (
	ErrAgain = Error(-C.EAGAIN)
	ErrNoMem = Error(-C.ENOMEM)
	ErrInval = Error(-C.EINVAL)

	ErrOptionNotFound = Error(C.AVERROR_OPTION_NOT_FOUND)
)

// arbitrary numbers that hopefully doesn't collide with any existing error codes
const (
	errInternalIOError Error = -1923754812 - iota
	errInternalFormatError
	errInternalCodecError
)

func (e Error) Error() string {
	switch e {
	case errInternalIOError:
		return "internal io error; placeholder for error in pinned file"
	case errInternalFormatError:
		return "internal format error; placeholder for error in pinned format context data"
	case errInternalCodecError:
		return "internal codec error; placeholder for error in pinned format context data"
	}

	var buf [64]C.char
	if C.av_strerror(C.int(e), &buf[0], 64) != 0 {
		return "unknown error"
	}

	return C.GoString(&buf[0])
}

func averror(code C.int) error {
	if code == 0 {
		return nil
	}

	if code == C.AVERROR_EOF {
		return io.EOF
	}

	if code == -C.ENOMEM {
		panic(ErrNoMem)
	}

	return errors.WithStack(Error(code))
}

func avreturn(code C.int) (int, error) {
	if code > 0 {
		return int(code), nil
	}

	return 0, averror(code)
}
