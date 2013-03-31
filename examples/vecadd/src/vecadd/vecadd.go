package main

import "fmt"

func add(a, b, c []int, d int) {
	for idx, val := range a {
		tmp := b[idx]
		a[idx] = val + tmp + c[tmp] + d
	}
}

func main() {
	a := make([]int, 1000)
	b := make([]int, 1000)
	c := make([]int, 1000)
	for i := 0; i < len(a); i++ {
		a[i] = i
		b[i] = i
	}
	add(a, b, c, 10)
	fmt.Println("done")
}
