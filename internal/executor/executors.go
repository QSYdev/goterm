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
	Touche(stepID, nodeID, delay int32)
	Stop() error
	Start(sender Sender) error
	Events() <-chan Event
}

type executor struct {
	sender Sender
	done   bool
	events chan Event

	stepTimer *time.Timer
	step      *step
	stepID    uint32
	steps     uint32

	routineTimer  *time.Timer
	duration      time.Duration
	stopOnTimeout bool
	getNextStep   func() *step
}

func (e *executor) start() {
	// TODO: routineTimeout if necessary
	if e.duration != 0 {
		e.routineTimer = time.AfterFunc(e.duration, e.routineTimeout)
	}
	e.sendStep()
}

func (e *executor) touche(stepID, nodeID, delay uint32) {
	if e.done || stepID != e.stepID {
		return
	}
	if done := e.step.done(nodeID); !done {
		return
	}
	e.nextStep()
}

func (e *executor) nextStep() {
	if e.stepTimer != nil {
		e.stepTimer.Stop()
	}
	if e.stepID == e.steps {
		e.done = true
		if e.routineTimer != nil {
			e.routineTimer.Stop()
		}
		e.routineEndEvent()
		return
	}
	e.sendStep()
}

func (e *executor) sendStep() {
	e.stepID++
	e.step = e.getNextStep()
	for _, nc := range e.step.NodeConfigs {
		e.sender.Send(*nc)
	}
	if e.step.GetTimeout() != 0 {
		e.stepTimer = time.AfterFunc(time.Duration(e.step.GetTimeout())*time.Millisecond, e.stepTimeout)
	}
}

// stepTimeout listens on the currentTimeout channel, if the
// step has timeout then it stops current step.
func (e *executor) stepTimeout() {
	e.stepTimeoutEvent()
	for _, nc := range e.step.NodeConfigs {
		e.sender.Send(NodeConfig{Id: nc.GetId(), Color: Color_NO_COLOR})
	}
	if e.stopOnTimeout {
		e.done = true
		e.routineEndEvent()
		return
	}
	e.nextStep()
}

func (e *executor) routineTimeout() {
	// TODO
}

func (e *executor) routineTimeoutEvent() {
	e.events <- Event{
		Type: Event_RoutineTimeout,
		Step: e.stepID,
	}
}

func (e *executor) stepTimeoutEvent() {
	e.events <- Event{
		Type: Event_StepTimeout,
		Step: e.stepID,
	}
}

func (e *executor) toucheEvent(nodeID, delay uint32) {
	e.events <- Event{
		Type:  Event_Touche,
		Color: e.step.nodeColor(nodeID),
		Delay: delay,
		Step:  e.stepID,
		Node:  nodeID,
	}
}

func (e *executor) routineEndEvent() {
	e.events <- Event{
		Type: Event_End,
		Step: e.steps,
	}
}
