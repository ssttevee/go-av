package av

// #include <libavcodec/avcodec.h>
import "C"
import (
	"runtime"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avutil"
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
			_codecContext: ctx,
		},
	}

	runtime.SetFinalizer(ret, func(ctx *EncoderContext) {
		ctx.finalizedPinnedData()
		avcodec.FreeContext(&ctx._codecContext)
	})

	return ret, nil
}

func (ctx *EncoderContext) SendFrame(frame *Frame) error {
	if err := ctx.init(); err != nil {
		return err
	}

	defer runtime.KeepAlive(frame)

	return averror(avcodec.SendFrame(ctx._codecContext, frame._frame))
}

func (ctx *EncoderContext) ReceivePacketReuse(packet *Packet) error {
	return averror(avcodec.ReceivePacket(ctx._codecContext, packet.prepare()))
}

func (ctx *EncoderContext) ReceivePacket() (*Packet, error) {
	packet := NewPacket()
	if err := averror(avcodec.ReceivePacket(ctx._codecContext, packet._packet)); err != nil {
		return nil, err
	}

	return packet, nil
}

func (ctx *EncoderContext) FramePackets(frame *Frame) ([]*Packet, error) {
	if ctx._codecContext.HwFramesCtx != nil && frame._frame.HwFramesCtx == nil {
		hwFrame := NewFrame()
		defer hwFrame.Unref()

		if err := averror(avutil.GetHWFrameBuffer(ctx._codecContext.HwFramesCtx, hwFrame._frame, 0)); err != nil {
			return nil, err
		}

		if err := averror(avutil.TransferHWFrameData(hwFrame._frame, frame._frame, 0)); err != nil {
			return nil, err
		}

		if err := averror(avutil.CopyFrameProps(hwFrame._frame, frame._frame)); err != nil {
			return nil, err
		}

		frame = hwFrame
	} else if ctx._codecContext.HwFramesCtx == nil && frame._frame.HwFramesCtx != nil {
		swFrame := NewFrame()
		defer swFrame.Unref()

		if err := averror(avutil.TransferHWFrameData(swFrame._frame, frame._frame, 0)); err != nil {
			return nil, err
		}

		if err := averror(avutil.CopyFrameProps(swFrame._frame, frame._frame)); err != nil {
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
		if errors.Is(err, avutil.ErrAgain) {
			// ErrAgain means there are no more packets to receive
			break
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		packets = append(packets, packet)
	}

	return packets, nil
}
