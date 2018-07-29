package executor

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
)

type empty struct{}

// Random wraps a RandomExecutor with the functionality
// necessary to execute.
type Random struct {
	*executor
	*RandomExecutor
}

// Start starts the executor using sender to send actions.
func (r *Random) Start(sender Sender) error {
	if r.RandomExecutor == nil {
		return errors.New("invalid configuration")
	}
	r.executor = &executor{
		events:        make(chan Event, 30),
		sender:        sender,
		stopOnTimeout: r.GetStopOnTimeout(),
		duration:      time.Duration(r.GetDuration()) * time.Millisecond,
		getNextStep:   r.generateNextStep,
		steps:         r.GetSteps(),
	}
	r.start()
	return nil
}

// Touche registers a touche with the specified node config.
// If the nodeConfig is not part of the current step then
// it's a nop.
func (r *Random) Touche(stepID, nodeID, delay uint32) {
	r.touche(stepID, nodeID, delay)
}

// generateNextStep generates a new random step.
func (r *Random) generateNextStep() *Step {
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
			Id:    uint32(n),
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
	return &Step{
		NodeConfigs:   nodeConfigs,
		Expression:    exp,
		Timeout:       r.RandomExecutor.Timeout,
		StopOnTimeout: r.RandomExecutor.StopOnTimeout,
	}
}
