package executor

import (
	"testing"
	"time"
)

type s struct {
	r chan uint32
}

func (s *s) Send(config NodeConfig) {
	s.r <- config.GetId()
}

func TestStepTimeout(t *testing.T) {
	t.Parallel()

	schan := make(chan uint32, 1)
	e := &executor{
		sender:        &s{r: schan},
		events:        make(chan Event, 2),
		stopOnTimeout: true,
		step:          newStep(&Step{NodeConfigs: []*NodeConfig{&NodeConfig{Id: 1}}}),
	}
	e.stepTimer = time.AfterFunc(10*time.Millisecond, e.stepTimeout)
	if event := <-e.events; event.GetType() != Event_StepTimeout {
		t.Fatalf("expected step timeout event but got %s", event.GetType())
	}
	if event := <-e.events; event.GetType() != Event_End {
		t.Fatalf("expected routine end event but got %s", event.GetType())
	}
	if nid := <-schan; nid != 1 {
		t.Fatalf("expected node id 1 to be sent but go %d", nid)
	}

	e = &executor{
		sender:        &s{r: schan},
		events:        make(chan Event, 1),
		stopOnTimeout: false,
		step:          newStep(&Step{NodeConfigs: []*NodeConfig{&NodeConfig{Id: 1}}}),
		getNextStep:   func() *step { return newStep(&Step{NodeConfigs: []*NodeConfig{&NodeConfig{Id: 2}}}) },
		steps:         3,
	}
	e.stepTimer = time.AfterFunc(10*time.Millisecond, e.stepTimeout)
	if event := <-e.events; event.GetType() != Event_StepTimeout {
		t.Fatalf("expected step timeout event but got %s", event.GetType())
	}
	if nid := <-schan; nid != 1 {
		t.Fatalf("expected node id 1 to be sent but go %d", nid)
	}
	if nid := <-schan; nid != 2 {
		t.Fatalf("expected node id 2 to be sent but got %d", nid)
	}
}

func TestSendStep(t *testing.T) {
	t.Parallel()

	schan := make(chan uint32, 2)
	e := &executor{
		stepID: 1,
		sender: &s{r: schan},
		getNextStep: func() *step {
			return newStep(&Step{Timeout: 1, NodeConfigs: []*NodeConfig{&NodeConfig{Id: 1}, &NodeConfig{Id: 2}}})
		},
	}
	e.sendStep()
	if nid := <-schan; nid != 1 {
		t.Fatalf("expected node id 1 to be sent but got %d", nid)
	}
	if nid := <-schan; nid != 2 {
		t.Fatalf("expected node id 2 to be sent but got %d", nid)
	}
	if e.stepTimer == nil {
		t.Fatalf("step timer should be set")
	}
}

func TestNextStep(t *testing.T) {
	e := &executor{
		sender: &s{},
		stepID: 1,
		steps:  1,
		events: make(chan Event, 1),
	}
	e.nextStep()
	if !e.done {
		t.Fatalf("expected routine to be done")
	}
	if event := <-e.events; event.GetType() != Event_End {
		t.Fatalf("expected event to be routine end but got %s", event.GetType())
	}
}
