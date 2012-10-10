package parallel

import "testing"

func TestFor(t *testing.T) {
	m := make([]int, 4)
	For(0, 4, 1, func(index int) { m[index] = index })
	if m[0] != 0 || m[1] != 1 || m[2] != 2 || m[3] != 3 {
		t.Errorf("Failed setting arry m using For\n")
	}
}

func TestDo(t *testing.T) {
	m := make([]int, 4)
	Do(
		func() { m[0] = 0 },
		func() { m[1] = 1 },
		func() { m[2] = 2 },
		func() { m[3] = 3 })
	if m[0] != 0 || m[1] != 1 || m[2] != 2 || m[3] != 3 {
		t.Errorf("Failed setting arry m using Do. m=%v\n", m)
	}
}
