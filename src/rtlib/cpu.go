// CPU-parallel (using goroutines)
//
//      var a, b []int
//      rtlib.CPUParallel(func (_idx int) {
//        <loop.variables>
//        a[idx] += b[idx]
//      }, start, stop)
package rtlib

import (
	"fmt"
	"runtime"
	"sync"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type ParallelFunc func(int)

func CPUParallel(f ParallelFunc, start, stop int) {
	var wg sync.WaitGroup

	cpus := runtime.NumCPU()
	iterations := stop - start
	fmt.Println("CPUParallel", cpus, "iters", iterations)
	for i := 0; i < cpus; i++ {
		wg.Add(1)
		go func(i int) {
			offset := start + i
			for idx := offset; idx < stop; idx = idx + cpus {
				f(idx)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return
}
