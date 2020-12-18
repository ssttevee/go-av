package av

// #include <libavcodec/avcodec.h>
import "C"
import (
	"fmt"
	"runtime"
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
			ctx: ctx,
		},
	}

	runtime.SetFinalizer(ret, func(ctx *DecoderContext) {
		ctx.finalizedPinnedData()
		C.avcodec_free_context(&ctx.ctx)
	})

	return ret, nil
}

func (ctx *DecoderContext) BufferSourceArgs() string {
	var framerateArg string
	if framerate := ctx.Framerate(); !framerate.IsZero() {
		framerateArg = ":frame_rate=" + framerate.String()
	}

	return fmt.Sprintf("video_size=%dx%d:pix_fmt=%d:time_base=%s:pixel_aspect=%s", ctx.Width(), ctx.Height(), int(ctx.PixelFormat()), ctx.TimeBase(), ctx.SampleAspectRatio()) + framerateArg
}

func (ctx *DecoderContext) SendPacket(packet *Packet) error {
	if err := ctx.init(); err != nil {
		return err
	}

	return averror(C.avcodec_send_packet(ctx.ctx, packet.packet))
}

func (ctx *DecoderContext) ReceiveFrameReuse(frame *Frame) error {
	return averror(C.avcodec_receive_frame(ctx.ctx, frame.prepare()))
}

func (ctx *DecoderContext) ReceiveFrame() (*Frame, error) {
	frame := NewFrame()
	if err := averror(C.avcodec_receive_frame(ctx.ctx, frame.frame)); err != nil {
		return nil, err
	}

	return frame, nil
}
