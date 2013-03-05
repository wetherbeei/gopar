package main

import "fmt"

func add(a, b []int) {
	for idx := range a {
		tmp := b[idx]
		a[idx] += tmp
	}
}

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	for i := 0; i < len(a); i++ {
		a[i] = i
		b[i] = i
	}
	add(a, b)
	fmt.Println("done")
}
