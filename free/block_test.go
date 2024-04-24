package free

import "testing"

func TestBlockEnd(t *testing.T) {
	alloc := Block{offset: 1, size: 2}

	if alloc.end() != 3 {
		t.Error("Block.end()")
	}
}