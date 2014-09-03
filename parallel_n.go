package parallel

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// ForN is similar to For, but constraint the maximum number of
// parallel goroutines.
func ForN(low, high, step, parallelism int, worker interface{}) error {
	if low > high {
		return fmt.Errorf("low (%d) > high (%d)", low, high)
	}
	if step <= 0 {
		return fmt.Errorf("step (%d) must be positive", step)
	}

	typ := reflect.TypeOf(worker)
	if typ.Kind() != reflect.Func {
		return errors.New("parallel.ForN worker must be a function.")
	}
	if typ.NumIn() != 1 {
		return errors.New("parallel.ForN worker must have 1 parameter")
	}
	if typ.In(0).Kind() != reflect.Int {
		return errors.New("parallel.ForN worker must have a int param")
	}
	if typ.NumOut() > 1 {
		return errors.New("parallel.ForN worker must return nothing or error")
	}
	if typ.NumOut() == 1 && typ.Out(0).Name() != "error" {
		return errors.New("parallel.ForN worker's return type must be error")
	}

	chin := make(chan int)
	chout := make(chan error)
	var wg sync.WaitGroup
	wg.Add(parallelism)
	var errs string

	go func() {
		for i := low; i < high; i += step {
			chin <- i
		}
		close(chin)
	}()

	val := reflect.ValueOf(worker)
	for i := 0; i < parallelism; i++ {
		go func(val reflect.Value) {
			defer wg.Done()
			for input := range chin {
				if typ.NumOut() == 0 { // worker returns nothing
					val.Call([]reflect.Value{reflect.ValueOf(input)})
				} else { // worker returns an error
					r := val.Call([]reflect.Value{reflect.ValueOf(input)})
					if r[0].Interface() != nil {
						chout <- r[0].Interface().(error)
					}
				}
			}
		}(val)
	}

	go func() {
		wg.Wait()
		close(chout)
	}()

	for e := range chout {
		errs += e.Error() + "\n"
	}
	if len(errs) > 0 {
		return errors.New(errs)
	}
	return nil
}
