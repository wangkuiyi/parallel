// The package parallel provides OpenMP like syntax for Go.
//
// For
//
// The function parallel.For mimics the for structure, but the loop
// body is executed in parallel.  For example, the following code snippet
// fills slice m by numbers {1,0,3,0}.
//     m := make([]int, 4)
//     for i := 0; i < 4; i += 2 {
//         m[i] = i+1
//     }
// The following snippet does the same thing, but in parallel:
//     parallel.For(0, 4, 2, func(i int) { m[i] = i+1 })
//
// Do
//
// The function parallel.Do accepts a variable number of parameters,
// each should be a function with no parameter nor return value.  It
// executes these functions in parallel.  For example, the following code
// snippet
//     m := make([]int, 4)
//     Do(
//         func() { m[0] = 0 },
//         func() { m[1] = 1 },
//         func() { m[2] = 2 },
//         func() { m[3] = 3 })
// fills m by {0,1,2,3}.
//
package parallel

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// For accepts a worker function with an integer paramter and returns
// either nothing or an error.  It then makes (high-low)/step
// simultaneous invocations to the worker.  These simultaneous
// invocations mimic the for loop: i := low; i < high; i += step,
// where i is the parameter passed to worker.  For examples, please
// refer to unit tests.
func For(low, high, step int, worker interface{}) error {
	if low > high {
		return fmt.Errorf("low (%d) > high (%d)", low, high)
	}
	if step <= 0 {
		return fmt.Errorf("step (%d) must be positive", step)
	}

	t := reflect.TypeOf(worker)
	if t.Kind() != reflect.Func {
		return errors.New("parallel.For worker must be a function.")
	}
	if t.NumIn() != 1 {
		return errors.New("parallel.For worker must have 1 parameter")
	}
	if t.In(0).Kind() != reflect.Int {
		return errors.New("parallel.For worker must have a int param")
	}
	if t.NumOut() > 1 {
		return errors.New("parallel.For worker must return nothing or error")
	}
	if t.NumOut() == 1 && t.Out(0).Name() != "error" {
		return errors.New("parallel.For worker's return type must be error")
	}

	// sem has sufficiently large buffer that prevents blocking.
	sem := make(chan int, high-low)
	v := reflect.ValueOf(worker)
	var errs string
	var mutex sync.Mutex

	for i := low; i < high; i += step {
		go func(v reflect.Value, i int) {
			if t.NumOut() == 0 { // worker returns nothing
				v.Call([]reflect.Value{reflect.ValueOf(i)})
			} else { // worker returns an error
				r := v.Call([]reflect.Value{reflect.ValueOf(i)})
				if r[0].Interface() != nil {
					mutex.Lock()
					defer mutex.Unlock()
					errs += fmt.Sprintf("%v\n", r[0].Interface().(error))
				}
			}
			sem <- 1
		}(v, i)
	}

	for i := low; i < high; i += step {
		<-sem
	}

	if len(errs) > 0 {
		return errors.New(errs)
	}
	return nil
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
		if t[i].NumOut() == 1 && t[i].Out(0).Name() != "error" {
			return errors.New("Return type is not error")
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

	if len(r) > 0 {
		return errors.New(r)
	}
	return nil
}

// RangeMap calls f for each key-value pairs in map m, and collect
// non-nil errors returned by f as its return value.
func RangeMap(m interface{}, f func(k, v reflect.Value) error) error {
	if reflect.TypeOf(m).Kind() != reflect.Map {
		panic(fmt.Sprintf("%+v is not a map", m))
	}

	keys := reflect.ValueOf(m).MapKeys()
	es := make([]error, len(keys))

	var wg sync.WaitGroup
	for i, k := range keys {
		wg.Add(1) // Do not put this line into go func(k,v,i);
		// otherwise wg.Wait might be executed before
		// wg.Add(1)
		go func(k, v reflect.Value, i int) {
			defer wg.Done()
			es[i] = f(k, v)
		}(k, reflect.ValueOf(m).MapIndex(k), i)
	}
	wg.Wait()

	r := ""
	for _, e := range es {
		if e != nil {
			r += fmt.Sprintf("%v\n", e)
		}
	}
	if len(r) > 0 {
		return errors.New(r)
	}
	return nil
}
