package avformat

// #include <libavformat/avformat.h>
import "C"

type MediaType C.enum_AVMediaType

const (
	Audio = MediaType(C.AVMEDIA_TYPE_AUDIO)
	Video = MediaType(C.AVMEDIA_TYPE_VIDEO)
)
