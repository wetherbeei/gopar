// CPU-parallel (using goroutines)
//
//      var a, b []int
//      rtlib.CPUParallel(func (_idx int) {
//        <loop.variables>
//        a[idx] += b[idx]
//      }, start, stop)
package rtlib

import (
	"math"
	"runtime"
	"sync"
)

type ParallelFunc func(int)

func CPUParallel(f ParallelFunc, start, stop int) {
	var wg sync.WaitGroup

	cpus := runtime.NumCPU()
	iterations := start - stop
	perCPU := int(math.Ceil(float64(iterations) / float64(cpus)))

	for i := 0; i < cpus; i++ {
		wg.Add(1)
		go func() {
			for j := perCPU * i; j < perCPU*(i+1); j++ {
				idx := j + start
				if idx < stop {
					f(idx)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
