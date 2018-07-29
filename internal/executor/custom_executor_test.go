package executor

import "testing"

func TestCustomGenerateNextStep(t *testing.T) {
	t.Parallel()

	c := &Custom{
		CustomExecutor: &CustomExecutor{
			Steps: []*Step{&Step{NodeConfigs: []*NodeConfig{&NodeConfig{Id: 1, Delay: 500, Color: Color_BLUE}}, Expression: "1"}, &Step{NodeConfigs: []*NodeConfig{&NodeConfig{Id: 2, Delay: 500, Color: Color_BLUE}}, Expression: "2"}},
		},
		executor: &executor{stepID: 1},
	}
	s := c.generateNextStep()
	if s.GetExpression() != "1" {
		t.Fatalf("expected expression to be 1 but got %s", s.GetExpression())
	}
	if len(s.GetNodeConfigs()) != 1 {
		t.Fatalf("expected to have only one node config in step")
	}
	if nc := s.GetNodeConfigs()[0]; nc.Id != 1 || nc.Color != Color_BLUE {
		t.Fatalf("expected id and color to be 1 and Blue, got %d and %s", nc.Id, nc.Color)
	}
	c.stepID++
	s = c.generateNextStep()
	if s.GetExpression() != "2" {
		t.Fatalf("expected expression to be 2 but got %s", s.GetExpression())
	}
	if len(s.GetNodeConfigs()) != 1 {
		t.Fatalf("expected to have only one node config in step")
	}
	if nc := s.GetNodeConfigs()[0]; nc.Id != 2 || nc.Color != Color_BLUE {
		t.Fatalf("expected id and color to be 2 and Blue, got %d and %s", nc.Id, nc.Color)
	}
}
