package executor

import "qsydev.com/term/internal/tree"

type step struct {
	*Step
	tree    tree.Node
	touched []bool
}

func newStep(s *Step) *step {
	return &step{
		Step:    s,
		tree:    tree.Parse(s.GetExpression()),
		touched: make([]bool, len(s.NodeConfigs)),
	}
}

// Done checks with the step expression if this step is done.
func (s *step) done(nodeID uint32) bool {
	s.touched[nodeID] = true
	return s.tree.Eval(s.touched)
}

// nodeColor returns the color of nodeID. If nodeID is not in
// nodeConfigs then it Color_NO_COLOR.
func (s *step) nodeColor(nodeID uint32) Color {
	for _, nc := range s.NodeConfigs {
		if nc.GetId() == nodeID {
			return nc.GetColor()
		}
	}
	return Color_NO_COLOR
}
