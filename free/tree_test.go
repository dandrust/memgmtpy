package free

import "testing"

func buildNode(offset uint, size uint) *Node {
	block := Block{offset: offset, size: size}
	return NewNode(&block) 
}

// Add this to prod code?
func dfsNodeCount(root *Node) int {
	if root == nil { return 0 }

	return 1 + dfsNodeCount(root.left) + dfsNodeCount(root.right)
}

func assertInOrder(root *Node) bool {
	actual := []uint{}
	dfsInOrder(root, &actual)

	for i := 1; i < len(actual); i++ {
		if actual[i - 1] > actual[i] { return false }
	}

	return true
}

// Add this to prod code with closure?
func dfsInOrder(root *Node, ary *[]uint) {
	if root == nil { return }

	dfsInOrder(root.left, ary)
	*ary = append(*ary, root.block.offset)
	dfsInOrder(root.right, ary)
}

func TestNewNode(t *testing.T) {
	block := Block{offset: 1, size: 1}
	node := NewNode(&block) 

	if node.block != &block {
		t.Error("NewNode()")
	}
}

func TestFirstWithNullInupt(t *testing.T) {
	output := first(nil)

	if output != nil {
		t.Error("first() with null input")
	}
}

func TestFirstWithNullLeftChild(t *testing.T) {
	node := buildNode(1, 1)

	output := first(node)

	if output != node {
		t.Error("first() with null left child")
	}
}
func TestFirstWithLeftChild(t *testing.T) {
	root := buildNode(5, 1)
	left_child := buildNode(1, 1)
	root.left = left_child

	output := first(root)

	if output != left_child {
		t.Error("first() with left child")
	}
}

func TestFirstWithDeepLeftDescendant(t *testing.T) {
	root := buildNode(5, 1)
	current := root

	for i := 0; i < 15; i++ {
		current.left = buildNode(1, 1)
		current = current.left	
	}

	output := first(root)

	if output != current {
		t.Error("first() with deep left descendant")
	}
}

func TestLastWithNullInupt(t *testing.T) {
	output := last(nil)

	if output != nil {
		t.Error("last() with null input")
	}
}

func TestLastWithNullLeftChild(t *testing.T) {
	node := buildNode(1, 1)

	output := last(node)

	if output != node {
		t.Error("last() with null right child")
	}
}
func TestLastWithLeftChild(t *testing.T) {
	root := buildNode(5, 1)
	right_child := buildNode(1, 1)
	root.right = right_child

	output := last(root)

	if output != right_child {
		t.Error("last() with right child")
	}
}

func TestLastWithDeepRightDescendant(t *testing.T) {
	root := buildNode(5, 1)
	current := root

	for i := 0; i < 15; i++ {
		current.right = buildNode(1, 1)
		current = current.right
	}

	output := last(root)

	if output != current {
		t.Error("last() with deep right descendant")
	}
}

func TestDfsAddWithContiguousLeftBlock(t *testing.T) {
	// |-------| <- expected size of a
	// + - + - + - + - + - + - + - + - + - +
	// |   | a |   | b |   | c |   |   |   |
	// + - + - + - + - + - + - + - + - + - +
	// 0   1   2   3   4   5   6   7   8   9
	//   ^ Add this block

	a := buildNode(1, 1)
	b := buildNode(3, 1)
	c := buildNode(5, 1)
	node_count := 3

	// Tree:
	// | |  b
	// | |a   c
	// | |
	//  ^ Add this block 

	b.left = a
	b.right = c

	root := b
	newBlock := Block{offset: 0, size: 1}

	dfsAdd(root, &newBlock)

	// it grows the existing block by increasing size and decreasing offset
	if a.block.offset != 0 && a.block.size != 2 {
		t.Error("dfsAdd() with contiguous left block: expected to grow existing node")
	}

	// it doesn't increase the number of nodes
	if dfsNodeCount(root) != node_count {
		t.Error("dfsAdd() with contiguous left block: expected to maintain node count")
	}	
}

func TestDfsAddWithContiguousLeftBlockAndContiguousDescendant(t *testing.T) {
	//                 |-----------| <- expected size of d
	// + - + - + - + - + - + - + - + - + - +   
	// | a |   | b |   | c |   | d |   | e |   
	// + - + - + - + - + - + - + - + - + - +   
	// 0   1   2   3   4   5   6   7   8   9   
	//                       ^ Add this block

	a := buildNode(0, 1)
	b := buildNode(2, 1)
	c := buildNode(4, 1)
	d := buildNode(6, 1)
	e := buildNode(8, 1)
	node_count := 5

	// Tree:
	//      | |d
	//   b  | |  e
	// a   c| |
	//       ^ Add this block 

	d.left = b
	d.right = e
	b.left = a
	b.right = c
	
	root := d
	newBlock := Block{offset: 5, size: 1}

	dfsAdd(root, &newBlock)

	// it grows the existing block by BOTH the added entry and the contiguous descendant
	if root.block.offset != 4 && root.block.size != 3 {
		t.Error("dfsAdd() with contiguous left block and contiguous descentant: expected to grow existing node")
	}

	// it descreased the number of nodes by 1
	if dfsNodeCount(root) != node_count - 1 {
		t.Error("dfsAdd() with contiguous left block and contiguous descentant: expected to remove contiguous descentant")
	}
}
func TestDfsAddWithContiguousLeftBlockAndContiguousDescendantWithChild(t *testing.T) {
	//                         |-----------| <- expected size of e
	// + - + - + - + - + - + - + - + - + - + - + - +  
	// | a |   | b |   | c |   | d |   | e |   | f |  
	// + - + - + - + - + - + - + - + - + - + - + - +  
	// 0   1   2   3   4   5   6   7   8   9   10     
	//                               ^ Add this block         

	a := buildNode(0, 1)
	b := buildNode(2, 1)
	c := buildNode(4, 1)
	d := buildNode(6, 1)
	e := buildNode(8, 1)
	f := buildNode(10, 1)
	node_count := 6

	// Tree:
	//        | |e
	//   b    | |  f
	// a     d| |
	//     c  | |
	//         ^ Add this block 

	e.left = b
	e.right = f
	b.left = a
	b.right = d
	d.left = c
	
	root := e
	newBlock := Block{offset: 7, size: 1}

	dfsAdd(root, &newBlock)

	// it grows the existing block by BOTH the added entry and the contiguous descendant
	if root.block.offset != 4 && root.block.size != 3 {
		t.Error("dfsAdd() with contiguous left block and contiguous descentant with child: expected to grow existing node")
	}

	// it descreased the number of nodes by 1
	if dfsNodeCount(root) != node_count - 1 {
		t.Error("dfsAdd() with contiguous left block and contiguous descentant with child: expected to remove contiguous descentant")
	}

	// it maintains bst ordering
	if !assertInOrder(root) {
		t.Error("dfsAdd() with contiguous left block and contiguous descentant with child: expected to maintain ordering")
	}

}

