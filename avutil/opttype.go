package avutil

// #include <libavutil/avutil.h>
// #include <libavutil/opt.h>
import "C"

type OptionType C.enum_AVOptionType

const (
	OptionTypeBool      = OptionType(C.AV_OPT_TYPE_BOOL)
	OptionTypeFlags     = OptionType(C.AV_OPT_TYPE_FLAGS)
	OptionTypeInt       = OptionType(C.AV_OPT_TYPE_INT)
	OptionTypeInt64     = OptionType(C.AV_OPT_TYPE_INT64)
	OptionTypeUint64    = OptionType(C.AV_OPT_TYPE_UINT64)
	OptionTypeFloat     = OptionType(C.AV_OPT_TYPE_FLOAT)
	OptionTypeDouble    = OptionType(C.AV_OPT_TYPE_DOUBLE)
	OptionTypeVideoRate = OptionType(C.AV_OPT_TYPE_VIDEO_RATE)
	OptionTypeRational  = OptionType(C.AV_OPT_TYPE_RATIONAL)
	OptionTypeConst     = OptionType(C.AV_OPT_TYPE_CONST)
)
