package av

// #include <libavutil/hwcontext.h>
// #include <libavutil/hwcontext_cuda.h>
import "C"
import (
	"runtime"
	"unsafe"
)

type HWDeviceType C.enum_AVHWDeviceType

const (
	Cuda = HWDeviceType(C.AV_HWDEVICE_TYPE_CUDA)
)

func (t HWDeviceType) ctype() C.enum_AVHWDeviceType {
	return C.enum_AVHWDeviceType(t)
}

type HWDeviceContext struct {
	ctx *C.AVBufferRef

	frames *HWFramesContext
}

func newHWDeviceContext(ctx *C.AVBufferRef) *HWDeviceContext {
	ret := &HWDeviceContext{ctx: ctx}

	runtime.SetFinalizer(ret, func(ctx *HWDeviceContext) {
		C.av_buffer_unref(&ctx.ctx)
	})

	return ret
}

func NewHWDeviceContext(deviceType HWDeviceType, device string) (*HWDeviceContext, error) {
	cdevice := C.CString(device)
	defer C.free(unsafe.Pointer(cdevice))

	var ctx *C.AVBufferRef
	if err := averror(C.av_hwdevice_ctx_create(&ctx, deviceType.ctype(), cdevice, nil, 0)); err != nil {
		return nil, err
	}

	return newHWDeviceContext(ctx), nil
}

func (ctx *HWDeviceContext) ref() *C.AVBufferRef {
	ref := C.av_buffer_ref(ctx.ctx)
	if ref == nil {
		panic(ErrNoMem)
	}

	return ref
}

type HWFramesContext struct {
	ctx *C.AVBufferRef
}

func newHWFramesContext(ctx *C.AVBufferRef) *HWFramesContext {
	if ctx == nil {
		panic(ErrNoMem)
	}

	ret := &HWFramesContext{ctx: ctx}

	runtime.SetFinalizer(ret, func(ctx *HWFramesContext) {
		C.av_buffer_unref(&ctx.ctx)
	})

	return ret
}

func NewHWFramesContext(deviceCtx *HWDeviceContext) *HWFramesContext {
	return newHWFramesContext(C.av_hwframe_ctx_alloc(deviceCtx.ctx))
}

func (ctx *HWFramesContext) ctype() *C.AVHWFramesContext {
	return (*C.AVHWFramesContext)(unsafe.Pointer(ctx.ctx.data))
}

func (ctx *HWFramesContext) PixelFormat() PixelFormat {
	return PixelFormat(ctx.ctype().format)
}

func (ctx *HWFramesContext) SetPixelFormat(format PixelFormat) {
	hwFramesContext := ctx.ctype()
	hwFramesContext.format = format.ctype()
}

func (ctx *HWFramesContext) SWPixelFormat() PixelFormat {
	return PixelFormat(ctx.ctype().sw_format)
}

func (ctx *HWFramesContext) SetSWPixelFormat(format PixelFormat) {
	hwFramesContext := ctx.ctype()
	hwFramesContext.sw_format = format.ctype()
}

func (ctx *HWFramesContext) Width() int {
	return int(ctx.ctype().width)
}

func (ctx *HWFramesContext) SetWidth(width int) {
	hwFramesContext := ctx.ctype()
	hwFramesContext.width = C.int(width)
}

func (ctx *HWFramesContext) Height() int {
	return int(ctx.ctype().height)
}

func (ctx *HWFramesContext) SetHeight(height int) {
	hwFramesContext := ctx.ctype()
	hwFramesContext.height = C.int(height)
}

func (ctx *HWFramesContext) Init() error {
	return averror(C.av_hwframe_ctx_init(ctx.ctx))
}

func (ctx *HWFramesContext) Eq(ctx2 *HWFramesContext) bool {
	if (ctx == nil) != (ctx2 == nil) {
		return false
	}

	if ctx == nil && ctx2 == nil {
		return true
	}

	return ctx.ctx.data == ctx2.ctx.data
}

func (ctx *HWFramesContext) ref() *C.AVBufferRef {
	ref := C.av_buffer_ref(ctx.ctx)
	if ref == nil {
		panic(ErrNoMem)
	}

	return ref
}
