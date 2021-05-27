package avcodec

// #include <libavcodec/codec_id.h>
import "C"

type ID C.enum_AVCodecID

const (
	H264 = ID(C.AV_CODEC_ID_H264)
	AAC  = ID(C.AV_CODEC_ID_AAC)
)

func (id ID) String() string {
	return getName(id).String()
}
