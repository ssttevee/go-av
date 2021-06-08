package av

import (
	"sync/atomic"
	"unsafe"
)

type pinType uint32

var pinCounter uint32

func randPin() pinType {
	return pinType(atomic.AddUint32(&pinCounter, 1))
}

func pin(p unsafe.Pointer) pinType {
	return pinType(uintptr(p))
}

func pinptr(p pinType) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p))
}
