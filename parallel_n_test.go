package parallel

import (
	"errors"
	"testing"
)

func TestForN(t *testing.T) {
	m := make([]int, 40000)
	ForN(0, len(m), 1, 10, func(i int) {
		if i%2 == 0 {
			m[i] = 1
		}
	})
	sum := 0
	for i := 0; i < len(m); i++ {
		sum += m[i]
	}
	if sum != len(m)/2 {
		t.Errorf("Expecting %d, got %d", len(m)/2, sum)
	}
}

func TestForNOutput(t *testing.T) {
	loop := 40000
	e := ForN(0, loop, 1, 10, func(i int) error {
		if i%2 == 0 {
			return errors.New("A")
		}
		return nil
	})
	if l := len(e.Error()); l != loop {
		t.Errorf("Expecting %d, got %d", loop, l)
	}
}
