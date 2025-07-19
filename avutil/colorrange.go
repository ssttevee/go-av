package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/pixfmt.h>
import "C"

type ColorRange C.enum_AVColorRange

const (
	ColorRangeUnspecified = ColorRange(C.AVCOL_RANGE_UNSPECIFIED)
)

func (cr ColorRange) String() string {
	return getColorRangeName(uint32(cr)).String()
}
