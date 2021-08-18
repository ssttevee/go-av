package av

// #include <stdint.h>
//
// struct AVCodecContext;
//
// extern int goavCodecContextGetFormat(struct AVCodecContext *, int *);
import "C"
import (
	"reflect"
	"runtime/cgo"
	"sync"
	"unsafe"

	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avutil"
)

type CodecNotFoundError string

func (e CodecNotFoundError) Error() string {
	return "codec not found: " + string(e)
}

type _codec = avcodec.Codec

type Codec struct {
	*_codec
}

func FindDecoderCodecByID(codecID avcodec.ID) (*Codec, error) {
	codec := avcodec.FindDecoder(codecID)
	if codec == nil {
		return nil, CodecNotFoundError(codecID.String())
	}

	return &Codec{
		_codec: codec,
	}, nil
}

func FindDecoderCodecByName(name string) (*Codec, error) {
	codec := avcodec.FindDecoderByName(name)
	if codec == nil {
		return nil, CodecNotFoundError(name)
	}

	return &Codec{
		_codec: codec,
	}, nil
}

func FindEncoderCodecByID(codecID avcodec.ID) (*Codec, error) {
	codec := avcodec.FindEncoder(codecID)
	if codec == nil {
		return nil, CodecNotFoundError(codecID.String())
	}

	return &Codec{
		_codec: codec,
	}, nil
}

func FindEncoderCodecByName(name string) (*Codec, error) {
	codec := avcodec.FindEncoderByName(name)
	if codec == nil {
		return nil, CodecNotFoundError(name)
	}

	return &Codec{
		_codec: codec,
	}, nil
}

func (c *Codec) Name() string {
	return c._codec.Name.String()
}

func countPixelFormats(fmts *avutil.PixelFormat) int {
	if fmts == nil {
		return 0
	}

	size := unsafe.Sizeof(*fmts)

	var n int
	ptr := uintptr(unsafe.Pointer(fmts))
	for *(*avutil.PixelFormat)(unsafe.Pointer(ptr)) != -1 {
		n++
		ptr += size
	}

	return n
}

func pixelFormatSlice(ptr *avutil.PixelFormat) []avutil.PixelFormat {
	n := countPixelFormats(ptr)
	return *(*[]avutil.PixelFormat)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(ptr)),
		Len:  n,
		Cap:  n,
	}))
}

func (c *Codec) NbPixFmts() int {
	return countPixelFormats(c._codec.PixFmts)
}

func (c *Codec) PixFmts() []avutil.PixelFormat {
	return pixelFormatSlice(c._codec.PixFmts)
}

func (c *Codec) PixFmt(i int) avutil.PixelFormat {
	return *(*avutil.PixelFormat)(unsafe.Pointer(uintptr(unsafe.Pointer(c._codec.PixFmts)) + unsafe.Sizeof(*c._codec.PixFmts)*uintptr(i)))
}

func countSampleFormats(fmts *avutil.SampleFormat) int {
	if fmts == nil {
		return 0
	}

	size := unsafe.Sizeof(*fmts)

	var n int
	ptr := uintptr(unsafe.Pointer(fmts))
	for *(*avutil.SampleFormat)(unsafe.Pointer(ptr)) != -1 {
		n++
		ptr += size
	}

	return n
}

func sampleFormatSlice(ptr *avutil.SampleFormat) []avutil.SampleFormat {
	n := countSampleFormats(ptr)
	return *(*[]avutil.SampleFormat)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(ptr)),
		Len:  n,
		Cap:  n,
	}))
}

func (c *Codec) NbSampleFmts() int {
	if c._codec.SampleFmts == nil {
		return 0
	}

	size := unsafe.Sizeof(*c._codec.SampleFmts)

	var n int
	ptr := uintptr(unsafe.Pointer(c._codec.SampleFmts))
	for *(*avutil.SampleFormat)(unsafe.Pointer(ptr)) != -1 {
		n++
		ptr += size
	}

	return n
}

func (c *Codec) SampleFmts() []avutil.SampleFormat {
	return sampleFormatSlice(c._codec.SampleFmts)
}

func (c *Codec) SampleFmt(i int) avutil.SampleFormat {
	return *(*avutil.SampleFormat)(unsafe.Pointer(uintptr(unsafe.Pointer(c._codec.SampleFmts)) + unsafe.Sizeof(*c._codec.SampleFmts)*uintptr(i)))
}

type _codecParameters = avcodec.Parameters

type CodecParameters struct {
	*_codecParameters
}

type pinnedCodecContextData struct {
	err error

	getFormatFunc func([]avutil.PixelFormat) avutil.PixelFormat
}

