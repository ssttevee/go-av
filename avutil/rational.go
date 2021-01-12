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

// Rat is a convenience function for creating a Rational
func Rat(num, den int32) Rational {
	return Rational{
		Num: num,
		Den: den,
	}
}

func (q Rational) String() string {
	return strconv.Itoa(int(q.Num)) + "/" + strconv.Itoa(int(q.Den))
}

func (q Rational) IsZero() bool {
	return q.Num == 0
}

func (q Rational) Float64() float64 {
	return q2d(q)
}

func (q Rational) Inverse() Rational {
	return Rational{
		Num: q.Den,
		Den: q.Num,
	}
}
