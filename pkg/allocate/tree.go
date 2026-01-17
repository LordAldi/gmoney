package allocate

import (
	"fmt"
	"strings"

	"github.com/LordAldi/gmoney/pkg/money"
)

// Node represents a bucket in the allocation hierarchy.
type Node struct {
	Name     string
	Weight   int
	Children []*Node

	// Allocated is the result value. It is populated after calculation.
	Allocated money.Money
}

// NewNode is a helper to build trees cleanly.
func NewNode(name string, weight int, children ...*Node) *Node {
	return &Node{
		Name:     name,
		Weight:   weight,
		Children: children,
	}
}

// PrintTree helps visualize the distribution (DFS traversal).
func (n *Node) PrintTree(level int) {
	indent := strings.Repeat("  ", level)
	fmt.Printf("%s- %s: %s (Weight: %d)\n", indent, n.Name, n.Allocated, n.Weight)
	for _, child := range n.Children {
		child.PrintTree(level + 1)
	}
}

// AllocateTree distributes the total amount down the hierarchy.
func AllocateTree(root *Node, total money.Money) error {
	// 1. Assign the total to the root (The root holds the full bag)
	root.Allocated = total

	// 2. Begin recursion
	return distribute(root)
}

func distribute(parent *Node) error {
	// Base case: Leaf node (no children to split money with)
	if len(parent.Children) == 0 {
		return nil
	}

	// Step A: Collect weights
	weights := make([]int, len(parent.Children))
	for i, child := range parent.Children {
		weights[i] = child.Weight
	}

	// Step B: Perform the "Penny Perfect" split on the Parent's money
	// This ensures Parent.Allocated == Sum(Children.Allocated)
	amounts, err := Split(parent.Allocated, weights)
	if err != nil {
		return err
	}

	// Step C: Assign results and Recurse
	for i, child := range parent.Children {
		child.Allocated = amounts[i]

		// Recursive call: The child now becomes the parent for its own subtree
		if err := distribute(child); err != nil {
			return err
		}
	}

	return nil
}
