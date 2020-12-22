package avutil

import (
	"strconv"
)

func RescaleQRound(a int64, b, c Rational, flags Rounding) int64 {
	return RescaleRound(a, int64(b.Num*c.Den), int64(c.Num*b.Den), uint32(flags))
}

func RescaleQ(a int64, b, c Rational) int64 {
	return RescaleQRound(a, b, c, RoundingNearInfinity)
}

func (q Rational) String() string {
	return strconv.Itoa(int(q.Num)) + "/" + strconv.Itoa(int(q.Den))
}

func (q Rational) IsZero() bool {
	return q.Num == 0
}
