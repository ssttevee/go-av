package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/samplefmt.h>
import "C"

type SampleFormat C.enum_AVSampleFormat

const (
	SampleFormatNone = SampleFormat(C.AV_SAMPLE_FMT_NONE)
	SampleFormatU8   = SampleFormat(C.AV_SAMPLE_FMT_U8)
	SampleFormatS16  = SampleFormat(C.AV_SAMPLE_FMT_S16)
	SampleFormatS32  = SampleFormat(C.AV_SAMPLE_FMT_S32)
	SampleFormatFLT  = SampleFormat(C.AV_SAMPLE_FMT_FLT)
	SampleFormatDBL  = SampleFormat(C.AV_SAMPLE_FMT_DBL)
	SampleFormatU8P  = SampleFormat(C.AV_SAMPLE_FMT_U8P)
	SampleFormatS16P = SampleFormat(C.AV_SAMPLE_FMT_S16P)
	SampleFormatS32P = SampleFormat(C.AV_SAMPLE_FMT_S32P)
	SampleFormatFLTP = SampleFormat(C.AV_SAMPLE_FMT_FLTP)
	SampleFormatDBLP = SampleFormat(C.AV_SAMPLE_FMT_DBLP)
	SampleFormatS64  = SampleFormat(C.AV_SAMPLE_FMT_S64)
	SampleFormatS64P = SampleFormat(C.AV_SAMPLE_FMT_S64P)
)

func (f SampleFormat) String() string {
	return getSampleFormatName(f).String()
}
