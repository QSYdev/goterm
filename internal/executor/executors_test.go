package executor

import "testing"

type s struct{}

func (s) Send(config NodeConfig) {
	// do nothing
}

func TestRandomTouche(t *testing.T) {
	e := NewRandom(&RandomExecutor{
		Colors:            []Color{Color_BLUE, Color_RED},
		Timeout:           int64(3000),
		Delay:             int64(500),
		Duration:          int64(10000),
		Steps:             5,
		Nodes:             2,
		StopOnTimeout:     false,
		WaitForAllPlayers: false,
	}, s{})
	e.events = make(chan Event, 1)
	e.step = &Step{
		NodeConfigs: []*NodeConfig{&NodeConfig{Id: 1, Delay: int64(500), Color: Color_BLUE}},
	}
	if err := e.Touche(*e.step.NodeConfigs[0]); err != nil {
		t.Fatalf("touche of existing node config should succeed")
	}
	event := <-e.events
	if _, ok := <-e.events; ok {
		t.Fatalf("there should only be one event")
	}
	if event.GetType() != Event_Touche {
		t.Fatalf("expected event to be touche, got %s", event.GetType())
	}
	if err := e.Touche(NodeConfig{Id: 2, Delay: int64(500), Color: Color_CYAN}); err == nil {
		t.Fatalf("touche of non existing node config should not succeed")
	}
}
