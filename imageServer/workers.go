package imageServer

import "runtime"

var workers chan int

func init() {
	workerCount := runtime.NumCPU() - 1
	workers = make(chan int, workerCount)

	for i := 0; i < workerCount; i++ {
		workers <- i
	}
}
