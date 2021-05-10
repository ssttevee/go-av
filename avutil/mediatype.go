package avutil

// #include <libavutil/avutil.h>
import "C"

type MediaType C.enum_AVMediaType

const (
	Audio = MediaType(C.AVMEDIA_TYPE_AUDIO)
	Video = MediaType(C.AVMEDIA_TYPE_VIDEO)
)

func (t MediaType) String() string {
	return getMediaTypeString(t).String()
}
