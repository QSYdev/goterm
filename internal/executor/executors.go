package executor

import (
	"errors"
	"time"
)

var (
	// ErrNotExist is the error that happens when the node
	// does not exist in the current step.
	ErrNotExist = errors.New("NodeConfig does not exist in current step")
)

// Sender knows how to send a node config.
type Sender interface {
	Send(config NodeConfig)
}

// Executor knows how to advance after each touche
// and exposes the events that happen durinig the
// execution.
type Executor interface {
	Touche(node NodeConfig) error
	Stop() error
	Start()
	Events() <-chan Event
}

type executor struct {
	events         chan Event
	done           bool
	step           *Step
	currentTimeout *time.Timer
	stepID         int32
}

func (e executor) routineTimeoutEvent() {
	e.events <- Event{
		Type: Event_RoutineTimeout,
		Step: e.stepID,
	}
}

func (e executor) stepTimeoutEvent() {
	e.events <- Event{
		Type: Event_StepTimeout,
		Step: e.stepID,
	}
}

func (e executor) toucheEvent(node NodeConfig) {
	e.events <- Event{
		Type:  Event_Touche,
		Color: node.GetColor(),
		Delay: node.Delay,
		Step:  e.stepID,
		Node:  node.GetId(),
	}
}

func (e executor) routineEndEvent(steps int32) {
	e.events <- Event{
		Type: Event_End,
		Step: steps,
	}
}

// Done checks with the step expression if this step is done.
func (s *Step) Done(node NodeConfig) (bool, error) {
	return false, nil
}
