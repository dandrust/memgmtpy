from typing import Optional, Callable

class Block:
    def __init__(self, offset: int, size: int) -> None:
        self.offset = offset
        self.size = size

    def end(self) -> int:
        return self.offset + self.size

    def __str__(self) -> str:
        return f"[{self.offset}..{self.offset + self.size}] {self.size} bytes free"

class Node:
    def __init__(self, block: Block) -> None:
        self.block = block
        self.left = None
        self.right = None

class Allocation:
    def __init__(self, offset: int, size: int) -> None:
        self.offset = offset
        self.size = size

    def __str__(self) -> str:
        return f"[{self.offset}..{self.offset + self.size}] {self.size} allocated"


# free
# available
# allocate_slot
# allocate
# compact
class Manager:
    FREE_BLOCK_ENTRY_SIZE = 4
    SLOT_ENTRY_SIZE = 2

    def __init__(self) -> None:
        self.root = None

    def dfs_lookup_by_end(self, root: Node, target: int) -> Optional[Node]:
        if not root:
            return None

        if (right := self.dfs_lookup_by_end(root.right, target)): return right
        if root.block.end() == target: return root
        return self.dfs_lookup_by_end(root.left, target)
    
    def dfs_lookup_by_offset(self, root: Node, target: int):
        if not root: return None

        if (left := self.dfs_lookup_by_offset(root.left, target)): return left
        if root.block.offset == target: return root
        return self.dfs_lookup_by_offset(root.right, target)

    def add(self, offset: int, size: int) -> None:
        b = Block(offset, size)

        print(self.root)
        if not self.root:
            print("there is no root!")
            self.root = Node(b)
            return

        print("there's....a root?")
        self.dfs_add(self.root, b)

    def dfs_remove(self, root: Node, target: int) -> None:
        # TODO: Account for removing root. Should probably try compaction!
        print(f"Attempting to remove a block starting at {target}")
        if not root: return

        # left
        if (found := self.dfs_remove(root.left, target)):
            print(f"found the block as a left child of {root.block}")
            to_remove = root.left
            if to_remove.right: # there is a right child
                to_remove.left = to_remove.right.left
                root.left = to_remove.right
            else:
                root.left = to_remove.left
            return
        
        # self
        if root.block.offset == target: return True
            
        # right
        if (found := self.dfs_remove(root.right, target)):
            print(f"found the block as a right child of {root.block}")
            to_remove = root.right
            if to_remove.left: # there is a left child

                to_remove.right = to_remove.left.right
                root.right = to_remove.left
            else:
                root.right = to_remove.right
            return

    def dfs_add(self, root: Node, block: Block) -> None:    
        if not root: 
            return
        
        print(f"block is `{block}`, Node's block is `{root.block}`")
        print(f"block end is {block.end()}; root's block's offset is {root.block.offset}")
        if block.end() <= root.block.offset: # Go left
            print("going left!")
            
            if block.end() == root.block.offset:
                print("contiguous! Adjusting right to cover left")
                root.block.offset -= block.size
                root.block.size += block.size
                print(f"can I find a block in my left subtree that ends at {root.block.offset}?")
                # TODO: This could check ONLY the rightmost leaf - that's the only place this could be
                if (contiguous := self.dfs_lookup_by_end(root.left, root.block.offset)):
                    print(f"found one! {contiguous.block}") 
                    # remove it from the tree and return it
                    self.dfs_remove(root.left, contiguous.block.offset)
                    # expand the root
                    root.block.offset -= contiguous.block.size
                    root.block.size += contiguous.block.size
                return
            
            if not root.left:
                root.left = Node(block)
                return
            
            return self.dfs_add(root.left, block)

        if block.offset >= root.block.end(): # Go right
            print("going right")

            if block.offset == root.block.end():
                print("contigous! Extending left to cover right")
                root.block.size += block.size

                print(f"can I find a block in my right subtree that ends at {root.block.end()}?")
                # TODO: This could check ONLY the leftmost leaf - that's the only place this could be
                if (contiguous := self.dfs_lookup_by_offset(root.right, root.block.end())):
                    print(f"found one! {contiguous.block}") 
                    # remove it from the tree and return it
                    self.dfs_remove(root.right, contiguous.block.offset)
                    # expand the root
                    root.block.size += contiguous.block.size
                return

            if not root.right:
                root.right = Node(block)
                return
            
            return self.dfs_add(root.right, block)

        raise("Contiguous handling not yet implemented")

    def print(self) -> None:
        print("Memory Manager ----")
        self.dfs_print(self.root, 0)
        print("--- end ---")

    def dfs_print(self, root: Node, depth: int) -> None:
        if not root:
            return
        
        tab = "\t"
        print(f"{tab * depth}{root.block}")
        self.dfs_print(root.left, depth + 1)
        self.dfs_print(root.right, depth + 1)

    # If this fails we could try compaction. But I think the caller
    # should do this, because if slot allocation fails we want to know
    # if the data will fit before compacting (because compacting will
    # be expensive)
    def allocate_slot(self) -> Optional[Allocation]:
        # The first free block is guaranteed to follow the slot
        # entries, so we'll always allocate slots from the FRONT
        # of the first free block

        if self.root.block.size <= self.SLOT_ENTRY_SIZE + self.FREE_BLOCK_ENTRY_SIZE:
            return None

        alloc = Allocation(self.root.block.offset, self.SLOT_ENTRY_SIZE)

        self.root.block.offset += self.SLOT_ENTRY_SIZE
        self.root.block.size -= self.SLOT_ENTRY_SIZE

        return alloc
    
    def allocate(self, requested: int) -> Optional[Allocation]:
        # We must have either a perfect fit or leave enough space
        # to keep a free block entry (reporting 0 bytes availabe) 
        # that occupies FREE_BLOCK_ENTRY_SIZE bytes
        def is_sufficient(available: int) -> bool:
            return (available == requested) or available >= (requested + self.FREE_BLOCK_ENTRY_SIZE)
        
        if (node := self.dfs_lookup_by_size_closure(self.root, is_sufficient)):
            if node.block.size == requested:
                self.dfs_remove(self.root, node.block.offset)
                return Allocation(node.block.offset, node.block.size)
            else:
                node.block.size -= requested
                return Allocation(node.block.end() - requested, requested)     
        
    def dfs_lookup_by_size_closure(self, root: Node, predicate: Callable[[int], bool]) -> Optional[Node]:
        if not root: return

        if predicate(root.block.size):
            return root
        
        if (found := self.dfs_lookup_by_size_closure(root.left, predicate)):
            return found
        return self.dfs_lookup_by_size_closure(root.right, predicate)
    
    def available(self) -> int:
        return self.dfs_sum(self.root)
    
    def dfs_sum(self, root: Node) -> int:
        if not root: return 0

        return (
            root.block.size + 
            self.dfs_sum(root.left) + 
            self.dfs_sum(root.right)
        )

    # Concentrates all available free space at the offset
    # of the root node
    def compact(self) -> None:
        self.root = Node(
            Block(
                self.root.block.offset,
                self.available()
            )
        )

if __name__ == "__main__":
    m = Manager()
    m.add(100, 9)
    m.add(30, 20)
    m.add(10, 2)
    m.add(18, 7)
    m.add(15, 1)
    m.print()

    print(m.available())

    print("allocating a slot...")
    if (a := m.allocate_slot()):
        print(a)
    else:
        print("allocation failed")

    m.print()

    print("allocating 20")
    a = m.allocate(20)
    print(a)
    m.print()

    print("allocating 6")
    if (a := m.allocate(6)):
        print(a)
    else:
        print("allocation (6) failed")

    print("allocating 5")
    if (a := m.allocate(5)):
        print(a)
    else:
        print("allocation (5) failed")

    print("allocating 4")
    if (a := m.allocate(4)):
        print(a)
    else:
        print("allocation (4) failed")

    print("allocating 3")
    if (a := m.allocate(3)):
        print(a)
    else:
        print("allocation (3) failed")

    m.print()
    print(m.available())


    m.compact()
    print(m.available())
    m.print()




