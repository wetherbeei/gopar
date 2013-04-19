package main

import "fmt"

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	c := make([]int, 1000000)
	for i := range c {
		c[i] = a[i] + b[i]
	}
	fmt.Println("done")
}
