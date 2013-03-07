package main

import "fmt"

func doAdd(aArg, bArg []int, i int) {
	// don't follow the i back to idx, because i could be incremented first
	aArg[i] = aArg[i] + bArg[i]
}

func doAddP(aP, bP *int) {
	*aP.b = *aP.b + *bP.b
}

func add(a, b []int) {
	for idx := range a {
		doAdd(a, b, idx)
	}
	for idx := range a {
		doAddP(&a[idx].b.c, &b[idx].b.c)
	}
	for idx := range a {
		a[idx].b.c += b[idx].b.c
	}
}

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	add(a, b)
	fmt.Println("done")
}
