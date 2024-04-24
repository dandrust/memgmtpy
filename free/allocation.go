package free

import "fmt"

type Allocation struct {
	Offset uint
	Size uint
}

func (a Allocation) String() string {
	return fmt.Sprintf("[%d..%d] %d bytes", a.Offset, a.end(), a.Size)
}

func (a Allocation) end() uint {
	return a.Offset + a.Size
}
