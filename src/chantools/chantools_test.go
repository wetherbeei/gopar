package chantools

import "testing"

//import "unsafe"

func TestLock(t *testing.T) {
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	//ptr := unsafe.Pointer(&c)
	ChanDebug(c)
}
