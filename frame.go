package av

// #include <libavutil/frame.h>
import "C"
import (
	"runtime"
)

type Frame struct {
	frame *C.AVFrame
}

func NewFrame() *Frame {
	frame := C.av_frame_alloc()
	if frame == nil {
		panic(ErrNoMem)
	}

	ret := &Frame{
		frame: frame,
	}

	runtime.SetFinalizer(ret, func(f *Frame) {
		C.av_frame_free(&f.frame)
	})

	return ret
}

func (f *Frame) prepare() *C.AVFrame {
	f.Unref()

	return f.frame
}

func (f *Frame) Unref() {
	defer runtime.KeepAlive(f)

	C.av_frame_unref(f.frame)
}

func (f *Frame) CopyTo(f2 *Frame) error {
	return averror(C.av_frame_ref(f.frame, f2.frame))
}

func (f *Frame) Clone() (*Frame, error) {
	clone := NewFrame()
	if err := f.CopyTo(clone); err != nil {
		return nil, err
	}

	return clone, nil
}

func (f *Frame) HWFramesContext() *HWFramesContext {
	if f.frame.hw_frames_ctx == nil {
		return nil
	}

	return newHWFramesContext(C.av_buffer_ref(f.frame.hw_frames_ctx))
}
