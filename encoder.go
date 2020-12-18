package av

// #include <libavcodec/avcodec.h>
import "C"
import (
	"runtime"

	"github.com/pkg/errors"
)

type EncoderContext struct {
	codecContext
}

func NewEncoderContext(codec *Codec, params *CodecParameters) (*EncoderContext, error) {
	ctx, err := newCodecContext(codec, params)
	if err != nil {
		return nil, err
	}

	ret := &EncoderContext{
		codecContext: codecContext{
			ctx: ctx,
		},
	}

	runtime.SetFinalizer(ret, func(ctx *EncoderContext) {
		ctx.finalizedPinnedData()
		C.avcodec_free_context(&ctx.ctx)
	})

	return ret, nil
}

func (ctx *EncoderContext) SendFrame(frame *Frame) error {
	if err := ctx.init(); err != nil {
		return err
	}

	defer runtime.KeepAlive(frame)

	return averror(C.avcodec_send_frame(ctx.ctx, frame.frame))
}

func (ctx *EncoderContext) ReceivePacketReuse(packet *Packet) error {
	return averror(C.avcodec_receive_packet(ctx.ctx, packet.prepare()))
}

func (ctx *EncoderContext) ReceivePacket() (*Packet, error) {
	packet := NewPacket()
	if err := averror(C.avcodec_receive_packet(ctx.ctx, packet.packet)); err != nil {
		return nil, err
	}

	return packet, nil
}

func (ctx *EncoderContext) FramePackets(frame *Frame) ([]*Packet, error) {
	if hwFrameCtx := ctx.HWFramesContext(); hwFrameCtx != nil && frame.frame.hw_frames_ctx == nil {
		hwFrame := NewFrame()
		defer hwFrame.Unref()

		if err := averror(C.av_hwframe_get_buffer(hwFrameCtx.ctx, hwFrame.frame, 0)); err != nil {
			return nil, err
		}

		if err := averror(C.av_hwframe_transfer_data(hwFrame.frame, frame.frame, 0)); err != nil {
			return nil, err
		}

		if err := averror(C.av_frame_copy_props(hwFrame.frame, frame.frame)); err != nil {
			return nil, err
		}

		frame = hwFrame
	} else if hwFrameCtx == nil && frame.frame.hw_frames_ctx != nil {
		swFrame := NewFrame()
		defer swFrame.Unref()

		if err := averror(C.av_hwframe_transfer_data(swFrame.frame, frame.frame, 0)); err != nil {
			return nil, err
		}

		if err := averror(C.av_frame_copy_props(swFrame.frame, frame.frame)); err != nil {
			return nil, err
		}

		frame = swFrame
	}

	if err := ctx.SendFrame(frame); err != nil {
		return nil, errors.WithStack(err)
	}

	var packets []*Packet
	for {
		packet, err := ctx.ReceivePacket()
		if errors.Is(err, ErrAgain) {
			// ErrAgain means there are no more packets to receive
			break
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		packets = append(packets, packet)
	}

	return packets, nil
}
