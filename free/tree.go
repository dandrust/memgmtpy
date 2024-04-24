package free

import "fmt"

type Node struct {
	block *Block
	left *Node
	right *Node
}

func NewNode(block *Block) *Node {
	return &Node{block: block, left: nil, right: nil}
}

func dfsPrint(root *Node, depth uint) {
	for i := uint(0); i < depth; i++ {
		fmt.Printf("  ")
	}

	if root == nil {
		fmt.Printf("[EMPTY]\n")
		return
	}

	fmt.Printf("%s\n", root.block)
	dfsPrint(root.left, depth + 1)
	dfsPrint(root.right, depth + 1)
}

func first(root *Node) *Node {
	if root == nil { return root }
	
	var out *Node

	for root != nil {
		out = root
		root = root.left
	}

	return out
}

func last(root *Node) *Node {
	if root == nil { return root }
	
	var out *Node

	for root != nil {
		out = root
		root = root.right
	}

	return out
}

func dfsRemove(root *Node, target uint) bool {
	if root == nil { return false }

	if root.block.offset == target { return true }
	
	if dfsRemove(root.left, target) {
		// Remove the left child
		to_remove := root.left
		
		if to_remove.right != nil {
			// TODO: There's a bug here. This should remove and re-add
			// on of the children
			root.left = to_remove.right
			root.left.left = to_remove.left
		} else {
			root.left = to_remove.left
		}

		return false
	}

	if dfsRemove(root.right, target) {
		// Remove the right child
		to_remove := root.right
		
		if to_remove.left != nil {
			// TODO: There's a bug here. This should remove and re-add
			// on of the children
			root.right = to_remove.left
			root.right.right = to_remove.right
		} else {
			root.right = to_remove.right
		}

		return false
	}

	return false
}

func dfsAdd(root *Node, block *Block) {
	if root == nil { return }

	if root.block.offset == block.end() {
		fmt.Printf("We're in the contiguous merge condition\n")
		// Shares left border with root
		root.block.offset -= block.size
		root.block.size += block.size

		// Can we find a block in the left subtree that
		// shares it's end with our new offset?
		found := last(root.left)
		if found != nil && found.block.end() == root.block.offset {
			fmt.Printf("found %s\n", found.block)
			// remove it from the tree
			dfsRemove(root.left, found.block.offset)
			// add it to the current node
			root.block.offset -= found.block.size
			root.block.size += found.block.size
		}

	} else if root.block.offset > block.end() {
		// Less than root, go left
		if root.left == nil {
			root.left = NewNode(block)
		} else {
			dfsAdd(root.left, block)
		}

	} else if block.offset == root.block.end() {
		// Shares right border with root
		root.block.size += block.size

		// Can we find a block in the right subtree that
		// shares it's offset with our new end?
		found := first(root.right)
		if found != nil && found.block.offset == root.block.end() { 
			// fmt.Printf("found %s\n", found.block)
			// remove it from the tree
			dfsRemove(root.right, found.block.offset)
			// add it to the current node
			root.block.size += found.block.size
		}

	} else if block.end() > root.block.offset {
		// Greater than root, go right
		if root.right == nil {
			root.right = NewNode(block)
		} else {
			dfsAdd(root.right, block)
		}
	}
}

func dfsLookupBySizeClosure(root *Node, requirement func(uint) bool) *Node {
	if root == nil { return nil }

	if requirement(root.block.size) {
		return root
	}

	left := dfsLookupBySizeClosure(root.left, requirement)
	if left != nil { return left }

	return dfsLookupBySizeClosure(root.right, requirement)
}

func dfsLookupByOffset(root *Node, target uint) *Node {
	if root == nil { return nil }

	if root.block.offset == target { return root }
	left := dfsLookupByOffset(root.left, target)
	if left != nil {
		return left
	}
	return dfsLookupByOffset(root.right, target)
}

func dfsSum(root *Node) uint {
	if root == nil { return 0 }
	return root.block.size + dfsSum(root.left) + dfsSum(root.right)
}

// Don't export, eventually
func DfsInOrderFreeHeader(root *Node, previous *Node, ch chan<- FreeHeader) *Node {
	if root == nil { return nil }

	res := DfsInOrderFreeHeader(root.left, previous, ch) // provide previous because we haven't visited this node yet
	// Visit self. If visit to left child returned a node,, use that as the previous value.  Otherwise, use the node passed in
	if res != nil {
		fmt.Printf("Write [%d, %d] at offset %d\n", res.block.size, root.block.offset, res.block.offset)
		ch <- FreeHeader{Offset: res.block.offset, Size: res.block.size, NextOffset: root.block.offset}
	} else if previous != nil {
		fmt.Printf("Write [%d, %d] at offset %d\n", previous.block.size, root.block.offset, previous.block.offset)
		ch <- FreeHeader{Offset: previous.block.offset, Size: previous.block.size, NextOffset: root.block.offset}
	}

	// if a node bubbled up from visiting the right subtree continue to bubble it. Otherwise, provide
	// self as the previous node
	res = DfsInOrderFreeHeader(root.right, root, ch)
	if res != nil {
		return res
	} else {
		return root
	}
}

