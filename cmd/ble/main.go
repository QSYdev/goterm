package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/qsydev/goterm/pkg/idk"
	"github.com/qsydev/goterm/pkg/qsy"
)

var srv *qsy.Server
var ctx, cancel = context.WithCancel(context.Background())

type r struct{}

func (r r) Receive(p qsy.Packet) {
	if p.T == qsy.KeepAliveT {
		log.Printf("keep alive node: %v", p.ID)
	}
}

func (r r) LostNode(id uint16) {
	log.Printf("lost node: %v", id)
}

func (r r) NewNode(id uint16) {
	log.Printf("new node: %v", id)
}

func (r r) NewPlayerExecutor(p *idk.PlayerExecutor) {
	log.Printf("Player executor: %s", p.String())
}

func (r r) NewCustomExecutor(c *idk.CustomExecutor) {
	log.Printf("Custom executor: %s", c.String())
}

func (r r) NotifyStep() {
}

func (r r) NotifyDone() <-chan *idk.Result {
	c := make(chan *idk.Result, 1)
	e := []*idk.Event{
		&idk.Event{
			Type:  idk.Event_Start,
			Delay: int64(0),
			Step:  int32(0),
			Node:  int32(0),
		},
		&idk.Event{
			Type:  idk.Event_Touche,
			Delay: int64(1000),
			Step:  int32(1),
			Node:  int32(1),
		},
		&idk.Event{
			Type:  idk.Event_Touche,
			Delay: int64(500),
			Step:  int32(2),
			Node:  int32(1),
		},
		&idk.Event{
			Type:  idk.Event_StepTimeout,
			Delay: int64(1001),
			Step:  int32(2),
			Node:  int32(1),
		},
		&idk.Event{
			Type:  idk.Event_End,
			Delay: int64(0),
			Step:  int32(2),
			Node:  int32(0),
		},
	}
	res := &idk.Result{
		Events:   e,
		Steps:    3,
		Duration: int64(10000),
	}
	c <- res
	return c
}

func main() {
	var err error
	client := r{}
	srv, err = qsy.NewServer(ctx, os.Stdout, "wlan0", net.IP{224, 0, 0, 12}, "", "10.0.0.1", client)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}
	if err = srv.ListenAndAccept(); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
	d, err := gatt.NewDevice(option.DefaultServerOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	// Register optional handlers.
	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { log.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { log.Println("Disconnect: ", c.ID()) }),
	)

	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			s := idk.NewService(client)
			d.AddService(s)
			uuids := []gatt.UUID{s.UUID()}
			d.AdvertiseNameAndServices("terminal", uuids)
		default:
		}
	}
	d.Init(onStateChanged)
	<-ctx.Done()
	log.Printf("exiting")
}
