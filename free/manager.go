package free

import "fmt"

type Configuration struct { 
	MinBlockSize uint
	SlotEntrySize uint
}

type Manager struct {
	configuration *Configuration
	root *Node
}

func NewManager() *Manager {
	return &Manager{root: nil, configuration: nil}
}

func (m *Manager) Configure(c *Configuration) {
	m.configuration = c
}

func (m *Manager) isConfigured() bool {
	return m.configuration != nil
}

func (m *Manager) minBlockSize() uint {
	return m.configuration.MinBlockSize
}

func (m *Manager) slotEntrySize() uint {
	return m.configuration.SlotEntrySize
}

func (m *Manager) Print() {
	fmt.Printf("Available Memory ---\n")
	dfsPrint(m.root, 0)
	fmt.Printf("--- end ---\n")
}

func (m *Manager) Free(offset uint, size uint) (uint, error) {
	if !m.isConfigured() { panic("Cannot add blocks until configured") }
	
	if size < m.minBlockSize() {
		// Is there free space following that could absorb this less-than-minimum block?
		found := dfsLookupByOffset(m.root, offset + size)
		if found == nil {
			return 0, fmt.Errorf("Cannot free blocks less than %d bytes (%d byte(s) requested)", m.minBlockSize, size)
		}
	}

	block := &Block{offset: offset, size: size}
	if m.root == nil {
		m.root = NewNode(block)
	} else {
		dfsAdd(m.root, block)
	}
	
	return size, nil
}

func (m *Manager) AllocateSlot() *Allocation {
	required := m.minBlockSize() + m.slotEntrySize()
	block := m.root.block

	if block.size < required {
		return nil
	}

	alloc := Allocation{Offset: block.offset, Size: m.slotEntrySize()}

	block.offset += m.slotEntrySize()
	block.size -= m.slotEntrySize()

	return &alloc
}

func (m *Manager) isBlockSizeSufficient(requested uint) func(uint) bool {
	return func(size uint) bool {
		return size == requested || (size >= requested + m.minBlockSize())
	}
}

func (m *Manager) Allocate(requested uint) *Allocation {
	found := dfsLookupBySizeClosure(m.root, m.isBlockSizeSufficient(requested))

	if found == nil { return nil }

	fmt.Printf("sanity...found %s\n", found.block)
	if found.block.size == requested { 
		fmt.Printf("sanity...we're removing\n")
		dfsRemove(m.root, found.block.offset) 
	}

	alloc := Allocation{Offset: found.block.offset + found.block.size - requested, Size: requested}

	found.block.size -= requested

	return &alloc
}

func (m *Manager) TryAllocateAt(offset uint, requested uint) *Allocation {
	found := dfsLookupByOffset(m.root, offset)

	if found != nil && m.isBlockSizeSufficient(requested)(found.block.size) {
		block := found.block

		alloc := Allocation{Offset: block.offset, Size: requested}

		block.offset += requested
		block.size -= requested
		
		return &alloc
	}
	
	return nil
}

func (m *Manager) Available() uint {
	return dfsSum(m.root)
}

func (m *Manager) RootOffset() uint {
	return m.root.block.offset
}

func (m *Manager) FreeHeaders(ch chan<- FreeHeader) {
	defer close(ch)
	last := DfsInOrderFreeHeader(m.root, nil, ch)
	ch <- FreeHeader{Offset: last.block.offset, Size: last.block.size, NextOffset: 0}	
}
