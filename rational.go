package av

// #include <libavutil/mathematics.h>
import "C"
import (
	"strconv"
)

type Rational struct {
	Num int
	Den int
}

func Rat(num, den int) Rational {
	return Rational{
		Num: num,
		Den: den,
	}
}

func (r Rational) String() string {
	return strconv.Itoa(r.Num) + "/" + strconv.Itoa(r.Den)
}

func (r Rational) Inverse() Rational {
	return Rational{
		Num: r.Den,
		Den: r.Num,
	}
}

func (r Rational) IsZero() bool {
	return r.Num == 0 || r.Den == 0
}

func (r Rational) Gt(r2 Rational) bool {
	return r.Num * r2.Den > r2.Num * r.Den
}

func (r Rational) ctype() C.AVRational {
	return C.AVRational{
		num: C.int(r.Num),
		den: C.int(r.Den),
	}
}

func rational(r C.AVRational) Rational {
	return Rational{
		Num: int(r.num),
		Den: int(r.den),
	}
}

type RoundingFlag C.enum_AVRounding

const (
	RoundingPassMinMax   = RoundingFlag(C.AV_ROUND_PASS_MINMAX)
	RoundingNearInfinity = RoundingFlag(C.AV_ROUND_NEAR_INF)
)

func (f RoundingFlag) ctype() C.enum_AVRounding {
	return C.enum_AVRounding(f)
}

func RescaleRound(a, b, c int64, flags RoundingFlag) int64 {
	return int64(C.av_rescale_rnd(C.int64_t(a), C.int64_t(b), C.int64_t(c), flags.ctype()))
}

func RescaleQRound(a int64, b, c Rational, flags RoundingFlag) int64 {
	return RescaleRound(a, int64(b.Num * c.Den), int64(c.Num * b.Den), flags)
}

func RescaleQ(a int64, b, c Rational) int64 {
	return RescaleQRound(a, b, c, RoundingNearInfinity)
}
