package av

// #include <libavformat/avformat.h>
import "C"

type Stream struct {
	stream    *C.AVStream
	formatCtx *C.AVFormatContext
}

func (s *Stream) CodecParameters() *CodecParameters {
	return &CodecParameters{
		parameters: s.stream.codecpar,
	}
}

func (s *Stream) Index() int {
	return int(s.stream.index)
}

func (s *Stream) GuessFramerate() Rational {
	return rational(C.av_guess_frame_rate(s.formatCtx, s.stream, nil))
}

func (s *Stream) AverageFrameRate() Rational {
	return rational(s.stream.avg_frame_rate)
}

func (s *Stream) RealFrameRate() Rational {
	return rational(s.stream.r_frame_rate)
}

func (s *Stream) SetCodecParameters(params *CodecParameters) {
	if err := averror(C.avcodec_parameters_copy(s.stream.codecpar, params.parameters)); err != nil {
		panic(err)
	}
}

func (s *Stream) TimeBase() Rational {
	return rational(s.stream.time_base)
}

func (s *Stream) SetTimeBase(r Rational) {
	s.stream.time_base = r.ctype()
}
