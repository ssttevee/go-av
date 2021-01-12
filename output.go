package av

// #include <libavformat/avformat.h>
import "C"
import (
	"errors"
	"io"
	"runtime"
	"sync"

	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
	"github.com/ssttevee/go-av/internal/common"
)

type outputDest interface {
	initIOContext(pb **avformat.IOContext) (func() error, error)
}

type writerOutputDest struct {
	w io.Writer
}

func (dst writerOutputDest) initIOContext(pb **avformat.IOContext) (func() error, error) {
	*pb = allocAvioContext(dst.w, true)
	return nil, nil
}

type fileOutputDest string

func (dst fileOutputDest) initIOContext(pb **avformat.IOContext) (func() error, error) {
	var ctx *avformat.IOContext
	if err := averror(avformat.OpenIO(&ctx, string(dst), C.AVIO_FLAG_WRITE)); err != nil {
		return nil, err
	}

	*pb = ctx

	return func() error {
		return averror(avformat.CloseIO(ctx))
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
	var ctx *avformat.Context
	if err := averror(avformat.NewOutputContext(&ctx, nil, formatName, filename)); err != nil {
		return nil, err
	}

	ctx.Opaque = nil

	ret := &OutputFormatContext{
		formatContext: formatContext{
			_formatContext: ctx,
		},
	}

	runtime.SetFinalizer(ret, func(ctx *OutputFormatContext) {
		ctx.finalizePinnedData()
		avformat.FreeContext(ctx._formatContext)
	})

	return ret, nil
}

func NewOutputContext(formatName string) (*OutputFormatContext, error) {
	return newOutputContext(formatName, "")
}

func (ctx *OutputFormatContext) NewStream(codec *Codec) *Stream {
	var c *avcodec.Codec
	if codec != nil {
		c = codec._codec
	}

	stream := avformat.NewStream(ctx._formatContext, c)
	if stream == nil {
		panic(avutil.ErrNoMem)
	}

	stream.ID = int32(ctx._formatContext.NbStreams - 1)

	return &Stream{
		_stream:   stream,
		formatCtx: ctx._formatContext,
	}
}

func (ctx *OutputFormatContext) init() error {
	ctx.initOnce.Do(func() {
		if ctx.Flags&C.AVFMT_NOFILE == 0 && ctx.dst != nil {
			if ctx.dst == nil {
				ctx.initErr = errors.New("missing output dest")
				return
			}

			ctx.closeFunc, ctx.initErr = ctx.dst.initIOContext(&ctx.Pb)
			if ctx.initErr != nil {
				return
			}
		}

		if ctx.initErr = averror(avformat.WriteHeader(ctx._formatContext, nil)); ctx.initErr != nil {
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

	var pkt *avcodec.Packet
	if packet != nil {
		pkt = packet._packet
	}

	return ctx.realError(averror(avformat.WriteInterleavedFrame(ctx._formatContext, pkt)))
}

func (ctx *OutputFormatContext) realError(err error) error {
	if averr, ok := errors.Unwrap(err).(avutil.Error); ok {
		switch int(averr) {
		case common.IOError:
			return pinnedFiles[pin(ctx.Pb.Opaque)].err

		case common.FormatError:
			return ctx.pinnedData().err
		}
	}

	return err
}

func (ctx *OutputFormatContext) Close() error {
	ctx.closeOnce.Do(func() {
		if ctx.closeErr = averror(avformat.WriteTrailer(ctx._formatContext)); ctx.closeErr != nil {
			return
		}

		if ctx.Flags&C.AVFMT_NOFILE == 0 {
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
