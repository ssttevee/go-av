package av

// #include <libavcodec/avcodec.h>
import "C"
import (
	"fmt"
	"runtime"

	"github.com/ssttevee/go-av/avcodec"
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
		avcodec.FreeContext(&ctx._codecContext)
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
	return averror(avcodec.ReceiveFrame(ctx._codecContext, frame.prepare()))
}

func (ctx *DecoderContext) ReceiveFrame() (*Frame, error) {
	frame := NewFrame()
	if err := averror(avcodec.ReceiveFrame(ctx._codecContext, frame._frame)); err != nil {
		return nil, err
	}

	return frame, nil
}
