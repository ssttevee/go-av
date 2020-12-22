package av

// #include <libavformat/avformat.h>
import "C"
import (
	"github.com/ssttevee/go-av/avcodec"
	"github.com/ssttevee/go-av/avformat"
	"github.com/ssttevee/go-av/avutil"
)

type _stream = avformat.Stream

type Stream struct {
	*_stream
	formatCtx *avformat.Context
}

func (s *Stream) GuessFramerate() avutil.Rational {
	return avformat.GuessFrameRate(s.formatCtx, s._stream, nil)
}

func (s *Stream) Codecpar() *CodecParameters {
	return &CodecParameters{
		_codecParameters: s._stream.Codecpar,
	}
}

func (s *Stream) SetCodecpar(params *CodecParameters) {
	if err := averror(avcodec.CopyParameters(s._stream.Codecpar, params._codecParameters)); err != nil {
		panic(err)
	}
}
