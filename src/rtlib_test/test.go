package main

import (
	"chantools"
	"rtlib"
)

func main() {
	rtlib.PrintDebug()
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	chantools.ChanDebug(c)
}

func translatedKernel() {
	done := make(chan int)
	a := make([]int, 1000000)
	for i, _ := range a {
		a[i] = i
	}

	go func() {
		// Get device
		GPU := rtlib.GetAccelerator()
		// Transfer copies of local vars
		aG := GPU.MirrorGPU(a)
		// Launch kernel
		result := rtlib.Kernel["reduce"].Launch(aG)
		result.Wait()
		// Copy back modified vars
		aG.CopyBack()
		done <- 1
	}()

	<-done
}
