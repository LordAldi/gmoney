// pkg/allocate/tree_test.go
package allocate

import (
	"testing"

	"github.com/LordAldi/gmoney/pkg/money"
)

func TestAllocateTree_ComplexHierarchy(t *testing.T) {
	//    ROOT ($0.10)
	//    /     |     \
	//   A(1)  B(1)  C(1)  <- Weights
	//  / | \
	// x  y  z (Weights 1, 1, 1)

	// Construct the tree
	deptA := NewNode("Dept A", 1,
		NewNode("Team X", 1),
		NewNode("Team Y", 1),
		NewNode("Team Z", 1),
	)
	deptB := NewNode("Dept B", 1)
	deptC := NewNode("Dept C", 1)

	root := NewNode("CEO Fund", 1, deptA, deptB, deptC)
	total := money.New(10, "USD") // 10 cents

	// Run Allocation
	err := AllocateTree(root, total)
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}

	// --- VERIFICATION ---

	// 1. Root Integrity
	if root.Allocated.Amount() != 10 {
		t.Errorf("Root lost money! Got %d", root.Allocated.Amount())
	}

	// 2. First Level Split (10 split 3 ways -> 4, 3, 3)
	// Dept A should have 4 because it was the first child (stability in our split logic)
	if deptA.Allocated.Amount() != 4 {
		t.Errorf("Dept A expected 4, got %d", deptA.Allocated.Amount())
	}

	// 3. Second Level Split (4 split 3 ways -> 2, 1, 1)
	x := deptA.Children[0]
	y := deptA.Children[1]
	z := deptA.Children[2]

	if x.Allocated.Amount() != 2 {
		t.Errorf("Team X expected 2, got %d", x.Allocated.Amount())
	}
	if y.Allocated.Amount() != 1 {
		t.Errorf("Team Y expected 1, got %d", y.Allocated.Amount())
	}

	// 4. Conservation of Money Check
	sumChildrenA := x.Allocated.Amount() + y.Allocated.Amount() + z.Allocated.Amount()
	if sumChildrenA != deptA.Allocated.Amount() {
		t.Errorf("Dept A Leak! Parent has %d, Children sum to %d", deptA.Allocated.Amount(), sumChildrenA)
	}

	// Visual check for debugging
	root.PrintTree(0)
}
