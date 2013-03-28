package main

import "fmt"

func add(a, b, c []int) {
	for idx, val := range a {
		tmp := b[idx]
		a[idx] = val + tmp + c[tmp]
	}
}

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	c := make([]int, 1000000)
	for i := 0; i < len(a); i++ {
		a[i] = i
		b[i] = i
	}
	add(a, b, c)
	fmt.Println("done")
}
