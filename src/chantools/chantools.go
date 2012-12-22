package chantools

import "unsafe"

func chanDebug(c interface{}) {
	chanDebug(unsafe.Pointer(&c))
}

func ChanDebug(c interface{})
