package av

// #include <libavcodec/packet.h>
import "C"
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
		panic(ErrNoMem)
	}

	ret := &Packet{
		_packet: packet,
	}

	runtime.SetFinalizer(ret, func(p *Packet) {
		avcodec.FreePacket(&p._packet)
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
	defer runtime.KeepAlive(p)

	avcodec.UnrefPacket(p._packet)
}
