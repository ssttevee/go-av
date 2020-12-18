package av

// #cgo pkg-config: libavformat libavcodec libavfilter
// #include <libavcodec/avcodec.h>
// #include <libavutil/opt.h>
//
// extern enum AVPixelFormat goavCodecContextGetFormat(AVCodecContext *, enum AVPixelFormat *);
import "C"
import (
	"reflect"
	"sync"
	"unsafe"
)

type CodecNotFoundError string

func (e CodecNotFoundError) Error() string {
	return "codec not found: " + string(e)
}

type CodecID C.enum_AVCodecID

const (
	H264 = CodecID(C.AV_CODEC_ID_H264)
)

func (id CodecID) String() string {
	return C.GoString(C.avcodec_get_name(id.ctype()))
}

func (id CodecID) ctype() C.enum_AVCodecID {
	return C.enum_AVCodecID(id)
}

type Codec struct {
	codec *C.AVCodec
}

func FindDecoderCodecByName(name string) (*Codec, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	codec := C.avcodec_find_decoder_by_name(cname)
	if codec == nil {
		return nil, CodecNotFoundError(name)
	}

	return &Codec{
		codec: codec,
	}, nil
}

func FindEncoderCodecByName(name string) (*Codec, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	codec := C.avcodec_find_encoder_by_name(cname)
	if codec == nil {
		return nil, CodecNotFoundError(name)
	}

	return &Codec{
		codec: codec,
	}, nil
}

func (c *Codec) Name() string {
	return C.GoString(c.codec.name)
}

func countPixelFormats(fmts *C.enum_AVPixelFormat) int {
	if fmts == nil {
		return 0
	}

	size := unsafe.Sizeof(*fmts)

	var n int
	ptr := uintptr(unsafe.Pointer(fmts))
	for *(*PixelFormat)(unsafe.Pointer(ptr)) != -1 {
		n++
		ptr += size
	}

	return n
}

func (c *Codec) NumPixelFormat() int {
	return countPixelFormats(c.codec.pix_fmts)
}

func (c *Codec) PixelFormat(i int) PixelFormat {
	return *(*PixelFormat)(unsafe.Pointer(uintptr(unsafe.Pointer(c.codec.pix_fmts)) + unsafe.Sizeof(*c.codec.pix_fmts)*uintptr(i)))
}

func (c *Codec) NumSampleFormat() int {
	if c.codec.sample_fmts == nil {
		return 0
	}

	size := unsafe.Sizeof(*c.codec.sample_fmts)

	var n int
	ptr := uintptr(unsafe.Pointer(c.codec.sample_fmts))
	for *(*PixelFormat)(unsafe.Pointer(ptr)) != -1 {
		n++
		ptr += size
	}

	return n
}

func (c *Codec) SampleFormat(i int) SampleFormat {
	return *(*SampleFormat)(unsafe.Pointer(uintptr(unsafe.Pointer(c.codec.sample_fmts)) + unsafe.Sizeof(*c.codec.sample_fmts)*uintptr(i)))
}

type CodecParameters struct {
	parameters *C.AVCodecParameters
}

type pinnedCodecContextData struct {
	err error

	getFormatFunc func([]PixelFormat) PixelFormat
}

var pinnedCodecContextDataEntries = map[pinType]*pinnedCodecContextData{}

func returnPinnedCodecContextDataError(p unsafe.Pointer, err error) C.int {
	pinnedCodecContextDataEntries[pin(p)].err = err
	return C.int(errInternalFormatError)
}

//export goavCodecContextGetFormat
func goavCodecContextGetFormat(ctx *C.AVCodecContext, choices *C.enum_AVPixelFormat) C.enum_AVPixelFormat {
	n := countPixelFormats(choices)
	return pinnedCodecContextDataEntries[pin(ctx.opaque)].getFormatFunc(*(*[]PixelFormat)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(choices)),
		Len:  n,
		Cap:  n,
	}))).ctype()
}

type codecContext struct {
	ctx *C.AVCodecContext

	pinnedDataOnce sync.Once

	initOnce sync.Once
	initErr  error
}

