package main

import "math"

type Point struct {
	x, y float64 // cartesian
	r, θ float64 // polar
}

func main() {
	a := make([]Point, 10000)
	for i := range a {
		a[i].Conv()
	}
	return
}

func (p *Point) Conv() {
	p.θ = math.Atan(p.y / p.x)
	rSquared := math.Pow(p.x, 2) + math.Pow(p.y, 2)
	p.r = math.Sqrt(rSquared)
	return
}
