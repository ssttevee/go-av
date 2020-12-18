package av

// #include <libavutil/samplefmt.h>
import "C"

type SampleFormat C.enum_AVSampleFormat

func (f SampleFormat) ctype() C.enum_AVSampleFormat {
	return C.enum_AVSampleFormat(f)
}
