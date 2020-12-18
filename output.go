package av

// #include <libavformat/avformat.h>
import "C"
import (
	"errors"
	"io"
	"runtime"
	"sync"
	"unsafe"
)

type outputDest interface {
	initIOContext(pb **C.AVIOContext) (func() error, error)
}

type writerOutputDest struct {
	w io.Writer
}

func (dst writerOutputDest) initIOContext(pb **C.AVIOContext) (func() error, error) {
	*pb = allocAvioContext(dst.w, true)
	return nil, nil
}

type fileOutputDest string

func (dst fileOutputDest) initIOContext(pb **C.AVIOContext) (func() error, error) {
	filename := C.CString(string(dst))
	defer C.free(unsafe.Pointer(filename))

	var ctx *C.AVIOContext
	if err := averror(C.avio_open(&ctx, filename, C.AVIO_FLAG_WRITE)); err != nil {
		return nil, err
	}

	*pb = ctx

	return func() error {
		return averror(C.avio_close(ctx))
	}, nil
}

type OutputFormatContext struct {
	formatContext

	dst outputDest

	initOnce sync.Once
	initErr  error

	closeFunc func() error
	closeOnce sync.Once
	closeErr  error
}

func NewFileOutputContext(formatName string, filename string) (*OutputFormatContext, error) {
	ctx, err := newOutputContext(formatName, filename)
	if err != nil {
		return nil, err
	}

	ctx.dst = fileOutputDest(filename)

	return ctx, nil
}

func NewWriterOutputContext(formatName string, w io.Writer) (*OutputFormatContext, error) {
	ctx, err := newOutputContext(formatName, "")
	if err != nil {
		return nil, err
	}

	ctx.dst = writerOutputDest{w: w}

	return ctx, nil
}

func newOutputContext(formatName string, filename string) (*OutputFormatContext, error) {
	cformatName := C.CString(formatName)
	defer C.free(unsafe.Pointer(cformatName))

	var cfilename *C.char
	if filename != "" {
		cfilename = C.CString(filename)
		defer C.free(unsafe.Pointer(cfilename))
	}

	var ctx *C.AVFormatContext
	if err := averror(C.avformat_alloc_output_context2(&ctx, nil, cformatName, cfilename)); err != nil {
		return nil, err
	}

	ctx.opaque = nil

	ret := &OutputFormatContext{
		formatContext: formatContext{
			ctx: ctx,
		},
	}

	runtime.SetFinalizer(ret, func(ctx *OutputFormatContext) {
		ctx.finalizedPinnedData()
		C.avformat_free_context(ctx.ctx)
	})

	return ret, nil
}

func NewOutputContext(formatName string) (*OutputFormatContext, error) {
	return newOutputContext(formatName, "")
}

func (ctx *OutputFormatContext) NewStream(codec *Codec) *Stream {
	var c *C.AVCodec
	if codec != nil {
		c = codec.codec
	}

	stream := C.avformat_new_stream(ctx.ctx, c)
	if stream == nil {
		panic(ErrNoMem)
	}

	stream.id = C.int(ctx.ctx.nb_streams) - 1

	return &Stream{
		stream:    stream,
		formatCtx: ctx.ctx,
	}
}

func (ctx *OutputFormatContext) init() error {
	ctx.initOnce.Do(func() {
		if ctx.ctx.flags&C.AVFMT_NOFILE == 0 && ctx.dst != nil {
			if ctx.dst == nil {
				ctx.initErr = errors.New("missing output dest")
				return
			}

			ctx.closeFunc, ctx.initErr = ctx.dst.initIOContext(&ctx.ctx.pb)
			if ctx.initErr != nil {
				return
			}
		}

		if ctx.initErr = averror(C.avformat_write_header(ctx.ctx, nil)); ctx.initErr != nil {
			return
		}
	})

	return ctx.initErr
}

func (ctx *OutputFormatContext) WritePacket(packet *Packet) error {
	if err := ctx.realError(ctx.init()); err != nil {
		return err
	}

	defer runtime.KeepAlive(packet)

	var pkt *C.AVPacket
	if packet != nil {
		pkt = packet.packet
	}

	return ctx.realError(averror(C.av_interleaved_write_frame(ctx.ctx, pkt)))
}

func (ctx *OutputFormatContext) realError(err error) error {
	switch err {
	case errInternalIOError:
		return pinnedFiles[pin(ctx.ctx.pb.opaque)].err

	case errInternalFormatError:
		return ctx.pinnedData().err
	}

	return err
}

func (ctx *OutputFormatContext) Close() error {
	ctx.closeOnce.Do(func() {
		if ctx.closeErr = averror(C.av_write_trailer(ctx.ctx)); ctx.closeErr != nil {
			return
		}

		if ctx.ctx.flags&C.AVFMT_NOFILE == 0 {
			if ctx.closeFunc == nil {
				return
			}

			if ctx.closeErr = ctx.closeFunc(); ctx.closeErr != nil {
				return
			}
		}
	})

	return ctx.closeErr
}
