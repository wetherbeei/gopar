package main

import (
	"chantools"
	"rtlib"
	"unsafe"
)

func main() {
	rtlib.PrintDebug()
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	ptr := unsafe.Pointer(&c)
	chantools.ChanDebug(ptr)
}

func translatedKernel() {
	done := make(chan int)
	a := make([]int, 1000000)
	for i, _ := range a {
		a[i] = i
	}

	go func() {
		// Transfer copies of local vars
		aGPU := rtlib.MirrorGPU(a)
		// Launch kernel
		result := rtlib.Kernel["reduce"].Launch(aGPU)
		result.Wait()
		// Copy back modified vars
		aGPU.CopyBack()
		done <- 1
	}()

	<-done
}
