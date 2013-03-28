package main

import (
	"fmt"
	"rtlib"
	"unsafe"
)

type Data struct {
	a int
	b int
	c byte
}

func main() {
	rtlib.PrintDebug()
	var a Data
	var b interface{} = a
	var c int
	var d interface{} = c
	fmt.Println(unsafe.Sizeof(a), unsafe.Sizeof(b), unsafe.Sizeof(c), unsafe.Sizeof(d))
	fmt.Println("Hello, playground")
}

/*
func translatedKernel() {
	done := make(chan int)
	a := make([]int, 1000000)
	for i, _ := range a {
		a[i] = i
	}

	go func() {
		// Get device
		GPU := rtlib.GetAccelerator()
		kernel := GPU.MakeKernel("__kernel foo() {}")
		aG := GPU.MirrorGPU(a, false)
		kernel.Run(aG)
		// Copy back modified vars
		a = aG.CopyBack()
		done <- 1
	}()

	<-done
}
*/
