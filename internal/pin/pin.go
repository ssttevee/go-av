package pin

import (
	"math/rand"
	"unsafe"
)

type Pin uint32

func Rand() Pin {
	return Pin(rand.Uint32())
}

func Of(p unsafe.Pointer) Pin {
	return Pin(uintptr(p))
}

func (p Pin) Ptr() unsafe.Pointer {
	return unsafe.Pointer(uintptr(p))
}
