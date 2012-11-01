package parallel

import (
	"fmt"
	"testing"
)

func TestFor(t *testing.T) {
	m := make([]int, 4)
	For(0, 4, 1, func(index int) { m[index] = index })
	if m[0] != 0 || m[1] != 1 || m[2] != 2 || m[3] != 3 {
		t.Errorf("Failed setting arry m using For\n")
	}
}

func TestForWithErrors(t *testing.T) {
	m := make([]int, 4)
	e := For(0, 4, 1, func(index int) error {
		m[index] = index
		if index < 2 {
			return fmt.Errorf("Error %d", index)
		}
		return nil
	})

	if m[0] != 0 || m[1] != 1 || m[2] != 2 || m[3] != 3 {
		t.Errorf("Failed setting arry m using For\n")
	}

	if e.Error() != "Error 0\nError 1\n" &&
		e.Error() != "Error 1\nError 0\n" {
		t.Errorf("Got unexpected errors:%v", e)
	}
}

func TestNestedFor(t *testing.T) {
	m := make([]int, 4)
	For(0, 2, 1, func(i int) {
		For(0, 2, 1, func(j int) {
			m[i*2+j] = i*2 + j
		})
	})

	if m[0] != 0 || m[1] != 1 || m[2] != 2 || m[3] != 3 {
		t.Errorf("Failed setting arry m using For\n")
	}
}

func TestDo(t *testing.T) {
	m := make([]int, 4)

	e := Do(
		func() error { m[0] = 0; return fmt.Errorf("First error") },
		func() error { m[1] = 1; return nil },
		func() error { m[2] = 2; return fmt.Errorf("Second error") },
		func() { m[3] = 3 })

	if e == nil {
		t.Errorf("Failed capturing errors")
	}

	if e.Error() != "First error\nSecond error\n" {
		t.Errorf("Captured wrong errors")
	}

	if m[0] != 0 || m[1] != 1 || m[2] != 2 || m[3] != 3 {
		t.Errorf("Failed setting arry m using Do. m=%v\n", m)
	}
}
