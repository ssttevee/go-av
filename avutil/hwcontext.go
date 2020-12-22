package avutil

// #include <libavutil/hwcontext.h>
import "C"

type HWDeviceType C.enum_AVHWDeviceType

const (
	Cuda = HWDeviceType(C.AV_HWDEVICE_TYPE_CUDA)
)

func (f HWDeviceType) String() string {
	return getHWDeviceTypeName(f).String()
}
