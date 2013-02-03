package main

import "fmt"

func operate(idx int, a []int, b []int) (outcome bool) {
	answer := 100 * a[idx]
	b[idx] = answer
	if a := b[idx]; a > 0 {
		b[idx] += 1
	} else if bx := a; a == 0 {
		b[idx] = bx
	} else {
		b[idx] -= 1
	}
	return
}

type DataB struct {
	c int
}
type DataA struct {
	a int
	b DataB
}

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	done := make(chan int)
	go func() {
		// Listen for new data on work channel 
		// Kernel copy channel buffer to mem
		// Launch kernel
		for idx, _ := range a {
			operate(idx, a, b)
		}
		// Kernel copy back
		done <- 1
	}()
	<-done // b should be done by this point
	var x DataA
	fmt.Println(x.b.c)
	fmt.Println("done")
}
