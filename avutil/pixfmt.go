package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/pixfmt.h>
import "C"

type PixelFormat C.enum_AVPixelFormat

const (
	PixelFormatCuda = PixelFormat(C.AV_PIX_FMT_CUDA)
	PixelFormatNV12 = PixelFormat(C.AV_PIX_FMT_NV12)
)

func (f PixelFormat) String() string {
	return getPixelFormatName(f).String()
}
