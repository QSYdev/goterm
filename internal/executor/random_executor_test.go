package executor

import (
	"testing"
)

func TestGenerateNextStep(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		re             *Random
		nNodeConfigs   int
		expectedColors []Color
	}{
		{name: "repeated colors", re: &Random{RandomExecutor: &RandomExecutor{Colors: []Color{Color_BLUE, Color_BLUE}, Nodes: 4}}, expectedColors: []Color{Color_BLUE, Color_BLUE}},
		{name: "different colors", re: &Random{RandomExecutor: &RandomExecutor{Colors: []Color{Color_BLUE, Color_RED}, Nodes: 5}}, expectedColors: []Color{Color_BLUE, Color_RED}},
	}
	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			s := c.re.generateNextStep()
			if len(s.NodeConfigs) != len(c.expectedColors) {
				tt.Fatalf("expected len of node configs to be 2 but got %d", len(s.NodeConfigs))
			}
			for _, nc := range s.NodeConfigs {
				if !hasColors(*nc, c.expectedColors) {
					tt.Fatalf("node config color is not expected, got %s", nc.Color)
				}
			}

		})
	}
}

func hasColors(nc NodeConfig, colors []Color) bool {
	for _, c := range colors {
		if nc.Color == c {
			return true
		}
	}
	return false
}
