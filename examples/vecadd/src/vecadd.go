package main

func main() {
	a := make([]int, 1000000)
	b := make([]int, 1000000)
	work := make(chan int)
	done := make(chan int)
	go func() {
		// Listen for new data on work channel 
		// Kernel copy channel buffer to mem
		// Launch kernel
		for idx := range work {
			b[idx] = 100 * a[idx]
		}
		// Kernel copy back
		done <- 1
	}()
	<-done // b should be done by this point
}
