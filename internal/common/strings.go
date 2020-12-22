package common

import "C"

type CChar C.char

func (s *CChar) String() string {
	return C.GoString((*C.char)(s))
}
