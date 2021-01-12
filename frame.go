package av

// #include <libavutil/frame.h>
import "C"
import (
	"runtime"

	"github.com/ssttevee/go-av/avutil"
)

type _frame = avutil.Frame

type Frame struct {
	*_frame
}

func NewFrame() *Frame {
	frame := avutil.NewFrame()
	if frame == nil {
		panic(avutil.ErrNoMem)
	}

	ret := &Frame{
		_frame: frame,
	}

	runtime.SetFinalizer(ret, func(f *Frame) {
		avutil.FreeFrame(&f._frame)
	})

	return ret
}

func (f *Frame) prepare() *avutil.Frame {
	f.Unref()

	return f._frame
}

func (f *Frame) Unref() {
	avutil.UnrefFrame(f._frame)
}

func (f *Frame) CopyTo(f2 *Frame) error {
	return averror(avutil.RefFrame(f._frame, f2._frame))
}

func (f *Frame) Clone() (*Frame, error) {
	clone := NewFrame()
	if err := f.CopyTo(clone); err != nil {
		return nil, err
	}

	return clone, nil
}

func (f *Frame) HwFramesCtx() *HWFramesContext {
	if f._frame.HwFramesCtx == nil {
		return nil
	}

	return newHWFramesContext(avutil.RefBuffer(f._frame.HwFramesCtx))
}
