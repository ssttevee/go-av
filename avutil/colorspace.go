package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/pixfmt.h>
import "C"

type ColorSpace C.enum_AVColorSpace

const (
	ColorSpaceUnspecified = ColorSpace(C.AVCOL_SPC_UNSPECIFIED)
)

func (cs ColorSpace) String() string {
	return getColorSpaceName(uint32(cs)).String()
}
