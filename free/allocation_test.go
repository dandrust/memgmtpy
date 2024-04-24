package free

import "testing"

func TestAllocationEnd(t *testing.T) {
	alloc := Allocation{Offset: 1, Size: 2}

	if alloc.end() != 3 {
		t.Error("Allocation.end()")
	}
}