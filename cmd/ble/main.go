package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	proto "github.com/golang/protobuf/proto"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"qsydev.com/term/internal/ble/fragmenter"
	"qsydev.com/term/internal/executor"
	"qsydev.com/term/pkg/qsy"
)

var srv *qsy.Server
var ctx, cancel = context.WithCancel(context.Background())

type r struct {
	executor executor.E
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
	t := data[0]
	switch t {
	case executor.CustomExecID:
		ce := &executor.Custom{}
		if err := proto.Unmarshal(data[1:], ce); err != nil {
			log.Printf("failed decoding custom : %s", err)
			return errors.New("failed to decode custom executor")
		}
		fmt.Println(ce)
		return nil
	case executor.RandomExecID:
		re := &executor.Random{}
		if err := proto.Unmarshal(data, re); err != nil {
			log.Printf("failed decoding random executor: %s", err)
			return errors.New("failed to decode random executor")
		}
		fmt.Println(re)
		return nil
	case executor.StopExecID:
	}
	return nil
}

func (r r) Notify() <-chan []byte {
	c := make(chan []byte, 1)
	return c
}

func main() {
	var err error
	client := r{}
	srv, err = qsy.NewServer(ctx, os.Stdout, "wlan0", "10.0.0.1", client)
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
			svc := fragmenter.New(client)
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
