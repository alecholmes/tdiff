package b

import "unsafe"

func Hello() *unsafe.Pointer {
	return new(unsafe.Pointer)

}
