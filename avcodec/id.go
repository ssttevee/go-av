package avcodec

// #include <libavcodec/codec_id.h>
import "C"

type ID C.enum_AVCodecID

const (
	H264 = ID(C.AV_CODEC_ID_H264)
)

func (id ID) String() string {
	return getName(uint32(id)).String()
}
