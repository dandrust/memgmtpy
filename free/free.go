// A free space manager for pages
package free

type FreeHeader struct {
	Offset uint
	Size uint
	NextOffset uint
}
