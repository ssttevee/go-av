package av

import (
	"runtime"

	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avutil"
)

type _packet = avcodec.Packet

type Packet struct {
	*_packet
}

func NewPacket() *Packet {
	packet := avcodec.NewPacket()
	if packet == nil {
		panic(avutil.ErrNoMem)
	}

	ret := &Packet{
		_packet: packet,
	}

	runtime.SetFinalizer(ret, func(p *Packet) {
		// heap pointer may not be passed to cgo, so use a stack pointer instead :D
		packet := (*avcodec.Packet)(p._packet)
		avcodec.FreePacket(&packet)
		p._packet = packet
	})

	return ret
}

func (p *Packet) Rescale(src, dst avutil.Rational) {
	p.Pts = avutil.RescaleQRound(p.Pts, src, dst, avutil.RoundingNearInfinity|avutil.RoundingPassMinMax)
	p.Dts = avutil.RescaleQRound(p.Dts, src, dst, avutil.RoundingNearInfinity|avutil.RoundingPassMinMax)
	p.Duration = avutil.RescaleQ(p.Duration, src, dst)
}

func (p *Packet) prepare() *avcodec.Packet {
	p.Unref()

	return p._packet
}

func (p *Packet) CopyTo(p2 *Packet) error {
	return averror(avcodec.RefPacket(p2._packet, p._packet))
}

func (p *Packet) Clone() (*Packet, error) {
	clone := NewPacket()
	if err := p.CopyTo(clone); err != nil {
		return nil, err
	}

	return clone, nil
}

func (p *Packet) Unref() {
	avcodec.UnrefPacket(p._packet)
}
