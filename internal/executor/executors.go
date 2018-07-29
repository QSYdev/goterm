package executor

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrNotExist is the error that happens when the node
	// does not exist in the current step.
	ErrNotExist = errors.New("NodeConfig does not exist in current step")
)

// Sender knows how to send a node config.
type Sender interface {
	Send(stepID uint32, node NodeConfig)
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

	mu        sync.RWMutex
	stepID    uint32
	steps     uint32
	stepTimer *time.Timer
	step      *step

	routineTimer  *time.Timer
	duration      time.Duration
	stopOnTimeout bool
	getNextStep   func() *step
}

func (e *executor) start() {
	if e.duration != 0 {
		e.routineTimer = time.AfterFunc(e.duration, e.routineTimeout)
	}
	e.stepID = 1
	e.sendStep()
}

func (e *executor) touche(stepID, nodeID, delay uint32) {
	e.mu.RLock()
	if e.done || stepID != e.stepID {
		e.mu.RUnlock()
		return
	}
	e.mu.RUnlock()
	if done := e.step.done(nodeID); !done {
		return
	}
	e.mu.Lock()
	e.stepID++
	e.mu.Unlock()
	e.nextStep()
}

func (e *executor) nextStep() {
	if e.stepTimer != nil {
		e.stepTimer.Stop()
	}
	e.mu.RLock()
	if e.stepID == e.steps {
		e.mu.RUnlock()
		e.done = true
		if e.routineTimer != nil {
			e.routineTimer.Stop()
		}
		e.routineEndEvent()
		return
	}
	e.mu.RUnlock()
	e.sendStep()
}

func (e *executor) sendStep() {
	e.step = e.getNextStep()
	e.mu.RLock()
	for _, nc := range e.step.NodeConfigs {
		e.sender.Send(e.stepID, *nc)
	}
	e.mu.RUnlock()
	if e.step.GetTimeout() != 0 {
		e.stepTimer = time.AfterFunc(time.Duration(e.step.GetTimeout())*time.Millisecond, e.stepTimeout)
	}
}

// stepTimeout listens on the currentTimeout channel, if the
// step has timeout then it stops current step.
func (e *executor) stepTimeout() {
	e.mu.Lock()
	if e.stepID < e.steps {
		e.stepID++
	}
	e.mu.Unlock()
	e.stepTimeoutEvent()
	e.cancelStep()
	if e.stopOnTimeout {
		e.done = true
		e.routineEndEvent()
		return
	}
	e.nextStep()
}

func (e *executor) cancelStep() {
	for _, nc := range e.step.NodeConfigs {
		e.sender.Send(0, *nc)
	}
}

func (e *executor) routineTimeout() {
	if e.stepTimer != nil {
		e.stepTimer.Stop()
	}
	e.done = true
	e.cancelStep()
	e.routineTimeoutEvent()
}

func (e *executor) routineTimeoutEvent() {
	e.mu.RLock()
	e.events <- Event{
		Type: Event_RoutineTimeout,
		Step: e.stepID,
	}
	e.mu.RUnlock()
}

func (e *executor) stepTimeoutEvent() {
	e.mu.RLock()
	e.events <- Event{
		Type: Event_StepTimeout,
		Step: e.stepID,
	}
	e.mu.RUnlock()
}

func (e *executor) toucheEvent(nodeID, delay uint32) {
	e.mu.RLock()
	e.events <- Event{
		Type:  Event_Touche,
		Color: e.step.nodeColor(nodeID),
		Delay: delay,
		Step:  e.stepID,
		Node:  nodeID,
	}
	e.mu.RUnlock()
}

func (e *executor) routineEndEvent() {
	e.events <- Event{
		Type: Event_End,
		Step: e.steps,
	}
}
