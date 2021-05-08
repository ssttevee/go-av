package av

// #include <libavcodec/avcodec.h>
import "C"
import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avutil"
)

type DecoderContext struct {
	codecContext
}

func NewDecoderContext(codec *Codec, params *CodecParameters) (*DecoderContext, error) {
	ctx, err := newCodecContext(codec, params)
	if err != nil {
		return nil, err
	}

	ret := &DecoderContext{
		codecContext: codecContext{
			_codecContext: ctx,
		},
	}

	runtime.SetFinalizer(ret, func(ctx *DecoderContext) {
		ctx.finalizedPinnedData()
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		codecContext := (*avcodec.Context)(ctx._codecContext)
		avcodec.FreeContext(&codecContext)
		ctx._codecContext = codecContext
	})

	return ret, nil
}

func (ctx *DecoderContext) BufferSourceArgs() string {
	var framerateArg string
	if !ctx.Framerate.IsZero() {
		framerateArg = ":frame_rate=" + ctx.Framerate.String()
	}

	return fmt.Sprintf("video_size=%dx%d:pix_fmt=%d:time_base=%s:pixel_aspect=%s", ctx.Width, ctx.Height, int(ctx.PixFmt), ctx.TimeBase, ctx.SampleAspectRatio) + framerateArg
}

func (ctx *DecoderContext) SendPacket(packet *Packet) error {
	if err := ctx.init(); err != nil {
		return err
	}

	return averror(avcodec.SendPacket(ctx._codecContext, packet._packet))
}

func (ctx *DecoderContext) ReceiveFrameReuse(frame *Frame) error {
	if err := ctx.init(); err != nil {
		return err
	}

	return averror(avcodec.ReceiveFrame(ctx._codecContext, frame.prepare()))
}

func (ctx *DecoderContext) ReceiveFrame() (*Frame, error) {
	if err := ctx.init(); err != nil {
		return nil, err
	}

	frame := NewFrame()
	if err := averror(avcodec.ReceiveFrame(ctx._codecContext, frame._frame)); err != nil {
		return nil, err
	}

	return frame, nil
}

type FrameIterator struct {
	ifc         *InputFormatContext
	dc          *DecoderContext
	streamIndex int32

	pkt *Packet
}

func (ctx *DecoderContext) NewFrameIterator(ifc *InputFormatContext, streamIndex int32) *FrameIterator {
	return &FrameIterator{
		ifc:         ifc,
		dc:          ctx,
		streamIndex: streamIndex,
		pkt:         NewPacket(),
	}
}

func NewFrameIterator(ifc *InputFormatContext, streamIndex int32) (*FrameIterator, error) {
	stream := ifc.Streams()[streamIndex]
	if stream == nil {
		return nil, errors.Errorf("stream index %d not found in input", streamIndex)
	}

	codec, err := FindDecoderCodecByID(stream._stream.Codecpar.CodecID)
	if err != nil {
		return nil, err
	}

	dc, err := NewDecoderContext(codec, stream.Codecpar())
	if err != nil {
		return nil, err
	}

	return &FrameIterator{
		ifc: ifc,
		dc:  dc,
		pkt: NewPacket(),
	}, nil
}

func (it *FrameIterator) Next(frame *Frame) error {
	for {
		if err := it.dc.ReceiveFrameReuse(frame); err == nil {
			return nil
		} else if !errors.Is(err, avutil.ErrAgain) {
			return err
		}

		for {
			if err := it.ifc.ReadPacketReuse(it.pkt); err != nil {
				return err
			} else if it.pkt.StreamIndex == it.streamIndex {
				break
			}
		}

		if err := it.dc.SendPacket(it.pkt); err != nil {
			return err
		}
	}
}
