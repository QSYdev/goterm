package terminal

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/qsydev/goterm/internal/ble"
	"github.com/qsydev/goterm/internal/ble/fragmenter"
	"github.com/qsydev/goterm/internal/executor"
	"github.com/qsydev/goterm/pkg/qsy"
)

const (
	inf   = "wlan0"
	laddr = "10.0.0.1"
)

var (
	// ErrExecutorRunning is the error returned when
	// a new executor wants to be created but it is already
	// running
	ErrExecutorRunning = errors.New("executor is arleady running")
	// ErrUnsupportedCommand is the error returned when the command
	// sent by BLE is not known.
	ErrUnsupportedCommand = errors.New("executor command is not supported")
)

type nodeEvent struct {
	id   uint32
	lost bool
}

// T is the terminal that puts together all the
// modules of this app.
type T struct {
	ctx context.Context

	device *ble.Device
	server *qsy.Server

	mu        sync.RWMutex
	executing bool
	executor  executor.E
	events    chan []byte

	nodesChan   chan nodeEvent
	packetsChan chan qsy.Packet
}

// Run initializes all the modules and starts listening
// for all events. Run is a blocking function.
func (t *T) Run(ctx context.Context) error {
	var err error
	t.device, err = ble.Init(t, fragmenter.New(t))
	if err != nil {
		return errors.Wrap(err, "failed to initialize BLE device")
	}
	t.server, err = qsy.NewServer(ctx, os.Stdout, inf, laddr, t)
	if err != nil {
		return errors.Wrap(err, "failed to create QSY server")
	}
	if err = t.server.ListenAndAccept(); err != nil {
		return errors.Wrap(err, "failed to start QSY server")
	}
	t.events = make(chan []byte)
	t.nodesChan = make(chan nodeEvent)
	t.ctx = ctx
	return t.run()
}

func (t *T) run() error {
	for {
		select {
		case ne := <-t.nodesChan:
			t.handleNodeEvent(ne)
		case pkt := <-t.packetsChan:
			t.handlePacket(pkt)
		case <-t.ctx.Done():
			t.mu.Lock()
			if t.executing {
				t.executing = false
				t.mu.Unlock()
				t.executor.Stop()
			}
			t.mu.Unlock()
			if err := t.device.Close(); err != nil {
				log.Printf("failed to close BLE device: %s", err)
			}
			close(t.nodesChan)
			close(t.packetsChan)
			close(t.events)
			return nil
		}
	}
}

func (t *T) exec() {
	for event := range t.executor.Events() {
		b, err := proto.Marshal(&event)
		if err != nil {
			continue
		}
		select {
		case t.events <- b:
		case <-t.ctx.Done():
			return
		}
	}
}

// TODO: when BLE service is implemented we'll need to
// send the event.
func (t *T) handleNodeEvent(event nodeEvent) {
}

func (t *T) handlePacket(pkt qsy.Packet) {
	t.mu.Lock()
	if !t.executing {
		t.mu.Unlock()
		return
	}
	t.mu.Unlock()
	t.executor.Touche(uint32(pkt.Step), uint32(pkt.ID), pkt.Delay)
}

// ConnState implements the ble.ConnListener interface.
func (t *T) ConnState(state ble.State) {
}

// Write implements the fragmenter.Client interface.
func (t *T) Write(data []byte) error {
	if data[0] == executor.StopExecID {
		t.mu.Lock()
		if t.executing {
			t.executing = false
			t.mu.Unlock()
			t.executor.Stop()
		}
		t.mu.Unlock()
		return nil
	}
	if data[0] != executor.CustomExecID && data[0] != executor.RandomExecID {
		return ErrUnsupportedCommand
	}
	t.mu.Lock()
	if t.executing {
		t.mu.Unlock()
		return ErrExecutorRunning
	}
	t.mu.Unlock()
	if data[0] == executor.CustomExecID {
		c := &executor.Custom{}
		if err := proto.Unmarshal(data[1:], c); err != nil {
			return errors.Wrap(err, "failed to unmarshal bytes")
		}
		t.executor = c
	} else {
		r := &executor.Random{}
		if err := proto.Unmarshal(data[1:], r); err != nil {
			return errors.Wrap(err, "failed to unmarshal bytes")
		}
		t.executor = r
	}
	t.mu.Lock()
	t.executing = true
	t.mu.Unlock()
	t.executor.Start(t)
	go t.exec()
	return nil
}

// Notify implements the fragmenter.Client interface.
func (t *T) Notify() <-chan []byte {
	return t.events
}

// Receive implements the receive method of qsy.Listener.
func (t *T) Receive(pkt qsy.Packet) {
	select {
	case t.packetsChan <- pkt:
	case <-t.ctx.Done():
		return
	}
}

// LostNode implements the LostNode method of qsy.Listener.
func (t *T) LostNode(id uint16) {
	select {
	case t.nodesChan <- nodeEvent{id: uint32(id), lost: true}:
	case <-t.ctx.Done():
		return
	}
}

// NewNode implements the receive NewNode of qsy.Listener.
func (t *T) NewNode(id uint16) {
	select {
	case t.nodesChan <- nodeEvent{id: uint32(id), lost: false}:
	case <-t.ctx.Done():
		return
	}
}

// Send implements the send method of executor.Sender.
func (t *T) Send(stepID uint32, nc executor.NodeConfig) {
	select {
	case <-t.ctx.Done():
		return
	default:
		if err := t.server.Send(
			qsy.NewPacket(qsy.ToucheT, uint16(nc.GetId()), parseColor(nc.GetColor()),
				nc.GetDelay(), uint16(stepID), false, false)); err != nil {
			// TODO: what happens when we can't send
		}
	}
}

// parseColor parses the executor.Color to a qsy.Color.
func parseColor(color executor.Color) qsy.Color {
	switch color {
	case executor.Color_RED:
		return qsy.Red
	case executor.Color_GREEN:
		return qsy.Green
	case executor.Color_BLUE:
		return qsy.Blue
	case executor.Color_CYAN:
		return qsy.Cyan
	case executor.Color_MAGENTA:
		return qsy.Magenta
	case executor.Color_YELLOW:
		return qsy.Yellow
	case executor.Color_WHITE:
		return qsy.White
	default:
		return qsy.NoColor
	}
}
