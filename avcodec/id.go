package avcodec

// #include <libavcodec/codec_id.h>
import "C"

type ID C.enum_AVCodecID

const (
	AV1  = ID(C.AV_CODEC_ID_AV1)
	HEVC = ID(C.AV_CODEC_ID_HEVC)
	H264 = ID(C.AV_CODEC_ID_H264)
	AAC  = ID(C.AV_CODEC_ID_AAC)
)

func (id ID) String() string {
	return getName(id).String()
}
