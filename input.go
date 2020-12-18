package av

// #include <libavformat/avformat.h>
import "C"
import (
	"io"
	"runtime"
	"unsafe"
)

type InputFormatContext struct {
	formatContext

	ioctx *ioContext
}

func finalizeInputFormatContext(ctx *InputFormatContext) {
	ctx.finalizedPinnedData()
	C.avformat_close_input(&ctx.ctx)
}

func OpenInputFile(input string) (*InputFormatContext, error) {
	cinput := C.CString(input)
	defer C.free(unsafe.Pointer(cinput))

	var ctx *C.AVFormatContext
	if err := averror(C.avformat_open_input(&ctx, cinput, nil, nil)); err != nil {
		return nil, err
	}

	if err := averror(C.avformat_find_stream_info(ctx, nil)); err != nil {
		return nil, err
	}

	ret := &InputFormatContext{
		formatContext: formatContext{
			ctx: ctx,
		},
	}

	runtime.SetFinalizer(ret, finalizeInputFormatContext)

	return ret, nil
}

func OpenInputReader(r io.Reader) (*InputFormatContext, error) {
	ioctx := newIOContext(r, false)

	ctx := C.avformat_alloc_context()
	if ctx == nil {
		panic(ErrNoMem)
	}

	ctx.opaque = nil
	ctx.pb = ioctx.ctx

	ret := &InputFormatContext{
		formatContext: formatContext{
			ctx: ctx,
		},
		ioctx: ioctx,
	}

	runtime.SetFinalizer(ret, finalizeInputFormatContext)

	if err := averror(C.avformat_open_input(&ctx, nil, nil, nil)); err != nil {
		return nil, err
	}

	if err := averror(C.avformat_find_stream_info(ctx, nil)); err != nil {
		return nil, err
	}

	return ret, nil
}

func (ctx *InputFormatContext) ReadPacketReuse(packet *Packet) error {
	return averror(C.av_read_frame(ctx.ctx, packet.prepare()))
}

func (ctx *InputFormatContext) ReadPacket() (*Packet, error) {
	packet := NewPacket()
	if err :=  averror(C.av_read_frame(ctx.ctx, packet.packet)); err != nil {
		return nil, err
	}

	return packet, nil
}
