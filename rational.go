package av

import (
	"github.com/ssttevee/go-av/avutil"
)

func Rat(num, den int32) avutil.Rational {
	return avutil.Rational{
		Num: num,
		Den: den,
	}
}
