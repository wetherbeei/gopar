package main

import "fmt"

func structTest(b []DataB) {
	for idx := range b {
		b[idx].c += 1
	}
}

func operate(idx int, a []int, b []int) (outcome bool) {
	answer := 100 * a[idx]
	b[idx] = answer
	if a := b[idx]; a > 0 {
		b[idx+1] += 1
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
	var z int
	z = <-done
	z++
	z += 1
	for i, val := range a {
		fmt.Println(i, val)
		break
	}
	for i := 0; i < len(a); i++ {
		fmt.Println(i)
		break
	}
	var x DataA
	fmt.Println(x.b.c, z)

	// write-first trigger
	var y = 1
	for i := 1; i < 5; i++ {
		// read-write trigger
		y += 1
	}
	fmt.Println("done")
}
