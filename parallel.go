package parallel

import (
	"errors"
	"fmt"
	"reflect"
)

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
		<-sem
	}
}

// Do accepts a varadic parameter of functions and execute them in
// parallel.  These functions must have no parameter, and return
// either nothing or an error.  For examples, please refer to
// corresponding unit test.
func Do(functions ...interface{}) error {
	t := make([]reflect.Type, len(functions))

	for i, f := range functions {
		t[i] = reflect.TypeOf(f)
		if t[i].Kind() != reflect.Func {
			return fmt.Errorf("The #%d param of Do is not a function", i+1)
		}
		if t[i].NumIn() != 0 {
			return fmt.Errorf(
				"The #%d param of Do is not a function with out param", i+1)
		}
		if t[i].NumOut() > 1 {
			return fmt.Errorf(
				"The #%d param of Do must return nothing or an error", i+1)
		}
	}

	// sem has sufficiently large buffer that prevents blocking.
	sem := make(chan int, len(functions))
	errs := make([]error, len(functions))

	for i, f := range functions {
		v := reflect.ValueOf(f)
		go func(v reflect.Value, i int) {
			if t[i].NumOut() == 0 { // f returns nothing
				v.Call(nil)
			} else { // f returns an error
				r := v.Call(nil)
				if r[0].Interface() != nil {
					errs[i] = r[0].Interface().(error)
				}
			}
			sem <- 1
		}(v, i)
	}

	for i := 0; i < len(functions); i++ {
		<-sem
	}

	r := ""
	for _, e := range errs {
		if e != nil {
			r = r + fmt.Sprintf("%v\n", e)
		}
	}
	return errors.New(r)
}
