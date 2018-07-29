package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	proto "github.com/golang/protobuf/proto"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/qsydev/goterm/internal/ble"
	"github.com/qsydev/goterm/internal/executor"
	"github.com/qsydev/goterm/pkg/qsy"
)

var srv *qsy.Server
var ctx, cancel = context.WithCancel(context.Background())

type r struct {
	executor executor.Executor
}

func (r r) Receive(p qsy.Packet) {
}

func (r r) LostNode(id uint16) {
	log.Printf("lost node: %v", id)
}

func (r r) NewNode(id uint16) {
	log.Printf("new node: %v", id)
}

func (r r) Write(data []byte) error {
	// TODO: data will have the type of packet in the first byte
	// Type could be RandomExecutor, CustomExecutor or StopExecutor
	// this types will be defined later in the executors package.
	// 0x14 is custom executor, move to variable
	// 0x15 is random executor, move to variable
	// 0xFF is stop executor, move to variable
	t := data[0]
	switch t {
	case 0x14:
		// TODO: this will not be used this way, once we have
		// the executor package we defer the creation and unmarshling
		// to that pacakge
		ce := &executor.CustomExecutor{}
		if err := proto.Unmarshal(data[1:], ce); err != nil {
			log.Printf("failed decoding custom : %s", err)
			return errors.New("failed to decode custom executor")
		}
		fmt.Println(ce)
		return nil
	case 0x15:
		// TODO: this will not be used this way, once we have
		// the executor package we defer the creation and unmarshling
		// to that pacakge
		re := &executor.Random{}
		if err := proto.Unmarshal(data, re); err != nil {
			log.Printf("failed decoding random executor: %s", err)
			return errors.New("failed to decode random executor")
		}
		fmt.Println(re)
		return nil
	case 0xFF:
		// TODO: stop executor if running
	}
	return nil
}

func (r r) Notify() <-chan []byte {
	// Sends bytes on the data, the first byte
	// is used to specify if there was an error or if
	// all went well.
	c := make(chan []byte, 1)
	e := []*executor.Event{
		&executor.Event{
			Type:  executor.Event_Start,
			Delay: uint32(0),
			Step:  uint32(0),
			Node:  uint32(0),
		},
		&executor.Event{
			Type:  executor.Event_Touche,
			Delay: uint32(1000),
			Step:  uint32(1),
			Node:  uint32(1),
		},
		&executor.Event{
			Type:  executor.Event_Touche,
			Delay: uint32(500),
			Step:  uint32(2),
			Node:  uint32(1),
		},
		&executor.Event{
			Type:  executor.Event_StepTimeout,
			Delay: uint32(1001),
			Step:  uint32(2),
			Node:  uint32(1),
		},
		&executor.Event{
			Type:  executor.Event_End,
			Delay: uint32(0),
			Step:  uint32(2),
			Node:  uint32(0),
		},
	}
	res := &executor.Result{
		Events:   e,
		Steps:    3,
		Duration: uint32(10000),
	}
	b, err := proto.Marshal(res)
	if err != nil {
		// TODO: 0x01 indicates some error happened
		// extract this to a variable in some package
		b = []byte{0x01}
	}
	// TODO: 0x00 indicates success, should be a variable
	// in some package
	c <- append([]byte{0x00}, b...)
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
		log.Fatalf("failed to open device, err: %s", err)
	}
	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { log.Println("Connect: ", c.ID()) }),
		gatt.CentralDisconnected(func(c gatt.Central) { log.Println("Disconnect: ", c.ID()) }),
	)
	onStateChanged := func(d gatt.Device, s gatt.State) {
		switch s {
		case gatt.StatePoweredOn:
			svc := ble.NewService(client)
			d.AddService(svc)
			uuids := []gatt.UUID{svc.UUID()}
			d.AdvertiseNameAndServices("terminal", uuids)
		default:
		}
	}
	d.Init(onStateChanged)
	<-ctx.Done()
	log.Printf("exiting")
}
