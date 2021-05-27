package av

import (
	"runtime"
	"unsafe"

	"github.com/ssttevee/go-av/avutil"
)

type HWDeviceContext struct {
	ctx *avutil.BufferRef

	frames *HWFramesContext
}

func newHWDeviceContext(ctx *avutil.BufferRef) *HWDeviceContext {
	ret := &HWDeviceContext{ctx: ctx}

	runtime.SetFinalizer(ret, func(ctx *HWDeviceContext) {
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		buf := (*avutil.BufferRef)(ctx.ctx)
		avutil.UnrefBuffer(&buf)
		ctx.ctx = buf
	})

	return ret
}

func NewHWDeviceContext(deviceType avutil.HWDeviceType, device string) (*HWDeviceContext, error) {
	var ctx *avutil.BufferRef
	if err := averror(avutil.NewHWDeviceContext(&ctx, deviceType, device, nil, 0)); err != nil {
		return nil, err
	}

	return newHWDeviceContext(ctx), nil
}

func (ctx *HWDeviceContext) ref() *avutil.BufferRef {
	ref := avutil.RefBuffer(ctx.ctx)
	if ref == nil {
		panic(avutil.ErrNoMem)
	}

	return ref
}

type _hwFramesContext = avutil.HWFramesContext

type HWFramesContext struct {
	*_hwFramesContext
	buf *avutil.BufferRef
}

func newHWFramesContext(buf *avutil.BufferRef) *HWFramesContext {
	if buf == nil {
		panic(avutil.ErrNoMem)
	}

	ret := &HWFramesContext{
		_hwFramesContext: (*avutil.HWFramesContext)(unsafe.Pointer(buf.Data)),
		buf:              buf,
	}

	runtime.SetFinalizer(ret, func(ctx *HWFramesContext) {
		ctx._hwFramesContext = nil
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		buf := (*avutil.BufferRef)(ctx.buf)
		avutil.UnrefBuffer(&buf)
		ctx.buf = buf
	})

	return ret
}

func NewHWFramesContext(deviceCtx *HWDeviceContext) *HWFramesContext {
	return newHWFramesContext(avutil.NewHWFramesContext(deviceCtx.ctx))
}

func (ctx *HWFramesContext) Init() error {
	return averror(avutil.InitHWFramesContext(ctx.buf))
}

func (ctx *HWFramesContext) Eq(ctx2 *HWFramesContext) bool {
	if (ctx == nil) != (ctx2 == nil) {
		return false
	}

	if ctx == nil && ctx2 == nil {
		return true
	}

	return ctx.buf.Data == ctx2.buf.Data
}

func (ctx *HWFramesContext) ref() *avutil.BufferRef {
	ref := avutil.RefBuffer(ctx.buf)
	if ref == nil {
		panic(avutil.ErrNoMem)
	}

	return ref
}
