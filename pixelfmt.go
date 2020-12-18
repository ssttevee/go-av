package av

// #include <libavutil/pixfmt.h>
// #include <libavutil/pixdesc.h>
import "C"

type PixelFormat C.enum_AVPixelFormat

const (
	PixelFormatCuda = PixelFormat(C.AV_PIX_FMT_CUDA)
	PixelFormatNV12 = PixelFormat(C.AV_PIX_FMT_NV12)
)

func (f PixelFormat) String() string {
	return C.GoString(C.av_get_pix_fmt_name(f.ctype()))
}

func (f PixelFormat) ctype() C.enum_AVPixelFormat {
	return C.enum_AVPixelFormat(f)
}
