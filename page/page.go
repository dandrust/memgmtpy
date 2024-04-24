// A slotted-page implementation
package page

import (
	"memgmtgo/free"
	"fmt"
	"encoding/binary"
	"memgmtgo/tuple"
)

const (
	schemaVersion = 1
	magicNumber = 0xBEEFFACE
	magicNumberSize = 4
	versionOffset = magicNumberSize
	versionSize = 2
	freeOffset = versionOffset + versionSize
	freeSize = 2
	slotCountOffset = freeOffset + freeSize
	slotCountSize = 2
	slotArrayOffset = slotCountOffset + slotCountSize

	pageSize = 4096
	headerSize = magicNumberSize + versionSize + freeSize + slotCountSize
	slotEntrySize = 2
	freeBlockHeaderSize = 4
)

type Offset uint16

type TuplePeek struct {
	offset Offset
	size uint
}

type Page struct {
	magicNumber uint32
	version uint16
	free Offset
	slots []Offset    // slot_count and slot entries are derived from this slice
	data *[4096]uint8
	mem *free.Manager
}

func NewPage() *Page {
	p := Page{}

	p.magicNumber = magicNumber
	p.version = 1
	p.slots = []Offset{}
	p.data = &[4096]uint8{}

	p.mem = free.NewManager()
	p.mem.Configure(&free.Configuration{MinBlockSize: freeBlockHeaderSize, SlotEntrySize: slotEntrySize})
	p.mem.Free(headerSize, pageSize - headerSize)

	return &p
}

func LoadPage(buffer *[4096]byte) *Page {
	// Capture data
	p := Page{data: buffer}
	
	// Parse metadata
	p.magicNumber = binary.BigEndian.Uint32(p.data[:magicNumberSize])
	p.version = binary.BigEndian.Uint16(p.data[versionOffset:(versionOffset + versionSize)])
	p.free = Offset(binary.BigEndian.Uint16(p.data[freeOffset:(freeOffset + freeSize)]))

	slotCount := int(binary.BigEndian.Uint16(p.data[slotCountOffset:(slotCountOffset + slotCountSize)]))
	p.slots = make([]Offset, slotCount)

	// Populate slots
	var slotEntryOffset uint
	for i := 0; i < slotCount; i++ {
		slotEntryOffset = uint(slotArrayOffset + (i * slotEntrySize))
		p.slots[i] = Offset(binary.BigEndian.Uint16(p.data[slotEntryOffset:(slotEntryOffset + slotEntrySize)]))
	}

	// Configure memory manager
	p.mem = free.NewManager()
	p.mem.Configure(&free.Configuration{MinBlockSize: freeBlockHeaderSize, SlotEntrySize: slotEntrySize})

	// Walk the linked list
	offset := p.free
	var nextOffset Offset
	var size uint16

	for offset != 0 {
		size = binary.BigEndian.Uint16(p.data[offset:(offset + 2)]) // TODO: declare width as a const somewhere
		nextOffset = Offset(binary.BigEndian.Uint16(p.data[(offset + 2):(offset + 4)])) // TODO: declare width as a const somewhere
		p.mem.Free(uint(offset), uint(size))
		offset = nextOffset
	}

	return &p
}

func (p *Page) WriteMetadata() {
	// Write magic number
	magicBytes := [4]byte{
		byte(p.magicNumber >> 24),
		byte(p.magicNumber >> 16),
		byte(p.magicNumber >> 8),
		byte(p.magicNumber)}

	copy(p.data[:magicNumberSize], magicBytes[:])

	// Write version
	binary.BigEndian.PutUint16(p.data[versionOffset:], uint16(schemaVersion))

	// Write slot count and slots
	binary.BigEndian.PutUint16(p.data[slotCountOffset:], uint16(len(p.slots)))
	for i, slotEntry := range p.slots {
		offset := slotArrayOffset + (i * slotEntrySize)
		binary.BigEndian.PutUint16(p.data[offset:], uint16(slotEntry))
	}

	// Write free pointer and free block headers
	binary.BigEndian.PutUint16(p.data[freeOffset:], uint16(p.mem.RootOffset()))

	freeHeaders := make(chan free.FreeHeader, 16)
	go p.mem.FreeHeaders(freeHeaders)

	for header := range freeHeaders {
		bytes := [4]byte{}
		binary.BigEndian.PutUint16(bytes[:], uint16(header.Size))
		binary.BigEndian.PutUint16(bytes[2:], uint16(header.NextOffset))

		copy(p.data[header.Offset:], bytes[:])
	}
}

