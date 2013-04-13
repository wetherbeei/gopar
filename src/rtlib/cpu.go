// CPU-parallel (using goroutines)
//
//      var a, b []int
//      rtlib.CPUParallel(func (_idx int) {
//        <loop.variables>
//        a[idx] += b[idx]
//      }, start, stop)
package rtlib

import (
	"runtime"
	"sync"
)

func init() {
	// TODO: remove this
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type ParallelFunc func(int)

func CPUParallel(f ParallelFunc, start, stop int) {
	var wg sync.WaitGroup

	procs := runtime.GOMAXPROCS(-1)
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func(i int) {
			offset := start + i
			for idx := offset; idx < stop; idx = idx + procs {
				f(idx)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return
}
