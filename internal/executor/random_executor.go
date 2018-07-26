package executor

import (
	"math/rand"
	"strconv"
	"time"
)

type empty struct{}

// Random wraps a RandomExecutor with the functionality
// necessary to execute.
type Random struct {
	executor
	*RandomExecutor
	sender Sender
}

// NewRandom returns a Random that wraps RandomExecutor.
func NewRandom(re *RandomExecutor, sender Sender) *Random {
	return &Random{
		RandomExecutor: re,
		executor: executor{
			events: make(chan Event, 30),
		},
		sender: sender,
	}
}

// Touche registers a touche with the specified node config.
// If the nodeConfig is not part of the current step then
// it's a nop.
func (r *Random) Touche(node NodeConfig) error {
	done, err := r.step.Done(node)
	if err != nil {
		return ErrNotExist
	}
	r.toucheEvent(node)
	if !done {
		return nil
	}
	r.nextStep()
	return nil
}

func (r *Random) nextStep() {
	r.currentTimeout.Stop()
	if r.stepID == r.RandomExecutor.Steps {
		r.routineEndEvent(r.RandomExecutor.Steps)
		return
	}
	r.generateNextStep()
	for _, nc := range r.step.NodeConfigs {
		// TODO: maybe goroutine?
		r.sender.Send(*nc)
	}
	r.currentTimeout = time.AfterFunc(time.Duration(r.RandomExecutor.Timeout)*time.Millisecond, r.stepTimeout)
}

func (r *Random) generateNextStep() {
	t := int(r.RandomExecutor.Nodes)
	nodes := make([]bool, t)
	colors := make(map[Color]bool)
	for _, v := range r.RandomExecutor.Colors {
		colors[v] = false
	}
	nodeConfigs := []*NodeConfig{}
	exp := ""
	for i := 0; i < t; i++ {
		n := rand.Intn(i - 1)
		if nodes[n] {
			continue
		}
		nodes[n] = true
		nc := &NodeConfig{
			Id:    int32(n),
			Delay: r.RandomExecutor.Delay,
		}
		for k, v := range colors {
			if !v {
				continue
			}
			colors[k] = true
			nc.Color = k
		}
		nodeConfigs = append(nodeConfigs, nc)
		exp = exp + strconv.Itoa(n)
		if i < t-1 {
			exp = exp + "&"
		}
	}
	r.step = &Step{
		NodeConfigs:   nodeConfigs,
		Expression:    exp,
		Timeout:       r.RandomExecutor.Timeout,
		StopOnTimeout: r.RandomExecutor.StopOnTimeout,
	}
	r.stepID++
}

func (r *Random) stepTimeout() {
	if _, ok := <-r.currentTimeout.C; !ok {
		return
	}
	// stop everything
	r.stepTimeoutEvent()
	for _, v := range r.step.NodeConfigs {
		r.sender.Send(NodeConfig{Id: v.GetId(), Color: Color_NO_COLOR})
	}
	if r.RandomExecutor.StopOnTimeout {
		r.done = true
		r.routineTimeoutEvent()
		return
	}
	r.nextStep()
}
