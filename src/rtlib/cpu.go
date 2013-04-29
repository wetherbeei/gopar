// CPU-parallel (using goroutines)
package rtlib

import (
	"runtime"
	"sync"
)

type ParallelFunc func(int)

// CPUParallel launches GOMAXPROCS goroutines to distribute the indexes between
// [start, stop) evenly.
// TODO: account for some cores being "faster" than others, so use an atomic int
// for each goroutine to get a new chunk of work when it's ready.
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
