package av

// #include <libavcodec/bsf.h>
import "C"
import (
	"errors"
	"runtime"
	"sync"
	"unsafe"
)

type BitstreamFilterNotFoundError string

func (e BitstreamFilterNotFoundError) Error() string {
	return "bitstream not found: " + string(e)
}

type BitstreamFilter struct {
	filter *C.AVBitStreamFilter
}

func FindBitstreamFilterByName(name string) (*BitstreamFilter, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	filter := C.av_bsf_get_by_name(cname)
	if filter == nil {
		return nil, BitstreamFilterNotFoundError(name)
	}

	return &BitstreamFilter{filter: filter}, nil
}

type BitstreamFilterContext struct {
	ctx *C.AVBSFContext

	initOnce sync.Once
	initErr  error
}

func NewBitstreamFilterContext(filter *BitstreamFilter) (*BitstreamFilterContext, error) {
	var ctx *C.AVBSFContext
	if err := averror(C.av_bsf_alloc(filter.filter, &ctx)); err != nil {
		return nil, err
	}

	ret := &BitstreamFilterContext{ctx: ctx}

	runtime.SetFinalizer(ret, func(ctx *BitstreamFilterContext) {
		C.av_bsf_free(&ctx.ctx)
	})

	return ret, nil
}

func (ctx *BitstreamFilterContext) SetInputCodecParameters(params *CodecParameters) {
	if err := averror(C.avcodec_parameters_copy(ctx.ctx.par_in, params.parameters)); err != nil {
		panic(err)
	}
}

func (ctx *BitstreamFilterContext) SetOutputCodecParameters(params *CodecParameters) {
	if err := averror(C.avcodec_parameters_copy(ctx.ctx.par_out, params.parameters)); err != nil {
		panic(err)
	}
}

func (ctx *BitstreamFilterContext) init() error {
	ctx.initOnce.Do(func() {
		if ctx.initErr = averror(C.av_bsf_init(ctx.ctx)); ctx.initErr != nil {
			return
		}
	})

	return ctx.initErr
}

func (ctx *BitstreamFilterContext) FilterPacket(inPacket *Packet) ([]*Packet, error) {
	if err := ctx.init(); err != nil {
		return nil, err
	}

	if err := averror(C.av_bsf_send_packet(ctx.ctx, inPacket.packet)); err != nil {
		return nil, err
	}

	var outPackets []*Packet
	for {
		outPacket := NewPacket()
		if err := averror(C.av_bsf_receive_packet(ctx.ctx, outPacket.packet)); errors.Is(err, ErrAgain) {
			break
		} else if err != nil {
			// TODO: consider freeing previously received packets here
			return nil, err
		}

		outPackets = append(outPackets, outPacket)
	}

	return outPackets, nil
}
