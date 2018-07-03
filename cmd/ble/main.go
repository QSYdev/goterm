package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
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

func NewTerminalService() *gatt.Service {
	s := gatt.NewService(gatt.MustParseUUID(uuid.New().String()))
	s.AddCharacteristic(gatt.MustParseUUID(uuid.New().String())).HandleReadFunc(
		func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
			go func() {
				t := time.NewTicker(time.Second)
				color := false
				for {
					select {
					case <-t.C:
						var c qsy.Color
						if !color {
							c = qsy.NoColor
						} else {
							c = qsy.Blue
						}
						color = !color
						srv.Send(qsy.NewPacket(qsy.CommandT, uint16(19), c, uint32(0), uint16(0), false, false))
						srv.Send(qsy.NewPacket(qsy.CommandT, uint16(20), c, uint32(0), uint16(0), false, false))
					}
				}
			}()
		})
	s.AddCharacteristic(gatt.MustParseUUID(uuid.New().String())).HandleReadFunc(
		func(rsp gatt.ResponseWriter, req *gatt.ReadRequest) {
			cancel()
		})
	return s
}

func main() {
	var err error
	srv, err = qsy.NewServer(ctx, os.Stdout, "wlan0", net.IP{224, 0, 0, 12}, "", "10.0.0.1", r{})
	if err != nil {
		log.Printf("failed to create server: %s", err)
		os.Exit(1)
	}
	if err := srv.ListenAndAccept(); err != nil {
		log.Printf("failed to start server: %s", err)
		os.Exit(1)
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
			s := NewTerminalService()
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
