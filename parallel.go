package parallel

import "fmt"

func For(low, high, step int, worker func(index int)) {
	if low > high {
		panic(fmt.Sprintf("low (%d) > high (%d)", low, high))
	}

	sem := make(chan int, high-low)

	for i := low; i < high; i += step {
		go func(i int) {
			worker(i)
			sem <- 1
		}(i)
	}

	for i := low; i < high; i += step {
		<- sem
	}
}

func Do(workers ...func()) {
	sem := make(chan int, len(workers))

	for _, worker := range workers {
		go func(w func()) {
			w()
			sem <- 1
		}(worker)
	}

	for i := 0; i < len(workers); i++ {
		<- sem
	}
}
