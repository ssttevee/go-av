package av

// #include <libavformat/avformat.h>
import "C"
import (
	"io"
	"runtime"

	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
)

type InputFormatContext struct {
	formatContext

	ioctx *ioContext
}

func finalizeInputFormatContext(ctx *InputFormatContext) {
	ctx.finalizePinnedData()
	// heap pointer may not be passed to cgo, so use a stack pointer instead :D
	formatCtx := (*avformat.Context)(ctx._formatContext)
	avformat.CloseInput(&formatCtx)
	ctx._formatContext = formatCtx
}

func OpenInputFile(input string) (*InputFormatContext, error) {
	var ctx *avformat.Context
	if err := averror(avformat.OpenInput(&ctx, input, nil, nil)); err != nil {
		return nil, err
	}

	if err := averror(avformat.FindStreamInfo(ctx, nil)); err != nil {
		return nil, err
	}

	ret := &InputFormatContext{
		formatContext: formatContext{
			_formatContext: ctx,
		},
	}

	runtime.SetFinalizer(ret, finalizeInputFormatContext)

	return ret, nil
}

func OpenInputReader(r io.Reader) (*InputFormatContext, error) {
	ioctx := newIOContext(r, false)

	ctx := avformat.NewContext()
	if ctx == nil {
		panic(avutil.ErrNoMem)
	}

	ctx.Opaque = nil
	ctx.Pb = ioctx._ioContext

	ret := &InputFormatContext{
		formatContext: formatContext{
			_formatContext: ctx,
		},
		ioctx: ioctx,
	}

	if err := averror(avformat.OpenInput(&ctx, "", nil, nil)); err != nil {
		return nil, err
	}

	runtime.SetFinalizer(ret, finalizeInputFormatContext)

	if err := averror(avformat.FindStreamInfo(ctx, nil)); err != nil {
		return nil, err
	}

	return ret, nil
}

func (ctx *InputFormatContext) ReadPacketReuse(packet *Packet) error {
	return averror(avformat.ReadFrame(ctx._formatContext, packet.prepare()))
}

func (ctx *InputFormatContext) ReadPacket() (*Packet, error) {
	packet := NewPacket()
	if err := averror(avformat.ReadFrame(ctx._formatContext, packet._packet)); err != nil {
		return nil, err
	}

	return packet, nil
}

func (ctx *InputFormatContext) SeekFile(streamIndex int32, minTimestamp, timestamp, maxTimestamp int64, flags int32) error {
	return averror(avformat.SeekFile(ctx._formatContext, streamIndex, minTimestamp, timestamp, maxTimestamp, flags))
}