func newCodecContext(codec *Codec, params *CodecParameters) (*C.AVCodecContext, error) {
	ctx := C.avcodec_alloc_context3(codec.codec)
	if ctx == nil {
		panic(ErrNoMem)
	}

	if params != nil {
		if err := averror(C.avcodec_parameters_to_context(ctx, params.parameters)); err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

func (ctx *codecContext) pinnedData() *pinnedCodecContextData {
	ctx.pinnedDataOnce.Do(func() {
		var p pinType
		for {
			p = randPin()
			if _, ok := pinnedCodecContextDataEntries[p]; !ok {
				break
			}
		}

		pinnedCodecContextDataEntries[p] = &pinnedCodecContextData{}

		ctx.ctx.opaque = pinptr(p)
	})

	return pinnedCodecContextDataEntries[pin(ctx.ctx.opaque)]
}

func (ctx *codecContext) finalizedPinnedData() {
	if ctx.ctx.opaque == nil {
		return
	}

	delete(pinnedCodecContextDataEntries, pin(ctx.ctx.opaque))
}

func (ctx *codecContext) Codec() *Codec {
	return &Codec{codec: ctx.ctx.codec}
}

func (ctx *codecContext) CodecID() CodecID {
	return CodecID(ctx.ctx.codec_id)
}

func (ctx *codecContext) CodecParameters() *CodecParameters {
	var parameters C.AVCodecParameters
	if err := averror(C.avcodec_parameters_from_context(&parameters, ctx.ctx)); err != nil {
		panic(err)
	}

	return &CodecParameters{
		parameters: &parameters,
	}
}

func (ctx *codecContext) SetGetFormat(f func([]PixelFormat) PixelFormat) {
	if f == nil {
		ctx.ctx.get_format = (*[0]byte)(C.avcodec_default_get_format)
	} else {
		ctx.pinnedData().getFormatFunc = f
		ctx.ctx.get_format = (*[0]byte)(C.goavCodecContextGetFormat)
	}
}

func (ctx *codecContext) HWDeviceContext() *HWDeviceContext {
	if ctx.ctx.hw_device_ctx == nil {
		return nil
	}

	return newHWDeviceContext(C.av_buffer_ref(ctx.ctx.hw_device_ctx))
}

func (ctx *codecContext) SetHWDeviceContext(deviceCtx *HWDeviceContext) {
	if ctx.ctx.hw_device_ctx != nil {
		C.av_buffer_unref(&ctx.ctx.hw_device_ctx)
	}

	if deviceCtx == nil {
		ctx.ctx.hw_device_ctx = nil
	} else {
		ctx.ctx.hw_device_ctx = deviceCtx.ref()
	}
}

func (ctx *codecContext) HWFramesContext() *HWFramesContext {
	if ctx.ctx.hw_frames_ctx == nil {
		return nil
	}

	return newHWFramesContext(C.av_buffer_ref(ctx.ctx.hw_frames_ctx))
}

func (ctx *codecContext) SetHWFramesContext(framesCtx *HWFramesContext) {
	if ctx.ctx.hw_frames_ctx != nil {
		C.av_buffer_unref(&ctx.ctx.hw_frames_ctx)
	}

	if framesCtx == nil {
		ctx.ctx.hw_frames_ctx = nil
	} else {
		ctx.ctx.hw_frames_ctx = framesCtx.ref()
	}
}

func (ctx *codecContext) Framerate() Rational {
	return rational(ctx.ctx.framerate)
}

func (ctx *codecContext) SetFramerate(framerate Rational) {
	ctx.ctx.framerate = framerate.ctype()
}

func (ctx *codecContext) PacketTimeBase() Rational {
	return rational(ctx.ctx.pkt_timebase)
}

func (ctx *codecContext) SetPacketTimeBase(timeBase Rational) {
	ctx.ctx.pkt_timebase = timeBase.ctype()
}

func (ctx *codecContext) TimeBase() Rational {
	return rational(ctx.ctx.time_base)
}

func (ctx *codecContext) SetTimeBase(timeBase Rational) {
	ctx.ctx.time_base = timeBase.ctype()
}

func (ctx *codecContext) PixelFormat() PixelFormat {
	return PixelFormat(ctx.ctx.pix_fmt)
}

func (ctx *codecContext) SetPixelFormat(format PixelFormat) {
	ctx.ctx.pix_fmt = format.ctype()
}

func (ctx *codecContext) SampleAspectRatio() Rational {
	return rational(ctx.ctx.sample_aspect_ratio)
}

func (ctx *codecContext) SetSampleAspectRatio(ratio Rational) {
	ctx.ctx.sample_aspect_ratio = ratio.ctype()
}

func (ctx *codecContext) ExtraDataSize() int {
	return int(ctx.ctx.extradata_size)
}

func (ctx *codecContext) SetExtraDataSize(size int) {
	ctx.ctx.extradata_size = C.int(size)
}

func (ctx *codecContext) ExtraData() int {
	return int(ctx.ctx.extradata_size)
}

func (ctx *codecContext) SetExtraData(size int) {
	ctx.ctx.extradata_size = C.int(size)
}

func (ctx *codecContext) Width() int {
	return int(ctx.ctx.width)
}

func (ctx *codecContext) SetWidth(width int) {
	ctx.ctx.width = C.int(width)
}

func (ctx *codecContext) Height() int {
	return int(ctx.ctx.height)
}

func (ctx *codecContext) SetHeight(height int) {
	ctx.ctx.height = C.int(height)
}

func (ctx *codecContext) Bitrate() int64 {
	return int64(ctx.ctx.bit_rate)
}

func (ctx *codecContext) SetBitrate(bitrate int64) {
	ctx.ctx.bit_rate = C.int64_t(bitrate)
}

func (ctx *codecContext) BufferSize() int {
	return int(ctx.ctx.rc_buffer_size)
}

func (ctx *codecContext) SetBufferSize(size int) {
	ctx.ctx.rc_buffer_size = C.int(size)
}

func (ctx *codecContext) MaxBitrate() int64 {
	return int64(ctx.ctx.rc_max_rate)
}

func (ctx *codecContext) SetMaxBitrate(bitrate int64) {
	ctx.ctx.rc_max_rate = C.int64_t(bitrate)
}

func (ctx *codecContext) MinBitrate() int64 {
	return int64(ctx.ctx.rc_min_rate)
}

func (ctx *codecContext) SetMinBitrate(bitrate int64) {
	ctx.ctx.rc_min_rate = C.int64_t(bitrate)
}

func (ctx *codecContext) Channels() int {
	return int(ctx.ctx.channels)
}

func (ctx *codecContext) SetChannels(channels int) {
	ctx.ctx.channels = C.int(channels)
}

func (ctx *codecContext) ChannelLayout() uint64 {
	return uint64(ctx.ctx.channel_layout)
}

func (ctx *codecContext) SetChannelLayout(layout uint64) {
	ctx.ctx.channel_layout = C.uint64_t(layout)
}

func (ctx *codecContext) SampleRate() int {
	return int(ctx.ctx.sample_rate)
}

func (ctx *codecContext) SetSampleRate(rate int) {
	ctx.ctx.sample_rate = C.int(rate)
}

func (ctx *codecContext) SampleFormat() SampleFormat {
	return SampleFormat(ctx.ctx.sample_fmt)
}

func (ctx *codecContext) SetSampleFormat(fmt SampleFormat) {
	ctx.ctx.sample_fmt = fmt.ctype()
}

func (ctx *codecContext) SetOption(name string, value interface{}) error {
	return setOption(ctx.ctx.priv_data, name, value, 0)
}

func (ctx *codecContext) GetOption(name string) (interface{}, error) {
	return getOption(ctx.ctx.priv_data, name, 0)
}

func (ctx *codecContext) init() error {
	ctx.initOnce.Do(func() {
		ctx.initErr = averror(C.avcodec_open2(ctx.ctx, nil, nil))
	})

	return ctx.initErr
}

func (ctx *codecContext) Open() error {
	return ctx.init()
}
