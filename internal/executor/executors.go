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
	// ErrInvalidExecutor is the error returned when the
	// executor is not valid.
	ErrInvalidExecutor = errors.New("Executor can't be nil")

	// CustomExecID is the identifier that identifies a custom
	// executor.
	CustomExecID byte = 0x14
	// RandomExecID is the identifier that identifies a random
	// executor.
	RandomExecID byte = 0x15
	// StopExecID is the identifier that identifies the stop
	// executor operation.
	StopExecID byte = 0xFF
)

const (
	// eventChannelSize is the size of the buffered channel
	// used to send events.
	eventChannelSize = 30
)

// Sender knows how to send a node config.
type Sender interface {
	Send(stepID uint32, node NodeConfig)
}

// E knows how to advance after each touche
// and exposes the events that happen durinig the
// execution.
type E interface {
	Touche(stepID, nodeID, delay uint32)
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

// Stop stops the current execution, if there is no execution
// it returns an error.
func (e *executor) Stop() error {
	if e.done {
		return errors.New("can't stop stopped executor")
	}
	e.done = true
	if e.stepTimer != nil {
		e.stepTimer.Stop()
	}
	if e.routineTimer != nil {
		e.routineTimer.Stop()
	}
	e.cancelStep()
	e.routineEndEvent()
	close(e.events)
	return nil
}

// Events returns the channel were events are sent.
func (e *executor) Events() <-chan Event {
	return e.events
}

// Touche adds the nodeID as touched. If stepID is not the
// same one than the current step then this does nothing.
func (e *executor) Touche(stepID, nodeID, delay uint32) {
	e.touche(stepID, nodeID, delay)
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
		close(e.events)
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
		close(e.events)
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
	close(e.events)
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
