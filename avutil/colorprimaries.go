package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/pixfmt.h>
import "C"

type ColorPrimaries C.enum_AVColorPrimaries

const (
	ColorPrimariesUnspecified = ColorPrimaries(C.AVCOL_PRI_UNSPECIFIED)
)

func (cp ColorPrimaries) String() string {
	return getColorPrimariesName(uint32(cp)).String()
}