func (p *Page) SlotCount() uint {
	return uint(len(p.slots))
}

func (p *Page) Insert(t *tuple.PersonTuple) uint {
	// Allocate space for a slot
	slotAlloc := p.mem.AllocateSlot()

	// Capture slot number before we mutate the slots array, plus some sanity checking
	slotNumber := p.SlotCount()
	fmt.Println("slot number claim is ", )
	if slotAlloc.Offset != headerSize + (slotNumber * slotEntrySize) { panic("Slot array and data not aligned!") }

	// Allocate space for the tuple
	heapAlloc := p.mem.Allocate(t.Len())

	// Record the offset where we'll store the tuple
	p.slots = append(p.slots, Offset(heapAlloc.Offset))

	// Write the tuple to the page
	copy(p.data[heapAlloc.Offset:], *t.Bytes())

	return slotNumber
}

func (p *Page) Delete(slotNumber uint) {
	peek := p.tuplePeek(slotNumber)

	// Clear out the slot entry at the slot number
	p.slots[slotNumber] = 0

	// Free the tuple memory
	p.mem.Free(uint(peek.offset), peek.size)
}

func (p *Page) Update(slotNumber uint, t *tuple.PersonTuple) {
	newTupleSize := t.Len()
	existingTuplePeek := p.tuplePeek(slotNumber)

	if newTupleSize == existingTuplePeek.size {
		// write the new tuple to the offset, no need to fuss with memory
		copy(p.data[existingTuplePeek.offset:], *t.Bytes())
		return
	} else if newTupleSize < existingTuplePeek.size {
		// Try to free the excess space.  If we can, write the updated
		// (abbreviated) tuple
		diff := existingTuplePeek.size - newTupleSize
		_, err := p.mem.Free(uint(existingTuplePeek.offset) + newTupleSize, diff)

		if err == nil {
			copy(p.data[existingTuplePeek.offset:], *t.Bytes())
			return
		}
		
	} else if newTupleSize > existingTuplePeek.size {
		// Try to allocate space immediately after. If we can, wreit eh
		// updated (expanded) tuple
		diff := newTupleSize - existingTuplePeek.size
		existingTupleEnd := uint(existingTuplePeek.offset) + existingTuplePeek.size
		alloc := p.mem.TryAllocateAt(uint(existingTupleEnd), diff)

		if alloc != nil {
			copy(p.data[existingTuplePeek.offset:], *t.Bytes())
			return
		}
	}

	// If we end up here we cannot keep the tuple offset constant. Allocate
	// a new block of memory and "move" the slot entry to the new offset, and
	// free the memory at the old offset
	alloc := p.mem.Allocate(newTupleSize)
	p.mem.Free(uint(existingTuplePeek.offset), existingTuplePeek.size)

	p.slots[slotNumber] = Offset(alloc.Offset)
	copy(p.data[alloc.Offset:], *t.Bytes())
}

func (p *Page) tuplePeek(slotNumber uint) *TuplePeek {
	// Read the tuple offset from the slot entry
	offset := p.slots[slotNumber]

	// Follow the offset to read the tuple length
	size := uint(binary.BigEndian.Uint16(p.data[offset:]))

	return &TuplePeek{offset: offset, size: size}
}

func (p *Page) DebugData() {
	fmt.Println(p.data)
}

func (p *Page) DebugMemory() {
	p.mem.Print()
}

func (p *Page) DebugSlots() {
	fmt.Println("Slots: ", p.slots)
}

func (p *Page) Data() []byte {
	return p.data[:]
}
