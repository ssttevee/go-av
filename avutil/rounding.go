package avutil

// #include <libavutil/mathematics.h>
import "C"

type Rounding C.enum_AVRounding

const (
	RoundingPassMinMax   = Rounding(C.AV_ROUND_PASS_MINMAX)
	RoundingNearInfinity = Rounding(C.AV_ROUND_NEAR_INF)
)
