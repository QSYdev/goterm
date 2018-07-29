package tree

// Node has an eval method that returns true depending
// on the visited elements.
type Node interface {
	Eval(visited []bool) bool
}

// And is a node that implements the and binary expression.
type And struct {
	Value       int
	Left, Right Node
}

// Eval implements the Node interface for and.
func (and And) Eval(visited []bool) bool {
	if and.Left == nil && and.Right == nil {
		return visited[and.Value]
	}
	return and.Left.Eval(visited) && and.Right.Eval(visited)

}

// Or is a node that implements the or binary expression.
type Or struct {
	Value       int
	Left, Right Node
}

// Eval implements the Node interface for or.
func (or Or) Eval(visited []bool) bool {
	if or.Left == nil && or.Right == nil {
		return visited[or.Value]
	}
	return or.Left.Eval(visited) || or.Right.Eval(visited)
}

// Parse parses the expression and returns the tree.
func Parse(expression string) Node {
	// TODO
	return And{}
}
