package main

import "fmt"

type Body struct {
	weight int
	x, y   int
}

func nbody(a, b []Body) {
	for idx, this := range a {
		for _, val := range a {
			dist := this.x - val.x + this.y - val.y + 100
			b[idx].weight += val.weight / dist
		}
	}
}

func main() {
	a := make([]Body, 20000)
	b := make([]Body, 20000)
	for i := 0; i < len(a); i++ {
		a[i].weight = i
		a[i].x = i % 50
		a[i].y = i % 50
	}
	for i := 0; i < 5; i++ {
		nbody(a, b)
	}
	fmt.Println("done")
}
