package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/pixfmt.h>
import "C"

type ColorTransferCharacteristic C.enum_AVColorTransferCharacteristic

const (
	ColorTransferCharacteristicUnspecified = ColorTransferCharacteristic(C.AVCOL_TRC_UNSPECIFIED)
)

func (ctr ColorTransferCharacteristic) String() string {
	return getColorTransferName(uint32(ctr)).String()
}
