package free

import "fmt"

type Block struct {
	offset uint
	size uint
}

func (b Block) String() string {
	return fmt.Sprintf("[%d..%d] %d bytes free", b.offset, b.end(), b.size)
}

func (b Block) end() uint {
	return b.offset + b.size
}