func TestDfsAddWithContiguousRightBlock(t *testing.T) {
	//                     |-------| <- expected size of c
	// + - + - + - + - + - + - + - + - + - +
	// |   | a |   | b |   | c |   |   |   |
	// + - + - + - + - + - + - + - + - + - +
	// 0   1   2   3   4   5   6   7   8   9
	//   ^ Add this block

	a := buildNode(1, 1)
	b := buildNode(3, 1)
	c := buildNode(5, 1)
	node_count := 3

	// Tree:
	//   b  | |
	// a   c| |
	//      | |
	//       ^ Add this block 

	b.left = a
	b.right = c

	root := b
	newBlock := Block{offset: 6, size: 1}

	dfsAdd(root, &newBlock)

	// it grows the existing block by increasing size and decreasing offset
	if c.block.offset != 5 && c.block.size != 2 {
		t.Error("dfsAdd() with contiguous right block: expected to grow existing node")
	}

	// it doesn't increase the number of nodes
	if dfsNodeCount(root) != node_count {
		t.Error("dfsAdd() with contiguous right block: expected to maintain node count")
	}	
}

func TestDfsAddWithContiguousRightBlockAndContiguousDescendant(t *testing.T) {
	//         |-----------| <- expected size of b
	// + - + - + - + - + - + - + - + - + - +   
	// | a |   | b |   | c |   | d |   | e |   
	// + - + - + - + - + - + - + - + - + - +   
	// 0   1   2   3   4   5   6   7   8   9   
	//               ^ Add this block

	a := buildNode(0, 1)
	b := buildNode(2, 1)
	c := buildNode(4, 1)
	d := buildNode(6, 1)
	e := buildNode(8, 1)
	node_count := 5

	// Tree:
	//   b| |
	// a  | |  d
	//    | |c   e
	//     ^ Add this block 

	b.left = a
	b.right = d
	d.left = c
	d.right = e
	
	root := b
	newBlock := Block{offset: 3, size: 1}

	dfsAdd(root, &newBlock)

	// it grows the existing block by BOTH the added entry and the contiguous descendant
	if root.block.offset != 2 && root.block.size != 3 {
		t.Error("dfsAdd() with contiguous right block and contiguous descentant: expected to grow existing node")
	}

	// it descreased the number of nodes by 1
	if dfsNodeCount(root) != node_count - 1 {
		t.Error("dfsAdd() with contiguous right block and contiguous descentant: expected to remove contiguous descentant")
	}
}

func TestDfsAddWithContiguousRightBlockAndContiguousDescendantWithChild(t *testing.T) {
	//         |-----------| <- expected size of b
	// + - + - + - + - + - + - + - + - + - + - + - +  
	// | a |   | b |   | c |   | d |   | e |   | f |  
	// + - + - + - + - + - + - + - + - + - + - + - +  
	// 0   1   2   3   4   5   6   7   8   9   10     
	//               ^ Add this block         

	a := buildNode(0, 1)
	b := buildNode(2, 1)
	c := buildNode(4, 1)
	d := buildNode(6, 1)
	e := buildNode(8, 1)
	f := buildNode(10, 1)
	node_count := 6

	// Tree:
	//   b| |
	// a  | |    e
	//    | |c     f
	//         d
	//     ^ Add this block 

	b.left = a
	b.right = e
	e.left = c
	e.right = f
	c.right = d
	
	root := b
	newBlock := Block{offset: 3, size: 1}

	dfsAdd(root, &newBlock)

	// it grows the existing block by BOTH the added entry and the contiguous descendant
	if root.block.offset != 2 && root.block.size != 3 {
		t.Error("dfsAdd() with contiguous right block and contiguous descentant with child: expected to grow existing node")
	}

	// it descreased the number of nodes by 1
	if dfsNodeCount(root) != node_count - 1 {
		t.Error("dfsAdd() with contiguous right block and contiguous descentant with child: expected to remove contiguous descentant")
	}

	// it maintains bst ordering
	if !assertInOrder(root) {
		t.Error("dfsAdd() with contiguous right block and contiguous descentant with child: expected to maintain ordering")
	}

}