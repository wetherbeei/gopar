/*
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

    * Redistributions of source code must retain the above copyright
    notice, this list of conditions and the following disclaimer.

    * Redistributions in binary form must reproduce the above copyright
    notice, this list of conditions and the following disclaimer in the
    documentation and/or other materials provided with the distribution.

    * Neither the name of "The Computer Language Benchmarks Game" nor the
    name of "The Computer Language Shootout Benchmarks" nor the names of
    its contributors may be used to endorse or promote products derived
    from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

/* The Computer Language Benchmarks Game
 * http://shootout.alioth.debian.org/
 *
 * contributed by The Go Authors.
 * based on C program by Christoph Bauer
 */

package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
)

var n = flag.Int("n", 1000, "number of iterations")
var v = flag.Int("v", 600, "number of planets")

type Body struct {
	x, y, z, vx, vy, vz, mass float64
}

const (
	solarMass   = 4 * math.Pi * math.Pi
	daysPerYear = 365.24
)

func (b *Body) offsetMomentum(px, py, pz float64) {
	b.vx = -px / solarMass
	b.vy = -py / solarMass
	b.vz = -pz / solarMass
}

type System struct {
	planets []Body
	results []Body
}

func NewSystem(planets []Body) *System {
	var px, py, pz float64
	for _, body := range planets {
		px += body.vx * body.mass
		py += body.vy * body.mass
		pz += body.vz * body.mass
	}
	planets[0].offsetMomentum(px, py, pz)
	return &System{
		planets: planets,
		results: make([]Body, len(planets)),
	}
}

func (sys *System) energy() float64 {
	var e float64
	for i, body := range sys.planets {
		e += 0.5 * body.mass *
			(body.vx*body.vx + body.vy*body.vy + body.vz*body.vz)
		for j := i + 1; j < len(sys.planets); j++ {
			body2 := sys.planets[j]
			dx := body.x - body2.x
			dy := body.y - body2.y
			dz := body.z - body2.z
			distance := math.Sqrt(dx*dx + dy*dy + dz*dz)
			e -= (body.mass * body2.mass) / distance
		}
	}
	return e
}

func (sys *System) advance(dt float64) {
	for i := range sys.results {
		body := sys.planets[i]
		for j := 0; j < len(sys.planets); j++ {
			if j == i {
				continue // don't advance ourselves
			}
			body2 := sys.planets[j]
			dx := body.x - body2.x
			dy := body.y - body2.y
			dz := body.z - body2.z

			dSquared := dx*dx + dy*dy + dz*dz
			distance := math.Sqrt(dSquared)
			mag := dt / (dSquared * distance)

			body.vx -= dx * body2.mass * mag
			body.vy -= dy * body2.mass * mag
			body.vz -= dz * body2.mass * mag
		}
		sys.results[i] = body
	}

	for i, body := range sys.results {
		body.x += dt * body.vx
		body.y += dt * body.vy
		body.z += dt * body.vz
		sys.results[i] = body
	}
	// swap lists
	sys.planets, sys.results = sys.results, sys.planets
}

func main() {
	flag.Parse()
	var planets []Body
	r := rand.New(rand.NewSource(1234))
	for i := 0; i < *v; i++ {
		planets = append(planets, Body{
			x:    r.Float64(),
			y:    r.Float64(),
			z:    r.Float64(),
			vx:   r.Float64(),
			vy:   r.Float64(),
			vz:   r.Float64(),
			mass: r.Float64(),
		})
	}
	system := NewSystem(planets)
	fmt.Printf("%.9f\n", system.energy())
	for i := 0; i < *n; i++ {
		system.advance(0.01)
	}
	fmt.Printf("%.9f\n", system.energy())
}
