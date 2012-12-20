package chantools

import "unsafe"

func ChanDebug(c interface{}) {
	chanDebug(unsafe.Pointer(&c))
}

func chanDebug(c unsafe.Pointer)
