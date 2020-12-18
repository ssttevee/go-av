package av

// #include <libavcodec/packet.h>
import "C"
import (
	"runtime"
)

type Packet struct {
	packet *C.AVPacket
}

func NewPacket() *Packet {
	packet := C.av_packet_alloc()
	if packet == nil {
		panic(ErrNoMem)
	}

	ret := &Packet{
		packet: packet,
	}

	runtime.SetFinalizer(ret, func(p *Packet) {
		C.av_packet_free(&p.packet)
	})

	return ret
}

func (p *Packet) Rescale(src, dst Rational) {
	//C.av_packet_rescale_ts(p.packet, src.ctype(), dst.ctype())

	p.SetPresentationTimestamp(RescaleQRound(p.PresentationTimestamp(), src, dst, RoundingNearInfinity | RoundingPassMinMax))
	p.SetDecompressionTimestamp(RescaleQRound(p.DecompressionTimestamp(), src, dst, RoundingNearInfinity | RoundingPassMinMax))
	p.SetDuration(RescaleQ(p.Duration(), src, dst))
}

func (p *Packet) StreamIndex() int {
	return int(p.packet.stream_index)
}

func (p *Packet) SetStreamIndex(i int) {
	p.packet.stream_index = C.int(i)
}

func (p *Packet) SetPosition(pos int64) {
	p.packet.pos = C.int64_t(pos)
}

func (p *Packet) PresentationTimestamp() int64 {
	return int64(p.packet.pts)
}

func (p *Packet) SetPresentationTimestamp(pts int64) {
	p.packet.pts = C.int64_t(pts)
}

func (p *Packet) DecompressionTimestamp() int64 {
	return int64(p.packet.dts)
}

func (p *Packet) SetDecompressionTimestamp(dts int64) {
	p.packet.dts = C.int64_t(dts)
}

func (p *Packet) Duration() int64 {
	return int64(p.packet.duration)
}

func (p *Packet) SetDuration(duration int64) {
	p.packet.duration = C.int64_t(duration)
}

func (p *Packet) prepare() *C.AVPacket {
	p.Unref()

	return p.packet
}

func (p *Packet) CopyTo(p2 *Packet) error {
	return averror(C.av_packet_ref(p2.packet, p.packet))
}

func (p *Packet) Clone() (*Packet, error) {
	clone := NewPacket()
	if err := p.CopyTo(clone); err != nil {
		return nil, err
	}

	return clone, nil
}

func (p *Packet) Unref() {
	defer runtime.KeepAlive(p)

	C.av_packet_unref(p.packet)
}