func unwrapPinnedCodecContextData(p unsafe.Pointer) *pinnedCodecContextData {
	return cgo.Handle(p).Value().(*pinnedCodecContextData)
}

//export goavCodecContextGetFormat
func goavCodecContextGetFormat(ctx *C.struct_AVCodecContext, choices *C.int) int32 {
	return int32(unwrapPinnedCodecContextData((*avcodec.Context)(unsafe.Pointer(ctx)).Opaque).getFormatFunc(pixelFormatSlice((*avutil.PixelFormat)(unsafe.Pointer(choices)))))
}

type _codecContext = avcodec.Context

type codecContext struct {
	*_codecContext

	pinnedDataOnce sync.Once

	initOnce sync.Once
	initErr  error
}

func newCodecContext(codec *Codec, params *CodecParameters) (*avcodec.Context, error) {
	ctx := avcodec.NewContext(codec._codec)
	if ctx == nil {
		panic(avutil.ErrNoMem)
	}

	if params != nil {
		if err := averror(avcodec.ParametersToContext(ctx, params._codecParameters)); err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

func (ctx *codecContext) pinnedData() *pinnedCodecContextData {
	ctx.pinnedDataOnce.Do(func() {
		ctx.Opaque = unsafe.Pointer(cgo.NewHandle(&pinnedCodecContextData{}))
	})

	return unwrapPinnedCodecContextData(ctx.Opaque)
}

func (ctx *codecContext) finalizedPinnedData() {
	if ctx.Opaque == nil {
		return
	}

	cgo.Handle(ctx.Opaque).Delete()
}

func (ctx *codecContext) Codec() *Codec {
	return &Codec{
		_codec: ctx._codecContext.Codec,
	}
}

func (ctx *codecContext) CodecID() avcodec.ID {
	return avcodec.ID(ctx._codecContext.CodecID)
}

func (ctx *codecContext) CodecParameters() *CodecParameters {
	var parameters avcodec.Parameters
	if err := averror(avcodec.ParametersFromContext(&parameters, ctx._codecContext)); err != nil {
		panic(err)
	}

	return &CodecParameters{
		_codecParameters: &parameters,
	}
}

func (ctx *codecContext) SetGetFormat(f func([]avutil.PixelFormat) avutil.PixelFormat) {
	if f == nil {
		// ctx.GetFormat = (*[0]byte)(C.avcodec_default_get_format)
		ctx.GetFormat = nil
	} else {
		ctx.pinnedData().getFormatFunc = f
		ctx.GetFormat = (*[0]byte)(unsafe.Pointer(C.goavCodecContextGetFormat))
	}
}

func (ctx *codecContext) HwDeviceCtx() *HWDeviceContext {
	if ctx._codecContext.HwDeviceCtx == nil {
		return nil
	}

	return newHWDeviceContext(avutil.RefBuffer(ctx._codecContext.HwDeviceCtx))
}

func (ctx *codecContext) SetHwDeviceCtx(deviceCtx *HWDeviceContext) {
	if ctx._codecContext.HwDeviceCtx != nil {
		avutil.UnrefBuffer(&ctx._codecContext.HwDeviceCtx)
	}

	if deviceCtx == nil {
		ctx._codecContext.HwDeviceCtx = nil
	} else {
		ctx._codecContext.HwDeviceCtx = deviceCtx.ref()
	}
}

func (ctx *codecContext) HwFramesCtx() *HWFramesContext {
	if ctx._codecContext.HwFramesCtx == nil {
		return nil
	}

	return newHWFramesContext(avutil.RefBuffer(ctx._codecContext.HwFramesCtx))
}

func (ctx *codecContext) SetHwFramesCtx(framesCtx *HWFramesContext) {
	if ctx._codecContext.HwFramesCtx != nil {
		avutil.UnrefBuffer(&ctx._codecContext.HwFramesCtx)
	}

	if framesCtx == nil {
		ctx._codecContext.HwFramesCtx = nil
	} else {
		ctx._codecContext.HwFramesCtx = framesCtx.ref()
	}
}

func (ctx *codecContext) SetOption(name string, value interface{}) error {
	return setOption(ctx._codecContext.PrivData, name, value, 0)
}

func (ctx *codecContext) GetOption(name string) (interface{}, error) {
	return getOption(ctx._codecContext.PrivData, name, 0)
}

func (ctx *codecContext) init() error {
	ctx.initOnce.Do(func() {
		ctx.initErr = averror(avcodec.Open(ctx._codecContext, nil, nil))
	})

	return ctx.initErr
}

func (ctx *codecContext) Open() error {
	return ctx.init()
}
