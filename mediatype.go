package av

// #include <libavformat/avformat.h>
import "C"

type MediaType C.enum_AVMediaType

const (
	Audio = MediaType(C.AVMEDIA_TYPE_AUDIO)
	Video = MediaType(C.AVMEDIA_TYPE_VIDEO)
)

func (t MediaType) ctype() C.enum_AVMediaType{
	return C.enum_AVMediaType(t)
}
