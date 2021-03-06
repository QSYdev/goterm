package executor

import (
	"math/rand"
	"strconv"
	"time"
)

// Random wraps a RandomExecutor with the functionality
// necessary to execute.
type Random struct {
	*executor
	*RandomExecutor
}

// Start starts the executor using sender to send actions.
func (r *Random) Start(sender Sender) error {
	if r.RandomExecutor == nil {
		return ErrInvalidExecutor
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

// generateNextStep generates a new random step.
func (r *Random) generateNextStep() *step {
	nodeConfigs := []*NodeConfig{}
	exp := ""
	nodes := rand.Perm(int(r.RandomExecutor.Nodes))
	for i, c := range r.RandomExecutor.Colors {
		nodeConfigs = append(nodeConfigs, &NodeConfig{
			Id:    uint32(nodes[i]),
			Delay: r.RandomExecutor.Delay,
			Color: c,
		})
		exp += strconv.Itoa(nodes[i]) + "&"
	}
	return newStep(&Step{
		NodeConfigs:   nodeConfigs,
		Expression:    exp[:len(exp)-1],
		Timeout:       r.RandomExecutor.Timeout,
		StopOnTimeout: r.RandomExecutor.StopOnTimeout,
	})
}
