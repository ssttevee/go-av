package av

import (
	"errors"
	"runtime"
	"sync"

	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avutil"
)

type BitstreamFilterNotFoundError string

func (e BitstreamFilterNotFoundError) Error() string {
	return "bitstream not found: " + string(e)
}

type BitstreamFilter struct {
	filter *avcodec.BitstreamFilter
}

func FindBitstreamFilterByName(name string) (*BitstreamFilter, error) {
	filter := avcodec.GetBitstreamFilterByName(name)
	if filter == nil {
		return nil, BitstreamFilterNotFoundError(name)
	}

	return &BitstreamFilter{filter: filter}, nil
}

type BitstreamFilterContext struct {
	ctx *avcodec.BitstreamFilterContext

	initOnce sync.Once
	initErr  error
}

func NewBitstreamFilterContext(filter *BitstreamFilter) (*BitstreamFilterContext, error) {
	var ctx *avcodec.BitstreamFilterContext
	if err := averror(avcodec.NewBitstreamFilter(filter.filter, &ctx)); err != nil {
		return nil, err
	}

	ret := &BitstreamFilterContext{ctx: ctx}

	runtime.SetFinalizer(ret, func(ctx *BitstreamFilterContext) {
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		bsfCtx := ctx.ctx
		avcodec.FreeBitstreamFilter(&bsfCtx)
		ctx.ctx = bsfCtx
	})

	return ret, nil
}

func (ctx *BitstreamFilterContext) SetInputCodecParameters(params *CodecParameters) {
	if err := averror(avcodec.CopyParameters(ctx.ctx.ParIn, params._codecParameters)); err != nil {
		panic(err)
	}
}

func (ctx *BitstreamFilterContext) SetOutputCodecParameters(params *CodecParameters) {
	if err := averror(avcodec.CopyParameters(ctx.ctx.ParOut, params._codecParameters)); err != nil {
		panic(err)
	}
}

func (ctx *BitstreamFilterContext) init() error {
	ctx.initOnce.Do(func() {
		if ctx.initErr = averror(avcodec.InitBitstreamFilter(ctx.ctx)); ctx.initErr != nil {
			return
		}
	})

	return ctx.initErr
}

func (ctx *BitstreamFilterContext) FilterPacket(inPacket *Packet) ([]*Packet, error) {
	if err := ctx.init(); err != nil {
		return nil, err
	}

	if err := averror(avcodec.SendBitstreamFilterPacket(ctx.ctx, inPacket._packet)); err != nil {
		return nil, err
	}

	var outPackets []*Packet
	for {
		outPacket := NewPacket()
		if err := averror(avcodec.ReceiveBitstreamFilterPacket(ctx.ctx, outPacket._packet)); errors.Is(err, avutil.ErrAgain) {
			break
		} else if err != nil {
			// TODO: consider freeing previously received packets here
			return nil, err
		}

		outPackets = append(outPackets, outPacket)
	}

	return outPackets, nil
}
