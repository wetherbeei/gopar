package main

import "fmt"

func add(a, b []int) {
	for idx := range a {
		a[idx] += b[idx]
	}
}

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	add(a, b)
	fmt.Println("done")
}
