package av

import (
	"io"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-av/avutil"
)

func averror(code int32) error {
	if code == 0 {
		return nil
	}

	switch avutil.Error(code) {
	case avutil.ErrEOF:
		return io.EOF

	case avutil.ErrNoMem:
		panic(avutil.ErrNoMem)
	}

	return errors.WithStack(avutil.Error(code))
}

func avreturn(code int32) (int, error) {
	if code > 0 {
		return int(code), nil
	}

	return 0, averror(code)
}
