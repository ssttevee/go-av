package av

import (
	"math/rand"
	"unsafe"
)

type pinType uint32

func randPin() pinType {
	return pinType(rand.Uint32())
}

func pin(p unsafe.Pointer) pinType {
	return pinType(uintptr(p))
}

func pinptr(p pinType) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p))
}
