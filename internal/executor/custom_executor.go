package executor

import "time"

// Custom wraps a CustomExecutor with the functionality
// necessary to be executed.
type Custom struct {
	*executor
	*CustomExecutor
}

// Start starts the executor using sender to send commands.
func (c *Custom) Start(sender Sender) error {
	if c.CustomExecutor == nil {
		return ErrInvalidExecutor
	}
	c.executor = &executor{
		events:      make(chan Event, eventChannelSize),
		sender:      sender,
		duration:    time.Duration(c.GetDuration()) * time.Millisecond,
		getNextStep: c.generateNextStep,
		steps:       uint32(len(c.GetSteps())),
	}
	c.start()
	return nil
}

// generateNextSteps returns the next step to be executed.
func (c *Custom) generateNextStep() *step {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return newStep(c.GetSteps()[c.stepID-1])
}
